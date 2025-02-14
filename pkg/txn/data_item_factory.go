package txn

import (
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
)

type ItemOptions struct {
	Key          string
	Value        string
	GroupKeyList string
	TxnState     config.State
	TValid       int64
	TLease       time.Time
	Prev         string
	LinkedLen    int
	IsDeleted    bool
	Version      string

	DataItem DataItem
}

type DataItemFactory interface {
	NewDataItem(ItemOptions) DataItem
}
