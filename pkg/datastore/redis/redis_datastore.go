package redis

import (
	"cmp"
	"errors"
	"slices"
	"time"

	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"
)

type RedisDatastore struct {
	txn.BaseDataStore
	conn       RedisConnectionInterface
	readCache  map[string]RedisItem
	writeCache map[string]RedisItem
	se         serializer.Serializer
}

func NewRedisDatastore(name string, conn RedisConnectionInterface) *RedisDatastore {
	return &RedisDatastore{
		BaseDataStore: txn.BaseDataStore{Name: name},
		conn:          conn,
		readCache:     make(map[string]RedisItem),
		writeCache:    make(map[string]RedisItem),
		se:            serializer.NewJSONSerializer(),
	}
}

func (r *RedisDatastore) Start() error {
	return r.conn.Connect()
}

func (r *RedisDatastore) Read(key string, value any) error {

	// if the record is in the writeCache
	if item, ok := r.writeCache[key]; ok {
		// if the record is marked as deleted
		if item.IsDeleted {
			return txn.KeyNotFound
		}
		return r.se.Deserialize([]byte(item.Value), value)
	}

	// if the record is in the readCache
	if item, ok := r.readCache[key]; ok {
		return r.se.Deserialize([]byte(item.Value), value)
	}

	// else get if from connection
	item, err := r.conn.GetItem(key)
	if err != nil {
		return err
	}

	err = r.basicVisibilityProcessor(item, value)
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisDatastore) basicVisibilityProcessor(item RedisItem, value any) error {
	if item.TxnState == config.COMMITTED {
		return r.readAsCommitted(item, value)
	}

	if item.TxnState == config.PREPARED {
		state, err := r.Txn.GetTSRState(item.TxnId)
		if err == nil {
			switch state {
			// if TSR exists and the TSR is in COMMITTED state
			// roll forward the record
			case config.COMMITTED:
				item, err = r.rollForward(item)
				if err != nil {
					return err
				}
				return r.readAsCommitted(item, value)
			// if TSR exists and the TSR is in ABORTED state
			// we should rollback the record
			// because the transaction that modified the record has been aborted
			case config.ABORTED:
				item, err := r.rollback(item)
				if err != nil {
					return err
				}
				if item.Equal(RedisItem{}) {
					return txn.KeyNotFound
				}
				return r.readAsCommitted(item, value)
			}
		}
		// if TSR does not exist
		// and if t_lease has expired
		// we should rollback the record
		if item.TLease.Before(r.Txn.TxnStartTime) {
			// the corresponding transaction is considered ABORTED
			err := r.Txn.WriteTSR(item.TxnId, config.ABORTED)
			if err != nil {
				return err
			}
			item, err := r.rollback(item)
			if err != nil {
				return err
			}
			if item.Equal(RedisItem{}) {
				return txn.KeyNotFound
			}
			return r.readAsCommitted(item, value)
		}
		return errors.New("dirty Read")
	}
	return txn.KeyNotFound
}

func (r *RedisDatastore) readAsCommitted(item RedisItem, value any) error {
	curItem := item
	for i := 1; i <= config.Config.MaxRecordLength; i++ {

		if curItem.TValid.Before(r.Txn.TxnStartTime) {
			// if the record has been deleted
			if curItem.IsDeleted {
				return txn.KeyNotFound
			}
			err := r.se.Deserialize([]byte(curItem.Value), value)
			if err != nil {
				return err
				// return txn.DeserializeError.Join(err)
			}
			r.readCache[curItem.Key] = curItem
			return nil
		}
		if i == config.Config.MaxRecordLength {
			break
		}

		// if prev is empty
		if curItem.Prev == "" {
			return txn.KeyNotFound
		}

		// get the previous record
		var preItem RedisItem
		err := r.se.Deserialize([]byte(curItem.Prev), &preItem)
		if err != nil {
			// The transaction needs to be aborted
			return err
		}
		curItem = preItem
	}

	return txn.KeyNotFound
}

func (r *RedisDatastore) Write(key string, value any) error {
	bs, err := r.se.Serialize(value)
	if err != nil {
		return err
	}
	str := string(bs)
	// if the record is in the writeCache
	if item, ok := r.writeCache[key]; ok {
		item.Value, item.IsDeleted = str, false
		r.writeCache[key] = item
		return nil
	}

	var version int
	if item, ok := r.readCache[key]; ok {
		version = item.Version
	} else {
		// TODO: we should handler version problem in conditionalUpdate?
		// oldItem, err := r.conn.GetItem(key)
		// if err != nil {
		// 	version = 0
		// } else {
		// 	version = oldItem.Version
		// }
	}
	// else Write a record to the cache
	r.writeCache[key] = RedisItem{
		Key:       key,
		Value:     str,
		TxnId:     r.Txn.TxnId,
		TValid:    time.Now(),
		TLease:    time.Now().Add(config.Config.LeaseTime),
		Version:   version,
		IsDeleted: false,
	}
	return nil
}

