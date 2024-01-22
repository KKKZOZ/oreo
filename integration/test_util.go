package integration

import (
	"github.com/kkkzoz/oreo/internal/mock"
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

func NewConnectionWithSetup(dsType string) txn.Connector {

	var conn txn.Connector

	if dsType == "redis" {
		conn = redis.NewRedisConnection(&redis.ConnectionOptions{
			Address: "localhost:6379",
		})
	}
	if dsType == "mongo" {
		conn = mongo.NewMongoConnection(&mongo.ConnectionOptions{
			Address:        "mongodb://localhost:27017",
			Username:       "",
			Password:       "",
			DBName:         "oreo",
			CollectionName: "records",
		})
	}

	if dsType == "kvrocks" {
		conn = redis.NewRedisConnection(&redis.ConnectionOptions{
			Address: "localhost:6666",
		})
	}
	conn.Connect()
	return conn
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

func NewTransactionWithMockConn(dsType string, limit int,
	isReturned bool, debugFunc func() error) *txn.Transaction {

	txn := txn.NewTransaction()
	if dsType == "redis" {

		mockConn := mock.NewMockRedisConnection(
			"localhost", 6379, limit, isReturned, debugFunc)
		rds := redis.NewRedisDatastore("redis", mockConn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
	}

	if dsType == "mongo" {
		mockConn := mock.NewMockMongoConnection(
			"localhost", 27017, limit, isReturned, debugFunc)
		mds := mongo.NewMongoDatastore("mongo", mockConn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)
	}

	if dsType == "kvrocks" {
		mockConn := mock.NewMockRedisConnection(
			"localhost", 6666, limit, isReturned, debugFunc)
		rds := redis.NewRedisDatastore("kvrocks", mockConn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
	}

	return txn
}
