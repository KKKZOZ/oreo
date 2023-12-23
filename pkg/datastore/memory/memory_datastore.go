package memory

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

type MemoryDatastore struct {
	txn.BaseDataStore
	conn       MemoryConnectionInterface
	readCache  map[string]MemoryItem
	writeCache map[string]MemoryItem
}

type MemoryItem struct {
	Key       string
	Value     string
	TxnId     string
	TxnState  config.State
	TValid    time.Time
	TLease    time.Time
	Prev      string
	isDeleted bool
	Version   int
}

func (m MemoryItem) GetKey() string {
	return m.Key
}

func NewMemoryDatastore(name string, conn MemoryConnectionInterface) *MemoryDatastore {
	return &MemoryDatastore{
		BaseDataStore: txn.BaseDataStore{Name: name},
		conn:          conn,
		readCache:     make(map[string]MemoryItem),
		writeCache:    make(map[string]MemoryItem),
	}
}

func (m *MemoryDatastore) Start() error {
	err := m.conn.Connect()
	if err != nil {
		return err
	}
	return nil
}

func (m *MemoryDatastore) Read(key string, value any) error {

	var item MemoryItem

	// if the record is in the writeCache
	if item, ok := m.writeCache[key]; ok {
		// if the record is marked as deleted
		if item.isDeleted {
			return errors.New("key not found")
		}
		return json.Unmarshal([]byte(item.Value), value)
	}

	// if the record is in the readCache
	if item, ok := m.readCache[key]; ok {
		return json.Unmarshal([]byte(item.Value), value)
	}

	// else get if from connection
	err := m.conn.Get(key, &item)
	if err != nil {
		return err
	}

	if item.TxnState == config.COMMITTED {
		return m.readAsCommitted(item, value)
	}
	if item.TxnState == config.PREPARED {

		//TODO: what if the state is ABORTED
		_, err := m.Txn.GetTSRState(item.TxnId)
		if err == nil {
			// if TSR exists
			// roll forward the record
			item, err = m.rollForward(item)
			if err != nil {
				return err
			}
			return m.readAsCommitted(item, value)
		}
		// if TSR does not exist
		// and if t_lease has expired
		// we should rollback the record
		// because the transaction that modified the record has been aborted
		if item.TLease.Before(m.Txn.TxnStartTime) {
			item, err := m.rollback(item)
			if err != nil {
				return err
			}
			return m.readAsCommitted(item, value)
		}
		return errors.New("dirty Read")
	}
	return errors.New("key not found")

}

func (m *MemoryDatastore) readAsCommitted(item MemoryItem, value any) error {
	if item.TValid.Before(m.Txn.TxnStartTime) {
		// if the record has been deleted
		if item.isDeleted {
			return errors.New("key not found")
		}
		err := json.Unmarshal([]byte(item.Value), value)
		if err != nil {
			return err
		}
		m.readCache[item.Key] = item
		return nil
	}
	var preItem MemoryItem
	err := json.Unmarshal([]byte(item.Prev), &preItem)
	if err != nil {
		// The transaction needs to be aborted
		return err
	}
	if preItem.TValid.Before(m.Txn.TxnStartTime) {
		// if the record has been deleted
		if preItem.isDeleted {
			return errors.New("key not found")
		}
		err := json.Unmarshal([]byte(preItem.Value), value)
		if err != nil {
			return err
		}
		m.readCache[item.Key] = preItem
		return nil
	} else {
		return errors.New("key not found")
	}
}

func (m *MemoryDatastore) Write(key string, value any) error {
	jsonString := util.ToJSONString(value)
	// if the record is in the writeCache
	if item, ok := m.writeCache[key]; ok {
		item.Value, item.isDeleted = jsonString, false
		m.writeCache[key] = item
		return nil
	}

	var version int
	if item, ok := m.readCache[key]; ok {
		version = item.Version
	} else {
		version = 1
	}
	// else Write a record to the cache
	m.writeCache[key] = MemoryItem{
		Key:       key,
		Value:     jsonString,
		TxnId:     m.Txn.TxnId,
		TValid:    time.Now(),
		TLease:    time.Now().Add(config.LeastTime * time.Millisecond),
		Version:   version,
		isDeleted: false,
	}
	return nil
}

func (m *MemoryDatastore) Prev(key string, record string) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryDatastore) Delete(key string) error {
	// if the record is in the writeCache
	if item, ok := m.writeCache[key]; ok {
		if item.isDeleted {
			return errors.New("key not found")
		}
		item.isDeleted = true
		m.writeCache[key] = item
		return nil
	}

	// if the record is in the readCache
	// we can get the corresponding version

	version := 0
	if item, ok := m.readCache[key]; ok {
		version = item.Version
	} else {
		// else write a Delete record to the writeCache
		// first we have to get the corresponding version
		var item MemoryItem
		err := m.conn.Get(key, &item)
		if err != nil {
			return err
		}
		version = item.Version
	}

	m.writeCache[key] = MemoryItem{
		Key:       key,
		isDeleted: true,
		TxnId:     m.Txn.TxnId,
		TxnState:  config.COMMITTED,
		TValid:    time.Now(),
		TLease:    time.Now(),
		Version:   version,
	}
	return nil
}

