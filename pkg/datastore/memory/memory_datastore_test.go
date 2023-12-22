package memory

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

func NewTransactionWithSetup() *txn.Transaction {
	txn := txn.NewTransaction()
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)
	return txn
}

func TestSimpleReadInCache(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	memoryPerson := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(memoryPerson),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Put a item in cache
	cachePerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	cacheMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(cachePerson),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
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
	var result testutil.Person
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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedStr := util.ToJSONString(expected)
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    expectedStr,
		TxnId:    "123123",
		TxnState: config.COMMITTED,
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
	var result testutil.Person
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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	preMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "99",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  1,
	}
	curMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(curPerson),
		TxnId:    "100",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  2,
		Prev:     util.ToJSONString(preMemoryItem),
	}

	key := "John"
	conn.Put(key, curMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)
	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	preMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "99",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  1,
	}
	curMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(curPerson),
		TxnId:    "100",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(20 * time.Second),
		TLease:   time.Now().Add(15 * time.Second),
		Version:  2,
		Prev:     util.ToJSONString(preMemoryItem),
	}

	key := "John"
	conn.Put(key, curMemoryItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("memory", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from memory datastore: %s", err)
	}
}

func TestSimpleReadWhenPreparedWithTSR(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a connection to the memory database
	conn := NewMemoryConnection("localhost", 8321)
	conn.Connect()

	// Create a new memory datastore
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "100",
		TxnState: config.PREPARED,
		TValid:   time.Now(),
		TLease:   time.Now(),
		Version:  2,
	}

	key := "John"
	conn.Put(key, expectedMemoryItem)

	// Write the TSR
	conn.Put("100", config.COMMITTED)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a connection to the memory database
	conn := NewMemoryConnection("localhost", 8321)
	conn.Connect()

	// Create a new memory datastore
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "100",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	expectedStr := util.ToJSONString(expectedMemoryItem)

	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}

	curMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(curPerson),
		TxnId:    "101",
		TxnState: config.PREPARED,
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
	var result testutil.Person
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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a connection to the memory database
	conn := NewMemoryConnection("localhost", 8321)
	conn.Connect()

	// Create a new memory datastore
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "100",
		TxnState: config.PREPARED,
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
	var result testutil.Person
	err = txn.Read("memory", key, &result)
	if err.Error() != errors.New("dirty Read").Error() {
		t.Errorf("Error reading from memory datastore: %s", err)
	}
}

