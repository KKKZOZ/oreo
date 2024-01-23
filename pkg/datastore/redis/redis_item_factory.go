package redis

import "github.com/kkkzoz/oreo/pkg/txn"

type RedisItemFactory struct{}

func (r *RedisItemFactory) NewDataItem(options txn.ItemOptions) txn.DataItem {
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
