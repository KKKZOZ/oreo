package main

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimpleReadInCache(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	memoryPerson := Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(memoryPerson),
		TxnId:    "123123",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Put a item in cache
	cachePerson := Person{
		Name: "John",
		Age:  31,
	}
	cacheMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(cachePerson),
		TxnId:    "123123",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  2,
	}
	mds.writeCache[key] = cacheMemoryItem

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	// Check the result
	if result != cachePerson {
		t.Errorf("got %v want %v", result, cachePerson)
	}
}

func TestSimpleReadWhenCommitted(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

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
	conn.Put(key, expectedMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
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
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

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
	conn.Put(key, curMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
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
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)
	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

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
	conn.Put(key, curMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from memory datastore: %s", err)
	}
}

func TestSimpleReadWhenPreparedWithTSR(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a connection to the memory database
	conn := NewMemoryConnection("localhost", 8321)
	conn.Connect()

	// Create a new memory datastore
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "100",
		TxnState: PREPARED,
		TValid:   time.Now(),
		TLease:   time.Now(),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Write the TSR
	txn.globalDataStore.Write("100", COMMITTED)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenPrepareExpired(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a connection to the memory database
	conn := NewMemoryConnection("localhost", 8321)
	conn.Connect()

	// Create a new memory datastore
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "100",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	expectedStr := toJSONString(expectedMemoryItem)

	curPerson := Person{
		Name: "John",
		Age:  31,
	}

	curMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(curPerson),
		TxnId:    "101",
		TxnState: PREPARED,
		TValid:   time.Now().Add(-3 * time.Second),
		TLease:   time.Now().Add(-1 * time.Second),
		Version:  3,
		Prev:     expectedStr,
	}

	key := "John"
	conn.Put(key, curMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenPrepareNotExpired(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a connection to the memory database
	conn := NewMemoryConnection("localhost", 8321)
	conn.Connect()

	// Create a new memory datastore
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "100",
		TxnState: PREPARED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err.Error() != errors.New("dirty Read").Error() {
		t.Errorf("Error reading from memory datastore: %s", err)
	}
}

func TestSimpleWriteAndRead(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Write the value
	key := "John"
	person := Person{
		Name: "John",
		Age:  30,
	}
	err = txn.Write("memory", key, person)
	if err != nil {
		t.Errorf("Error writing to memory datastore: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleReadModifyWriteThenRead(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "123123",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	// Modify the value
	result.Age = 31

	// Write the value
	err = txn.Write("memory", key, result)
	if err != nil {
		t.Errorf("Error writing to memory datastore: %s", err)
	}

	// Read the value
	var result2 Person
	err = txn.Read("memory", key, &result2)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	// Check the result
	if result2 != result {
		t.Errorf("got %v want %v", result2, result)
	}
}

func TestSimpleOverwriteAndRead(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "123123",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Write the value
	person := Person{
		Name: "John",
		Age:  31,
	}
	err = txn.Write("memory", key, person)
	if err != nil {
		t.Errorf("Error writing to memory datastore: %s", err)
	}
	person.Age = 32
	err = txn.Write("memory", key, person)
	if err != nil {
		t.Errorf("Error writing to memory datastore: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleDeleteAndRead(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "123123",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Delete the value
	err = txn.Delete("memory", key)
	if err != nil {
		t.Errorf("Error deleting from memory datastore: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from memory datastore: %s", err)
	}
}

func TestSimpleDeleteTwice(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "123123",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Delete the value
	err = txn.Delete("memory", key)
	if err != nil {
		t.Errorf("Error deleting from memory datastore: %s", err)
	}
	err = txn.Delete("memory", key)
	if err.Error() != "key not found" {
		t.Errorf("Error deleting from memory datastore: %s", err)
	}
}

func TestDeleteWithRead(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	txn.Start()
	var person Person
	txn.Read("memory", "John", &person)
	err := txn.Delete("memory", "John")
	assert.NoError(t, err)

	err = txn.Commit()
	assert.NoError(t, err)
}

func TestDeleteWithoutRead(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	txn.Start()
	err := txn.Delete("memory", "John")
	if err != nil {
		t.Errorf("Error deleting from memory datastore: %s", err)
	}

	err = txn.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

}

func TestSimpleReadWriteDeleteThenRead(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "123123",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var person Person
	err = txn.Read("memory", key, &person)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	person.Age = 31

	// Write the value
	err = txn.Write("memory", key, person)
	if err != nil {
		t.Errorf("Error writing to memory datastore: %s", err)
	}

	// Delete the value
	err = txn.Delete("memory", key)
	if err != nil {
		t.Errorf("Error deleting from memory datastore: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from memory datastore: %s", err)
	}
}

func TestSimpleWriteDeleteWriteThenRead(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    toJSONString(expected),
		TxnId:    "123123",
		TxnState: COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Write the value
	person := Person{
		Name: "John",
		Age:  31,
	}
	err = txn.Write("memory", key, person)
	if err != nil {
		t.Errorf("Error writing to memory datastore: %s", err)
	}

	// Delete the value
	err = txn.Delete("memory", key)
	if err != nil {
		t.Errorf("Error deleting from memory datastore: %s", err)
	}

	// Write the value
	person.Age = 32
	err = txn.Write("memory", key, person)
	if err != nil {
		t.Errorf("Error writing to memory datastore: %s", err)
	}

	// Read the value
	var result Person
	err = txn.Read("memory", key, &result)
	if err != nil {
		t.Errorf("Error reading from memory datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}

}

func TestMockConnectionInTxn(t *testing.T) {
	t.Run("less than limit", func(t *testing.T) {
		// every record needs two `conn.Put()` call
		// Write TSR needs one `conn.Put()` call
		// So write X records needs 2X+1 `conn.Put()` call
		memoryDatabase := NewMemoryDatabase("localhost", 8321)
		go memoryDatabase.start()
		defer func() { <-memoryDatabase.msgChan }()
		defer func() { go memoryDatabase.stop() }()
		time.Sleep(100 * time.Millisecond)

		preTxn := NewTransactionWithSetup()
		preTxn.Start()
		for _, item := range inputItemList {
			preTxn.Write("memory", item.Value, item)
		}
		preTxn.Commit()

		txn := NewTransaction()
		conn := NewMockMemoryConnection("localhost", 8321, 11, true,
			func() error { return errors.New("debug error") })
		mds := NewMemoryDatastore("memory", conn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)

		txn.Start()
		for _, item := range inputItemList {
			var res TestItem
			txn.Read("memory", item.Value, &res)
			res.Value = item.Value + "-new"
			txn.Write("memory", item.Value, res)
		}

		err := txn.Commit()
		assert.NoError(t, err)

		// fmt.Printf("Call Times: %d\n", conn.callTimes)
	})

	t.Run("more than limit", func(t *testing.T) {
		// every record needs two `conn.Put()` call
		// Write TSR needs one `conn.Put()` call
		// So write X records needs 2X+1 `conn.Put()` call
		memoryDatabase := NewMemoryDatabase("localhost", 8321)
		go memoryDatabase.start()
		defer func() { <-memoryDatabase.msgChan }()
		defer func() { go memoryDatabase.stop() }()
		time.Sleep(100 * time.Millisecond)

		preTxn := NewTransactionWithSetup()
		preTxn.Start()
		for _, item := range inputItemList {
			preTxn.Write("memory", item.Value, item)
		}
		preTxn.Commit()

		txn := NewTransaction()
		conn := NewMockMemoryConnection("localhost", 8321, 3, true,
			func() error { return errors.New("debug error") })
		mds := NewMemoryDatastore("memory", conn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)

		txn.Start()
		for _, item := range inputItemList {
			var res TestItem
			txn.Read("memory", item.Value, &res)
			res.Value = item.Value + "-new"
			txn.Write("memory", item.Value, res)
		}

		err := txn.Commit()
		assert.EqualError(t, err, "prepare phase failed: write conflicted")

		// addtionally, we can check data consistency
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		for _, item := range inputItemList {
			var res TestItem
			postTxn.Read("memory", item.Value, &res)
			assert.Equal(t, item.Value, res.Value)
		}
		postTxn.Commit()
		fmt.Printf("Call Times: %d\n", conn.callTimes)

	})
}

// Error Test
// func TestSlowTransactionRecordExpiredConsistency(t *testing.T) {
// 	// run a memory database
// 	memoryDatabase := NewMemoryDatabase("localhost", 8321)
// 	go memoryDatabase.start()
// 	defer func() { <-memoryDatabase.msgChan }()
// 	defer func() { go memoryDatabase.stop() }()
// 	time.Sleep(100 * time.Millisecond)

// 	preTxn := NewTransactionWithSetup()
// 	preTxn.Start()
// 	for _, item := range inputItemList {
// 		preTxn.Write("memory", item.Value, item)
// 	}
// 	preTxn.Commit()

// 	go func() {
// 		slowTxn := NewTransaction()
// 		conn := NewMockMemoryConnection("localhost", 8321, 4, false,
// 			func() error { time.Sleep(3 * time.Second); return nil })
// 		mds := NewMemoryDatastore("memory", conn)
// 		slowTxn.AddDatastore(mds)
// 		slowTxn.SetGlobalDatastore(mds)

// 		slowTxn.Start()
// 		for _, item := range inputItemList {
// 			var result TestItem
// 			slowTxn.Read("memory", item.Value, &result)
// 			result.Value = item.Value + "-slow"
// 			slowTxn.Write("memory", item.Value, result)
// 		}
// 		err := slowTxn.Commit()
// 		assert.EqualError(t, err, "prepare phase failed: write conflicted")
// 	}()
// 	time.Sleep(2 * time.Second)

// 	testConn := NewMemoryConnection("localhost", 8321)
// 	testConn.Connect()
// 	var memItem1 MemoryItem
// 	testConn.Get("item1", &memItem1)
// 	assert.Equal(t, toJSONString(NewTestItem("item1-slow")), memItem1.Value)
// 	assert.Equal(t, memItem1.TxnState, PREPARED)

// 	var memItem2 MemoryItem
// 	testConn.Get("item2", &memItem2)
// 	assert.Equal(t, toJSONString(NewTestItem("item2-slow")), memItem2.Value)
// 	assert.Equal(t, memItem2.TxnState, PREPARED)

// 	var memItem3 MemoryItem
// 	testConn.Get("item3", &memItem3)
// 	assert.Equal(t, toJSONString(NewTestItem("item3")), memItem3.Value)
// 	assert.Equal(t, memItem3.TxnState, COMMITTED)

// 	fastTxn := NewTransactionWithSetup()
// 	fastTxn.Start()
// 	for i := 2; i <= 4; i++ {
// 		var result TestItem
// 		fastTxn.Read("memory", inputItemList[i].Value, &result)
// 		result.Value = inputItemList[i].Value + "-fast"
// 		fastTxn.Write("memory", inputItemList[i].Value, result)
// 	}
// 	err := fastTxn.Commit()
// 	assert.NoError(t, err)

// 	postTxn := NewTransactionWithSetup()
// 	postTxn.Start()

// 	var res1 TestItem
// 	postTxn.Read("memory", inputItemList[0].Value, &res1)
// 	assert.Equal(t, inputItemList[0], res1)

// 	var res2 TestItem
// 	postTxn.Read("memory", inputItemList[1].Value, &res2)
// 	assert.Equal(t, inputItemList[1], res2)

// 	for i := 2; i <= 4; i++ {
// 		var res TestItem
// 		postTxn.Read("memory", inputItemList[i].Value, &res)
// 		assert.Equal(t, inputItemList[i].Value+"-fast", res.Value)
// 	}

// 	err = postTxn.Commit()
// 	assert.NoError(t, err)

// }
