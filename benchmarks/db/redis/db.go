package redis

import (
	"benchmark/ycsb"
	"context"

	"github.com/kkkzoz/oreo/pkg/datastore/redis"
)

var _ ycsb.DB = (*Redis)(nil)

type Redis struct {
	conn *redis.RedisConnection
}

func NewRedis(conn *redis.RedisConnection) *Redis {
	return &Redis{
		conn: conn,
	}
}

func (r *Redis) Close() error {
	return nil
}

func (r *Redis) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return ctx
}

func (r *Redis) CleanupThread(ctx context.Context) {
}

func (r *Redis) Read(ctx context.Context, table string, key string) (string, error) {
	keyName := getKeyName(table, key)
	return r.conn.Get(keyName)
}

func (r *Redis) Update(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.conn.Put(keyName, value)
}

func (r *Redis) Insert(ctx context.Context, table string, key string, value string) error {
	keyName := getKeyName(table, key)
	return r.conn.Put(keyName, value)
}

func (r *Redis) Delete(ctx context.Context, table string, key string) error {
	keyName := getKeyName(table, key)
	return r.conn.Delete(keyName)
}

func getKeyName(table string, key string) string {
	return table + "/" + key
}
