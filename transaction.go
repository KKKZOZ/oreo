package main

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	TxnId           string
	TxnStartTime    time.Time
	TxnCommitTime   time.Time
	globalDataStore Datastore
	dataStoreMap    map[string]Datastore
}

func NewTransaction() *Transaction {
	return &Transaction{
		TxnId:        "",
		dataStoreMap: make(map[string]Datastore),
	}
}

func (t *Transaction) Start() error {
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
	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Read(key, value)
	}
	return errors.New("datastore not found")
}

func (t *Transaction) Write(dsName string, key string, value any) error {
	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Write(key, value)
	}
	return errors.New("datastore not found")
}

func (t *Transaction) Delete(dsName string, key string) error {
	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Delete(key)
	}
	return errors.New("datastore not found")
}

func (t *Transaction) Commit() error {
	// Prepare phase
	t.TxnCommitTime = time.Now()

	success := true
	for _, ds := range t.dataStoreMap {
		err := ds.Prepare()
		if err != nil {
			success = false
			break
		}
	}
	if !success {
		for _, ds := range t.dataStoreMap {
			ds.Abort()
		}
		return errors.New("prepare phase failed")
	}

	// Commit phase
	// The sync point
	t.globalDataStore.Write(t.TxnId, COMMITTED)

	for _, ds := range t.dataStoreMap {
		// TODO: do not allow abort after Commit
		// try indefinitely until success
		ds.Commit()
	}

	return nil
}
