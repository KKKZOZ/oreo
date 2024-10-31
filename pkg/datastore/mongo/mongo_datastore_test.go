package mongo

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/testutil"
	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	trxn "github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

// NewDefaultMongoConnection creates a new instance of MongoConnection with default settings.
// It establishes a connection to the MongoDB server running on localhost:27017.
// Returns a pointer to the MongoConnection.
func NewDefaultConnection() *MongoConnection {
	conn := NewMongoConnection(&ConnectionOptions{
		Address:        "mongodb://localhost:27017",
		DBName:         "oreo",
		CollectionName: "records",
		Username:       "admin",
		Password:       "admin",
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
	conn := NewMongoConnection(&ConnectionOptions{
		Address:        "mongodb://localhost:27017",
		DBName:         "oreo",
		CollectionName: "records",
		Username:       "admin",
		Password:       "admin",
	})
	txn := trxn.NewTransaction()
	rds := NewMongoDatastore("mongo", conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)
	return txn
}

func TestSimpleReadWhenCommitted(t *testing.T) {

	txn := trxn.NewTransaction()
	conn := NewDefaultConnection()
	rds := NewMongoDatastore("mongo", conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "123123",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-5 * time.Second),
		MVersion:      "2",
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
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenCommittedFindEmpty(t *testing.T) {

	txn1 := trxn.NewTransaction()
	conn := NewDefaultConnection()
	rds := NewMongoDatastore("mongo", conn)
	txn1.AddDatastore(rds)
	txn1.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "TestSimpleReadWhenCommittedFindEmpty",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(+10 * time.Second),
		MTLease:       time.Now().Add(+5 * time.Second),
		MVersion:      "2",
	}

	key := "John"
	conn.PutItem(key, expectedRedisItem)

	// Start the transaction
	err := txn1.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn1.Read("mongo", key, &result)
	assert.EqualError(t, err, trxn.KeyNotFound.Error())

}

func TestSimpleReadWhenCommittedFindPrevious(t *testing.T) {
	// Create a new transaction
	txn := trxn.NewTransaction()

	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewMongoDatastore("mongo", conn)
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
	preRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "99",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-5 * time.Second),
		MVersion:      "1",
	}
	curRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(curPerson),
		MGroupKeyList: "100",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(10 * time.Second),
		MTLease:       time.Now().Add(5 * time.Second),
		MVersion:      "2",
		MPrev:         util.ToJSONString(preRedisItem),
	}

	key := "John"
	conn.PutItem(key, curRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
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
	rds := NewMongoDatastore("mongo", conn)
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
	preRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "99",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(10 * time.Second),
		MTLease:       time.Now().Add(5 * time.Second),
		MVersion:      "1",
	}
	curRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(curPerson),
		MGroupKeyList: "100",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(20 * time.Second),
		MTLease:       time.Now().Add(15 * time.Second),
		MVersion:      "2",
		MPrev:         util.ToJSONString(preRedisItem),
	}

	key := "John"
	conn.PutItem(key, curRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from redis datastore: %s", err)
	}
}

