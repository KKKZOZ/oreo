package couchdb

import "github.com/oreo-dtx-lab/oreo/pkg/txn"

var _ txn.DataItemFactory = (*CouchDBItemFactory)(nil)

type CouchDBItemFactory struct{}

func (m *CouchDBItemFactory) NewDataItem(options txn.ItemOptions) txn.DataItem {
	return NewCouchDBItem(options)
}