func (m *MemoryDatastore) conditionalUpdate(item txn.Item) error {
	memItem := item.(MemoryItem)

	key := "memory" + memItem.Key
	err := m.Txn.Lock(key, memItem.TxnId, 100*time.Millisecond)
	if err != nil {
		return err
	}
	defer m.Txn.Unlock(key, memItem.TxnId)

	var oldItem MemoryItem
	err = m.conn.Get(memItem.Key, &oldItem)
	if err != nil {
		// this is a new record
		newItem := m.updateMetadata(memItem, MemoryItem{})
		// Write the new item to the data store
		if err = m.conn.Put(newItem.Key, newItem); err != nil {
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
		newItem := m.updateMetadata(memItem, oldItem)
		// Write the new item to the data store
		if err = m.conn.Put(newItem.Key, newItem); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return errors.New("write conflicted: the record has been modified by others")
	}
}

func (m *MemoryDatastore) updateMetadata(newItem MemoryItem, oldItem MemoryItem) MemoryItem {
	// clear the Prev field of the old item
	oldItem.Prev = ""
	// update record's metadata
	newItem.Prev = util.ToJSONString(oldItem)
	newItem.Version++
	newItem.TxnState = config.PREPARED
	newItem.TValid = m.Txn.TxnCommitTime
	newItem.TLease = m.Txn.TxnCommitTime.Add(config.LeastTime * time.Millisecond)

	return newItem
}

func (m *MemoryDatastore) Prepare() error {
	records := make([]MemoryItem, 0, len(m.writeCache))
	for _, v := range m.writeCache {
		records = append(records, v)
	}
	// sort records by key
	// TODO: global consistent hash order
	slices.SortFunc(
		records, func(i, j MemoryItem) int {
			return cmp.Compare(i.Key, j.Key)
		},
	)
	for _, v := range records {
		err := m.conditionalUpdate(v)
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
func (m *MemoryDatastore) Commit() error {
	// update record's state to the COMMITTED state in the data store
	for _, v := range m.writeCache {
		var item MemoryItem
		err := m.conn.Get(v.Key, &item)
		if err != nil {
			return err
		}
		item.TxnState = config.COMMITTED
		err = m.conn.Put(v.Key, item)
		if err != nil {
			return err
		}
	}
	// clear the cache
	m.writeCache = make(map[string]MemoryItem)
	m.readCache = make(map[string]MemoryItem)
	return nil
}

// Abort discards the changes made in the current transaction.
// If hasCommitted is false, it clears the write cache.
// If hasCommitted is true, it rolls back the changes made by the current transaction.
// It returns an error if there is any issue during the rollback process.
func (m *MemoryDatastore) Abort(hasCommitted bool) error {
	if !hasCommitted {
		m.writeCache = make(map[string]MemoryItem)
		return nil
	}

	for _, v := range m.writeCache {
		var item MemoryItem
		err := m.conn.Get(v.Key, &item)
		if err != nil {
			return err
		}
		// if the record has been modified by this transaction
		curTxnId := m.Txn.TxnId
		if item.TxnId == curTxnId {
			m.rollback(item)
		}
	}
	m.writeCache = make(map[string]MemoryItem)
	return nil
}

func (m *MemoryDatastore) Recover(key string) {
	//TODO implement me
	panic("implement me")
}

// rollback overwrites the record with the application data and metadata that found in field Prev
func (m *MemoryDatastore) rollback(item MemoryItem) (MemoryItem, error) {
	var newItem MemoryItem
	err := json.Unmarshal([]byte(item.Prev), &newItem)
	if err != nil {
		return MemoryItem{}, err
	}
	err = m.conn.Put(item.Key, newItem)
	return newItem, err
}

// rollForward makes the record metadata with COMMITTED state
func (m *MemoryDatastore) rollForward(item MemoryItem) (MemoryItem, error) {
	// var oldItem MemoryItem
	// m.conn.Get(item.Key, &oldItem)
	item.TxnState = config.COMMITTED
	err := m.conn.Put(item.Key, item)
	return item, err
}

// GetName returns the name of the MemoryDatastore.
func (m *MemoryDatastore) GetName() string {
	return m.Name
}

// SetTxn sets the transaction for the MemoryDatastore.
// It takes a pointer to a Transaction as input and assigns it to the Txn field of the MemoryDatastore.
func (m *MemoryDatastore) SetTxn(txn *txn.Transaction) {
	m.Txn = txn
}

func (m *MemoryDatastore) ReadTSR(txnId string) (config.State, error) {
	var txnState config.State
	err := m.conn.Get(txnId, &txnState)
	if err != nil {
		return txnState, err
	}
	return txnState, nil
}

// WriteTSR writes the transaction state (txnState) associated with the given transaction ID (txnId) to the memory datastore.
// It returns an error if the write operation fails.
func (m *MemoryDatastore) WriteTSR(txnId string, txnState config.State) error {
	return m.conn.Put(txnId, txnState)
}

// DeleteTSR deletes a transaction with the given transaction ID from the memory datastore.
// It returns an error if the deletion operation fails.
func (m *MemoryDatastore) DeleteTSR(txnId string) error {
	return m.conn.Delete(txnId)
}

func (m *MemoryDatastore) Copy() txn.Datastore {
	return NewMemoryDatastore(m.Name, m.conn)
}
