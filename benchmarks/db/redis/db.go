package redis

import (
	"context"
	"sync"
	"time"

	"benchmark/ycsb"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/redis/go-redis/v9"
)

var _ ycsb.DBCreator = (*RedisCreator)(nil)

type RedisCreator struct {
	mu      sync.Mutex
	RdbList []*redis.Client
	next    int
}

func (rc *RedisCreator) Create() (ycsb.DB, error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rdb := rc.RdbList[rc.next]
	rc.next++
	if rc.next >= len(rc.RdbList) {
		rc.next = 0
	}
	return NewRedis(rdb), nil
}

var _ ycsb.DB = (*Redis)(nil)

type Redis struct {
	Rdb *redis.Client
}

func NewRedis(rdb *redis.Client) *Redis {
	return &Redis{
		Rdb: rdb,
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
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	return r.Rdb.Get(context.Background(), key).Result()
}

func (r *Redis) Update(ctx context.Context, table string, key string, value string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	return r.Rdb.Set(context.Background(), key, value, 0).Err()
}

func (r *Redis) Insert(ctx context.Context, table string, key string, value string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	return r.Rdb.Set(context.Background(), key, value, 0).Err()
}

func (r *Redis) Delete(ctx context.Context, table string, key string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	return r.Rdb.Del(context.Background(), key).Err()
}
