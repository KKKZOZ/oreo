package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/pkg/datastore/memory"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

// Testing cases for the NewTransactionFactory method
func TestNewTransactionFactory(t *testing.T) {
	memDst1 := memory.NewMemoryDatastore("mem1", memory.NewMemoryConnection("localhost", 8321))
	memDst2 := memory.NewMemoryDatastore("mem2", memory.NewMemoryConnection("localhost", 8321))
	testCases := []struct {
		name        string
		config      *txn.TransactionConfig
		expectedErr error
	}{

		{
			name: "Test case 1: with empty DatastoreList",
			config: &txn.TransactionConfig{
				DatastoreList: []txn.Datastore{},
			},
			expectedErr: errors.New("DatastoreList is empty"),
		},
		{
			name: "Test case 2: with nil GlobalDatastore",
			config: &txn.TransactionConfig{
				DatastoreList: []txn.Datastore{memDst1, memDst2},
			},
			expectedErr: errors.New("GlobalDatastore is empty"),
		},
		{
			name: "Test case 3: with empty OracleURL",
			config: &txn.TransactionConfig{
				DatastoreList:    []txn.Datastore{memDst1, memDst2},
				GlobalDatastore:  memDst1,
				TimeOracleSource: txn.GLOBAL,
			},
			expectedErr: errors.New("OracleURL is empty"),
		},
		{
			name: "Test case 4: with local locker source and global time oracle source",
			config: &txn.TransactionConfig{
				DatastoreList:    []txn.Datastore{memDst1},
				GlobalDatastore:  memDst1,
				OracleURL:        "http://localhost:8300",
				TimeOracleSource: txn.GLOBAL,
			},
			expectedErr: errors.New("LockerSource must be GLOBAL when using a global time oracle"),
		},
		{
			name: "Test case 5: with global locker source and empty OracleURL",
			config: &txn.TransactionConfig{
				DatastoreList:   []txn.Datastore{memDst1, memDst2},
				GlobalDatastore: memDst1,
				LockerSource:    txn.GLOBAL,
			},
			expectedErr: errors.New("OracleURL is empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := txn.NewTransactionFactory(tc.config)
			if err != nil && err.Error() != tc.expectedErr.Error() {
				t.Errorf("expected error %s, got %s", tc.expectedErr, err)
			}
		})
	}
}

func TestTransactionCreatedByFactory(t *testing.T) {
	// Create a new memory database instance
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new memory datastore instance
	memDst1 := memory.NewMemoryDatastore("mem1", memory.NewMemoryConnection("localhost", 8321))

	txnFactory, err := txn.NewTransactionFactory(&txn.TransactionConfig{
		DatastoreList:    []txn.Datastore{memDst1},
		GlobalDatastore:  memDst1,
		TimeOracleSource: txn.LOCAL,
		LockerSource:     txn.LOCAL,
	})
	assert.NoError(t, err)

	txn1 := txnFactory.NewTransaction()
	txn1.Start()
	person := testutil.NewDefaultPerson()
	txn1.Write("mem1", "John", person)
	err = txn1.Commit()
	assert.NoError(t, err)

	txn2 := txnFactory.NewTransaction()
	txn2.Start()
	var person2 testutil.Person
	txn2.Read("mem1", "John", &person2)
	person2.Age = 31
	txn2.Write("mem1", "John", person2)
	err = txn2.Commit()
	assert.NoError(t, err)

	postTxn := txnFactory.NewTransaction()
	postTxn.Start()
	var person3 testutil.Person
	postTxn.Read("mem1", "John", &person3)
	assert.Equal(t, 31, person3.Age)
	err = postTxn.Commit()
	assert.NoError(t, err)
}

func TestConcurrentTransactionCreatedByFactory(t *testing.T) {

}
