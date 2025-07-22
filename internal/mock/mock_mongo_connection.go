package mock

import (
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ txn.Connector = (*MockMongoConnection)(nil)

// MockMongoConnection is a mock of MongoConnection
// When Put is called, it will return error when debugCounter is 0
// Semantically, it means `Put()` call will succeed X times
type MockMongoConnection struct {
	*mongo.MongoConnection
	debugCounter int
	debugFunc    func() error
	isReturned   bool
	networkDelay time.Duration
	PutTimes     int
	GetTimes     int
}

func NewMockMongoConnection(address string, port int, username string, password string, limit int,
	isReturned bool, networkDelay time.Duration, debugFunc func() error,
) *MockMongoConnection {
	conn := mongo.NewMongoConnection(&mongo.ConnectionOptions{
		Address:        fmt.Sprintf("mongodb://%s:%d", address, port),
		Username:       username,
		Password:       password,
		DBName:         "oreo",
		CollectionName: "records",
	})
	return &MockMongoConnection{
		MongoConnection: conn,
		debugCounter:    limit,
		debugFunc:       debugFunc,
		isReturned:      isReturned,
		networkDelay:    networkDelay,
		PutTimes:        0,
		GetTimes:        0,
	}
}

func (m *MockMongoConnection) GetItem(key string) (txn.DataItem, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.GetTimes++ }()
	return m.MongoConnection.GetItem(key)
}

func (m *MockMongoConnection) Get(name string) (string, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.GetTimes++ }()
	return m.MongoConnection.Get(name)
}

func (m *MockMongoConnection) ConditionalUpdate(
	key string,
	value txn.DataItem,
	doCreate bool,
) (string, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return "", m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.MongoConnection.ConditionalUpdate(key, value, doCreate)
}

func (m *MockMongoConnection) PutItem(key string, value txn.DataItem) (string, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return "", m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.MongoConnection.PutItem(key, value)
}

func (m *MockMongoConnection) Put(name string, value any) error {
	time.Sleep(m.networkDelay)
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
	time.Sleep(m.networkDelay)
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
