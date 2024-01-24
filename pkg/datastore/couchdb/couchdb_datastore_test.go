package couchdb

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	trxn "github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

// NewDefaultCouchDBConnection creates a new instance of CouchDBConnection with default settings.
// It establishes a connection to the CouchDBDB server running on localhost:27017.
// Returns a pointer to the CouchDBConnection.
func NewDefaultConnection() *CouchDBConnection {
	conn := NewCouchDBConnection(&ConnectionOptions{
		Address: "http://admin:password@localhost:5984",
		DBName:  "oreo",
	})
	err := conn.Connect()
	if err != nil {
		panic(err)
	}
	return conn
}

// NewTransactionWithSetup creates a new transaction with a Redis datastore setup.
// It initializes a Redis connection, creates a new transaction, and adds a Redis datastore to the transaction.
// The Redis connection is established with the provided address.
// The created transaction is returned.
func NewTransactionWithSetup() *trxn.Transaction {
	conn := NewCouchDBConnection(&ConnectionOptions{
		Address: "http://admin:password@localhost:5984",
		DBName:  "oreo",
	})
	txn := trxn.NewTransaction()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)
	return txn
}

func TestSimpleReadWhenCommitted(t *testing.T) {

	txn := trxn.NewTransaction()
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	conn.Delete("John")

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "123123",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-5 * time.Second),
		// CVersion:  "2",
	}

	key := "John"
	conn.PutItem(key, expectedRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err != nil {
		t.Errorf("Error reading from couchdb datastore: %s", err)
	}

	// Check the result
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenCommittedFindEmpty(t *testing.T) {

	txn1 := trxn.NewTransaction()
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn1.AddDatastore(rds)
	txn1.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "TestSimpleReadWhenCommittedFindEmpty",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(+10 * time.Second),
		CTLease:   time.Now().Add(+5 * time.Second),
		CVersion:  "2",
	}

	key := "John"
	conn.Delete(key)
	conn.PutItem(key, expectedRedisItem)

	// Start the transaction
	err := txn1.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn1.Read("couchdb", key, &result)
	assert.EqualError(t, err, trxn.KeyNotFound.Error())

}

func TestSimpleReadWhenCommittedFindPrevious(t *testing.T) {
	// Create a new transaction
	txn := trxn.NewTransaction()

	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	preRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "99",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-5 * time.Second),
		// CVersion:  "1",
	}
	curRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(curPerson),
		CTxnId:    "100",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(10 * time.Second),
		CTLease:   time.Now().Add(5 * time.Second),
		// CVersion:  "2",
		CPrev: util.ToJSONString(preRedisItem),
	}

	key := "John"
	conn.Delete(key)

	conn.PutItem(key, curRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenCommittedFindNone(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	preRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "99",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(10 * time.Second),
		CTLease:   time.Now().Add(5 * time.Second),
		CVersion:  "1",
	}
	curRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(curPerson),
		CTxnId:    "100",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(20 * time.Second),
		CTLease:   time.Now().Add(15 * time.Second),
		CVersion:  "2",
		CPrev:     util.ToJSONString(preRedisItem),
	}

	key := "John"
	conn.Delete(key)
	conn.PutItem(key, curRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from redis datastore: %s", err)
	}
}

// TestSimpleReadWhenPreparedWithTSRInCOMMITTED tests the scenario where a simple read operation is performed
// on a record which is in PREPARED state and has a TSR in COMMITTED state.
func TestSimpleReadWhenPreparedWithTSRInCOMMITTED(t *testing.T) {
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "100",
		CTxnState: config.PREPARED,
		CTValid:   time.Now(),
		CTLease:   time.Now(),
		// CVersion:  "2",
	}

	key := "John"
	conn.Delete(key)
	conn.PutItem(key, expectedRedisItem)

	conn.Delete("100")
	// Write the TSR
	err := conn.Put("100", config.COMMITTED)
	assert.NoError(t, err)

	// Start the transaction
	err = txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}

	// Delete the TSR
	conn.Delete("100")
}

// TestSimpleReadWhenPreparedWithTSRInABORTED tests the scenario where a simple read operation is performed
// on a record which is in PREPARED state and has a TSR in ABORTED state.
func TestSimpleReadWhenPreparedWithTSRInABORTED(t *testing.T) {
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	tarMemItem := &CouchDBItem{
		CKey:      "item1",
		CValue:    util.ToJSONString(testutil.NewTestItem("item1")),
		CTxnId:    "99",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-9 * time.Second),
		CVersion:  "1",
	}

	curMemItem := &CouchDBItem{
		CKey:      "item1",
		CValue:    util.ToJSONString(testutil.NewTestItem("item1-prepared")),
		CTxnId:    "TestSimpleReadWhenPreparedWithTSRInABORTED",
		CTxnState: config.PREPARED,
		CTValid:   time.Now().Add(-5 * time.Second),
		CTLease:   time.Now().Add(-4 * time.Second),
		CPrev:     util.ToJSONString(tarMemItem),
		// CVersion:  "2",
	}

	key := "item1"
	conn.Delete(key)
	conn.PutItem(key, curMemItem)

	// Write the TSR
	conn.Put("TestSimpleReadWhenPreparedWithTSRInABORTED", config.ABORTED)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.TestItem
	err = txn.Read("couchdb", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}
	expected := testutil.NewTestItem("item1")
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}

	// Delete the TSR
	conn.Delete("TestSimpleReadWhenPreparedWithTSRInABORTED")
}

