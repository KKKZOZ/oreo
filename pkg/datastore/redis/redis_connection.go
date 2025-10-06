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
	atomicCreateItemSHA  string
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
	return redis.call('SET',ARGV[1] , ARGV[2])
else
	return redis.error_reply('already exists')
end
`

const AtomicCreateItemScript = `
if redis.call('EXISTS', KEYS[1]) == 0 then
	redis.call('HSET', KEYS[1], 'Key', ARGV[2])
	redis.call('HSET', KEYS[1], 'Value', ARGV[3])
	redis.call('HSET', KEYS[1], 'GroupKeyList', ARGV[4])
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
	redis.call('HSET', KEYS[1], 'GroupKeyList', ARGV[4])
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
	redis.call('HSET', KEYS[1], 'TValid', ARGV[4])
	return redis.call('HGETALL', KEYS[1])
else
	return redis.error_reply('version mismatch')
end
`

var defaultOptions = ConnectionOptions{
	Address:  "localhost:6379",
	Password: "",
	se:       serializer.NewJSONSerializer(),
	PoolSize: 60,
}

// NewRedisConnection creates a new Redis connection.
func NewRedisConnection(config *ConnectionOptions) *RedisConnection {
	// Start with a copy of the default configuration.
	finalConfig := defaultOptions

	// If the user provided a configuration, layer their values on top.
	// This avoids mutating the original 'config' object.
	if config != nil {
		if config.Address != "" {
			finalConfig.Address = config.Address
		}
		if config.Password != "" {
			finalConfig.Password = config.Password
		}
		if config.se != nil {
			finalConfig.se = config.se
		}
		// Note: This logic assumes 0 is not a valid user-set value for PoolSize.
		if config.PoolSize != 0 {
			finalConfig.PoolSize = config.PoolSize
		}
	}

	return &RedisConnection{
		rdb: redis.NewClient(&redis.Options{
			Addr:        finalConfig.Address,
			Password:    finalConfig.Password,
			PoolSize:    finalConfig.PoolSize,
			PoolTimeout: 10 * time.Second,
		}),
		Address: finalConfig.Address,
		se:      finalConfig.se,
	}
}

// Connect establishes a connection to the Redis server and loads Lua scripts.
func (r *RedisConnection) Connect() error {
	if r.connected {
		return nil
	}
	r.connected = true

	logger.Log.Debugw("Start Connect", "address", r.Address)
	defer logger.Log.Debugw("End   Connect", "address", r.Address)

	// Define all scripts and their destinations in one place.
	scriptsToLoad := []struct {
		scriptContent string
		destSHA       *string
	}{
		{scriptContent: AtomicCreateScript, destSHA: &r.atomicCreateSHA},
		{scriptContent: AtomicCreateItemScript, destSHA: &r.atomicCreateItemSHA},
		{scriptContent: ConditionalUpdateScript, destSHA: &r.conditionalUpdateSHA},
		{scriptContent: ConditionalCommitScript, destSHA: &r.conditionalCommitSHA},
	}

	var eg errgroup.Group
	ctx := context.Background()

	// Iterate and load each script concurrently.
	for _, s := range scriptsToLoad {
		script := s // Capture loop variable for the closure.
		eg.Go(func() error {
			sha, err := r.rdb.ScriptLoad(ctx, script.scriptContent).Result()
			if err != nil {
				return err
			}
			*script.destSHA = sha // Store the result in the destination field.
			return nil
		})
	}

	return eg.Wait()
}

// GetItem retrieves a structured transaction item stored as a Redis Hash.
// It returns a txn.DataItem, which encapsulates all transaction-related metadata.
func (r *RedisConnection) GetItem(key string) (txn.DataItem, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
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

// PutItem inserts or updates an item in Redis.
func (r *RedisConnection) PutItem(key string, value txn.DataItem) (string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx := context.Background()
	_, err := r.rdb.Pipelined(ctx, func(rdb redis.Pipeliner) error {
		rdb.HSet(ctx, key, "Key", value.Key())
		rdb.HSet(ctx, key, "Value", value.Value())
		rdb.HSet(ctx, key, "GroupKeyList", value.GroupKeyList())
		rdb.HSet(ctx, key, "TxnState", value.TxnState())
		rdb.HSet(ctx, key, "TValid", value.TValid())
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
	return value.Version(), nil
}

// ConditionalUpdate atomically updates an item if the version matches.
// If doCreate is true, it will create the item if it does not exist.
func (r *RedisConnection) ConditionalUpdate(
	key string,
	value txn.DataItem,
	doCreate bool,
) (string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	if doCreate {

		ctx := context.Background()
		newVer := util.AddToString(value.Version(), 1)

		_, err := r.rdb.EvalSha(ctx, r.atomicCreateItemSHA, []string{value.Key()}, value.Version(), value.Key(),
			value.Value(), value.GroupKeyList(), value.TxnState(), value.TValid(), value.TLease(),
			newVer, value.Prev(), value.LinkedLen(), value.IsDeleted()).
			Result()
		if err != nil {
			if err.Error() == "version mismatch" {
				logger.Log.Warnw(
					"Version mismatch in atomic create",
					"key", value.Key(),
					"current version",
					value.Version(),
				)
				return "", errors.New(txn.VersionMismatch)
			}
			return "", err
		}

		// fmt.Printf("Created new item with key %s and version %s\n", value.Key(), newVer)

		return newVer, nil
	}

	ctx := context.Background()
	newVer := util.AddToString(value.Version(), 1)

	_, err := r.rdb.EvalSha(ctx, r.conditionalUpdateSHA, []string{value.Key()}, value.Version(), value.Key(),
		value.Value(), value.GroupKeyList(), value.TxnState(), value.TValid(), value.TLease(),
		newVer, value.Prev(), value.LinkedLen(), value.IsDeleted()).
		Result()
	if err != nil {
		if err.Error() == "version mismatch" {
			logger.Log.Warnw(
				"Version mismatch in conditional update",
				"key", value.Key(),
				"current version",
				value.Version(),
			)
			return "", errors.New(txn.VersionMismatch)
		}
		return "", err
	}

	return newVer, nil
}

// ConditionalCommit atomically commits a transaction if the version matches.
func (r *RedisConnection) ConditionalCommit(
	key string,
	version string,
	tCommit int64,
) (string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx := context.Background()
	newVer := util.AddToString(version, 1)

	_, err := r.rdb.EvalSha(ctx, r.conditionalCommitSHA,
		[]string{key}, version, config.COMMITTED, newVer, tCommit).Result()
	if err != nil {
		if err.Error() == "version mismatch" {
			return "", errors.New(txn.VersionMismatch)
		}
		return "", err
	}
	return newVer, nil
}

// AtomicCreate creates a key-value pair if the key does not already exist.
func (r *RedisConnection) AtomicCreate(name string, value any) (string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx := context.Background()
	_, err := r.rdb.
		EvalSha(ctx, r.atomicCreateSHA, []string{name}, name, value).Result()
	if err != nil {
		if err.Error() == "already exists" {
			old, err := r.Get(name)
			if err != nil {
				return "", errors.New("get old state failed")
			}
			return old, errors.New(txn.KeyExists)
		}
		return "", err
	}
	return "", nil
}

// Get retrieves a simple string value for a given key.
// This is a general-purpose getter, distinct from GetItem, which retrieves a structured txn.DataItem.
func (r *RedisConnection) Get(name string) (string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
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

// Put sets the value for a given key.
func (r *RedisConnection) Put(name string, value any) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	return r.rdb.Set(context.Background(), name, value, 0).Err()
}

// Delete removes a key-value pair.
func (r *RedisConnection) Delete(name string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	return r.rdb.Del(context.Background(), name).Err()
}

// Close disconnects from the Redis server.
func (r *RedisConnection) Close() error {
	return r.rdb.Close()
}
