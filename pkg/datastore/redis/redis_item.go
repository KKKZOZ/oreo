package redis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ txn.DataItem = (*RedisItem)(nil)

type RedisItem struct {
	RKey          string       `redis:"Key" json:"Key"`
	RValue        string       `redis:"Value" json:"Value"`
	RGroupKeyList string       `redis:"GroupKeyList" json:"GroupKeyList"`
	RTxnState     config.State `redis:"TxnState" json:"TxnState"`
	RTValid       int64        `redis:"TValid" json:"TValid"`
	RTLease       time.Time    `redis:"TLease" json:"TLease"`
	RPrev         string       `redis:"Prev" json:"Prev"`
	RLinkedLen    int          `redis:"LinkedLen" json:"LinkedLen"`
	RIsDeleted    bool         `redis:"IsDeleted" json:"IsDeleted"`
	RVersion      string       `redis:"Version" json:"Version"`
}

func NewRedisItem(options txn.ItemOptions) *RedisItem {
	return &RedisItem{
		RKey:          options.Key,
		RValue:        options.Value,
		RGroupKeyList: options.GroupKeyList,
		RTxnState:     options.TxnState,
		RTValid:       options.TValid,
		RTLease:       options.TLease,
		RPrev:         options.Prev,
		RLinkedLen:    options.LinkedLen,
		RIsDeleted:    options.IsDeleted,
		RVersion:      options.Version,
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

func (r *RedisItem) GroupKeyList() string {
	return r.RGroupKeyList
}

func (r *RedisItem) SetGroupKeyList(g string) {
	r.RGroupKeyList = g
}

func (r *RedisItem) TxnState() config.State {
	return r.RTxnState
}

func (r *RedisItem) SetTxnState(s config.State) {
	r.RTxnState = s
}

func (r *RedisItem) TValid() int64 {
	return r.RTValid
}

func (r *RedisItem) SetTValid(t int64) {
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

func (r *RedisItem) Version() string {
	return r.RVersion
}

func (r *RedisItem) SetVersion(v string) {
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
    Version:   %s,
}`, r.RKey, r.RValue, r.RGroupKeyList, util.ToString(r.RTxnState),
		util.ToString(r.RTValid), r.RTLease.Format(time.RFC3339),
		r.RPrev, r.RLinkedLen, r.RIsDeleted, r.RVersion)
}

func (r *RedisItem) Empty() bool {
	return r.RKey == "" && r.RValue == "" &&
		r.RGroupKeyList == "" && r.RTxnState == config.State(0) &&
		r.RTValid == 0 && r.RTLease.IsZero() &&
		r.RPrev == "" && r.RLinkedLen == 0 &&
		!r.RIsDeleted && r.RVersion == ""
}

func (r *RedisItem) Equal(other txn.DataItem) bool {
	return r.Key() == other.Key() &&
		r.Value() == other.Value() &&
		r.GroupKeyList() == other.GroupKeyList() &&
		r.TxnState() == other.TxnState() &&
		r.TValid() == other.TValid() &&
		r.TLease().Equal(other.TLease()) &&
		r.Prev() == other.Prev() &&
		r.LinkedLen() == other.LinkedLen() &&
		r.IsDeleted() == other.IsDeleted() &&
		r.Version() == other.Version()
}

func (r RedisItem) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}
