package redis

import (
	"github.com/kkkzoz/oreo/pkg/txn"
)

// RedisDatastore represents a datastore implementation using Redis as the underlying storage.
type RedisDatastore struct {
	*txn.Datastore
}

// NewRedisDatastore creates a new instance of RedisDatastore with the given name and Redis connection.
func NewRedisDatastore(name string, conn txn.Connector) txn.Datastorer {
	return txn.NewDatastore(name, conn)
}
