package txn

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

func NewFlatDataItem(options ItemOptions) *FlatDataItem {
	f := options.DataItem.(*FlatDataItem)
	logicItem := LogicItem{
		Value:        options.Value,
		TxnState:     options.TxnState,
		GroupKeyList: options.GroupKeyList,
		Version:      options.Version,
		TCommit:      options.TValid,
		TLease:       options.TLease,
		IsDeleted:    options.IsDeleted,
	}
	f.ItemList = append([]LogicItem{logicItem}, f.ItemList...)
	if len(f.ItemList) > config.Config.MaxRecordLength {
		f.ItemList = f.ItemList[:len(f.ItemList)-1]
	}
	return f
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

func (f *FlatDataItem) LinkedLen() int {
	return len(f.ItemList)
}

func (f *FlatDataItem) SetLinkedLen(l int) {
}

func (f *FlatDataItem) Prev() string {
	return ""
}

func (f *FlatDataItem) SetPrev(p string) {
}

func (f *FlatDataItem) Equal(other DataItem) bool {
	return f.FKey == other.Key()
}

func (f *FlatDataItem) Empty() bool {
	return f.FKey == ""
}
