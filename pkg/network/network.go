package network

import (
	"time"

	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/txn"
)

type Response[T any] struct {
	Status string
	ErrMsg string
	Data   T
}

type ReadResponse struct {
	Status string
	ErrMsg string
	Data   *redis.RedisItem
}

type ReadRequest struct {
	Key       string
	StartTime time.Time
	Config    txn.RecordConfig
}

type PrepareRequest struct {
	ItemList   []redis.RedisItem
	StartTime  time.Time
	CommitTime time.Time
	Config     txn.RecordConfig
}

type CommitRequest struct {
	List []txn.CommitInfo
}

type AbortRequest struct {
	KeyList []string
	TxnId   string
}
