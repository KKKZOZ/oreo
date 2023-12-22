package config

type DataStoreType string

const (
	MEMORY DataStoreType = "memory"
)

const LeastTime = 1000

type State int

const (
	EMPTY     State = 0
	STARTED   State = 1
	PREPARED  State = 2
	COMMITTED State = 3
	ABORTED   State = 4
)
