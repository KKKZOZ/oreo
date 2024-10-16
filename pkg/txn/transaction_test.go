package txn

import (
	"testing"

	"github.com/oreo-dtx-lab/oreo/internal/testutil"
)

func NewTransactionWithSetup() *Transaction {
	txn := NewTransaction()
	mds := &MockDatastore{}
	mds.On("GetName").Return("memory")
	mds.On("SetTxn", txn).Return(nil)
	mds.On("Start").Return(nil)
	txn.AddDatastore(mds)
	txn.SetGlobalDatastore(mds)
	return txn
}

// TestTxnStartAgain tests the behavior of starting a transaction multiple times.
func TestTxnStartAgain(t *testing.T) {
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
	txn := NewTransactionWithSetup()
	err := txn.Commit()
	if err == nil {
		t.Errorf("Expected error committing transaction")
	}
}

// TestTxnAbortWithoutStart tests the behavior of aborting a transaction without starting it.
func TestTxnAbortWithoutStart(t *testing.T) {
	txn := NewTransactionWithSetup()
	err := txn.Abort()
	if err == nil {
		t.Errorf("Expected error aborting transaction")
	}
}

// TestTxnOperateWithoutStart tests the behavior of transaction operations
// when the database has not been started.
func TestTxnOperateWithoutStart(t *testing.T) {

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
