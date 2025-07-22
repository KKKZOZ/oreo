package workload

import (
	"context"

	"benchmark/ycsb"
)

type operationType int64

const (
	read operationType = iota + 1
	update
	insert
	scan
	readModifyWrite
	doubleSeqCommit
)

type datastoreType int64

const (
	redisDatastore1 datastoreType = iota + 1
	mongoDatastore1
	mongoDatastore2
	couchDatastore1
	kvrocksDatastore1
	cassandraDatastore1
	dynamodbDatastore1
	tikvDatastore1
)

const (
	MAX_VALUE_LENGTH = 100
)

type Workload interface {
	ResetKeySequence()
	Load(ctx context.Context, opCount int, db ycsb.DB)
	Run(ctx context.Context, opCount int, db ycsb.DB)
	NeedPostCheck() bool
	NeedRawDB() bool
	PostCheck(ctx context.Context, db ycsb.DB, resChan chan int)
	DisplayCheckResult()
}
