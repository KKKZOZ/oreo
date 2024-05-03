package redis

import (
	"context"
	"time"

	"github.com/go-errors/errors"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/logger"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

// RedisConnection implements the txn.Connector interface.
var _ txn.Connector = (*RedisConnection)(nil)

type RedisConnection struct {
	rdb                  *redis.Client
	Address              string
	se                   serializer.Serializer
	connected            bool
	atomicCreateSHA      string
	conditionalUpdateSHA string
	conditionalCommitSHA string
}

type ConnectionOptions struct {
	Address  string
	Password string
	se       serializer.Serializer
	PoolSize int
}

const AtomicCreateScript = `
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
`

const ConditionalUpdateScript = `
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
`

const ConditionalCommitScript = `
if redis.call('HGET', KEYS[1], 'Version') == ARGV[1] then
	redis.call('HSET', KEYS[1], 'TxnState', ARGV[2])
	redis.call('HSET', KEYS[1], 'Version', ARGV[3])
	return redis.call('HGETALL', KEYS[1])
else
	return redis.error_reply('version mismatch')
end
`

// NewRedisConnection creates a new Redis connection using the provided configuration options.
// If the config parameter is nil, default values will be used.
//
// The Redis connection is established using the specified address and password.
// The address format should be in the form "host:port".
//
// The se parameter is used for data serialization and deserialization.
// If se is nil, a default JSON serializer will be used.
//
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

	if config.PoolSize == 0 {
		config.PoolSize = 60
	}

	return &RedisConnection{
		rdb: redis.NewClient(&redis.Options{
			Addr:     config.Address,
			Password: config.Password,
			PoolSize: config.PoolSize,
		}),
		Address: config.Address,
		se:      config.se,
	}
}

// Connect establishes a connection to the Redis server.
// It returns an error if the connection cannot be established.
func (r *RedisConnection) Connect() error {

	if r.connected {
		return nil
	}

	r.connected = true

	logger.Log.Debugw("Start Connect", "address", r.Address)
	defer logger.Log.Debugw("End   Connect", "address", r.Address)

	var eg errgroup.Group
	ctx := context.Background()

	eg.Go(func() error {
		sha, err := r.rdb.ScriptLoad(ctx, AtomicCreateScript).Result()
		if err != nil {
			return err
		}
		r.atomicCreateSHA = sha
		return nil
	})

	eg.Go(func() error {
		sha, err := r.rdb.ScriptLoad(ctx, ConditionalUpdateScript).Result()
		if err != nil {
			return err
		}
		r.conditionalUpdateSHA = sha
		return nil
	})

	eg.Go(func() error {
		sha, err := r.rdb.ScriptLoad(ctx, ConditionalCommitScript).Result()
		if err != nil {
			return err
		}
		r.conditionalCommitSHA = sha
		return nil
	})

	return eg.Wait()
}

// GetItem retrieves a txn.DataItem from the Redis database based on the specified key.
// If the key is not found, it returns an empty txn.DataItem and an error.
func (r *RedisConnection) GetItem(key string) (txn.DataItem, error) {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.AdditionalLatency)
	}

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

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.AdditionalLatency)
	}

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

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.AdditionalLatency)
	}

	logger.Log.Debugw("Start  ConditionalUpdate", "key", key)
	defer logger.Log.Debugw("End    ConditionalUpdate", "key", key)

	if doCreate {
		ctx := context.Background()
		newVer := util.AddToString(value.Version(), 1)

		_, err := r.rdb.EvalSha(ctx, r.atomicCreateSHA, []string{value.Key()}, value.Version(), value.Key(),
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

	ctx := context.Background()
	newVer := util.AddToString(value.Version(), 1)

	_, err := r.rdb.EvalSha(ctx, r.conditionalUpdateSHA, []string{value.Key()}, value.Version(), value.Key(),
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

// ConditionalCommit updates the txnState and version of a Redis item if the version matches the provided value.
// It takes a key string and a version string as parameters.
// If the item's version does not match, it returns a version mismatch error.
// Otherwise, it updates the item with the provided values and returns the updated item.
func (r *RedisConnection) ConditionalCommit(key string, version string) (string, error) {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.AdditionalLatency)
	}

	logger.Log.Debugw("Start  ConditionalCommit", "key", key)
	defer logger.Log.Debugw("End    ConditionalCommit", "key", key)

	ctx := context.Background()
	newVer := util.AddToString(version, 1)

	_, err := r.rdb.EvalSha(ctx, r.conditionalCommitSHA,
		[]string{key}, version, config.COMMITTED, newVer).Result()
	if err != nil {
		if err.Error() == "version mismatch" {
			return "", errors.New(txn.VersionMismatch)
		}
		return "", err
	}
	return newVer, nil

}

// Get retrieves the value associated with the given key from the Redis database.
// If the key is not found, it returns an empty string and an error indicating the key was not found.
// If an error occurs during the retrieval, it returns an empty string and the error.
// Otherwise, it returns the retrieved value and nil error.
func (r *RedisConnection) Get(name string) (string, error) {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.AdditionalLatency)
	}

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

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.AdditionalLatency)
	}

	return r.rdb.Set(context.Background(), name, value, 0).Err()
}

// Delete removes the specified key from Redis.
// It allows for the deletion of a key that does not exist.
func (r *RedisConnection) Delete(name string) error {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.AdditionalLatency)
	}

	return r.rdb.Del(context.Background(), name).Err()
}
