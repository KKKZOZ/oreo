package main

type MockMemoryDatastore struct {
	MemoryDatastore
}

func NewMockMemoryDatastore(name string, conn *MemoryConnection) *MockMemoryDatastore {
	return &MockMemoryDatastore{
		MemoryDatastore: MemoryDatastore{
			dataStore:  dataStore{Name: name},
			conn:       conn,
			readCache:  make(map[string]MemoryItem),
			writeCache: make(map[string]MemoryItem),
		},
	}
}

func (m *MockMemoryDatastore) conditionalUpdate(item Item) bool {
	return true
}
