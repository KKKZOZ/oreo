package cassandra

import "github.com/kkkzoz/oreo/pkg/txn"

var _ txn.DataItemFactory = (*CassandraItemFactory)(nil)

type CassandraItemFactory struct{}

func (c *CassandraItemFactory) NewDataItem(options txn.ItemOptions) txn.DataItem {
	return NewCassandraItem(options)
}
