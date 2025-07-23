package cassandra

import "github.com/kkkzoz/oreo/pkg/txn"

type CassandraDatastore struct {
	*txn.Datastore
}

func NewCassandraDatastore(name string, conn txn.Connector) txn.Datastorer {
	return txn.NewDatastore(name, conn, &CassandraItemFactory{})
}
