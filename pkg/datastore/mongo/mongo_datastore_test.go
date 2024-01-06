package mongo

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

// NewDefaultMongoConnection creates a new instance of MongoConnection with default settings.
// It establishes a connection to the MongoDB server running on localhost:27017.
// Returns a pointer to the MongoConnection.
func NewDefaultMongoConnection() *MongoConnection {
	conn := NewMongoConnection(&ConnectionOptions{
		Address:        "mongodb://localhost:27017",
		DBName:         "oreo",
		CollectionName: "records",
	})
	conn.Connect()
	return conn
}

func NewTransactionWithSetup() *txn.Transaction {
	conn := NewMongoConnection(&ConnectionOptions{
		Address:        "mongodb://localhost:27017",
		DBName:         "oreo",
		CollectionName: "records",
	})
	txn := txn.NewTransaction()
	mds := NewMongoDatastore("mongo", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)
	return txn
}

func TestSimpleReadInCache(t *testing.T) {
	txn := txn.NewTransaction()
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	memoryPerson := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(memoryPerson),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	// conn.Delete(key)
	err := conn.PutItem(key, expectedMongoItem)
	if err != nil {
		t.Errorf("Error putting item to Mongo datastore: %s", err)
	}

	// Put a item in cache
	cachePerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	cacheMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(cachePerson),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  2,
	}
	mds.writeCache[key] = cacheMongoItem

	// Start the transaction
	err = txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	// Check the result
	if result != cachePerson {
		t.Errorf("got %v want %v", result, cachePerson)
	}
}

func TestSimpleReadWhenCommitted(t *testing.T) {

	txn := txn.NewTransaction()
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.PutItem(key, expectedMongoItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	// Check the result
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenCommittedFindPrevious(t *testing.T) {
	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	preMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "99",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  1,
	}
	curMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(curPerson),
		TxnId:    "100",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  2,
		Prev:     util.ToJSONString(preMongoItem),
	}

	key := "John"
	conn.PutItem(key, curMongoItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	// Check the result
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenCommittedFindNone(t *testing.T) {
	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	preMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "99",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  1,
	}
	curMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(curPerson),
		TxnId:    "100",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(20 * time.Second),
		TLease:   time.Now().Add(15 * time.Second),
		Version:  2,
		Prev:     util.ToJSONString(preMongoItem),
	}

	key := "John"
	conn.PutItem(key, curMongoItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}
}

// TestSimpleReadWhenPreparedWithTSRInCOMMITTED tests the scenario where a simple read operation is performed
// on a record which is in PREPARED state and has a TSR in COMMITTED state.
func TestSimpleReadWhenPreparedWithTSRInCOMMITTED(t *testing.T) {
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "100",
		TxnState: config.PREPARED,
		TValid:   time.Now(),
		TLease:   time.Now(),
		Version:  2,
	}

	key := "John"
	conn.PutItem(key, expectedMongoItem)

	// Write the TSR
	conn.Put("100", config.COMMITTED)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
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
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	tarMemItem := MongoItem{
		Key:      "item1",
		Value:    util.ToJSONString(testutil.NewTestItem("item1")),
		TxnId:    "99",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-9 * time.Second),
		Version:  1,
	}

	curMemItem := MongoItem{
		Key:      "item1",
		Value:    util.ToJSONString(testutil.NewTestItem("item1-prepared")),
		TxnId:    "TestSimpleReadWhenPreparedWithTSRInABORTED",
		TxnState: config.PREPARED,
		TValid:   time.Now().Add(-5 * time.Second),
		TLease:   time.Now().Add(-4 * time.Second),
		Prev:     util.ToJSONString(tarMemItem),
		Version:  2,
	}

	key := "item1"
	err := conn.PutItem(key, curMemItem)
	assert.NoError(t, err)

	// Write the TSR
	err = conn.Put("TestSimpleReadWhenPreparedWithTSRInABORTED", config.ABORTED)
	assert.NoError(t, err)

	// Start the transaction
	err = txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.TestItem
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}
	expected := testutil.NewTestItem("item1")
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}

	// Delete the TSR
	conn.Delete("TestSimpleReadWhenPreparedWithTSRInABORTED")
}

