package mongo

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

type MongoDatastore struct {
	txn.BaseDataStore
	conn       MongoConnectionInterface
	readCache  map[string]MongoItem
	writeCache map[string]MongoItem
	se         serializer.Serializer
}

func NewMongoDatastore(name string, conn MongoConnectionInterface) *MongoDatastore {
	return &MongoDatastore{
		BaseDataStore: txn.BaseDataStore{Name: name},
		conn:          conn,
		readCache:     make(map[string]MongoItem),
		writeCache:    make(map[string]MongoItem),
		se:            serializer.NewJSONSerializer(),
	}
}

func (r *MongoDatastore) Start() error {
	return r.conn.Connect()
}

func (r *MongoDatastore) Read(key string, value any) error {

	// if the record is in the writeCache
	if item, ok := r.writeCache[key]; ok {
		// if the record is marked as deleted
		if item.IsDeleted {
			return errors.New("key not found")
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
			return r.readAsCommitted(item, value)
		}
		return errors.New("dirty Read")
	}
	return errors.New("key not found")

}

func (r *MongoDatastore) readAsCommitted(item MongoItem, value any) error {
	if item.TValid.Before(r.Txn.TxnStartTime) {
		// if the record has been deleted
		if item.IsDeleted {
			return errors.New("key not found")
		}
		err := r.se.Deserialize([]byte(item.Value), value)
		if err != nil {
			return err
		}
		r.readCache[item.Key] = item
		return nil
	}
	var preItem MongoItem
	err := r.se.Deserialize([]byte(item.Prev), &preItem)
	if err != nil {
		// The transaction needs to be aborted
		return err
	}
	if preItem.TValid.Before(r.Txn.TxnStartTime) {
		// if the record has been deleted
		if preItem.IsDeleted {
			return errors.New("key not found")
		}
		err := r.se.Deserialize([]byte(preItem.Value), value)
		if err != nil {
			return err
		}
		r.readCache[item.Key] = preItem
		return nil
	} else {
		return errors.New("key not found")
	}
}

func (r *MongoDatastore) Write(key string, value any) error {
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
		oldItem, err := r.conn.GetItem(key)
		if err != nil {
			version = 0
		} else {
			version = oldItem.Version
		}
	}
	// else Write a record to the cache
	r.writeCache[key] = MongoItem{
		Key:       key,
		Value:     str,
		TxnId:     r.Txn.TxnId,
		TValid:    time.Now(),
		TLease:    time.Now().Add(config.DefaultConfig.LeaseTime),
		Version:   version,
		IsDeleted: false,
	}
	return nil
}

func (r *MongoDatastore) Prev(key string, record string) {
	//TODO implement me
	panic("implement me")
}

func (r *MongoDatastore) Delete(key string) error {
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

	r.writeCache[key] = MongoItem{
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

func (r *MongoDatastore) conditionalUpdate(item txn.Item) error {
	memItem := item.(MongoItem)
	oldItem, err := r.conn.GetItem(memItem.Key)
	if err != nil {
		// this is a new record
		newItem, _ := r.updateMetadata(memItem, MongoItem{})
		// Write the new item to the data store
		return r.conn.PutItem(newItem.Key, newItem)
	}

	newItem, err := r.updateMetadata(memItem, oldItem)
	if err != nil {
		return err
	}
	if err = r.conn.ConditionalUpdate(newItem.Key, newItem); err != nil {
		return errors.New("write conflicted: the record has been modified by others")
	}
	return nil
}

func (r *MongoDatastore) updateMetadata(newItem MongoItem, oldItem MongoItem) (MongoItem, error) {
	// clear the Prev field of the old item
	oldItem.Prev = ""
	// update record's metadata
	bs, err := r.se.Serialize(oldItem)
	if err != nil {
		return newItem, err
	}
	newItem.Prev = string(bs)
	newItem.TxnState = config.PREPARED
	newItem.TValid = r.Txn.TxnCommitTime
	newItem.TLease = r.Txn.TxnCommitTime.Add(config.DefaultConfig.LeaseTime)
	return newItem, nil
}

func (r *MongoDatastore) Prepare() error {
	records := make([]MongoItem, 0, len(r.writeCache))
	for _, v := range r.writeCache {
		records = append(records, v)
	}
	// sort records by key
	// TODO: global consistent hash order
	slices.SortFunc(
		records, func(i, j MongoItem) int {
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
func (r *MongoDatastore) Commit() error {
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
	r.writeCache = make(map[string]MongoItem)
	r.readCache = make(map[string]MongoItem)
	return nil
}

// Abort discards the changes made in the current transaction.
// If hasCommitted is false, it clears the write cache.
// If hasCommitted is true, it rolls back the changes made by the current transaction.
// It returns an error if there is any issue during the rollback process.
func (r *MongoDatastore) Abort(hasCommitted bool) error {

	if !hasCommitted {
		r.writeCache = make(map[string]MongoItem)
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
	r.readCache = make(map[string]MongoItem)
	r.writeCache = make(map[string]MongoItem)
	return nil
}

func (r *MongoDatastore) Recover(key string) {
	//TODO implement me
	panic("implement me")
}

// rollback overwrites the record with the application data and metadata that found in field Prev
func (r *MongoDatastore) rollback(item MongoItem) (MongoItem, error) {
	var newItem MongoItem
	err := r.se.Deserialize([]byte(item.Prev), &newItem)
	if err != nil {
		return MongoItem{}, errors.Join(errors.New("rollback failed"), err)
	}
	err = r.conn.PutItem(item.Key, newItem)
	if err != nil {
		return MongoItem{}, err
	}

	return newItem, err
}

// rollForward makes the record metadata with COMMITTED state
func (r *MongoDatastore) rollForward(item MongoItem) (MongoItem, error) {
	// var oldItem MongoItem
	// r.conn.Get(item.Key, &oldItem)
	item.TxnState = config.COMMITTED
	err := r.conn.PutItem(item.Key, item)
	return item, err
}

// GetName returns the name of the MemoryDatastore.
func (r *MongoDatastore) GetName() string {
	return r.Name
}

// SetTxn sets the transaction for the MemoryDatastore.
// It takes a pointer to a Transaction as input and assigns it to the Txn field of the MemoryDatastore.
func (r *MongoDatastore) SetTxn(txn *txn.Transaction) {
	r.Txn = txn
}

func (r *MongoDatastore) ReadTSR(txnId string) (config.State, error) {
	var txnState config.State
	state, err := r.conn.Get(txnId)
	if err != nil {
		return txnState, err
	}
	txnState = config.State(util.ToInt(state))
	return txnState, nil
}

// WriteTSR writes the transaction state (txnState) associated with the given transaction ID (txnId) to the Mongo datastore.
// It returns an error if the write operation fails.
func (r *MongoDatastore) WriteTSR(txnId string, txnState config.State) error {
	return r.conn.Put(txnId, util.ToString(txnState))
}

// DeleteTSR deletes a transaction with the given transaction ID from the Mongo datastore.
// It returns an error if the deletion operation fails.
func (r *MongoDatastore) DeleteTSR(txnId string) error {
	return r.conn.Delete(txnId)
}

// Copy returns a new instance of MongoDatastore with the same name and connection.
// It is used to create a copy of the MongoDatastore object.
func (r *MongoDatastore) Copy() txn.Datastore {
	return NewMongoDatastore(r.Name, r.conn)
}
