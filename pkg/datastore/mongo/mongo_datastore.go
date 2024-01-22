package mongo

import (
	"github.com/kkkzoz/oreo/pkg/txn"
)

type MongoDatastore struct {
	*txn.Datastore
}

func NewMongoDatastore(name string, conn txn.Connector) txn.Datastorer {
	return txn.NewDatastore(name, conn)
}
