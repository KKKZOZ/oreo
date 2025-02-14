package txn

import (
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
)

type DataItemType int

const (
	NestedItemType    DataItemType = iota
	FlattenedItemType DataItemType = iota
)

type DataItemNext interface {
	TxnMetadata

	TxnValueOperator
}

type TxnMetadata interface {
	Key() string

	GroupKeyList() string
	SetGroupKeyList(string)

	TxnState() config.State
	SetTxnState(config.State)

	Version() string
	SetVersion(string)

	TValid() int64
	SetTValid(int64)

	TLease() time.Time
	SetTLease(time.Time)

	IsDeleted() bool
	SetIsDeleted(bool)
	Empty() bool
}

type TxnValueOperator interface {
	// Deserialize the value from the given pointer
	ParseValue(any) error
	// Serialize the value to []byte
	SetValue(any) error

	UpdateMetadata(DataItemNext, int64, time.Time) error

	// if the valid item is found in the logical chain, return the item and true
	//
	// else return nil and false
	GetValidItem(TStart int64) (DataItemNext, bool)

	// - if prev is empty, return "key not found in AVC" Error
	//
	// - if deserialize failed, return "deserialize failed" Error
	//
	// Note the prev item contains the remaining logical chain
	GetPrevItem() (DataItemNext, error)
}

type ItemOptionsNext struct {
	Key          string
	Value        any
	GroupKeyList string
	TxnState     config.State
	TValid       int64
	TLease       time.Time
	Prev         []byte
	LinkedLen    int
	IsDeleted    bool
	Version      string

	DataItem DataItem
}

type DataItemFactoryNext interface {
	NewDataItem(ItemOptionsNext) DataItemNext
}