// TestSimpleReadWhenPreparedWithTSRInCOMMITTED tests the scenario where a simple read operation is performed
// on a record which is in PREPARED state and has a TSR in COMMITTED state.
func TestSimpleReadWhenPreparedWithTSRInCOMMITTED(t *testing.T) {
	conn := NewDefaultConnection()
	rds := NewMongoDatastore("mongo", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "100",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now(),
		MTLease:       time.Now(),
		MVersion:      "2",
	}

	key := "John"
	conn.PutItem(key, expectedRedisItem)

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
	rds := NewMongoDatastore("mongo", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	tarMemItem := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1")),
		MGroupKeyList: "99",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-9 * time.Second),
		MVersion:      "1",
	}

	curMemItem := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-prepared")),
		MGroupKeyList: "TestSimpleReadWhenPreparedWithTSRInABORTED",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now().Add(-5 * time.Second),
		MTLease:       time.Now().Add(-4 * time.Second),
		MPrev:         util.ToJSONString(tarMemItem),
		MVersion:      "2",
	}

	key := "item1"
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
	err = txn.Read("mongo", key, &result)
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
	rds := NewMongoDatastore("mongo", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "100",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-5 * time.Second),
		MVersion:      "2",
	}

	expectedStr := util.ToJSONString(expectedRedisItem)

	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}

	curRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(curPerson),
		MGroupKeyList: "101",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now().Add(-3 * time.Second),
		MTLease:       time.Now().Add(-1 * time.Second),
		MVersion:      "3",
		MPrev:         expectedStr,
	}

	key := "John"
	conn.PutItem(key, curRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenPrepareNotExpired(t *testing.T) {

	dbItem1 := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre1")),
		MGroupKeyList: "TestSimpleReadWhenPrepareNotExpired1",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-2 * time.Second),
		MTLease:       time.Now().Add(-1 * time.Second),
		MLinkedLen:    1,
		MVersion:      "1",
	}

	dbItem2 := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre2")),
		MGroupKeyList: "TestSimpleReadWhenPrepareNotExpired2",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now().Add(1 * time.Second),
		MTLease:       time.Now().Add(2 * time.Second),
		MPrev:         util.ToJSONString(dbItem1),
		MLinkedLen:    2,
		MVersion:      "2",
	}

	t.Run("when the item has a valid Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()
		conn.PutItem("item1", dbItem2)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var item testutil.TestItem
		err := txn1.Read("mongo", "item1", &item)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre1", item.Value)
	})

	t.Run("when the item has an empty Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()
		dbItem := dbItem2
		dbItem.SetPrev("")
		dbItem.SetLinkedLen(1)
		conn.PutItem("item1", dbItem)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var item testutil.TestItem
		err := txn1.Read("mongo", "item1", &item)
		assert.EqualError(t, err, trxn.KeyNotFound.Error())
	})
}

func TestSimpleReadWhenDeleted(t *testing.T) {
	conn := NewDefaultConnection()
	dbItem := &MongoItem{
		MKey:       "item2",
		MValue:     util.ToJSONString(testutil.NewTestItem("item2-db")),
		MTxnState:  config.COMMITTED,
		MTValid:    time.Now().Add(-2 * time.Second),
		MTLease:    time.Now().Add(-1 * time.Second),
		MLinkedLen: 1,
		MVersion:   "1",
		MIsDeleted: true,
	}

	conn.PutItem(dbItem.Key(), dbItem)

	txn1 := NewTransactionWithSetup()
	txn1.Start()

	var item testutil.TestItem
	err := txn1.Read("mongo", "item2", &item)
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
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
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
		t.Errorf("Error writing to redis datastore: %s", err)
	}
	err = txn.Commit()
	assert.NoError(t, err)
}

func TestSimpleWriteAndReadLocal(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewMongoDatastore("mongo", conn)
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
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
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
	rds := NewMongoDatastore("mongo", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "123123",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-5 * time.Second),
		MVersion:      "2",
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
	err = txn.Read("mongo", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Modify the value
	result.Age = 31

	// Write the value
	err = txn.Write("mongo", key, result)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result2 testutil.Person
	err = txn.Read("mongo", key, &result2)
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
	rds := NewMongoDatastore("mongo", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "123123",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-5 * time.Second),
		MVersion:      "2",
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
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}
	person.Age = 32
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
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
	rds := NewMongoDatastore("mongo", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "123123",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-5 * time.Second),
		MVersion:      "2",
	}

	key := "John"
	conn.PutItem(key, expectedRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Delete the value
	err = txn.Delete("mongo", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from redis datastore: %s", err)
	}
}

func TestSimpleDeleteTwice(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewMongoDatastore("mongo", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "123123",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-5 * time.Second),
		MVersion:      "2",
	}

	key := "John"
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
	err = txn.Delete("mongo", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}
	err = txn.Delete("mongo", key)
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
	rds := NewMongoDatastore("mongo", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "123123",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-5 * time.Second),
		MVersion:      "2",
	}

	key := "John"
	conn.PutItem(key, expectedRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var person testutil.Person
	err = txn.Read("mongo", key, &person)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	person.Age = 31

	// Write the value
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Delete the value
	err = txn.Delete("mongo", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from redis datastore: %s", err)
	}
}

