package memory

import (
	"cmp"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

type MemoryDatastore struct {
	Name       string
	Txn        *txn.Transaction
	conn       MemoryConnectionInterface
	readCache  map[string]txn.DataItem2
	writeCache map[string]txn.DataItem2
}

func NewMemoryDatastore(name string, conn MemoryConnectionInterface) *MemoryDatastore {
	return &MemoryDatastore{
		Name:       name,
		conn:       conn,
		readCache:  make(map[string]txn.DataItem2),
		writeCache: make(map[string]txn.DataItem2),
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

	var item txn.DataItem2

	// if the record is in the writeCache
	if item, ok := m.writeCache[key]; ok {
		// if the record is marked as deleted
		if item.IsDeleted {
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
		groupKey, err := m.Txn.GetGroupKey(item.TxnId)
		if err == nil {
			// if TSR exists and the TSR is in COMMITTED state
			if groupKey.TxnState == config.COMMITTED {
				// roll forward the record
				item, err = m.rollForward(item)
				if err != nil {
					return err
				}
				return m.readAsCommitted(item, value)
			}
			if groupKey.TxnState == config.ABORTED {
				// if TSR exists and the TSR is in ABORTED state
				// we should rollback the record
				// because the transaction that modified the record has been aborted
				item, err := m.rollback(item)
				if err != nil {
					return err
				}
				return m.readAsCommitted(item, value)
			}
		}
		// if TSR does not exist
		// and if t_lease has expired
		// we should rollback the record
		// because the transaction that modified the record has been aborted
		// TODO: Not sure...
		if item.TLease.Before(time.Now()) {
			// the corresponding transaction is considered ABORTED
			m.Txn.CreateGroupKey(item.TxnId, config.ABORTED, 0)
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

func (m *MemoryDatastore) readAsCommitted(item txn.DataItem2, value any) error {

	curItem := item
	for i := 1; i <= config.Config.MaxRecordLength; i++ {

		if curItem.TValid < m.Txn.TxnStartTime {
			// if the record has been deleted
			if curItem.IsDeleted {
				return errors.New("key not found")
			}
			err := json.Unmarshal([]byte(curItem.Value), value)
			if err != nil {
				return err
			}
			m.readCache[curItem.Key] = curItem
			return nil
		}
		if i == config.Config.MaxRecordLength {
			break
		}
		// get the previous record
		var preItem txn.DataItem2
		err := json.Unmarshal([]byte(curItem.Prev), &preItem)
		if err != nil {
			// The transaction needs to be aborted
			return err
		}
		curItem = preItem
	}

	return errors.New("key not found")
}

func (m *MemoryDatastore) Write(key string, value any) error {
	jsonString := util.ToJSONString(value)
	// if the record is in the writeCache
	if item, ok := m.writeCache[key]; ok {
		item.Value, item.IsDeleted = jsonString, false
		m.writeCache[key] = item
		return nil
	}

	var version int
	// if the record is in the readCache
	if item, ok := m.readCache[key]; ok {
		version = item.Version
	} else {
		var oldItem txn.DataItem2
		err := m.conn.Get(key, &oldItem)
		if err != nil {
			version = 0
		} else {
			version = oldItem.Version
		}
	}
	// else Write a record to the cache
	m.writeCache[key] = txn.DataItem2{
		Key:       key,
		Value:     jsonString,
		TxnId:     m.Txn.TxnId,
		TValid:    time.Now().UnixMicro(),
		TLease:    time.Now().Add(config.Config.LeaseTime),
		Version:   version,
		IsDeleted: false,
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
		if item.IsDeleted {
			return errors.New("key not found")
		}
		item.IsDeleted = true
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
		// TODO: should we first read into the cache?
		var item txn.DataItem2
		err := m.conn.Get(key, &item)
		if err != nil {
			return err
		}
		version = item.Version
	}

	m.writeCache[key] = txn.DataItem2{
		Key:       key,
		IsDeleted: true,
		TxnId:     m.Txn.TxnId,
		TxnState:  config.COMMITTED,
		TValid:    time.Now().UnixMicro(),
		TLease:    time.Now(),
		Version:   version,
	}
	return nil
}

func (m *MemoryDatastore) conditionalUpdate(item txn.DataItem2) error {
	memItem := item

	// key := "memory" + memItem.Key
	// err := m.Txn.Lock(key, memItem.TxnId, 100*time.Millisecond)
	// if err != nil {
	// 	return err
	// }
	// defer m.Txn.Unlock(key, memItem.TxnId)

	var oldItem txn.DataItem2
	err := m.conn.Get(memItem.Key, &oldItem)
	if err != nil {
		// this is a new record
		newItem, err := m.updateMetadata(memItem, txn.DataItem2{})
		if err != nil {
			return err
		}
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
		newItem, err := m.updateMetadata(memItem, oldItem)
		if err != nil {
			return err
		}
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

func (m *MemoryDatastore) truncate(newItem *txn.DataItem2) (txn.DataItem2, error) {
	maxLen := config.Config.MaxRecordLength

	if newItem.LinkedLen > maxLen {
		stack := util.NewStack[txn.DataItem2]()
		stack.Push(*newItem)
		curItem := newItem
		for i := 1; i <= maxLen-1; i++ {
			var preItem txn.DataItem2
			err := json.Unmarshal([]byte(curItem.Prev), &preItem)
			if err != nil {
				return txn.DataItem2{}, errors.New("Unmarshal error: " + err.Error())
			}
			curItem = &preItem
			stack.Push(*curItem)
		}

		tarItem, err := stack.Pop()
		if err != nil {
			return txn.DataItem2{}, errors.New("Pop error: " + err.Error())
		}
		tarItem.Prev = ""
		tarItem.LinkedLen = 1

		for !stack.IsEmpty() {
			item, err := stack.Pop()
			if err != nil {
				return txn.DataItem2{}, errors.New("Pop error: " + err.Error())
			}
			item.Prev = util.ToJSONString(tarItem)
			item.LinkedLen = tarItem.LinkedLen + 1
			tarItem = item
		}
		return tarItem, nil
	} else {
		return *newItem, nil
	}
}

func (m *MemoryDatastore) updateMetadata(newItem txn.DataItem2, oldItem txn.DataItem2) (txn.DataItem2, error) {
	// update record's metadata
	if oldItem == (txn.DataItem2{}) {
		newItem.LinkedLen = 1
	} else {
		newItem.LinkedLen = oldItem.LinkedLen + 1
		newItem.Prev = util.ToJSONString(oldItem)
	}

	// truncate the record
	newItem, err := m.truncate(&newItem)
	if err != nil {
		return txn.DataItem2{}, err
	}

	newItem.Version++
	newItem.TxnState = config.PREPARED
	newItem.TValid = m.Txn.TxnCommitTime
	// TODO: time.Now() is temporary
	newItem.TLease = time.Now().Add(config.Config.LeaseTime)

	return newItem, nil
}

func (m *MemoryDatastore) Prepare() (int64, error) {
	records := make([]txn.DataItem2, 0, len(m.writeCache))
	for _, v := range m.writeCache {
		records = append(records, v)
	}
	// sort records by key
	// TODO: global consistent hash order
	slices.SortFunc(
		records, func(i, j txn.DataItem2) int {
			return cmp.Compare(i.Key, j.Key)
		},
	)
	for _, v := range records {
		err := m.conditionalUpdate(v)
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
}

// Commit updates the state of records in the data store to COMMITTED.
// It iterates over the write cache and updates each record's state to COMMITTED.
// After updating the records, it clears the write cache.
// Returns an error if there is any issue updating the records or clearing the cache.
func (m *MemoryDatastore) Commit() error {
	// update record's state to the COMMITTED state in the data store
	for _, v := range m.writeCache {
		var item txn.DataItem2
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
	m.writeCache = make(map[string]txn.DataItem2)
	m.readCache = make(map[string]txn.DataItem2)
	return nil
}

// Abort discards the changes made in the current transaction.
// If hasCommitted is false, it clears the write cache.
// If hasCommitted is true, it rolls back the changes made by the current transaction.
// It returns an error if there is any issue during the rollback process.
func (m *MemoryDatastore) Abort(hasCommitted bool) error {

	if !hasCommitted {
		m.writeCache = make(map[string]txn.DataItem2)
		return nil
	}

	for _, v := range m.writeCache {
		var item txn.DataItem2
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
	m.readCache = make(map[string]txn.DataItem2)
	m.writeCache = make(map[string]txn.DataItem2)
	return nil
}

func (m *MemoryDatastore) OnePhaseCommit() error {
	// TODO: implement me
	return nil
}

func (m *MemoryDatastore) Recover(key string) {
	//TODO implement me
	panic("implement me")
}

// rollback overwrites the record with the application data and metadata that found in field Prev
func (m *MemoryDatastore) rollback(item txn.DataItem2) (txn.DataItem2, error) {
	var newItem txn.DataItem2
	err := json.Unmarshal([]byte(item.Prev), &newItem)
	if err != nil {
		return txn.DataItem2{}, err
	}
	err = m.conn.Put(item.Key, newItem)
	if err != nil {
		return txn.DataItem2{}, err
	}

	return newItem, err
}

// rollForward makes the record metadata with COMMITTED state
func (m *MemoryDatastore) rollForward(item txn.DataItem2) (txn.DataItem2, error) {
	// var oldItem txn.DataItem
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

// Copy returns a new instance of MemoryDatastore that is a copy of the current datastore.
// The copy shares the same name and connection as the original datastore.
func (m *MemoryDatastore) Copy() txn.Datastorer {
	return NewMemoryDatastore(m.Name, m.conn)
}
