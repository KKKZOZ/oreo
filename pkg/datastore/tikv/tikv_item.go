package tikv

import (
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ txn.DataItem = (*TiKVItem)(nil)

type TiKVItem struct {
	KKey          string       `json:"Key"`
	KValue        string       `json:"Value"`
	KGroupKeyList string       `json:"GroupKeyList"`
	KTxnState     config.State `json:"State"`
	KTValid       int64        `json:"TValid"`
	KTLease       time.Time    `json:"TLease"`
	KPrev         string       `json:"Prev"`
	KLinkedLen    int          `json:"LinkedLen"`
	KIsDeleted    bool         `json:"IsDeleted"`
	KVersion      string       `json:"Version,omitempty"`
}

func NewTiKVItem(options txn.ItemOptions) *TiKVItem {
	if options.Value == nil {
		options.Value = ""
	}

	return &TiKVItem{
		KKey:          options.Key,
		KValue:        options.Value.(string),
		KGroupKeyList: options.GroupKeyList,
		KTxnState:     options.TxnState,
		KTValid:       options.TValid,
		KTLease:       options.TLease,
		KPrev:         options.Prev,
		KLinkedLen:    options.LinkedLen,
		KIsDeleted:    options.IsDeleted,
		KVersion:      options.Version,
	}
}

func (k *TiKVItem) Key() string {
	return k.KKey
}

func (k *TiKVItem) Value() string {
	return k.KValue
}

func (k *TiKVItem) SetValue(value string) {
	k.KValue = value
}

func (k *TiKVItem) GroupKeyList() string {
	return k.KGroupKeyList
}

func (k *TiKVItem) SetGroupKeyList(groupKeyList string) {
	k.KGroupKeyList = groupKeyList
}

func (k *TiKVItem) TxnState() config.State {
	return k.KTxnState
}

func (k *TiKVItem) SetTxnState(state config.State) {
	k.KTxnState = state
}

func (k *TiKVItem) TValid() int64 {
	return k.KTValid
}

func (k *TiKVItem) SetTValid(tValid int64) {
	k.KTValid = tValid
}

func (k *TiKVItem) TLease() time.Time {
	return k.KTLease
}

func (k *TiKVItem) SetTLease(tLease time.Time) {
	k.KTLease = tLease
}

func (k *TiKVItem) Prev() string {
	return k.KPrev
}

func (k *TiKVItem) SetPrev(prev string) {
	k.KPrev = prev
}

func (k *TiKVItem) LinkedLen() int {
	return k.KLinkedLen
}

func (k *TiKVItem) SetLinkedLen(linkedLen int) {
	k.KLinkedLen = linkedLen
}

func (k *TiKVItem) IsDeleted() bool {
	return k.KIsDeleted
}

func (k *TiKVItem) SetIsDeleted(isDeleted bool) {
	k.KIsDeleted = isDeleted
}

func (k *TiKVItem) Version() string {
	return k.KVersion
}

func (k *TiKVItem) SetVersion(version string) {
	k.KVersion = version
}

func (k *TiKVItem) Equal(other txn.DataItem) bool {
	if other == nil {
		return false
	}
	otherItem, ok := other.(*TiKVItem)
	if !ok {
		return false
	}

	return k.Key() == otherItem.Key() &&
		k.Value() == otherItem.Value() &&
		k.GroupKeyList() == otherItem.GroupKeyList() &&
		k.TxnState() == otherItem.TxnState() &&
		k.TValid() == otherItem.TValid() &&
		k.TLease().Equal(otherItem.TLease()) &&
		k.Prev() == otherItem.Prev() &&
		k.LinkedLen() == otherItem.LinkedLen() &&
		k.IsDeleted() == otherItem.IsDeleted() &&
		k.Version() == otherItem.Version()
}

func (k *TiKVItem) Empty() bool {
	return k == nil || (k.Key() == "" && k.Value() == "")
}

func (k *TiKVItem) String() string {
	return fmt.Sprintf(`TiKVItem{
    Key:       %s,
    Value:     %s,
    GroupKeyList:     %s,
    TxnState:  %s,
    TValid:    %v,
    TLease:    %s,
    Prev:      %s,
    LinkedLen: %d,
    IsDeleted: %v,
    Version:   %s,
}`, k.KKey, k.KValue, k.KGroupKeyList, util.ToString(k.KTxnState),
		k.KTValid, k.KTLease.Format(time.RFC3339),
		k.KPrev, k.KLinkedLen, k.KIsDeleted, k.KVersion)
}
