package txn

import "github.com/kkkzoz/vanilla-icecream/locker"

type TransactionFactory struct {
	oracleURL string
	locker    locker.Locker
}

type TransactionConfig struct {
	TimeOracleSource SourceType
	LockerSource     SourceType
	Options          map[string]string
}

// NewTransactionFactory creates a new TransactionFactory object.
// It initializes the TransactionFactory with default values and returns a pointer to the newly created object.
func NewTransactionFactory(config *TransactionConfig) *TransactionFactory {
	if config == nil {
		config = &TransactionConfig{
			TimeOracleSource: LOCAL,
			LockerSource:     LOCAL,
		}
	}
	return &TransactionFactory{
		oracleURL: "",
		locker:    locker.NewMemoryLocker(),
	}
}
