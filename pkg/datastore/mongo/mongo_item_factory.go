package mongo

import "github.com/kkkzoz/oreo/pkg/txn"

type MongoItemFactory struct{}

func (m *MongoItemFactory) NewDataItem(options txn.ItemOptions) txn.DataItem {
	return NewMongoItem(options)
}
