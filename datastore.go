package main

type Datastore interface {
	// Start a transaction, including initializing the connection
	Start() error
	// Read a record, if the record is not in the cache, read from the connection,
	// then put it into the cache
	Read(key string, value any) error
	Write(key string, value any) error
	Prev(key string, record string)
	Delete(key string) error
	Prepare() error
	Commit() error
	// abort the transaction
	Abort(hasCommitted bool) error
	Recover(key string)

	GetName() string
	SetTxn(txn *Transaction)
	WriteTSR(key string) error
	DeleteTSR(key string) error
}

type dataStore struct {
	Name string
	Txn  *Transaction
}

type State int

const (
	EMPTY     State = 0
	STARTED   State = 1
	PREPARED  State = 2
	COMMITTED State = 3
	ABORTED   State = 4
)
