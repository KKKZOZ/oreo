package memory

import "github.com/kkkzoz/oreo/pkg/txn"

type MockMemoryDatastore struct {
	MemoryDatastore
}

func NewMockMemoryDatastore(name string, conn *MemoryConnection) *MockMemoryDatastore {
	return &MockMemoryDatastore{
		MemoryDatastore: MemoryDatastore{
			BaseDataStore: txn.BaseDataStore{Name: name},
			conn:          conn,
			readCache:     make(map[string]MemoryItem),
			writeCache:    make(map[string]MemoryItem),
		},
	}
}
