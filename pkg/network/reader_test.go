package network

// Please make sure the stateless component
// is running on port 8000 before running this test

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	trxn "github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

// NewDefaultRedisConnection creates and connect a new RedisConnection with default connection options.
// It uses the localhost address and the default Redis port (6379).
func NewDefaultRedisConnection() *redis.RedisConnection {
	conn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address: "localhost:6380",
	})
	conn.Connect()
	return conn
}

// NewTransactionWithSetup creates a new transaction with a Redis datastore setup.
// It initializes a Redis connection, creates a new transaction, and adds a Redis datastore to the transaction.
// The Redis connection is established with the provided address.
// The created transaction is returned.
func NewTransactionWithSetup() *trxn.Transaction {
	conn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address: "localhost:6380",
	})
	client := NewClient("localhost:8000")
	txn := trxn.NewTransactionWithRemote(client)
	rds := redis.NewRedisDatastore("redis", conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)
	return txn
}

func TestSimpleReadWhenCommitted(t *testing.T) {

	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "123123",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-5 * time.Second),
		RVersion:  "2",
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
	err = txn.Read("redis", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenCommittedFindEmpty(t *testing.T) {

	txn1 := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "TestSimpleReadWhenCommittedFindEmpty",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(+10 * time.Second),
		RTLease:   time.Now().Add(+5 * time.Second),
		RVersion:  "2",
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
	err = txn1.Read("redis", key, &result)
	assert.EqualError(t, err, trxn.KeyNotFound.Error())

}

func TestSimpleReadWhenCommittedFindPrevious(t *testing.T) {

	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	preRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "99",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-5 * time.Second),
		RVersion:  "1",
	}
	curRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(curPerson),
		RTxnId:    "100",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(10 * time.Second),
		RTLease:   time.Now().Add(5 * time.Second),
		RVersion:  "2",
		RPrev:     util.ToJSONString(preRedisItem),
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
	err = txn.Read("redis", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenCommittedFindNone(t *testing.T) {
	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	preRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "99",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(10 * time.Second),
		RTLease:   time.Now().Add(5 * time.Second),
		RVersion:  "1",
	}
	curRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(curPerson),
		RTxnId:    "100",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(20 * time.Second),
		RTLease:   time.Now().Add(15 * time.Second),
		RVersion:  "2",
		RPrev:     util.ToJSONString(preRedisItem),
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
	err = txn.Read("redis", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from redis datastore: %s", err)
	}
}

// TestSimpleReadWhenPreparedWithTSRInCOMMITTED tests the scenario where a simple read operation is performed
// on a record which is in PREPARED state and has a TSR in COMMITTED state.
func TestSimpleReadWhenPreparedWithTSRInCOMMITTED(t *testing.T) {
	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "100",
		RTxnState: config.PREPARED,
		RTValid:   time.Now(),
		RTLease:   time.Now(),
		RVersion:  "2",
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
	err = txn.Read("redis", key, &result)
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
	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	tarMemItem := &redis.RedisItem{
		RKey:      "item1",
		RValue:    util.ToJSONString(testutil.NewTestItem("item1")),
		RTxnId:    "99",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-9 * time.Second),
		RVersion:  "1",
	}

	curMemItem := &redis.RedisItem{
		RKey:      "item1",
		RValue:    util.ToJSONString(testutil.NewTestItem("item1-prepared")),
		RTxnId:    "TestSimpleReadWhenPreparedWithTSRInABORTED",
		RTxnState: config.PREPARED,
		RTValid:   time.Now().Add(-5 * time.Second),
		RTLease:   time.Now().Add(-4 * time.Second),
		RPrev:     util.ToJSONString(tarMemItem),
		RVersion:  "2",
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
	err = txn.Read("redis", key, &result)
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
	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "100",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-5 * time.Second),
		RVersion:  "2",
	}

	expectedStr := util.ToJSONString(expectedRedisItem)

	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}

	curRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(curPerson),
		RTxnId:    "101",
		RTxnState: config.PREPARED,
		RTValid:   time.Now().Add(-3 * time.Second),
		RTLease:   time.Now().Add(-1 * time.Second),
		RVersion:  "3",
		RPrev:     expectedStr,
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
	err = txn.Read("redis", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	if result != expected {
		t.Errorf("got %v want %v", result, expected)
	}
}