func TestSimpleReadWhenPrepareExpired(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "100",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-5 * time.Second),
		CVersion:  "2",
	}

	expectedStr := util.ToJSONString(expectedRedisItem)

	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}

	curRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(curPerson),
		CTxnId:    "101",
		CTxnState: config.PREPARED,
		CTValid:   time.Now().Add(-3 * time.Second),
		CTLease:   time.Now().Add(-1 * time.Second),
		// CVersion:  "3",
		CPrev: expectedStr,
	}

	key := "John"
	conn.Delete(key)
	conn.PutItem(key, curRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenPrepareNotExpired(t *testing.T) {

	dbItem1 := &CouchDBItem{
		CKey:       "item1",
		CValue:     util.ToJSONString(testutil.NewTestItem("item1-pre1")),
		CTxnId:     "TestSimpleReadWhenPrepareNotExpired1",
		CTxnState:  config.COMMITTED,
		CTValid:    time.Now().Add(-2 * time.Second),
		CTLease:    time.Now().Add(-1 * time.Second),
		CLinkedLen: 1,
		CVersion:   "1",
	}

	dbItem2 := &CouchDBItem{
		CKey:       "item1",
		CValue:     util.ToJSONString(testutil.NewTestItem("item1-pre2")),
		CTxnId:     "TestSimpleReadWhenPrepareNotExpired2",
		CTxnState:  config.PREPARED,
		CTValid:    time.Now().Add(1 * time.Second),
		CTLease:    time.Now().Add(2 * time.Second),
		CPrev:      util.ToJSONString(dbItem1),
		CLinkedLen: 2,
		// CVersion:   "2",
	}

	t.Run("when the item has a valid Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()
		conn.Delete("item1")
		conn.PutItem("item1", dbItem2)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var item testutil.TestItem
		err := txn1.Read("couchdb", "item1", &item)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre1", item.Value)
	})

	t.Run("when the item has an empty Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()
		dbItem := dbItem2
		dbItem.SetPrev("")
		dbItem.SetLinkedLen(1)
		conn.Delete("item1")
		conn.PutItem("item1", dbItem)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var item testutil.TestItem
		err := txn1.Read("couchdb", "item1", &item)
		assert.EqualError(t, err, trxn.KeyNotFound.Error())
	})
}

func TestSimpleReadWhenDeleted(t *testing.T) {
	conn := NewDefaultConnection()
	dbItem := &CouchDBItem{
		CKey:       "item2",
		CValue:     util.ToJSONString(testutil.NewTestItem("item2-db")),
		CTxnState:  config.COMMITTED,
		CTValid:    time.Now().Add(-2 * time.Second),
		CTLease:    time.Now().Add(-1 * time.Second),
		CLinkedLen: 1,
		CVersion:   "1",
		CIsDeleted: true,
	}

	conn.Delete("item2")
	conn.PutItem(dbItem.Key(), dbItem)

	txn1 := NewTransactionWithSetup()
	txn1.Start()

	var item testutil.TestItem
	err := txn1.Read("couchdb", "item2", &item)
	assert.EqualError(t, err, trxn.KeyNotFound.Error())
}

func TestSimpleWriteAndRead(t *testing.T) {

	// Start the transaction
	txn := NewTransactionWithSetup()
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
	err = txn.Write("couchdb", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleDirectWrite(t *testing.T) {

	conn := NewDefaultConnection()
	conn.Delete("John")

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	key := "John"
	prePerson := testutil.NewPerson("John-pre")
	preTxn.Write("couchdb", key, prePerson)
	err := preTxn.Commit()
	assert.NoError(t, err)

	// Start the transaction
	txn := NewTransactionWithSetup()
	err = txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Write the value
	person := testutil.Person{
		Name: "John",
		Age:  30,
	}
	err = txn.Write("couchdb", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}
	err = txn.Commit()
	assert.NoError(t, err)
}

func TestSimpleWriteAndReadLocal(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

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
	err = txn.Write("couchdb", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleReadModifyWriteThenRead(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "123123",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-5 * time.Second),
		CVersion:  "2",
	}

	key := "John"
	conn.PutItem(key, expectedRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Modify the value
	result.Age = 31

	// Write the value
	err = txn.Write("couchdb", key, result)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result2 testutil.Person
	err = txn.Read("couchdb", key, &result2)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result2 != result {
		t.Errorf("got %v want %v", result2, result)
	}
}

func TestSimpleOverwriteAndRead(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "123123",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-5 * time.Second),
		CVersion:  "2",
	}

	key := "John"
	conn.PutItem(key, expectedRedisItem)

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
	err = txn.Write("couchdb", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}
	person.Age = 32
	err = txn.Write("couchdb", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleDeleteAndRead(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "123123",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-5 * time.Second),
		CVersion:  "2",
	}

	key := "John"
	conn.PutItem(key, expectedRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Delete the value
	err = txn.Delete("couchdb", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from redis datastore: %s", err)
	}
}

func TestSimpleDeleteTwice(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "123123",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-5 * time.Second),
		// CVersion:  "2",
	}

	key := "John"
	conn.Delete(key)
	_, err := conn.PutItem(key, expectedRedisItem)
	if err != nil {
		t.Errorf("Error putting item to redis datastore: %s", err)
	}

	// Start the transaction
	err = txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Delete the value
	err = txn.Delete("couchdb", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}
	err = txn.Delete("couchdb", key)
	if err.Error() != "key not found" {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}
}

