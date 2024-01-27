package integration

import (
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/mock"
	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/factory"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

func TestRedis_TxnWrite(t *testing.T) {

	txn1 := NewTransactionWithSetup(REDIS)
	expected := testutil.NewDefaultPerson()

	// clear the data
	conn := NewConnectionWithSetup(REDIS)
	conn.Delete("John")

	// Txn1 writes the record
	err := txn1.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	err = txn1.Write(REDIS, "John", expected)
	if err != nil {
		t.Errorf("Error writing record: %s", err)
	}
	err = txn1.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

	txn2 := NewTransactionWithSetup(REDIS)

	// Txn2 reads the record
	err = txn2.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var actual testutil.Person
	err = txn2.Read(REDIS, "John", &actual)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}
	err = txn2.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

	// compare
	if actual != expected {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestRedis_ReadOwnWrite(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(REDIS, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(REDIS)

	// Txn reads the record
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var person testutil.Person
	err = txn.Read(REDIS, "John", &person)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}

	expected := person
	expected.Age = 31
	txn.Write(REDIS, "John", expected)

	var actual testutil.Person
	err = txn.Read(REDIS, "John", &actual)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}
	err = txn.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

	if actual != expected {
		t.Errorf("Expected %v, got %v", expected, actual)
	}

}

func TestRedis_SingleKeyWriteConflict(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(REDIS, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(REDIS)
	txn.Start()
	var person testutil.Person
	txn.Read(REDIS, "John", &person)
	person.Age = 31
	txn.Write(REDIS, "John", person)
	var anotherPerson testutil.Person
	txn.Read(REDIS, "John", &anotherPerson)

	if person != anotherPerson {
		t.Errorf("Expected two read to be the same")
	}
	txn.Commit()

	postTxn := NewTransactionWithSetup(REDIS)
	postTxn.Start()
	var postPerson testutil.Person
	postTxn.Read(REDIS, "John", &postPerson)
	if postPerson != person {
		t.Errorf("got %v want %v", postPerson, person)
	}

}

func TestRedis_MultileKeyWriteConflict(t *testing.T) {

	// clear the data
	conn := NewConnectionWithSetup(REDIS)
	conn.Delete("item1")
	conn.Delete("item2")

	preTxn := NewTransactionWithSetup(REDIS)
	item1 := testutil.NewTestItem("item1")
	item2 := testutil.NewTestItem("item2")
	preTxn.Start()
	preTxn.Write(REDIS, "item1", item1)
	preTxn.Write(REDIS, "item2", item2)
	err := preTxn.Commit()
	assert.Nil(t, err)

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(REDIS)
		txn1.Start()
		var item testutil.TestItem
		txn1.Read(REDIS, "item1", &item)
		item.Value = "item1-updated-by-txn1"
		txn1.Write(REDIS, "item1", item)

		txn1.Read(REDIS, "item2", &item)
		item.Value = "item2-updated-by-txn1"
		txn1.Write(REDIS, "item2", item)

		err := txn1.Commit()
		if err != nil {
			t.Logf("txn1 commit err: %s", err)
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup(REDIS)
		txn2.Start()
		var item testutil.TestItem
		txn2.Read(REDIS, "item2", &item)
		item.Value = "item2-updated-by-txn2"
		txn2.Write(REDIS, "item2", item)

		txn2.Read(REDIS, "item1", &item)
		item.Value = "item1-updated-by-txn2"
		txn2.Write(REDIS, "item1", item)

		err := txn2.Commit()
		if err != nil {
			t.Logf("txn2 commit err: %s", err)
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	res1 := <-resChan
	res2 := <-resChan

	// only one transaction should succeed
	if res1 == res2 {
		t.Logf("res1: %v, res2: %v", res1, res2)
		t.Errorf("Expected only one transaction to succeed")
		postTxn := NewTransactionWithSetup(REDIS)
		postTxn.Start()
		var item testutil.TestItem
		postTxn.Read(REDIS, "item1", &item)
		t.Logf("item1: %v", item)
		postTxn.Read(REDIS, "item2", &item)
		t.Logf("item2: %v", item)
		postTxn.Commit()
	}
}

func TestRedis_RepeatableReadWhenRecordDeleted(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(REDIS, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(REDIS)
	manualTxn := NewTransactionWithSetup(REDIS)
	txn.Start()
	manualTxn.Start()

	var person1 testutil.Person
	txn.Read(REDIS, "John", &person1)

	// manualTxn deletes John and commits
	manualTxn.Delete("memory", "John")
	manualTxn.Commit()

	var person2 testutil.Person
	txn.Read(REDIS, "John", &person2)

	// two read in txn should be the same
	if person1 != person2 {
		t.Errorf("Expected two read in txn to be the same")
	}
}

func TestRedis_RepeatableReadWhenRecordUpdatedTwice(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(REDIS, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(REDIS)
	manualTxn1 := NewTransactionWithSetup(REDIS)
	txn.Start()
	manualTxn1.Start()

	var person1 testutil.Person
	txn.Read(REDIS, "John", &person1)

	// manualTxn1 updates John and commits
	var manualPerson1 testutil.Person
	manualTxn1.Read(REDIS, "John", &manualPerson1)
	manualPerson1.Age = 31
	manualTxn1.Write(REDIS, "John", manualPerson1)
	manualTxn1.Commit()

	manualTxn2 := NewTransactionWithSetup(REDIS)
	manualTxn2.Start()
	// manualTxn updates John again and commits
	var manualPerson2 testutil.Person
	manualTxn2.Read(REDIS, "John", &manualPerson2)
	manualPerson2.Age = 32
	manualTxn2.Write(REDIS, "John", manualPerson2)
	manualTxn2.Commit()

	var person2 testutil.Person
	err := txn.Read(REDIS, "John", &person2)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}

	// two read in txn should be the same
	if person1 != person2 {
		t.Errorf("Expected two read in txn to be the same")
	}
}

//   - txn1 starts  txn2 starts
//   - txn1 reads John
//   - txn2 reads John
//   - txn1 writes John
//   - txn2 read John again
//
// two read in txn2 should be the same
func TestRedis_RepeatableReadWhenAnotherUncommitted(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(REDIS, "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(REDIS)
		txn1.Start()
		var person testutil.Person
		txn1.Read(REDIS, "John", &person)
		time.Sleep(50 * time.Millisecond)

		// txn1 writes John
		person.Age = 31
		txn1.Write(REDIS, "John", person)

		// wait for txn2 to read John
		time.Sleep(100 * time.Millisecond)
		err := txn1.Commit()
		if err != nil {
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup(REDIS)
		txn2.Start()
		var person1 testutil.Person
		// txn2 reads John
		txn2.Read(REDIS, "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 testutil.Person
		// txn2 reads John again
		txn2.Read(REDIS, "John", &person2)

		// two read in txn2 should be the same
		if person1 != person2 {
			t.Errorf("Expected two read in txn2 to be the same")
		}

		err := txn2.Commit()
		if err != nil {
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	res1 := <-resChan
	res2 := <-resChan

	// both transactions should succeed
	if !res1 || !res2 {
		t.Errorf("Expected both transactions to succeed")
	}
}

//   - txn1 starts  txn2 starts
//   - txn1 reads John
//   - txn2 reads John
//   - txn1 writes John
//   - txn1 commits
//   - txn2 read John again
//
// two read in txn2 should be the same
func TestRedis_RepeatableReadWhenAnotherCommitted(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(REDIS, "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(REDIS)
		txn1.Start()
		var person testutil.Person
		txn1.Read(REDIS, "John", &person)
		time.Sleep(50 * time.Millisecond)

		// txn1 writes John
		person.Age = 31
		txn1.Write(REDIS, "John", person)

		err := txn1.Commit()
		if err != nil {
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup(REDIS)
		txn2.Start()
		var person1 testutil.Person
		// txn2 reads John
		txn2.Read(REDIS, "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 testutil.Person
		// txn2 reads John again
		txn2.Read(REDIS, "John", &person2)

		// two read in txn2 should be the same
		if person1 != person2 {
			t.Errorf("Expected two read in txn2 to be the same")
		}

		err := txn2.Commit()
		if err != nil {
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	res1 := <-resChan
	res2 := <-resChan

	// both transactions should succeed
	if !res1 || !res2 {
		t.Errorf("Expected both transactions to succeed")
	}
}

func TestRedis_TxnAbort(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	preTxn.Start()
	expected := testutil.NewDefaultPerson()
	preTxn.Write(REDIS, "John", expected)
	preTxn.Commit()

	txn := NewTransactionWithSetup(REDIS)
	var person testutil.Person
	txn.Start()
	txn.Read(REDIS, "John", &person)
	person.Age = 31
	txn.Write(REDIS, "John", person)
	txn.Abort()

	postTxn := NewTransactionWithSetup(REDIS)
	postTxn.Start()
	var postPerson testutil.Person
	postTxn.Read(REDIS, "John", &postPerson)
	postTxn.Commit()
	if postPerson != expected {
		t.Errorf("got %v want %v", postPerson, expected)
	}
}

// TODO: WTF why this test failed when using CLI
func TestRedis_TxnAbortCausedByWriteConflict(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(REDIS, item.Value, item)
	}
	err := preTxn.Commit()
	assert.Nil(t, err)

	txn := NewTransactionWithSetup(REDIS)
	manualTxn := NewTransactionWithSetup(REDIS)
	txn.Start()
	manualTxn.Start()

	// txn reads all items and modify them
	for _, item := range testutil.InputItemList {
		var actual testutil.TestItem
		txn.Read(REDIS, item.Value, &actual)
		actual.Value = item.Value + "updated"
		txn.Write(REDIS, item.Value, actual)
	}

	// manualTxn reads one item and modify it
	var manualItem testutil.TestItem
	manualTxn.Read(REDIS, "item4", &manualItem)
	manualItem.Value = "item4updated"
	manualTxn.Write(REDIS, "item4", manualItem)
	err = manualTxn.Commit()
	assert.Nil(t, err)

	err = txn.Commit()
	if err == nil {
		t.Errorf("Expected error committing transaction")
	}

	postTxn := NewTransactionWithSetup(REDIS)
	postTxn.Start()
	for _, item := range testutil.InputItemList {
		var actual testutil.TestItem
		postTxn.Read(REDIS, item.Value, &actual)
		if item.Value != "item4" {
			assert.Equal(t, item, actual)
		} else {
			assert.Equal(t, manualItem, actual)
		}
	}
	postTxn.Commit()
}

func TestRedis_ConcurrentTransaction(t *testing.T) {

	// Create a new redis datastore instance
	redisDst1 := redis.NewRedisDatastore("redis1", NewConnectionWithSetup(REDIS))

	txnFactory, err := factory.NewTransactionFactory(&factory.TransactionConfig{
		DatastoreList:    []txn.Datastorer{redisDst1},
		GlobalDatastore:  redisDst1,
		TimeOracleSource: txn.LOCAL,
		LockerSource:     txn.LOCAL,
	})
	assert.NoError(t, err)

	preTxn := txnFactory.NewTransaction()
	preTxn.Start()
	person := testutil.NewDefaultPerson()
	preTxn.Write(REDIS, "John", person)
	err = preTxn.Commit()
	assert.NoError(t, err)

	resChan := make(chan bool)
	conNum := 100

	conn := NewConnectionWithSetup(REDIS)

	for i := 1; i <= conNum; i++ {
		go func(id int) {
			txn := txn.NewTransaction()
			rds := redis.NewRedisDatastore(REDIS, conn)
			txn.AddDatastore(rds)
			txn.SetGlobalDatastore(rds)
			txn.Start()
			var person testutil.Person
			txn.Read(REDIS, "John", &person)
			person.Age = person.Age + id
			txn.Write(REDIS, "John", person)
			time.Sleep(1000 * time.Millisecond)
			err = txn.Commit()
			if err != nil {
				resChan <- false
			} else {
				resChan <- true
			}
		}(i)
	}

	successNum := 0
	for i := 1; i <= conNum; i++ {
		res := <-resChan
		if res {
			successNum++
		}
	}
	assert.Equal(t, 1, successNum)
}

// TestSimpleExpiredRead tests the scenario where a read operation is performed on an expired redis item.
// It inserts a redis item with an expired lease and a PREPARED state
// Then, it starts a transaction, reads the redis item,
// and verifies that the read item matches the expected value.
// Finally, it commits the transaction and checks that
// the redis item has been updated to the committed state.
func TestRedis_SimpleExpiredRead(t *testing.T) {
	tarMemItem := redis.NewRedisItem(txn.ItemOptions{
		Key:      "item1",
		Value:    util.ToJSONString(testutil.NewTestItem("item1")),
		TxnId:    "TestRedis_SimpleExpiredRead1",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-9 * time.Second),
		Version:  "1",
	})

	curMemItem := redis.NewRedisItem(txn.ItemOptions{
		Key:      "item1",
		Value:    util.ToJSONString(testutil.NewTestItem("item1-prepared")),
		TxnId:    "TestRedis_SimpleExpiredRead2",
		TxnState: config.PREPARED,
		TValid:   time.Now().Add(-5 * time.Second),
		TLease:   time.Now().Add(-4 * time.Second),
		Prev:     util.ToJSONString(tarMemItem),
		Version:  "2",
	})

	conn := NewConnectionWithSetup(REDIS)
	conn.PutItem("item1", curMemItem)

	txn := NewTransactionWithSetup(REDIS)
	txn.Start()

	var item testutil.TestItem
	txn.Read(REDIS, "item1", &item)
	assert.Equal(t, testutil.NewTestItem("item1"), item)
	err := txn.Commit()
	assert.NoError(t, err)
	actual, err := conn.GetItem("item1")
	assert.NoError(t, err)
	tarMemItem.SetVersion("3")
	if !tarMemItem.Equal(actual) {
		t.Errorf("\ngot\n%v\nwant\n%v", actual, tarMemItem)
	}
	// assert.Equal(t, util.ToJSONString(tarMemItem), util.ToJSONString(actual))

}

// A complex test
// preTxn writes data to the redis database
// slowTxn read all data and write all data, but it will block when conditionalUpdate item3 (sleep 4s)
// so when slowTxn blocks, the internal state of redis database:
//
//   - item1-slow PREPARED
//   - item2-slow PREPARED
//   - item3 COMMITTED
//   - item4 COMMITTED
//   - item5 COMMITTED
//
// fastTxn read item3, item4, item5 and write them, then commit
// the internal state of redis database:
//   - item1-slow PREPARED
//   - item2-slow PREPARED
//   - item3-fast COMMITTED
//   - item4-fast COMMITTED
//   - item5-fast COMMITTED
//
// then, slowTxn unblocks, it starts to conditionalUpdate item3
// and it detects a version mismatch,so it aborts(with rolling back all changes)
// postTxn reads all data and verify them
// so the final internal state of redis database:
//   - item1 rollback to COMMITTED
//   - item2 rollback to COMMITTED
//   - item3-fast COMMITTED
//   - item4-fast COMMITTED
//   - item5-fast COMMITTED
func TestRedis_SlowTransactionRecordExpiredWhenPrepare_Conflict(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(REDIS, item.Value, item)
	}
	preTxn.Commit()
	time.Sleep(500 * time.Millisecond)

	go func() {
		slowTxn := NewTransactionWithMockConn(REDIS, 2, false,
			0, func() error { time.Sleep(2 * time.Second); return nil })

		slowTxn.Start()
		for _, item := range testutil.InputItemList {
			var result testutil.TestItem
			slowTxn.Read(REDIS, item.Value, &result)
			result.Value = item.Value + "-slow"
			slowTxn.Write(REDIS, item.Value, result)
		}
		err := slowTxn.Commit()
		assert.EqualError(t, err, "prepare phase failed: version mismatch")
	}()
	time.Sleep(1 * time.Second)

	// ensure the internal state of redis database
	testConn := NewConnectionWithSetup(REDIS)
	testConn.Connect()
	memItem1, _ := testConn.GetItem("item1")
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-slow")), memItem1.Value())
	assert.Equal(t, memItem1.TxnState(), config.PREPARED)

	memItem2, _ := testConn.GetItem("item2")
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item2-slow")), memItem2.Value())
	assert.Equal(t, memItem2.TxnState(), config.PREPARED)

	memItem3, _ := testConn.GetItem("item3")
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item3")), memItem3.Value())
	assert.Equal(t, memItem3.TxnState(), config.COMMITTED)

	fastTxn := NewTransactionWithSetup(REDIS)
	fastTxn.Start()
	for i := 2; i <= 4; i++ {
		var result testutil.TestItem
		fastTxn.Read(REDIS, testutil.InputItemList[i].Value, &result)
		result.Value = testutil.InputItemList[i].Value + "-fast"
		fastTxn.Write(REDIS, testutil.InputItemList[i].Value, result)
	}
	err := fastTxn.Commit()
	assert.NoError(t, err)

	// wait for slowTxn to complete
	time.Sleep(1 * time.Second)
	postTxn := NewTransactionWithSetup(REDIS)
	postTxn.Start()

	var res1 testutil.TestItem
	postTxn.Read(REDIS, testutil.InputItemList[0].Value, &res1)
	assert.Equal(t, testutil.InputItemList[0], res1)

	var res2 testutil.TestItem
	postTxn.Read(REDIS, testutil.InputItemList[1].Value, &res2)
	assert.Equal(t, testutil.InputItemList[1], res2)

	for i := 2; i <= 4; i++ {
		var res testutil.TestItem
		postTxn.Read(REDIS, testutil.InputItemList[i].Value, &res)
		assert.Equal(t, testutil.InputItemList[i].Value+"-fast", res.Value)
	}

	err = postTxn.Commit()
	assert.NoError(t, err)

}

// A complex test
// preTxn writes data to the redis database
// slowTxn read all data and write all data,
// but it will block when conditionalUpdate item5 (sleep 2s)
// so when slowTxn blocks, the internal state of redis database:
//   - item1-slow PREPARED
//   - item2-slow PREPARED
//   - item3-slow PREPARED
//   - item4-slow PREPARED
//   - item5 COMMITTED
//
// fastTxn read item3, item4 and write them, then commit
// (fastTxn realize item3 and item4 are expired, so it will first rollback, and write the TSR with ABORTED)
// the internal state of redis database:
//   - item1-slow PREPARED
//   - item2-slow PREPARED
//   - item3-fast COMMITTED
//   - item4-fast COMMITTED
//   - item5 COMMITTED
//
// then, slowTxn unblocks, it conditionalUpdate item5 then check the TSR state
// the TSR is marked as ABORTED, so it aborts(with rolling back all changes)
// postTxn reads all data and verify them
// so the final internal state of redis database:
//   - item1 rollback to COMMITTED
//   - item2 rollback to COMMITTED
//   - item3-fast COMMITTED
//   - item4-fast COMMITTED
//   - item5 COMMITTED
func TestRedis_SlowTransactionRecordExpiredWhenPrepare_NoConflict(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(REDIS, item.Value, item)
	}
	err := preTxn.Commit()
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)

	go func() {
		slowTxn := NewTransactionWithMockConn(REDIS, 4, false,
			0, func() error { time.Sleep(2 * time.Second); return nil })
		slowTxn.Start()
		for _, item := range testutil.InputItemList {
			var result testutil.TestItem
			slowTxn.Read(REDIS, item.Value, &result)
			result.Value = item.Value + "-slow"
			slowTxn.Write(REDIS, item.Value, result)
		}
		err := slowTxn.Commit()
		assert.EqualError(t, err, "transaction is aborted by other transaction")
	}()

	time.Sleep(1 * time.Second)
	testConn := NewConnectionWithSetup(REDIS)
	testConn.Connect()

	// all records should be PREPARED state except item5
	for _, item := range testutil.InputItemList {
		memItem, err := testConn.GetItem(item.Value)
		assert.NoError(t, err)
		if item.Value == "item5" {
			assert.Equal(t, util.ToJSONString(testutil.NewTestItem(item.Value)), memItem.Value())
			assert.Equal(t, memItem.TxnState(), config.COMMITTED)
			continue
		}
		itemValue := item.Value + "-slow"
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem(itemValue)), memItem.Value())
		assert.Equal(t, memItem.TxnState(), config.PREPARED)
	}

	fastTxn := NewTransactionWithSetup(REDIS)
	err = fastTxn.Start()
	assert.NoError(t, err)
	for i := 2; i <= 3; i++ {
		var result testutil.TestItem
		fastTxn.Read(REDIS, testutil.InputItemList[i].Value, &result)
		result.Value = testutil.InputItemList[i].Value + "-fast"
		fastTxn.Write(REDIS, testutil.InputItemList[i].Value, result)
	}
	err = fastTxn.Commit()
	assert.NoError(t, err)

	// wait for slowTxn to complete
	time.Sleep(1 * time.Second)
	postTxn := NewTransactionWithSetup(REDIS)
	postTxn.Start()

	var res1 testutil.TestItem
	postTxn.Read(REDIS, testutil.InputItemList[0].Value, &res1)
	assert.Equal(t, testutil.InputItemList[0], res1)

	var res2 testutil.TestItem
	postTxn.Read(REDIS, testutil.InputItemList[1].Value, &res2)
	assert.Equal(t, testutil.InputItemList[1], res2)

	for i := 2; i <= 3; i++ {
		var res testutil.TestItem
		postTxn.Read(REDIS, testutil.InputItemList[i].Value, &res)
		assert.Equal(t, testutil.InputItemList[i].Value+"-fast", res.Value)
	}

	var res5 testutil.TestItem
	postTxn.Read(REDIS, testutil.InputItemList[4].Value, &res5)
	assert.Equal(t, testutil.InputItemList[4].Value, res5.Value)

	err = postTxn.Commit()
	assert.NoError(t, err)

	// for debug only
	// time.Sleep(10 * time.Second)
}

// A complex test
// preTxn writes data to the redis database
// slowTxn read all data and write all data,
// but it will block for 3s and **fail** when writing the TSR
// so when slowTxn blocks, the internal state of redis database:
//   - item1-slow PREPARED
//   - item2-slow PREPARED
//   - item3-slow PREPARED
//   - item4-slow PREPARED
//   - item5-slow PREPARED
//
// testTxn read item1,item2,item3, item4
// (testTxn realize item1,item2,item3, item4 are expired, so it will first rollback, and write the TSR with ABORTED)
// the internal state of redis database:
//   - item1 rollback to COMMITTED
//   - item2 rollback to COMMITTED
//   - item3 rollback to COMMITTED
//   - item4 rollback to COMMITTED
//   - item5-slow PREPARED
//
// then, slowTxn unblocks, it fails to write the TSR, and it aborts(it tries to rollback all the items)
// so slowTxn will abort(with rolling back all changes)
// postTxn reads all data and verify them
// so the final internal state of redis database:
//   - item1 rollback to COMMITTED
//   - item2 rollback to COMMITTED
//   - item3 rollback to COMMITTED
//   - item4 rollback to COMMITTED
//   - item5 rollback to COMMITTED
func TestRedis_TransactionAbortWhenWritingTSR(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(REDIS, item.Value, item)
	}
	err := preTxn.Commit()
	if err != nil {
		t.Errorf("preTxn commit err: %s", err)
	}
	time.Sleep(1 * time.Second)

	txn := NewTransactionWithMockConn(REDIS, 5, true,
		0, func() error { time.Sleep(3 * time.Second); return errors.New("fail to write TSR") })
	txn.Start()
	for _, item := range testutil.InputItemList {
		var result testutil.TestItem
		txn.Read(REDIS, item.Value, &result)
		result.Value = item.Value + "-slow"
		txn.Write(REDIS, item.Value, result)
	}
	err = txn.Commit()
	assert.EqualError(t, err, "fail to write TSR")

	testTxn := NewTransactionWithSetup(REDIS)
	testTxn.Start()

	for i := 0; i <= 3; i++ {
		item := testutil.InputItemList[i]
		var memItem testutil.TestItem
		testTxn.Read(REDIS, item.Value, &memItem)
	}
	err = testTxn.Commit()
	assert.NoError(t, err)
	postTxn := NewTransactionWithSetup(REDIS)
	postTxn.Start()
	for i := 0; i <= 3; i++ {
		item := testutil.InputItemList[i]
		var memItem testutil.TestItem
		postTxn.Read(REDIS, item.Value, &memItem)
		assert.Equal(t, item.Value, memItem.Value)
	}

	conn := NewConnectionWithSetup(REDIS)
	memItem, err := conn.GetItem("item5")
	assert.NoError(t, err)
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item5")), memItem.Value())
	assert.Equal(t, config.COMMITTED, memItem.TxnState())
}

func TestRedis_LinkedRecord(t *testing.T) {

	t.Cleanup(func() {
		config.Config.MaxRecordLength = 2
	})

	t.Run("commit time less than MaxLen", func(t *testing.T) {

		preTxn := NewTransactionWithSetup(REDIS)
		preTxn.Start()
		person := testutil.NewDefaultPerson()
		preTxn.Write(REDIS, "John", person)
		err := preTxn.Commit()
		assert.NoError(t, err)

		slowTxn := NewTransactionWithSetup(REDIS)
		slowTxn.Start()

		config.Config.MaxRecordLength = 4
		// 1+2=3 < 4, including origin
		commitTime := 2
		for i := 1; i <= commitTime; i++ {
			txn := NewTransactionWithSetup(REDIS)
			txn.Start()
			var p testutil.Person
			txn.Read(REDIS, "John", &p)
			p.Age = p.Age + 1
			txn.Write(REDIS, "John", p)
			err = txn.Commit()
			assert.NoError(t, err)
		}

		var p testutil.Person
		err = slowTxn.Read(REDIS, "John", &p)
		assert.NoError(t, err)
		assert.Equal(t, 30, p.Age)
	})

	t.Run("commit time equals MaxLen", func(t *testing.T) {

		preTxn := NewTransactionWithSetup(REDIS)
		preTxn.Start()
		person := testutil.NewDefaultPerson()
		preTxn.Write(REDIS, "John", person)
		err := preTxn.Commit()
		assert.NoError(t, err)

		slowTxn := NewTransactionWithSetup(REDIS)
		slowTxn.Start()

		config.Config.MaxRecordLength = 4
		// 1+3=4 == 4, including origin
		commitTime := 3
		for i := 1; i <= commitTime; i++ {
			txn := NewTransactionWithSetup(REDIS)
			txn.Start()
			var p testutil.Person
			txn.Read(REDIS, "John", &p)
			p.Age = p.Age + 1
			txn.Write(REDIS, "John", p)
			err = txn.Commit()
			assert.NoError(t, err)
		}

		var p testutil.Person
		err = slowTxn.Read(REDIS, "John", &p)
		assert.NoError(t, err)
		assert.Equal(t, 30, p.Age)
	})

	t.Run("commit times bigger than MaxLen", func(t *testing.T) {

		preTxn := NewTransactionWithSetup(REDIS)
		preTxn.Start()
		person := testutil.NewDefaultPerson()
		preTxn.Write(REDIS, "John", person)
		err := preTxn.Commit()
		assert.NoError(t, err)

		slowTxn := NewTransactionWithSetup(REDIS)
		slowTxn.Start()

		config.Config.MaxRecordLength = 4
		// 1+4=5 > 4, including origin
		commitTime := 4
		for i := 1; i <= commitTime; i++ {
			txn := NewTransactionWithSetup(REDIS)
			txn.Start()
			var p testutil.Person
			txn.Read(REDIS, "John", &p)
			p.Age = p.Age + 1
			txn.Write(REDIS, "John", p)
			err = txn.Commit()
			assert.NoError(t, err)
		}

		var p testutil.Person
		err = slowTxn.Read(REDIS, "John", &p)
		assert.EqualError(t, err, "key not found")
	})
}

func TestRedis_RollbackConflict(t *testing.T) {

	// there is a broken item
	//   - txnA reads the item, decides to roll back
	//   - txnB reads the item, decides to roll back
	//   - txnB rollbacks the item
	//   - txnB update the item and commit
	//   - txnA rollbacks the item -> should fail
	t.Run("the broken item has a valid Prev field", func(t *testing.T) {
		conn := NewConnectionWithSetup(REDIS)

		redisItem1 := &redis.RedisItem{
			RKey:       "item1",
			RValue:     util.ToJSONString(testutil.NewTestItem("item1-pre")),
			RTxnId:     "TestRedis_RollbackConflict1",
			RTxnState:  config.COMMITTED,
			RTValid:    time.Now().Add(-5 * time.Second),
			RTLease:    time.Now().Add(-4 * time.Second),
			RLinkedLen: 1,
			RVersion:   "1",
		}
		redisItem2 := &redis.RedisItem{
			RKey:       "item1",
			RValue:     util.ToJSONString(testutil.NewTestItem("item1-broken")),
			RTxnId:     "TestRedis_RollbackConflict2",
			RTxnState:  config.PREPARED,
			RTValid:    time.Now().Add(-5 * time.Second),
			RTLease:    time.Now().Add(-4 * time.Second),
			RPrev:      util.ToJSONString(redisItem1),
			RLinkedLen: 2,
			RVersion:   "2",
		}
		conn.PutItem("item1", redisItem2)

		go func() {
			time.Sleep(100 * time.Millisecond)
			txnA := NewTransactionWithMockConn(REDIS, 0, false,
				0, func() error { time.Sleep(2 * time.Second); return nil })
			txnA.Start()

			var item testutil.TestItem
			err := txnA.Read(REDIS, "item1", &item)
			assert.NotNil(t, err)
		}()

		time.Sleep(100 * time.Millisecond)
		txnB := NewTransactionWithSetup(REDIS)
		txnB.Start()
		var item testutil.TestItem
		err := txnB.Read(REDIS, "item1", &item)
		assert.NoError(t, err)
		assert.Equal(t, testutil.NewTestItem("item1-pre"), item)
		item.Value = "item1-B"
		txnB.Write(REDIS, "item1", item)
		err = txnB.Commit()
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)
		resItem, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())
		redisItem1.RVersion = "3"
		assert.Equal(t, util.ToJSONString(redisItem1), resItem.Prev())
	})

	// there is a broken item
	//   - txnA reads the item, decides to roll back
	//   - txnB reads the item, decides to roll back
	//   - txnB rollbacks the item
	//   - txnB update the item and commit
	//   - txnA rollbacks the item -> should fail
	t.Run("the broken item has an empty Prev field", func(t *testing.T) {
		conn := NewConnectionWithSetup(REDIS)

		redisItem2 := &redis.RedisItem{
			RKey:       "item1",
			RValue:     util.ToJSONString(testutil.NewTestItem("item1-broken")),
			RTxnId:     "TestRedis_RollbackConflict2-emptyField",
			RTxnState:  config.PREPARED,
			RTValid:    time.Now().Add(-5 * time.Second),
			RTLease:    time.Now().Add(-4 * time.Second),
			RPrev:      "",
			RLinkedLen: 1,
			RVersion:   "1",
		}
		conn.PutItem("item1", redisItem2)

		go func() {
			time.Sleep(100 * time.Millisecond)
			txnA := NewTransactionWithMockConn(REDIS, 0, false,
				0, func() error { time.Sleep(2 * time.Second); return nil })
			txnA.Start()

			var item testutil.TestItem
			err := txnA.Read(REDIS, "item1", &item)
			assert.NotNil(t, err)
		}()

		time.Sleep(100 * time.Millisecond)
		txnB := NewTransactionWithSetup(REDIS)
		txnB.Start()
		var item testutil.TestItem
		err := txnB.Read(REDIS, "item1", &item)
		assert.EqualError(t, err, txn.KeyNotFound.Error())

		item.Value = "item1-B"
		txnB.Write(REDIS, "item1", item)
		err = txnB.Commit()
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)
		resItem, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())
		redisItem2.RIsDeleted = true
		redisItem2.RTxnState = config.COMMITTED
		redisItem2.RVersion = util.AddToString(redisItem2.RVersion, 1)
		assert.Equal(t, util.ToJSONString(redisItem2), resItem.Prev())
	})

}

// there is a broken item
//   - txnA reads the item, decides to roll forward
//   - txnB reads the item, decides to roll forward
//   - txnB rolls forward the item
//   - txnB update the item and commit
//   - txnA rolls forward the item -> should fail
func TestRedis_RollForwardConflict(t *testing.T) {
	conn := NewConnectionWithSetup(REDIS)

	redisItem1 := &redis.RedisItem{
		RKey:      "item1",
		RValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
		RTxnId:    "TestRedis_RollForwardConflict1",
		RTxnState: config.COMMITTED,
		RTValid:   time.Now().Add(-5 * time.Second),
		RTLease:   time.Now().Add(-4 * time.Second),
		RVersion:  "1",
	}
	redisItem2 := &redis.RedisItem{
		RKey:       "item1",
		RValue:     util.ToJSONString(testutil.NewTestItem("item1-broken")),
		RTxnId:     "TestRedis_RollForwardConflict2",
		RTxnState:  config.PREPARED,
		RTValid:    time.Now().Add(-5 * time.Second),
		RTLease:    time.Now().Add(-4 * time.Second),
		RPrev:      util.ToJSONString(redisItem1),
		RLinkedLen: 2,
		RVersion:   "2",
	}
	conn.PutItem("item1", redisItem2)
	conn.Put("TestRedis_RollForwardConflict2", config.COMMITTED)

	go func() {
		time.Sleep(100 * time.Millisecond)
		txnA := NewTransactionWithMockConn(REDIS, 0, false,
			0, func() error { time.Sleep(2 * time.Second); return nil })
		txnA.Start()
		var item testutil.TestItem
		err := txnA.Read(REDIS, "item1", &item)
		assert.NotNil(t, err)

	}()

	time.Sleep(100 * time.Millisecond)
	txnB := NewTransactionWithSetup(REDIS)
	txnB.Start()
	var item testutil.TestItem
	txnB.Read(REDIS, "item1", &item)
	assert.Equal(t, testutil.NewTestItem("item1-broken"), item)
	item.Value = "item1-B"
	txnB.Write(REDIS, "item1", item)
	err := txnB.Commit()
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)
	resItem, err := conn.GetItem("item1")
	assert.NoError(t, err)
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())
	redisItem2.RTxnState = config.COMMITTED
	redisItem2.RPrev = ""
	redisItem2.RLinkedLen = 1
	redisItem2.RVersion = "3"
	assert.Equal(t, util.ToJSONString(redisItem2), resItem.Prev())

}

