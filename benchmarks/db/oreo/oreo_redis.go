package oreo

import (
	"benchmark/pkg/config"
	"benchmark/ycsb"
	"context"
	"sync"

	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/network"
	"github.com/kkkzoz/oreo/pkg/txn"
)

var _ ycsb.DBCreator = (*OreoRedisCreator)(nil)

type OreoRedisCreator struct {
	mu       sync.Mutex
	ConnList []*redis.RedisConnection
	next     int
	IsRemote bool
}

func (rc *OreoRedisCreator) Create() (ycsb.DB, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	conn := rc.ConnList[rc.next]
	rc.next++
	if rc.next >= len(rc.ConnList) {
		rc.next = 0
	}
	return NewRedisDatastore(conn, rc.IsRemote), nil
}

var _ ycsb.DB = (*RedisDatastore)(nil)

var _ ycsb.TransactionDB = (*RedisDatastore)(nil)

type RedisDatastore struct {
	conn     *redis.RedisConnection
	txn      *txn.Transaction
	isRemote bool
}

func NewRedisDatastore(conn *redis.RedisConnection, isRemote bool) *RedisDatastore {

	return &RedisDatastore{
		conn:     conn,
		isRemote: isRemote,
	}
}

func (r *RedisDatastore) Start() error {
	var txn1 *txn.Transaction
	if r.isRemote {
		client := network.NewClient(config.RemoteAddress)
		txn1 = txn.NewTransactionWithRemote(client)
	} else {
		txn1 = txn.NewTransaction()
	}
	rds := redis.NewRedisDatastore("redis", r.conn)
	txn1.AddDatastore(rds)
	txn1.SetGlobalDatastore(rds)
	r.txn = txn1
	return r.txn.Start()
}

func (r *RedisDatastore) Commit() error {
	return r.txn.Commit()
}

func (r *RedisDatastore) Abort() error {
	return r.txn.Abort()
}

func (r *RedisDatastore) Close() error {
	return nil
}

func (r *RedisDatastore) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return ctx
}

func (r *RedisDatastore) CleanupThread(ctx context.Context) {
}

func (r *RedisDatastore) Read(ctx context.Context, table string, key string) (string, error) {
	keyName := getKeyName(table, key)
	var value string
	err := r.txn.Read("redis", keyName, &value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (r *RedisDatastore) Update(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.txn.Write("redis", keyName, value)
}

func (r *RedisDatastore) Insert(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.txn.Write("redis", keyName, value)
}

func (r *RedisDatastore) Delete(ctx context.Context, table string, key string) error {
	keyName := getKeyName(table, key)
	return r.txn.Delete("redis", keyName)
}
