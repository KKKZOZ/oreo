package oreo

import (
	"benchmark/ycsb"
	"context"

	"github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/txn"
)

var _ ycsb.DBCreator = (*OreoRedisCreator)(nil)

type OreoCreator struct {
	ConnMap             map[string]txn.Connector
	GlobalDatastoreName string
}

func (oc *OreoCreator) Create() (ycsb.DB, error) {
	return NewOreoDatastore(oc.ConnMap, oc.GlobalDatastoreName), nil
}

var _ ycsb.DB = (*OreoDatastore)(nil)

var _ ycsb.TransactionDB = (*OreoDatastore)(nil)

type OreoDatastore struct {
	connMap             map[string]txn.Connector
	globalDatastoreName string
	txn                 *txn.Transaction
}

func NewOreoDatastore(connMap map[string]txn.Connector, globalDatastoreName string) *OreoDatastore {

	return &OreoDatastore{
		connMap:             connMap,
		globalDatastoreName: globalDatastoreName,
	}
}

func (r *OreoDatastore) Start() error {
	txn := txn.NewTransaction()

	for dbName, conn := range r.connMap {
		switch dbName {
		case "redis":
			rds := redis.NewRedisDatastore("redis", conn)
			txn.AddDatastore(rds)
			if r.globalDatastoreName == "redis" {
				txn.SetGlobalDatastore(rds)
			}
		case "mongo":
			mds := mongo.NewMongoDatastore("mongo", conn)
			txn.AddDatastore(mds)
			if r.globalDatastoreName == "mongo" {
				txn.SetGlobalDatastore(mds)
			}
		default:
			panic("unknown datastore")
		}
	}

	r.txn = txn
	return r.txn.Start()
}

func (r *OreoDatastore) Commit() error {
	return r.txn.Commit()
}

func (r *OreoDatastore) Abort() error {
	return r.txn.Abort()
}

func (r *OreoDatastore) Close() error {
	return nil
}

func (r *OreoDatastore) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return ctx
}

func (r *OreoDatastore) CleanupThread(ctx context.Context) {
}

func (r *OreoDatastore) Read(ctx context.Context, table string, key string) (string, error) {
	keyName := getKeyName(table, key)
	var value string
	err := r.txn.Read(table, keyName, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *OreoDatastore) Update(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.txn.Write(table, keyName, value)
}

func (r *OreoDatastore) Insert(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.txn.Write(table, keyName, value)
}

func (r *OreoDatastore) Delete(ctx context.Context, table string, key string) error {
	keyName := getKeyName(table, key)
	return r.txn.Delete(table, keyName)
}
