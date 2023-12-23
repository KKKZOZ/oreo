package integration

import (
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/datastore/memory"
	"github.com/kkkzoz/oreo/pkg/factory"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

func NewTransactionWithSetup() *txn.Transaction {
	txn := txn.NewTransaction()
	conn := memory.NewMemoryConnection("localhost", 8321)
	mds := memory.NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)
	return txn
}

// TestTxnStartAgain tests the behavior of starting a transaction multiple times.
func TestTxnStartAgain(t *testing.T) {
	// Create a new memory database instance
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new transaction with setup
	txn := NewTransactionWithSetup()

	// Start the transaction
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}

	// Try starting the transaction again and expect an error
	err = txn.Start()
	if err == nil {
		t.Errorf("Expected error starting transaction")
	}
}

// TestTxnCommitWithoutStart tests the scenario where a transaction is committed without being started.
func TestTxnCommitWithoutStart(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	txn := NewTransactionWithSetup()
	err := txn.Commit()
	if err == nil {
		t.Errorf("Expected error committing transaction")
	}
}

// TestTxnAbortWithoutStart tests the behavior of aborting a transaction without starting it.
func TestTxnAbortWithoutStart(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	txn := NewTransactionWithSetup()
	err := txn.Abort()
	if err == nil {
		t.Errorf("Expected error aborting transaction")
	}
}

// TestTxnOperateWithoutStart tests the behavior of transaction operations
// when the database has not been started.
func TestTxnOperateWithoutStart(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	txn := NewTransactionWithSetup()
	var person testutil.Person
	err := txn.Read("memory", "John", &person)
	if err == nil {
		t.Errorf("Expected error reading record")
	}
	err = txn.Write("memory", "John", person)
	if err == nil {
		t.Errorf("Expected error writing record")
	}
	err = txn.Delete("memory", "John")
	if err == nil {
		t.Errorf("Expected error deleting record")
	}
}