func (r *RedisDatastore) Prev(key string, record string) {
	//TODO implement me
	panic("implement me")
}

func (r *RedisDatastore) Delete(key string) error {
	// if the record is in the writeCache
	if item, ok := r.writeCache[key]; ok {
		if item.IsDeleted {
			return errors.New("key not found")
		}
		item.IsDeleted = true
		r.writeCache[key] = item
		return nil
	}

	// if the record is in the readCache
	// we can get the corresponding version
	version := 0
	if item, ok := r.readCache[key]; ok {
		version = item.Version
	} else {
		// else write a Delete record to the writeCache
		// first we have to get the corresponding version
		// TODO: should we first read into the cache?
		item, err := r.conn.GetItem(key)
		if err != nil {
			return err
		}
		version = item.Version
	}

	r.writeCache[key] = RedisItem{
		Key:       key,
		IsDeleted: true,
		TxnId:     r.Txn.TxnId,
		TxnState:  config.COMMITTED,
		TValid:    time.Now(),
		TLease:    time.Now(),
		Version:   version,
	}
	return nil
}

func (r *RedisDatastore) doConditionalUpdate(cacheItem RedisItem, dbItem RedisItem) error {
	// newItem, err := r.updateMetadata(cacheItem, dbItem)

	if dbItem.Equal(RedisItem{}) {
		newItem, err := r.updateMetadata(cacheItem, dbItem)
		if err != nil {
			return err
		}
		if err = r.conn.ConditionalUpdate(newItem.Key, newItem, true); err != nil {
			return errors.New("write conflicted: the record has been modified by others")
		}
		return nil
	}

	curItem := dbItem
	for i := 1; i <= config.Config.MaxRecordLength; i++ {

		if curItem.TValid.Before(r.Txn.TxnStartTime) {
			// if the record has been deleted
			if curItem.IsDeleted {
				newItem, err := r.updateMetadata(cacheItem, curItem)
				if err != nil {
					return err
				}
				return r.conn.ConditionalUpdate(newItem.Key, newItem, false)
			}
			newItem, err := r.updateMetadata(cacheItem, curItem)
			if err != nil {
				return err
			}
			return r.conn.ConditionalUpdate(newItem.Key, newItem, false)
		}
		if i == config.Config.MaxRecordLength {
			break
		}

		// if prev is empty
		if curItem.Prev == "" {
			newItem, err := r.updateMetadata(cacheItem, curItem)
			if err != nil {
				return err
			}
			return r.conn.ConditionalUpdate(newItem.Key, newItem, true)
		}

		// get the previous record
		var preItem RedisItem
		err := r.se.Deserialize([]byte(curItem.Prev), &preItem)
		if err != nil {
			// The transaction needs to be aborted
			return err
		}
		curItem = preItem
	}

	// if it can not find a corresponding version in linked list
	return txn.VersionMismatch

	// newItem, err := r.updateMetadata(cacheItem, dbItem)
	// if err != nil {
	// 	return err
	// }

	// return r.conn.ConditionalUpdate(newItem.Key, newItem, true)
}

