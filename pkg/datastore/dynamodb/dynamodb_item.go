package dynamodb

import (
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ txn.DataItem = (*DynamoDBItem)(nil)

type DynamoDBItem struct {
	DKey          string       `dynamodbav:"ID" json:"Key"`                    // 分区键
	DValue        string       `dynamodbav:"Value" json:"Value"`               // 值
	DGroupKeyList string       `dynamodbav:"GroupKeyList" json:"GroupKeyList"` // 组键列表
	DTxnState     config.State `dynamodbav:"TxnState" json:"TxnState"`         // 事务状态
	DTValid       int64        `dynamodbav:"TValid" json:"TValid"`             // 有效时间戳
	DTLease       time.Time    `dynamodbav:"TLease" json:"TLease"`             // 租约时间
	DPrev         string       `dynamodbav:"Prev" json:"Prev"`                 // 前驱
	DLinkedLen    int          `dynamodbav:"LinkedLen" json:"LinkedLen"`       // 链接长度
	DIsDeleted    bool         `dynamodbav:"IsDeleted" json:"IsDeleted"`       // 删除标记
	DVersion      string       `dynamodbav:"Version" json:"Version"`           // 版本号
}

func NewDynamoDBItem(options txn.ItemOptions) *DynamoDBItem {
	return &DynamoDBItem{
		DKey:          options.Key,
		DValue:        options.Value,
		DGroupKeyList: options.GroupKeyList,
		DTxnState:     options.TxnState,
		DTValid:       options.TValid,
		DTLease:       options.TLease,
		DPrev:         options.Prev,
		DLinkedLen:    options.LinkedLen,
		DIsDeleted:    options.IsDeleted,
		DVersion:      options.Version,
	}
}

func (d *DynamoDBItem) Key() string {
	return d.DKey
}

func (d *DynamoDBItem) Value() string {
	return d.DValue
}

func (d *DynamoDBItem) SetValue(value string) {
	d.DValue = value
}

func (d *DynamoDBItem) GroupKeyList() string {
	return d.DGroupKeyList
}

func (d *DynamoDBItem) SetGroupKeyList(groupKeyList string) {
	d.DGroupKeyList = groupKeyList
}

func (d *DynamoDBItem) TxnState() config.State {
	return d.DTxnState
}

func (d *DynamoDBItem) SetTxnState(state config.State) {
	d.DTxnState = state
}

func (d *DynamoDBItem) TValid() int64 {
	return d.DTValid
}

func (d *DynamoDBItem) SetTValid(tValid int64) {
	d.DTValid = tValid
}

func (d *DynamoDBItem) TLease() time.Time {
	return d.DTLease
}

func (d *DynamoDBItem) SetTLease(tLease time.Time) {
	d.DTLease = tLease
}

func (d *DynamoDBItem) Prev() string {
	return d.DPrev
}

func (d *DynamoDBItem) SetPrev(prev string) {
	d.DPrev = prev
}

func (d *DynamoDBItem) LinkedLen() int {
	return d.DLinkedLen
}

func (d *DynamoDBItem) SetLinkedLen(linkedLen int) {
	d.DLinkedLen = linkedLen
}

func (d *DynamoDBItem) IsDeleted() bool {
	return d.DIsDeleted
}

func (d *DynamoDBItem) SetIsDeleted(isDeleted bool) {
	d.DIsDeleted = isDeleted
}

func (d *DynamoDBItem) Version() string {
	return d.DVersion
}

func (d *DynamoDBItem) SetVersion(version string) {
	d.DVersion = version
}

func (d *DynamoDBItem) Equal(other txn.DataItem) bool {
	if other == nil {
		return false
	}
	otherItem, ok := other.(*DynamoDBItem)
	if !ok {
		return false
	}

	return d.Key() == otherItem.Key() &&
		d.Value() == otherItem.Value() &&
		d.GroupKeyList() == otherItem.GroupKeyList() &&
		d.TxnState() == otherItem.TxnState() &&
		d.TValid() == otherItem.TValid() &&
		d.TLease().Equal(otherItem.TLease()) &&
		d.Prev() == otherItem.Prev() &&
		d.LinkedLen() == otherItem.LinkedLen() &&
		d.IsDeleted() == otherItem.IsDeleted() &&
		d.Version() == otherItem.Version()
}

func (d *DynamoDBItem) Empty() bool {
	return d == nil || (d.Key() == "" && d.Value() == "")
}

func (d *DynamoDBItem) String() string {
	return fmt.Sprintf(`DynamoDBItem{
    Key:          %s,
    Value:        %s,
    GroupKeyList: %s,
    TxnState:     %s,
    TValid:       %v,
    TLease:       %s,
    Prev:         %s,
    LinkedLen:    %d,
    IsDeleted:    %v,
    Version:      %s,
}`, d.DKey, d.DValue, d.DGroupKeyList, util.ToString(d.DTxnState),
		d.DTValid, d.DTLease.Format(time.RFC3339),
		d.DPrev, d.DLinkedLen, d.DIsDeleted, d.DVersion)
}
