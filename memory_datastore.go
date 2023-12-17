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
	TStart     time.Time
}

type MemoryItem struct {
	Key      string
	Value    string
	TxnId    string
	TxnState State
	TValid   time.Time
	TLease   time.Time
	Prev     string
	DoDelete bool
	Version  int
}

func NewMemoryDatastore(conn *MemoryConnection) *MemoryDatastore {
	return &MemoryDatastore{
		dataStore:  dataStore{Type: MEMORY},
		conn:       conn,
		cache:      make(map[string]MemoryItem),
		versionMap: make(map[string]int),
	}
}

func (m *MemoryDatastore) start() error {
	err := m.conn.connect()
	if err != nil {
		return err
	}
	m.TStart = time.Now()
	return nil
}

func (m *MemoryDatastore) read(key string, value any) error {

	var item MemoryItem

	// if the record is in the cache
	if item, ok := m.cache[key]; ok {
		err := json.Unmarshal([]byte(item.Value), value)
		return err
	}

	// else get if from data
	err := m.conn.get(key, &item)
	if err != nil {
		return err
	}

	if item.TxnState == COMMITTED {
		return m.readAsCommitted(item, value)
	}
	if item.TxnState == PREPARED {
		var state State
		err := m.Txn.globalDataStore.read(key, &state)
		if err == nil {
			// if TSR exists
			return m.readAsCommitted(item, value)
		}
		// if TSR does not exist
		if item.TLease.After(m.TStart) {
			err := m.rollForward(item)
			if err != nil {
				return err
			}
			return m.readAsCommitted(item, value)
		}
		return errors.New("key not found")
	}
	return errors.New("key not found")

}

func (m *MemoryDatastore) readAsCommitted(item MemoryItem, value any) error {
	if item.TValid.Before(m.TStart) {
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
	if preItem.TValid.Before(m.TStart) {
		m.versionMap[item.Key] = preItem.Version
		err := json.Unmarshal([]byte(preItem.Value), value)
		return err
	} else {
		return errors.New("key not found")
	}
}

func (m *MemoryDatastore) write(key string, value any) error {
	jsonString := toJSONString(value)
	// if the record is in the cache
	if item, ok := m.cache[key]; ok {
		item.Value = jsonString
		m.cache[key] = item
		return nil
	}
	m.cache[key] = MemoryItem{
		Key:     key,
		Value:   jsonString,
		TxnId:   m.Txn.TxnId,
		TValid:  time.Now(),
		TLease:  time.Now().Add(leastTime * time.Millisecond),
		Version: m.versionMap[key],
	}
	return nil
}

func (m *MemoryDatastore) prev(key string, record string) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryDatastore) delete(key string) {
	// if the record is in the cache
	if item, ok := m.cache[key]; ok {
		item.DoDelete = true
		m.cache[key] = item
		return
	}
}

func (m *MemoryDatastore) conditionalUpdate(item MemoryItem) bool {

	// TODO: 需不需要根据情况 roll forward ?
	if item.TxnState == PREPARED {
		return false
	}

	var oldItem MemoryItem
	err := m.conn.get(item.Key, &oldItem)
	if err != nil {
		return false
	}
	if oldItem.Version == item.Version {
		if item.DoDelete {
			m.conn.delete(item.Key)
			return true
		}
		item.Prev = toJSONString(oldItem)
		item.Version++
		m.conn.put(item.Key, item)
		return true
	} else {
		return false
	}
}

func (m *MemoryDatastore) prepare() error {
	records := make([]MemoryItem, 0, len(m.cache))
	for _, v := range m.cache {
		records = append(records, v)
	}
	// sort records by key
	// TODO: global consistent hash order
	slices.SortFunc(records, func(i, j MemoryItem) int {
		return cmp.Compare(i.Key, j.Key)
	})
	for i, v := range records {
		ok := m.conditionalUpdate(v)
		if !ok {
			// rollback itself, then report it to the Transaction
			// 0 ~ i-1 should be rollbacked
			for j := 0; j < i; j++ {
				itemKey := records[j].Key
				var oldItem MemoryItem
				err := m.conn.get(itemKey, &oldItem)
				if err != nil {
					return err
				}
				var item MemoryItem
				err = json.Unmarshal([]byte(oldItem.Prev), &item)
				if err != nil {
					return err
				}
				m.conn.put(itemKey, item)
			}
			return errors.New("prepare failed")
		}
	}
	return nil
}

func (m *MemoryDatastore) commit() error {
	// update record's state to the COMMITTED state in the data store
	for _, v := range m.cache {
		var item MemoryItem
		err := m.conn.get(v.Key, &item)
		if err != nil {
			return err
		}
		item.TxnState = COMMITTED
		err = m.conn.put(v.Key, item)
		if err != nil {
			return err
		}
	}
	// clear the cache
	m.cache = make(map[string]MemoryItem)
	return nil
}

func (m *MemoryDatastore) abort() {
	for _, v := range m.cache {
		var item MemoryItem
		err := m.conn.get(v.Key, &item)
		if err != nil {
			panic(err)
		}
		if item.TxnState == PREPARED {
			m.rollback(item)
		}
	}
}

func (m *MemoryDatastore) recover(key string) {
	//TODO implement me
	panic("implement me")
}

// overwrite the record with the application data and metadata that found in field Prev
func (m *MemoryDatastore) rollback(item MemoryItem) {
	var oldItem MemoryItem
	m.conn.get(item.Key, &oldItem)
	var preItem MemoryItem
	err := json.Unmarshal([]byte(oldItem.Prev), &preItem)
	if err != nil {
		panic(err)
	}
	m.conn.put(item.Key, preItem)
}

// make the record metadata with COMMITTED state
func (m *MemoryDatastore) rollForward(item MemoryItem) error {
	var oldItem MemoryItem
	m.conn.get(item.Key, &oldItem)
	oldItem.TxnState = COMMITTED
	return m.conn.put(item.Key, oldItem)
}

func (m *MemoryDatastore) getType() DataStoreType {
	return m.Type
}

func (m *MemoryDatastore) setTxn(txn *Transaction) {
	m.Txn = txn
}
