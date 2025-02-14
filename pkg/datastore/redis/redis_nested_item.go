package redis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/errs"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ txn.NestedItem = (*NestedRedisItem)(nil)

type NestedRedisItem struct {
	RKey          string       `redis:"Key" json:"Key"`
	RValue        []byte       `redis:"Value" json:"Value"`
	RGroupKeyList string       `redis:"GroupKeyList" json:"GroupKeyList"`
	RTxnState     config.State `redis:"TxnState" json:"TxnState"`
	RTValid       int64        `redis:"TValid" json:"TValid"`
	RTLease       time.Time    `redis:"TLease" json:"TLease"`
	RPrev         []byte       `redis:"Prev" json:"Prev"`
	RLinkedLen    int          `redis:"LinkedLen" json:"LinkedLen"`
	RIsDeleted    bool         `redis:"IsDeleted" json:"IsDeleted"`
	RVersion      string       `redis:"Version" json:"Version"`
}

func NewNestedRedisItem(options txn.ItemOptionsNext) *NestedRedisItem {
	value, _ := config.Config.Serializer.Serialize(options.Value)
	return &NestedRedisItem{
		RKey:          options.Key,
		RValue:        value,
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

func (r *NestedRedisItem) Key() string {
	return r.RKey
}

func (r *NestedRedisItem) Value() []byte {
	return r.RValue
}

func (ni *NestedRedisItem) ParseValue(valuePtr any) error {
	err := config.Config.Serializer.Deserialize([]byte(ni.RValue), valuePtr)
	if err != nil {
		return errs.NewSerializerError(errs.DeserializeFailed, err)
	}
	return nil
}

func (r *NestedRedisItem) SetValue(v []byte) error {
	value, err := config.Config.Serializer.Serialize(v)
	if err != nil {
		return errs.NewSerializerError(errs.SerializeFailed, err)
	}
	r.RValue = value
	return nil
}

func (r *NestedRedisItem) GroupKeyList() string {
	return r.RGroupKeyList
}

func (r *NestedRedisItem) SetGroupKeyList(g string) {
	r.RGroupKeyList = g
}

func (r *NestedRedisItem) TxnState() config.State {
	return r.RTxnState
}

func (r *NestedRedisItem) SetTxnState(s config.State) {
	r.RTxnState = s
}

func (r *NestedRedisItem) TValid() int64 {
	return r.RTValid
}

func (r *NestedRedisItem) SetTValid(t int64) {
	r.RTValid = t
}

func (r *NestedRedisItem) TLease() time.Time {
	return r.RTLease
}

func (r *NestedRedisItem) SetTLease(t time.Time) {
	r.RTLease = t
}

func (r *NestedRedisItem) Prev() []byte {
	return r.RPrev
}

func (r *NestedRedisItem) SetPrev(p []byte) {
	r.RPrev = p
}

func (r *NestedRedisItem) LinkedLen() int {
	return r.RLinkedLen
}

func (r *NestedRedisItem) SetLinkedLen(l int) {
	r.RLinkedLen = l
}

func (r *NestedRedisItem) IsDeleted() bool {
	return r.RIsDeleted
}

func (r *NestedRedisItem) SetIsDeleted(d bool) {
	r.RIsDeleted = d
}

func (r *NestedRedisItem) Version() string {
	return r.RVersion
}

func (r *NestedRedisItem) SetVersion(v string) {
	r.RVersion = v
}

func (r NestedRedisItem) String() string {
	return fmt.Sprintf(`NestedRedisItem{
    Key:       %s,
    Value:     %s,
    GroupKeyList:     %s,
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

func (r *NestedRedisItem) Empty() bool {
	return r.RKey == "" && string(r.RValue) == "" &&
		r.RGroupKeyList == "" && r.RTxnState == config.State(0) &&
		r.RTValid == 0 && r.RTLease.IsZero() &&
		string(r.RPrev) == "" && r.RLinkedLen == 0 &&
		!r.RIsDeleted && r.RVersion == ""
}

func (r *NestedRedisItem) Equal(other txn.DataItemNext) bool {
	return r.Key() == other.Key() &&
		// r.Value() == other.Value() &&
		r.GroupKeyList() == other.GroupKeyList() &&
		r.TxnState() == other.TxnState()
	// r.TValid() == other.TValid() &&
	// r.TLease().Equal(other.TLease()) &&
	// r.Prev() == other.Prev() &&
	// r.LinkedLen() == other.LinkedLen() &&
	// r.IsDeleted() == other.IsDeleted() &&
	// r.Version() == other.Version()
}

func (r NestedRedisItem) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

func (ni *NestedRedisItem) GetPrevItem() (txn.
	NestedItem, error) {
	if string(ni.Prev()) == "" {
		return nil, errs.NewKeyNotFoundError(ni.Key(), errs.NotFoundInAVC)
	}
	var prevItem NestedRedisItem
	prevBytes := []byte(ni.Prev())
	err := config.Config.Serializer.Deserialize(prevBytes, &prevItem)
	if err != nil {
		return nil, errs.NewSerializerError(errs.DeserializeFailed, err)
	}
	return &prevItem, nil
}
