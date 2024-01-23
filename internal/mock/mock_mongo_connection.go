package mock

import (
	"fmt"

	"github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"github.com/kkkzoz/oreo/pkg/txn"
)

// MockMongoConnection is a mock of MongoConnection
// When Put is called, it will return error when debugCounter is 0
// Semantically, it means `Put()` call will succeed X times
type MockMongoConnection struct {
	*mongo.MongoConnection
	debugCounter int
	debugFunc    func() error
	isReturned   bool
	PutTimes     int
	GetTimes     int
}

func NewMockMongoConnection(address string, port int, limit int,
	isReturned bool, debugFunc func() error) *MockMongoConnection {
	conn := mongo.NewMongoConnection(&mongo.ConnectionOptions{
		Address:        fmt.Sprintf("mongodb://%s:%d", address, port),
		DBName:         "oreo",
		CollectionName: "records",
	})
	return &MockMongoConnection{
		MongoConnection: conn,
		debugCounter:    limit,
		debugFunc:       debugFunc,
		isReturned:      isReturned,
		PutTimes:        0,
		GetTimes:        0,
	}
}

func (m *MockMongoConnection) GetItem(key string) (txn.DataItem2, error) {
	defer func() { m.GetTimes++ }()
	return m.MongoConnection.GetItem(key)
}

func (m *MockMongoConnection) Get(name string) (string, error) {
	defer func() { m.GetTimes++ }()
	return m.MongoConnection.Get(name)
}

func (m *MockMongoConnection) ConditionalUpdate(key string, value txn.DataItem2, doCreate bool) error {
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.MongoConnection.ConditionalUpdate(key, value, doCreate)
}

func (m *MockMongoConnection) PutItem(key string, value txn.DataItem2) error {
	defer func() { m.debugCounter--; m.PutTimes++ }()
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
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.MongoConnection.Put(name, value)
}

func (m *MockMongoConnection) Delete(name string) error {
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.MongoConnection.Delete(name)
}