func TestRedis_ConcurrentDirectWrite(t *testing.T) {

	conn := NewConnectionWithSetup(REDIS)
	conn.Delete("item1")

	conNumber := 5
	mu := sync.Mutex{}
	globalId := 0
	resChan := make(chan bool)
	for i := 1; i <= conNumber; i++ {
		go func(id int) {
			txn := NewTransactionWithSetup(REDIS)
			item := testutil.NewTestItem("item1-" + strconv.Itoa(id))
			txn.Start()
			txn.Write(REDIS, "item1", item)
			err := txn.Commit()
			if err != nil {
				resChan <- false
			} else {
				mu.Lock()
				globalId = id
				mu.Unlock()
				resChan <- true
			}
		}(i)
	}

	successNum := 0
	for i := 1; i <= conNumber; i++ {
		res := <-resChan
		if res {
			successNum++
		}
	}

	assert.Equal(t, 1, successNum)

	postTxn := NewTransactionWithSetup(REDIS)
	postTxn.Start()
	var item testutil.TestItem
	postTxn.Read(REDIS, "item1", &item)
	assert.Equal(t, testutil.NewTestItem("item1-"+strconv.Itoa(globalId)), item)
	// assert.Equal(t, 0, globalId)
	t.Logf("item: %v", item)
}

