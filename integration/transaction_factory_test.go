package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/internal/testutil"
	"github.com/kkkzoz/oreo/pkg/datastore/memory"
	"github.com/kkkzoz/oreo/pkg/factory"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/stretchr/testify/assert"
)

// Testing cases for the NewTransactionFactory method
func TestNewTransactionFactory(t *testing.T) {
	memDst1 := memory.NewMemoryDatastore("mem1", memory.NewMemoryConnection("localhost", 8321))
	memDst2 := memory.NewMemoryDatastore("mem2", memory.NewMemoryConnection("localhost", 8321))
	testCases := []struct {
		name        string
		config      *factory.TransactionConfig
		expectedErr error
	}{

		{
			name: "Test case 1: with empty DatastoreList",
			config: &factory.TransactionConfig{
				DatastoreList: []txn.Datastorer{},
			},
			expectedErr: errors.New("DatastoreList is empty"),
		},
		{
			name: "Test case 2: with nil GlobalDatastore",
			config: &factory.TransactionConfig{
				DatastoreList: []txn.Datastorer{memDst1, memDst2},
			},
			expectedErr: errors.New("GlobalDatastore is empty"),
		},
		{
			name: "Test case 3: with empty OracleURL",
			config: &factory.TransactionConfig{
				DatastoreList:    []txn.Datastorer{memDst1, memDst2},
				GlobalDatastore:  memDst1,
				TimeOracleSource: txn.GLOBAL,
			},
			expectedErr: errors.New("OracleURL is empty"),
		},
		{
			name: "Test case 4: with local locker source and global time oracle source",
			config: &factory.TransactionConfig{
				DatastoreList:    []txn.Datastorer{memDst1},
				GlobalDatastore:  memDst1,
				OracleURL:        "http://localhost:8300",
				TimeOracleSource: txn.GLOBAL,
			},
			expectedErr: errors.New("LockerSource must be GLOBAL when using a global time oracle"),
		},
		{
			name: "Test case 5: with global locker source and empty OracleURL",
			config: &factory.TransactionConfig{
				DatastoreList:   []txn.Datastorer{memDst1, memDst2},
				GlobalDatastore: memDst1,
				LockerSource:    txn.GLOBAL,
			},
			expectedErr: errors.New("OracleURL is empty"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := factory.NewTransactionFactory(tc.config)
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
	memDst1 := memory.NewMemoryDatastore("memory", memory.NewMemoryConnection("localhost", 8321))

	txnFactory, err := factory.NewTransactionFactory(&factory.TransactionConfig{
		DatastoreList:    []txn.Datastorer{memDst1},
		GlobalDatastore:  memDst1,
		TimeOracleSource: txn.LOCAL,
		LockerSource:     txn.LOCAL,
	})
	assert.NoError(t, err)

	txn1 := txnFactory.NewTransaction()
	txn1.Start()
	person := testutil.NewDefaultPerson()
	txn1.Write("memory", "John", person)
	err = txn1.Commit()
	assert.NoError(t, err)

	txn2 := txnFactory.NewTransaction()
	txn2.Start()
	var person2 testutil.Person
	txn2.Read("memory", "John", &person2)
	person2.Age = 31
	txn2.Write("memory", "John", person2)
	err = txn2.Commit()
	assert.NoError(t, err)

	postTxn := txnFactory.NewTransaction()
	postTxn.Start()
	var person3 testutil.Person
	postTxn.Read("memory", "John", &person3)
	assert.Equal(t, 31, person3.Age)
	err = postTxn.Commit()
	assert.NoError(t, err)
}

// TODO: This one will fail due to the fact that current implementation of datastore
// is not thread safe.
// TestConcurrentTransactionCreatedByFactory tests the concurrent creation of transactions using a transaction factory.
func TestConcurrentTransactionCreatedByFactory(t *testing.T) {
	// Create a new memory database instance
	memoryDatabase := memory.NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	// Create a new memory datastore instance
	memDst1 := memory.NewMemoryDatastore("mem1", memory.NewMemoryConnection("localhost", 8321))

	txnFactory, err := factory.NewTransactionFactory(&factory.TransactionConfig{
		DatastoreList:    []txn.Datastorer{memDst1},
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
			txn := txnFactory.NewTransaction()
			txn.Start()
			var person testutil.Person
			txn.Read("mem1", "John", &person)
			person.Age = person.Age + id
			txn.Write("mem1", "John", person)
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
