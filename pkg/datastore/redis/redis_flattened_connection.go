package redis

import (
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/redis/go-redis/v9"
)

// RedisNestedConnection implements the txn.Connector interface.
var _ txn.Connector = (*RedisFlattenedConnection)(nil)

type RedisFlattenedConnection struct {
	rdb                  *redis.Client
	Address              string
	se                   serializer.Serializer
	connected            bool
	atomicCreateSHA      string
	atomicCreateItemSHA  string
	conditionalUpdateSHA string
	conditionalCommitSHA string
}

// AtomicCreate implements txn.Connector.
func (r *RedisFlattenedConnection) AtomicCreate(name string, value any) (string, error) {
	panic("unimplemented")
}

// ConditionalCommit implements txn.Connector.
func (r *RedisFlattenedConnection) ConditionalCommit(key string, version string, tCommit int64) (string, error) {
	panic("unimplemented")
}

// ConditionalUpdate implements txn.Connector.
func (r *RedisFlattenedConnection) ConditionalUpdate(key string, value txn.DataItem, doCreate bool) (string, error) {
	panic("unimplemented")
}

// Connect implements txn.Connector.
func (r *RedisFlattenedConnection) Connect() error {
	panic("unimplemented")
}

// Delete implements txn.Connector.
func (r *RedisFlattenedConnection) Delete(name string) error {
	panic("unimplemented")
}

// Get implements txn.Connector.
func (r *RedisFlattenedConnection) Get(name string) (string, error) {
	panic("unimplemented")
}

// GetItem implements txn.Connector.
func (r *RedisFlattenedConnection) GetItem(key string) (txn.DataItem, error) {
	panic("unimplemented")
}

// Put implements txn.Connector.
func (r *RedisFlattenedConnection) Put(name string, value any) error {
	panic("unimplemented")
}

// PutItem implements txn.Connector.
func (r *RedisFlattenedConnection) PutItem(key string, value txn.DataItem) (string, error) {
	panic("unimplemented")
}

func NewRedisFlattenedConnection(config *ConnectionOptions) *RedisFlattenedConnection {
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

	return &RedisFlattenedConnection{
		rdb: redis.NewClient(&redis.Options{
			Addr:     config.Address,
			Password: config.Password,
			PoolSize: config.PoolSize,
		}),
		Address: config.Address,
		se:      config.se,
	}
}