func TestTxnWrite(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	txn1 := NewTransactionWithSetup()

	expected := testutil.NewDefaultPerson()

	// Txn1 writes the record
	err := txn1.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	err = txn1.Write("memory", "John", expected)
	if err != nil {
		t.Errorf("Error writing record: %s", err)
	}
	err = txn1.Commit()
	if err != nil {
		t.Errorf("Error committing transaction: %s", err)
	}

	txn2 := NewTransactionWithSetup()

	// Txn2 reads the record
	err = txn2.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var actual testutil.Person
	err = txn2.Read("memory", "John", &actual)
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

func TestReadOwnWrite(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
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

	// Txn reads the record
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var person testutil.Person
	err = txn.Read("memory", "John", &person)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}

	expected := person
	expected.Age = 31
	txn.Write("memory", "John", expected)

	var actual testutil.Person
	err = txn.Read("memory", "John", &actual)
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

func TestSingleKeyWriteConflict(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
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
	var person testutil.Person
	txn.Read("memory", "John", &person)
	person.Age = 31
	txn.Write("memory", "John", person)
	var anotherPerson testutil.Person
	txn.Read("memory", "John", &anotherPerson)

	if person != anotherPerson {
		t.Errorf("Expected two read to be the same")
	}
	txn.Commit()

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	var postPerson testutil.Person
	postTxn.Read("memory", "John", &postPerson)
	if postPerson != person {
		t.Errorf("got %v want %v", postPerson, person)
	}

}

func TestMultileKeyWriteConflict(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	item1 := testutil.NewTestItem("item1")
	item2 := testutil.NewTestItem("item2")
	preTxn.Start()
	preTxn.Write("memory", "item1", item1)
	preTxn.Write("memory", "item2", item2)
	err := preTxn.Commit()
	assert.Nil(t, err)

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var item testutil.TestItem
		txn1.Read("memory", "item1", &item)
		item.Value = "item1-updated-by-txn1"
		txn1.Write("memory", "item1", item)

		txn1.Read("memory", "item2", &item)
		item.Value = "item2-updated-by-txn1"
		txn1.Write("memory", "item2", item)

		err := txn1.Commit()
		if err != nil {
			t.Logf("txn1 commit err: %s", err)
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup()
		txn2.Start()
		var item testutil.TestItem
		txn2.Read("memory", "item2", &item)
		item.Value = "item2-updated-by-txn2"
		txn2.Write("memory", "item2", item)

		txn2.Read("memory", "item1", &item)
		item.Value = "item1-updated-by-txn2"
		txn2.Write("memory", "item1", item)

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
		postTxn := NewTransactionWithSetup()
		postTxn.Start()
		var item testutil.TestItem
		postTxn.Read("memory", "item1", &item)
		t.Logf("item1: %v", item)
		postTxn.Read("memory", "item2", &item)
		t.Logf("item2: %v", item)
		postTxn.Commit()
	}
}

func TestRepeatableReadWhenRecordDeleted(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
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
	manualTxn := NewTransactionWithSetup()
	txn.Start()
	manualTxn.Start()

	var person1 testutil.Person
	txn.Read("memory", "John", &person1)

	// manualTxn deletes John and commits
	manualTxn.Delete("memory", "John")
	manualTxn.Commit()

	var person2 testutil.Person
	txn.Read("memory", "John", &person2)

	// two read in txn should be the same
	if person1 != person2 {
		t.Errorf("Expected two read in txn to be the same")
	}
}

func TestRepeatableReadWhenRecordUpdatedTwice(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
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
	manualTxn1 := NewTransactionWithSetup()
	txn.Start()
	manualTxn1.Start()

	var person1 testutil.Person
	txn.Read("memory", "John", &person1)

	// manualTxn1 updates John and commits
	var manualPerson1 testutil.Person
	manualTxn1.Read("memory", "John", &manualPerson1)
	manualPerson1.Age = 31
	manualTxn1.Write("memory", "John", manualPerson1)
	manualTxn1.Commit()

	manualTxn2 := NewTransactionWithSetup()
	manualTxn2.Start()
	// manualTxn updates John again and commits
	var manualPerson2 testutil.Person
	manualTxn2.Read("memory", "John", &manualPerson2)
	manualPerson2.Age = 32
	manualTxn2.Write("memory", "John", manualPerson2)
	manualTxn2.Commit()

	var person2 testutil.Person
	err := txn.Read("memory", "John", &person2)
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
func TestRepeatableReadWhenAnotherUncommitted(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var person testutil.Person
		txn1.Read("memory", "John", &person)
		time.Sleep(50 * time.Millisecond)

		// txn1 writes John
		person.Age = 31
		txn1.Write("memory", "John", person)

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
		txn2 := NewTransactionWithSetup()
		txn2.Start()
		var person1 testutil.Person
		// txn2 reads John
		txn2.Read("memory", "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 testutil.Person
		// txn2 reads John again
		txn2.Read("memory", "John", &person2)

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
func TestRepeatableReadWhenAnotherCommitted(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := testutil.NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var person testutil.Person
		txn1.Read("memory", "John", &person)
		time.Sleep(50 * time.Millisecond)

		// txn1 writes John
		person.Age = 31
		txn1.Write("memory", "John", person)

		err := txn1.Commit()
		if err != nil {
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	go func() {
		txn2 := NewTransactionWithSetup()
		txn2.Start()
		var person1 testutil.Person
		// txn2 reads John
		txn2.Read("memory", "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 testutil.Person
		// txn2 reads John again
		txn2.Read("memory", "John", &person2)

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

func TestTxnAbort(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	expected := testutil.NewDefaultPerson()
	preTxn.Write("memory", "John", expected)
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	var person testutil.Person
	txn.Start()
	txn.Read("memory", "John", &person)
	person.Age = 31
	txn.Write("memory", "John", person)
	txn.Abort()

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	var postPerson testutil.Person
	postTxn.Read("memory", "John", &postPerson)
	postTxn.Commit()
	if postPerson != expected {
		t.Errorf("got %v want %v", postPerson, expected)
	}
}

// TODO: WTF why this test failed when using CLI
func TestTxnAbortCausedByWriteConflict(t *testing.T) {
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	err := testutil.WaitForServer("localhost", 8321, 100*time.Millisecond)
	assert.Nil(t, err)

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	for _, item := range testutil.InputItemList {
		preTxn.Write("memory", item.Value, item)
	}
	err = preTxn.Commit()
	assert.Nil(t, err)

	txn := NewTransactionWithSetup()
	manualTxn := NewTransactionWithSetup()
	txn.Start()
	manualTxn.Start()

	// txn reads all items and modify them
	for _, item := range testutil.InputItemList {
		var actual testutil.TestItem
		txn.Read("memory", item.Value, &actual)
		actual.Value = item.Value + "updated"
		txn.Write("memory", item.Value, actual)
	}

	// manualTxn reads one item and modify it
	var manualItem testutil.TestItem
	manualTxn.Read("memory", "item4", &manualItem)
	manualItem.Value = "item4updated"
	manualTxn.Write("memory", "item4", manualItem)
	err = manualTxn.Commit()
	assert.Nil(t, err)

	err = txn.Commit()
	if err == nil {
		t.Errorf("Expected error committing transaction")
	}

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	for _, item := range testutil.InputItemList {
		var actual testutil.TestItem
		postTxn.Read("memory", item.Value, &actual)
		if item.Value != "item4" {
			assert.Equal(t, item, actual)
		} else {
			assert.Equal(t, manualItem, actual)
		}
	}
	postTxn.Commit()
}

// TODO: Dangetous test due to use of an unstable version of TransactionFactory
func TestConcurrentTransaction(t *testing.T) {
	// Create a new memory database instance
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new memory datastore instance
	memDst1 := memory.NewMemoryDatastore("mem1", memory.NewMemoryConnection("localhost", 8321))

	txnFactory, err := factory.NewTransactionFactory(&factory.TransactionConfig{
		DatastoreList:    []txn.Datastore{memDst1},
		GlobalDatastore:  memDst1,
		TimeOracleSource: txn.LOCAL,
		LockerSource:     txn.LOCAL,
	})
	assert.NoError(t, err)

	preTxn := txnFactory.NewTransaction()
	preTxn.Start()
	person := testutil.NewDefaultPerson()
	preTxn.Write("mem1", "John", person)
	err = preTxn.Commit()
	assert.NoError(t, err)

	resChan := make(chan bool)

	conNum := 10

	for i := 1; i <= conNum; i++ {
		go func(id int) {
			txn := NewTransactionWithSetup()
			txn.Start()
			var person testutil.Person
			txn.Read("memory", "John", &person)
			person.Age = person.Age + id
			txn.Write("memory", "John", person)
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

func TestSlowTransactionRecordExpiredWhenPrepare(t *testing.T) {
	// run a memory database
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
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

	go func() {
		slowTxn := txn.NewTransaction()
		conn := memory.NewMockMemoryConnection("localhost", 8321, 2, false,
			func() error { time.Sleep(3 * time.Second); return nil })
		mds := memory.NewMemoryDatastore("memory", conn)
		slowTxn.AddDatastore(mds)
		slowTxn.SetGlobalDatastore(mds)

		slowTxn.Start()
		for _, item := range testutil.InputItemList {
			var result testutil.TestItem
			slowTxn.Read("memory", item.Value, &result)
			result.Value = item.Value + "-slow"
			slowTxn.Write("memory", item.Value, result)
		}
		err := slowTxn.Commit()
		assert.EqualError(t, err, "prepare phase failed: write conflicted")
	}()
	time.Sleep(2 * time.Second)

	testConn := memory.NewMemoryConnection("localhost", 8321)
	testConn.Connect()
	var memItem1 memory.MemoryItem
	testConn.Get("item1", &memItem1)
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item1-slow")), memItem1.Value)
	assert.Equal(t, memItem1.TxnState, config.PREPARED)

	var memItem2 memory.MemoryItem
	testConn.Get("item2", &memItem2)
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item2-slow")), memItem2.Value)
	assert.Equal(t, memItem2.TxnState, config.PREPARED)

	var memItem3 memory.MemoryItem
	testConn.Get("item3", &memItem3)
	assert.Equal(t, util.ToJSONString(testutil.NewTestItem("item3")), memItem3.Value)
	assert.Equal(t, memItem3.TxnState, config.COMMITTED)

	fastTxn := NewTransactionWithSetup()
	fastTxn.Start()
	for i := 2; i <= 4; i++ {
		var result testutil.TestItem
		fastTxn.Read("memory", testutil.InputItemList[i].Value, &result)
		result.Value = testutil.InputItemList[i].Value + "-fast"
		fastTxn.Write("memory", testutil.InputItemList[i].Value, result)
	}
	err := fastTxn.Commit()
	assert.NoError(t, err)

	postTxn := NewTransactionWithSetup()
	postTxn.Start()

	var res1 testutil.TestItem
	postTxn.Read("memory", testutil.InputItemList[0].Value, &res1)
	assert.Equal(t, testutil.InputItemList[0], res1)

	var res2 testutil.TestItem
	postTxn.Read("memory", testutil.InputItemList[1].Value, &res2)
	assert.Equal(t, testutil.InputItemList[1], res2)

	for i := 2; i <= 4; i++ {
		var res testutil.TestItem
		postTxn.Read("memory", testutil.InputItemList[i].Value, &res)
		assert.Equal(t, testutil.InputItemList[i].Value+"-fast", res.Value)
	}

	err = postTxn.Commit()
	assert.NoError(t, err)

}

func TestSlowTransactionRecordExpiredWhenWriteTSR(t *testing.T) {
	// run a memory database
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
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

	go func() {
		slowTxn := txn.NewTransaction()
		conn := memory.NewMockMemoryConnection("localhost", 8321, 5, false,
			func() error { time.Sleep(3 * time.Second); return nil })
		mds := memory.NewMemoryDatastore("memory", conn)
		slowTxn.AddDatastore(mds)
		slowTxn.SetGlobalDatastore(mds)

		slowTxn.Start()
		for _, item := range testutil.InputItemList {
			var result testutil.TestItem
			slowTxn.Read("memory", item.Value, &result)
			result.Value = item.Value + "-slow"
			slowTxn.Write("memory", item.Value, result)
		}
		err := slowTxn.Commit()
		assert.EqualError(t, err, "transaction is aborted by other transaction")
	}()
	time.Sleep(2 * time.Second)

	testConn := memory.NewMemoryConnection("localhost", 8321)
	testConn.Connect()

	// all records should be PREPARED state
	for _, item := range testutil.InputItemList {
		var memItem memory.MemoryItem
		testConn.Get(item.Value, &memItem)
		itemValue := item.Value + "-slow"
		assert.Equal(t, util.ToJSONString(testutil.NewTestItem(itemValue)), memItem.Value)
		assert.Equal(t, memItem.TxnState, config.PREPARED)
	}

	fastTxn := NewTransactionWithSetup()
	fastTxn.Start()
	for i := 2; i <= 4; i++ {
		var result testutil.TestItem
		fastTxn.Read("memory", testutil.InputItemList[i].Value, &result)
		result.Value = testutil.InputItemList[i].Value + "-fast"
		fastTxn.Write("memory", testutil.InputItemList[i].Value, result)
	}
	err := fastTxn.Commit()
	assert.NoError(t, err)

	postTxn := NewTransactionWithSetup()
	postTxn.Start()

	var res1 testutil.TestItem
	postTxn.Read("memory", testutil.InputItemList[0].Value, &res1)
	assert.Equal(t, testutil.InputItemList[0], res1)

	var res2 testutil.TestItem
	postTxn.Read("memory", testutil.InputItemList[1].Value, &res2)
	assert.Equal(t, testutil.InputItemList[1], res2)

	for i := 2; i <= 4; i++ {
		var res testutil.TestItem
		postTxn.Read("memory", testutil.InputItemList[i].Value, &res)
		assert.Equal(t, testutil.InputItemList[i].Value+"-fast", res.Value)
	}

	err = postTxn.Commit()
	assert.NoError(t, err)

}
