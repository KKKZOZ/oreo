package memory

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/kkkzoz/vanilla-icecream/util"
)

// MockMemoryConnection is a mock of MemoryConnection
// When Put is called, it will return error when debugCounter is 0
// Semantically, it means `Put()` call will succeed X times
type MockMemoryConnection struct {
	*MemoryConnection
	debugCounter int
	debugFunc    func() error
	isReturned   bool
	callTimes    int
}

func NewMockMemoryConnection(address string, port int, limit int,
	isReturned bool, debugFunc func() error) *MockMemoryConnection {
	conn := NewMemoryConnection(address, port)
	return &MockMemoryConnection{
		MemoryConnection: conn,
		debugCounter:     limit,
		debugFunc:        debugFunc,
		isReturned:       isReturned,
		callTimes:        0,
	}
}

func (m *MockMemoryConnection) Put(key string, value any) error {
	defer func() { m.debugCounter--; m.callTimes++ }()
	if m.debugCounter == 0 {
		if m.isReturned {
			return m.debugFunc()
		} else {
			m.debugFunc()
		}
	}

	jsonStr := util.ToJSONString(value)
	path := fmt.Sprintf("%s/put/%s", m.baseURL, key)
	baseURL, err := url.Parse(path)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Add("value", jsonStr)
	baseURL.RawQuery = params.Encode()

	req, _ := http.NewRequest("POST", baseURL.String(), nil)

	httpClient := &http.Client{}
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("put failed")
	}
	return nil
}