func TestDeleteWithRead(t *testing.T) {

	conn := NewDefaultConnection()
	// clear the test data
	conn.Delete("John")

	preTxn := NewTransactionWithSetup()
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("couchdb", "John", dataPerson)
	err := preTxn.Commit()
	assert.NoError(t, err)

	txn := NewTransactionWithSetup()
	txn.Start()
	var person testutil.Person
	err = txn.Read("couchdb", "John", &person)
	assert.NoError(t, err)
	err = txn.Delete("couchdb", "John")
	assert.NoError(t, err)

	err = txn.Commit()
	assert.NoError(t, err)
}

func TestDeleteWithoutRead(t *testing.T) {

	preTxn := NewTransactionWithSetup()
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("couchdb", "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	txn.Start()
	err := txn.Delete("couchdb", "John")
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	err = txn.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

}

func TestSimpleReadWriteDeleteThenRead(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "123123",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-5 * time.Second),
		// CVersion:  "2",
	}

	key := "John"
	conn.Delete(key)
	conn.PutItem(key, expectedRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var person testutil.Person
	err = txn.Read("couchdb", key, &person)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	person.Age = 31

	// Write the value
	err = txn.Write("couchdb", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Delete the value
	err = txn.Delete("couchdb", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from redis datastore: %s", err)
	}
}

func TestSimpleWriteDeleteWriteThenRead(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewCouchDBDatastore("couchdb", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &CouchDBItem{
		CKey:      "John",
		CValue:    util.ToJSONString(expected),
		CTxnId:    "123123",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-5 * time.Second),
		CVersion:  "2",
	}

	key := "John"
	conn.PutItem(key, expectedRedisItem)

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
	err = txn.Write("couchdb", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Delete the value
	err = txn.Delete("couchdb", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	// Write the value
	person.Age = 32
	err = txn.Write("couchdb", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("couchdb", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}

}

func TestRedisDatastore_ConcurrentWriteConflicts(t *testing.T) {

	// clear the test data
	conn := NewDefaultConnection()
	for _, item := range testutil.InputItemList {
		conn.Delete(item.Value)
	}

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write("couchdb", item.Value, item)
	}
	err := preTxn.Commit()
	assert.NoError(t, err)

	resChan := make(chan bool)
	successId := 0

	concurrentCount := 10

	for i := 1; i <= concurrentCount; i++ {
		go func(id int) {
			txn := NewTransactionWithSetup()
			txn.Start()
			for _, item := range testutil.InputItemList {
				var res testutil.TestItem
				txn.Read("couchdb", item.Value, &res)
				res.Value = item.Value + "-new-" + strconv.Itoa(id)
				txn.Write("couchdb", item.Value, res)
			}

			time.Sleep(100 * time.Millisecond)
			err := txn.Commit()
			if err != nil {
				if err.Error() != "prepare phase failed: write conflicted: the record is in PREPARED state" &&
					err.Error() != "prepare phase failed: write conflicted: the record has been modified by others" &&
					err.Error() != "prepare phase failed: version mismatch" {
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
		postTxn.Read("couchdb", item.Value, &res)
		assert.Equal(t, item.Value+"-new-"+strconv.Itoa(successId), res.Value)
	}
	err = postTxn.Commit()
	assert.NoError(t, err)

}

func TestTxnWriteMultiRecord(t *testing.T) {

	// clear the test data
	conn := NewDefaultConnection()
	conn.Delete("item1")
	conn.Delete("item2")

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	preTxn.Write("couchdb", "item1", testutil.NewTestItem("item1"))
	preTxn.Write("couchdb", "item2", testutil.NewTestItem("item2"))
	err := preTxn.Commit()
	assert.Nil(t, err)

	txn := NewTransactionWithSetup()
	txn.Start()
	var item testutil.TestItem
	txn.Read("couchdb", "item1", &item)
	item.Value = "item1_new"
	txn.Write("couchdb", "item1", item)

	txn.Read("couchdb", "item2", &item)
	item.Value = "item2_new"
	txn.Write("couchdb", "item2", item)

	err = txn.Commit()
	assert.Nil(t, err)

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	var resItem testutil.TestItem
	postTxn.Read("couchdb", "item1", &resItem)
	assert.Equal(t, "item1_new", resItem.Value)
	postTxn.Read("couchdb", "item2", &resItem)
	assert.Equal(t, "item2_new", resItem.Value)

}

// ---|---------|--------|---------|------> time
// item1_1  T_Start   item1_2   item1_3
func TestLinkedReadAsCommitted(t *testing.T) {

	item1_1 := testutil.NewTestItem("item1_1")
	memItem1_1 := &CouchDBItem{
		CKey:       "item1",
		CValue:     util.ToJSONString(item1_1),
		CTxnId:     "txn1",
		CTxnState:  config.COMMITTED,
		CTValid:    time.Now().Add(-10 * time.Second),
		CTLease:    time.Now().Add(-9 * time.Second),
		CVersion:   "1",
		CLinkedLen: 1,
	}

	item1_2 := testutil.NewTestItem("item1_2")
	memItem1_2 := &CouchDBItem{
		CKey:       "item1",
		CValue:     util.ToJSONString(item1_2),
		CTxnId:     "txn2",
		CTxnState:  config.COMMITTED,
		CTValid:    time.Now().Add(5 * time.Second),
		CTLease:    time.Now().Add(6 * time.Second),
		CVersion:   "2",
		CPrev:      util.ToJSONString(memItem1_1),
		CLinkedLen: 2,
	}

	item1_3 := testutil.NewTestItem("item1_3")
	memItem1_3 := &CouchDBItem{
		CKey:      "item1",
		CValue:    util.ToJSONString(item1_3),
		CTxnId:    "txn3",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(10 * time.Second),
		CTLease:   time.Now().Add(11 * time.Second),
		// CVersion:   "3",
		CPrev:      util.ToJSONString(memItem1_2),
		CLinkedLen: 3,
	}

	t.Run("read will fail due to MaxRecordLength=2", func(t *testing.T) {

		conn := NewDefaultConnection()
		conn.Delete("item1")
		_, err := conn.PutItem("item1", memItem1_3)
		assert.NoError(t, err)

		config.Config.MaxRecordLength = 2
		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err = txn.Read("couchdb", "item1", &item)
		assert.EqualError(t, err, "key not found")
	})

	t.Run("read will success due to MaxRecordLength=3", func(t *testing.T) {

		conn := NewDefaultConnection()
		conn.Delete("item1")
		conn.PutItem("item1", memItem1_3)

		config.Config.MaxRecordLength = 3
		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("couchdb", "item1", &item)
		assert.Nil(t, err)

		assert.Equal(t, "item1_1", item.Value)
	})

	t.Run("read will success due to MaxRecordLength > 3", func(t *testing.T) {

		conn := NewDefaultConnection()
		conn.Delete("item1")
		conn.PutItem("item1", memItem1_3)

		config.Config.MaxRecordLength = 3 + 1
		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("couchdb", "item1", &item)
		assert.Nil(t, err)

		assert.Equal(t, "item1_1", item.Value)
	})
}

func TestLinkedTruncate(t *testing.T) {

	t.Cleanup(func() {
		config.Config.MaxRecordLength = 2
	})

	t.Run("4 commits immediately after txn.Start() when MaxRecordLength = 2", func(t *testing.T) {

		config.Config.MaxRecordLength = 2

		conn := NewDefaultConnection()
		conn.Delete("item1")

		for i := 1; i <= 4; i++ {
			time.Sleep(10 * time.Millisecond)
			item := testutil.NewTestItem("item1_" + strconv.Itoa(i))
			txn := NewTransactionWithSetup()
			txn.Start()
			txn.Write("couchdb", "item1", item)
			err := txn.Commit()
			assert.Nil(t, err)
		}

		// check the linked record length
		item, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, config.Config.MaxRecordLength, item.LinkedLen())

		tarItem := item
		for i := 1; i <= config.Config.MaxRecordLength-1; i++ {
			var preItem CouchDBItem
			err := json.Unmarshal([]byte(tarItem.Prev()), &preItem)
			assert.Nil(t, err)
			tarItem = &preItem
		}
		assert.Equal(t, "", tarItem.Prev())

		err = conn.Delete("item1")
		assert.NoError(t, err)
	})

	t.Run("4 commits immediately after txn.Start() when MaxRecordLength = 4",
		func(t *testing.T) {
			config.Config.MaxRecordLength = 4
			for i := 1; i <= 4; i++ {
				time.Sleep(10 * time.Millisecond)
				item := testutil.NewTestItem("item1_" + strconv.Itoa(i))
				txn := NewTransactionWithSetup()
				txn.Start()
				txn.Write("couchdb", "item1", item)
				err := txn.Commit()
				assert.Nil(t, err)
			}

			// check the linked record length
			conn := NewDefaultConnection()
			item, err := conn.GetItem("item1")
			assert.NoError(t, err)
			assert.Equal(t, config.Config.MaxRecordLength, item.LinkedLen())
			t.Logf("item: %+v", item)

			tarItem := item
			for i := 1; i <= config.Config.MaxRecordLength-1; i++ {
				var preItem CouchDBItem
				err := json.Unmarshal([]byte(tarItem.Prev()), &preItem)
				assert.Nil(t, err)
				tarItem = &preItem
			}
			assert.Equal(t, "", tarItem.Prev())

			err = conn.Delete("item1")
			assert.NoError(t, err)
		})

	t.Run("4 commits immediately after txn.Start() when MaxRecordLength = 5", func(t *testing.T) {

		config.Config.MaxRecordLength = 5
		expectedLen := min(4, config.Config.MaxRecordLength)
		for i := 1; i <= 4; i++ {
			time.Sleep(10 * time.Millisecond)
			item := testutil.NewTestItem("item1_" + strconv.Itoa(i))
			txn := NewTransactionWithSetup()
			txn.Start()
			txn.Write("couchdb", "item1", item)
			err := txn.Commit()
			assert.Nil(t, err)
		}

		// check the linked record length
		conn := NewDefaultConnection()
		item, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, expectedLen, item.LinkedLen())

		tarItem := item
		for i := 1; i <= expectedLen-1; i++ {
			var preItem CouchDBItem
			err := json.Unmarshal([]byte(tarItem.Prev()), &preItem)
			assert.Nil(t, err)
			tarItem = &preItem
		}
		assert.Equal(t, "", tarItem.Prev())

		err = conn.Delete("item1")
		assert.NoError(t, err)
	})
}

// The transcation should ***roll back*** the record then conditionalUpdate properly
func TestDirectWriteOnOutdatedPreparedRecordWithoutTSR(t *testing.T) {

	// final linked record should be "item1-cur" -> "item1-pre2"
	t.Run("the record has a valid Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()

		tarItem := &CouchDBItem{
			CKey:       "item1",
			CValue:     util.ToJSONString(testutil.NewTestItem("item1-pre2")),
			CTxnId:     "99",
			CTxnState:  config.COMMITTED,
			CTValid:    time.Now().Add(-10 * time.Second),
			CTLease:    time.Now().Add(-9 * time.Second),
			CLinkedLen: 1,
			// CVersion:   "1",
		}

		curItem := &CouchDBItem{
			CKey:       "item1",
			CValue:     util.ToJSONString(testutil.NewTestItem("item1-pre")),
			CTxnId:     "100",
			CTxnState:  config.PREPARED,
			CTValid:    time.Now().Add(-5 * time.Second),
			CTLease:    time.Now().Add(-4 * time.Second),
			CPrev:      util.ToJSONString(tarItem),
			CLinkedLen: 2,
			// CVersion:   "2",
		}

		conn.Delete(tarItem.Key())
		conn.PutItem(tarItem.Key(), curItem)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item1-cur")
		txn.Write("couchdb", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("couchdb", "item1", &resItem)
		assert.NoError(t, err)
		assert.Equal(t, "item1-cur", resItem.Value)

		// Second, we check the final record's Prev field is correct
		finalRedisItem, err := conn.GetItem("item1")
		assert.NoError(t, err)

		var resPreItem CouchDBItem
		err = json.Unmarshal([]byte(finalRedisItem.Prev()), &resPreItem)
		assert.NoError(t, err)

		tarItem.SetVersion(resPreItem.Version())
		assert.Equal(t, util.ToJSONString(tarItem), finalRedisItem.Prev())
	})

	// final linked record should be "item1-cur" -> "item1-pre(deleted)"
	t.Run("the record has an empty Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()

		tarItem := &CouchDBItem{
			CKey:      "item1",
			CValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
			CTxnId:    "99",
			CTxnState: config.PREPARED,
			CTValid:   time.Now().Add(-10 * time.Second),
			CTLease:   time.Now().Add(-9 * time.Second),
			// CVersion:  "1",
		}

		conn.Delete(tarItem.Key())
		conn.PutItem(tarItem.Key(), tarItem)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item1-cur")
		txn.Write("couchdb", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("couchdb", "item1", &resItem)
		assert.NoError(t, err)
		assert.Equal(t, "item1-cur", resItem.Value)

		// Second, we check the final record's Prev field is correct
		finalRedisItem, err := conn.GetItem("item1")
		assert.NoError(t, err)

		var resPreItem CouchDBItem
		err = json.Unmarshal([]byte(finalRedisItem.Prev()), &resPreItem)
		assert.NoError(t, err)

		tarItem.SetIsDeleted(true)
		tarItem.SetTxnState(config.COMMITTED)
		tarItem.SetVersion(resPreItem.Version())
		assert.Equal(t, util.ToJSONString(tarItem), finalRedisItem.Prev())
	})

}

// The transcation should ***roll forward*** the record then conditionalUpdate properly
func TestDirectWriteOnOutdatedPreparedRecordWithTSR(t *testing.T) {

	// final linked record should be "item2-cur" -> "item2-pre"
	t.Run("the record has a valid Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()

		tarItem := &CouchDBItem{
			CKey:       "item2",
			CValue:     util.ToJSONString(testutil.NewTestItem("item2-pre2")),
			CTxnId:     "TestDirectWriteOnOutdatedPreparedRecordWithTSR2",
			CTxnState:  config.COMMITTED,
			CTValid:    time.Now().Add(-10 * time.Second),
			CTLease:    time.Now().Add(-9 * time.Second),
			CLinkedLen: 1,
			CVersion:   "1",
		}

		curItem := &CouchDBItem{
			CKey:       "item2",
			CValue:     util.ToJSONString(testutil.NewTestItem("item2-pre")),
			CTxnId:     "TestDirectWriteOnOutdatedPreparedRecordWithTSR",
			CTxnState:  config.PREPARED,
			CTValid:    time.Now().Add(-5 * time.Second),
			CTLease:    time.Now().Add(-4 * time.Second),
			CLinkedLen: 2,
			// CVersion:   "2",
			CPrev: util.ToJSONString(tarItem),
		}

		conn.Delete(curItem.Key())
		conn.Delete("TestDirectWriteOnOutdatedPreparedRecordWithTSR")
		conn.PutItem(curItem.Key(), curItem)
		conn.Put("TestDirectWriteOnOutdatedPreparedRecordWithTSR", config.COMMITTED)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item2-cur")
		txn.Write("couchdb", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("couchdb", "item2", &resItem)
		assert.NoError(t, err)
		assert.Equal(t, "item2-cur", resItem.Value)

		// Second, we check the final record's Prev field is correct
		finalRedisItem, err := conn.GetItem("item2")
		assert.NoError(t, err)

		var resPreItem CouchDBItem
		err = json.Unmarshal([]byte(finalRedisItem.Prev()), &resPreItem)
		assert.NoError(t, err)

		curItem.SetPrev("")
		curItem.SetLinkedLen(1)
		curItem.SetTxnState(config.COMMITTED)
		curItem.SetVersion(resPreItem.Version())
		assert.Equal(t, util.ToJSONString(curItem), finalRedisItem.Prev())
	})

	// final linked record should be "item1-cur" -> "item1-pre"
	t.Run("the record has an empty Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()

		tarItem := &CouchDBItem{
			CKey:      "item1",
			CValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
			CTxnId:    "TestDirectWriteOnOutdatedPreparedRecordWithTSR",
			CTxnState: config.PREPARED,
			CTValid:   time.Now().Add(-10 * time.Second),
			CTLease:   time.Now().Add(-9 * time.Second),
			// CVersion:  "1",
		}

		conn.Delete(tarItem.Key())
		conn.Delete("TestDirectWriteOnOutdatedPreparedRecordWithTSR")
		conn.PutItem(tarItem.Key(), tarItem)
		conn.Put("TestDirectWriteOnOutdatedPreparedRecordWithTSR", config.COMMITTED)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item1-cur")
		txn.Write("couchdb", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("couchdb", "item1", &resItem)
		assert.NoError(t, err)
		assert.Equal(t, "item1-cur", resItem.Value)

		// Second, we check the final record's Prev field is correct
		finalRedisItem, err := conn.GetItem("item1")
		assert.NoError(t, err)

		var resPreItem CouchDBItem
		err = json.Unmarshal([]byte(finalRedisItem.Prev()), &resPreItem)
		assert.NoError(t, err)

		tarItem.SetVersion(resPreItem.Version())
		tarItem.SetTxnState(config.COMMITTED)
		assert.Equal(t, util.ToJSONString(tarItem), finalRedisItem.Prev())
	})

}

// The transaction should abort because version mismatch
func TestDirectWriteOnPreparingRecord(t *testing.T) {

	conn := NewDefaultConnection()

	tarItem := &CouchDBItem{
		CKey:      "item1",
		CValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		CTxnId:    "TestDirectWriteOnPreparingRecord",
		CTxnState: config.PREPARED,
		CTValid:   time.Now().Add(2 * time.Second),
		CTLease:   time.Now().Add(1 * time.Second),
		// CVersion:  "1",
	}

	conn.Delete(tarItem.Key())
	conn.PutItem(tarItem.Key(), tarItem)

	txn := NewTransactionWithSetup()
	txn.Start()
	item := testutil.NewTestItem("item1-cur")
	txn.Write("couchdb", tarItem.Key(), item)
	err := txn.Commit()
	assert.EqualError(t, err, "prepare phase failed: version mismatch")
}

func TestDirectWriteOnInvisibleRecord(t *testing.T) {
	conn := NewDefaultConnection()

	dbItem1 := &CouchDBItem{
		CKey:       "item1",
		CValue:     util.ToJSONString(testutil.NewTestItem("item1-pre1")),
		CTxnId:     "TestDirectWriteOnInvisibleRecord1",
		CTxnState:  config.COMMITTED,
		CTValid:    time.Now().Add(3 * time.Second),
		CTLease:    time.Now().Add(4 * time.Second),
		CLinkedLen: 1,
		// CVersion:   "2",
	}
	conn.Delete(dbItem1.Key())
	conn.PutItem(dbItem1.Key(), dbItem1)

	txn := NewTransactionWithSetup()
	txn.Start()
	item := testutil.NewTestItem("item1-cur")
	txn.Write("couchdb", dbItem1.Key(), item)
	err := txn.Commit()
	assert.EqualError(t, err, "prepare phase failed: version mismatch")

	// post check
	resItem, err := conn.GetItem("item1")
	assert.NoError(t, err)
	dbItem1.SetVersion(resItem.Version())
	if !resItem.Equal(dbItem1) {
		t.Errorf("\ngot\n %v\n \nwant\n %v\n", resItem, dbItem1)
	}
}

func TestDirectWriteOnDeletedRecord(t *testing.T) {
	// TODO:
}

func TestRollbackWhenReading(t *testing.T) {

	item1Pre := &CouchDBItem{
		CKey:      "item1",
		CValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		CTxnId:    "TestRollback",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-9 * time.Second),
		CVersion:  "1",
	}

	item1 := &CouchDBItem{
		CKey:      "item1",
		CValue:    util.ToJSONString(testutil.NewTestItem("item1")),
		CTxnId:    "TestRollback",
		CTxnState: config.PREPARED,
		CTValid:   time.Now().Add(-5 * time.Second),
		CTLease:   time.Now().Add(-4 * time.Second),
		// CVersion:  "2",
	}

	t.Run("rollback an item with a valid Prev field when reading", func(t *testing.T) {
		conn := NewDefaultConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev(util.ToJSONString(item1Pre))
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.Delete(todoRedisItem.Key())
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("couchdb", todoRedisItem.Key(), &item)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre", item.Value)
	})

	t.Run("rollback an item with an invalid Prev field when reading", func(t *testing.T) {
		conn := NewDefaultConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev("a broken prev field")
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.Delete(todoRedisItem.Key())
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("couchdb", todoRedisItem.Key(), &item)
		assert.NotNil(t, err)
	})

	t.Run("rollback an item with an empty Prev field when reading", func(t *testing.T) {
		conn := NewDefaultConnection()
		item1.SetPrev("")
		conn.Delete(item1.Key())
		conn.PutItem(item1.Key(), item1)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var item testutil.TestItem
		err := txn1.Read("couchdb", item1.Key(), &item)
		assert.EqualError(t, err, trxn.KeyNotFound.Error())
	})
}

func TestRollbackWhenWriting(t *testing.T) {
	item1Pre := &CouchDBItem{
		CKey:      "item1",
		CValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		CTxnId:    "TestRollback",
		CTxnState: config.COMMITTED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-9 * time.Second),
		CVersion:  "1",
	}

	item1 := &CouchDBItem{
		CKey:      "item1",
		CValue:    util.ToJSONString(testutil.NewTestItem("item1")),
		CTxnId:    "TestRollback",
		CTxnState: config.PREPARED,
		CTValid:   time.Now().Add(-5 * time.Second),
		CTLease:   time.Now().Add(-4 * time.Second),
		// CVersion:  "2",
	}

	t.Run("rollback an item with a valid Prev field when writing", func(t *testing.T) {
		conn := NewDefaultConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev(util.ToJSONString(item1Pre))
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.Delete(todoRedisItem.Key())
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		item := testutil.NewTestItem("item1-cur")
		txn.Write("couchdb", todoRedisItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		res, err := conn.GetItem(todoRedisItem.Key())
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(item), res.Value())

		var resPrevItem CouchDBItem
		err = json.Unmarshal([]byte(res.Prev()), &resPrevItem)
		assert.NoError(t, err)

		tarItem := item1Pre
		tarItem.SetVersion(resPrevItem.Version())
		assert.Equal(t, util.ToJSONString(tarItem), res.Prev())

	})

	t.Run("rollback an item with an invalid Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev("a broken prev field")
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.Delete(todoRedisItem.Key())
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		item := testutil.NewTestItem("item1-cur")
		txn.Write("couchdb", todoRedisItem.Key(), item)
		err := txn.Commit()
		assert.NotNil(t, err)
	})

	t.Run("rollback an item with an empty Prev field when writing", func(t *testing.T) {
		conn := NewDefaultConnection()
		item1.SetPrev("")
		item1.SetLinkedLen(1)
		conn.Delete(item1.Key())
		conn.PutItem(item1.Key(), item1)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		item := testutil.NewTestItem("item1-cur")
		txn1.Write("couchdb", item1.Key(), item)
		err := txn1.Commit()
		assert.NoError(t, err)

		res, err := conn.GetItem(item1.Key())
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(item), res.Value())

		var resPreItem CouchDBItem
		err = json.Unmarshal([]byte(res.Prev()), &resPreItem)
		assert.NoError(t, err)

		item1.SetIsDeleted(true)
		item1.SetTxnState(config.COMMITTED)
		item1.SetVersion(resPreItem.Version())
		assert.Equal(t, util.ToJSONString(item1), res.Prev())
	})
}

