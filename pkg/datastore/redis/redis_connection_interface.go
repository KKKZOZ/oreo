package redis

type RedisConnectionInterface interface {
	Connect() error
	GetItem(key string) (RedisItem, error)
	PutItem(key string, value RedisItem) error
	ConditionalUpdate(key string, value RedisItem) error
	Get(name string) (string, error)
	Put(name string, value any) error
	Delete(name string) error
}
