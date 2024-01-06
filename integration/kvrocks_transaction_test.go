package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

func TestKvrocks_TxnWrite(t *testing.T) {

	txn1 := NewTransactionWithSetup(KVROCKS)
	expected := testutil.NewDefaultPerson()

	// clear the data
	conn := NewRedisConnection()
	conn.Delete("John")

	// Txn1 writes the record
	err := txn1.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	err = txn1.Write(KVROCKS, "John", expected)
	if err != nil {
		t.Errorf("Error writing record: %s", err)
	}
	err = txn1.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

	txn2 := NewTransactionWithSetup(KVROCKS)

	// Txn2 reads the record
	err = txn2.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var actual testutil.Person
	err = txn2.Read(KVROCKS, "John", &actual)
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

func TestKvrocks_ReadOwnWrite(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(KVROCKS, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(KVROCKS)

	// Txn reads the record
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var person testutil.Person
	err = txn.Read(KVROCKS, "John", &person)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}

	expected := person
	expected.Age = 31
	txn.Write(KVROCKS, "John", expected)

	var actual testutil.Person
	err = txn.Read(KVROCKS, "John", &actual)
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

func TestKvrocks_SingleKeyWriteConflict(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(KVROCKS, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(KVROCKS)
	txn.Start()
	var person testutil.Person
	txn.Read(KVROCKS, "John", &person)
	person.Age = 31
	txn.Write(KVROCKS, "John", person)
	var anotherPerson testutil.Person
	txn.Read(KVROCKS, "John", &anotherPerson)

	if person != anotherPerson {
		t.Errorf("Expected two read to be the same")
	}
	txn.Commit()

	postTxn := NewTransactionWithSetup(KVROCKS)
	postTxn.Start()
	var postPerson testutil.Person
	postTxn.Read(KVROCKS, "John", &postPerson)
	if postPerson != person {
		t.Errorf("got %v want %v", postPerson, person)
	}

}

func TestKvrocks_MultileKeyWriteConflict(t *testing.T) {

	// clear the data
	conn := NewRedisConnection()
	conn.Delete("item1")
	conn.Delete("item2")

	preTxn := NewTransactionWithSetup(KVROCKS)
	item1 := testutil.NewTestItem("item1")
	item2 := testutil.NewTestItem("item2")
	preTxn.Start()
	preTxn.Write(KVROCKS, "item1", item1)
	preTxn.Write(KVROCKS, "item2", item2)
	err := preTxn.Commit()
	assert.Nil(t, err)

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(KVROCKS)
		txn1.Start()
		var item testutil.TestItem
		txn1.Read(KVROCKS, "item1", &item)
		item.Value = "item1-updated-by-txn1"
		txn1.Write(KVROCKS, "item1", item)

		txn1.Read(KVROCKS, "item2", &item)
		item.Value = "item2-updated-by-txn1"
		txn1.Write(KVROCKS, "item2", item)

		err := txn1.Commit()
		if err != nil {
			t.Logf("txn1 commit err: %s", err)
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup(KVROCKS)
		txn2.Start()
		var item testutil.TestItem
		txn2.Read(KVROCKS, "item2", &item)
		item.Value = "item2-updated-by-txn2"
		txn2.Write(KVROCKS, "item2", item)

		txn2.Read(KVROCKS, "item1", &item)
		item.Value = "item1-updated-by-txn2"
		txn2.Write(KVROCKS, "item1", item)

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
		postTxn := NewTransactionWithSetup(KVROCKS)
		postTxn.Start()
		var item testutil.TestItem
		postTxn.Read(KVROCKS, "item1", &item)
		t.Logf("item1: %v", item)
		postTxn.Read(KVROCKS, "item2", &item)
		t.Logf("item2: %v", item)
		postTxn.Commit()
	}
}

func TestKvrocks_RepeatableReadWhenRecordDeleted(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(KVROCKS, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(KVROCKS)
	manualTxn := NewTransactionWithSetup(KVROCKS)
	txn.Start()
	manualTxn.Start()

	var person1 testutil.Person
	txn.Read(KVROCKS, "John", &person1)

	// manualTxn deletes John and commits
	manualTxn.Delete("memory", "John")
	manualTxn.Commit()

	var person2 testutil.Person
	txn.Read(KVROCKS, "John", &person2)

	// two read in txn should be the same
	if person1 != person2 {
		t.Errorf("Expected two read in txn to be the same")
	}
}

func TestKvrocks_RepeatableReadWhenRecordUpdatedTwice(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(KVROCKS, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(KVROCKS)
	manualTxn1 := NewTransactionWithSetup(KVROCKS)
	txn.Start()
	manualTxn1.Start()

	var person1 testutil.Person
	txn.Read(KVROCKS, "John", &person1)

	// manualTxn1 updates John and commits
	var manualPerson1 testutil.Person
	manualTxn1.Read(KVROCKS, "John", &manualPerson1)
	manualPerson1.Age = 31
	manualTxn1.Write(KVROCKS, "John", manualPerson1)
	manualTxn1.Commit()

	manualTxn2 := NewTransactionWithSetup(KVROCKS)
	manualTxn2.Start()
	// manualTxn updates John again and commits
	var manualPerson2 testutil.Person
	manualTxn2.Read(KVROCKS, "John", &manualPerson2)
	manualPerson2.Age = 32
	manualTxn2.Write(KVROCKS, "John", manualPerson2)
	manualTxn2.Commit()

	var person2 testutil.Person
	err := txn.Read(KVROCKS, "John", &person2)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}

	// two read in txn should be the same
	if person1 != person2 {
		t.Errorf("Expected two read in txn to be the same")
	}
}

// txn1 starts  txn2 starts
// txn1 reads John
// txn2 reads John
// txn1 writes John
// txn2 read John again
// two read in txn2 should be the same
func TestKvrocks_RepeatableReadWhenAnotherUncommitted(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(KVROCKS, "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(KVROCKS)
		txn1.Start()
		var person testutil.Person
		txn1.Read(KVROCKS, "John", &person)
		time.Sleep(50 * time.Millisecond)

		// txn1 writes John
		person.Age = 31
		txn1.Write(KVROCKS, "John", person)

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
		txn2 := NewTransactionWithSetup(KVROCKS)
		txn2.Start()
		var person1 testutil.Person
		// txn2 reads John
		txn2.Read(KVROCKS, "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 testutil.Person
		// txn2 reads John again
		txn2.Read(KVROCKS, "John", &person2)

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

// txn1 starts  txn2 starts
// txn1 reads John
// txn2 reads John
// txn1 writes John
// txn1 commits
// txn2 read John again
// two read in txn2 should be the same
func TestKvrocks_RepeatableReadWhenAnotherCommitted(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(KVROCKS, "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(KVROCKS)
		txn1.Start()
		var person testutil.Person
		txn1.Read(KVROCKS, "John", &person)
		time.Sleep(50 * time.Millisecond)

		// txn1 writes John
		person.Age = 31
		txn1.Write(KVROCKS, "John", person)

		err := txn1.Commit()
		if err != nil {
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup(KVROCKS)
		txn2.Start()
		var person1 testutil.Person
		// txn2 reads John
		txn2.Read(KVROCKS, "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 testutil.Person
		// txn2 reads John again
		txn2.Read(KVROCKS, "John", &person2)

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

func TestKvrocks_TxnAbort(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	preTxn.Start()
	expected := testutil.NewDefaultPerson()
	preTxn.Write(KVROCKS, "John", expected)
	preTxn.Commit()

	txn := NewTransactionWithSetup(KVROCKS)
	var person testutil.Person
	txn.Start()
	txn.Read(KVROCKS, "John", &person)
	person.Age = 31
	txn.Write(KVROCKS, "John", person)
	txn.Abort()

	postTxn := NewTransactionWithSetup(KVROCKS)
	postTxn.Start()
	var postPerson testutil.Person
	postTxn.Read(KVROCKS, "John", &postPerson)
	postTxn.Commit()
	if postPerson != expected {
		t.Errorf("got %v want %v", postPerson, expected)
	}
}

// TODO: WTF why this test failed when using CLI
func TestKvrocks_TxnAbortCausedByWriteConflict(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(KVROCKS, item.Value, item)
	}
	err := preTxn.Commit()
	assert.Nil(t, err)

	txn := NewTransactionWithSetup(KVROCKS)
	manualTxn := NewTransactionWithSetup(KVROCKS)
	txn.Start()
	manualTxn.Start()

	// txn reads all items and modify them
	for _, item := range testutil.InputItemList {
		var actual testutil.TestItem
		txn.Read(KVROCKS, item.Value, &actual)
		actual.Value = item.Value + "updated"
		txn.Write(KVROCKS, item.Value, actual)
	}

	// manualTxn reads one item and modify it
	var manualItem testutil.TestItem
	manualTxn.Read(KVROCKS, "item4", &manualItem)
	manualItem.Value = "item4updated"
	manualTxn.Write(KVROCKS, "item4", manualItem)
	err = manualTxn.Commit()
	assert.Nil(t, err)

	err = txn.Commit()
	if err == nil {
		t.Errorf("Expected error committing transaction")
	}

	postTxn := NewTransactionWithSetup(KVROCKS)
	postTxn.Start()
	for _, item := range testutil.InputItemList {
		var actual testutil.TestItem
		postTxn.Read(KVROCKS, item.Value, &actual)
		if item.Value != "item4" {
			assert.Equal(t, item, actual)
		} else {
			assert.Equal(t, manualItem, actual)
		}
	}
	postTxn.Commit()
}

func TestKvrocks_ConcurrentTransaction(t *testing.T) {
	preTxn := NewTransactionWithSetup(KVROCKS)
	preTxn.Start()
	person := testutil.NewDefaultPerson()
	preTxn.Write(KVROCKS, "John", person)
	err := preTxn.Commit()
	assert.NoError(t, err)

	resChan := make(chan bool)
	conNum := 100

	for i := 1; i <= conNum; i++ {
		go func(id int) {
			txn := NewTransactionWithSetup(KVROCKS)

			txn.Start()
			var person testutil.Person
			txn.Read(KVROCKS, "John", &person)
			person.Age = person.Age + id
			txn.Write(KVROCKS, "John", person)
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
func TestKvrocks_SimpleExpiredRead(t *testing.T) {
	tarMemItem := redis.RedisItem{
		Key:      "item1",
		Value:    util.ToJSONString(testutil.NewTestItem("item1")),
		TxnId:    "99",
		TxnState: config.COMMITTED,
		TValid:   time.Now().Add(-10 * time.Second),
		TLease:   time.Now().Add(-9 * time.Second),
		Version:  1,
	}

	curMemItem := redis.RedisItem{
		Key:      "item1",
		Value:    util.ToJSONString(testutil.NewTestItem("item1-prepared")),
		TxnId:    "TestKvrocks_SimpleExpiredRead",
		TxnState: config.PREPARED,
		TValid:   time.Now().Add(-5 * time.Second),
		TLease:   time.Now().Add(-4 * time.Second),
		Prev:     util.ToJSONString(tarMemItem),
		Version:  2,
	}

	conn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address: "localhost:6666",
	})
	conn.PutItem("item1", curMemItem)

	txn := NewTransactionWithSetup(KVROCKS)
	txn.Start()

	var item testutil.TestItem
	txn.Read(KVROCKS, "item1", &item)
	assert.Equal(t, testutil.NewTestItem("item1"), item)
	err := txn.Commit()
	assert.NoError(t, err)
	actual, err := conn.GetItem("item1")
	assert.NoError(t, err)
	assert.Equal(t, util.ToJSONString(tarMemItem), util.ToJSONString(actual))

}

// A complex test
// preTxn writes data to the redis database
// slowTxn read all data and write all data, but it will block when conditionalUpdate item3 (sleep 4s)
// so when slowTxn blocks, the internal state of redis database:
// item1-slow PREPARED
// item2-slow PREPARED
// item3 COMMITTED
// item4 COMMITTED
// item5 COMMITTED
// fastTxn read item3, item4, item5 and write them, then commit
// the internal state of redis database:
// item1-slow PREPARED
// item2-slow PREPARED
// item3-fast COMMITTED
// item4-fast COMMITTED
// item5-fast COMMITTED
// then, slowTxn unblocks, it starts to conditionalUpdate item3
// and it detects a version mismatch,so it aborts(with rolling back all changes)
// postTxn reads all data and verify them
// so the final internal state of redis database:
// item1 rollback to COMMITTED
// item2 rollback to COMMITTED
// item3-fast COMMITTED
// item4-fast COMMITTED
// item5-fast COMMITTED
func TestKvrocks_SlowTransactionRecordExpiredWhenPrepare_Conflict(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(KVROCKS, item.Value, item)
	}
	preTxn.Commit()

	go func() {
		slowTxn := txn.NewTransaction()
		conn := redis.NewMockRedisConnection("localhost", 6666, 2, false,
			func() error { time.Sleep(2 * time.Second); return nil })
		rds := redis.NewRedisDatastore(KVROCKS, conn)
		slowTxn.AddDatastore(rds)
		slowTxn.SetGlobalDatastore(rds)

		slowTxn.Start()
		for _, item := range testutil.InputItemList {
			var result testutil.TestItem
			slowTxn.Read(KVROCKS, item.Value, &result)
			result.Value = item.Value + "-slow"
			slowTxn.Write(KVROCKS, item.Value, result)
		}
		err := slowTxn.Commit()
		assert.EqualError(t, err, "prepare phase failed: write conflicted: the record has been modified by others")
	}()
	time.Sleep(1 * time.Second)

	// ensure the internal state of redis database
	testConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address: "localhost:6666",
	})
	testConn.Connect()
	memItem1, _ := testConn.GetItem("item1")
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-slow")), memItem1.Value)
	assert.Equal(t, memItem1.TxnState, config.PREPARED)

	memItem2, _ := testConn.GetItem("item2")
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item2-slow")), memItem2.Value)
	assert.Equal(t, memItem2.TxnState, config.PREPARED)

	memItem3, _ := testConn.GetItem("item3")
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item3")), memItem3.Value)
	assert.Equal(t, memItem3.TxnState, config.COMMITTED)

	fastTxn := NewTransactionWithSetup(KVROCKS)
	fastTxn.Start()
	for i := 2; i <= 4; i++ {
		var result testutil.TestItem
		fastTxn.Read(KVROCKS, testutil.InputItemList[i].Value, &result)
		result.Value = testutil.InputItemList[i].Value + "-fast"
		fastTxn.Write(KVROCKS, testutil.InputItemList[i].Value, result)
	}
	err := fastTxn.Commit()
	assert.NoError(t, err)

	// wait for slowTxn to complete
	time.Sleep(1 * time.Second)
	postTxn := NewTransactionWithSetup(KVROCKS)
	postTxn.Start()

	var res1 testutil.TestItem
	postTxn.Read(KVROCKS, testutil.InputItemList[0].Value, &res1)
	assert.Equal(t, testutil.InputItemList[0], res1)

	var res2 testutil.TestItem
	postTxn.Read(KVROCKS, testutil.InputItemList[1].Value, &res2)
	assert.Equal(t, testutil.InputItemList[1], res2)

	for i := 2; i <= 4; i++ {
		var res testutil.TestItem
		postTxn.Read(KVROCKS, testutil.InputItemList[i].Value, &res)
		assert.Equal(t, testutil.InputItemList[i].Value+"-fast", res.Value)
	}

	err = postTxn.Commit()
	assert.NoError(t, err)

}

// A complex test
// preTxn writes data to the redis database
// slowTxn read all data and write all data,
// but it will block when conditionalUpdate item5 (sleep 5s)
// so when slowTxn blocks, the internal state of redis database:
// item1-slow PREPARED
// item2-slow PREPARED
// item3-slow PREPARED
// item4-slow PREPARED
// item5 COMMITTED
// fastTxn read item3, item4 and write them, then commit
// (fastTxn realize item3 and item4 are expired, so it will first rollback, and write the TSR with ABORTED)
// the internal state of redis database:
// item1-slow PREPARED
// item2-slow PREPARED
// item3-fast COMMITTED
// item4-fast COMMITTED
// item5 COMMITTED
// then, slowTxn unblocks, it conditionalUpdate item5 then check the TSR state
// the TSR is marked as ABORTED, so it aborts(with rolling back all changes)
// postTxn reads all data and verify them
// so the final internal state of redis database:
// item1 rollback to COMMITTED
// item2 rollback to COMMITTED
// item3-fast COMMITTED
// item4-fast COMMITTED
// item5 COMMITTED
func TestKvrocks_SlowTransactionRecordExpiredWhenPrepare_NoConflict(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(KVROCKS, item.Value, item)
	}
	err := preTxn.Commit()
	assert.NoError(t, err)

	go func() {
		slowTxn := txn.NewTransaction()
		conn := redis.NewMockRedisConnection("localhost", 6666, 4, false,
			func() error { time.Sleep(2 * time.Second); return nil })
		rds := redis.NewRedisDatastore(KVROCKS, conn)
		slowTxn.AddDatastore(rds)
		slowTxn.SetGlobalDatastore(rds)

		slowTxn.Start()
		for _, item := range testutil.InputItemList {
			var result testutil.TestItem
			slowTxn.Read(KVROCKS, item.Value, &result)
			result.Value = item.Value + "-slow"
			slowTxn.Write(KVROCKS, item.Value, result)
		}
		err := slowTxn.Commit()
		assert.EqualError(t, err, "transaction is aborted by other transaction")
	}()

	time.Sleep(1 * time.Second)
	testConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address: "localhost:6666",
	})
	testConn.Connect()

	// all records should be PREPARED state except item5
	for _, item := range testutil.InputItemList {
		memItem, err := testConn.GetItem(item.Value)
		assert.NoError(t, err)
		if item.Value == "item5" {
			assert.Equal(t, util.ToJSONString(testutil.NewTestItem(item.Value)), memItem.Value)
			assert.Equal(t, memItem.TxnState, config.COMMITTED)
			continue
		}
		itemValue := item.Value + "-slow"
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem(itemValue)), memItem.Value)
		assert.Equal(t, memItem.TxnState, config.PREPARED)
	}

	fastTxn := NewTransactionWithSetup(KVROCKS)
	err = fastTxn.Start()
	assert.NoError(t, err)
	for i := 2; i <= 3; i++ {
		var result testutil.TestItem
		fastTxn.Read(KVROCKS, testutil.InputItemList[i].Value, &result)
		result.Value = testutil.InputItemList[i].Value + "-fast"
		fastTxn.Write(KVROCKS, testutil.InputItemList[i].Value, result)
	}
	err = fastTxn.Commit()
	assert.NoError(t, err)

	// wait for slowTxn to complete
	time.Sleep(1 * time.Second)
	postTxn := NewTransactionWithSetup(KVROCKS)
	postTxn.Start()

	var res1 testutil.TestItem
	postTxn.Read(KVROCKS, testutil.InputItemList[0].Value, &res1)
	assert.Equal(t, testutil.InputItemList[0], res1)

	var res2 testutil.TestItem
	postTxn.Read(KVROCKS, testutil.InputItemList[1].Value, &res2)
	assert.Equal(t, testutil.InputItemList[1], res2)

	for i := 2; i <= 3; i++ {
		var res testutil.TestItem
		postTxn.Read(KVROCKS, testutil.InputItemList[i].Value, &res)
		assert.Equal(t, testutil.InputItemList[i].Value+"-fast", res.Value)
	}

	var res5 testutil.TestItem
	postTxn.Read(KVROCKS, testutil.InputItemList[4].Value, &res5)
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
// item1-slow PREPARED
// item2-slow PREPARED
// item3-slow PREPARED
// item4-slow PREPARED
// item5-slow PREPARED
// testTxn read item1,item2,item3, item4
// (testTxn realize item1,item2,item3, item4 are expired, so it will first rollback, and write the TSR with ABORTED)
// the internal state of redis database:
// item1 rollback to COMMITTED
// item2 rollback to COMMITTED
// item3 rollback to COMMITTED
// item4 rollback to COMMITTED
// item5-slow PREPARED
// then, slowTxn unblocks, it fails to write the TSR, and it aborts(it tries to rollback all the items)
// so slowTxn will abort(with rolling back all changes)
// postTxn reads all data and verify them
// so the final internal state of redis database:
// item1 rollback to COMMITTED
// item2 rollback to COMMITTED
// item3 rollback to COMMITTED
// item4 rollback to COMMITTED
// item5 rollback to COMMITTED
func TestKvrocks_TransactionAbortWhenWritingTSR(t *testing.T) {

	preTxn := NewTransactionWithSetup(KVROCKS)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(KVROCKS, item.Value, item)
	}
	err := preTxn.Commit()
	if err != nil {
		t.Errorf("preTxn commit err: %s", err)
	}

	txn := txn.NewTransaction()
	conn := redis.NewMockRedisConnection("localhost", 6666, 5, true,
		func() error { time.Sleep(3 * time.Second); return errors.New("fail to write TSR") })
	mds := redis.NewRedisDatastore(KVROCKS, conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)

	txn.Start()
	for _, item := range testutil.InputItemList {
		var result testutil.TestItem
		txn.Read(KVROCKS, item.Value, &result)
		result.Value = item.Value + "-slow"
		txn.Write(KVROCKS, item.Value, result)
	}
	err = txn.Commit()
	assert.EqualError(t, err, "fail to write TSR")

	testTxn := NewTransactionWithSetup(KVROCKS)
	testTxn.Start()

	for i := 0; i <= 3; i++ {
		item := testutil.InputItemList[i]
		var memItem testutil.TestItem
		testTxn.Read(KVROCKS, item.Value, &memItem)
	}
	err = testTxn.Commit()
	assert.NoError(t, err)
	postTxn := NewTransactionWithSetup(KVROCKS)
	postTxn.Start()
	for i := 0; i <= 3; i++ {
		item := testutil.InputItemList[i]
		var memItem testutil.TestItem
		postTxn.Read(KVROCKS, item.Value, &memItem)
		assert.Equal(t, item.Value, memItem.Value)
	}

	memItem, err := conn.GetItem("item5")
	assert.NoError(t, err)
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item5")), memItem.Value)
	assert.Equal(t, config.COMMITTED, memItem.TxnState)
}