func TestSimpleReadWhenPrepareNotExpired(t *testing.T) {

	dbItem1 := &redis.RedisItem{
		RKey:       "item1",
		RValue:     util.ToJSONString(testutil.NewTestItem("item1-pre1")),
		RTxnId:     "TestSimpleReadWhenPrepareNotExpired1",
		RTxnState:  config.COMMITTED,
		RTValid:    time.Now().Add(-2 * time.Second),
		RTLease:    time.Now().Add(-1 * time.Second),
		RLinkedLen: 1,
		RVersion:   "1",
	}

	dbItem2 := &redis.RedisItem{
		RKey:       "item1",
		RValue:     util.ToJSONString(testutil.NewTestItem("item1-pre2")),
		RTxnId:     "TestSimpleReadWhenPrepareNotExpired2",
		RTxnState:  config.PREPARED,
		RTValid:    time.Now().Add(1 * time.Second),
		RTLease:    time.Now().Add(2 * time.Second),
		RPrev:      util.ToJSONString(dbItem1),
		RLinkedLen: 2,
		RVersion:   "2",
	}

	t.Run("when the item has a valid Prev field", func(t *testing.T) {
		conn := NewDefaultRedisConnection()
		conn.PutItem("item1", dbItem2)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var item testutil.TestItem
		err := txn1.Read("redis", "item1", &item)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre1", item.Value)
	})

	t.Run("when the item has an empty Prev field", func(t *testing.T) {
		conn := NewDefaultRedisConnection()
		dbItem := dbItem2
		dbItem.SetPrev("")
		dbItem.SetLinkedLen(1)
		conn.PutItem("item1", dbItem)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var item testutil.TestItem
		err := txn1.Read("redis", "item1", &item)
		assert.EqualError(t, err, trxn.KeyNotFound.Error())
	})
}

