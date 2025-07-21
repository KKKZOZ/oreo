package mongo

import (
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ txn.DataItemFactory = (*MongoItemFactory)(nil)

type MongoItemFactory struct{}

func (m *MongoItemFactory) NewDataItem(options txn.ItemOptions) txn.DataItem {
	return NewMongoItem(options)
}