func TestSimpleReadWhenPrepareExpired(t *testing.T) {
	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "100",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	expectedStr := util.ToJSONString(expectedMongoItem)

	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}

	curMongoItem := MongoItem{
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
	conn.PutItem(key, curMongoItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenPrepareNotExpired(t *testing.T) {
	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "100",
		TxnState: config.PREPARED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  2,
	}

	key := "John"
	err := conn.PutItem(key, expectedMongoItem)
	if err != nil {
		t.Errorf("Error putting item to Mongo datastore: %s", err)
	}

	// Start the transaction
	err = txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err.Error() != errors.New("dirty Read").Error() {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}
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
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to Mongo datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleDirectWrite(t *testing.T) {

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	key := "John"
	prePerson := testutil.NewPerson("John-pre")
	preTxn.Write("mongo", key, prePerson)
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
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to Mongo datastore: %s", err)
	}
	err = txn.Commit()
	assert.NoError(t, err)
}

func TestSimpleWriteAndReadLocal(t *testing.T) {
	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
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
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to Mongo datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleReadModifyWriteThenRead(t *testing.T) {
	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.PutItem(key, expectedMongoItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	// Modify the value
	result.Age = 31

	// Write the value
	err = txn.Write("mongo", key, result)
	if err != nil {
		t.Errorf("Error writing to Mongo datastore: %s", err)
	}

	// Read the value
	var result2 testutil.Person
	err = txn.Read("mongo", key, &result2)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	// Check the result
	if result2 != result {
		t.Errorf("got %v want %v", result2, result)
	}
}

func TestSimpleOverwriteAndRead(t *testing.T) {
	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.PutItem(key, expectedMongoItem)

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
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to Mongo datastore: %s", err)
	}
	person.Age = 32
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to Mongo datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleDeleteAndRead(t *testing.T) {
	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.PutItem(key, expectedMongoItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Delete the value
	err = txn.Delete("mongo", key)
	if err != nil {
		t.Errorf("Error deleting from Mongo datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}
}

func TestSimpleDeleteTwice(t *testing.T) {
	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	err := conn.PutItem(key, expectedMongoItem)
	if err != nil {
		t.Errorf("Error putting item to Mongo datastore: %s", err)
	}

	// Start the transaction
	err = txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Delete the value
	err = txn.Delete("mongo", key)
	if err != nil {
		t.Errorf("Error deleting from Mongo datastore: %s", err)
	}
	err = txn.Delete("mongo", key)
	if err.Error() != "key not found" {
		t.Errorf("Error deleting from Mongo datastore: %s", err)
	}
}

func TestDeleteWithRead(t *testing.T) {

	conn := NewDefaultMongoConnection()
	// clear the test data
	conn.Delete("John")

	preTxn := NewTransactionWithSetup()
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("mongo", "John", dataPerson)
	err := preTxn.Commit()
	assert.NoError(t, err)

	txn := NewTransactionWithSetup()
	txn.Start()
	var person testutil.Person
	err = txn.Read("mongo", "John", &person)
	assert.NoError(t, err)
	err = txn.Delete("mongo", "John")
	assert.NoError(t, err)

	err = txn.Commit()
	assert.NoError(t, err)
}

func TestDeleteWithoutRead(t *testing.T) {

	preTxn := NewTransactionWithSetup()
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("mongo", "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	txn.Start()
	err := txn.Delete("mongo", "John")
	if err != nil {
		t.Errorf("Error deleting from Mongo datastore: %s", err)
	}

	err = txn.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

}

func TestSimpleReadWriteDeleteThenRead(t *testing.T) {
	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.PutItem(key, expectedMongoItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var person testutil.Person
	err = txn.Read("mongo", key, &person)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	person.Age = 31

	// Write the value
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to Mongo datastore: %s", err)
	}

	// Delete the value
	err = txn.Delete("mongo", key)
	if err != nil {
		t.Errorf("Error deleting from Mongo datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}
}

func TestSimpleWriteDeleteWriteThenRead(t *testing.T) {
	// Create a new Mongo datastore
	conn := NewDefaultMongoConnection()
	mds := NewMongoDatastore("mongo", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	// initialize the Mongo database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedMongoItem := MongoItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	conn.PutItem(key, expectedMongoItem)

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
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to Mongo datastore: %s", err)
	}

	// Delete the value
	err = txn.Delete("mongo", key)
	if err != nil {
		t.Errorf("Error deleting from Mongo datastore: %s", err)
	}

	// Write the value
	person.Age = 32
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to Mongo datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from Mongo datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}

}

func TestMongoDatastore_ConcurrentWriteConflicts(t *testing.T) {

	// clear the test data
	conn := NewDefaultMongoConnection()
	for _, item := range testutil.InputItemList {
		conn.Delete(item.Value)
	}

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write("mongo", item.Value, item)
	}
	err := preTxn.Commit()
	assert.NoError(t, err)

	resChan := make(chan bool)
	successId := 0

	concurrentCount := 50

	for i := 1; i <= concurrentCount; i++ {
		go func(id int) {
			txn := NewTransactionWithSetup()
			txn.Start()
			for _, item := range testutil.InputItemList {
				var res testutil.TestItem
				txn.Read("mongo", item.Value, &res)
				res.Value = item.Value + "-new-" + strconv.Itoa(id)
				txn.Write("mongo", item.Value, res)
			}

			time.Sleep(100 * time.Millisecond)
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
		postTxn.Read("mongo", item.Value, &res)
		assert.Equal(t, item.Value+"-new-"+strconv.Itoa(successId), res.Value)
	}
	err = postTxn.Commit()
	assert.NoError(t, err)

}

func TestTxnWriteMultiRecord(t *testing.T) {

	// clear the test data
	conn := NewDefaultMongoConnection()
	conn.Delete("item1")
	conn.Delete("item2")

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	preTxn.Write("mongo", "item1", testutil.NewTestItem("item1"))
	preTxn.Write("mongo", "item2", testutil.NewTestItem("item2"))
	err := preTxn.Commit()
	assert.Nil(t, err)

	txn := NewTransactionWithSetup()
	txn.Start()
	var item testutil.TestItem
	txn.Read("mongo", "item1", &item)
	item.Value = "item1_new"
	txn.Write("mongo", "item1", item)

	txn.Read("mongo", "item2", &item)
	item.Value = "item2_new"
	txn.Write("mongo", "item2", item)

	err = txn.Commit()
	assert.Nil(t, err)

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	var resItem testutil.TestItem
	postTxn.Read("mongo", "item1", &resItem)
	assert.Equal(t, "item1_new", resItem.Value)
	postTxn.Read("mongo", "item2", &resItem)
	assert.Equal(t, "item2_new", resItem.Value)

}

func TestMongoDatastore_ReadTSR(t *testing.T) {
	conn := NewDefaultMongoConnection()

	conn.Put("ReadTSR_test", config.COMMITTED)

}
