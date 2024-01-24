package couchdb

import (
	"github.com/kkkzoz/oreo/pkg/txn"
)

type CouchDBDatastore struct {
	*txn.Datastore
}

func NewCouchDBDatastore(name string, conn txn.Connector) txn.Datastorer {
	return txn.NewDatastore(name, conn, &CouchDBItemFactory{})
}
