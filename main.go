package main

import (
	"time"

	"github.com/kkkzoz/oreo/datastore/memory"
	"github.com/kkkzoz/oreo/txn"
)

type TimeTest struct {
	Time time.Time
	Name string
}

func NewTransactionWithSetup() *txn.Transaction {
	txn := txn.NewTransaction()
	conn := memory.NewMemoryConnection("localhost", 8321)
	mds := memory.NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)
	return txn
}

func main() {

}
