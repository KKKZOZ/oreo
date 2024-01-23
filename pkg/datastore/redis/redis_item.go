package redis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/txn"
)

var _ txn.DataItem = (*RedisItem)(nil)

type RedisItem struct {
	RKey       string       `redis:"Key" json:"Key"`
	RValue     string       `redis:"Value" json:"Value"`
	RTxnId     string       `redis:"TxnId" json:"TxnId"`
	RTxnState  config.State `redis:"TxnState" json:"TxnState"`
	RTValid    time.Time    `redis:"TValid" json:"TValid"`
	RTLease    time.Time    `redis:"TLease" json:"TLease"`
	RPrev      string       `redis:"Prev" json:"Prev"`
	RLinkedLen int          `redis:"LinkedLen" json:"LinkedLen"`
	RIsDeleted bool         `redis:"IsDeleted" json:"IsDeleted"`
	RVersion   int          `redis:"Version" json:"Version"`
}

func NewRedisItem(options txn.ItemOptions) *RedisItem {
	return &RedisItem{
		RKey:       options.Key,
		RValue:     options.Value,
		RTxnId:     options.TxnId,
		RTxnState:  options.TxnState,
		RTValid:    options.TValid,
		RTLease:    options.TLease,
		RPrev:      options.Prev,
		RLinkedLen: options.LinkedLen,
		RIsDeleted: options.IsDeleted,
		RVersion:   options.Version,
	}
}

func (r *RedisItem) Key() string {
	return r.RKey
}

func (r *RedisItem) Value() string {
	return r.RValue
}

func (r *RedisItem) SetValue(v string) {
	r.RValue = v
}

func (r *RedisItem) TxnId() string {
	return r.RTxnId
}

func (r *RedisItem) TxnState() config.State {
	return r.RTxnState
}

func (r *RedisItem) SetTxnState(s config.State) {
	r.RTxnState = s
}

func (r *RedisItem) TValid() time.Time {
	return r.RTValid
}

func (r *RedisItem) SetTValid(t time.Time) {
	r.RTValid = t
}

func (r *RedisItem) TLease() time.Time {
	return r.RTLease
}

func (r *RedisItem) SetTLease(t time.Time) {
	r.RTLease = t
}

func (r *RedisItem) Prev() string {
	return r.RPrev
}

func (r *RedisItem) SetPrev(p string) {
	r.RPrev = p
}

func (r *RedisItem) LinkedLen() int {
	return r.RLinkedLen
}

func (r *RedisItem) SetLinkedLen(l int) {
	r.RLinkedLen = l
}

func (r *RedisItem) IsDeleted() bool {
	return r.RIsDeleted
}

func (r *RedisItem) SetIsDeleted(d bool) {
	r.RIsDeleted = d
}

func (r *RedisItem) Version() int {
	return r.RVersion
}

func (r *RedisItem) SetVersion(v int) {
	r.RVersion = v
}

func (r RedisItem) String() string {
	return fmt.Sprintf(`RedisItem{
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
}`, r.RKey, r.RValue, r.RTxnId, util.ToString(r.RTxnState),
		r.RTValid.Format(time.RFC3339), r.RTLease.Format(time.RFC3339),
		r.RPrev, r.RLinkedLen, r.RIsDeleted, r.RVersion)
}

func (r *RedisItem) Empty() bool {
	return r.RKey == "" && r.RValue == "" &&
		r.RTxnId == "" && r.RTxnState == config.State(0) &&
		r.RTValid.IsZero() && r.RTLease.IsZero() &&
		r.RPrev == "" && r.RLinkedLen == 0 &&
		!r.RIsDeleted && r.RVersion == 0
}

func (r *RedisItem) Equal(other txn.DataItem) bool {
	return r.Key() == other.Key() &&
		r.Value() == other.Value() &&
		r.TxnId() == other.TxnId() &&
		r.TxnState() == other.TxnState() &&
		r.TValid().Equal(other.TValid()) &&
		r.TLease().Equal(other.TLease()) &&
		r.Prev() == other.Prev() &&
		r.LinkedLen() == other.LinkedLen() &&
		r.IsDeleted() == other.IsDeleted() &&
		r.Version() == other.Version()
}

func (r RedisItem) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}
