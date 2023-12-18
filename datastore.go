package main

type Datastore interface {
	// Start a transaction, including initializing the connection
	Start() error
	Read(key string, value any) error
	Write(key string, value any) error
	Prev(key string, record string)
	Delete(key string) error
	Prepare() error
	Commit() error
	// abort the transaction
	Abort() error
	Recover(key string)

	GetName() string
	SetTxn(txn *Transaction)
}

type dataStore struct {
	Name string
	Txn  *Transaction
}

type State int

const (
	PREPARED  State = 0
	COMMITTED State = 1
)
