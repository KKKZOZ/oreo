package factory

import (
	"errors"

	"github.com/kkkzoz/oreo/pkg/locker"
	"github.com/kkkzoz/oreo/pkg/txn"
)

// TransactionFactory represents a factory for creating transactions.
type TransactionFactory struct {
	// oracleURL is the URL of the oracle used for transactions.
	// for example: http://localhost:8300
	oracleURL string
	// locker is the locker used for transaction synchronization.
	locker locker.Locker
	// TimeOracleSource is the source type for the time oracle.
	TimeOracleSource txn.SourceType
	// LockerSource is the source type for the locker.
	LockerSource txn.SourceType

	dateStoreList   []txn.Datastorer
	globalDatastore txn.Datastorer
}

type TransactionConfig struct {
	// TimeOracleSource specifies the source type for the time oracle.
	TimeOracleSource txn.SourceType

	// LockerSource specifies the source type for the locker.
	LockerSource txn.SourceType

	// OracleURL is the URL of the oracle.
	OracleURL string

	// Options is a map of additional options.
	Options map[string]string

	// DatastoreList is a list of datastores to be added to the transaction.
	DatastoreList []txn.Datastorer

	// GlobalDatastore is the global datastore to be added to the transaction.
	GlobalDatastore txn.Datastorer

	// LocalLocker is the local locker instance.
	LocalLocker locker.Locker
}

// NewTransactionFactory creates a new TransactionFactory object.
func NewTransactionFactory(config *TransactionConfig) (*TransactionFactory, error) {
	if config == nil {
		config = &TransactionConfig{
			TimeOracleSource: txn.LOCAL,
			LockerSource:     txn.LOCAL,
		}
	}
	if config.TimeOracleSource == "" {
		config.TimeOracleSource = txn.LOCAL
	}
	if config.LockerSource == "" {
		config.LockerSource = txn.LOCAL
	}

	if len(config.DatastoreList) == 0 {
		return nil, errors.New("DatastoreList is empty")
	}

	if config.GlobalDatastore == nil {
		return nil, errors.New("GlobalDatastore is empty")
	}

	if config.TimeOracleSource == txn.GLOBAL {
		if config.OracleURL == "" {
			return nil, errors.New("OracleURL is empty")
		}
	}

	if config.LockerSource == txn.GLOBAL {
		if config.OracleURL == "" {
			return nil, errors.New("OracleURL is empty")
		}
	}

	// when using a global time oracle, the locker must also be global
	if config.TimeOracleSource == txn.GLOBAL && config.LockerSource == txn.LOCAL {
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
func (t *TransactionFactory) NewTransaction() *txn.Transaction {
	// By default, time oracle and locker are local
	txn1 := txn.NewTransaction()

	// if t.TimeOracleSource == txn.GLOBAL {
	// 	txn1.SetGlobalTimeSource(t.oracleURL)
	// 	txn1.SetLocker(locker.NewHttpLocker(t.oracleURL))
	// }

	// if t.LockerSource == txn.GLOBAL {
	// 	txn1.SetLocker(locker.NewHttpLocker(t.oracleURL))
	// }

	for _, ds := range t.dateStoreList {
		copy := ds.Copy()
		txn1.AddDatastore(copy)
		if ds == t.globalDatastore {
			txn1.SetGlobalDatastore(copy)
		}
	}

	return txn1
}