func TestSimpleWriteDeleteWriteThenRead(t *testing.T) {
	// Create a new redis datastore
	conn := NewDefaultConnection()
	rds := NewMongoDatastore("mongo", conn)
	txn := trxn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &MongoItem{
		MKey:          "John",
		MValue:        util.ToJSONString(expected),
		MGroupKeyList: "123123",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-5 * time.Second),
		MVersion:      "2",
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
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Delete the value
	err = txn.Delete("mongo", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	// Write the value
	person.Age = 32
	err = txn.Write("mongo", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("mongo", key, &result)
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
		preTxn.Write("mongo", item.Value, item)
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
				txn.Read("mongo", item.Value, &res)
				res.Value = item.Value + "-new-" + strconv.Itoa(id)
				txn.Write("mongo", item.Value, res)
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
		postTxn.Read("mongo", item.Value, &res)
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

// ---|---------|--------|---------|------> time
// item1_1  T_Start   item1_2   item1_3
func TestLinkedReadAsCommitted(t *testing.T) {

	item1_1 := testutil.NewTestItem("item1_1")
	memItem1_1 := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(item1_1),
		MGroupKeyList: "txn1",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-9 * time.Second),
		MVersion:      "1",
		MLinkedLen:    1,
	}

	item1_2 := testutil.NewTestItem("item1_2")
	memItem1_2 := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(item1_2),
		MGroupKeyList: "txn2",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(5 * time.Second),
		MTLease:       time.Now().Add(6 * time.Second),
		MVersion:      "2",
		MPrev:         util.ToJSONString(memItem1_1),
		MLinkedLen:    2,
	}

	item1_3 := testutil.NewTestItem("item1_3")
	memItem1_3 := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(item1_3),
		MGroupKeyList: "txn3",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(10 * time.Second),
		MTLease:       time.Now().Add(11 * time.Second),
		MVersion:      "3",
		MPrev:         util.ToJSONString(memItem1_2),
		MLinkedLen:    3,
	}

	t.Run("read will fail due to MaxRecordLength=2", func(t *testing.T) {

		conn := NewDefaultConnection()
		_, err := conn.PutItem("item1", memItem1_3)
		assert.NoError(t, err)

		config.Config.MaxRecordLength = 2
		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err = txn.Read("mongo", "item1", &item)
		assert.EqualError(t, err, "key not found")
	})

	t.Run("read will success due to MaxRecordLength=3", func(t *testing.T) {

		conn := NewDefaultConnection()
		conn.PutItem("item1", memItem1_3)

		config.Config.MaxRecordLength = 3
		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("mongo", "item1", &item)
		assert.Nil(t, err)

		assert.Equal(t, "item1_1", item.Value)
	})

	t.Run("read will success due to MaxRecordLength > 3", func(t *testing.T) {

		conn := NewDefaultConnection()
		conn.PutItem("item1", memItem1_3)

		config.Config.MaxRecordLength = 3 + 1
		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("mongo", "item1", &item)
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
			txn.Write("mongo", "item1", item)
			err := txn.Commit()
			assert.Nil(t, err)
		}

		// check the linked record length
		item, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, config.Config.MaxRecordLength, item.LinkedLen())

		tarItem := item
		for i := 1; i <= config.Config.MaxRecordLength-1; i++ {
			var preItem MongoItem
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
				txn.Write("mongo", "item1", item)
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
				var preItem MongoItem
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
			txn.Write("mongo", "item1", item)
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
			var preItem MongoItem
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

		tarItem := &MongoItem{
			MKey:          "item1",
			MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre2")),
			MGroupKeyList: "99",
			MTxnState:     config.COMMITTED,
			MTValid:       time.Now().Add(-10 * time.Second),
			MTLease:       time.Now().Add(-9 * time.Second),
			MLinkedLen:    1,
			MVersion:      "1",
		}

		curItem := &MongoItem{
			MKey:          "item1",
			MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
			MGroupKeyList: "100",
			MTxnState:     config.PREPARED,
			MTValid:       time.Now().Add(-5 * time.Second),
			MTLease:       time.Now().Add(-4 * time.Second),
			MPrev:         util.ToJSONString(tarItem),
			MLinkedLen:    2,
			MVersion:      "2",
		}

		conn.PutItem(tarItem.Key(), curItem)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item1-cur")
		txn.Write("mongo", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("mongo", "item1", &resItem)
		assert.NoError(t, err)
		assert.Equal(t, "item1-cur", resItem.Value)

		// Second, we check the final record's Prev field is correct
		finalRedisItem, err := conn.GetItem("item1")
		assert.NoError(t, err)
		tarItem.SetVersion("3")
		assert.Equal(t, util.ToJSONString(tarItem), finalRedisItem.Prev())
	})

	// final linked record should be "item1-cur" -> "item1-pre(deleted)"
	t.Run("the record has an empty Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()

		tarItem := &MongoItem{
			MKey:          "item1",
			MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
			MGroupKeyList: "99",
			MTxnState:     config.PREPARED,
			MTValid:       time.Now().Add(-10 * time.Second),
			MTLease:       time.Now().Add(-9 * time.Second),
			MVersion:      "1",
		}

		conn.PutItem(tarItem.Key(), tarItem)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item1-cur")
		txn.Write("mongo", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("mongo", "item1", &resItem)
		assert.NoError(t, err)
		assert.Equal(t, "item1-cur", resItem.Value)

		// Second, we check the final record's Prev field is correct
		finalRedisItem, err := conn.GetItem("item1")
		assert.NoError(t, err)
		tarItem.SetIsDeleted(true)
		tarItem.SetTxnState(config.COMMITTED)
		tarItem.SetVersion(util.AddToString(tarItem.Version(), 1))
		assert.Equal(t, util.ToJSONString(tarItem), finalRedisItem.Prev())
	})

}

