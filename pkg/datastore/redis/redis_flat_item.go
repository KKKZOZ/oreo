package redis

import (
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
)

type FlatDataItem struct {
	FKey     string
	ItemList []LogicItem
}

type LogicItem struct {
	Value        string
	TxnState     config.State
	GroupKeyList string
	Version      string
	TCommit      int64
	TLease       time.Time
	IsDeleted    bool
}

// Insert into head
func (f *FlatDataItem) NewLogicItem() {
	f.ItemList = append([]LogicItem{{}}, f.ItemList...)
	if len(f.ItemList) > config.Config.MaxRecordLength {
		f.ItemList = f.ItemList[:len(f.ItemList)-1]
	}
}

func (f *FlatDataItem) Key() string {
	return f.FKey
}

func (f *FlatDataItem) Value() string {
	return f.ItemList[0].Value
}

func (f *FlatDataItem) SetValue(value string) {
	f.ItemList[0].Value = value
}

func (f *FlatDataItem) TxnState() config.State {
	return f.ItemList[0].TxnState
}

func (f *FlatDataItem) SetTxnState(txnState config.State) {
	f.ItemList[0].TxnState = txnState
}

func (f *FlatDataItem) TValid() int64 {
	return f.ItemList[0].TCommit
}

func (f *FlatDataItem) SetTValid(tValid int64) {
	f.ItemList[0].TCommit = tValid
}

func (f *FlatDataItem) TLease() time.Time {
	return f.ItemList[0].TLease
}

func (f *FlatDataItem) SetTLease(tLease time.Time) {
	f.ItemList[0].TLease = tLease
}

func (f *FlatDataItem) Version() string {
	return f.ItemList[0].Version
}

func (f *FlatDataItem) SetVersion(version string) {
	f.ItemList[0].Version = version
}

func (f *FlatDataItem) IsDeleted() bool {
	return f.ItemList[0].IsDeleted
}

func (f *FlatDataItem) SetIsDeleted(isDeleted bool) {
	f.ItemList[0].IsDeleted = isDeleted
}

func (f *FlatDataItem) GroupKeyList() string {
	return f.ItemList[0].GroupKeyList
}

func (f *FlatDataItem) SetGroupKeyList(groupKeyList string) {
	f.ItemList[0].GroupKeyList = groupKeyList
}

func (f *FlatDataItem) Equal(other FlatDataItem) bool {
	return f.FKey == other.FKey
}

func (f *FlatDataItem) Empty() bool {
	return f.FKey == ""
}
