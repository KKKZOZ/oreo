package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/redis/go-redis/v9"
)

type RedisConnection struct {
	rdb     *redis.Client
	Address string
	se      serializer.Serializer
}

type ConnectionOptions struct {
	Address  string
	Password string
	se       serializer.Serializer
}

func NewRedisConnection(config *ConnectionOptions) *RedisConnection {
	if config == nil {
		config = &ConnectionOptions{
			Address: "localhost:6379",
		}
	}
	if config.Address == "" {
		config.Address = "localhost:6379"
	}

	if config.se == nil {
		config.se = serializer.NewJSONSerializer()
	}

	return &RedisConnection{
		rdb: redis.NewClient(&redis.Options{
			Addr:     config.Address,
			Password: config.Password,
		}),
		Address: config.Address,
		se:      config.se,
	}
}

func (r *RedisConnection) Connect() error {
	return nil
}

// GetItem retrieves the RedisItem for the given key.
func (r *RedisConnection) GetItem(key string) (RedisItem, error) {
	var value RedisItem
	err := r.rdb.HGetAll(context.Background(), key).Scan(&value)
	if err != nil {
		return RedisItem{}, err
	}
	// Check if returned value is an empty struct
	if (RedisItem{}) == value {
		return RedisItem{}, fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

func (r *RedisConnection) PutItem(key string, value RedisItem) error {
	ctx := context.Background()
	_, err := r.rdb.Pipelined(ctx, func(rdb redis.Pipeliner) error {
		rdb.HSet(ctx, key, "Key", value.Key)
		rdb.HSet(ctx, key, "Value", value.Value)
		rdb.HSet(ctx, key, "TxnId", value.TxnId)
		rdb.HSet(ctx, key, "TxnState", value.TxnState)
		rdb.HSet(ctx, key, "TValid", value.TValid.Format(time.RFC3339Nano))
		rdb.HSet(ctx, key, "TLease", value.TLease.Format(time.RFC3339Nano))
		rdb.HSet(ctx, key, "Prev", value.Prev)
		rdb.HSet(ctx, key, "IsDeleted", value.IsDeleted)
		rdb.HSet(ctx, key, "Version", value.Version)
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (r *RedisConnection) ConditionalUpdate(key string, value RedisItem) error {
	ctx := context.Background()
	sha, err := r.rdb.ScriptLoad(ctx, `
	if redis.call('EXISTS', KEYS[1]) == 0 or redis.call('HGET', KEYS[1], 'Version') == ARGV[1] then
	redis.call('HSET', KEYS[1], 'Key', ARGV[2])
	redis.call('HSET', KEYS[1], 'Value', ARGV[3])
	redis.call('HSET', KEYS[1], 'TxnId', ARGV[4])
	redis.call('HSET', KEYS[1], 'TxnState', ARGV[5])
	redis.call('HSET', KEYS[1], 'TValid', ARGV[6])
	redis.call('HSET', KEYS[1], 'TLease', ARGV[7])
	redis.call('HSET', KEYS[1], 'Version', ARGV[8])
	redis.call('HSET', KEYS[1], 'Prev', ARGV[9])
	return redis.call('HGETALL', KEYS[1])
else
	return redis.error_reply('version mismatch')
end
    `).Result()
	if err != nil {
		return err
	}

	_, err = r.rdb.EvalSha(ctx, sha, []string{value.Key}, value.Version, value.Key,
		value.Value, value.TxnId, value.TxnState, value.TValid, value.TLease,
		value.Version+1, value.Prev).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisConnection) Get(name string) (string, error) {
	return r.rdb.Get(context.Background(), name).Result()
}

func (r *RedisConnection) Put(name string, value any) error {
	return r.rdb.Set(context.Background(), name, value, 0).Err()
}

func (r *RedisConnection) Delete(name string) error {
	return r.rdb.Del(context.Background(), name).Err()
}
