package main

type Datastore interface {
	// start a transaction, including initializing the connection
	start() error
	read(key string, value any) error
	write(key string, value any) error
	prev(key string, record string)
	delete(key string)
	prepare() error
	commit() error
	// abort the transaction
	abort()
	recover(key string)

	getType() DataStoreType
	setTxn(txn *Transaction)
}

type dataStore struct {
	Type DataStoreType
	Txn  *Transaction
}

type State int

const (
	PREPARED  State = 0
	COMMITTED State = 1
)