func TestRedis_TxnDelete(t *testing.T) {
	preTxn := NewTransactionWithSetup(REDIS)
	preTxn.Start()
	item := testutil.NewTestItem("item1-pre")
	preTxn.Write(REDIS, "item1", item)
	err := preTxn.Commit()
	assert.NoError(t, err)

	txn := NewTransactionWithSetup(REDIS)
	txn.Start()
	var item1 testutil.TestItem
	txn.Read(REDIS, "item1", &item1)
	txn.Delete(REDIS, "item1")
	err = txn.Commit()
	assert.NoError(t, err)

	postTxn := NewTransactionWithSetup(REDIS)
	postTxn.Start()
	var item2 testutil.TestItem
	err = postTxn.Read(REDIS, "item1", &item2)
	assert.EqualError(t, err, "key not found")
}

func TestRedis_PreventLostUpdatesValidation(t *testing.T) {

	t.Run("Case 1-1(with read): The target record has been updated by the concurrent transaction",
		func(t *testing.T) {

			preTxn := NewTransactionWithSetup(REDIS)
			preTxn.Start()
			item := testutil.NewTestItem("item1-pre")
			preTxn.Write(REDIS, "item1", item)
			err := preTxn.Commit()
			assert.NoError(t, err)

			txnA := NewTransactionWithSetup(REDIS)
			txnA.Start()
			txnB := NewTransactionWithSetup(REDIS)
			txnB.Start()

			var itemA testutil.TestItem
			txnA.Read(REDIS, "item1", &itemA)
			itemA.Value = "item1-A"
			txnA.Write(REDIS, "item1", itemA)
			err = txnA.Commit()
			assert.NoError(t, err)

			var itemB testutil.TestItem
			txnB.Read(REDIS, "item1", &itemB)
			itemB.Value = "item1-B"
			txnB.Write(REDIS, "item1", itemB)
			err = txnB.Commit()
			assert.EqualError(t, err,
				"prepare phase failed: version mismatch")

			postTxn := NewTransactionWithSetup(REDIS)
			postTxn.Start()
			var item1 testutil.TestItem
			postTxn.Read(REDIS, "item1", &item1)
			assert.Equal(t, itemA, item1)
		})

	t.Run("Case 1-2(without read): The target record has been updated by the concurrent transaction",
		func(t *testing.T) {

			preTxn := NewTransactionWithSetup(REDIS)
			preTxn.Start()
			item := testutil.NewTestItem("item1-pre")
			preTxn.Write(REDIS, "item1", item)
			err := preTxn.Commit()
			assert.NoError(t, err)

			txnA := NewTransactionWithSetup(REDIS)
			txnA.Start()
			txnB := NewTransactionWithSetup(REDIS)
			txnB.Start()

			itemA := testutil.NewTestItem("item1-A")
			txnA.Write(REDIS, "item1", itemA)
			err = txnA.Commit()
			assert.NoError(t, err)

			itemB := testutil.NewTestItem("item1-B")
			txnB.Write(REDIS, "item1", itemB)
			err = txnB.Commit()
			assert.EqualError(t, err,
				"prepare phase failed: version mismatch")

			postTxn := NewTransactionWithSetup(REDIS)
			postTxn.Start()
			var item1 testutil.TestItem
			postTxn.Read(REDIS, "item1", &item1)
			assert.Equal(t, itemA, item1)
		})

	t.Run("Case 2-1(with read): There is no conflict", func(t *testing.T) {

		preTxn := NewTransactionWithSetup(REDIS)
		preTxn.Start()
		item := testutil.NewTestItem("item1-pre")
		preTxn.Write(REDIS, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txnA := NewTransactionWithSetup(REDIS)
		txnA.Start()
		var itemA testutil.TestItem
		txnA.Read(REDIS, "item1", &itemA)
		itemA.Value = "item1-A"
		txnA.Write(REDIS, "item1", itemA)
		err = txnA.Commit()
		assert.NoError(t, err)

		txnB := NewTransactionWithSetup(REDIS)
		txnB.Start()
		var itemB testutil.TestItem
		txnB.Read(REDIS, "item1", &itemB)
		itemB.Value = "item1-B"
		txnB.Write(REDIS, "item1", itemB)
		err = txnB.Commit()
		assert.NoError(t, err)

		postTxn := NewTransactionWithSetup(REDIS)
		postTxn.Start()
		var item1 testutil.TestItem
		postTxn.Read(REDIS, "item1", &item1)
		assert.Equal(t, itemB, item1)
	})

	t.Run("Case 2-2(without read): There is no conflict", func(t *testing.T) {

		preTxn := NewTransactionWithSetup(REDIS)
		preTxn.Start()
		item := testutil.NewTestItem("item1-pre")
		preTxn.Write(REDIS, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)
		time.Sleep(1 * time.Second)

		txnA := NewTransactionWithSetup(REDIS)
		txnA.Start()
		itemA := testutil.NewTestItem("item1-A")
		txnA.Write(REDIS, "item1", itemA)
		err = txnA.Commit()
		assert.NoError(t, err)

		txnB := NewTransactionWithSetup(REDIS)
		txnB.Start()
		itemB := testutil.NewTestItem("item1-B")
		txnB.Write(REDIS, "item1", itemB)
		err = txnB.Commit()
		assert.NoError(t, err)

		postTxn := NewTransactionWithSetup(REDIS)
		postTxn.Start()
		var item1 testutil.TestItem
		postTxn.Read(REDIS, "item1", &item1)
		assert.Equal(t, itemB, item1)
	})
}

func TestRedis_RepeatableReadWhenDirtyRead(t *testing.T) {
	t.Run("the prepared item has a valid Prev", func(t *testing.T) {
		config.Config.LeaseTime = 3000 * time.Millisecond

		preTxn := NewTransactionWithSetup(REDIS)
		preTxn.Start()
		item := testutil.NewTestItem("item1-pre")
		preTxn.Write(REDIS, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txnA := NewTransactionWithMockConn(REDIS, 1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		txnB := NewTransactionWithSetup(REDIS)
		txnA.Start()
		txnB.Start()

		go func() {
			itemA := testutil.NewTestItem("item1-A")
			txnA.Write(REDIS, "item1", itemA)
			err = txnA.Commit()
			assert.NoError(t, err)
		}()

		// wait for txnA to write item into database
		time.Sleep(500 * time.Millisecond)
		var itemB testutil.TestItem
		err = txnB.Read(REDIS, "item1", &itemB)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre", itemB.Value)

		time.Sleep(2 * time.Second)
		err = txnB.Read(REDIS, "item1", &itemB)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre", itemB.Value)
	})

	t.Run("the prepared item has an empty Prev", func(t *testing.T) {
		config.Config.LeaseTime = 3000 * time.Millisecond

		testConn := NewConnectionWithSetup(REDIS)
		testConn.Delete("item1")

		txnA := NewTransactionWithMockConn(REDIS, 1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		txnB := NewTransactionWithSetup(REDIS)
		txnA.Start()
		txnB.Start()

		go func() {
			itemA := testutil.NewTestItem("item1-A")
			txnA.Write(REDIS, "item1", itemA)
			err := txnA.Commit()
			assert.NoError(t, err)
		}()

		// wait for txnA to write item into database
		time.Sleep(500 * time.Millisecond)
		var itemB testutil.TestItem
		err := txnB.Read(REDIS, "item1", &itemB)
		assert.EqualError(t, err, txn.KeyNotFound.Error())

		time.Sleep(2 * time.Second)

		// make sure txnA has committed
		resItem, err := testConn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-A")), resItem.Value())

		err = txnB.Read(REDIS, "item1", &itemB)
		assert.EqualError(t, err, txn.KeyNotFound.Error())

		// post check
		postTxn := NewTransactionWithSetup(REDIS)
		postTxn.Start()

		var itemPost testutil.TestItem
		err = postTxn.Read(REDIS, "item1", &itemPost)
		assert.NoError(t, err)
		assert.Equal(t, "item1-A", itemPost.Value)
	})

}

func TestRedis_DeleteTimingProblems(t *testing.T) {

	// conn Puts an item with an empty Prev field
	//  - txnA reads the item, decides to delete
	//  - txnB reads the item, updates and commmits
	//  - txnA tries to delete the item -> should fail
	t.Run("the item has an empty Prev field", func(t *testing.T) {
		testConn := NewConnectionWithSetup(REDIS)
		dbItem := &redis.RedisItem{
			RKey:      "item1",
			RValue:    util.ToJSONString(testutil.NewTestItem("item1-pre")),
			RTxnId:    "TestRedis_DeleteTimingProblems",
			RTxnState: config.COMMITTED,
			RTValid:   time.Now().Add(-5 * time.Second),
			RTLease:   time.Now().Add(-4 * time.Second),
			RVersion:  "1",
		}
		testConn.PutItem("item1", dbItem)

		txnA := NewTransactionWithMockConn(REDIS, 0, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		txnB := NewTransactionWithSetup(REDIS)
		txnA.Start()
		txnB.Start()

		go func() {
			var item testutil.TestItem
			txnA.Read(REDIS, "item1", &item)
			txnA.Delete(REDIS, "item1")
			err := txnA.Commit()
			assert.NotNil(t, err)
			// assert.NoError(t, err)
		}()

		time.Sleep(200 * time.Millisecond)
		var itemB testutil.TestItem
		txnB.Read(REDIS, "item1", &itemB)
		txnB.Delete(REDIS, "item1")
		itemB.Value = "item1-B"
		txnB.Write(REDIS, "item1", itemB)
		err := txnB.Commit()
		assert.NoError(t, err)

		time.Sleep(1 * time.Second)

		// post check
		resItem, err := testConn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())

	})
}

func TestRedis_VisibilityResults(t *testing.T) {

	t.Run("a normal chain", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(REDIS)
		preTxn.Start()
		item := testutil.NewTestItem("item1-V0")
		preTxn.Write(REDIS, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		chainNum := 5

		for i := 1; i <= chainNum; i++ {
			time.Sleep(10 * time.Millisecond)
			txn := NewTransactionWithSetup(REDIS)
			txn.Start()
			var item testutil.TestItem
			txn.Read(REDIS, "item1", &item)
			assert.Equal(t, "item1-V"+strconv.Itoa(i-1), item.Value)
			item.Value = "item1-V" + strconv.Itoa(i)
			txn.Write(REDIS, "item1", item)
			err = txn.Commit()
			assert.NoError(t, err)
		}

	})
}

// A network call times test
//
// when a transaction read-modify-commit X items at once, it will:
//  - call `Get` X+1 times
//  - call `Put` 2X+2 times

func TestRedis_ReadModifyWritePattern(t *testing.T) {

	t.Run("when X = 1", func(t *testing.T) {

		X := 1
		preTxn := NewTransactionWithSetup(REDIS)
		preTxn.Start()
		dbItem := testutil.NewTestItem("item1-pre")
		preTxn.Write(REDIS, "item1", dbItem)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txn := txn.NewTransaction()
		mockConn := mock.NewMockRedisConnection("localhost", 6379, -1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		rds := redis.NewRedisDatastore(REDIS, mockConn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
		txn.Start()

		var item1 testutil.TestItem
		err = txn.Read(REDIS, "item1", &item1)
		assert.NoError(t, err)
		item1.Value = "item1-modified"
		txn.Write(REDIS, "item1", item1)
		err = txn.Commit()
		assert.NoError(t, err)

		// t.Logf("Connection status:\nGetTimes: %d\nPutTimes: %d\n",
		// 	mockConn.GetTimes, mockConn.PutTimes)

		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, X+1, mockConn.GetTimes)
		assert.Equal(t, 2*X+2, mockConn.PutTimes)
	})

	t.Run("when X = 5", func(t *testing.T) {

		X := 5
		preTxn := NewTransactionWithSetup(REDIS)
		preTxn.Start()
		for i := 1; i <= X; i++ {
			dbItem := testutil.NewTestItem("item" + strconv.Itoa(i) + "-pre")
			preTxn.Write(REDIS, "item"+strconv.Itoa(i), dbItem)
		}
		err := preTxn.Commit()
		assert.NoError(t, err)
		time.Sleep(1 * time.Second)

		txn := txn.NewTransaction()
		mockConn := mock.NewMockRedisConnection("localhost", 6379, -1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		rds := redis.NewRedisDatastore(REDIS, mockConn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
		txn.Start()

		for i := 1; i <= X; i++ {
			var item testutil.TestItem
			err = txn.Read(REDIS, "item"+strconv.Itoa(i), &item)
			assert.NoError(t, err)
			item.Value = "item" + strconv.Itoa(i) + "-modified"
			txn.Write(REDIS, "item"+strconv.Itoa(i), item)
		}
		err = txn.Commit()
		assert.NoError(t, err)

		// t.Logf("Connection status:\nGetTimes: %d\nPutTimes: %d\n",
		// 	mockConn.GetTimes, mockConn.PutTimes)

		time.Sleep(1000 * time.Millisecond)
		assert.Equal(t, X+1, mockConn.GetTimes)
		assert.Equal(t, 2*X+2, mockConn.PutTimes)
	})

}

// A timing case.
//   - txnA reads,modifies,writes item1
//   - txnA returns before datastore.Commit() finishes
//   - txnB reads item1, and brings item1 to COMMITTED
//   - txnB modifies item1 and commit
//   - txnA tries to bring item1 to COMMITTED state, it should fail
func TestRedis_CommitTimingCase(t *testing.T) {

	preTxn := NewTransactionWithSetup(REDIS)
	preTxn.Start()
	item := testutil.NewTestItem("item1-pre")
	preTxn.Write(REDIS, "item1", item)
	err := preTxn.Commit()
	assert.NoError(t, err)

	go func() {

		txnA := NewTransactionWithMockConn(REDIS, 2, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		txnA.Start()
		var item1 testutil.TestItem
		err = txnA.Read(REDIS, "item1", &item1)
		assert.NoError(t, err)
		item1.Value = "item1-A"
		txnA.Write(REDIS, "item1", item1)
		err = txnA.Commit()
		assert.NoError(t, err)
	}()

	time.Sleep(200 * time.Millisecond)
	txnB := NewTransactionWithSetup(REDIS)
	txnB.Start()
	var item1 testutil.TestItem
	err = txnB.Read(REDIS, "item1", &item1)
	assert.NoError(t, err)
	item1.Value = "item1-B"
	txnB.Write(REDIS, "item1", item1)
	err = txnB.Commit()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	postTxn := NewTransactionWithSetup(REDIS)
	postTxn.Start()
	var resItem testutil.TestItem
	err = postTxn.Read(REDIS, "item1", &resItem)
	assert.NoError(t, err)
	assert.Equal(t, "item1-B", resItem.Value)

}

func TestRedis_ConcurrentOptimization(t *testing.T) {

	t.Run("test datastore.Commit() should be concurrent", func(t *testing.T) {

		// Params
		X := 5
		delay := 50 * time.Millisecond

		preTxn := NewTransactionWithSetup(REDIS)
		preTxn.Start()

		for i := 1; i <= X; i++ {
			dbItem := testutil.NewTestItem("item" + strconv.Itoa(i) + "-pre")
			preTxn.Write(REDIS, "item"+strconv.Itoa(i), dbItem)
		}
		err := preTxn.Commit()
		assert.NoError(t, err)

		startTime := time.Now()

		txn := txn.NewTransaction()
		mockConn := mock.NewMockRedisConnection("localhost", 6379, -1, false,
			delay, func() error { time.Sleep(1 * time.Second); return nil })
		rds := redis.NewRedisDatastore(REDIS, mockConn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
		txn.Start()

		for i := 1; i <= 5; i++ {
			var item testutil.TestItem
			txn.Read(REDIS, "item"+strconv.Itoa(i), &item)
			item.Value = "item" + strconv.Itoa(i) + "-modified"
			txn.Write(REDIS, "item"+strconv.Itoa(i), item)
		}
		err = txn.Commit()
		assert.NoError(t, err)

		endTime := time.Now()

		executionTime := endTime.Sub(startTime)

		threshold := 50 * time.Millisecond

		ok := testutil.RoughlyEqual(time.Duration(2*X+4)*delay, executionTime, threshold)
		assert.True(t, ok)
		t.Logf("Execution time: %v\n Expected Time: %v +-%v ",
			executionTime, time.Duration(2*X+4)*delay, threshold)
	})

	t.Run("test transaction.Commit() should be concurrent", func(t *testing.T) {

		// Params
		X := 5
		delay := 50 * time.Millisecond

		preTxn1 := NewTransactionWithSetup(REDIS)
		preTxn1.Start()
		for i := 1; i <= X; i++ {
			dbItem := testutil.NewTestItem("item" + strconv.Itoa(i) + "-pre")
			preTxn1.Write(REDIS, "item"+strconv.Itoa(i), dbItem)
		}
		err := preTxn1.Commit()
		assert.NoError(t, err)

		preTxn2 := NewTransactionWithSetup(REDIS)
		preTxn2.Start()
		for i := 1; i <= X; i++ {
			dbItem := testutil.NewTestItem("item" + strconv.Itoa(i) + "-pre")
			preTxn2.Write(REDIS, "item"+strconv.Itoa(i), dbItem)
		}
		err = preTxn2.Commit()
		assert.NoError(t, err)

		startTime := time.Now()

		txn := txn.NewTransaction()
		mockRedisConn := mock.NewMockRedisConnection("localhost", 6379, -1, false,
			delay, func() error { time.Sleep(1 * time.Second); return nil })
		rds := redis.NewRedisDatastore(REDIS, mockRedisConn)
		mockKvConn := mock.NewMockRedisConnection("localhost", 6666, -1, false,
			delay, func() error { time.Sleep(1 * time.Second); return nil })
		kds := redis.NewRedisDatastore(KVROCKS, mockKvConn)
		txn.AddDatastore(rds)
		txn.AddDatastore(kds)
		txn.SetGlobalDatastore(rds)
		txn.Start()

		for i := 1; i <= 5; i++ {
			var item testutil.TestItem
			txn.Read(REDIS, "item"+strconv.Itoa(i), &item)
			item.Value = "item" + strconv.Itoa(i) + "-modified"
			txn.Write(REDIS, "item"+strconv.Itoa(i), item)

			txn.Read(KVROCKS, "item"+strconv.Itoa(i), &item)
			item.Value = "item" + strconv.Itoa(i) + "-modified"
			txn.Write(KVROCKS, "item"+strconv.Itoa(i), item)
		}
		err = txn.Commit()
		assert.NoError(t, err)

		endTime := time.Now()

		executionTime := endTime.Sub(startTime)

		threshold := 100 * time.Millisecond

		ok := testutil.RoughlyEqual(time.Duration(4*X+4)*delay, executionTime, threshold)
		assert.True(t, ok)
		t.Logf("Execution time: %v\n Expected Time: %v +-%v ",
			executionTime, time.Duration(4*X+4)*delay, threshold)

	})
}
