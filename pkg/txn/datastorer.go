package txn

// Datastorer is an interface that defines the operations for interacting with a data store.
//
//go:generate mockery --name Datastore
type Datastorer interface {
	// Start starts a transaction, including initializing the connection.
	Start() error

	// Read reads a record from the data store. If the record is not in the cache (readCache/writeCache),
	// it reads the record from the connection and puts it into the cache.
	Read(key string, value any) error

	// Write writes records into the writeCache.
	Write(key string, value any) error

	// Delete marks a record as deleted.
	Delete(key string) error

	// Prepare executes the prepare phase of transaction commit.
	// In Oreo, it will return TCommit as well

	Prepare() (int64, error)

	// Commit executes the commit phase of transaction commit.
	// It updates the records in the writeCache to the COMMITTED state
	// in the data store.
	Commit() error

	// Abort aborts the transaction.
	// It rolls back the records in the writeCache to the state before the transaction.
	Abort(hasCommitted bool) error

	// OnePhaseCommit executes the one-phase commit protocol.
	OnePhaseCommit() error

	// GetName returns the name of the data store.
	GetName() string

	// SetTxn sets the current transaction for the data store.
	SetTxn(txn *Transaction)

	Copy() Datastorer
}
