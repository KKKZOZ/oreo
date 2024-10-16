package oreo

import (
	"benchmark/pkg/config"
	"benchmark/ycsb"
	"context"

	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/network"
	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ ycsb.DBCreator = (*OreoRedisCreator)(nil)

type OreoRealisticCreator struct {
	ConnMap             map[string]txn.Connector
	GlobalDatastoreName string
	IsRemote            bool
}

func (oc *OreoRealisticCreator) Create() (ycsb.DB, error) {
	return NewOreoRealisticDatastore(oc.ConnMap, oc.GlobalDatastoreName, oc.IsRemote), nil
}

var _ ycsb.DB = (*OreoRealisticDatastore)(nil)

var _ ycsb.TransactionDB = (*OreoRealisticDatastore)(nil)

type OreoRealisticDatastore struct {
	connMap             map[string]txn.Connector
	globalDatastoreName string
	txn                 *txn.Transaction
	isRemote            bool
}

func NewOreoRealisticDatastore(connMap map[string]txn.Connector, globalDatastoreName string, isRemote bool) *OreoRealisticDatastore {

	return &OreoRealisticDatastore{
		isRemote:            isRemote,
		connMap:             connMap,
		globalDatastoreName: globalDatastoreName,
	}
}

func (r *OreoRealisticDatastore) Start() error {
	var txn1 *txn.Transaction
	if r.isRemote {
		client := network.NewClient(config.RemoteAddressList)
		// oracle := timesource.NewGlobalTimeSource(config.TimeOracleUrl)
		// oracle := timesource.NewLocalTimeSource()
		oracle := timesource.NewSimpleTimeSource()
		txn1 = txn.NewTransactionWithRemote(client, oracle)
	} else {
		txn1 = txn.NewTransaction()
	}

	for dbName, conn := range r.connMap {
		switch dbName {
		case "Redis":
			rds := redis.NewRedisDatastore("Redis", conn)
			txn1.AddDatastore(rds)
			if r.globalDatastoreName == "Redis" {
				txn1.SetGlobalDatastore(rds)
			}
		case "KVRocks":
			rds := redis.NewRedisDatastore("KVRocks", conn)
			txn1.AddDatastore(rds)
			if r.globalDatastoreName == "KVRocks" {
				txn1.SetGlobalDatastore(rds)
			}
		case "MongoDB":
			mds := mongo.NewMongoDatastore("MongoDB", conn)
			txn1.AddDatastore(mds)
			if r.globalDatastoreName == "MongoDB" {
				txn1.SetGlobalDatastore(mds)
			}
		case "CouchDB":
			cds := couchdb.NewCouchDBDatastore("CouchDB", conn)
			txn1.AddDatastore(cds)
			if r.globalDatastoreName == "CouchDB" {
				txn1.SetGlobalDatastore(cds)
			}
		default:
			panic("unknown datastore")
		}
	}

	r.txn = txn1
	return r.txn.Start()
}

func (r *OreoRealisticDatastore) Commit() error {
	return r.txn.Commit()
}

func (r *OreoRealisticDatastore) Abort() error {
	return r.txn.Abort()
}

func (r *OreoRealisticDatastore) NewTransaction() ycsb.TransactionDB {
	panic("implement me")
}

func (r *OreoRealisticDatastore) Close() error {
	return nil
}

func (r *OreoRealisticDatastore) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return ctx
}

func (r *OreoRealisticDatastore) CleanupThread(ctx context.Context) {
}

func (r *OreoRealisticDatastore) Read(ctx context.Context, table string, key string) (string, error) {
	key = addPrefix(key)

	var value string
	err := r.txn.Read(table, key, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *OreoRealisticDatastore) Update(ctx context.Context, table string, key string, value string) error {
	key = addPrefix(key)
	return r.txn.Write(table, key, value)
}

func (r *OreoRealisticDatastore) Insert(ctx context.Context, table string, key string, value string) error {
	key = addPrefix(key)
	return r.txn.Write(table, key, value)
}

func (r *OreoRealisticDatastore) Delete(ctx context.Context, table string, key string) error {
	key = addPrefix(key)
	return r.txn.Delete(table, key)
}

func addPrefix(key string) string {
	return "oreo-" + key
}
