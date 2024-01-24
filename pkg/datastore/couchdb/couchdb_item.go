package couchdb

import (
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/txn"
)

var _ txn.DataItem = (*CouchDBItem)(nil)

type CouchDBItem struct {
	CKey       string       `json:"Key"`
	CValue     string       `json:"Value"`
	CTxnId     string       `json:"TxnId"`
	CTxnState  config.State `json:"State"`
	CTValid    time.Time    `json:"TValid"`
	CTLease    time.Time    `json:"TLease"`
	CPrev      string       `json:"Prev"`
	CLinkedLen int          `json:"LinkedLen"`
	CIsDeleted bool         `json:"IsDeleted"`
	CVersion   int          `json:"Version"`
	Rev        string       `json:"_rev,omitempty"`
}

func NewCouchDBItem(options txn.ItemOptions) *CouchDBItem {
	return &CouchDBItem{
		CKey:       options.Key,
		CValue:     options.Value,
		CTxnId:     options.TxnId,
		CTxnState:  options.TxnState,
		CTValid:    options.TValid,
		CTLease:    options.TLease,
		CPrev:      options.Prev,
		CLinkedLen: options.LinkedLen,
		CIsDeleted: options.IsDeleted,
		CVersion:   options.Version,
	}
}

func (c *CouchDBItem) Key() string {
	return c.CKey
}

func (c *CouchDBItem) Value() string {
	return c.CValue
}

func (c *CouchDBItem) SetValue(value string) {
	c.CValue = value
}

func (c *CouchDBItem) TxnId() string {
	return c.CTxnId
}

func (c *CouchDBItem) TxnState() config.State {
	return c.CTxnState
}

func (c *CouchDBItem) SetTxnState(state config.State) {
	c.CTxnState = state
}

func (c *CouchDBItem) TValid() time.Time {
	return c.CTValid
}

func (c *CouchDBItem) SetTValid(tValid time.Time) {
	c.CTValid = tValid
}

func (c *CouchDBItem) TLease() time.Time {
	return c.CTLease
}

func (c *CouchDBItem) SetTLease(tLease time.Time) {
	c.CTLease = tLease
}

func (c *CouchDBItem) Prev() string {
	return c.CPrev
}

func (c *CouchDBItem) SetPrev(prev string) {
	c.CPrev = prev
}

func (c *CouchDBItem) LinkedLen() int {
	return c.CLinkedLen
}

func (c *CouchDBItem) SetLinkedLen(linkedLen int) {
	c.CLinkedLen = linkedLen
}

func (c *CouchDBItem) IsDeleted() bool {
	return c.CIsDeleted
}

func (c *CouchDBItem) SetIsDeleted(isDeleted bool) {
	c.CIsDeleted = isDeleted
}

func (c *CouchDBItem) Version() int {
	return c.CVersion
}

func (c *CouchDBItem) SetVersion(version int) {
	c.CVersion = version
}

func (c *CouchDBItem) Equal(other txn.DataItem) bool {
	if other == nil {
		return false
	}
	otherItem, ok := other.(*CouchDBItem)
	if !ok {
		return false
	}

	// Compare properties.
	return c.Key() == otherItem.Key() &&
		c.Value() == otherItem.Value() &&
		c.TxnId() == otherItem.TxnId() &&
		c.TxnState() == otherItem.TxnState() &&
		c.TValid().Equal(otherItem.TValid()) &&
		c.TLease().Equal(otherItem.TLease()) &&
		c.Prev() == otherItem.Prev() &&
		c.LinkedLen() == otherItem.LinkedLen() &&
		c.IsDeleted() == otherItem.IsDeleted() &&
		c.Version() == otherItem.Version()
}

func (c *CouchDBItem) Empty() bool {
	return c == nil || (c.Key() == "" && c.Value() == "")
}

func (c *CouchDBItem) String() string {
	return fmt.Sprintf(`MongoItem{
    Key:       %s,
    Value:     %s,
    TxnId:     %s,
    TxnState:  %s,
    TValid:    %s,
    TLease:    %s,
    Prev:      %s,
	LinkedLen: %d,
    IsDeleted: %v,
    Version:   %d,
}`, c.CKey, c.CValue, c.CTxnId, util.ToString(c.CTxnState),
		c.CTValid.Format(time.RFC3339), c.CTLease.Format(time.RFC3339),
		c.CPrev, c.CLinkedLen, c.CIsDeleted, c.CVersion)
}
