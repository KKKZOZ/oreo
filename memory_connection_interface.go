package main

type MemoryConnectionInterface interface {
	Connect() error
	Get(key string, value any) error
	Put(key string, value any) error
	Delete(key string) error
}
