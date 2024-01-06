package integration

import (
	"github.com/kkkzoz/oreo/pkg/datastore/memory"
	"github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/txn"
)

const (
	MEMORY  = "memory"
	REDIS   = "redis"
	MONGO   = "mongo"
	KVROCKS = "kvrocks"
)

func NewRedisConnection() *redis.RedisConnection {
	return redis.NewRedisConnection(&redis.ConnectionOptions{
		Address: "localhost:6379",
	})
}

func NewTransactionWithSetup(dsType string) *txn.Transaction {
	txn := txn.NewTransaction()
	if dsType == "memory" {
		conn := memory.NewMemoryConnection("localhost", 8321)
		mds := memory.NewMemoryDatastore("memory", conn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)
	}
	if dsType == "redis" {
		conn := redis.NewRedisConnection(&redis.ConnectionOptions{
			Address: "localhost:6379",
		})
		rds := redis.NewRedisDatastore("redis", conn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
	}
	if dsType == "mongo" {
		conn := mongo.NewMongoConnection(&mongo.ConnectionOptions{
			Address:        "mongodb://localhost:27017",
			Username:       "",
			Password:       "",
			DBName:         "oreo",
			CollectionName: "records",
		})
		mds := mongo.NewMongoDatastore("mongo", conn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)
	}
	if dsType == "kvrocks" {
		conn := redis.NewRedisConnection(&redis.ConnectionOptions{
			Address: "localhost:6666",
		})
		rds := redis.NewRedisDatastore("kvrocks", conn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
	}

	return txn
}
