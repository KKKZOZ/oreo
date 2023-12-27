package redis

import (
	"cmp"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/txn"
)

type RedisDatastore struct {
	txn.BaseDataStore
	conn       RedisConnection
	readCache  map[string]RedisItem
	writeCache map[string]RedisItem
}

func NewRedisDatastore(name string, conn RedisConnection) *RedisDatastore {
	return &RedisDatastore{
		BaseDataStore: txn.BaseDataStore{Name: name},
		conn:          conn,
		readCache:     make(map[string]RedisItem),
		writeCache:    make(map[string]RedisItem),
	}
}

func (r *RedisDatastore) Start() error {
	err := r.conn.Connect()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisDatastore) Read(key string, value any) error {

	// if the record is in the writeCache
	if item, ok := r.writeCache[key]; ok {
		// if the record is marked as deleted
		if item.IsDeleted {
			return errors.New("key not found")
		}
		return json.Unmarshal([]byte(item.Value), value)
	}

	// if the record is in the readCache
	if item, ok := r.readCache[key]; ok {
		return json.Unmarshal([]byte(item.Value), value)
	}

	// else get if from connection
	item, err := r.conn.GetItem(key)
	if err != nil {
		return err
	}

	if item.TxnState == config.COMMITTED {
		return r.readAsCommitted(item, value)
	}
	if item.TxnState == config.PREPARED {

		//TODO: what if the state is ABORTED
		_, err := r.Txn.GetTSRState(item.TxnId)
		if err == nil {
			// if TSR exists
			// roll forward the record
			item, err = r.rollForward(item)
			if err != nil {
				return err
			}
			return r.readAsCommitted(item, value)
		}
		// if TSR does not exist
		// and if t_lease has expired
		// we should rollback the record
		// because the transaction that modified the record has been aborted
		if item.TLease.Before(r.Txn.TxnStartTime) {
			// the corresponding transaction is considered ABORTED
			r.Txn.WriteTSR(item.TxnId, config.ABORTED)
			item, err := r.rollback(item)
			if err != nil {
				return err
			}
			return r.readAsCommitted(item, value)
		}
		return errors.New("dirty Read")
	}
	return errors.New("key not found")

}

func (r *RedisDatastore) readAsCommitted(item RedisItem, value any) error {
	if item.TValid.Before(r.Txn.TxnStartTime) {
		// if the record has been deleted
		if item.IsDeleted {
			return errors.New("key not found")
		}
		err := json.Unmarshal([]byte(item.Value), value)
		if err != nil {
			return err
		}
		r.readCache[item.Key] = item
		return nil
	}
	var preItem RedisItem
	err := json.Unmarshal([]byte(item.Prev), &preItem)
	if err != nil {
		// The transaction needs to be aborted
		return err
	}
	if preItem.TValid.Before(r.Txn.TxnStartTime) {
		// if the record has been deleted
		if preItem.IsDeleted {
			return errors.New("key not found")
		}
		err := json.Unmarshal([]byte(preItem.Value), value)
		if err != nil {
			return err
		}
		r.readCache[item.Key] = preItem
		return nil
	} else {
		return errors.New("key not found")
	}
}

func (r *RedisDatastore) Write(key string, value any) error {
	jsonString := util.ToJSONString(value)
	// if the record is in the writeCache
	if item, ok := r.writeCache[key]; ok {
		item.Value, item.IsDeleted = jsonString, false
		r.writeCache[key] = item
		return nil
	}

	var version int
	if item, ok := r.readCache[key]; ok {
		version = item.Version
	} else {
		version = 1
	}
	// else Write a record to the cache
	r.writeCache[key] = RedisItem{
		Key:       key,
		Value:     jsonString,
		TxnId:     r.Txn.TxnId,
		TValid:    time.Now(),
		TLease:    time.Now().Add(config.LeastTime * time.Millisecond),
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

func (r *RedisDatastore) conditionalUpdate(item txn.Item) error {
	memItem := item.(RedisItem)

	oldItem, err := r.conn.GetItem(memItem.Key)
	if err != nil {
		// this is a new record
		newItem := r.updateMetadata(memItem, RedisItem{})
		// Write the new item to the data store
		if err = r.conn.PutItem(newItem.Key, newItem); err != nil {
			return err
		} else {
			return nil
		}
	}

	// TODO: 需不需要根据情况 roll forward ?
	if oldItem.TxnState == config.PREPARED {
		return errors.New("write conflicted: the record is in PREPARED state")
	}

	// the old item is in COMMITTED state
	if oldItem.Version == memItem.Version {
		// we can do nothing when the record is deleted

		// update record's metadata
		newItem := r.updateMetadata(memItem, oldItem)
		// Write the new item to the data store
		if err = r.conn.PutItem(newItem.Key, newItem); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return errors.New("write conflicted: the record has been modified by others")
	}
}

func (r *RedisDatastore) updateMetadata(newItem RedisItem, oldItem RedisItem) RedisItem {
	// clear the Prev field of the old item
	oldItem.Prev = ""
	// update record's metadata
	newItem.Prev = util.ToJSONString(oldItem)
	newItem.Version++
	newItem.TxnState = config.PREPARED
	newItem.TValid = r.Txn.TxnCommitTime
	newItem.TLease = r.Txn.TxnCommitTime.Add(config.LeastTime * time.Millisecond)

	return newItem
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
	r.writeCache = make(map[string]RedisItem)
	return nil
}

func (r *RedisDatastore) Recover(key string) {
	//TODO implement me
	panic("implement me")
}

// rollback overwrites the record with the application data and metadata that found in field Prev
func (r *RedisDatastore) rollback(item RedisItem) (RedisItem, error) {
	var newItem RedisItem
	err := json.Unmarshal([]byte(item.Prev), &newItem)
	if err != nil {
		return RedisItem{}, err
	}
	err = r.conn.PutItem(item.Key, newItem)
	if err != nil {
		return RedisItem{}, err
	}

	return newItem, err
}

// rollForward makes the record metadata with COMMITTED state
func (r *RedisDatastore) rollForward(item RedisItem) (RedisItem, error) {
	// var oldItem RedisItem
	// r.conn.Get(item.Key, &oldItem)
	item.TxnState = config.COMMITTED
	err := r.conn.PutItem(item.Key, item)
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
	txnState = config.State(util.ToInt(state))
	if err != nil {
		return txnState, err
	}
	return txnState, nil
}

// WriteTSR writes the transaction state (txnState) associated with the given transaction ID (txnId) to the memory datastore.
// It returns an error if the write operation fails.
func (r *RedisDatastore) WriteTSR(txnId string, txnState config.State) error {
	return r.conn.Put(txnId, util.ToString(txnState))
}

// DeleteTSR deletes a transaction with the given transaction ID from the memory datastore.
// It returns an error if the deletion operation fails.
func (r *RedisDatastore) DeleteTSR(txnId string) error {
	return r.conn.Delete(txnId)
}

func (r *RedisDatastore) Copy() txn.Datastore {
	return NewRedisDatastore(r.Name, r.conn)
}
