package redis

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/redis/go-redis/v9"
)

// RedisConnection implements the txn.Connector interface.
var _ txn.Connector = (*RedisConnection)(nil)

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

// NewRedisConnection creates a new Redis connection using the provided configuration options.
// If the config parameter is nil, default values will be used.
// The Redis connection is established using the specified address and password.
// The address format should be in the form "host:port".
// The se parameter is used for data serialization and deserialization.
// If se is nil, a default JSON serializer will be used.
// Returns a pointer to the created RedisConnection.
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

// Connect establishes a connection to the Redis server.
// It returns an error if the connection cannot be established.
func (r *RedisConnection) Connect() error {
	return nil
}

// GetItem retrieves a txn.DataItem from the Redis database based on the specified key.
// If the key is not found, it returns an empty txn.DataItem and an error.
func (r *RedisConnection) GetItem(key string) (txn.DataItem, error) {
	var value RedisItem
	err := r.rdb.HGetAll(context.Background(), key).Scan(&value)
	if err != nil {
		return &RedisItem{}, err
	}
	// Check if returned value is an empty struct
	if value.Empty() {
		return &RedisItem{}, errors.New(txn.KeyNotFound)
	}
	return &value, nil
}

// PutItem puts an item into the Redis database with the specified key and value.
// It sets various fields of the txn.DataItem struct as hash fields in the Redis hash.
// The function returns an error if there was a problem executing the Redis commands.
func (r *RedisConnection) PutItem(key string, value txn.DataItem) (string, error) {
	ctx := context.Background()
	_, err := r.rdb.Pipelined(ctx, func(rdb redis.Pipeliner) error {
		rdb.HSet(ctx, key, "Key", value.Key())
		rdb.HSet(ctx, key, "Value", value.Value())
		rdb.HSet(ctx, key, "TxnId", value.TxnId())
		rdb.HSet(ctx, key, "TxnState", value.TxnState())
		rdb.HSet(ctx, key, "TValid", value.TValid().Format(time.RFC3339Nano))
		rdb.HSet(ctx, key, "TLease", value.TLease().Format(time.RFC3339Nano))
		rdb.HSet(ctx, key, "Prev", value.Prev())
		rdb.HSet(ctx, key, "LinkedLen", value.LinkedLen())
		rdb.HSet(ctx, key, "IsDeleted", value.IsDeleted())
		rdb.HSet(ctx, key, "Version", value.Version())
		return nil
	})

	if err != nil {
		return "", err
	}
	return "", nil
}

// ConditionalUpdate updates the value of a Redis item if the version matches the provided value.
// It takes a key string and a txn.DataItem value as parameters.
// If the item's version does not match, it returns a version mismatch error.
// Otherwise, it updates the item with the provided values and returns the updated item.
func (r *RedisConnection) ConditionalUpdate(key string, value txn.DataItem, doCreate bool) (string, error) {

	if doCreate {
		ctx := context.Background()
		sha, err := r.rdb.ScriptLoad(ctx, `
	if redis.call('EXISTS', KEYS[1]) == 0 then
	redis.call('HSET', KEYS[1], 'Key', ARGV[2])
	redis.call('HSET', KEYS[1], 'Value', ARGV[3])
	redis.call('HSET', KEYS[1], 'TxnId', ARGV[4])
	redis.call('HSET', KEYS[1], 'TxnState', ARGV[5])
	redis.call('HSET', KEYS[1], 'TValid', ARGV[6])
	redis.call('HSET', KEYS[1], 'TLease', ARGV[7])
	redis.call('HSET', KEYS[1], 'Version', ARGV[8])
	redis.call('HSET', KEYS[1], 'Prev', ARGV[9])
	redis.call('HSET', KEYS[1], 'LinkedLen', ARGV[10])
	redis.call('HSET', KEYS[1], 'IsDeleted', ARGV[11])
	return redis.call('HGETALL', KEYS[1])
else
	return redis.error_reply('version mismatch')
end
    `).Result()
		if err != nil {
			return "", err
		}
		newVer := util.AddToString(value.Version(), 1)

		_, err = r.rdb.EvalSha(ctx, sha, []string{value.Key()}, value.Version(), value.Key(),
			value.Value(), value.TxnId(), value.TxnState(), value.TValid(), value.TLease(),
			newVer, value.Prev(), value.LinkedLen(), value.IsDeleted()).Result()
		if err != nil {
			if err.Error() == "version mismatch" {
				return "", errors.New(txn.VersionMismatch)
			}
			return "", err
		}
		return newVer, nil
	} else {
		ctx := context.Background()
		sha, err := r.rdb.ScriptLoad(ctx, `
		if redis.call('HGET', KEYS[1], 'Version') == ARGV[1] then
		redis.call('HSET', KEYS[1], 'Key', ARGV[2])
		redis.call('HSET', KEYS[1], 'Value', ARGV[3])
		redis.call('HSET', KEYS[1], 'TxnId', ARGV[4])
		redis.call('HSET', KEYS[1], 'TxnState', ARGV[5])
		redis.call('HSET', KEYS[1], 'TValid', ARGV[6])
		redis.call('HSET', KEYS[1], 'TLease', ARGV[7])
		redis.call('HSET', KEYS[1], 'Version', ARGV[8])
		redis.call('HSET', KEYS[1], 'Prev', ARGV[9])
		redis.call('HSET', KEYS[1], 'LinkedLen', ARGV[10])
		redis.call('HSET', KEYS[1], 'IsDeleted', ARGV[11])
		return redis.call('HGETALL', KEYS[1])
	else
		return redis.error_reply('version mismatch')
	end
		`).Result()
		if err != nil {
			return "", err
		}
		newVer := util.AddToString(value.Version(), 1)

		_, err = r.rdb.EvalSha(ctx, sha, []string{value.Key()}, value.Version(), value.Key(),
			value.Value(), value.TxnId(), value.TxnState(), value.TValid(), value.TLease(),
			newVer, value.Prev(), value.LinkedLen(), value.IsDeleted()).Result()
		if err != nil {
			if err.Error() == "version mismatch" {
				return "", errors.New(txn.VersionMismatch)
			}
			return "", err
		}
		return newVer, nil
	}

}

// Get retrieves the value associated with the given key from the Redis database.
// If the key is not found, it returns an empty string and an error indicating the key was not found.
// If an error occurs during the retrieval, it returns an empty string and the error.
// Otherwise, it returns the retrieved value and nil error.
func (r *RedisConnection) Get(name string) (string, error) {
	str, err := r.rdb.Get(context.Background(), name).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.New(txn.KeyNotFound)
		}
		return "", err
	}
	return str, nil
}

// Put stores the given value with the specified name in the Redis database.
// It will overwrite the value if the key already exists.
// It returns an error if the operation fails.
func (r *RedisConnection) Put(name string, value any) error {
	return r.rdb.Set(context.Background(), name, value, 0).Err()
}

// Delete removes the specified key from Redis.
// It allows for the deletion of a key that does not exist.
func (r *RedisConnection) Delete(name string) error {
	return r.rdb.Del(context.Background(), name).Err()
}
