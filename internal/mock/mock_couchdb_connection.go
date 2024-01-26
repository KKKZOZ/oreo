package mock

import (
	"fmt"
	"time"

	"github.com/kkkzoz/oreo/pkg/datastore/couchdb"
	"github.com/kkkzoz/oreo/pkg/txn"
)

var _ txn.Connector = (*MockCouchDBConnection)(nil)

// MockCouchDBConnection is a mock of CouchDBConnection
// When Put is called, it will return error when debugCounter is 0
// Semantically, it means `Put()` call will succeed X times
type MockCouchDBConnection struct {
	*couchdb.CouchDBConnection
	debugCounter int
	debugFunc    func() error
	isReturned   bool
	networkDelay time.Duration
	PutTimes     int
	GetTimes     int
}

func NewMockCouchDBConnection(address string, port int, limit int,
	isReturned bool, networkDelay time.Duration, debugFunc func() error) *MockCouchDBConnection {
	conn := couchdb.NewCouchDBConnection(&couchdb.ConnectionOptions{
		Address: fmt.Sprintf("http://admin:password@%s:%d", address, port),
		DBName:  "oreo",
	})
	return &MockCouchDBConnection{
		CouchDBConnection: conn,
		debugCounter:      limit,
		debugFunc:         debugFunc,
		isReturned:        isReturned,
		networkDelay:      networkDelay,
		PutTimes:          0,
		GetTimes:          0,
	}
}

func (m *MockCouchDBConnection) GetItem(key string) (txn.DataItem, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.GetTimes++ }()
	return m.CouchDBConnection.GetItem(key)
}

func (m *MockCouchDBConnection) Get(name string) (string, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.GetTimes++ }()
	return m.CouchDBConnection.Get(name)
}

func (m *MockCouchDBConnection) ConditionalUpdate(key string, value txn.DataItem, doCreate bool) (string, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return "", m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.CouchDBConnection.ConditionalUpdate(key, value, doCreate)
}

func (m *MockCouchDBConnection) PutItem(key string, value txn.DataItem) (string, error) {
	time.Sleep(m.networkDelay)
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return "", m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.CouchDBConnection.PutItem(key, value)
}

func (m *MockCouchDBConnection) Put(name string, value any) error {
	time.Sleep(m.networkDelay)
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.CouchDBConnection.Put(name, value)
}

func (m *MockCouchDBConnection) Delete(name string) error {
	time.Sleep(m.networkDelay)
	defer func() { m.debugCounter--; m.PutTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}
	return m.CouchDBConnection.Delete(name)
}
