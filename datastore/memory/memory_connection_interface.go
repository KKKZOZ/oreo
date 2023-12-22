package memory

// MemoryConnectionInterface represents an interface for a memory connection.
type MemoryConnectionInterface interface {
	// Connect establishes a connection to the memory database.
	Connect() error

	// Get retrieves the value associated with the given key from the memory database.
	// The value is stored in the 'value' parameter.
	Get(key string, value any) error

	// Put stores the given value in the memory database with the specified key.
	Put(key string, value any) error

	// Delete removes the value associated with the given key from the memory database.
	Delete(key string) error
}
