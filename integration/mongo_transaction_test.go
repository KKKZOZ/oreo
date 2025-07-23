package integration

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/mock"
	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

func TestMongo_TxnWrite(t *testing.T) {
	txn1 := NewTransactionWithSetup(MONGO)
	expected := testutil.NewDefaultPerson()

	// clear the data
	conn := NewConnectionWithSetup(MONGO)
	conn.Delete("John")

	// Txn1 writes the record
	err := txn1.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	err = txn1.Write(MONGO, "John", expected)
	if err != nil {
		t.Errorf("Error writing record: %s", err)
	}
	err = txn1.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

	txn2 := NewTransactionWithSetup(MONGO)

	// Txn2 reads the record
	err = txn2.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var actual testutil.Person
	err = txn2.Read(MONGO, "John", &actual)
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

func TestMongo_ReadOwnWrite(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(MONGO, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(MONGO)

	// Txn reads the record
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var person testutil.Person
	err = txn.Read(MONGO, "John", &person)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}

	expected := person
	expected.Age = 31
	txn.Write(MONGO, "John", expected)

	var actual testutil.Person
	err = txn.Read(MONGO, "John", &actual)
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

func TestMongo_SingleKeyWriteConflict(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(MONGO, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(MONGO)
	txn.Start()
	var person testutil.Person
	txn.Read(MONGO, "John", &person)
	person.Age = 31
	txn.Write(MONGO, "John", person)
	var anotherPerson testutil.Person
	txn.Read(MONGO, "John", &anotherPerson)

	if person != anotherPerson {
		t.Errorf("Expected two read to be the same")
	}
	txn.Commit()

	postTxn := NewTransactionWithSetup(MONGO)
	postTxn.Start()
	var postPerson testutil.Person
	postTxn.Read(MONGO, "John", &postPerson)
	if postPerson != person {
		t.Errorf("got %v want %v", postPerson, person)
	}
}

func TestMongo_MultileKeyWriteConflict(t *testing.T) {
	// clear the data
	conn := NewConnectionWithSetup(MONGO)
	conn.Delete("item1")
	conn.Delete("item2")

	preTxn := NewTransactionWithSetup(MONGO)
	item1 := testutil.NewTestItem("item1")
	item2 := testutil.NewTestItem("item2")
	preTxn.Start()
	preTxn.Write(MONGO, "item1", item1)
	preTxn.Write(MONGO, "item2", item2)
	err := preTxn.Commit()
	assert.Nil(t, err)

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(MONGO)
		txn1.Start()
		var item testutil.TestItem
		txn1.Read(MONGO, "item1", &item)
		item.Value = "item1-updated-by-txn1"
		txn1.Write(MONGO, "item1", item)

		txn1.Read(MONGO, "item2", &item)
		item.Value = "item2-updated-by-txn1"
		txn1.Write(MONGO, "item2", item)

		err := txn1.Commit()
		if err != nil {
			t.Logf("txn1 commit err: %s", err)
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup(MONGO)
		txn2.Start()
		var item testutil.TestItem
		txn2.Read(MONGO, "item2", &item)
		item.Value = "item2-updated-by-txn2"
		txn2.Write(MONGO, "item2", item)

		txn2.Read(MONGO, "item1", &item)
		item.Value = "item1-updated-by-txn2"
		txn2.Write(MONGO, "item1", item)

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
		postTxn := NewTransactionWithSetup(MONGO)
		postTxn.Start()
		var item testutil.TestItem
		postTxn.Read(MONGO, "item1", &item)
		t.Logf("item1: %v", item)
		postTxn.Read(MONGO, "item2", &item)
		t.Logf("item2: %v", item)
		postTxn.Commit()
	}
}

func TestMongo_RepeatableReadWhenRecordDeleted(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(MONGO, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(MONGO)
	manualTxn := NewTransactionWithSetup(MONGO)
	txn.Start()
	manualTxn.Start()

	var person1 testutil.Person
	txn.Read(MONGO, "John", &person1)

	// manualTxn deletes John and commits
	manualTxn.Delete("memory", "John")
	manualTxn.Commit()

	var person2 testutil.Person
	txn.Read(MONGO, "John", &person2)

	// two read in txn should be the same
	if person1 != person2 {
		t.Errorf("Expected two read in txn to be the same")
	}
}

func TestMongo_RepeatableReadWhenRecordUpdatedTwice(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(MONGO, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(MONGO)
	manualTxn1 := NewTransactionWithSetup(MONGO)
	txn.Start()
	manualTxn1.Start()

	var person1 testutil.Person
	txn.Read(MONGO, "John", &person1)

	// manualTxn1 updates John and commits
	var manualPerson1 testutil.Person
	manualTxn1.Read(MONGO, "John", &manualPerson1)
	manualPerson1.Age = 31
	manualTxn1.Write(MONGO, "John", manualPerson1)
	manualTxn1.Commit()

	manualTxn2 := NewTransactionWithSetup(MONGO)
	manualTxn2.Start()
	// manualTxn updates John again and commits
	var manualPerson2 testutil.Person
	manualTxn2.Read(MONGO, "John", &manualPerson2)
	manualPerson2.Age = 32
	manualTxn2.Write(MONGO, "John", manualPerson2)
	manualTxn2.Commit()

	var person2 testutil.Person
	err := txn.Read(MONGO, "John", &person2)
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
func TestMongo_RepeatableReadWhenAnotherUncommitted(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(MONGO, "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(MONGO)
		txn1.Start()
		var person testutil.Person
		txn1.Read(MONGO, "John", &person)
		time.Sleep(50 * time.Millisecond)

		// txn1 writes John
		person.Age = 31
		txn1.Write(MONGO, "John", person)

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
		txn2 := NewTransactionWithSetup(MONGO)
		txn2.Start()
		var person1 testutil.Person
		// txn2 reads John
		txn2.Read(MONGO, "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 testutil.Person
		// txn2 reads John again
		txn2.Read(MONGO, "John", &person2)

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
func TestMongo_RepeatableReadWhenAnotherCommitted(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(MONGO, "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(MONGO)
		txn1.Start()
		var person testutil.Person
		txn1.Read(MONGO, "John", &person)
		time.Sleep(50 * time.Millisecond)

		// txn1 writes John
		person.Age = 31
		txn1.Write(MONGO, "John", person)

		err := txn1.Commit()
		if err != nil {
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup(MONGO)
		txn2.Start()
		var person1 testutil.Person
		// txn2 reads John
		txn2.Read(MONGO, "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 testutil.Person
		// txn2 reads John again
		txn2.Read(MONGO, "John", &person2)

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

func TestMongo_TxnAbort(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	preTxn.Start()
	expected := testutil.NewDefaultPerson()
	preTxn.Write(MONGO, "John", expected)
	preTxn.Commit()

	txn := NewTransactionWithSetup(MONGO)
	var person testutil.Person
	txn.Start()
	txn.Read(MONGO, "John", &person)
	person.Age = 31
	txn.Write(MONGO, "John", person)
	txn.Abort()

	postTxn := NewTransactionWithSetup(MONGO)
	postTxn.Start()
	var postPerson testutil.Person
	postTxn.Read(MONGO, "John", &postPerson)
	postTxn.Commit()
	if postPerson != expected {
		t.Errorf("got %v want %v", postPerson, expected)
	}
}

// TODO: WTF why this test failed when using CLI
func TestMongo_TxnAbortCausedByWriteConflict(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(MONGO, item.Value, item)
	}
	err := preTxn.Commit()
	assert.Nil(t, err)

	txn := NewTransactionWithSetup(MONGO)
	manualTxn := NewTransactionWithSetup(MONGO)
	txn.Start()
	manualTxn.Start()

	// txn reads all items and modify them
	for _, item := range testutil.InputItemList {
		var actual testutil.TestItem
		txn.Read(MONGO, item.Value, &actual)
		actual.Value = item.Value + "updated"
		txn.Write(MONGO, item.Value, actual)
	}

	// manualTxn reads one item and modify it
	var manualItem testutil.TestItem
	manualTxn.Read(MONGO, "item4", &manualItem)
	manualItem.Value = "item4updated"
	manualTxn.Write(MONGO, "item4", manualItem)
	err = manualTxn.Commit()
	assert.Nil(t, err)

	err = txn.Commit()
	if err == nil {
		t.Errorf("Expected error committing transaction")
	}

	postTxn := NewTransactionWithSetup(MONGO)
	postTxn.Start()
	for _, item := range testutil.InputItemList {
		var actual testutil.TestItem
		postTxn.Read(MONGO, item.Value, &actual)
		if item.Value != "item4" {
			assert.Equal(t, item, actual)
		} else {
			assert.Equal(t, manualItem, actual)
		}
	}
	postTxn.Commit()
}

func TestMongo_ConcurrentTransaction(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	preTxn.Start()
	person := testutil.NewDefaultPerson()
	preTxn.Write(MONGO, "John", person)
	err := preTxn.Commit()
	assert.NoError(t, err)

	resChan := make(chan bool)
	conNum := 10

	conn := NewConnectionWithSetup(MONGO)

	for i := 1; i <= conNum; i++ {
		go func(id int) {
			txn := txn.NewTransaction()
			rds := mongo.NewMongoDatastore(MONGO, conn)
			txn.AddDatastore(rds)
			txn.SetGlobalDatastore(rds)
			txn.Start()
			var person testutil.Person
			txn.Read(MONGO, "John", &person)
			person.Age = person.Age + id
			txn.Write(MONGO, "John", person)
			time.Sleep(500 * time.Millisecond)
			err = txn.Commit()
			if err != nil {
				fmt.Printf("txn commit err: %s\n", err)
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
func TestMongo_SimpleExpiredRead(t *testing.T) {
	tarMemItem := &mongo.MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1")),
		MGroupKeyList: "TestMongo_SimpleExpiredRead1",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-10 * time.Second).UnixMicro(),
		MTLease:       time.Now().Add(-9 * time.Second),
		MVersion:      "1",
	}

	curMemItem := &mongo.MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-prepared")),
		MGroupKeyList: "TestMongo_SimpleExpiredRead2",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
		MTLease:       time.Now().Add(-4 * time.Second),
		MPrev:         util.ToJSONString(tarMemItem),
		MVersion:      "2",
	}

	conn := NewConnectionWithSetup(MONGO)
	conn.PutItem("item1", curMemItem)

	txn := NewTransactionWithSetup(MONGO)
	txn.Start()

	var item testutil.TestItem
	txn.Read(MONGO, "item1", &item)
	assert.Equal(t, testutil.NewTestItem("item1"), item)
	err := txn.Commit()
	assert.NoError(t, err)
	actual, err := conn.GetItem("item1")
	assert.NoError(t, err)
	tarMemItem.MVersion = "3"
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
func TestMongo_SlowTransactionRecordExpiredWhenPrepare_Conflict(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(MONGO, item.Value, item)
	}
	preTxn.Commit()

	go func() {
		slowTxn := NewTransactionWithMockConn(MONGO, 2, false,
			0, func() error { time.Sleep(2 * time.Second); return nil })

		slowTxn.Start()
		for _, item := range testutil.InputItemList {
			var result testutil.TestItem
			slowTxn.Read(MONGO, item.Value, &result)
			result.Value = item.Value + "-slow"
			slowTxn.Write(MONGO, item.Value, result)
		}
		err := slowTxn.Commit()
		assert.EqualError(t, err, "prepare phase failed: version mismatch")
	}()
	time.Sleep(1 * time.Second)

	// ensure the internal state of redis database
	testConn := NewConnectionWithSetup(MONGO)
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

	fastTxn := NewTransactionWithSetup(MONGO)
	fastTxn.Start()
	for i := 2; i <= 4; i++ {
		var result testutil.TestItem
		fastTxn.Read(MONGO, testutil.InputItemList[i].Value, &result)
		result.Value = testutil.InputItemList[i].Value + "-fast"
		fastTxn.Write(MONGO, testutil.InputItemList[i].Value, result)
	}
	err := fastTxn.Commit()
	assert.NoError(t, err)

	// wait for slowTxn to complete
	time.Sleep(1 * time.Second)
	postTxn := NewTransactionWithSetup(MONGO)
	postTxn.Start()

	var res1 testutil.TestItem
	postTxn.Read(MONGO, testutil.InputItemList[0].Value, &res1)
	assert.Equal(t, testutil.InputItemList[0], res1)

	var res2 testutil.TestItem
	postTxn.Read(MONGO, testutil.InputItemList[1].Value, &res2)
	assert.Equal(t, testutil.InputItemList[1], res2)

	for i := 2; i <= 4; i++ {
		var res testutil.TestItem
		postTxn.Read(MONGO, testutil.InputItemList[i].Value, &res)
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
func TestMongo_SlowTransactionRecordExpiredWhenPrepare_NoConflict(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(MONGO, item.Value, item)
	}
	err := preTxn.Commit()
	assert.NoError(t, err)

	go func() {
		slowTxn := NewTransactionWithMockConn(MONGO, 4, false,
			0, func() error { time.Sleep(2 * time.Second); return nil })
		slowTxn.Start()
		for _, item := range testutil.InputItemList {
			var result testutil.TestItem
			slowTxn.Read(MONGO, item.Value, &result)
			result.Value = item.Value + "-slow"
			slowTxn.Write(MONGO, item.Value, result)
		}
		err := slowTxn.Commit()
		assert.EqualError(t, err, "transaction is aborted by other transaction")
	}()

	time.Sleep(1 * time.Second)
	testConn := NewConnectionWithSetup(MONGO)
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

	fastTxn := NewTransactionWithSetup(MONGO)
	err = fastTxn.Start()
	assert.NoError(t, err)
	for i := 2; i <= 3; i++ {
		var result testutil.TestItem
		fastTxn.Read(MONGO, testutil.InputItemList[i].Value, &result)
		result.Value = testutil.InputItemList[i].Value + "-fast"
		fastTxn.Write(MONGO, testutil.InputItemList[i].Value, result)
	}
	err = fastTxn.Commit()
	assert.NoError(t, err)

	// wait for slowTxn to complete
	time.Sleep(1 * time.Second)
	postTxn := NewTransactionWithSetup(MONGO)
	postTxn.Start()

	var res1 testutil.TestItem
	postTxn.Read(MONGO, testutil.InputItemList[0].Value, &res1)
	assert.Equal(t, testutil.InputItemList[0], res1)

	var res2 testutil.TestItem
	postTxn.Read(MONGO, testutil.InputItemList[1].Value, &res2)
	assert.Equal(t, testutil.InputItemList[1], res2)

	for i := 2; i <= 3; i++ {
		var res testutil.TestItem
		postTxn.Read(MONGO, testutil.InputItemList[i].Value, &res)
		assert.Equal(t, testutil.InputItemList[i].Value+"-fast", res.Value)
	}

	var res5 testutil.TestItem
	postTxn.Read(MONGO, testutil.InputItemList[4].Value, &res5)
	assert.Equal(t, testutil.InputItemList[4].Value, res5.Value)

	err = postTxn.Commit()
	assert.NoError(t, err)

	// for debug only
	// time.Sleep(10 * time.Second)
}

// A complex test
// preTxn writes data to the redis database
// slowTxn reads all data and write all data,
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
func TestMongo_TransactionAbortWhenWritingTSR(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(MONGO, item.Value, item)
	}
	err := preTxn.Commit()
	if err != nil {
		t.Errorf("preTxn commit err: %s", err)
	}

	txn := NewTransactionWithMockConn(MONGO, 5, true,
		0, func() error { time.Sleep(3 * time.Second); return errors.New("fail to write TSR") })
	txn.Start()
	for _, item := range testutil.InputItemList {
		var result testutil.TestItem
		txn.Read(MONGO, item.Value, &result)
		result.Value = item.Value + "-slow"
		txn.Write(MONGO, item.Value, result)
	}
	err = txn.Commit()
	assert.EqualError(t, err, "fail to write TSR")

	testTxn := NewTransactionWithSetup(MONGO)
	testTxn.Start()

	for i := 0; i <= 3; i++ {
		item := testutil.InputItemList[i]
		var memItem testutil.TestItem
		testTxn.Read(MONGO, item.Value, &memItem)
	}
	err = testTxn.Commit()
	assert.NoError(t, err)
	postTxn := NewTransactionWithSetup(MONGO)
	postTxn.Start()
	for i := 0; i <= 3; i++ {
		item := testutil.InputItemList[i]
		var memItem testutil.TestItem
		postTxn.Read(MONGO, item.Value, &memItem)
		assert.Equal(t, item.Value, memItem.Value)
	}

	conn := NewConnectionWithSetup(MONGO)
	memItem, err := conn.GetItem("item5")
	assert.NoError(t, err)
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item5")), memItem.Value())
	assert.Equal(t, config.COMMITTED, memItem.TxnState())
}

func TestMongo_LinkedRecord(t *testing.T) {
	t.Cleanup(func() {
		config.Config.MaxRecordLength = 2
	})

	t.Run("commit time less than MaxLen", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(MONGO)
		preTxn.Start()
		person := testutil.NewDefaultPerson()
		preTxn.Write(MONGO, "John", person)
		err := preTxn.Commit()
		assert.NoError(t, err)

		slowTxn := NewTransactionWithSetup(MONGO)
		slowTxn.Start()

		config.Config.MaxRecordLength = 4
		// 1+2=3 < 4, including origin
		commitTime := 2
		for i := 1; i <= commitTime; i++ {
			txn := NewTransactionWithSetup(MONGO)
			txn.Start()
			var p testutil.Person
			txn.Read(MONGO, "John", &p)
			p.Age = p.Age + 1
			txn.Write(MONGO, "John", p)
			err = txn.Commit()
			assert.NoError(t, err)
		}

		var p testutil.Person
		err = slowTxn.Read(MONGO, "John", &p)
		assert.NoError(t, err)
		assert.Equal(t, 30, p.Age)
	})

	t.Run("commit time equals MaxLen", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(MONGO)
		preTxn.Start()
		person := testutil.NewDefaultPerson()
		preTxn.Write(MONGO, "John", person)
		err := preTxn.Commit()
		assert.NoError(t, err)

		slowTxn := NewTransactionWithSetup(MONGO)
		slowTxn.Start()

		config.Config.MaxRecordLength = 4
		// 1+3=4 == 4, including origin
		commitTime := 3
		for i := 1; i <= commitTime; i++ {
			txn := NewTransactionWithSetup(MONGO)
			txn.Start()
			var p testutil.Person
			txn.Read(MONGO, "John", &p)
			p.Age = p.Age + 1
			txn.Write(MONGO, "John", p)
			err = txn.Commit()
			assert.NoError(t, err)
		}

		var p testutil.Person
		err = slowTxn.Read(MONGO, "John", &p)
		assert.NoError(t, err)
		assert.Equal(t, 30, p.Age)
	})

	t.Run("commit times bigger than MaxLen", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(MONGO)
		preTxn.Start()
		person := testutil.NewDefaultPerson()
		preTxn.Write(MONGO, "John", person)
		err := preTxn.Commit()
		assert.NoError(t, err)

		slowTxn := NewTransactionWithSetup(MONGO)
		slowTxn.Start()

		config.Config.MaxRecordLength = 4
		// 1+4=5 > 4, including origin
		commitTime := 4
		for i := 1; i <= commitTime; i++ {
			txn := NewTransactionWithSetup(MONGO)
			txn.Start()
			var p testutil.Person
			txn.Read(MONGO, "John", &p)
			p.Age = p.Age + 1
			txn.Write(MONGO, "John", p)
			err = txn.Commit()
			assert.NoError(t, err)
		}

		var p testutil.Person
		err = slowTxn.Read(MONGO, "John", &p)
		assert.EqualError(t, err, "key not found")
	})
}

func TestMongo_RollbackConflict(t *testing.T) {
	// there is a broken item
	//   - txnA reads the item, decides to roll back
	//   - txnB reads the item, decides to roll back
	//   - txnB rollbacks the item
	//   - txnB update the item and commit
	//   - txnA rollbacks the item -> should fail
	t.Run("the broken item has a valid Prev field", func(t *testing.T) {
		conn := NewConnectionWithSetup(MONGO)

		redisItem1 := &mongo.MongoItem{
			MKey:          "item1",
			MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
			MGroupKeyList: "TestMongo_RollbackConflict1",
			MTxnState:     config.COMMITTED,
			MTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
			MTLease:       time.Now().Add(-4 * time.Second),
			MLinkedLen:    1,
			MVersion:      "1",
		}
		redisItem2 := &mongo.MongoItem{
			MKey:          "item1",
			MValue:        util.ToJSONString(testutil.NewTestItem("item1-broken")),
			MGroupKeyList: "TestMongo_RollbackConflict2",
			MTxnState:     config.PREPARED,
			MTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
			MTLease:       time.Now().Add(-4 * time.Second),
			MPrev:         util.ToJSONString(redisItem1),
			MLinkedLen:    2,
			MVersion:      "2",
		}
		conn.PutItem("item1", redisItem2)

		go func() {
			time.Sleep(100 * time.Millisecond)
			txnA := NewTransactionWithMockConn(MONGO, 0, false,
				0, func() error { time.Sleep(2 * time.Second); return nil })
			txnA.Start()

			var item testutil.TestItem
			err := txnA.Read(MONGO, "item1", &item)
			assert.NotNil(t, err)
		}()

		time.Sleep(100 * time.Millisecond)
		txnB := NewTransactionWithSetup(MONGO)
		txnB.Start()
		var item testutil.TestItem
		err := txnB.Read(MONGO, "item1", &item)
		assert.NoError(t, err)
		assert.Equal(t, testutil.NewTestItem("item1-pre"), item)
		item.Value = "item1-B"
		txnB.Write(MONGO, "item1", item)
		err = txnB.Commit()
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)
		resItem, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())
		redisItem1.MVersion = "3"
		assert.Equal(t, util.ToJSONString(redisItem1), resItem.Prev())
	})

	// there is a broken item
	//   - txnA reads the item, decides to roll back
	//   - txnB reads the item, decides to roll back
	//   - txnB rollbacks the item
	//   - txnB update the item and commit
	//   - txnA rollbacks the item -> should fail
	t.Run("the broken item has an empty Prev field", func(t *testing.T) {
		conn := NewConnectionWithSetup(MONGO)

		redisItem2 := &mongo.MongoItem{
			MKey:          "item1",
			MValue:        util.ToJSONString(testutil.NewTestItem("item1-broken")),
			MGroupKeyList: "TestMongo_RollbackConflict2-emptyField",
			MTxnState:     config.PREPARED,
			MTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
			MTLease:       time.Now().Add(-4 * time.Second),
			MPrev:         "",
			MLinkedLen:    1,
			MVersion:      "1",
		}
		conn.PutItem("item1", redisItem2)

		go func() {
			time.Sleep(100 * time.Millisecond)
			txnA := NewTransactionWithMockConn(MONGO, 0, false,
				0, func() error { time.Sleep(2 * time.Second); return nil })
			txnA.Start()

			var item testutil.TestItem
			err := txnA.Read(MONGO, "item1", &item)
			assert.NotNil(t, err)
		}()

		time.Sleep(100 * time.Millisecond)
		txnB := NewTransactionWithSetup(MONGO)
		txnB.Start()
		var item testutil.TestItem
		err := txnB.Read(MONGO, "item1", &item)
		assert.EqualError(t, err, txn.KeyNotFound.Error())

		item.Value = "item1-B"
		txnB.Write(MONGO, "item1", item)
		err = txnB.Commit()
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)
		resItem, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())
		redisItem2.MIsDeleted = true
		redisItem2.MTxnState = config.COMMITTED
		redisItem2.MVersion = util.AddToString(redisItem2.MVersion, 1)
		assert.Equal(t, util.ToJSONString(redisItem2), resItem.Prev())
	})
}

// there is a broken item
//   - txnA reads the item, decides to roll forward
//   - txnB reads the item, decides to roll forward
//   - txnB rolls forward the item
//   - txnB update the item and commit
//   - txnA rolls forward the item -> should fail
func TestMongo_RollForwardConflict(t *testing.T) {
	conn := NewConnectionWithSetup(MONGO)

	redisItem1 := &mongo.MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
		MGroupKeyList: "TestMongo_RollForwardConflict1",
		MTxnState:     config.COMMITTED,
		MTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
		MTLease:       time.Now().Add(-4 * time.Second),
		MVersion:      "1",
	}
	redisItem2 := &mongo.MongoItem{
		MKey:          "item1",
		MValue:        util.ToJSONString(testutil.NewTestItem("item1-broken")),
		MGroupKeyList: "TestMongo_RollForwardConflict2",
		MTxnState:     config.PREPARED,
		MTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
		MTLease:       time.Now().Add(-4 * time.Second),
		MPrev:         util.ToJSONString(redisItem1),
		MLinkedLen:    2,
		MVersion:      "2",
	}
	conn.PutItem("item1", redisItem2)
	conn.Put("TestMongo_RollForwardConflict2", config.COMMITTED)

	go func() {
		time.Sleep(100 * time.Millisecond)
		txnA := NewTransactionWithMockConn(MONGO, 0, false,
			0, func() error { time.Sleep(2 * time.Second); return nil })
		txnA.Start()
		var item testutil.TestItem
		err := txnA.Read(MONGO, "item1", &item)
		assert.NotNil(t, err)
	}()

	time.Sleep(100 * time.Millisecond)
	txnB := NewTransactionWithSetup(MONGO)
	txnB.Start()
	var item testutil.TestItem
	txnB.Read(MONGO, "item1", &item)
	assert.Equal(t, testutil.NewTestItem("item1-broken"), item)
	item.Value = "item1-B"
	txnB.Write(MONGO, "item1", item)
	err := txnB.Commit()
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)
	resItem, err := conn.GetItem("item1")
	assert.NoError(t, err)
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())
	redisItem2.MTxnState = config.COMMITTED
	redisItem2.MPrev = ""
	redisItem2.MLinkedLen = 1
	redisItem2.MVersion = "3"
	assert.Equal(t, util.ToJSONString(redisItem2), resItem.Prev())
}

func TestMongo_ConcurrentDirectWrite(t *testing.T) {
	conn := NewConnectionWithSetup(MONGO)
	conn.Delete("item1")

	conNumber := 5
	mu := sync.Mutex{}
	globalId := 0
	resChan := make(chan bool)
	for i := 1; i <= conNumber; i++ {
		go func(id int) {
			txn := NewTransactionWithSetup(MONGO)
			item := testutil.NewTestItem("item1-" + strconv.Itoa(id))
			txn.Start()
			txn.Write(MONGO, "item1", item)
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

	postTxn := NewTransactionWithSetup(MONGO)
	postTxn.Start()
	var item testutil.TestItem
	postTxn.Read(MONGO, "item1", &item)
	assert.Equal(t, testutil.NewTestItem("item1-"+strconv.Itoa(globalId)), item)
	// assert.Equal(t, 0, globalId)
	t.Logf("item: %v", item)
}

func TestMongo_TxnDelete(t *testing.T) {
	preTxn := NewTransactionWithSetup(MONGO)
	preTxn.Start()
	item := testutil.NewTestItem("item1-pre")
	preTxn.Write(MONGO, "item1", item)
	err := preTxn.Commit()
	assert.NoError(t, err)

	txn := NewTransactionWithSetup(MONGO)
	txn.Start()
	var item1 testutil.TestItem
	txn.Read(MONGO, "item1", &item1)
	txn.Delete(MONGO, "item1")
	err = txn.Commit()
	assert.NoError(t, err)

	postTxn := NewTransactionWithSetup(MONGO)
	postTxn.Start()
	var item2 testutil.TestItem
	err = postTxn.Read(MONGO, "item1", &item2)
	assert.EqualError(t, err, "key not found")
}

func TestMongo_PreventLostUpdatesValidation(t *testing.T) {
	t.Run("Case 1-1(with read): The target record has been updated by the concurrent transaction",
		func(t *testing.T) {
			preTxn := NewTransactionWithSetup(MONGO)
			preTxn.Start()
			item := testutil.NewTestItem("item1-pre")
			preTxn.Write(MONGO, "item1", item)
			err := preTxn.Commit()
			assert.NoError(t, err)

			txnA := NewTransactionWithSetup(MONGO)
			txnA.Start()
			txnB := NewTransactionWithSetup(MONGO)
			txnB.Start()

			var itemA testutil.TestItem
			txnA.Read(MONGO, "item1", &itemA)
			itemA.Value = "item1-A"
			txnA.Write(MONGO, "item1", itemA)
			err = txnA.Commit()
			assert.NoError(t, err)

			var itemB testutil.TestItem
			txnB.Read(MONGO, "item1", &itemB)
			itemB.Value = "item1-B"
			txnB.Write(MONGO, "item1", itemB)
			err = txnB.Commit()
			assert.EqualError(t, err,
				"prepare phase failed: version mismatch")

			postTxn := NewTransactionWithSetup(MONGO)
			postTxn.Start()
			var item1 testutil.TestItem
			postTxn.Read(MONGO, "item1", &item1)
			assert.Equal(t, itemA, item1)
		})

	t.Run(
		"Case 1-2(without read): The target record has been updated by the concurrent transaction",
		func(t *testing.T) {
			preTxn := NewTransactionWithSetup(MONGO)
			preTxn.Start()
			item := testutil.NewTestItem("item1-pre")
			preTxn.Write(MONGO, "item1", item)
			err := preTxn.Commit()
			assert.NoError(t, err)

			txnA := NewTransactionWithSetup(MONGO)
			txnA.Start()
			txnB := NewTransactionWithSetup(MONGO)
			txnB.Start()

			itemA := testutil.NewTestItem("item1-A")
			txnA.Write(MONGO, "item1", itemA)
			err = txnA.Commit()
			assert.NoError(t, err)

			itemB := testutil.NewTestItem("item1-B")
			txnB.Write(MONGO, "item1", itemB)
			err = txnB.Commit()
			assert.EqualError(t, err,
				"prepare phase failed: version mismatch")

			postTxn := NewTransactionWithSetup(MONGO)
			postTxn.Start()
			var item1 testutil.TestItem
			postTxn.Read(MONGO, "item1", &item1)
			assert.Equal(t, itemA, item1)
		},
	)

	t.Run("Case 2-1(with read): There is no conflict", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(MONGO)
		preTxn.Start()
		item := testutil.NewTestItem("item1-pre")
		preTxn.Write(MONGO, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txnA := NewTransactionWithSetup(MONGO)
		txnA.Start()
		var itemA testutil.TestItem
		txnA.Read(MONGO, "item1", &itemA)
		itemA.Value = "item1-A"
		txnA.Write(MONGO, "item1", itemA)
		err = txnA.Commit()
		assert.NoError(t, err)

		txnB := NewTransactionWithSetup(MONGO)
		txnB.Start()
		var itemB testutil.TestItem
		txnB.Read(MONGO, "item1", &itemB)
		itemB.Value = "item1-B"
		txnB.Write(MONGO, "item1", itemB)
		err = txnB.Commit()
		assert.NoError(t, err)

		postTxn := NewTransactionWithSetup(MONGO)
		postTxn.Start()
		var item1 testutil.TestItem
		postTxn.Read(MONGO, "item1", &item1)
		assert.Equal(t, itemB, item1)
	})

	t.Run("Case 2-2(without read): There is no conflict", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(MONGO)
		preTxn.Start()
		item := testutil.NewTestItem("item1-pre")
		preTxn.Write(MONGO, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txnA := NewTransactionWithSetup(MONGO)
		txnA.Start()
		itemA := testutil.NewTestItem("item1-A")
		txnA.Write(MONGO, "item1", itemA)
		err = txnA.Commit()
		assert.NoError(t, err)

		txnB := NewTransactionWithSetup(MONGO)
		txnB.Start()
		itemB := testutil.NewTestItem("item1-B")
		txnB.Write(MONGO, "item1", itemB)
		err = txnB.Commit()
		assert.NoError(t, err)

		postTxn := NewTransactionWithSetup(MONGO)
		postTxn.Start()
		var item1 testutil.TestItem
		postTxn.Read(MONGO, "item1", &item1)
		assert.Equal(t, itemB, item1)
	})
}

func TestMongo_RepeatableReadWhenDirtyRead(t *testing.T) {
	t.Run("the prepared item has a valid Prev", func(t *testing.T) {
		config.Config.LeaseTime = 3000 * time.Millisecond

		preTxn := NewTransactionWithSetup(MONGO)
		preTxn.Start()
		item := testutil.NewTestItem("item1-pre")
		preTxn.Write(MONGO, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txnA := NewTransactionWithMockConn(MONGO, 1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		txnB := NewTransactionWithSetup(MONGO)
		txnA.Start()
		txnB.Start()

		go func() {
			itemA := testutil.NewTestItem("item1-A")
			txnA.Write(MONGO, "item1", itemA)
			err = txnA.Commit()
			assert.NoError(t, err)
		}()

		// wait for txnA to write item into database
		time.Sleep(500 * time.Millisecond)
		var itemB testutil.TestItem
		err = txnB.Read(MONGO, "item1", &itemB)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre", itemB.Value)

		time.Sleep(2 * time.Second)
		err = txnB.Read(MONGO, "item1", &itemB)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre", itemB.Value)
	})

	t.Run("the prepared item has an empty Prev", func(t *testing.T) {
		config.Config.LeaseTime = 3000 * time.Millisecond

		testConn := NewConnectionWithSetup(MONGO)
		testConn.Delete("item1")

		txnA := NewTransactionWithMockConn(MONGO, 1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		txnB := NewTransactionWithSetup(MONGO)
		txnA.Start()
		txnB.Start()

		go func() {
			itemA := testutil.NewTestItem("item1-A")
			txnA.Write(MONGO, "item1", itemA)
			err := txnA.Commit()
			assert.NoError(t, err)
		}()

		// wait for txnA to write item into database
		time.Sleep(500 * time.Millisecond)
		var itemB testutil.TestItem
		err := txnB.Read(MONGO, "item1", &itemB)
		assert.EqualError(t, err, txn.KeyNotFound.Error())

		time.Sleep(2 * time.Second)

		// make sure txnA has committed
		resItem, err := testConn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-A")), resItem.Value())

		err = txnB.Read(MONGO, "item1", &itemB)
		assert.EqualError(t, err, txn.KeyNotFound.Error())

		// post check
		postTxn := NewTransactionWithSetup(MONGO)
		postTxn.Start()

		var itemPost testutil.TestItem
		err = postTxn.Read(MONGO, "item1", &itemPost)
		assert.NoError(t, err)
		assert.Equal(t, "item1-A", itemPost.Value)
	})
}

func TestMongo_DeleteTimingProblems(t *testing.T) {
	// conn Puts an item with an empty Prev field
	//  - txnA reads the item, decides to delete
	//  - txnB reads the item, updates and commmits
	//  - txnA tries to delete the item -> should fail
	t.Run("the item has an empty Prev field", func(t *testing.T) {
		testConn := NewConnectionWithSetup(MONGO)
		dbItem := &mongo.MongoItem{
			MKey:          "item1",
			MValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
			MGroupKeyList: "TestMongo_DeleteTimingProblems",
			MTxnState:     config.COMMITTED,
			MTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
			MTLease:       time.Now().Add(-4 * time.Second),
			MVersion:      "1",
		}
		testConn.PutItem("item1", dbItem)

		txnA := NewTransactionWithMockConn(MONGO, 0, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		txnB := NewTransactionWithSetup(MONGO)
		txnA.Start()
		txnB.Start()

		go func() {
			var item testutil.TestItem
			txnA.Read(MONGO, "item1", &item)
			txnA.Delete(MONGO, "item1")
			err := txnA.Commit()
			assert.NotNil(t, err)
			// assert.NoError(t, err)
		}()

		time.Sleep(200 * time.Millisecond)
		var itemB testutil.TestItem
		txnB.Read(MONGO, "item1", &itemB)
		txnB.Delete(MONGO, "item1")
		itemB.Value = "item1-B"
		txnB.Write(MONGO, "item1", itemB)
		err := txnB.Commit()
		assert.NoError(t, err)

		time.Sleep(1 * time.Second)

		// post check
		resItem, err := testConn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())
	})
}

func TestMongo_VisibilityResults(t *testing.T) {
	t.Run("a normal chain", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(MONGO)
		preTxn.Start()
		item := testutil.NewTestItem("item1-V0")
		preTxn.Write(MONGO, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		chainNum := 5

		for i := 1; i <= chainNum; i++ {
			time.Sleep(10 * time.Millisecond)
			txn := NewTransactionWithSetup(MONGO)
			txn.Start()
			var item testutil.TestItem
			txn.Read(MONGO, "item1", &item)
			assert.Equal(t, "item1-V"+strconv.Itoa(i-1), item.Value)
			item.Value = "item1-V" + strconv.Itoa(i)
			txn.Write(MONGO, "item1", item)
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

func TestMongo_ReadModifyWritePattern(t *testing.T) {
	t.Run("when X = 1", func(t *testing.T) {
		X := 1
		preTxn := NewTransactionWithSetup(MONGO)
		preTxn.Start()
		dbItem := testutil.NewTestItem("item1-pre")
		preTxn.Write(MONGO, "item1", dbItem)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txn := txn.NewTransaction()
		mockConn := mock.NewMockMongoConnection("localhost", 6379, "admin", "admin", -1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		mds := mongo.NewMongoDatastore(MONGO, mockConn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)
		txn.Start()

		var item1 testutil.TestItem
		err = txn.Read(MONGO, "item1", &item1)
		assert.NoError(t, err)
		item1.Value = "item1-modified"
		txn.Write(MONGO, "item1", item1)
		err = txn.Commit()
		assert.NoError(t, err)

		// t.Logf("Connection status:\nGetTimes: %d\nPutTimes: %d\n",
		// 	mockConn.GetTimes, mockConn.PutTimes)

		assert.Equal(t, X+1, mockConn.GetTimes)
		assert.Equal(t, 2*X+2, mockConn.PutTimes)
	})

	t.Run("when X = 5", func(t *testing.T) {
		X := 5
		preTxn := NewTransactionWithSetup(MONGO)
		preTxn.Start()
		for i := 1; i <= X; i++ {
			dbItem := testutil.NewTestItem("item" + strconv.Itoa(i) + "-pre")
			preTxn.Write(MONGO, "item"+strconv.Itoa(i), dbItem)
		}
		err := preTxn.Commit()
		assert.NoError(t, err)

		txn := txn.NewTransaction()
		mockConn := mock.NewMockRedisConnection("localhost", 6379, -1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		rds := redis.NewRedisDatastore(MONGO, mockConn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
		txn.Start()

		for i := 1; i <= X; i++ {
			var item testutil.TestItem
			err = txn.Read(MONGO, "item"+strconv.Itoa(i), &item)
			assert.NoError(t, err)
			item.Value = "item" + strconv.Itoa(i) + "-modified"
			txn.Write(MONGO, "item"+strconv.Itoa(i), item)
		}
		err = txn.Commit()
		assert.NoError(t, err)

		// t.Logf("Connection status:\nGetTimes: %d\nPutTimes: %d\n",
		// 	mockConn.GetTimes, mockConn.PutTimes)

		assert.Equal(t, X+1, mockConn.GetTimes)
		assert.Equal(t, 2*X+2, mockConn.PutTimes)
	})
}
