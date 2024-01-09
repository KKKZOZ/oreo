package config

import (
	"encoding/json"
	"time"
)

type DataStoreType string

const (
	MEMORY DataStoreType = "memory"
)

type config struct {
	LeaseTime       time.Duration
	MaxRecordLength int
	IdGenerator     IdGenerator
}

var Config = config{
	LeaseTime:       1000 * time.Millisecond,
	MaxRecordLength: 2,
	IdGenerator:     NewIncrementalGenerator(),
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
