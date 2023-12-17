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
	dataStoreList   []Datastore
}

func NewTransaction() *Transaction {
	return &Transaction{
		TxnId:         "",
		dataStoreList: make([]Datastore, 0),
	}
}

func (t *Transaction) start() error {
	t.TxnId = uuid.NewString()
	t.TxnStartTime = time.Now()
	for _, ds := range t.dataStoreList {
		err := ds.start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Transaction) addDatastore(ds Datastore) {
	ds.setTxn(t)
	t.dataStoreList = append(t.dataStoreList, ds)
}

func (t *Transaction) setGlobalDatastore(ds Datastore) {
	t.globalDataStore = ds
}

func (t *Transaction) read(dsType DataStoreType, key string, value any) error {
	for _, ds := range t.dataStoreList {
		if ds.getType() == dsType {
			return ds.read(key, value)
		}
	}
	return errors.New("Datastore not found")
}

func (t *Transaction) write(dsType DataStoreType, key string, value any) {
	for _, ds := range t.dataStoreList {
		if ds.getType() == dsType {
			ds.write(key, value)
		}
	}
}

func (t *Transaction) delete(dsType DataStoreType, key string) {
	for _, ds := range t.dataStoreList {
		if ds.getType() == dsType {
			ds.delete(key)
		}
	}
}

func (t *Transaction) commit() error {
	// Prepare phase
	t.TxnCommitTime = time.Now()
	// for i, ds := range t.dataStoreList {
	// 	ds.prepare()

	// }
	return nil
}
