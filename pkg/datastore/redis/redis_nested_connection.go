package redis

import (
	"context"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/errs"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

// RedisNestedConnection implements the txn.Connector interface.
var _ txn.Connector = (*RedisNestedConnection)(nil)

type RedisNestedConnection struct {
	rdb                  *redis.Client
	Address              string
	se                   serializer.Serializer
	connected            bool
	atomicCreateSHA      string
	atomicCreateItemSHA  string
	conditionalUpdateSHA string
	conditionalCommitSHA string
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

// NewRedisNestedConnection creates a new Redis connection using the provided configuration options.
func NewRedisNestedConnection(config *ConnectionOptions) *RedisNestedConnection {
	return &RedisNestedConnection{
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
func (r *RedisNestedConnection) Connect() error {

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
		sha, err := r.rdb.ScriptLoad(ctx, AtomicCreateItemScript).Result()
		if err != nil {
			return err
		}
		r.atomicCreateItemSHA = sha
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

	err := eg.Wait()
	if err != nil {
		return errs.NewConnectorError("Connect", err)
	}
	return nil
}

// GetItem retrieves a txn.DataItem from the Redis database based on the specified key.
// If the key is not found, it returns an empty txn.DataItem and an error.
func (r *RedisNestedConnection) GetItem(key string) (txn.DataItem, error) {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	var value RedisItem
	err := r.rdb.HGetAll(context.Background(), key).Scan(&value)
	if err != nil {
		return &RedisItem{}, errs.NewConnectorError("GetItem", err)
	}
	// Check if returned value is an empty struct
	if value.Empty() {
		return &RedisItem{}, errs.NewConnectorError("GetItem", errs.NewKeyNotFoundError(key, errs.NotFoundInDB))
	}
	return &value, nil
}

// PutItem puts an item into the Redis database with the specified key and value.
// It sets various fields of the txn.DataItem struct as hash fields in the Redis hash.
// The function returns an error if there was a problem executing the Redis commands.
func (r *RedisNestedConnection) PutItem(key string, value txn.DataItem) (string, error) {

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
		return "", errs.NewConnectorError("PutItem", err)
	}
	return "", nil
}

// ConditionalUpdate updates the value of a Redis item if the version matches the provided value.
// It takes a key string and a txn.DataItem value as parameters.
// If the item's version does not match, it returns a version mismatch error.
// Otherwise, it updates the item with the provided values and returns the updated item.
func (r *RedisNestedConnection) ConditionalUpdate(key string, value txn.DataItem, doCreate bool) (string, error) {

	debugStart := time.Now()

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	// logger.Log.Debugw("Start  ConditionalUpdate", "DataItem", value, "doCreate", doCreate, "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")
	defer func() {
		logger.Log.Debugw("End    ConditionalUpdate", "key", key, "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")
	}()

	if doCreate {
		ctx := context.Background()
		newVer := util.AddToString(value.Version(), 1)

		_, err := r.rdb.EvalSha(ctx, r.atomicCreateItemSHA, []string{value.Key()}, value.Version(), value.Key(),
			value.Value(), value.GroupKeyList(), value.TxnState(), value.TValid(), value.TLease(), newVer,
			value.Prev(), value.LinkedLen(), value.IsDeleted()).Result()

		if err != nil {
			if err.Error() == "version mismatch" {
				err = errs.NewVersionMismatchError(value.Key(), "unkown in ConditionalUpdate", value.Version(), err)
			}
			return "", errs.NewConnectorError("ConditionalUpdate", err)
		}
		return newVer, nil
	}

	ctx := context.Background()
	newVer := util.AddToString(value.Version(), 1)

	_, err := r.rdb.EvalSha(ctx, r.conditionalUpdateSHA, []string{value.Key()}, value.Version(), value.Key(),
		value.Value(), value.GroupKeyList(), value.TxnState(), value.TValid(), value.TLease(), newVer,
		value.Prev(), value.LinkedLen(), value.IsDeleted()).Result()

	if err != nil {
		if err.Error() == "version mismatch" {
			err = errs.NewVersionMismatchError(value.Key(), "unkown in ConditionalUpdate", value.Version(), err)
		}
		return "", errs.NewConnectorError("ConditionalUpdate", err)
	}

	return newVer, nil
}

// ConditionalCommit updates the txnState and version of a Redis item if the version matches the provided value.
// It takes a key string and a version string as parameters.
// If the item's version does not match, it returns a version mismatch error.
// Otherwise, it updates the item with the provided values and returns the updated item.
func (r *RedisNestedConnection) ConditionalCommit(key string, version string, tCommit int64) (string, error) {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	logger.Log.Debugw("Start  ConditionalCommit", "key", key)
	defer logger.Log.Debugw("End    ConditionalCommit", "key", key)

	ctx := context.Background()
	newVer := util.AddToString(version, 1)

	_, err := r.rdb.EvalSha(ctx, r.conditionalCommitSHA,
		[]string{key}, version, config.COMMITTED, newVer, tCommit).Result()
	if err != nil {
		if err.Error() == "version mismatch" {
			err = errs.NewVersionMismatchError(key, "unkown in ConditionalUpdate", version, err)
		}
		return "", errs.NewConnectorError("ConditionalCommit", err)
	}
	return newVer, nil

}

func (r *RedisNestedConnection) AtomicCreate(name string, value any) (string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx := context.Background()
	_, err := r.rdb.
		EvalSha(ctx, r.atomicCreateSHA, []string{name}, name, value).Result()
	if err != nil {
		// if err.Error() == "already exists" {
		// 	old, err := r.Get(name)
		// 	if err != nil {
		// 		return "", errors.New("get old state failed")
		// 	}
		// 	return old, errors.New(txn.KeyExists)
		// }
		return "", errs.NewConnectorError("AtomicCreate", err)
	}
	return "", nil
}

// Get retrieves the value associated with the given key from the Redis database.
// If the key is not found, it returns an empty string and an error indicating the key was not found.
// If an error occurs during the retrieval, it returns an empty string and the error.
// Otherwise, it returns the retrieved value and nil error.
func (r *RedisNestedConnection) Get(name string) (string, error) {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	str, err := r.rdb.Get(context.Background(), name).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errs.NewConnectorError("Get", errs.NewKeyNotFoundError(name, errs.NotFoundInDB))
		}
		return "", err
	}
	return str, nil
}

// Put stores the given value with the specified name in the Redis database.
// It will overwrite the value if the key already exists.
// It returns an error if the operation fails.
func (r *RedisNestedConnection) Put(name string, value any) error {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}
	err := r.rdb.Set(context.Background(), name, value, 0).Err()
	if err != nil {
		return errs.NewConnectorError("Put", err)
	}
	return nil
}

// Delete removes the specified key from Redis.
// It allows for the deletion of a key that does not exist.
func (r *RedisNestedConnection) Delete(name string) error {

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	err := r.rdb.Del(context.Background(), name).Err()
	if err != nil {
		return errs.NewConnectorError("Delete", err)
	}
	return nil
}
