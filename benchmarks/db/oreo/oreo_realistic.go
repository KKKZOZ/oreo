package oreo

import (
	"context"
	"log"

	"benchmark/pkg/benconfig"
	"benchmark/ycsb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/cassandra"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/dynamodb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/tikv"
	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ ycsb.DBCreator = (*OreoRedisCreator)(nil)

type OreoRealisticCreator struct {
	ConnMap             map[string]txn.Connector
	GlobalDatastoreName string
	IsRemote            bool
	Mode                string
}

func (oc *OreoRealisticCreator) Create() (ycsb.DB, error) {
	return NewOreoRealisticDatastore(oc.ConnMap, oc.GlobalDatastoreName, oc.IsRemote, oc.Mode), nil
}

var _ ycsb.DB = (*OreoRealisticDatastore)(nil)

var _ ycsb.TransactionDB = (*OreoRealisticDatastore)(nil)

type OreoRealisticDatastore struct {
	connMap             map[string]txn.Connector
	globalDatastoreName string
	txn                 *txn.Transaction
	isRemote            bool
	mode                string
}

func NewOreoRealisticDatastore(
	connMap map[string]txn.Connector,
	globalDatastoreName string,
	isRemote bool,
	mode string,
) *OreoRealisticDatastore {
	return &OreoRealisticDatastore{
		isRemote:            isRemote,
		connMap:             connMap,
		globalDatastoreName: globalDatastoreName,
		mode:                mode,
	}
}

func (r *OreoRealisticDatastore) Start() error {
	var txn1 *txn.Transaction
	oracle := timesource.NewGlobalTimeSource(benconfig.TimeOracleUrl)
	// oracle := timesource.NewLocalTimeSource()
	// oracle := timesource.NewSimpleTimeSource()
	if r.isRemote {
		txn1 = txn.NewTransactionWithRemote(benconfig.GlobalClient, oracle)
	} else {
		txn1 = txn.NewTransactionWithOracle(oracle)
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
		case "Cassandra":
			cds := cassandra.NewCassandraDatastore("Cassandra", conn)
			txn1.AddDatastore(cds)
			if r.globalDatastoreName == "Cassandra" {
				txn1.SetGlobalDatastore(cds)
			}
		case "DynamoDB":
			dds := dynamodb.NewDynamoDBDatastore("DynamoDB", conn)
			txn1.AddDatastore(dds)
			if r.globalDatastoreName == "DynamoDB" {
				txn1.SetGlobalDatastore(dds)
			}
		case "TiKV":
			tds := tikv.NewTiKVDatastore("TiKV", conn)
			txn1.AddDatastore(tds)
			if r.globalDatastoreName == "TiKV" {
				txn1.SetGlobalDatastore(tds)
			}
		default:
			log.Panicf("unknown datastore: %s", dbName)
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

func (r *OreoRealisticDatastore) InitThread(
	ctx context.Context,
	threadID int,
	threadCount int,
) context.Context {
	return ctx
}

func (r *OreoRealisticDatastore) CleanupThread(ctx context.Context) {
}

func (r *OreoRealisticDatastore) Read(
	ctx context.Context,
	table string,
	key string,
) (string, error) {
	key = r.addPrefix(key)

	var value string
	err := r.txn.Read(table, key, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *OreoRealisticDatastore) Update(
	ctx context.Context,
	table string,
	key string,
	value string,
) error {
	key = r.addPrefix(key)
	return r.txn.Write(table, key, value)
}

func (r *OreoRealisticDatastore) Insert(
	ctx context.Context,
	table string,
	key string,
	value string,
) error {
	key = r.addPrefix(key)
	return r.txn.Write(table, key, value)
}

func (r *OreoRealisticDatastore) Delete(ctx context.Context, table string, key string) error {
	key = r.addPrefix(key)
	return r.txn.Delete(table, key)
}

func (r *OreoRealisticDatastore) addPrefix(key string) string {
	prefix := ""
	switch r.mode {
	case "native":
		prefix = "native-"
	case "cg":
		prefix = "cg-"
	case "oreo":
		prefix = "oreo-"
	default:
		panic("unknown mode in OreoRealisticDatastore")
	}
	return prefix + key
}
