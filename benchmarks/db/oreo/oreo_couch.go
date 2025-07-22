package oreo

import (
	"context"
	"sync"

	"benchmark/ycsb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ ycsb.DBCreator = (*OreoCouchCreator)(nil)

type OreoCouchCreator struct {
	mu       sync.Mutex
	ConnList []*couchdb.CouchDBConnection
	next     int
}

func (rc *OreoCouchCreator) Create() (ycsb.DB, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	conn := rc.ConnList[rc.next]
	rc.next++
	if rc.next >= len(rc.ConnList) {
		rc.next = 0
	}
	return NewCouchDatastore(conn), nil
}

var _ ycsb.DB = (*CouchDatastore)(nil)

var _ ycsb.TransactionDB = (*CouchDatastore)(nil)

type CouchDatastore struct {
	conn *couchdb.CouchDBConnection
	txn  *txn.Transaction
}

func NewCouchDatastore(conn *couchdb.CouchDBConnection) *CouchDatastore {
	return &CouchDatastore{
		conn: conn,
	}
}

func (r *CouchDatastore) Start() error {
	txn := txn.NewTransaction()
	rds := couchdb.NewCouchDBDatastore("couch", r.conn)
	txn.AddDatastore(rds)
	txn.SetGlobalDatastore(rds)
	r.txn = txn
	return r.txn.Start()
}

func (r *CouchDatastore) Commit() error {
	return r.txn.Commit()
}

func (r *CouchDatastore) Abort() error {
	return r.txn.Abort()
}

func (r *CouchDatastore) NewTransaction() ycsb.TransactionDB {
	rds := NewCouchDatastore(r.conn)
	return rds
}

func (r *CouchDatastore) Close() error {
	return nil
}

func (r *CouchDatastore) InitThread(
	ctx context.Context,
	threadID int,
	threadCount int,
) context.Context {
	return ctx
}

func (r *CouchDatastore) CleanupThread(ctx context.Context) {
}

func (r *CouchDatastore) Read(ctx context.Context, table string, key string) (string, error) {
	keyName := getKeyName(table, key)
	var value string
	err := r.txn.Read("couch", keyName, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *CouchDatastore) Update(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.txn.Write("couch", keyName, value)
}

func (r *CouchDatastore) Insert(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.txn.Write("couch", keyName, value)
}

func (r *CouchDatastore) Delete(ctx context.Context, table string, key string) error {
	keyName := getKeyName(table, key)
	return r.txn.Delete("couch", keyName)
}
