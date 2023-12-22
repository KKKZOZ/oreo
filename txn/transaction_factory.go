package txn

import (
	"errors"

	"github.com/kkkzoz/oreo/locker"
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

	// LocalLocker is the local locker instance.
	LocalLocker locker.Locker
}

// NewTransactionFactory creates a new TransactionFactory object.
// It initializes the TransactionFactory with default values and returns a pointer to the newly created object.
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

	return txn
}
