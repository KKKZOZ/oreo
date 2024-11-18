package oreo

import (
	"benchmark/pkg/config"
	"benchmark/ycsb"
	"context"

	"github.com/oreo-dtx-lab/oreo/pkg/datastore/cassandra"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/dynamodb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/network"
	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ ycsb.DBCreator = (*OreoRedisCreator)(nil)

type OreoYCSBCreator struct {
	ConnMap             map[string]txn.Connector
	GlobalDatastoreName string
	Mode                string
}

func (oc *OreoYCSBCreator) Create() (ycsb.DB, error) {
	return NewOreoYCSBDatastore(oc.ConnMap, oc.GlobalDatastoreName, oc.Mode), nil
}

var _ ycsb.DB = (*OreoYCSBDatastore)(nil)

var _ ycsb.TransactionDB = (*OreoYCSBDatastore)(nil)

type OreoYCSBDatastore struct {
	connMap             map[string]txn.Connector
	globalDatastoreName string
	txn                 *txn.Transaction
	mode                string
}

func NewOreoYCSBDatastore(connMap map[string]txn.Connector, globalDatastoreName string, mode string) *OreoYCSBDatastore {

	return &OreoYCSBDatastore{
		connMap:             connMap,
		globalDatastoreName: globalDatastoreName,
		mode:                mode,
	}
}

func (r *OreoYCSBDatastore) Start() error {
	var txn1 *txn.Transaction
	oracle := timesource.NewGlobalTimeSource(config.TimeOracleUrl)
	// oracle := timesource.NewLocalTimeSource()
	// oracle := timesource.NewSimpleTimeSource()
	if r.mode == "oreo" {
		client := network.NewClient(config.RemoteAddressList)
		txn1 = txn.NewTransactionWithRemote(client, oracle)
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
		default:
			panic("unknown datastore")
		}
	}

	r.txn = txn1
	return r.txn.Start()
}

func (r *OreoYCSBDatastore) Commit() error {
	return r.txn.Commit()
}

func (r *OreoYCSBDatastore) Abort() error {
	return r.txn.Abort()
}

func (r *OreoYCSBDatastore) NewTransaction() ycsb.TransactionDB {
	panic("implement me")
}

func (r *OreoYCSBDatastore) Close() error {
	return nil
}

func (r *OreoYCSBDatastore) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return ctx
}

func (r *OreoYCSBDatastore) CleanupThread(ctx context.Context) {
}

func (r *OreoYCSBDatastore) Read(ctx context.Context, table string, key string) (string, error) {
	key = r.addPrefix(key)

	var value string
	err := r.txn.Read(table, key, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *OreoYCSBDatastore) Update(ctx context.Context, table string, key string, value string) error {
	key = r.addPrefix(key)
	return r.txn.Write(table, key, value)
}

func (r *OreoYCSBDatastore) Insert(ctx context.Context, table string, key string, value string) error {
	key = r.addPrefix(key)
	return r.txn.Write(table, key, value)
}

func (r *OreoYCSBDatastore) Delete(ctx context.Context, table string, key string) error {
	key = r.addPrefix(key)
	return r.txn.Delete(table, key)
}

func (r *OreoYCSBDatastore) addPrefix(key string) string {
	prefix := ""
	switch r.mode {
	case "native":
		prefix = "native-"
	case "cg":
		prefix = "cg-"
	case "oreo":
		prefix = "oreo-"
	default:
		msg := "unknown mode in OreoYCSBDatastore: " + r.mode
		panic(msg)
	}
	return prefix + key
}
