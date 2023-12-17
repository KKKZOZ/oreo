package main

import (
	"errors"
	"testing"
	"time"
)

func TestSimpleReadWhenCommitted(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore(conn)
	txn.addDatastore(mds)
	txn.setGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	expectedStr := toJSONString(expected)
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    expectedStr,
		TxnId:    "123123",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.put(key, expectedMemoryItem)

	// Start the transaction
	err := txn.start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.read(MEMORY, key, &result)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	// Check the result
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenCommittedFindPrevious(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8322)
	go memoryDatabase.start()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8322)
	mds := NewMemoryDatastore(conn)
	txn.addDatastore(mds)
	txn.setGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	curPerson := Person{
		Name: "John",
		Age:  31,
	}
	preMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "99",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  1,
	}
	curMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(curPerson),
		TxnId:    "100",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  2,
		Prev:     toJSONString(preMemoryItem),
	}

	key := "John"
	conn.put(key, curMemoryItem)

	// Start the transaction
	err := txn.start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.read(MEMORY, key, &result)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	// Check the result
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenCommittedFindNone(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8322)
	go memoryDatabase.start()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8322)
	mds := NewMemoryDatastore(conn)
	txn.addDatastore(mds)
	txn.setGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	curPerson := Person{
		Name: "John",
		Age:  31,
	}
	preMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "99",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  1,
	}
	curMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(curPerson),
		TxnId:    "100",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(20 * time.Second),
		TLease:   time.Now().Add(15 * time.Second),
		Version:  2,
		Prev:     toJSONString(preMemoryItem),
	}

	key := "John"
	conn.put(key, curMemoryItem)

	// Start the transaction
	err := txn.start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.read(MEMORY, key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from memory datastore: %s", err)
	}
}