func (r *RedisDatastore) conditionalUpdate(cacheItem RedisItem) error {

	dbItem, err := r.conn.GetItem(cacheItem.Key)
	if err != nil {
		if err == txn.KeyNotFound {
			return r.doConditionalUpdate(cacheItem, RedisItem{})
			// // this is a new record
			// newItem, _ := r.updateMetadata(cacheItem, RedisItem{})
			// // Write the new item to the data store
			// return r.conn.PutItem(newItem.Key, newItem)
			// // TODO: we should handle atomic create new record here
			// // return r.conn.ConditionalUpdate(newItem.Key, newItem)
		}
		return err
	}

	if dbItem.TxnState == config.COMMITTED {
		return r.doConditionalUpdate(cacheItem, dbItem)
	}

	if dbItem.TxnState == config.PREPARED {
		state, err := r.Txn.GetTSRState(dbItem.TxnId)
		if err == nil {
			switch state {
			// if TSR exists and the TSR is in COMMITTED state
			// roll forward the record
			case config.COMMITTED:
				dbItem, err = r.rollForward(dbItem)
				if err != nil {
					return err
				}
				return r.doConditionalUpdate(cacheItem, dbItem)
			// if TSR exists and the TSR is in ABORTED state
			// we should rollback the record
			// because the transaction that modified the record has been aborted
			case config.ABORTED:
				dbItem, err = r.rollback(dbItem)
				if err != nil {
					return err
				}
				// if item.Equal(RedisItem{}) {
				// 	return txn.KeyNotFound
				// }
				return r.doConditionalUpdate(cacheItem, dbItem)
			}
		}
		// if TSR does not exist
		// and if t_lease has expired
		// we should rollback the record
		if dbItem.TLease.Before(r.Txn.TxnStartTime) {
			// the corresponding transaction is considered ABORTED
			err := r.Txn.WriteTSR(dbItem.TxnId, config.ABORTED)
			if err != nil {
				return err
			}
			dbItem, err = r.rollback(dbItem)
			if err != nil {
				return err
			}
			// if dbItem.Equal(RedisItem{}) {
			// 	return txn.KeyNotFound
			// }
			return r.doConditionalUpdate(cacheItem, dbItem)
		}
		return txn.VersionMismatch
	}
	return nil
}

// truncate truncates the linked list of RedisItems if the length exceeds the maximum record length defined in the configuration.
// It takes a pointer to a RedisItem as input and returns the truncated RedisItem and an error, if any.
// If the length of the linked list is greater than the maximum record length, it creates a stack of RedisItems and pops the items from the stack until the length is reduced to the maximum record length.
// It then updates the Prev and LinkedLen fields of the RedisItems in the stack accordingly.
// Finally, it returns the last popped RedisItem as the truncated RedisItem.
// If the length of the linked list is less than or equal to the maximum record length, it returns the input RedisItem as is.
func (r *RedisDatastore) truncate(newItem *RedisItem) (RedisItem, error) {
	maxLen := config.Config.MaxRecordLength

	if newItem.LinkedLen > maxLen {
		stack := util.NewStack[RedisItem]()
		stack.Push(*newItem)
		curItem := newItem
		for i := 1; i <= maxLen-1; i++ {
			var preItem RedisItem
			err := r.se.Deserialize([]byte(curItem.Prev), &preItem)
			if err != nil {
				return RedisItem{}, errors.New("Unmarshal error: " + err.Error())
			}
			curItem = &preItem
			stack.Push(*curItem)
		}

		tarItem, err := stack.Pop()
		if err != nil {
			return RedisItem{}, errors.New("Pop error: " + err.Error())
		}
		tarItem.Prev = ""
		tarItem.LinkedLen = 1

		for !stack.IsEmpty() {
			item, err := stack.Pop()
			if err != nil {
				return RedisItem{}, errors.New("Pop error: " + err.Error())
			}
			bs, err := r.se.Serialize(tarItem)
			if err != nil {
				return RedisItem{}, errors.New("Serialize error: " + err.Error())
			}
			item.Prev = string(bs)
			item.LinkedLen = tarItem.LinkedLen + 1
			tarItem = item
		}
		return tarItem, nil
	} else {
		return *newItem, nil
	}
}

func (r *RedisDatastore) updateMetadata(newItem RedisItem, oldItem RedisItem) (RedisItem, error) {
	if oldItem == (RedisItem{}) {
		newItem.LinkedLen = 1
	} else {
		newItem.LinkedLen = oldItem.LinkedLen + 1
		bs, err := r.se.Serialize(oldItem)
		if err != nil {
			return RedisItem{}, err
		}
		newItem.Prev = string(bs)
		newItem.Version = oldItem.Version
	}

	// truncate the record
	newItem, err := r.truncate(&newItem)
	if err != nil {
		return RedisItem{}, err
	}

	newItem.TxnState = config.PREPARED
	newItem.TValid = r.Txn.TxnCommitTime
	newItem.TLease = r.Txn.TxnCommitTime.Add(config.Config.LeaseTime)
	return newItem, nil
}

func (r *RedisDatastore) Prepare() error {
	records := make([]RedisItem, 0, len(r.writeCache))
	for _, v := range r.writeCache {
		records = append(records, v)
	}
	// sort records by key
	// TODO: global consistent hash order
	slices.SortFunc(
		records, func(i, j RedisItem) int {
			return cmp.Compare(i.Key, j.Key)
		},
	)
	for _, v := range records {
		err := r.conditionalUpdate(v)
		if err != nil {
			return err
		}
	}
	return nil
}

