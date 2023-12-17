package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func (m *MemoryConnection) connect() error {
	return nil
}

func (m *MemoryConnection) get(key string, value any) error {
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

func (m *MemoryConnection) put(key string, value any) error {
	httpClient := &http.Client{}
	jsonByte, err := json.Marshal(value)
	if err != nil {
		return err
	}
	jsonStr := string(jsonByte)
	url := fmt.Sprintf("%s/put/%s?value=%s", m.baseURL, key, jsonStr)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	response, _ := httpClient.Do(req)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("put failed")
	}
	return nil
}

func (m *MemoryConnection) delete(key string) error {
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
