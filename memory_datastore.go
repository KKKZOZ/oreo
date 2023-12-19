package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"slices"
	"time"
)

type MemoryDatastore struct {
	dataStore
	conn       *MemoryConnection
	cache      map[string]MemoryItem
	versionMap map[string]int
}

type MemoryItem struct {
	Key       string
	Value     string
	TxnId     string
	TxnState  State
	TValid    time.Time
	TLease    time.Time
	Prev      string
	isDeleted bool
	Version   int
}

func NewMemoryDatastore(name string, conn *MemoryConnection) *MemoryDatastore {
	return &MemoryDatastore{
		dataStore:  dataStore{Name: name},
		conn:       conn,
		cache:      make(map[string]MemoryItem),
		versionMap: make(map[string]int),
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

	// if the record is in the cache
	if item, ok := m.cache[key]; ok {
		// if the record is marked as deleted
		if item.isDeleted {
			return errors.New("key not found")
		}
		err := json.Unmarshal([]byte(item.Value), value)
		return err
	}

	// else get if from data
	err := m.conn.Get(key, &item)
	if err != nil {
		return err
	}

	if item.TxnState == COMMITTED {
		return m.readAsCommitted(item, value)
	}
	if item.TxnState == PREPARED {
		var state State
		txnId := item.TxnId
		err := m.Txn.globalDataStore.Read(txnId, &state)
		if err == nil {
			// if TSR exists
			return m.readAsCommitted(item, value)
		}
		// if TSR does not exist
		// if t_lease has expired
		if item.TLease.Before(m.Txn.TxnStartTime) {
			err := m.rollForward(item)
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
		m.versionMap[item.Key] = item.Version
		err := json.Unmarshal([]byte(item.Value), value)
		return err
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
		m.versionMap[item.Key] = preItem.Version
		err := json.Unmarshal([]byte(preItem.Value), value)
		return err
	} else {
		return errors.New("key not found")
	}
}

func (m *MemoryDatastore) Write(key string, value any) error {
	jsonString := toJSONString(value)
	// if the record is in the cache
	if item, ok := m.cache[key]; ok {
		item.Value, item.isDeleted = jsonString, false
		m.cache[key] = item
		return nil
	}
	// else Write a record to the cache
	m.cache[key] = MemoryItem{
		Key:       key,
		Value:     jsonString,
		TxnId:     m.Txn.TxnId,
		TValid:    time.Now(),
		TLease:    time.Now().Add(leastTime * time.Millisecond),
		Version:   m.versionMap[key],
		isDeleted: false,
	}
	return nil
}

func (m *MemoryDatastore) Prev(key string, record string) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryDatastore) Delete(key string) error {
	// if the record is in the cache
	if item, ok := m.cache[key]; ok {
		if item.isDeleted {
			return errors.New("key not found")
		}
		item.isDeleted = true
		m.cache[key] = item
		return nil
	}
	// else Write a Delete record to the cache
	m.cache[key] = MemoryItem{
		Key:       key,
		isDeleted: true,
		TxnId:     m.Txn.TxnId,
		TxnState:  COMMITTED,
		TValid:    time.Now(),
		TLease:    time.Now(),
		Version:   0,
	}
	return nil
}

func (m *MemoryDatastore) conditionalUpdate(item MemoryItem) bool {

	var oldItem MemoryItem
	err := m.conn.Get(item.Key, &oldItem)
	if err != nil {
		// this is a new record
		newItem := m.updateMetadata(item, MemoryItem{})
		// Write the new item to the data store
		if err = m.conn.Put(newItem.Key, newItem); err != nil {
			return false
		} else {
			return true
		}
	}

	// TODO: 需不需要根据情况 roll forward ?
	if oldItem.TxnState == PREPARED {
		return false
	}

	// the old item is in COMMITTED state
	if oldItem.Version == item.Version {
		// we can do nothing when the record is deleted

		// update record's metadata
		newItem := m.updateMetadata(item, oldItem)
		// Write the new item to the data store
		if err = m.conn.Put(newItem.Key, newItem); err != nil {
			return false
		} else {
			return true
		}
	} else {
		return false
	}
}

func (m *MemoryDatastore) updateMetadata(newItem MemoryItem, oldItem MemoryItem) MemoryItem {
	// clear the Prev field of the old item
	oldItem.Prev = ""
	// update record's metadata
	newItem.Prev = toJSONString(oldItem)
	newItem.Version++
	newItem.TxnState = PREPARED
	newItem.TValid = m.Txn.TxnCommitTime
	newItem.TLease = m.Txn.TxnCommitTime.Add(leastTime * time.Millisecond)

	return newItem

}

func (m *MemoryDatastore) Prepare() error {
	records := make([]MemoryItem, 0, len(m.cache))
	for _, v := range m.cache {
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
		ok := m.conditionalUpdate(v)
		if !ok {
			return errors.New("Write conflicted, record key: " + v.Key)
		}
	}
	return nil
}

func (m *MemoryDatastore) Commit() error {
	// update record's state to the COMMITTED state in the data store
	for _, v := range m.cache {
		var item MemoryItem
		err := m.conn.Get(v.Key, &item)
		if err != nil {
			return err
		}
		item.TxnState = COMMITTED
		err = m.conn.Put(v.Key, item)
		if err != nil {
			return err
		}
	}
	// clear the cache
	m.cache = make(map[string]MemoryItem)
	return nil
}

func (m *MemoryDatastore) Abort(hasCommitted bool) error {
	if !hasCommitted {
		m.cache = make(map[string]MemoryItem)
		return nil
	}

	for _, v := range m.cache {
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
	m.cache = make(map[string]MemoryItem)
	return nil
}

func (m *MemoryDatastore) Recover(key string) {
	//TODO implement me
	panic("implement me")
}

// overwrite the record with the application data and metadata that found in field Prev
func (m *MemoryDatastore) rollback(item MemoryItem) {
	var oldItem MemoryItem
	m.conn.Get(item.Key, &oldItem)
	var preItem MemoryItem
	err := json.Unmarshal([]byte(oldItem.Prev), &preItem)
	if err != nil {
		panic(err)
	}
	m.conn.Put(item.Key, preItem)
}

// make the record metadata with COMMITTED state
func (m *MemoryDatastore) rollForward(item MemoryItem) error {
	var oldItem MemoryItem
	m.conn.Get(item.Key, &oldItem)
	oldItem.TxnState = COMMITTED
	return m.conn.Put(item.Key, oldItem)
}

func (m *MemoryDatastore) GetName() string {
	return m.Name
}

func (m *MemoryDatastore) SetTxn(txn *Transaction) {
	m.Txn = txn
}

func (m *MemoryDatastore) WriteTSR(key string) error {
	return m.conn.Put(key, COMMITTED)
}

func (m *MemoryDatastore) DeleteTSR(key string) error {
	return m.conn.Delete(key)
}
