package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type MemoryConnection struct {
	Address string
	Port    int
	baseURL string
}

func NewMemoryConnection(address string, port int) *MemoryConnection {
	return &MemoryConnection{
		Address: address,
		Port:    port,
		baseURL: fmt.Sprintf("http://%s:%d", address, port),
	}
}

func (m *MemoryConnection) Connect() error {
	return nil
}

func (m *MemoryConnection) Get(key string, value any) error {
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/get/%s", m.baseURL, key)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	response, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return fmt.Errorf("key not found")
	}

	body := getBodyString(response)
	err = json.Unmarshal([]byte(body), value)
	return err
}

func (m *MemoryConnection) Put(key string, value any) error {
	jsonStr := toJSONString(value)
	path := fmt.Sprintf("%s/put/%s", m.baseURL, key)
	baseURL, err := url.Parse(path)
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("value", jsonStr)
	baseURL.RawQuery = params.Encode()

	req, err := http.NewRequest("POST", baseURL.String(), nil)
	if err != nil {
		return err
	}
	httpClient := &http.Client{}
	response, _ := httpClient.Do(req)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("put failed")
	}
	return nil
}

func (m *MemoryConnection) Delete(key string) error {
	httpClient := &http.Client{}
	url := fmt.Sprintf("%s/delete/%s", m.baseURL, key)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	response, _ := httpClient.Do(req)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed")
	}
	return nil
}
