package integration

import (
	"github.com/kkkzoz/oreo/pkg/datastore/memory"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/txn"
)

const (
	MEMORY = "memory"
	REDIS  = "redis"
)

func NewRedisConnection() *redis.RedisConnection {
	return redis.NewRedisConnection(&redis.ConnectionOptions{
		Address: "localhost:6379",
	})
}

func NewTransactionWithSetup(tp string) *txn.Transaction {
	txn := txn.NewTransaction()
	if tp == "memory" {
		conn := memory.NewMemoryConnection("localhost", 8321)
		mds := memory.NewMemoryDatastore("memory", conn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)
	}
	if tp == "redis" {
		conn := redis.NewRedisConnection(&redis.ConnectionOptions{
			Address: "localhost:6379",
		})
		rds := redis.NewRedisDatastore("redis", conn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
	}

	return txn
}
