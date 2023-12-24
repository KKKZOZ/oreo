package txn

// Datastore is an interface that defines the operations for interacting with a data store.
//
//go:generate mockery --name Datastore
type Datastore interface {
	// Start starts a transaction, including initializing the connection.
	Start() error

	// Read reads a record from the data store. If the record is not in the cache (readCache/writeCache),
	// it reads the record from the connection and puts it into the cache.
	Read(key string, value any) error

	// Write writes records into the writeCache.
	Write(key string, value any) error

	// Prev retrieves the previous value of a record.
	Prev(key string, record string)

	// Delete marks a record as deleted.
	Delete(key string) error

	// Prepare executes the prepare phase of transaction commit.
	// It first marks the records in the writeCache with T_commit, TxnId, and TxnState,
	// then it performs `conditionalUpdate` in a global order.
	Prepare() error

	// Commit executes the commit phase of transaction commit.
	// It updates the records in the writeCache to the COMMITTED state
	// in the data store.
	Commit() error

	// Abort aborts the transaction.
	// It rolls back the records in the writeCache to the state before the transaction.
	Abort(hasCommitted bool) error

	// Recover recovers a record.
	Recover(key string)

	// GetName returns the name of the data store.
	GetName() string

	// SetTxn sets the current transaction for the data store.
	SetTxn(txn *Transaction)

	Copy() Datastore
}

// BaseDataStore represents a base data store with a name and a transaction.
type BaseDataStore struct {
	Name string
	Txn  *Transaction
}

type Item interface {
	GetKey() string
}