// The transcation should ***roll forward*** the record then conditionalUpdate properly
func TestDirectWriteOnOutdatedPreparedRecordWithTSR(t *testing.T) {

	// final linked record should be "item2-cur" -> "item2-pre"
	t.Run("the record has a valid Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()

		tarItem := &MongoItem{
			MKey:          "item2",
			MValue:        util.ToJSONString(testutil.NewTestItem("item2-pre2")),
			MGroupKeyList: "TestDirectWriteOnOutdatedPreparedRecordWithTSR2",
			MTxnState:     config.COMMITTED,
			MTValid:       time.Now().Add(-10 * time.Second),
			MTLease:       time.Now().Add(-9 * time.Second),
			MLinkedLen:    1,
			MVersion:      "1",
		}

		curItem := &MongoItem{
			MKey:          "item2",
			MValue:        util.ToJSONString(testutil.NewTestItem("item2-pre")),
			MGroupKeyList: "TestDirectWriteOnOutdatedPreparedRecordWithTSR",
			MTxnState:     config.PREPARED,
			MTValid:       time.Now().Add(-5 * time.Second),
			MTLease:       time.Now().Add(-4 * time.Second),
			MLinkedLen:    2,
			MVersion:      "2",
			MPrev:         util.ToJSONString(tarItem),
		}

		conn.PutItem(curItem.Key(), curItem)
		conn.Put("TestDirectWriteOnOutdatedPreparedRecordWithTSR", config.COMMITTED)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item2-cur")
		txn.Write("mongo", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("mongo", "item2", &resItem)
		assert.NoError(t, err)
		assert.Equal(t, "item2-cur", resItem.Value)

		// Second, we check the final record's Prev field is correct
		finalRedisItem, err := conn.GetItem("item2")
		assert.NoError(t, err)

		curItem.SetPrev("")
		curItem.SetLinkedLen(1)
		curItem.SetTxnState(config.COMMITTED)
		curItem.SetVersion("3")
		assert.Equal(t, util.ToJSONString(curItem), finalRedisItem.Prev())
	})

	// final linked record should be "item1-cur" -> "item1-pre"
	t.Run("the record has an empty Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()

		tarItem := &MongoItem{
			MKey:          "item1",
			MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
			MGroupKeyList: "TestDirectWriteOnOutdatedPreparedRecordWithTSR",
			MTxnState:     config.PREPARED,
			MTValid:       time.Now().Add(-10 * time.Second),
			MTLease:       time.Now().Add(-9 * time.Second),
			MVersion:      "1",
		}

		conn.PutItem(tarItem.Key(), tarItem)
		conn.Put("TestDirectWriteOnOutdatedPreparedRecordWithTSR", config.COMMITTED)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item1-cur")
		txn.Write("mongo", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("mongo", "item1", &resItem)
		assert.NoError(t, err)
		assert.Equal(t, "item1-cur", resItem.Value)

		// Second, we check the final record's Prev field is correct
		finalRedisItem, err := conn.GetItem("item1")
		assert.NoError(t, err)

		tarItem.SetVersion("2")
		tarItem.SetTxnState(config.COMMITTED)
		assert.Equal(t, util.ToJSONString(tarItem), finalRedisItem.Prev())
	})

}