func TestSimpleReadWhenDeleted(t *testing.T) {
	conn := NewDefaultRedisConnection()
	dbItem := &redis.RedisItem{
		RKey:       "item2",
		RValue:     util.ToJSONString(testutil.NewTestItem("item2-db")),
		RTxnState:  config.COMMITTED,
		RTValid:    time.Now().Add(-2 * time.Second),
		RTLease:    time.Now().Add(-1 * time.Second),
		RLinkedLen: 1,
		RVersion:   "1",
		RIsDeleted: true,
	}

	conn.PutItem(dbItem.Key(), dbItem)

	txn1 := NewTransactionWithSetup()
	txn1.Start()

	var item testutil.TestItem
	err := txn1.Read("redis", "item2", &item)
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
	err = txn.Write("redis", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("redis", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleDirectWrite(t *testing.T) {

	conn := NewDefaultRedisConnection()
	conn.Delete("John")

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	key := "John"
	prePerson := testutil.NewPerson("John-pre")
	preTxn.Write("redis", key, prePerson)
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
	err = txn.Write("redis", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}
	err = txn.Commit()
	assert.NoError(t, err)
}

func TestSimpleWriteAndReadLocal(t *testing.T) {
	txn := NewTransactionWithSetup()

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
	err = txn.Write("redis", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("redis", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleReadModifyWriteThenRead(t *testing.T) {
	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "123123",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-5 * time.Second),
		RVersion:  "2",
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
	err = txn.Read("redis", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Modify the value
	result.Age = 31

	// Write the value
	err = txn.Write("redis", key, result)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result2 testutil.Person
	err = txn.Read("redis", key, &result2)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result2 != result {
		t.Errorf("got %v want %v", result2, result)
	}
}

func TestSimpleOverwriteAndRead(t *testing.T) {
	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "123123",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-5 * time.Second),
		RVersion:  "2",
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
	err = txn.Write("redis", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}
	person.Age = 32
	err = txn.Write("redis", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("redis", key, &result)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	// Check the result
	if result != person {
		t.Errorf("got %v want %v", result, person)
	}
}

func TestSimpleDeleteAndRead(t *testing.T) {
	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "123123",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-5 * time.Second),
		RVersion:  "2",
	}

	key := "John"
	conn.PutItem(key, expectedRedisItem)

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Delete the value
	err = txn.Delete("redis", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("redis", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from redis datastore: %s", err)
	}
}

func TestSimpleDeleteTwice(t *testing.T) {
	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()
	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "123123",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-5 * time.Second),
		RVersion:  "2",
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
	err = txn.Delete("redis", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}
	err = txn.Delete("redis", key)
	if err.Error() != "key not found" {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}
}

func TestDeleteWithRead(t *testing.T) {

	conn := NewDefaultRedisConnection()
	// clear the test data
	conn.Delete("John")

	preTxn := NewTransactionWithSetup()
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("redis", "John", dataPerson)
	err := preTxn.Commit()
	assert.NoError(t, err)

	txn := NewTransactionWithSetup()
	txn.Start()
	var person testutil.Person
	err = txn.Read("redis", "John", &person)
	assert.NoError(t, err)
	err = txn.Delete("redis", "John")
	assert.NoError(t, err)

	err = txn.Commit()
	assert.NoError(t, err)
}

func TestDeleteWithoutRead(t *testing.T) {

	preTxn := NewTransactionWithSetup()
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("redis", "John", dataPerson)
	err := preTxn.Commit()
	assert.NoError(t, err)

	txn := NewTransactionWithSetup()
	txn.Start()
	err = txn.Delete("redis", "John")
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	err = txn.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

}

func TestSimpleReadWriteDeleteThenRead(t *testing.T) {
	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()
	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "123123",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-5 * time.Second),
		RVersion:  "2",
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
	err = txn.Read("redis", key, &person)
	if err != nil {
		t.Errorf("Error reading from redis datastore: %s", err)
	}

	person.Age = 31

	// Write the value
	err = txn.Write("redis", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Delete the value
	err = txn.Delete("redis", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("redis", key, &result)
	if err.Error() != errors.New("key not found").Error() {
		t.Errorf("Error reading from redis datastore: %s", err)
	}
}

func TestSimpleWriteDeleteWriteThenRead(t *testing.T) {
	txn := NewTransactionWithSetup()
	conn := NewDefaultRedisConnection()

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := &redis.RedisItem{
		RKey:      "John",
		RValue:    util.ToJSONString(expected),
		RTxnId:    "123123",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-5 * time.Second),
		RVersion:  "2",
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
	err = txn.Write("redis", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Delete the value
	err = txn.Delete("redis", key)
	if err != nil {
		t.Errorf("Error deleting from redis datastore: %s", err)
	}

	// Write the value
	person.Age = 32
	err = txn.Write("redis", key, person)
	if err != nil {
		t.Errorf("Error writing to redis datastore: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("redis", key, &result)
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
	conn := NewDefaultRedisConnection()
	for _, item := range testutil.InputItemList {
		conn.Delete(item.Value)
	}

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write("redis", item.Value, item)
	}
	err := preTxn.Commit()
	assert.NoError(t, err)

	resChan := make(chan bool)
	successId := 0

	concurrentCount := 1000
	client := NewClient("localhost:8000")
	txn := trxn.NewTransactionWithRemote(client)
	rds := redis.NewRedisDatastore("redis", conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	for i := 1; i <= concurrentCount; i++ {
		go func(id int) {
			txn := trxn.NewTransactionWithRemote(client)
			rds := redis.NewRedisDatastore("redis", conn)
			txn.AddDatastore(rds)
			txn.SetGlobalDatastore(rds)

			txn.Start()
			for _, item := range testutil.InputItemList {
				var res testutil.TestItem
				txn.Read("redis", item.Value, &res)
				res.Value = item.Value + "-new-" + strconv.Itoa(id)
				txn.Write("redis", item.Value, res)
			}

			time.Sleep(100 * time.Millisecond)
			err := txn.Commit()
			if err != nil {
				if err.Error() != "prepare phase failed: write conflicted: the record is in PREPARED state" &&
					err.Error() != "prepare phase failed: write conflicted: the record has been modified by others" &&
					err.Error() != "prepare phase failed: Remote prepare failed\nversion mismatch" {
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

	if commitCount != 0 && commitCount != 1 {
		t.Errorf("commitCount: %d", commitCount)
	}

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	for _, item := range testutil.InputItemList {
		var res testutil.TestItem
		postTxn.Read("redis", item.Value, &res)
		if commitCount == 1 {
			assert.Equal(t, item.Value+"-new-"+strconv.Itoa(successId), res.Value)
		} else {
			// the item supposes to be unchanged
			assert.Equal(t, item.Value, res.Value)
		}
	}
	err = postTxn.Commit()
	assert.NoError(t, err)

}

func TestTxnWriteMultiRecord(t *testing.T) {

	// clear the test data
	conn := NewDefaultRedisConnection()
	conn.Delete("item1")
	conn.Delete("item2")

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	preTxn.Write("redis", "item1", testutil.NewTestItem("item1"))
	preTxn.Write("redis", "item2", testutil.NewTestItem("item2"))
	err := preTxn.Commit()
	assert.Nil(t, err)

	txn := NewTransactionWithSetup()
	txn.Start()
	var item testutil.TestItem
	txn.Read("redis", "item1", &item)
	item.Value = "item1_new"
	txn.Write("redis", "item1", item)

	txn.Read("redis", "item2", &item)
	item.Value = "item2_new"
	txn.Write("redis", "item2", item)

	err = txn.Commit()
	assert.Nil(t, err)

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	var resItem testutil.TestItem
	postTxn.Read("redis", "item1", &resItem)
	assert.Equal(t, "item1_new", resItem.Value)
	postTxn.Read("redis", "item2", &resItem)
	assert.Equal(t, "item2_new", resItem.Value)

}

// ---|---------|--------|---------|------> time
// item1_1  T_Start   item1_2   item1_3
func TestLinkedReadAsCommitted(t *testing.T) {

	item1_1 := testutil.NewTestItem("item1_1")
	memItem1_1 := &redis.RedisItem{
		RKey:       "item1",
		RValue:     util.ToJSONString(item1_1),
		RTxnId:     "txn1",
		RTxnState:  config.COMMITTED,
		RTValid:    time.Now().Add(-10 * time.Second),
		RTLease:    time.Now().Add(-9 * time.Second),
		RVersion:   "1",
		RLinkedLen: 1,
	}

	item1_2 := testutil.NewTestItem("item1_2")
	memItem1_2 := &redis.RedisItem{
		RKey:       "item1",
		RValue:     util.ToJSONString(item1_2),
		RTxnId:     "txn2",
		RTxnState:  config.COMMITTED,
		RTValid:    time.Now().Add(5 * time.Second),
		RTLease:    time.Now().Add(6 * time.Second),
		RVersion:   "2",
		RPrev:      util.ToJSONString(memItem1_1),
		RLinkedLen: 2,
	}

	item1_3 := testutil.NewTestItem("item1_3")
	memItem1_3 := &redis.RedisItem{
		RKey:       "item1",
		RValue:     util.ToJSONString(item1_3),
		RTxnId:     "txn3",
		RTxnState:  config.COMMITTED,
		RTValid:    time.Now().Add(10 * time.Second),
		RTLease:    time.Now().Add(11 * time.Second),
		RVersion:   "3",
		RPrev:      util.ToJSONString(memItem1_2),
		RLinkedLen: 3,
	}

	t.Run("read will fail due to MaxRecordLength=2", func(t *testing.T) {

		txn := NewTransactionWithSetup()
		conn := NewDefaultRedisConnection()
		_, err := conn.PutItem("item1", memItem1_3)
		assert.NoError(t, err)

		config.Config.MaxRecordLength = 2
		txn.Start()
		var item testutil.TestItem
		err = txn.Read("redis", "item1", &item)
		assert.EqualError(t, err, "key not found")
	})

	t.Run("read will success due to MaxRecordLength=3", func(t *testing.T) {

		txn := NewTransactionWithSetup()
		conn := NewDefaultRedisConnection()
		conn.PutItem("item1", memItem1_3)

		config.Config.MaxRecordLength = 3
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("redis", "item1", &item)
		assert.Nil(t, err)

		assert.Equal(t, "item1_1", item.Value)
	})

	t.Run("read will success due to MaxRecordLength > 3", func(t *testing.T) {

		txn := NewTransactionWithSetup()
		conn := NewDefaultRedisConnection()
		conn.PutItem("item1", memItem1_3)

		config.Config.MaxRecordLength = 3 + 1
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("redis", "item1", &item)
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

		conn := NewDefaultRedisConnection()
		conn.Delete("item1")

		for i := 1; i <= 4; i++ {
			time.Sleep(10 * time.Millisecond)
			item := testutil.NewTestItem("item1_" + strconv.Itoa(i))
			txn := NewTransactionWithSetup()
			txn.Start()
			txn.Write("redis", "item1", item)
			err := txn.Commit()
			assert.Nil(t, err)
		}

		// check the linked record length
		item, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, config.Config.MaxRecordLength, item.LinkedLen())

		tarItem := item
		for i := 1; i <= config.Config.MaxRecordLength-1; i++ {
			var preItem redis.RedisItem
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
				txn.Write("redis", "item1", item)
				err := txn.Commit()
				assert.Nil(t, err)
			}

			// check the linked record length
			conn := NewDefaultRedisConnection()
			item, err := conn.GetItem("item1")
			assert.NoError(t, err)
			assert.Equal(t, config.Config.MaxRecordLength, item.LinkedLen())
			t.Logf("item: %+v", item)

			tarItem := item
			for i := 1; i <= config.Config.MaxRecordLength-1; i++ {
				var preItem redis.RedisItem
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
			txn.Write("redis", "item1", item)
			err := txn.Commit()
			assert.Nil(t, err)
		}

		// check the linked record length
		conn := NewDefaultRedisConnection()
		item, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, expectedLen, item.LinkedLen())

		tarItem := item
		for i := 1; i <= expectedLen-1; i++ {
			var preItem redis.RedisItem
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
		conn := NewDefaultRedisConnection()

		tarItem := &redis.RedisItem{
			RKey:       "item1",
			RValue:     util.ToJSONString(testutil.NewTestItem("item1-pre2")),
			RTxnId:     "99",
			RTxnState:  config.COMMITTED,
			RTValid:    time.Now().Add(-10 * time.Second),
			RTLease:    time.Now().Add(-9 * time.Second),
			RLinkedLen: 1,
			RVersion:   "1",
		}

		curItem := &redis.RedisItem{
			RKey:       "item1",
			RValue:     util.ToJSONString(testutil.NewTestItem("item1-pre")),
			RTxnId:     "100",
			RTxnState:  config.PREPARED,
			RTValid:    time.Now().Add(-5 * time.Second),
			RTLease:    time.Now().Add(-4 * time.Second),
			RPrev:      util.ToJSONString(tarItem),
			RLinkedLen: 2,
			RVersion:   "2",
		}

		conn.PutItem(tarItem.Key(), curItem)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item1-cur")
		txn.Write("redis", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("redis", "item1", &resItem)
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
		conn := NewDefaultRedisConnection()

		tarItem := &redis.RedisItem{
			RKey:      "item1",
			RValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
			RTxnId:    "99",
			RTxnState: config.PREPARED,
			RTValid:   time.Now().Add(-10 * time.Second),
			RTLease:   time.Now().Add(-9 * time.Second),
			RVersion:  "1",
		}

		conn.PutItem(tarItem.Key(), tarItem)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item1-cur")
		txn.Write("redis", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("redis", "item1", &resItem)
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
		conn := NewDefaultRedisConnection()

		tarItem := &redis.RedisItem{
			RKey:       "item2",
			RValue:     util.ToJSONString(testutil.NewTestItem("item2-pre2")),
			RTxnId:     "TestDirectWriteOnOutdatedPreparedRecordWithTSR2",
			RTxnState:  config.COMMITTED,
			RTValid:    time.Now().Add(-10 * time.Second),
			RTLease:    time.Now().Add(-9 * time.Second),
			RLinkedLen: 1,
			RVersion:   "1",
		}

		curItem := &redis.RedisItem{
			RKey:       "item2",
			RValue:     util.ToJSONString(testutil.NewTestItem("item2-pre")),
			RTxnId:     "TestDirectWriteOnOutdatedPreparedRecordWithTSR",
			RTxnState:  config.PREPARED,
			RTValid:    time.Now().Add(-5 * time.Second),
			RTLease:    time.Now().Add(-4 * time.Second),
			RLinkedLen: 2,
			RVersion:   "2",
			RPrev:      util.ToJSONString(tarItem),
		}

		conn.PutItem(curItem.Key(), curItem)
		conn.Put("TestDirectWriteOnOutdatedPreparedRecordWithTSR", config.COMMITTED)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item2-cur")
		txn.Write("redis", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("redis", "item2", &resItem)
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
		conn := NewDefaultRedisConnection()

		tarItem := &redis.RedisItem{
			RKey:      "item1",
			RValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
			RTxnId:    "TestDirectWriteOnOutdatedPreparedRecordWithTSR",
			RTxnState: config.PREPARED,
			RTValid:   time.Now().Add(-10 * time.Second),
			RTLease:   time.Now().Add(-9 * time.Second),
			RVersion:  "1",
		}

		conn.PutItem(tarItem.Key(), tarItem)
		conn.Put("TestDirectWriteOnOutdatedPreparedRecordWithTSR", config.COMMITTED)

		// Start the transaction
		txn := NewTransactionWithSetup()
		txn.Start()

		// Write the value
		item := testutil.NewTestItem("item1-cur")
		txn.Write("redis", tarItem.Key(), item)
		err := txn.Commit()
		assert.NoError(t, err)

		// First, we check the final record's Value
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var resItem testutil.TestItem
		err = postTxn.Read("redis", "item1", &resItem)
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

	conn := NewDefaultRedisConnection()

	tarItem := &redis.RedisItem{
		RKey:      "item1",
		RValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		RTxnId:    "TestDirectWriteOnPreparingRecord",
		RTxnState: config.PREPARED,
		RTValid:   time.Now().Add(2 * time.Second),
		RTLease:   time.Now().Add(1 * time.Second),
		RVersion:  "1",
	}

	conn.PutItem(tarItem.Key(), tarItem)

	txn := NewTransactionWithSetup()
	txn.Start()
	item := testutil.NewTestItem("item1-cur")
	txn.Write("redis", tarItem.Key(), item)
	err := txn.Commit()
	assert.EqualError(t, err,
		"prepare phase failed: Remote prepare failed\nversion mismatch")
}

func TestDirectWriteOnInvisibleRecord(t *testing.T) {
	conn := NewDefaultRedisConnection()

	dbItem1 := &redis.RedisItem{
		RKey:       "item1",
		RValue:     util.ToJSONString(testutil.NewTestItem("item1-pre1")),
		RTxnId:     "TestDirectWriteOnInvisibleRecord1",
		RTxnState:  config.COMMITTED,
		RTValid:    time.Now().Add(3 * time.Second),
		RTLease:    time.Now().Add(4 * time.Second),
		RLinkedLen: 1,
		RVersion:   "2",
	}
	conn.PutItem(dbItem1.Key(), dbItem1)

	txn := NewTransactionWithSetup()
	txn.Start()
	item := testutil.NewTestItem("item1-cur")
	txn.Write("redis", dbItem1.Key(), item)
	err := txn.Commit()
	assert.EqualError(t, err,
		"prepare phase failed: Remote prepare failed\nversion mismatch")

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

	item1Pre := &redis.RedisItem{
		RKey:      "item1",
		RValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		RTxnId:    "TestRollback",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-9 * time.Second),
		RVersion:  "1",
	}

	item1 := &redis.RedisItem{
		RKey:      "item1",
		RValue:    util.ToJSONString(testutil.NewTestItem("item1")),
		RTxnId:    "TestRollback",
		RTxnState: config.PREPARED,
		RTValid:   time.Now().Add(-5 * time.Second),
		RTLease:   time.Now().Add(-4 * time.Second),
		RVersion:  "2",
	}

	t.Run("rollback an item with a valid Prev field when reading", func(t *testing.T) {
		conn := NewDefaultRedisConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev(util.ToJSONString(item1Pre))
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("redis", todoRedisItem.Key(), &item)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre", item.Value)
	})

	t.Run("rollback an item with an invalid Prev field when reading", func(t *testing.T) {
		conn := NewDefaultRedisConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev("a broken prev field")
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("redis", todoRedisItem.Key(), &item)
		assert.NotNil(t, err)
	})

	t.Run("rollback an item with an empty Prev field when reading", func(t *testing.T) {
		conn := NewDefaultRedisConnection()
		item1.SetPrev("")
		conn.PutItem(item1.Key(), item1)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var item testutil.TestItem
		err := txn1.Read("redis", item1.Key(), &item)
		assert.EqualError(t, err, trxn.KeyNotFound.Error())
	})
}

func TestRollbackWhenWriting(t *testing.T) {
	item1Pre := &redis.RedisItem{
		RKey:      "item1",
		RValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		RTxnId:    "TestRollback",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-9 * time.Second),
		RVersion:  "1",
	}

	item1 := &redis.RedisItem{
		RKey:      "item1",
		RValue:    util.ToJSONString(testutil.NewTestItem("item1")),
		RTxnId:    "TestRollback",
		RTxnState: config.PREPARED,
		RTValid:   time.Now().Add(-5 * time.Second),
		RTLease:   time.Now().Add(-4 * time.Second),
		RVersion:  "2",
	}

	t.Run("rollback an item with a valid Prev field when writing", func(t *testing.T) {
		conn := NewDefaultRedisConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev(util.ToJSONString(item1Pre))
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		item := testutil.NewTestItem("item1-cur")
		txn.Write("redis", todoRedisItem.Key(), item)
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
		conn := NewDefaultRedisConnection()
		todoRedisItem := item1
		todoRedisItem.SetPrev("a broken prev field")
		// TODO: need a helper func
		todoRedisItem.SetLinkedLen(2)
		conn.PutItem(todoRedisItem.Key(), todoRedisItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		item := testutil.NewTestItem("item1-cur")
		txn.Write("redis", todoRedisItem.Key(), item)
		err := txn.Commit()
		assert.NotNil(t, err)
	})

	t.Run("rollback an item with an empty Prev field when writing", func(t *testing.T) {
		conn := NewDefaultRedisConnection()
		item1.SetPrev("")
		item1.SetLinkedLen(1)
		conn.PutItem(item1.Key(), item1)

		txn1 := NewTransactionWithSetup()
		txn1.Start()
		item := testutil.NewTestItem("item1-cur")
		txn1.Write("redis", item1.Key(), item)
		err := txn1.Commit()
		assert.NoError(t, err)

		res, err := conn.GetItem(item1.Key())
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(item), res.Value())

		item1.SetIsDeleted(true)
		item1.SetTxnState(config.COMMITTED)
		item1.SetVersion(util.AddToString(item1.Version(), 1))
		assert.Equal(t, util.ToJSONString(item1), res.Prev())
	})
}

func TestRollForwardWhenReading(t *testing.T) {

	conn := NewDefaultRedisConnection()

	tarItem := &redis.RedisItem{
		RKey:      "item1",
		RValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		RTxnId:    "TestRollForward",
		RTxnState: config.PREPARED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-9 * time.Second),
		RVersion:  "1",
	}

	conn.PutItem(tarItem.Key(), tarItem)
	conn.Put("TestRollForward", config.COMMITTED)

	// the transaction should roll forward the item
	txn := NewTransactionWithSetup()
	txn.Start()
	var item testutil.TestItem
	err := txn.Read("redis", tarItem.Key(), &item)
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

	conn := NewDefaultRedisConnection()

	tarItem := &redis.RedisItem{
		RKey:      "item1",
		RValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		RTxnId:    "TestRollForward",
		RTxnState: config.PREPARED,
		RTValid:   time.Now().Add(-10 * time.Second),
		RTLease:   time.Now().Add(-9 * time.Second),
		RVersion:  "1",
	}

	conn.PutItem(tarItem.Key(), tarItem)
	conn.Put("TestRollForward", config.COMMITTED)

	// the transaction should roll forward the item
	txn := NewTransactionWithSetup()
	txn.Start()
	item := testutil.NewTestItem("item1-cur")
	txn.Write("redis", tarItem.Key(), item)
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
		conn := NewDefaultRedisConnection()
		dbItem := &redis.RedisItem{
			RKey:       "item1",
			RValue:     util.ToJSONString(testutil.NewTestItem("item1-pre")),
			RTxnId:     "TestItemVersionUpdate",
			RTxnState:  config.COMMITTED,
			RTValid:    time.Now().Add(-10 * time.Second),
			RTLease:    time.Now().Add(-9 * time.Second),
			RLinkedLen: 1,
			RVersion:   "1",
		}
		conn.PutItem(dbItem.Key(), dbItem)

		txn := NewTransactionWithSetup()
		txn.Start()
		var item testutil.TestItem
		err := txn.Read("redis", dbItem.Key(), &item)
		assert.NoError(t, err)
		item.Value = "item1-cur"
		txn.Write("redis", dbItem.Key(), item)
		err = txn.Commit()
		assert.NoError(t, err)

		res, err := conn.GetItem(dbItem.Key())
		assert.NoError(t, err)
		assert.Equal(t, util.AddToString(dbItem.Version(), 2), res.Version())
	})
}
