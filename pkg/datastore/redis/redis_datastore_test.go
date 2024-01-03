package redis

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

func NewDefaultRedisConnection() *RedisConnection {
	return NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
}

func NewTransactionWithSetup() *txn.Transaction {
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	txn := txn.NewTransaction()
	rds := NewRedisDatastore("redis", conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)
	return txn
}

func TestSimpleReadInCache(t *testing.T) {
	txn := txn.NewTransaction()
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	memoryPerson := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
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
	err := conn.PutItem(key, expectedRedisItem)
	if err != nil {
		t.Errorf("Error putting item to redis datastore: %s", err)
	}

	// Put a item in cache
	cachePerson := testutil.Person{
		Name: "John",
		Age:  31,
	}
	cacheRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(cachePerson),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  2,
	}
	rds.writeCache[key] = cacheRedisItem

	// Start the transaction
	err = txn.Start()
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
	if result != cachePerson {
		t.Errorf("got %v want %v", result, cachePerson)
	}
}

func TestSimpleReadWhenCommitted(t *testing.T) {

	txn := txn.NewTransaction()
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
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

func TestSimpleReadWhenCommittedFindPrevious(t *testing.T) {
	// Create a new transaction
	txn := txn.NewTransaction()

	// Create a new redis datastore
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
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
	preRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "99",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  1,
	}
	curRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(curPerson),
		TxnId:    "100",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  2,
		Prev:     util.ToJSONString(preRedisItem),
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
	// Create a new redis datastore
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
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
	preRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "99",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  1,
	}
	curRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(curPerson),
		TxnId:    "100",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(20 * time.Second),
		TLease:   time.Now().Add(15 * time.Second),
		Version:  2,
		Prev:     util.ToJSONString(preRedisItem),
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
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "100",
		TxnState: config.PREPARED,
		TValid:   time.Now(),
		TLease:   time.Now(),
		Version:  2,
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
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	tarMemItem := RedisItem{
		Key:      "item1",
		Value:    util.ToJSONString(testutil.NewTestItem("item1")),
		TxnId:    "99",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-9 * time.Second),
		Version:  1,
	}

	curMemItem := RedisItem{
		Key:      "item1",
		Value:    util.ToJSONString(testutil.NewTestItem("item1-prepared")),
		TxnId:    "TestSimpleReadWhenPreparedWithTSRInABORTED",
		TxnState: config.PREPARED,
		TValid:   time.Now().Add(-5 * time.Second),
		TLease:   time.Now().Add(-4 * time.Second),
		Prev:     util.ToJSONString(tarMemItem),
		Version:  2,
	}

	key := "John"
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
	// Create a new redis datastore
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "100",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	expectedStr := util.ToJSONString(expectedRedisItem)

	curPerson := testutil.Person{
		Name: "John",
		Age:  31,
	}

	curRedisItem := RedisItem{
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
	// Create a new redis datastore
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "100",
		TxnState: config.PREPARED,
		TValid:   time.Now().Add(10 * time.Second),
		TLease:   time.Now().Add(5 * time.Second),
		Version:  2,
	}

	key := "John"
	err := conn.PutItem(key, expectedRedisItem)
	if err != nil {
		t.Errorf("Error putting item to redis datastore: %s", err)
	}

	// Start the transaction
	err = txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Read the value
	var result testutil.Person
	err = txn.Read("redis", key, &result)
	if err.Error() != errors.New("dirty Read").Error() {
		t.Errorf("Error reading from redis datastore: %s", err)
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
	// Create a new redis datastore
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
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
	// Create a new redis datastore
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
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
	// Create a new redis datastore
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
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
	// Create a new redis datastore
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
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
	// Create a new redis datastore
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
	}

	key := "John"
	err := conn.PutItem(key, expectedRedisItem)
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
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	txn.Start()
	err := txn.Delete("redis", "John")
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
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
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
	// Create a new redis datastore
	conn := NewRedisConnection(&ConnectionOptions{
		Address: "localhost:6379",
	})
	rds := NewRedisDatastore("redis", conn)
	txn := txn.NewTransaction()
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)

	// initialize the redis database
	expected := testutil.Person{
		Name: "John",
		Age:  30,
	}
	expectedRedisItem := RedisItem{
		Key:      "John",
		Value:    util.ToJSONString(expected),
		TxnId:    "123123",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-5 * time.Second),
		Version:  2,
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

	concurrentCount := 100

	for i := 1; i <= concurrentCount; i++ {
		go func(id int) {
			txn := NewTransactionWithSetup()
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
		postTxn.Read("redis", item.Value, &res)
		assert.Equal(t, item.Value+"-new-"+strconv.Itoa(successId), res.Value)
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
