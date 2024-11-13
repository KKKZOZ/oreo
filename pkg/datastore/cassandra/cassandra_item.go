package cassandra

import (
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ txn.DataItem = (*CassandraItem)(nil)

type CassandraItem struct {
	// 主键
	CKey string `cql:"key" json:"Key"`
	// 数据值
	CValue string `cql:"value" json:"Value"`
	// 组键列表
	CGroupKeyList string `cql:"group_key_list" json:"GroupKeyList"`
	// 事务状态
	CTxnState config.State `cql:"txn_state" json:"TxnState"`
	// 时间戳
	CTValid int64 `cql:"t_valid" json:"TValid"`
	// 租约时间
	CTLease time.Time `cql:"t_lease" json:"TLease"`
	// 前驱
	CPrev string `cql:"prev" json:"Prev"`
	// 链表长度
	CLinkedLen int `cql:"linked_len" json:"LinkedLen"`
	// 删除标记
	CIsDeleted bool `cql:"is_deleted" json:"IsDeleted"`
	// 版本控制 - Cassandra使用时间戳作为版本
	CVersion string `cql:"version" json:"Version"`
}

// CQL创建表的语句
const CreateTableCQL = `
CREATE TABLE IF NOT EXISTS items (
    key text,
    value text,
    group_key_list text,
    txn_state int,
    t_valid bigint,
    t_lease timestamp,
    prev text,
    linked_len int,
    is_deleted boolean,
    version text,
    PRIMARY KEY (key)
) WITH gc_grace_seconds = 172800`

func NewCassandraItem(options txn.ItemOptions) *CassandraItem {
	return &CassandraItem{
		CKey:          options.Key,
		CValue:        options.Value,
		CGroupKeyList: options.GroupKeyList,
		CTxnState:     options.TxnState,
		CTValid:       options.TValid,
		CTLease:       options.TLease,
		CPrev:         options.Prev,
		CLinkedLen:    options.LinkedLen,
		CIsDeleted:    options.IsDeleted,
		CVersion:      options.Version,
	}
}

// 实现所有必要的接口方法
func (c *CassandraItem) Key() string {
	return c.CKey
}

func (c *CassandraItem) Value() string {
	return c.CValue
}

func (c *CassandraItem) SetValue(value string) {
	c.CValue = value
}

func (c *CassandraItem) GroupKeyList() string {
	return c.CGroupKeyList
}

func (c *CassandraItem) SetGroupKeyList(groupKeyList string) {
	c.CGroupKeyList = groupKeyList
}

func (c *CassandraItem) TxnState() config.State {
	return c.CTxnState
}

func (c *CassandraItem) SetTxnState(state config.State) {
	c.CTxnState = state
}

func (c *CassandraItem) TValid() int64 {
	return c.CTValid
}

func (c *CassandraItem) SetTValid(tValid int64) {
	c.CTValid = tValid
}

func (c *CassandraItem) TLease() time.Time {
	return c.CTLease
}

func (c *CassandraItem) SetTLease(tLease time.Time) {
	c.CTLease = tLease
}

func (c *CassandraItem) Prev() string {
	return c.CPrev
}

func (c *CassandraItem) SetPrev(prev string) {
	c.CPrev = prev
}

func (c *CassandraItem) LinkedLen() int {
	return c.CLinkedLen
}

func (c *CassandraItem) SetLinkedLen(linkedLen int) {
	c.CLinkedLen = linkedLen
}

func (c *CassandraItem) IsDeleted() bool {
	return c.CIsDeleted
}

func (c *CassandraItem) SetIsDeleted(isDeleted bool) {
	c.CIsDeleted = isDeleted
}

func (c *CassandraItem) Version() string {
	return c.CVersion
}

func (c *CassandraItem) SetVersion(version string) {
	c.CVersion = version
}

func (c *CassandraItem) Equal(other txn.DataItem) bool {
	if other == nil {
		return false
	}
	otherItem, ok := other.(*CassandraItem)
	if !ok {
		return false
	}

	return c.Key() == otherItem.Key() &&
		c.Value() == otherItem.Value() &&
		c.GroupKeyList() == otherItem.GroupKeyList() &&
		c.TxnState() == otherItem.TxnState() &&
		c.TValid() == otherItem.TValid() &&
		c.TLease().Equal(otherItem.TLease()) &&
		c.Prev() == otherItem.Prev() &&
		c.LinkedLen() == otherItem.LinkedLen() &&
		c.IsDeleted() == otherItem.IsDeleted() &&
		c.Version() == otherItem.Version()
}

func (c *CassandraItem) Empty() bool {
	return c == nil || (c.Key() == "" && c.Value() == "")
}

func (c *CassandraItem) String() string {
	return fmt.Sprintf(`CassandraItem{
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
}`, c.CKey, c.CValue, c.CGroupKeyList, util.ToString(c.CTxnState),
		c.CTValid, c.CTLease.Format(time.RFC3339),
		c.CPrev, c.CLinkedLen, c.CIsDeleted, c.Version())
}
