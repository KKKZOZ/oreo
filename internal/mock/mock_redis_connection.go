package mock

import (
	"fmt"

	"github.com/kkkzoz/oreo/pkg/datastore/redis"
)

// MockRedisConnection is a mock of RedisConnection
// When Put is called, it will return error when debugCounter is 0
// Semantically, it means `Put()` call will succeed X times
type MockRedisConnection struct {
	*redis.RedisConnection
	debugCounter int
	debugFunc    func() error
	isReturned   bool
	callTimes    int
}

func NewMockRedisConnection(address string, port int, limit int,
	isReturned bool, debugFunc func() error) *MockRedisConnection {
	conn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address: fmt.Sprintf("%s:%d", address, port),
	})
	return &MockRedisConnection{
		RedisConnection: conn,
		debugCounter:    limit,
		debugFunc:       debugFunc,
		isReturned:      isReturned,
		callTimes:       0,
	}
}

func (m *MockRedisConnection) ConditionalUpdate(key string, value redis.RedisItem, doCreate bool) error {
	defer func() { m.debugCounter--; m.callTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.RedisConnection.ConditionalUpdate(key, value, doCreate)
}

func (m *MockRedisConnection) PutItem(key string, value redis.RedisItem) error {
	defer func() { m.debugCounter--; m.callTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.RedisConnection.PutItem(key, value)
}

func (m *MockRedisConnection) Put(name string, value any) error {
	defer func() { m.debugCounter--; m.callTimes++ }()
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
	defer func() { m.debugCounter--; m.callTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.RedisConnection.Delete(name)
}
