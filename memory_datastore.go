package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"slices"
	"sync"
	"time"
)

type MemoryDatastore struct {
	records    map[string]MemoryItem
	cache      map[string]MemoryItem
	versionMap map[string]int
	TStart     time.Time

	mu sync.Mutex
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

func NewMemoryDatastore() *MemoryDatastore {
	return &MemoryDatastore{
		records:    make(map[string]MemoryItem),
		cache:      make(map[string]MemoryItem),
		versionMap: make(map[string]int),
		mu:         sync.Mutex{},
	}
}

func (m *MemoryDatastore) start() {
	m.TStart = time.Now()
}

func (m *MemoryDatastore) read(key string) (string, error) {

	// if the record is in the cache
	if item, ok := m.cache[key]; ok {
		return item.Value, nil
	}

	if item, ok := m.records[key]; ok {
		if item.TxnState == COMMITTED {
			return m.readAsCommitted(item)
		}
		if item.TxnState == PREPARED {
			_, err := m.read(item.TxnId)
			if err == nil {
				// if TSR exists
				return m.readAsCommitted(item)
			}
			// if TSR does not exist
			if item.TLease.After(m.TStart) {
				err := m.rollForward(item)
				if err != nil {
					return "", err
				}
				return m.readAsCommitted(item)
			}
			return "", errors.New("key not found")
		}
		return "", errors.New("key not found")
	} else {
		return "", errors.New("key not found")
	}
}

func (m *MemoryDatastore) readAsCommitted(item MemoryItem) (string, error) {
	if item.TValid.Before(m.TStart) {
		m.versionMap[item.Key] = item.Version
		return item.Value, nil
	}
	var preItem MemoryItem
	err := json.Unmarshal([]byte(item.Prev), &preItem)
	if err != nil {
		// The transaction needs to be aborted
		return "", err
	}
	if preItem.TValid.Before(m.TStart) {
		m.versionMap[item.Key] = preItem.Version
		return preItem.Value, nil
	} else {
		return "", errors.New("key not found")
	}
}

func (m *MemoryDatastore) write(key string, value string) {
	// if the record is in the cache
	if item, ok := m.cache[key]; ok {
		item.Value = value
		m.cache[key] = item
		return
	}
	m.cache[key] = MemoryItem{
		Key:     key,
		Value:   value,
		TxnId:   txnManager.TxnId,
		TValid:  time.Now(),
		TLease:  time.Now().Add(leastTime * time.Millisecond),
		Version: m.versionMap[key],
	}
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
	m.mu.Lock()
	defer m.mu.Unlock()

	// TODO: 需不需要根据情况 roll forward ?
	if item.TxnState == PREPARED {
		return false
	}
	if m.records[item.Key].Version == item.Version {
		bytes, err := json.Marshal(m.records[item.Key])
		if err != nil {
			return false
		}
		item.Prev = string(bytes)
		m.records[item.Key] = item
		return true
	} else {
		return false
	}
}

func (m *MemoryDatastore) prepare(key string, record string) {
	records := make([]MemoryItem, 0, len(m.cache))
	for _, v := range m.cache {
		records = append(records, v)
	}
	// sort records by key
	slices.SortFunc(records, func(i, j MemoryItem) int {
		return cmp.Compare(i.Key, j.Key)
	})
	for i, v := range records {
		ok := m.conditionalUpdate(v)
		if !ok {
			// TODO: async rollback
		}
	}
}

func (m *MemoryDatastore) commit(key string, record string) {
	// write TSR first

}

func (m *MemoryDatastore) abort(key string, record string) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryDatastore) recover(key string) {
	//TODO implement me
	panic("implement me")
}

func (m *MemoryDatastore) rollForward(item MemoryItem) error {
	var preItem MemoryItem
	err := json.Unmarshal([]byte(item.Prev), &preItem)
	if err != nil {
		return errors.New("cannot unmarshal prev")
	}
	m.records[item.Key] = preItem
	return nil
}
