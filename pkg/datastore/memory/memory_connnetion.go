package memory

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
)

type MemoryConnection struct {
	Address   string
	Port      int
	baseURL   string
	transport *http.Transport
}

func NewMemoryConnection(address string, port int) *MemoryConnection {
	return &MemoryConnection{
		Address: address,
		Port:    port,
		baseURL: fmt.Sprintf("http://%s:%d", address, port),
		transport: &http.Transport{
			MaxIdleConns:        6000,
			MaxConnsPerHost:     6000,
			MaxIdleConnsPerHost: 6000,
			IdleConnTimeout:     60 * time.Second,
		},
	}
}

func (m *MemoryConnection) Connect() error {
	return nil
}

func (m *MemoryConnection) Get(key string, value any) error {
	url := fmt.Sprintf("%s/get/%s", m.baseURL, key)
	req, _ := http.NewRequest("GET", url, nil)

	httpClient := &http.Client{Transport: m.transport}
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		_, _ = io.CopyN(io.Discard, response.Body, 1024*4)
		return fmt.Errorf("key not found")
	}

	body := util.GetBodyString(response)
	err = json.Unmarshal([]byte(body), value)
	return err
}

func (m *MemoryConnection) Put(key string, value any) error {
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

	httpClient := &http.Client{Transport: m.transport}
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.CopyN(io.Discard, response.Body, 1024*4)
		return fmt.Errorf("put failed")
	}
	return nil
}

func (m *MemoryConnection) Delete(key string) error {
	httpClient := &http.Client{Transport: m.transport}
	url := fmt.Sprintf("%s/delete/%s", m.baseURL, key)
	req, _ := http.NewRequest("DELETE", url, nil)
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.CopyN(io.Discard, response.Body, 1024*4)
		return fmt.Errorf("delete failed")
	}
	return nil
}
