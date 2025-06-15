package txn

import (
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
)

type ItemOptions struct {
	Key          string
	Value        any
	GroupKeyList string
	TxnState     config.State
	TValid       int64
	TLease       time.Time
	Prev         string
	LinkedLen    int
	IsDeleted    bool
	Version      string
}

type DataItemFactory interface {
	NewDataItem(ItemOptions) DataItem
}
