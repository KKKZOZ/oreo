package main

type Datastore interface {
	start()
	read(key string) (string, error)
	write(key string, value string)
	prev(key string, record string)
	delete(key string)
	prepare(key string, record string)
	commit(key string, record string)
	abort(key string, record string)
	recover(key string)
}

type State int

const (
	PREPARED  State = 0
	COMMITTED State = 1
)