// Commit updates the state of records in the data store to COMMITTED.
// It iterates over the write cache and updates each record's state to COMMITTED.
// After updating the records, it clears the write cache.
// Returns an error if there is any issue updating the records or clearing the cache.
func (r *RedisDatastore) Commit() error {
	// update record's state to the COMMITTED state in the data store
	for _, v := range r.writeCache {
		item, err := r.conn.GetItem(v.Key)
		if err != nil {
			return err
		}
		item.TxnState = config.COMMITTED
		err = r.conn.PutItem(v.Key, item)
		if err != nil {
			return err
		}
	}
	// clear the cache
	r.writeCache = make(map[string]RedisItem)
	r.readCache = make(map[string]RedisItem)
	return nil
}

// Abort discards the changes made in the current transaction.
// If hasCommitted is false, it clears the write cache.
// If hasCommitted is true, it rolls back the changes made by the current transaction.
// It returns an error if there is any issue during the rollback process.
func (r *RedisDatastore) Abort(hasCommitted bool) error {

	if !hasCommitted {
		r.writeCache = make(map[string]RedisItem)
		return nil
	}

	for _, v := range r.writeCache {
		item, err := r.conn.GetItem(v.Key)
		if err != nil {
			return err
		}
		// if the record has been modified by this transaction
		curTxnId := r.Txn.TxnId
		if item.TxnId == curTxnId {
			r.rollback(item)
		}
	}
	r.readCache = make(map[string]RedisItem)
	r.writeCache = make(map[string]RedisItem)
	return nil
}

func (r *RedisDatastore) Recover(key string) {
	//TODO implement me
	panic("implement me")
}

// rollback overwrites the record with the application data
// and metadata that found in field Prev.
// if the `Prev` is empty, it simply deletes the record
func (r *RedisDatastore) rollback(item RedisItem) (RedisItem, error) {

	if item.Prev == "" {
		// TODO: mark the record as deleted?
		err := r.conn.Delete(item.Key)
		return RedisItem{}, err
	}

	var newItem RedisItem
	err := r.se.Deserialize([]byte(item.Prev), &newItem)
	if err != nil {
		return RedisItem{}, errors.Join(errors.New("rollback failed"), err)
	}
	// try to rollback through ConditionalUpdate
	newItem.Version = item.Version
	err = r.conn.ConditionalUpdate(item.Key, newItem, false)
	// err = r.conn.PutItem(item.Key, newItem)
	if err != nil {
		return RedisItem{}, errors.Join(errors.New("rollback failed"), err)
	}
	// update the version
	newItem.Version++
	return newItem, err
}

// rollForward makes the record metadata with COMMITTED state
func (r *RedisDatastore) rollForward(item RedisItem) (RedisItem, error) {
	// var oldItem RedisItem
	// r.conn.Get(item.Key, &oldItem)
	item.TxnState = config.COMMITTED
	err := r.conn.ConditionalUpdate(item.Key, item, false)
	if err != nil {
		return RedisItem{}, errors.Join(errors.New("rollForward failed"), err)
	}
	item.Version++
	// err := r.conn.PutItem(item.Key, item)
	return item, err
}

// GetName returns the name of the MemoryDatastore.
func (r *RedisDatastore) GetName() string {
	return r.Name
}

// SetTxn sets the transaction for the MemoryDatastore.
// It takes a pointer to a Transaction as input and assigns it to the Txn field of the MemoryDatastore.
func (r *RedisDatastore) SetTxn(txn *txn.Transaction) {
	r.Txn = txn
}

func (r *RedisDatastore) ReadTSR(txnId string) (config.State, error) {
	var txnState config.State
	state, err := r.conn.Get(txnId)
	if err != nil {
		return txnState, err
	}
	txnState = config.State(util.ToInt(state))
	return txnState, nil
}

// WriteTSR writes the transaction state (txnState) associated with the given transaction ID (txnId) to the Redis datastore.
// It returns an error if the write operation fails.
func (r *RedisDatastore) WriteTSR(txnId string, txnState config.State) error {
	return r.conn.Put(txnId, util.ToString(txnState))
}

// DeleteTSR deletes a transaction with the given transaction ID from the Redis datastore.
// It returns an error if the deletion operation fails.
func (r *RedisDatastore) DeleteTSR(txnId string) error {
	return r.conn.Delete(txnId)
}

// Copy returns a new instance of RedisDatastore with the same name and connection.
// It is used to create a copy of the RedisDatastore object.
func (r *RedisDatastore) Copy() txn.Datastore {
	return NewRedisDatastore(r.Name, r.conn)
}
