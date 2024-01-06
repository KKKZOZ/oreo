package mongo

import (
	"fmt"
)

// MockMongoConnection is a mock of MongoConnection
// When Put is called, it will return error when debugCounter is 0
// Semantically, it means `Put()` call will succeed X times
type MockMongoConnection struct {
	*MongoConnection
	debugCounter int
	debugFunc    func() error
	isReturned   bool
	callTimes    int
}

func NewMockMongoConnection(address string, port int, limit int,
	isReturned bool, debugFunc func() error) *MockMongoConnection {
	conn := NewMongoConnection(&ConnectionOptions{
		Address:        fmt.Sprintf("mongodb://%s:%d", address, port),
		DBName:         "oreo",
		CollectionName: "records",
	})
	return &MockMongoConnection{
		MongoConnection: conn,
		debugCounter:    limit,
		debugFunc:       debugFunc,
		isReturned:      isReturned,
		callTimes:       0,
	}
}

func (m *MockMongoConnection) ConditionalUpdate(key string, value MongoItem) error {
	defer func() { m.debugCounter--; m.callTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.MongoConnection.ConditionalUpdate(key, value)
}

func (m *MockMongoConnection) PutItem(key string, value MongoItem) error {
	defer func() { m.debugCounter--; m.callTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.MongoConnection.PutItem(key, value)
}

func (m *MockMongoConnection) Put(name string, value any) error {
	defer func() { m.debugCounter--; m.callTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.MongoConnection.Put(name, value)
}
