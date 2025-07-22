package redis

import (
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ txn.DataItemFactory = (*RedisItemFactory)(nil)

type RedisItemFactory struct{}

func (r *RedisItemFactory) NewDataItem(options txn.ItemOptions) txn.DataItem {
	if options.Value == nil {
		options.Value = ""
	}

	return &RedisItem{
		RKey:          options.Key,
		RValue:        options.Value.(string),
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
