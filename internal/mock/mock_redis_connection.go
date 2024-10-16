package mock

import (
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

// MockRedisConnection implements the txn.Connector interface.
var _ txn.Connector = (*MockRedisConnection)(nil)

// MockRedisConnection is a mock of RedisConnection
// When Put is called, it will return error when debugCounter is 0
// Semantically, it means `Put()` call will succeed X times
// If debugCounter is a negative number, it will never return errors
type MockRedisConnection struct {
	*redis.RedisConnection
	debugCounter int
	debugFunc    func() error
	isReturned   bool
	networkDelay time.Duration
	PutTimes     int
	GetTimes     int
}

func NewMockRedisConnection(address string, port int, limit int,
	isReturned bool, networkDelay time.Duration, debugFunc func() error) *MockRedisConnection {
	conn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address: fmt.Sprintf("%s:%d", address, port),
	})
	return &MockRedisConnection{
		RedisConnection: conn,
		debugCounter:    limit,
		debugFunc:       debugFunc,
		isReturned:      isReturned,
		networkDelay:    networkDelay,
		PutTimes:        0,
		GetTimes:        0,
	}
}

func (m *MockRedisConnection) GetItem(key string) (txn.DataItem, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.GetTimes++ }()
	return m.RedisConnection.GetItem(key)
}

func (m *MockRedisConnection) Get(name string) (string, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.GetTimes++ }()
	return m.RedisConnection.Get(name)
}

func (m *MockRedisConnection) ConditionalUpdate(key string, value txn.DataItem, doCreate bool) (string, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return "", m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.RedisConnection.ConditionalUpdate(key, value, doCreate)
}

func (m *MockRedisConnection) PutItem(key string, value txn.DataItem) (string, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return "", m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.RedisConnection.PutItem(key, value)
}

func (m *MockRedisConnection) Put(name string, value any) error {
	time.Sleep(m.networkDelay)
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.RedisConnection.Put(name, value)
}

func (m *MockRedisConnection) Delete(name string) error {
	time.Sleep(m.networkDelay)
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.RedisConnection.Delete(name)
}
