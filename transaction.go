package main

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	TxnId           string
	TxnState        State
	TxnStartTime    time.Time
	TxnCommitTime   time.Time
	globalDataStore Datastore
	dataStoreMap    map[string]Datastore
}

func NewTransaction() *Transaction {
	return &Transaction{
		TxnId:        "",
		TxnState:     EMPTY,
		dataStoreMap: make(map[string]Datastore),
	}
}

func (t *Transaction) Start() error {
	if t.TxnState != EMPTY {
		return errors.New("transaction is already started")
	}
	t.TxnState = STARTED

	// check nessary datastores are added
	if t.globalDataStore == nil {
		return errors.New("global datastore not set")
	}
	if len(t.dataStoreMap) == 0 {
		return errors.New("no datastores added")
	}
	t.TxnId = uuid.NewString()
	t.TxnStartTime = time.Now()
	for _, ds := range t.dataStoreMap {
		err := ds.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Transaction) AddDatastore(ds Datastore) error {
	// if name is duplicated
	if _, ok := t.dataStoreMap[ds.GetName()]; ok {
		return errors.New("duplicated datastore name")
	}
	ds.SetTxn(t)
	t.dataStoreMap[ds.GetName()] = ds
	return nil
}

func (t *Transaction) SetGlobalDatastore(ds Datastore) {
	t.globalDataStore = ds
}

func (t *Transaction) Read(dsName string, key string, value any) error {
	if t.TxnState != STARTED {
		return errors.New("transaction is not in STARTED state")
	}

	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Read(key, value)
	}
	return errors.New("datastore not found")
}

func (t *Transaction) Write(dsName string, key string, value any) error {
	if t.TxnState != STARTED {
		return errors.New("transaction is not in STARTED state")
	}

	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Write(key, value)
	}
	return errors.New("datastore not found")
}

func (t *Transaction) Delete(dsName string, key string) error {
	if t.TxnState != STARTED {
		return errors.New("transaction is not in STARTED state")
	}

	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Delete(key)
	}
	return errors.New("datastore not found")
}

func (t *Transaction) Commit() error {
	if t.TxnState == COMMITTED {
		return errors.New("transaction is already being committed")
	}
	if t.TxnState == ABORTED {
		return errors.New("transaction is already aborted")
	}
	if t.TxnState == EMPTY {
		return errors.New("transaction is not started")
	}
	t.TxnState = COMMITTED

	// Prepare phase
	t.TxnCommitTime = time.Now()

	success := true
	var cause error
	for _, ds := range t.dataStoreMap {
		err := ds.Prepare()
		if err != nil {
			success = false
			cause = err
			break
		}
	}
	if !success {
		for _, ds := range t.dataStoreMap {
			ds.Abort(true)
		}
		return errors.New("prepare phase failed: " + cause.Error())
	}

	// Commit phase
	// The sync point

	txnState, err := t.GetTSRState()
	if err == nil && txnState == ABORTED {
		t.Abort()
		return errors.New("transaction is aborted by other transaction")
	}

	err = t.WriteTSR(COMMITTED)
	if err != nil {
		return err
	}

	for _, ds := range t.dataStoreMap {
		// TODO: do not allow abort after Commit
		// try indefinitely until success
		ds.Commit()
	}

	t.DeleteTSR()

	return nil
}

func (t *Transaction) Abort() error {
	if t.TxnState == COMMITTED {
		return errors.New("transaction is already committed")
	}
	if t.TxnState == ABORTED {
		return errors.New("transaction is already aborted")
	}
	if t.TxnState == EMPTY {
		return errors.New("transaction is not started")
	}

	t.TxnState = ABORTED
	for _, ds := range t.dataStoreMap {
		ds.Abort(false)
	}
	return nil
}

func (t *Transaction) WriteTSR(txnState State) error {
	return t.globalDataStore.WriteTSR(t.TxnId, txnState)
}

func (t *Transaction) DeleteTSR() error {
	return t.globalDataStore.DeleteTSR(t.TxnId)
}

func (t *Transaction) GetTSRState() (State, error) {
	var state State
	err := t.globalDataStore.Read(t.TxnId, &state)
	return state, err
}
