package txn

import "github.com/oreo-dtx-lab/oreo/pkg/config"

//go:generate mockery --name Datastore
type GroupKeyMaintainer interface {
	// ReadGroupKey reads the transaction state record (TSR) for a transaction.
	ReadGroupKey(key string) (GroupKey, error)

	// WriteGroupKey writes the transaction state record (TSR) for a transaction.
	WriteGroupKey(key string, txnState config.State, tCommit int64) error

	// CreateGroupKey atomically create the transaction state record (TSR) for a transaction.
	// if the Group Key already exists, it will return an error.
	CreateGroupKey(key string, txnState config.State, tCommit int64) (config.State, error)

	// DeleteTSR deletes the transaction state record (TSR) for a transaction.
	DeleteGroupKey(key string) error
}
