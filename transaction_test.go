package main

import (
	"testing"
	"time"
)

func NewTransactionWithSetup() *Transaction {
	txn := NewTransaction()
	conn := NewMemoryConnection("localhost", 8321)
	mds := NewMemoryDatastore("memory", conn)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)
	return txn
}

func TestTxnStartAgain(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	txn := NewTransactionWithSetup()
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	err = txn.Start()
	if err == nil {
		t.Errorf("Expected error starting transaction")
	}
}

func TestTxnCommitWithoutStart(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	txn := NewTransactionWithSetup()
	err := txn.Commit()
	if err == nil {
		t.Errorf("Expected error committing transaction")
	}
}

func TestTxnAbortWithoutStart(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	txn := NewTransactionWithSetup()
	err := txn.Abort()
	if err == nil {
		t.Errorf("Expected error aborting transaction")
	}
}

func TestTxnOperateWithoutStart(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	txn := NewTransactionWithSetup()
	var person Person
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
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	txn1 := NewTransactionWithSetup()

	expected := NewDefaultPerson()

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
	var actual Person
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
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup()

	// Txn reads the record
	err := txn.Start()
	if err != nil {
		t.Errorf("Error starting transaction: %s", err)
	}
	var person Person
	err = txn.Read("memory", "John", &person)
	if err != nil {
		t.Errorf("Error reading record: %s", err)
	}

	expected := person
	expected.Age = 31
	txn.Write("memory", "John", expected)

	var actual Person
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
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	txn.Start()
	var person Person
	txn.Read("memory", "John", &person)
	person.Age = 31
	txn.Write("memory", "John", person)
	var anotherPerson Person
	txn.Read("memory", "John", &anotherPerson)

	if person != anotherPerson {
		t.Errorf("Expected two read to be the same")
	}
	txn.Commit()

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	var postPerson Person
	postTxn.Read("memory", "John", &postPerson)
	if postPerson != person {
		t.Errorf("got %v want %v", postPerson, person)
	}

}

func TestMultileKeyWriteConflict(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Write("memory", "Jane", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var person Person
		txn1.Read("memory", "John", &person)
		person.Age = 31
		txn1.Write("memory", "John", person)

		txn1.Read("memory", "Jane", &person)
		person.Age = 31
		txn1.Write("memory", "Jane", person)

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
		var person Person
		txn2.Read("memory", "Jane", &person)
		person.Age = 32
		txn2.Write("memory", "Jane", person)

		txn2.Read("memory", "John", &person)
		person.Age = 32
		txn2.Write("memory", "John", person)

		err := txn2.Commit()
		if err != nil {
			resChan <- false
		} else {
			resChan <- true
		}
	}()

	res1 := <-resChan
	res2 := <-resChan

	// only one transaction should succeed
	if res1 != res2 {
		t.Errorf("Expected only one transaction to succeed")
	}
}

func TestRepeatableReadWhenRecordDeleted(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	manualTxn := NewTransactionWithSetup()
	txn.Start()
	manualTxn.Start()

	var person1 Person
	txn.Read("memory", "John", &person1)

	// manualTxn deletes John and commits
	manualTxn.Delete("memory", "John")
	manualTxn.Commit()

	var person2 Person
	txn.Read("memory", "John", &person2)

	// two read in txn should be the same
	if person1 != person2 {
		t.Errorf("Expected two read in txn to be the same")
	}
}

func TestRepeatableReadWhenRecordUpdatedTwice(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	manualTxn1 := NewTransactionWithSetup()
	txn.Start()
	manualTxn1.Start()

	var person1 Person
	txn.Read("memory", "John", &person1)

	// manualTxn1 updates John and commits
	var manualPerson1 Person
	manualTxn1.Read("memory", "John", &manualPerson1)
	manualPerson1.Age = 31
	manualTxn1.Write("memory", "John", manualPerson1)
	manualTxn1.Commit()

	manualTxn2 := NewTransactionWithSetup()
	manualTxn2.Start()
	// manualTxn updates John again and commits
	var manualPerson2 Person
	manualTxn2.Read("memory", "John", &manualPerson2)
	manualPerson2.Age = 32
	manualTxn2.Write("memory", "John", manualPerson2)
	manualTxn2.Commit()

	var person2 Person
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
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var person Person
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
		var person1 Person
		// txn2 reads John
		txn2.Read("memory", "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 Person
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
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	dataPerson := NewDefaultPerson()
	preTxn.Start()
	preTxn.Write("memory", "John", dataPerson)
	preTxn.Commit()

	resChan := make(chan bool)

	go func() {
		txn1 := NewTransactionWithSetup()
		txn1.Start()
		var person Person
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
		var person1 Person
		// txn2 reads John
		txn2.Read("memory", "John", &person1)
		time.Sleep(100 * time.Millisecond)

		var person2 Person
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
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	expected := NewDefaultPerson()
	preTxn.Write("memory", "John", expected)
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	var person Person
	txn.Start()
	txn.Read("memory", "John", &person)
	person.Age = 31
	txn.Write("memory", "John", person)
	txn.Abort()

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	var postPerson Person
	postTxn.Read("memory", "John", &postPerson)
	postTxn.Commit()
	if postPerson != expected {
		t.Errorf("got %v want %v", postPerson, expected)
	}
}

var inputItemList = []TestItem{
	NewTestItem("item1"),
	NewTestItem("item2"),
	NewTestItem("item3"),
	NewTestItem("item4"),
	NewTestItem("item5"),
}

func TestTxnAbortCausedByWriteConflict(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()
	defer func() { <-memoryDatabase.msgChan }()
	defer func() { go memoryDatabase.stop() }()
	time.Sleep(100 * time.Millisecond)

	preTxn := NewTransactionWithSetup()
	preTxn.Start()
	for _, item := range inputItemList {
		preTxn.Write("memory", item.Value, item)
	}
	preTxn.Commit()

	txn := NewTransactionWithSetup()
	manualTxn := NewTransactionWithSetup()
	txn.Start()
	manualTxn.Start()

	// txn reads all items and modify them
	for _, item := range inputItemList {
		var actual TestItem
		txn.Read("memory", item.Value, &actual)
		actual.Value = item.Value + "updated"
		txn.Write("memory", item.Value, actual)
	}

	// manualTxn reads one item and modify it
	var manualItem TestItem
	manualTxn.Read("memory", "item4", &manualItem)
	manualItem.Value = "item4updated"
	manualTxn.Write("memory", "item4", manualItem)
	manualTxn.Commit()

	err := txn.Commit()
	if err == nil {
		t.Errorf("Expected error committing transaction")
	}

	postTxn := NewTransactionWithSetup()
	postTxn.Start()
	for _, item := range inputItemList {
		var actual TestItem
		postTxn.Read("memory", item.Value, &actual)
		if item.Value != "item4" {
			if actual != item {
				t.Errorf("got %v want %v", actual, item)
			}
		} else {
			if actual != manualItem {
				t.Errorf("got %v want %v", actual, manualItem)
			}
		}
	}
	postTxn.Commit()
}