func TestRollForwardWhenReading(t *testing.T) {

	conn := NewDefaultConnection()

	tarItem := &CouchDBItem{
		CKey:      "item1",
		CValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		CTxnId:    "TestRollForward",
		CTxnState: config.PREPARED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-9 * time.Second),
		// CVersion:  "1",
	}

	conn.Delete(tarItem.Key())
	conn.Delete("TestRollForward")
	conn.PutItem(tarItem.Key(), tarItem)
	conn.Put("TestRollForward", config.COMMITTED)

	// the transaction should roll forward the item
	txn := NewTransactionWithSetup()
	txn.Start()
	var item testutil.TestItem
	err := txn.Read("couchdb", tarItem.Key(), &item)
	assert.NoError(t, err)

	res, err := conn.GetItem(tarItem.Key())
	assert.NoError(t, err)
	tarItem.SetTxnState(config.COMMITTED)
	tarItem.SetVersion(res.Version())

	if !res.Equal(tarItem) {
		t.Errorf("\ngot\n %v \nwant\n %v", res, tarItem)
	}
}

func TestRollForwardWhenWriting(t *testing.T) {

	conn := NewDefaultConnection()

	tarItem := &CouchDBItem{
		CKey:      "item1",
		CValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		CTxnId:    "TestRollForward",
		CTxnState: config.PREPARED,
		CTValid:   time.Now().Add(-10 * time.Second),
		CTLease:   time.Now().Add(-9 * time.Second),
		// CVersion:  "1",
	}

	conn.Delete(tarItem.Key())
	conn.Delete("TestRollForward")
	_, err := conn.PutItem(tarItem.Key(), tarItem)
	assert.NoError(t, err)
	conn.Put("TestRollForward", config.COMMITTED)

	// the transaction should roll forward the item
	txn := NewTransactionWithSetup()
	txn.Start()
	item := testutil.NewTestItem("item1-cur")
	txn.Write("couchdb", tarItem.Key(), item)
	err = txn.Commit()
	assert.NoError(t, err)

	res, err := conn.GetItem(tarItem.Key())
	assert.NoError(t, err)
	assert.Equal(t, util.ToJSONString(item), res.Value())

	var resPrevItem CouchDBItem
	err = json.Unmarshal([]byte(res.Prev()), &resPrevItem)
	assert.NoError(t, err)

	tarItem.SetTxnState(config.COMMITTED)
	tarItem.SetVersion(resPrevItem.Version())
	tarPrev := util.ToJSONString(tarItem)
	assert.Equal(t, tarPrev, res.Prev())
}

