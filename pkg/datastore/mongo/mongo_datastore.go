package mongo

import (
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

type MongoDatastore struct {
	*txn.Datastore
}

func NewMongoDatastore(name string, conn txn.Connector) txn.Datastorer {
	return txn.NewDatastore(name, conn, &MongoItemFactory{})
}