func TestSimpleWriteAndRead(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

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
	person := testutil.Person{
		Name: "John",
		Age:  30,
	}
	err = txn.Write("memory", key, person)
	if err != nil {
		t.Errorf("Error writing to memory datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
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
	var result testutil.Person
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
	var result2 testutil.Person
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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
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
	person := testutil.Person{
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
	var result testutil.Person
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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
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
	var result testutil.Person
	err = txn.Read("memory", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from memory datastore: %s", err)
	}
}

func TestSimpleDeleteTwice(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	err := preTxn.Commit()
	assert.NoError(t, err)

	txn := NewTransactionWithSetup()
	txn.Start()
	var person testutil.Person
	err = txn.Read("memory", "John", &person)
	assert.NoError(t, err)
	err = txn.Delete("memory", "John")
	assert.NoError(t, err)

	err = txn.Commit()
	assert.NoError(t, err)
}

func TestDeleteWithoutRead(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := testutil.NewDefaultPerson()
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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
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
	var person testutil.Person
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
	var result testutil.Person
	err = txn.Read("memory", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from memory datastore: %s", err)
	}
}

func TestSimpleWriteDeleteWriteThenRead(t *testing.T) {
	// run a memory database
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new memory datastore
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the memory database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMemoryItem := MemoryItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
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
	person := testutil.Person{
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
	var result testutil.Person
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
		go memoryDatabase.Start()
		defer func() { <-memoryDatabase.MsgChan }()
		defer func() { go memoryDatabase.Stop() }()
		time.Sleep(100 * time.Millisecond)

		preTxn := NewTransactionWithSetup()
		preTxn.Start()
		for _, item := range testutil.InputItemList {
			preTxn.Write("memory", item.Value, item)
		}
		preTxn.Commit()

		txn := txn.NewTransaction()
		conn := NewMockMemoryConnection("localhost", 8321, 11, true,
			func() error { return errors.New("debug error") })
		mds := NewMemoryDatastore("memory", conn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)

		txn.Start()
		for _, item := range testutil.InputItemList {
			var res testutil.TestItem
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
		go memoryDatabase.Start()
		defer func() { <-memoryDatabase.MsgChan }()
		defer func() { go memoryDatabase.Stop() }()
		time.Sleep(100 * time.Millisecond)

		preTxn := NewTransactionWithSetup()
		preTxn.Start()
		for _, item := range testutil.InputItemList {
			preTxn.Write("memory", item.Value, item)
		}
		preTxn.Commit()

		txn := txn.NewTransaction()
		conn := NewMockMemoryConnection("localhost", 8321, 3, true,
			func() error { return errors.New("debug error") })
		mds := NewMemoryDatastore("memory", conn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)

		txn.Start()
		for _, item := range testutil.InputItemList {
			var res testutil.TestItem
			txn.Read("memory", item.Value, &res)
			res.Value = item.Value + "-new"
			txn.Write("memory", item.Value, res)
		}

		err := txn.Commit()
		assert.EqualError(t, err, "prepare phase failed: debug error")

		// addtionally, we can check data consistency
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		for _, item := range testutil.InputItemList {
			var res testutil.TestItem
			postTxn.Read("memory", item.Value, &res)
			assert.Equal(t, item.Value, res.Value)
		}
		postTxn.Commit()
		fmt.Printf("Call Times: %d\n", conn.callTimes)

	})
}

func TestMemoryDatastore_ConcurrentWriteConflicts(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	testutil.WaitForServer("localhost", 8321, 100*time.Millisecond)

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write("memory", item.Value, item)
	}
	err := preTxn.Commit()
	assert.NoError(t, err)

	resChan := make(chan bool)
	successId := 0

	concurrentCount := 100

	for i := 1; i <= concurrentCount; i++ {
		go func(id int) {
			txn := NewTransactionWithSetup()
			txn.Start()
			time.Sleep(100 * time.Millisecond)
			for _, item := range testutil.InputItemList {
				var res testutil.TestItem
				txn.Read("memory", item.Value, &res)
				res.Value = item.Value + "-new-" + strconv.Itoa(id)
				txn.Write("memory", item.Value, res)
			}

			err := txn.Commit()
			if err != nil {
				if err.Error() != "prepare phase failed: write conflicted: the record is in PREPARED state" &&
					err.Error() != "prepare phase failed: write conflicted: the record has been modified by others" {
					t.Errorf("Unexpected error: %s", err)
				}
				resChan <- false
			} else {
				resChan <- true
				successId = id
			}
		}(i)
	}
	commitCount := 0

	for i := 1; i <= concurrentCount; i++ {
		res := <-resChan
		if res {
			commitCount++
		}
	}

	assert.Equal(t, 1, commitCount)

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	for _, item := range testutil.InputItemList {
		var res testutil.TestItem
		postTxn.Read("memory", item.Value, &res)
		assert.Equal(t, item.Value+"-new-"+strconv.Itoa(successId), res.Value)
	}
	err = postTxn.Commit()
	assert.NoError(t, err)

}

func TestTxnWriteMultiRecord(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	err := testutil.WaitForServer("localhost", 8321, 100*time.Millisecond)
	assert.Nil(t, err)

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	preTxn.Write("memory", "item1", testutil.NewTestItem("item1"))
	preTxn.Write("memory", "item2", testutil.NewTestItem("item2"))
	err = preTxn.Commit()
	assert.Nil(t, err)

	txn := NewTransactionWithSetup()
	txn.Start()
	var item testutil.TestItem
	txn.Read("memory", "item1", &item)
	item.Value = "item1_new"
	txn.Write("memory", "item1", item)

	txn.Read("memory", "item2", &item)
	item.Value = "item2_new"
	txn.Write("memory", "item2", item)

	err = txn.Commit()

	assert.Nil(t, err)

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	var resItem testutil.TestItem
	postTxn.Read("memory", "item1", &resItem)
	assert.Equal(t, "item1_new", resItem.Value)
	postTxn.Read("memory", "item2", &resItem)
	assert.Equal(t, "item2_new", resItem.Value)

}

// Error Test
// func TestSlowTransactionRecordExpiredConsistency(t *testing.T) {
// 	// run a memory database
// 	memoryDatabase := NewMemoryDatabase("localhost", 8321)
// 	go memoryDatabase.Start()
// 	defer func() { <-memoryDatabase.MsgChan }()
// 	defer func() { go memoryDatabase.Stop() }()
// 	time.Sleep(100 * time.Millisecond)

// 	preTxn := NewTransactionWithSetup()
// 	preTxn.Start()
// 	for _, item := range testutil.InputItemList {
// 		preTxn.Write("memory", item.Value, item)
// 	}
// 	preTxn.Commit()

// 	go func() {
// 		slowTxn := txn.NewTransaction()
// 		conn := NewMockMemoryConnection("localhost", 8321, 4, false,
// 			func() error { time.Sleep(3 * time.Second); return nil })
// 		mds := NewMemoryDatastore("memory", conn)
// 		slowTxn.AddDatastore(mds)
// 		slowTxn.SetGlobalDatastore(mds)

// 		slowTxn.Start()
// 		for _, item := range testutil.InputItemList {
// 			var result testutil.TestItem
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
// 	assert.Equal(t, util.ToJSONString(NewTestItem("item1-slow")), memItem1.Value)
// 	assert.Equal(t, memItem1.TxnState, PREPARED)

// 	var memItem2 MemoryItem
// 	testConn.Get("item2", &memItem2)
// 	assert.Equal(t, util.ToJSONString(NewTestItem("item2-slow")), memItem2.Value)
// 	assert.Equal(t, memItem2.TxnState, PREPARED)

// 	var memItem3 MemoryItem
// 	testConn.Get("item3", &memItem3)
// 	assert.Equal(t, util.ToJSONString(NewTestItem("item3")), memItem3.Value)
// 	assert.Equal(t, memItem3.TxnState, COMMITTED)

// 	fastTxn := NewTransactionWithSetup()
// 	fastTxn.Start()
// 	for i := 2; i <= 4; i++ {
// 		var result testutil.TestItem
// 		fastTxn.Read("memory", testutil.InputItemList[i].Value, &result)
// 		result.Value = testutil.InputItemList[i].Value + "-fast"
// 		fastTxn.Write("memory", testutil.InputItemList[i].Value, result)
// 	}
// 	err := fastTxn.Commit()
// 	assert.NoError(t, err)

// 	postTxn := NewTransactionWithSetup()
// 	postTxn.Start()

// 	var res1 testutil.TestItem
// 	postTxn.Read("memory", testutil.InputItemList[0].Value, &res1)
// 	assert.Equal(t, testutil.InputItemList[0], res1)

// 	var res2 testutil.TestItem
// 	postTxn.Read("memory", testutil.InputItemList[1].Value, &res2)
// 	assert.Equal(t, testutil.InputItemList[1], res2)

// 	for i := 2; i <= 4; i++ {
// 		var res testutil.TestItem
// 		postTxn.Read("memory", testutil.InputItemList[i].Value, &res)
// 		assert.Equal(t, testutil.InputItemList[i].Value+"-fast", res.Value)
// 	}

// 	err = postTxn.Commit()
// 	assert.NoError(t, err)

// }
