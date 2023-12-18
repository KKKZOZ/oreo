package main

import (
	"errors"
	"testing"
	"time"
)

func TestSimpleReadInCache(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
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
	mds.cache[key] = cacheMemoryItem

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

	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenPrepareNotExpired(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
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

func TestSimpleReadWriteDeleteThenRead(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
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
