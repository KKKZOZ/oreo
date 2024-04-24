package config

import (
	"encoding/json"
	"time"

	"github.com/kkkzoz/oreo/pkg/serializer"
	"go.uber.org/zap/zapcore"
)

type DataStoreType string

type Mode string

const (
	REMOTE Mode = "remote"
	LOCAL  Mode = "local"

	MEMORY DataStoreType = "memory"

	// DEFAULT sets the concurrent optimization level to 0, which means that
	// when a write conflict occurs, one transaction will success
	DEFAULT int = 0

	// PARALLELIZE_ON_UPDATE sets the concurrent optimization level to 1, which
	// means there is NO guarantee that any transaction will success when a write
	// conflict happens in a single datastore
	PARALLELIZE_ON_UPDATE int = 1

	// PARALLELIZE_ON_PREPARE sets the concurrent optimization level to 2, which
	// means there is NO guarantee that any transaction will success when a write
	// conflict happens in different datastores
	PARALLELIZE_ON_PREPARE int = 2

	// AsyncLevelZero means no async commit
	AsyncLevelZero int = 0

	// AsyncLevelOne means async delete the TSR after return
	AsyncLevelOne int = 1

	// AsyncLevelTwo means async ds.Commit() and delete the TSR after return
	AsyncLevelTwo int = 2
)

type config struct {

	// Mode specifies the mode of the transaction.
	// It can be either REMOTE or LOCAL.
	Mode Mode

	// LeaseTime specifies the duration of time for which a record is leased.
	LeaseTime time.Duration

	// MaxRecordLength specifies the maximum length of a linked record.
	MaxRecordLength int

	// IdGenerator generates unique IDs for records.
	IdGenerator IdGenerator

	// Serializer serializes and deserializes records.
	Serializer serializer.Serializer

	// LogLevel specifies the logging level for the application.
	LogLevel zapcore.Level

	// ConcurrentUpdate specifies whether to allow concurrent conditional updates
	// in the datastore.Prepare() phase
	ConcurrentOptimizationLevel int

	// AsyncLevel specifies the level of asynchronous commit in the transaction
	AsyncLevel int

	// MaxOutstandingRequest specifies the maximum number of outstanding requests
	MaxOutstandingRequest int
}

var Config = config{
	LeaseTime:                   1000 * time.Millisecond,
	MaxRecordLength:             2,
	IdGenerator:                 NewUUIDGenerator(),
	Serializer:                  serializer.NewJSONSerializer(),
	LogLevel:                    zapcore.InfoLevel,
	ConcurrentOptimizationLevel: DEFAULT,
	AsyncLevel:                  AsyncLevelZero,
	MaxOutstandingRequest:       5,
}

type State int

// const (
// 	EMPTY     State = "EMPTY"
// 	STARTED   State = "STARTED"
// 	PREPARED  State = "PREPARED"
// 	COMMITTED State = "COMMITTED"
// 	ABORTED   State = "ABORTED"
// )

const (
	EMPTY     State = 0
	STARTED   State = 1
	PREPARED  State = 2
	COMMITTED State = 3
	ABORTED   State = 4
)

func (s State) MarshalBinary() (data []byte, err error) {
	// Use your preferred way to convert s to bytes (i.e., json.Marshal, gob.Encode, etc.)
	return json.Marshal(s)
}

func (s State) UnmarshalBinary(data []byte) error {
	// Use your preferred way to convert bytes in data back into struct s (i.e., json.Unmarshal, gob.Decode, etc.)
	return json.Unmarshal(data, &s)
}