// The transaction should abort because version mismatch
func TestDirectWriteOnPreparingRecord(t *testing.T) {

	conn := NewDefaultConnection()

	tarItem := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
		MGroupKeyList: "TestDirectWriteOnPreparingRecord",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now().Add(2 * time.Second),
		MTLease:       time.Now().Add(1 * time.Second),
		MVersion:      "1",
	}

	conn.PutItem(tarItem.Key(), tarItem)

	txn := NewTransactionWithSetup()
	txn.Start()
	item := testutil.NewTestItem("item1-cur")
	txn.Write("mongo", tarItem.Key(), item)
	err := txn.Commit()
	assert.EqualError(t, err, "prepare phase failed: version mismatch")
}

func TestDirectWriteOnInvisibleRecord(t *testing.T) {
	conn := NewDefaultConnection()

	dbItem1 := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre1")),
		MGroupKeyList: "TestDirectWriteOnInvisibleRecord1",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(3 * time.Second),
		MTLease:       time.Now().Add(4 * time.Second),
		MLinkedLen:    1,
		MVersion:      "2",
	}
	conn.PutItem(dbItem1.Key(), dbItem1)

	txn := NewTransactionWithSetup()
	txn.Start()
	item := testutil.NewTestItem("item1-cur")
	txn.Write("mongo", dbItem1.Key(), item)
	err := txn.Commit()
	assert.EqualError(t, err, "prepare phase failed: version mismatch")

	// post check
	resItem, err := conn.GetItem("item1")
	assert.NoError(t, err)
	if !resItem.Equal(dbItem1) {
		t.Errorf("\ngot\n %v\n \nwant\n %v\n", resItem, dbItem1)
	}
}

func TestDirectWriteOnDeletedRecord(t *testing.T) {
	// TODO:
}

func TestRollbackWhenReading(t *testing.T) {

	item1Pre := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
		MGroupKeyList: "TestRollback",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-9 * time.Second),
		MVersion:      "1",
	}

	item1 := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1")),
		MGroupKeyList: "TestRollback",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now().Add(-5 * time.Second),
		MTLease:       time.Now().Add(-4 * time.Second),
		MVersion:      "2",
	}

	t.Run("rollback an item with a valid Prev field when reading", func(t *testing.T) {
		conn := NewDefaultConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev(util.ToJSONString(item1Pre))
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("mongo", todoRedisItem.Key(), &item)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre", item.Value)
	})

	t.Run("rollback an item with an invalid Prev field when reading", func(t *testing.T) {
		conn := NewDefaultConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev("a broken prev field")
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("mongo", todoRedisItem.Key(), &item)
		assert.NotNil(t, err)
	})

	t.Run("rollback an item with an empty Prev field when reading", func(t *testing.T) {
		conn := NewDefaultConnection()
		item1.SetPrev("")
		conn.PutItem(item1.Key(), item1)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var item testutil.TestItem
		err := txn1.Read("mongo", item1.Key(), &item)
		assert.EqualError(t, err, trxn.KeyNotFound.Error())
	})
}

