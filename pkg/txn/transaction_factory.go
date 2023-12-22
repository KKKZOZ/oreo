package txn

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/kkkzoz/oreo/pkg/locker"
)

// TransactionFactory represents a factory for creating transactions.
type TransactionFactory struct {
	// oracleURL is the URL of the oracle used for transactions.
	// for example: http://localhost:8300
	oracleURL string
	// locker is the locker used for transaction synchronization.
	locker locker.Locker
	// TimeOracleSource is the source type for the time oracle.
	TimeOracleSource SourceType
	// LockerSource is the source type for the locker.
	LockerSource SourceType

	dateStoreList   []Datastore
	globalDatastore Datastore
}

type TransactionConfig struct {
	// TimeOracleSource specifies the source type for the time oracle.
	TimeOracleSource SourceType

	// LockerSource specifies the source type for the locker.
	LockerSource SourceType

	// OracleURL is the URL of the oracle.
	OracleURL string

	// Options is a map of additional options.
	Options map[string]string

	// DatastoreList is a list of datastores to be added to the transaction.
	DatastoreList []Datastore

	// GlobalDatastore is the global datastore to be added to the transaction.
	GlobalDatastore Datastore

	// LocalLocker is the local locker instance.
	LocalLocker locker.Locker
}

// NewTransactionFactory creates a new TransactionFactory object.
func NewTransactionFactory(config *TransactionConfig) (*TransactionFactory, error) {
	if config == nil {
		config = &TransactionConfig{
			TimeOracleSource: LOCAL,
			LockerSource:     LOCAL,
		}
	}
	if config.TimeOracleSource == "" {
		config.TimeOracleSource = LOCAL
	}
	if config.LockerSource == "" {
		config.LockerSource = LOCAL
	}

	if len(config.DatastoreList) == 0 {
		return nil, errors.New("DatastoreList is empty")
	}

	if config.GlobalDatastore == nil {
		return nil, errors.New("GlobalDatastore is empty")
	}

	if config.TimeOracleSource == GLOBAL {
		if config.OracleURL == "" {
			return nil, errors.New("OracleURL is empty")
		}
	}

	if config.LockerSource == GLOBAL {
		if config.OracleURL == "" {
			return nil, errors.New("OracleURL is empty")
		}
	}

	// when using a global time oracle, the locker must also be global
	if config.TimeOracleSource == GLOBAL && config.LockerSource == LOCAL {
		return nil, errors.New("LockerSource must be GLOBAL when using a global time oracle")
	}

	return &TransactionFactory{
		TimeOracleSource: config.TimeOracleSource,
		LockerSource:     config.LockerSource,
		oracleURL:        config.OracleURL,
		locker:           config.LocalLocker,
		dateStoreList:    config.DatastoreList,
		globalDatastore:  config.GlobalDatastore,
	}, nil

}

// NewTransaction creates a new Transaction object.
func (t *TransactionFactory) NewTransaction() *Transaction {
	// By default, time oracle and locker are local
	txn := NewTransaction()

	if t.TimeOracleSource == GLOBAL {
		txn.oracleURL = t.oracleURL
		txn.timeSource = GLOBAL
		txn.locker = locker.NewHttpLocker(t.oracleURL)
	}

	if t.LockerSource == GLOBAL {
		txn.locker = locker.NewHttpLocker(t.oracleURL)
	}

	for _, ds := range t.dateStoreList {
		// var tar Datastore
		// deepcopy(&ds, &tar)
		txn.AddDatastore(ds)
		if ds == t.globalDatastore {
			txn.SetGlobalDatastore(ds)
		}
	}

	return txn
}

func deepcopy(src, dest *Datastore) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(src)
	if err != nil {
		return err
	}
	return dec.Decode(dest)
}
