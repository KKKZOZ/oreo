package oreo

import (
	"context"

	"benchmark/pkg/benconfig"
	"benchmark/ycsb"
	"github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/timesource"
	"github.com/kkkzoz/oreo/pkg/txn"
)

var _ ycsb.DBCreator = (*OreoRedisCreator)(nil)

type OreoCreator struct {
	ConnMap             map[string]txn.Connector
	GlobalDatastoreName string
	IsRemote            bool
}

func (oc *OreoCreator) Create() (ycsb.DB, error) {
	return NewOreoDatastore(oc.ConnMap, oc.GlobalDatastoreName, oc.IsRemote), nil
}

var _ ycsb.DB = (*OreoDatastore)(nil)

var _ ycsb.TransactionDB = (*OreoDatastore)(nil)

type OreoDatastore struct {
	connMap             map[string]txn.Connector
	globalDatastoreName string
	txn                 *txn.Transaction
	isRemote            bool
}

func NewOreoDatastore(
	connMap map[string]txn.Connector,
	globalDatastoreName string,
	isRemote bool,
) *OreoDatastore {
	return &OreoDatastore{
		isRemote:            isRemote,
		connMap:             connMap,
		globalDatastoreName: globalDatastoreName,
	}
}

func (r *OreoDatastore) Start() error {
	var txn1 *txn.Transaction
	if r.isRemote {
		oracle := timesource.NewGlobalTimeSource(benconfig.TimeOracleUrl)
		txn1 = txn.NewTransactionWithRemote(benconfig.GlobalClient, oracle)
	} else {
		txn1 = txn.NewTransaction()
	}

	for dbName, conn := range r.connMap {
		switch dbName {
		case "redis1":
			rds := redis.NewRedisDatastore("redis1", conn)
			txn1.AddDatastore(rds)
			if r.globalDatastoreName == "redis1" {
				txn1.SetGlobalDatastore(rds)
			}
		case "mongo1":
			mds := mongo.NewMongoDatastore("mongo1", conn)
			txn1.AddDatastore(mds)
			if r.globalDatastoreName == "mongo1" {
				txn1.SetGlobalDatastore(mds)
			}
		case "mongo2":
			mds := mongo.NewMongoDatastore("mongo2", conn)
			txn1.AddDatastore(mds)
			if r.globalDatastoreName == "mongo2" {
				txn1.SetGlobalDatastore(mds)
			}
		default:
			panic("unknown datastore")
		}
	}

	r.txn = txn1
	return r.txn.Start()
}

func (r *OreoDatastore) Commit() error {
	return r.txn.Commit()
}

func (r *OreoDatastore) Abort() error {
	return r.txn.Abort()
}

func (r *OreoDatastore) NewTransaction() ycsb.TransactionDB {
	panic("implement me")
}

func (r *OreoDatastore) Close() error {
	return nil
}

func (r *OreoDatastore) InitThread(
	ctx context.Context,
	threadID int,
	threadCount int,
) context.Context {
	return ctx
}

func (r *OreoDatastore) CleanupThread(ctx context.Context) {
}

func (r *OreoDatastore) Read(ctx context.Context, table string, key string) (string, error) {
	keyName := getKeyName("", key)
	// fmt.Printf("Read key: %s Table: %s\n", keyName, table)

	var value string
	err := r.txn.Read(table, keyName, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *OreoDatastore) Update(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName("", key)
	return r.txn.Write(table, keyName, value)
}

func (r *OreoDatastore) Insert(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName("", key)
	return r.txn.Write(table, keyName, value)
}

func (r *OreoDatastore) Delete(ctx context.Context, table string, key string) error {
	keyName := getKeyName("", key)
	return r.txn.Delete(table, keyName)
}
