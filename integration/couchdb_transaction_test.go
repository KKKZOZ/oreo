package integration

import (
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/mock"
	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/datastore/couchdb"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/factory"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

func TestCouchDB_TxnWrite(t *testing.T) {
	txn1 := NewTransactionWithSetup(COUCHDB)
	expected := testutil.NewDefaultPerson()

	// clear the data
	conn := NewConnectionWithSetup(COUCHDB)
	conn.Delete("John")

	// Txn1 writes the record
	err := txn1.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	err = txn1.Write(COUCHDB, "John", expected)
	if err != nil {
		t.Errorf("Error writing record: %s", err)
	}
	err = txn1.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

	txn2 := NewTransactionWithSetup(COUCHDB)

	// Txn2 reads the record
	err = txn2.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var actual testutil.Person
	err = txn2.Read(COUCHDB, "John", &actual)
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

func TestCouchDB_ReadOwnWrite(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(COUCHDB, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(COUCHDB)

	// Txn reads the record
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var person testutil.Person
	err = txn.Read(COUCHDB, "John", &person)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}

	expected := person
	expected.Age = 31
	txn.Write(COUCHDB, "John", expected)

	var actual testutil.Person
	err = txn.Read(COUCHDB, "John", &actual)
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

func TestCouchDB_SingleKeyWriteConflict(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(COUCHDB, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(COUCHDB)
	txn.Start()
	var person testutil.Person
	txn.Read(COUCHDB, "John", &person)
	person.Age = 31
	txn.Write(COUCHDB, "John", person)
	var anotherPerson testutil.Person
	txn.Read(COUCHDB, "John", &anotherPerson)

	if person != anotherPerson {
		t.Errorf("Expected two read to be the same")
	}
	txn.Commit()

	postTxn := NewTransactionWithSetup(COUCHDB)
	postTxn.Start()
	var postPerson testutil.Person
	postTxn.Read(COUCHDB, "John", &postPerson)
	if postPerson != person {
		t.Errorf("got %v want %v", postPerson, person)
	}
}

func TestCouchDB_MultileKeyWriteConflict(t *testing.T) {
	// clear the data
	conn := NewConnectionWithSetup(COUCHDB)
	conn.Delete("item1")
	conn.Delete("item2")

	preTxn := NewTransactionWithSetup(COUCHDB)
	item1 := testutil.NewTestItem("item1")
	item2 := testutil.NewTestItem("item2")
	preTxn.Start()
	preTxn.Write(COUCHDB, "item1", item1)
	preTxn.Write(COUCHDB, "item2", item2)
	err := preTxn.Commit()
	assert.Nil(t, err)

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(COUCHDB)
		txn1.Start()
		var item testutil.TestItem
		txn1.Read(COUCHDB, "item1", &item)
		item.Value = "item1-updated-by-txn1"
		txn1.Write(COUCHDB, "item1", item)

		txn1.Read(COUCHDB, "item2", &item)
		item.Value = "item2-updated-by-txn1"
		txn1.Write(COUCHDB, "item2", item)

		err := txn1.Commit()
		if err != nil {
			t.Logf("txn1 commit err: %s", err)
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup(COUCHDB)
		txn2.Start()
		var item testutil.TestItem
		txn2.Read(COUCHDB, "item2", &item)
		item.Value = "item2-updated-by-txn2"
		txn2.Write(COUCHDB, "item2", item)

		txn2.Read(COUCHDB, "item1", &item)
		item.Value = "item1-updated-by-txn2"
		txn2.Write(COUCHDB, "item1", item)

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
		postTxn := NewTransactionWithSetup(COUCHDB)
		postTxn.Start()
		var item testutil.TestItem
		postTxn.Read(COUCHDB, "item1", &item)
		t.Logf("item1: %v", item)
		postTxn.Read(COUCHDB, "item2", &item)
		t.Logf("item2: %v", item)
		postTxn.Commit()
	}
}

func TestCouchDB_RepeatableReadWhenRecordDeleted(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(COUCHDB, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(COUCHDB)
	manualTxn := NewTransactionWithSetup(COUCHDB)
	txn.Start()
	manualTxn.Start()

	var person1 testutil.Person
	txn.Read(COUCHDB, "John", &person1)

	// manualTxn deletes John and commits
	manualTxn.Delete("memory", "John")
	manualTxn.Commit()

	var person2 testutil.Person
	txn.Read(COUCHDB, "John", &person2)

	// two read in txn should be the same
	if person1 != person2 {
		t.Errorf("Expected two read in txn to be the same")
	}
}

func TestCouchDB_RepeatableReadWhenRecordUpdatedTwice(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(COUCHDB, "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup(COUCHDB)
	manualTxn1 := NewTransactionWithSetup(COUCHDB)
	txn.Start()
	manualTxn1.Start()

	var person1 testutil.Person
	txn.Read(COUCHDB, "John", &person1)

	// manualTxn1 updates John and commits
	var manualPerson1 testutil.Person
	manualTxn1.Read(COUCHDB, "John", &manualPerson1)
	manualPerson1.Age = 31
	manualTxn1.Write(COUCHDB, "John", manualPerson1)
	manualTxn1.Commit()

	manualTxn2 := NewTransactionWithSetup(COUCHDB)
	manualTxn2.Start()
	// manualTxn updates John again and commits
	var manualPerson2 testutil.Person
	manualTxn2.Read(COUCHDB, "John", &manualPerson2)
	manualPerson2.Age = 32
	manualTxn2.Write(COUCHDB, "John", manualPerson2)
	manualTxn2.Commit()

	var person2 testutil.Person
	err := txn.Read(COUCHDB, "John", &person2)
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
func TestCouchDB_RepeatableReadWhenAnotherUncommitted(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(COUCHDB, "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(COUCHDB)
		txn1.Start()
		var person testutil.Person
		txn1.Read(COUCHDB, "John", &person)
		time.Sleep(50 * time.Millisecond)

		// txn1 writes John
		person.Age = 31
		txn1.Write(COUCHDB, "John", person)

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
		txn2 := NewTransactionWithSetup(COUCHDB)
		txn2.Start()
		var person1 testutil.Person
		// txn2 reads John
		txn2.Read(COUCHDB, "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 testutil.Person
		// txn2 reads John again
		txn2.Read(COUCHDB, "John", &person2)

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
func TestCouchDB_RepeatableReadWhenAnotherCommitted(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write(COUCHDB, "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup(COUCHDB)
		txn1.Start()
		var person testutil.Person
		txn1.Read(COUCHDB, "John", &person)
		time.Sleep(50 * time.Millisecond)

		// txn1 writes John
		person.Age = 31
		txn1.Write(COUCHDB, "John", person)

		err := txn1.Commit()
		if err != nil {
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup(COUCHDB)
		txn2.Start()
		var person1 testutil.Person
		// txn2 reads John
		txn2.Read(COUCHDB, "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 testutil.Person
		// txn2 reads John again
		txn2.Read(COUCHDB, "John", &person2)

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

func TestCouchDB_TxnAbort(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	preTxn.Start()
	expected := testutil.NewDefaultPerson()
	preTxn.Write(COUCHDB, "John", expected)
	preTxn.Commit()

	txn := NewTransactionWithSetup(COUCHDB)
	var person testutil.Person
	txn.Start()
	txn.Read(COUCHDB, "John", &person)
	person.Age = 31
	txn.Write(COUCHDB, "John", person)
	txn.Abort()

	postTxn := NewTransactionWithSetup(COUCHDB)
	postTxn.Start()
	var postPerson testutil.Person
	postTxn.Read(COUCHDB, "John", &postPerson)
	postTxn.Commit()
	if postPerson != expected {
		t.Errorf("got %v want %v", postPerson, expected)
	}
}

// TODO: WTF why this test failed when using CLI
func TestCouchDB_TxnAbortCausedByWriteConflict(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(COUCHDB, item.Value, item)
	}
	err := preTxn.Commit()
	assert.Nil(t, err)

	txn := NewTransactionWithSetup(COUCHDB)
	manualTxn := NewTransactionWithSetup(COUCHDB)
	txn.Start()
	manualTxn.Start()

	// txn reads all items and modify them
	for _, item := range testutil.InputItemList {
		var actual testutil.TestItem
		txn.Read(COUCHDB, item.Value, &actual)
		actual.Value = item.Value + "updated"
		txn.Write(COUCHDB, item.Value, actual)
	}

	// manualTxn reads one item and modify it
	var manualItem testutil.TestItem
	manualTxn.Read(COUCHDB, "item4", &manualItem)
	manualItem.Value = "item4updated"
	manualTxn.Write(COUCHDB, "item4", manualItem)
	err = manualTxn.Commit()
	assert.Nil(t, err)

	err = txn.Commit()
	if err == nil {
		t.Errorf("Expected error committing transaction")
	}

	postTxn := NewTransactionWithSetup(COUCHDB)
	postTxn.Start()
	for _, item := range testutil.InputItemList {
		var actual testutil.TestItem
		postTxn.Read(COUCHDB, item.Value, &actual)
		if item.Value != "item4" {
			assert.Equal(t, item, actual)
		} else {
			assert.Equal(t, manualItem, actual)
		}
	}
	postTxn.Commit()
}

func TestCouchDB_ConcurrentTransaction(t *testing.T) {
	// Create a new redis datastore instance
	redisDst1 := redis.NewRedisDatastore("redis1", NewConnectionWithSetup(COUCHDB))

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
	preTxn.Write(COUCHDB, "John", person)
	err = preTxn.Commit()
	assert.NoError(t, err)

	resChan := make(chan bool)
	conNum := 2001

	conn := NewConnectionWithSetup(COUCHDB)

	for i := 1; i <= conNum; i++ {
		go func(id int) {
			txn := txn.NewTransaction()
			rds := couchdb.NewCouchDBDatastore(COUCHDB, conn)
			txn.AddDatastore(rds)
			txn.SetGlobalDatastore(rds)
			txn.Start()
			var person testutil.Person
			txn.Read(COUCHDB, "John", &person)
			person.Age = person.Age + id
			txn.Write(COUCHDB, "John", person)
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
func TestCouchDB_SimpleExpiredRead(t *testing.T) {
	tarMemItem := &couchdb.CouchDBItem{
		CKey:          "item1",
		CValue:        util.ToJSONString(testutil.NewTestItem("item1")),
		CGroupKeyList: "TestCouchDB_SimpleExpiredRead1",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-10 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-9 * time.Second),
		CVersion:      "1",
	}

	curMemItem := &couchdb.CouchDBItem{
		CKey:          "item1",
		CValue:        util.ToJSONString(testutil.NewTestItem("item1-prepared")),
		CGroupKeyList: "TestCouchDB_SimpleExpiredRead2",
		CTxnState:     config.PREPARED,
		CTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-4 * time.Second),
		CPrev:         util.ToJSONString(tarMemItem),
		// CVersion:  "2",
	}

	conn := NewConnectionWithSetup(COUCHDB)
	conn.Delete("item1")
	conn.PutItem("item1", curMemItem)

	txn := NewTransactionWithSetup(COUCHDB)
	txn.Start()

	var item testutil.TestItem
	txn.Read(COUCHDB, "item1", &item)
	assert.Equal(t, testutil.NewTestItem("item1"), item)
	err := txn.Commit()
	assert.NoError(t, err)
	actual, err := conn.GetItem("item1")
	assert.NoError(t, err)

	tarMemItem.SetVersion(actual.Version())
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
func TestCouchDB_SlowTransactionRecordExpiredWhenPrepare_Conflict(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(COUCHDB, item.Value, item)
	}
	preTxn.Commit()

	go func() {
		slowTxn := NewTransactionWithMockConn(COUCHDB, 2, false,
			0, func() error { time.Sleep(2 * time.Second); return nil })

		slowTxn.Start()
		for _, item := range testutil.InputItemList {
			var result testutil.TestItem
			slowTxn.Read(COUCHDB, item.Value, &result)
			result.Value = item.Value + "-slow"
			slowTxn.Write(COUCHDB, item.Value, result)
		}
		err := slowTxn.Commit()
		assert.EqualError(t, err, "prepare phase failed: version mismatch")
	}()
	time.Sleep(1 * time.Second)

	// ensure the internal state of redis database
	testConn := NewConnectionWithSetup(COUCHDB)
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

	fastTxn := NewTransactionWithSetup(COUCHDB)
	fastTxn.Start()
	for i := 2; i <= 4; i++ {
		var result testutil.TestItem
		fastTxn.Read(COUCHDB, testutil.InputItemList[i].Value, &result)
		result.Value = testutil.InputItemList[i].Value + "-fast"
		fastTxn.Write(COUCHDB, testutil.InputItemList[i].Value, result)
	}
	err := fastTxn.Commit()
	assert.NoError(t, err)

	// wait for slowTxn to complete
	time.Sleep(1 * time.Second)
	postTxn := NewTransactionWithSetup(COUCHDB)
	postTxn.Start()

	var res1 testutil.TestItem
	postTxn.Read(COUCHDB, testutil.InputItemList[0].Value, &res1)
	assert.Equal(t, testutil.InputItemList[0], res1)

	var res2 testutil.TestItem
	postTxn.Read(COUCHDB, testutil.InputItemList[1].Value, &res2)
	assert.Equal(t, testutil.InputItemList[1], res2)

	for i := 2; i <= 4; i++ {
		var res testutil.TestItem
		postTxn.Read(COUCHDB, testutil.InputItemList[i].Value, &res)
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
func TestCouchDB_SlowTransactionRecordExpiredWhenPrepare_NoConflict(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(COUCHDB, item.Value, item)
	}
	err := preTxn.Commit()
	assert.NoError(t, err)

	go func() {
		slowTxn := NewTransactionWithMockConn(COUCHDB, 4, false,
			0, func() error { time.Sleep(2 * time.Second); return nil })
		slowTxn.Start()
		for _, item := range testutil.InputItemList {
			var result testutil.TestItem
			slowTxn.Read(COUCHDB, item.Value, &result)
			result.Value = item.Value + "-slow"
			slowTxn.Write(COUCHDB, item.Value, result)
		}
		err := slowTxn.Commit()
		assert.EqualError(t, err, "transaction is aborted by other transaction")
	}()

	time.Sleep(1 * time.Second)
	testConn := NewConnectionWithSetup(COUCHDB)
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

	fastTxn := NewTransactionWithSetup(COUCHDB)
	err = fastTxn.Start()
	assert.NoError(t, err)
	for i := 2; i <= 3; i++ {
		var result testutil.TestItem
		fastTxn.Read(COUCHDB, testutil.InputItemList[i].Value, &result)
		result.Value = testutil.InputItemList[i].Value + "-fast"
		fastTxn.Write(COUCHDB, testutil.InputItemList[i].Value, result)
	}
	err = fastTxn.Commit()
	assert.NoError(t, err)

	// wait for slowTxn to complete
	time.Sleep(1 * time.Second)
	postTxn := NewTransactionWithSetup(COUCHDB)
	postTxn.Start()

	var res1 testutil.TestItem
	postTxn.Read(COUCHDB, testutil.InputItemList[0].Value, &res1)
	assert.Equal(t, testutil.InputItemList[0], res1)

	var res2 testutil.TestItem
	postTxn.Read(COUCHDB, testutil.InputItemList[1].Value, &res2)
	assert.Equal(t, testutil.InputItemList[1], res2)

	for i := 2; i <= 3; i++ {
		var res testutil.TestItem
		postTxn.Read(COUCHDB, testutil.InputItemList[i].Value, &res)
		assert.Equal(t, testutil.InputItemList[i].Value+"-fast", res.Value)
	}

	var res5 testutil.TestItem
	postTxn.Read(COUCHDB, testutil.InputItemList[4].Value, &res5)
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
func TestCouchDB_TransactionAbortWhenWritingTSR(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write(COUCHDB, item.Value, item)
	}
	err := preTxn.Commit()
	if err != nil {
		t.Errorf("preTxn commit err: %s", err)
	}

	txn := NewTransactionWithMockConn(COUCHDB, 5, true,
		0, func() error { time.Sleep(3 * time.Second); return errors.New("fail to write TSR") })
	txn.Start()
	for _, item := range testutil.InputItemList {
		var result testutil.TestItem
		txn.Read(COUCHDB, item.Value, &result)
		result.Value = item.Value + "-slow"
		txn.Write(COUCHDB, item.Value, result)
	}
	err = txn.Commit()
	assert.EqualError(t, err, "fail to write TSR")

	testTxn := NewTransactionWithSetup(COUCHDB)
	testTxn.Start()

	for i := 0; i <= 3; i++ {
		item := testutil.InputItemList[i]
		var memItem testutil.TestItem
		testTxn.Read(COUCHDB, item.Value, &memItem)
	}
	err = testTxn.Commit()
	assert.NoError(t, err)
	postTxn := NewTransactionWithSetup(COUCHDB)
	postTxn.Start()
	for i := 0; i <= 3; i++ {
		item := testutil.InputItemList[i]
		var memItem testutil.TestItem
		postTxn.Read(COUCHDB, item.Value, &memItem)
		assert.Equal(t, item.Value, memItem.Value)
	}

	conn := NewConnectionWithSetup(COUCHDB)
	memItem, err := conn.GetItem("item5")
	assert.NoError(t, err)
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item5")), memItem.Value())
	assert.Equal(t, config.COMMITTED, memItem.TxnState())
}

func TestCouchDB_LinkedRecord(t *testing.T) {
	t.Cleanup(func() {
		config.Config.MaxRecordLength = 2
	})

	t.Run("commit time less than MaxLen", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(COUCHDB)
		preTxn.Start()
		person := testutil.NewDefaultPerson()
		preTxn.Write(COUCHDB, "John", person)
		err := preTxn.Commit()
		assert.NoError(t, err)

		slowTxn := NewTransactionWithSetup(COUCHDB)
		slowTxn.Start()

		config.Config.MaxRecordLength = 4
		// 1+2=3 < 4, including origin
		commitTime := 2
		for i := 1; i <= commitTime; i++ {
			txn := NewTransactionWithSetup(COUCHDB)
			txn.Start()
			var p testutil.Person
			txn.Read(COUCHDB, "John", &p)
			p.Age = p.Age + 1
			txn.Write(COUCHDB, "John", p)
			err = txn.Commit()
			assert.NoError(t, err)
		}

		var p testutil.Person
		err = slowTxn.Read(COUCHDB, "John", &p)
		assert.NoError(t, err)
		assert.Equal(t, 30, p.Age)
	})

	t.Run("commit time equals MaxLen", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(COUCHDB)
		preTxn.Start()
		person := testutil.NewDefaultPerson()
		preTxn.Write(COUCHDB, "John", person)
		err := preTxn.Commit()
		assert.NoError(t, err)

		slowTxn := NewTransactionWithSetup(COUCHDB)
		slowTxn.Start()

		config.Config.MaxRecordLength = 4
		// 1+3=4 == 4, including origin
		commitTime := 3
		for i := 1; i <= commitTime; i++ {
			txn := NewTransactionWithSetup(COUCHDB)
			txn.Start()
			var p testutil.Person
			txn.Read(COUCHDB, "John", &p)
			p.Age = p.Age + 1
			txn.Write(COUCHDB, "John", p)
			err = txn.Commit()
			assert.NoError(t, err)
		}

		var p testutil.Person
		err = slowTxn.Read(COUCHDB, "John", &p)
		assert.NoError(t, err)
		assert.Equal(t, 30, p.Age)
	})

	t.Run("commit times bigger than MaxLen", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(COUCHDB)
		preTxn.Start()
		person := testutil.NewDefaultPerson()
		preTxn.Write(COUCHDB, "John", person)
		err := preTxn.Commit()
		assert.NoError(t, err)

		slowTxn := NewTransactionWithSetup(COUCHDB)
		slowTxn.Start()

		config.Config.MaxRecordLength = 4
		// 1+4=5 > 4, including origin
		commitTime := 4
		for i := 1; i <= commitTime; i++ {
			txn := NewTransactionWithSetup(COUCHDB)
			txn.Start()
			var p testutil.Person
			txn.Read(COUCHDB, "John", &p)
			p.Age = p.Age + 1
			txn.Write(COUCHDB, "John", p)
			err = txn.Commit()
			assert.NoError(t, err)
		}

		var p testutil.Person
		err = slowTxn.Read(COUCHDB, "John", &p)
		assert.EqualError(t, err, "key not found")
	})
}

func TestCouchDB_RollbackConflict(t *testing.T) {
	// there is a broken item
	//   - txnA reads the item, decides to roll back
	//   - txnB reads the item, decides to roll back
	//   - txnB rollbacks the item
	//   - txnB update the item and commit
	//   - txnA rollbacks the item -> should fail
	t.Run("the broken item has a valid Prev field", func(t *testing.T) {
		conn := NewConnectionWithSetup(COUCHDB)

		redisItem1 := &couchdb.CouchDBItem{
			CKey:          "item1",
			CValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
			CGroupKeyList: "TestCouchDB_RollbackConflict1",
			CTxnState:     config.COMMITTED,
			CTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
			CTLease:       time.Now().Add(-4 * time.Second),
			CLinkedLen:    1,
			CVersion:      "1",
		}
		redisItem2 := &couchdb.CouchDBItem{
			CKey:          "item1",
			CValue:        util.ToJSONString(testutil.NewTestItem("item1-broken")),
			CGroupKeyList: "TestCouchDB_RollbackConflict2",
			CTxnState:     config.PREPARED,
			CTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
			CTLease:       time.Now().Add(-4 * time.Second),
			CPrev:         util.ToJSONString(redisItem1),
			CLinkedLen:    2,
			// CVersion:   "2",
		}
		conn.Delete("item1")
		conn.PutItem("item1", redisItem2)

		go func() {
			time.Sleep(100 * time.Millisecond)
			txnA := NewTransactionWithMockConn(COUCHDB, 0, false,
				0, func() error { time.Sleep(2 * time.Second); return nil })
			txnA.Start()

			var item testutil.TestItem
			err := txnA.Read(COUCHDB, "item1", &item)
			assert.NotNil(t, err)
		}()

		time.Sleep(100 * time.Millisecond)
		txnB := NewTransactionWithSetup(COUCHDB)
		txnB.Start()

		var item testutil.TestItem
		err := txnB.Read(COUCHDB, "item1", &item)
		assert.NoError(t, err)
		assert.Equal(t, testutil.NewTestItem("item1-pre"), item)
		item.Value = "item1-B"
		txnB.Write(COUCHDB, "item1", item)
		err = txnB.Commit()
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)
		resItem, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())

		var previousItem couchdb.CouchDBItem
		err = json.Unmarshal([]byte(resItem.Prev()), &previousItem)
		assert.NoError(t, err)

		redisItem1.SetVersion(previousItem.Version())
		assert.Equal(t, util.ToJSONString(redisItem1), resItem.Prev())
	})

	// there is a broken item
	//   - txnA reads the item, decides to roll back
	//   - txnB reads the item, decides to roll back
	//   - txnB rollbacks the item
	//   - txnB update the item and commit
	//   - txnA rollbacks the item -> should fail
	t.Run("the broken item has an empty Prev field", func(t *testing.T) {
		conn := NewConnectionWithSetup(COUCHDB)

		redisItem2 := &couchdb.CouchDBItem{
			CKey:          "item1",
			CValue:        util.ToJSONString(testutil.NewTestItem("item1-broken")),
			CGroupKeyList: "TestCouchDB_RollbackConflict2-emptyField",
			CTxnState:     config.PREPARED,
			CTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
			CTLease:       time.Now().Add(-4 * time.Second),
			CPrev:         "",
			CLinkedLen:    1,
			// CVersion:   "1",
		}
		conn.Delete("item1")
		conn.PutItem("item1", redisItem2)

		go func() {
			time.Sleep(100 * time.Millisecond)
			txnA := NewTransactionWithMockConn(COUCHDB, 0, false,
				0, func() error { time.Sleep(2 * time.Second); return nil })
			txnA.Start()

			var item testutil.TestItem
			err := txnA.Read(COUCHDB, "item1", &item)
			assert.NotNil(t, err)
		}()

		time.Sleep(100 * time.Millisecond)
		txnB := NewTransactionWithSetup(COUCHDB)
		txnB.Start()
		var item testutil.TestItem
		err := txnB.Read(COUCHDB, "item1", &item)
		assert.EqualError(t, err, txn.KeyNotFound.Error())

		item.Value = "item1-B"
		txnB.Write(COUCHDB, "item1", item)
		err = txnB.Commit()
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)
		resItem, err := conn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())

		var previousItem couchdb.CouchDBItem
		err = json.Unmarshal([]byte(resItem.Prev()), &previousItem)
		assert.NoError(t, err)

		redisItem2.CIsDeleted = true
		redisItem2.CTxnState = config.COMMITTED
		redisItem2.SetVersion(previousItem.Version())
		assert.Equal(t, util.ToJSONString(redisItem2), resItem.Prev())
	})
}

// there is a broken item
//   - txnA reads the item, decides to roll forward
//   - txnB reads the item, decides to roll forward
//   - txnB rolls forward the item
//   - txnB update the item and commit
//   - txnA rolls forward the item -> should fail
func TestCouchDB_RollForwardConflict(t *testing.T) {
	conn := NewConnectionWithSetup(COUCHDB)

	redisItem1 := &couchdb.CouchDBItem{
		CKey:          "item1",
		CValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
		CGroupKeyList: "TestCouchDB_RollForwardConflict1",
		CTxnState:     config.COMMITTED,
		CTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-4 * time.Second),
		CVersion:      "1",
	}
	redisItem2 := &couchdb.CouchDBItem{
		CKey:          "item1",
		CValue:        util.ToJSONString(testutil.NewTestItem("item1-broken")),
		CGroupKeyList: "TestCouchDB_RollForwardConflict2",
		CTxnState:     config.PREPARED,
		CTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
		CTLease:       time.Now().Add(-4 * time.Second),
		CPrev:         util.ToJSONString(redisItem1),
		CLinkedLen:    2,
		// CVersion:   "2",
	}
	conn.Delete("item1")
	conn.Delete("TestCouchDB_RollForwardConflict2")
	conn.PutItem("item1", redisItem2)
	conn.Put("TestCouchDB_RollForwardConflict2", config.COMMITTED)

	go func() {
		time.Sleep(100 * time.Millisecond)
		txnA := NewTransactionWithMockConn(COUCHDB, 0, false,
			0, func() error { time.Sleep(2 * time.Second); return nil })
		txnA.Start()
		var item testutil.TestItem
		err := txnA.Read(COUCHDB, "item1", &item)
		assert.NotNil(t, err)
	}()

	time.Sleep(100 * time.Millisecond)
	txnB := NewTransactionWithSetup(COUCHDB)
	txnB.Start()
	var item testutil.TestItem
	txnB.Read(COUCHDB, "item1", &item)
	assert.Equal(t, testutil.NewTestItem("item1-broken"), item)
	item.Value = "item1-B"
	txnB.Write(COUCHDB, "item1", item)
	err := txnB.Commit()
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)
	resItem, err := conn.GetItem("item1")
	assert.NoError(t, err)
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())

	var previousItem couchdb.CouchDBItem
	err = json.Unmarshal([]byte(resItem.Prev()), &previousItem)
	assert.NoError(t, err)

	redisItem2.CTxnState = config.COMMITTED
	redisItem2.CPrev = ""
	redisItem2.CLinkedLen = 1
	redisItem2.SetVersion(previousItem.Version())
	assert.Equal(t, util.ToJSONString(redisItem2), resItem.Prev())
}

func TestCouchDB_ConcurrentDirectWrite(t *testing.T) {
	conn := NewConnectionWithSetup(COUCHDB)
	conn.Delete("item1")

	conNumber := 5
	mu := sync.Mutex{}
	globalId := 0
	resChan := make(chan bool)
	for i := 1; i <= conNumber; i++ {
		go func(id int) {
			txn := NewTransactionWithSetup(COUCHDB)
			item := testutil.NewTestItem("item1-" + strconv.Itoa(id))
			txn.Start()
			txn.Write(COUCHDB, "item1", item)
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

	postTxn := NewTransactionWithSetup(COUCHDB)
	postTxn.Start()
	var item testutil.TestItem
	postTxn.Read(COUCHDB, "item1", &item)
	assert.Equal(t, testutil.NewTestItem("item1-"+strconv.Itoa(globalId)), item)
	// assert.Equal(t, 0, globalId)
	t.Logf("item: %v", item)
}

func TestCouchDB_TxnDelete(t *testing.T) {
	preTxn := NewTransactionWithSetup(COUCHDB)
	preTxn.Start()
	item := testutil.NewTestItem("item1-pre")
	preTxn.Write(COUCHDB, "item1", item)
	err := preTxn.Commit()
	assert.NoError(t, err)

	txn := NewTransactionWithSetup(COUCHDB)
	txn.Start()
	var item1 testutil.TestItem
	txn.Read(COUCHDB, "item1", &item1)
	txn.Delete(COUCHDB, "item1")
	err = txn.Commit()
	assert.NoError(t, err)

	postTxn := NewTransactionWithSetup(COUCHDB)
	postTxn.Start()
	var item2 testutil.TestItem
	err = postTxn.Read(COUCHDB, "item1", &item2)
	assert.EqualError(t, err, "key not found")
}

func TestCouchDB_PreventLostUpdatesValidation(t *testing.T) {
	t.Run("Case 1-1(with read): The target record has been updated by the concurrent transaction",
		func(t *testing.T) {
			preTxn := NewTransactionWithSetup(COUCHDB)
			preTxn.Start()
			item := testutil.NewTestItem("item1-pre")
			preTxn.Write(COUCHDB, "item1", item)
			err := preTxn.Commit()
			assert.NoError(t, err)

			txnA := NewTransactionWithSetup(COUCHDB)
			txnA.Start()
			txnB := NewTransactionWithSetup(COUCHDB)
			txnB.Start()

			var itemA testutil.TestItem
			txnA.Read(COUCHDB, "item1", &itemA)
			itemA.Value = "item1-A"
			txnA.Write(COUCHDB, "item1", itemA)
			err = txnA.Commit()
			assert.NoError(t, err)

			var itemB testutil.TestItem
			txnB.Read(COUCHDB, "item1", &itemB)
			itemB.Value = "item1-B"
			txnB.Write(COUCHDB, "item1", itemB)
			err = txnB.Commit()
			assert.EqualError(t, err,
				"prepare phase failed: version mismatch")

			postTxn := NewTransactionWithSetup(COUCHDB)
			postTxn.Start()
			var item1 testutil.TestItem
			postTxn.Read(COUCHDB, "item1", &item1)
			assert.Equal(t, itemA, item1)
		})

	t.Run(
		"Case 1-2(without read): The target record has been updated by the concurrent transaction",
		func(t *testing.T) {
			preTxn := NewTransactionWithSetup(COUCHDB)
			preTxn.Start()
			item := testutil.NewTestItem("item1-pre")
			preTxn.Write(COUCHDB, "item1", item)
			err := preTxn.Commit()
			assert.NoError(t, err)

			txnA := NewTransactionWithSetup(COUCHDB)
			txnA.Start()
			txnB := NewTransactionWithSetup(COUCHDB)
			txnB.Start()

			itemA := testutil.NewTestItem("item1-A")
			txnA.Write(COUCHDB, "item1", itemA)
			err = txnA.Commit()
			assert.NoError(t, err)

			itemB := testutil.NewTestItem("item1-B")
			txnB.Write(COUCHDB, "item1", itemB)
			err = txnB.Commit()
			assert.EqualError(t, err,
				"prepare phase failed: version mismatch")

			postTxn := NewTransactionWithSetup(COUCHDB)
			postTxn.Start()
			var item1 testutil.TestItem
			postTxn.Read(COUCHDB, "item1", &item1)
			assert.Equal(t, itemA, item1)
		},
	)

	t.Run("Case 2-1(with read): There is no conflict", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(COUCHDB)
		preTxn.Start()
		item := testutil.NewTestItem("item1-pre")
		preTxn.Write(COUCHDB, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txnA := NewTransactionWithSetup(COUCHDB)
		txnA.Start()
		var itemA testutil.TestItem
		txnA.Read(COUCHDB, "item1", &itemA)
		itemA.Value = "item1-A"
		txnA.Write(COUCHDB, "item1", itemA)
		err = txnA.Commit()
		assert.NoError(t, err)

		txnB := NewTransactionWithSetup(COUCHDB)
		txnB.Start()
		var itemB testutil.TestItem
		txnB.Read(COUCHDB, "item1", &itemB)
		itemB.Value = "item1-B"
		txnB.Write(COUCHDB, "item1", itemB)
		err = txnB.Commit()
		assert.NoError(t, err)

		postTxn := NewTransactionWithSetup(COUCHDB)
		postTxn.Start()
		var item1 testutil.TestItem
		postTxn.Read(COUCHDB, "item1", &item1)
		assert.Equal(t, itemB, item1)
	})

	t.Run("Case 2-2(without read): There is no conflict", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(COUCHDB)
		preTxn.Start()
		item := testutil.NewTestItem("item1-pre")
		preTxn.Write(COUCHDB, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txnA := NewTransactionWithSetup(COUCHDB)
		txnA.Start()
		itemA := testutil.NewTestItem("item1-A")
		txnA.Write(COUCHDB, "item1", itemA)
		err = txnA.Commit()
		assert.NoError(t, err)

		txnB := NewTransactionWithSetup(COUCHDB)
		txnB.Start()
		itemB := testutil.NewTestItem("item1-B")
		txnB.Write(COUCHDB, "item1", itemB)
		err = txnB.Commit()
		assert.NoError(t, err)

		postTxn := NewTransactionWithSetup(COUCHDB)
		postTxn.Start()
		var item1 testutil.TestItem
		postTxn.Read(COUCHDB, "item1", &item1)
		assert.Equal(t, itemB, item1)
	})
}

func TestCouchDB_RepeatableReadWhenDirtyRead(t *testing.T) {
	t.Run("the prepared item has a valid Prev", func(t *testing.T) {
		config.Config.LeaseTime = 3000 * time.Millisecond

		preTxn := NewTransactionWithSetup(COUCHDB)
		preTxn.Start()
		item := testutil.NewTestItem("item1-pre")
		preTxn.Write(COUCHDB, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txnA := NewTransactionWithMockConn(COUCHDB, 1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		txnB := NewTransactionWithSetup(COUCHDB)
		txnA.Start()
		txnB.Start()

		go func() {
			itemA := testutil.NewTestItem("item1-A")
			txnA.Write(COUCHDB, "item1", itemA)
			err = txnA.Commit()
			assert.NoError(t, err)
		}()

		// wait for txnA to write item into database
		time.Sleep(500 * time.Millisecond)
		var itemB testutil.TestItem
		err = txnB.Read(COUCHDB, "item1", &itemB)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre", itemB.Value)

		time.Sleep(2 * time.Second)
		err = txnB.Read(COUCHDB, "item1", &itemB)
		assert.NoError(t, err)
		assert.Equal(t, "item1-pre", itemB.Value)
	})

	t.Run("the prepared item has an empty Prev", func(t *testing.T) {
		config.Config.LeaseTime = 3000 * time.Millisecond

		testConn := NewConnectionWithSetup(COUCHDB)
		testConn.Delete("item1")

		txnA := NewTransactionWithMockConn(COUCHDB, 1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		txnB := NewTransactionWithSetup(COUCHDB)
		txnA.Start()
		txnB.Start()

		go func() {
			itemA := testutil.NewTestItem("item1-A")
			txnA.Write(COUCHDB, "item1", itemA)
			err := txnA.Commit()
			assert.NoError(t, err)
		}()

		// wait for txnA to write item into database
		time.Sleep(500 * time.Millisecond)
		var itemB testutil.TestItem
		err := txnB.Read(COUCHDB, "item1", &itemB)
		assert.EqualError(t, err, txn.KeyNotFound.Error())

		time.Sleep(2 * time.Second)

		// make sure txnA has committed
		resItem, err := testConn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-A")), resItem.Value())

		err = txnB.Read(COUCHDB, "item1", &itemB)
		assert.EqualError(t, err, txn.KeyNotFound.Error())

		// post check
		postTxn := NewTransactionWithSetup(COUCHDB)
		postTxn.Start()

		var itemPost testutil.TestItem
		err = postTxn.Read(COUCHDB, "item1", &itemPost)
		assert.NoError(t, err)
		assert.Equal(t, "item1-A", itemPost.Value)
	})
}

func TestCouchDB_DeleteTimingProblems(t *testing.T) {
	// conn Puts an item with an empty Prev field
	//  - txnA reads the item, decides to delete
	//  - txnB reads the item, updates and commmits
	//  - txnA tries to delete the item -> should fail
	t.Run("the item has an empty Prev field", func(t *testing.T) {
		testConn := NewConnectionWithSetup(COUCHDB)
		dbItem := &couchdb.CouchDBItem{
			CKey:          "item1",
			CValue:        util.ToJSONString(testutil.NewTestItem("item1-pre")),
			CGroupKeyList: "TestCouchDB_DeleteTimingProblems",
			CTxnState:     config.COMMITTED,
			CTValid:       time.Now().Add(-5 * time.Second).UnixMicro(),
			CTLease:       time.Now().Add(-4 * time.Second),
			CVersion:      "1",
		}
		testConn.PutItem("item1", dbItem)

		txnA := NewTransactionWithMockConn(COUCHDB, 0, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		txnB := NewTransactionWithSetup(COUCHDB)
		txnA.Start()
		txnB.Start()

		go func() {
			var item testutil.TestItem
			txnA.Read(COUCHDB, "item1", &item)
			txnA.Delete(COUCHDB, "item1")
			err := txnA.Commit()
			assert.NotNil(t, err)
			// assert.NoError(t, err)
		}()

		time.Sleep(200 * time.Millisecond)
		var itemB testutil.TestItem
		txnB.Read(COUCHDB, "item1", &itemB)
		txnB.Delete(COUCHDB, "item1")
		itemB.Value = "item1-B"
		txnB.Write(COUCHDB, "item1", itemB)
		err := txnB.Commit()
		assert.NoError(t, err)

		time.Sleep(1 * time.Second)

		// post check
		resItem, err := testConn.GetItem("item1")
		assert.NoError(t, err)
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-B")), resItem.Value())
	})
}

func TestCouchDB_VisibilityResults(t *testing.T) {
	t.Run("a normal chain", func(t *testing.T) {
		preTxn := NewTransactionWithSetup(COUCHDB)
		preTxn.Start()
		item := testutil.NewTestItem("item1-V0")
		preTxn.Write(COUCHDB, "item1", item)
		err := preTxn.Commit()
		assert.NoError(t, err)

		chainNum := 5

		for i := 1; i <= chainNum; i++ {
			time.Sleep(10 * time.Millisecond)
			txn := NewTransactionWithSetup(COUCHDB)
			txn.Start()
			var item testutil.TestItem
			txn.Read(COUCHDB, "item1", &item)
			assert.Equal(t, "item1-V"+strconv.Itoa(i-1), item.Value)
			item.Value = "item1-V" + strconv.Itoa(i)
			txn.Write(COUCHDB, "item1", item)
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

func TestCouchDB_ReadModifyWritePattern(t *testing.T) {
	t.Run("when X = 1", func(t *testing.T) {
		X := 1
		preTxn := NewTransactionWithSetup(COUCHDB)
		preTxn.Start()
		dbItem := testutil.NewTestItem("item1-pre")
		preTxn.Write(COUCHDB, "item1", dbItem)
		err := preTxn.Commit()
		assert.NoError(t, err)

		txn := txn.NewTransaction()
		mockConn := mock.NewMockRedisConnection("localhost", 6379, -1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		rds := redis.NewRedisDatastore(COUCHDB, mockConn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
		txn.Start()

		var item1 testutil.TestItem
		err = txn.Read(COUCHDB, "item1", &item1)
		assert.NoError(t, err)
		item1.Value = "item1-modified"
		txn.Write(COUCHDB, "item1", item1)
		err = txn.Commit()
		assert.NoError(t, err)

		// t.Logf("Connection status:\nGetTimes: %d\nPutTimes: %d\n",
		// 	mockConn.GetTimes, mockConn.PutTimes)

		assert.Equal(t, X+1, mockConn.GetTimes)
		assert.Equal(t, 2*X+2, mockConn.PutTimes)
	})

	t.Run("when X = 5", func(t *testing.T) {
		X := 5
		preTxn := NewTransactionWithSetup(COUCHDB)
		preTxn.Start()
		for i := 1; i <= X; i++ {
			dbItem := testutil.NewTestItem("item" + strconv.Itoa(i) + "-pre")
			preTxn.Write(COUCHDB, "item"+strconv.Itoa(i), dbItem)
		}
		err := preTxn.Commit()
		assert.NoError(t, err)

		txn := txn.NewTransaction()
		mockConn := mock.NewMockRedisConnection("localhost", 6379, -1, false,
			0, func() error { time.Sleep(1 * time.Second); return nil })
		rds := redis.NewRedisDatastore(COUCHDB, mockConn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
		txn.Start()

		for i := 1; i <= X; i++ {
			var item testutil.TestItem
			err = txn.Read(COUCHDB, "item"+strconv.Itoa(i), &item)
			assert.NoError(t, err)
			item.Value = "item" + strconv.Itoa(i) + "-modified"
			txn.Write(COUCHDB, "item"+strconv.Itoa(i), item)
		}
		err = txn.Commit()
		assert.NoError(t, err)

		// t.Logf("Connection status:\nGetTimes: %d\nPutTimes: %d\n",
		// 	mockConn.GetTimes, mockConn.PutTimes)

		assert.Equal(t, X+1, mockConn.GetTimes)
		assert.Equal(t, 2*X+2, mockConn.PutTimes)
	})
}
