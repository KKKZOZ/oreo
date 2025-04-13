package oreo

import (
	"benchmark/pkg/benconfig"
	"benchmark/ycsb"
	"context"
	"sync"

	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ ycsb.DBCreator = (*OreoMongoCreator)(nil)

type OreoMongoCreator struct {
	mu       sync.Mutex
	ConnList []*mongo.MongoConnection
	next     int
	IsRemote bool
}

func (rc *OreoMongoCreator) Create() (ycsb.DB, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	conn := rc.ConnList[rc.next]
	rc.next++
	if rc.next >= len(rc.ConnList) {
		rc.next = 0
	}
	return NewMongoDatastore(conn, rc.IsRemote), nil
}

var _ ycsb.DB = (*MongoDatastore)(nil)

var _ ycsb.TransactionDB = (*MongoDatastore)(nil)

type MongoDatastore struct {
	conn     *mongo.MongoConnection
	txn      *txn.Transaction
	isRemote bool
}

func NewMongoDatastore(conn *mongo.MongoConnection, isRemote bool) *MongoDatastore {

	return &MongoDatastore{
		isRemote: isRemote,
		conn:     conn,
	}
}

func (r *MongoDatastore) Start() error {
	var txn1 *txn.Transaction
	if r.isRemote {
		oracle := timesource.NewGlobalTimeSource(benconfig.TimeOracleUrl)
		txn1 = txn.NewTransactionWithRemote(benconfig.Client, oracle)
	} else {
		txn1 = txn.NewTransaction()
	}
	rds := mongo.NewMongoDatastore("mongo", r.conn)
	txn1.AddDatastore(rds)
	txn1.SetGlobalDatastore(rds)
	r.txn = txn1
	return r.txn.Start()
}

func (r *MongoDatastore) Commit() error {
	return r.txn.Commit()
}

func (r *MongoDatastore) Abort() error {
	return r.txn.Abort()
}

func (r *MongoDatastore) NewTransaction() ycsb.TransactionDB {
	rds := NewMongoDatastore(r.conn, r.isRemote)
	return rds
}

func (r *MongoDatastore) Close() error {
	return nil
}

func (r *MongoDatastore) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return ctx
}

func (r *MongoDatastore) CleanupThread(ctx context.Context) {
}

func (r *MongoDatastore) Read(ctx context.Context, table string, key string) (string, error) {
	keyName := getKeyName(table, key)
	var value string
	err := r.txn.Read("mongo", keyName, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *MongoDatastore) Update(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.txn.Write("mongo", keyName, value)
}

func (r *MongoDatastore) Insert(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.txn.Write("mongo", keyName, value)
}

func (r *MongoDatastore) Delete(ctx context.Context, table string, key string) error {
	keyName := getKeyName(table, key)
	return r.txn.Delete("mongo", keyName)
}
