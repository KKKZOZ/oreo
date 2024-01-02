package txn

import "github.com/kkkzoz/oreo/pkg/config"

//go:generate mockery --name Datastore
type TSRMaintainer interface {
	// ReadTSR reads the transaction state record (TSR) for a transaction.
	ReadTSR(txnId string) (config.State, error)

	// WriteTSR writes the transaction state record (TSR) for a transaction.
	WriteTSR(txnId string, txnState config.State) error

	// DeleteTSR deletes the transaction state record (TSR) for a transaction.
	DeleteTSR(txnId string) error
}
