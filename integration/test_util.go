package integration

import (
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/mock"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

const (
	MEMORY  = "memory"
	REDIS   = "redis"
	MONGO   = "mongo"
	KVROCKS = "kvrocks"
	COUCHDB = "couchdb"
)

func NewConnectionWithSetup(dsType string) txn.Connector {
	var conn txn.Connector

	if dsType == "redis" {
		conn = redis.NewRedisConnection(&redis.ConnectionOptions{
			Address:  "localhost:6379",
			Password: "password",
		})
	}
	if dsType == "mongo" {
		conn = mongo.NewMongoConnection(&mongo.ConnectionOptions{
			Address:        "mongodb://localhost:27017",
			Username:       "admin",
			Password:       "admin",
			DBName:         "oreo",
			CollectionName: "records",
		})
	}
	if dsType == "couchdb" {
		conn = couchdb.NewCouchDBConnection(&couchdb.ConnectionOptions{
			Address: "http://admin:password@localhost:5984",
			DBName:  "oreo",
		})
	}

	if dsType == "kvrocks" {
		conn = redis.NewRedisConnection(&redis.ConnectionOptions{
			Address:  "localhost:6666",
			Password: "password",
		})
	}
	err := conn.Connect()
	logger.CheckAndLogError("Failed to connect to datastore", err)
	return conn
}

func NewTransactionWithSetup(dsType string) *txn.Transaction {
	txn := txn.NewTransaction()
	if dsType == "redis" {
		conn := redis.NewRedisConnection(&redis.ConnectionOptions{
			Address:  "localhost:6379",
			Password: "password",
		})
		rds := redis.NewRedisDatastore("redis", conn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
	}
	if dsType == "mongo" {
		conn := mongo.NewMongoConnection(&mongo.ConnectionOptions{
			Address:        "mongodb://localhost:27017",
			Username:       "admin",
			Password:       "admin",
			DBName:         "oreo",
			CollectionName: "records",
		})
		mds := mongo.NewMongoDatastore("mongo", conn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)
	}
	if dsType == "couchdb" {
		conn := couchdb.NewCouchDBConnection(&couchdb.ConnectionOptions{
			Address: "http://admin:password@localhost:5984",
			DBName:  "oreo",
		})
		cds := couchdb.NewCouchDBDatastore("couchdb", conn)
		txn.AddDatastore(cds)
		txn.SetGlobalDatastore(cds)
	}
	if dsType == "kvrocks" {
		conn := redis.NewRedisConnection(&redis.ConnectionOptions{
			Address:  "localhost:6666",
			Password: "password",
		})
		rds := redis.NewRedisDatastore("kvrocks", conn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
	}

	return txn
}

func NewTransactionWithMockConn(dsType string, limit int,
	isReturned bool, networkDelay time.Duration, debugFunc func() error,
) *txn.Transaction {
	txn := txn.NewTransaction()
	if dsType == "redis" {
		mockConn := mock.NewMockRedisConnection(
			"localhost", 6379, limit, isReturned, networkDelay, debugFunc)
		rds := redis.NewRedisDatastore("redis", mockConn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
	}

	if dsType == "mongo" {
		mockConn := mock.NewMockMongoConnection(
			"localhost", 27017, "admin", "admin", limit, isReturned, networkDelay, debugFunc)
		mds := mongo.NewMongoDatastore("mongo", mockConn)
		txn.AddDatastore(mds)
		txn.SetGlobalDatastore(mds)
	}

	if dsType == "couchdb" {
		mockConn := mock.NewMockCouchDBConnection(
			"localhost", 5984, limit, isReturned, networkDelay, debugFunc)
		cds := couchdb.NewCouchDBDatastore("couchdb", mockConn)
		txn.AddDatastore(cds)
		txn.SetGlobalDatastore(cds)
	}

	if dsType == "kvrocks" {
		mockConn := mock.NewMockRedisConnection(
			"localhost", 6666, limit, isReturned, networkDelay, debugFunc)
		rds := redis.NewRedisDatastore("kvrocks", mockConn)
		txn.AddDatastore(rds)
		txn.SetGlobalDatastore(rds)
	}

	return txn
}
