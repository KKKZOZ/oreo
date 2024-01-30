package oreo

import (
	"context"

	"github.com/kkkzoz/oreo/pkg/datastore/redis"
)

type RedisDatastore struct {
	conn *redis.RedisConnection
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
	return r.conn.Get(keyName)
}

func (r *RedisDatastore) Update(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.conn.Put(keyName, value)
}

func (r *RedisDatastore) Insert(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.conn.Put(keyName, value)
}

func (r *RedisDatastore) Delete(ctx context.Context, table string, key string) error {
	keyName := getKeyName(table, key)
	return r.conn.Delete(keyName)
}

func getKeyName(table string, key string) string {
	return table + "/" + key
}