func TestItemVersionUpdate(t *testing.T) {

	t.Run("item version ++ after updated", func(t *testing.T) {
		conn := NewDefaultConnection()
		dbItem := &CouchDBItem{
			CKey:       "item1",
			CValue:     util.ToJSONString(testutil.NewTestItem("item1-pre")),
			CTxnId:     "TestItemVersionUpdate",
			CTxnState:  config.COMMITTED,
			CTValid:    time.Now().Add(-10 * time.Second),
			CTLease:    time.Now().Add(-9 * time.Second),
			CLinkedLen: 1,
			CVersion:   "1",
		}
		conn.PutItem(dbItem.Key(), dbItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("couchdb", dbItem.Key(), &item)
		assert.NoError(t, err)
		item.Value = "item1-cur"
		txn.Write("couchdb", dbItem.Key(), item)
		err = txn.Commit()
		assert.NoError(t, err)

		res, err := conn.GetItem(dbItem.Key())
		assert.NoError(t, err)

		assert.Equal(t, res.Version(), res.Version())
	})
}

func TestTSROperations(t *testing.T) {

	t.Run("test WriteTSR", func(t *testing.T) {
		conn := NewDefaultConnection()
		cds := NewCouchDBDatastore("couchdb", conn)
		ds := cds.(trxn.TSRMaintainer)
		conn.Delete("txn-1")

		err := ds.WriteTSR("txn-1", config.COMMITTED)
		assert.NoError(t, err)
	})

	t.Run("test DeleteTSR", func(t *testing.T) {
		conn := NewDefaultConnection()
		cds := NewCouchDBDatastore("couchdb", conn)
		ds := cds.(trxn.TSRMaintainer)

		err := ds.DeleteTSR("txn-1")
		assert.NoError(t, err)
	})

	t.Run("test ReadTSR", func(t *testing.T) {
		conn := NewDefaultConnection()
		cds := NewCouchDBDatastore("couchdb", conn)
		ds := cds.(trxn.TSRMaintainer)

		ds.DeleteTSR("txn-1")

		err := ds.WriteTSR("txn-1", config.COMMITTED)
		assert.NoError(t, err)

		state, err := ds.ReadTSR("txn-1")
		assert.NoError(t, err)
		assert.Equal(t, config.COMMITTED, state)
	})
}
