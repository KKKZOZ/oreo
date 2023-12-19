package main

import (
	"testing"
	"time"
)

func TestTwoTxnWriteThenRead(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn1 := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn1.AddDatastore(mds)
	txn1.SetGlobalDatastore(mds)

	expected := Person{
		Name: "John",
		Age:  30,
	}

	// Txn1 writes the record
	err := txn1.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	err=txn1.Write("memory", "John", expected)
	if err != nil {
		t.Errorf("Error writing record: %s", err)
	}
	err=txn1.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}


	txn2 := NewTransaction()
	txn2.AddDatastore(mds)
	txn2.SetGlobalDatastore(mds)

	// Txn2 reads the record
	err = txn2.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var actual Person
	err = txn2.Read("memory", "John", actual)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}
	err=txn2.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

	// compare
	if actual != expected {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
