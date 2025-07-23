package tikv

import "github.com/kkkzoz/oreo/pkg/txn"

type DynamoDBDatastore struct {
	*txn.Datastore
}

func NewTiKVDatastore(name string, conn txn.Connector) txn.Datastorer {
	return txn.NewDatastore(name, conn, &TiKVItemFactory{})
}
