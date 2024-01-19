package txn

type Connector interface {
	Connect() error
	GetItem(key string) (DataItem, error)
	PutItem(key string, value DataItem) error
	ConditionalUpdate(key string, value DataItem, doCreate bool) error
	Get(name string) (string, error)
	Put(name string, value any) error
	Delete(name string) error
}
