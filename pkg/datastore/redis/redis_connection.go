package redis

import (
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

type ConnectionOptions struct {
	Address      string
	Password     string
	se           serializer.Serializer
	PoolSize     int
	DataItemType txn.DataItemType
}

var defaultConfig = &ConnectionOptions{
	Address:      "localhost:6379",
	Password:     "kkkzoz",
	PoolSize:     60,
	se:           serializer.NewJSONSerializer(),
	DataItemType: txn.NestedItemType,
}

func NewRedisConnection(opts *ConnectionOptions) txn.Connector {

	if opts == nil {
		opts = defaultConfig
	} else {
		if opts.Address == "" {
			opts.Address = defaultConfig.Address
		}
		if opts.PoolSize == 0 {
			opts.PoolSize = defaultConfig.PoolSize
		}
		if opts.se == nil {
			opts.se = defaultConfig.se
		}
	}

	if opts.DataItemType == txn.NestedItemType {
		return NewRedisNestedConnection(opts)
	}
	if opts.DataItemType == txn.FlattenedItemType {
		return NewRedisFlattenedConnection(opts)
	}
	return nil
}