func TestRollbackWhenWriting(t *testing.T) {
	item1Pre := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
		MGroupKeyList: "TestRollback",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-9 * time.Second),
		MVersion:      "1",
	}

	item1 := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1")),
		MGroupKeyList: "TestRollback",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now().Add(-5 * time.Second),
		MTLease:       time.Now().Add(-4 * time.Second),
		MVersion:      "2",
	}

	t.Run("rollback an item with a valid Prev field when writing", func(t *testing.T) {
		conn := NewDefaultConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev(util.ToJSONString(item1Pre))
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		item := testutil.NewTestItem("item1-cur")
		txn.Write("mongo", todoRedisItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		res, err := conn.GetItem(todoRedisItem.Key())
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(item), res.Value())
		tarItem := item1Pre
		tarItem.SetVersion("3")
		assert.Equal(t, util.ToJSONString(tarItem), res.Prev())

	})

	t.Run("rollback an item with an invalid Prev field", func(t *testing.T) {
		conn := NewDefaultConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev("a broken prev field")
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		item := testutil.NewTestItem("item1-cur")
		txn.Write("mongo", todoRedisItem.Key(), item)
		err := txn.Commit()
		assert.NotNil(t, err)
	})

	t.Run("rollback an item with an empty Prev field when writing", func(t *testing.T) {
		conn := NewDefaultConnection()
		item1.SetPrev("")
		item1.SetLinkedLen(1)
		conn.PutItem(item1.Key(), item1)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		item := testutil.NewTestItem("item1-cur")
		txn1.Write("mongo", item1.Key(), item)
		err := txn1.Commit()
		assert.NoError(t, err)

		res, err := conn.GetItem(item1.Key())
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(item), res.Value())

		item1.SetIsDeleted(true)
		item1.SetTxnState(config.COMMITTED)
		item1.SetVersion(util.AddToString(item1.MVersion, 1))
		assert.Equal(t, util.ToJSONString(item1), res.Prev())
	})
}

func TestRollForwardWhenReading(t *testing.T) {

	conn := NewDefaultConnection()

	tarItem := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
		MGroupKeyList: "TestRollForward",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-9 * time.Second),
		MVersion:      "1",
	}

	conn.PutItem(tarItem.Key(), tarItem)
	conn.Put("TestRollForward", config.COMMITTED)

	// the transaction should roll forward the item
	txn := NewTransactionWithSetup()
	txn.Start()
	var item testutil.TestItem
	err := txn.Read("mongo", tarItem.Key(), &item)
	assert.NoError(t, err)

	res, err := conn.GetItem(tarItem.Key())
	assert.NoError(t, err)
	tarItem.SetTxnState(config.COMMITTED)
	tarItem.SetVersion("2")
	if !res.Equal(tarItem) {
		t.Errorf("\ngot\n %v \nwant\n %v", res, tarItem)
	}
}

func TestRollForwardWhenWriting(t *testing.T) {

	conn := NewDefaultConnection()

	tarItem := &MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
		MGroupKeyList: "TestRollForward",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now().Add(-10 * time.Second),
		MTLease:       time.Now().Add(-9 * time.Second),
		MVersion:      "1",
	}

	conn.PutItem(tarItem.Key(), tarItem)
	conn.Put("TestRollForward", config.COMMITTED)

	// the transaction should roll forward the item
	txn := NewTransactionWithSetup()
	txn.Start()
	item := testutil.NewTestItem("item1-cur")
	txn.Write("mongo", tarItem.Key(), item)
	err := txn.Commit()
	assert.NoError(t, err)

	res, err := conn.GetItem(tarItem.Key())
	assert.NoError(t, err)
	assert.Equal(t, util.ToJSONString(item), res.Value())
	tarItem.SetTxnState(config.COMMITTED)
	tarItem.SetVersion("2")
	tarPrev := util.ToJSONString(tarItem)
	assert.Equal(t, tarPrev, res.Prev())
}

func TestItemVersionUpdate(t *testing.T) {

	t.Run("item version ++ after updated", func(t *testing.T) {
		conn := NewDefaultConnection()
		dbItem := &MongoItem{
			MKey:          "item1",
			MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
			MGroupKeyList: "TestItemVersionUpdate",
			MTxnState:     config.COMMITTED,
			MTValid:       time.Now().Add(-10 * time.Second),
			MTLease:       time.Now().Add(-9 * time.Second),
			MLinkedLen:    1,
			MVersion:      "1",
		}
		conn.PutItem(dbItem.Key(), dbItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("mongo", dbItem.Key(), &item)
		assert.NoError(t, err)
		item.Value = "item1-cur"
		txn.Write("mongo", dbItem.Key(), item)
		err = txn.Commit()
		assert.NoError(t, err)

		res, err := conn.GetItem(dbItem.Key())
		assert.NoError(t, err)
		assert.Equal(t, util.AddToString(dbItem.Version(), 2), res.Version())
	})
}
