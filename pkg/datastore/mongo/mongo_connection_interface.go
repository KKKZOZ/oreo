package mongo

type MongoConnectionInterface interface {
	Connect() error
	GetItem(key string) (MongoItem, error)
	PutItem(key string, value MongoItem) error
	ConditionalUpdate(key string, value MongoItem) error
	Get(name string) (string, error)
	Put(name string, value any) error
	Delete(name string) error
}
