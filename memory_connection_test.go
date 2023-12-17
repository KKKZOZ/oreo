package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestConnectionGetNormal(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	memConn := NewMemoryConnection("localhost", 8321)
	memConn.connect()

	key := "1"
	value := Person{
		Name: "John",
		Age:  30,
	}
	jsonByte, _ := json.Marshal(value)
	memoryDatabase.records[key] = string(jsonByte)

	var expected Person
	err := memConn.get(key, &expected)
	if err != nil {
		t.Error(err)
	}

	if expected != value {
		t.Errorf("got %v want %v", value, expected)
	}
}

func TestConnectionGetNotFound(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8322)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	memConn := NewMemoryConnection("localhost", 8322)
	memConn.connect()

	key := "1"
	var expected Person
	err := memConn.get(key, &expected)
	if err == nil {
		t.Errorf("got %v want %v", nil, err)
	}
}

func TestConnectionGetBrokenJSON(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8322)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	memConn := NewMemoryConnection("localhost", 8322)
	memConn.connect()

	key := "1"
	memoryDatabase.records[key] = "broken json"

	var expected Person
	err := memConn.get(key, &expected)
	if err == nil {
		t.Errorf("got %v want %v", nil, err)
	}
}

func TestConnectionPutNormal(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8323)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	memConn := NewMemoryConnection("localhost", 8323)
	memConn.connect()

	key := "1"
	value := Person{
		Name: "John",
		Age:  30,
	}
	err := memConn.put(key, value)
	if err != nil {
		t.Error(err)
	}

	var expected Person
	err = json.Unmarshal([]byte(memoryDatabase.records[key]), &expected)
	if err != nil {
		t.Error(err)
	}

	if expected != value {
		t.Errorf("got %v want %v", value, expected)
	}
}

func TestConnectionPutAndGet(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8324)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	memConn := NewMemoryConnection("localhost", 8324)
	memConn.connect()

	key := "1"
	value := Person{
		Name: "John",
		Age:  30,
	}
	err := memConn.put(key, value)
	if err != nil {
		t.Error(err)
	}

	var expected Person
	err = memConn.get(key, &expected)
	if err != nil {
		t.Error(err)
	}

	if expected != value {
		t.Errorf("got %v want %v", value, expected)
	}
}

func TestConnectionReplaceAndGet(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8325)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	memConn := NewMemoryConnection("localhost", 8325)
	memConn.connect()

	key := "1"
	value := Person{
		Name: "John",
		Age:  30,
	}
	err := memConn.put(key, value)
	if err != nil {
		t.Error(err)
	}

	value = Person{
		Name: "John",
		Age:  31,
	}
	err = memConn.put(key, value)
	if err != nil {
		t.Error(err)
	}

	var expected Person
	err = memConn.get(key, &expected)
	if err != nil {
		t.Error(err)
	}

	if expected != value {
		t.Errorf("got %v want %v", value, expected)
	}
}

func TestConnectionPutAndDelete(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8326)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	memConn := NewMemoryConnection("localhost", 8326)
	memConn.connect()

	key := "1"
	value := Person{
		Name: "John",
		Age:  30,
	}
	err := memConn.put(key, value)
	if err != nil {
		t.Error(err)
	}

	err = memConn.delete(key)
	if err != nil {
		t.Error(err)
	}

	var expected Person
	err = memConn.get(key, &expected)
	if err == nil {
		t.Errorf("got %v want %v", nil, err)
	}
}

func TestConnectionDeleteNotFound(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8327)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	memConn := NewMemoryConnection("localhost", 8327)
	memConn.connect()

	key := "1"
	err := memConn.delete(key)
	if err == nil {
		t.Errorf("got %v want %v", nil, err)
	}
}

func TestConnectionDeleteTwice(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8328)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	memConn := NewMemoryConnection("localhost", 8328)
	memConn.connect()

	key := "1"
	value := Person{
		Name: "John",
		Age:  30,
	}
	err := memConn.put(key, value)
	if err != nil {
		t.Error(err)
	}

	err = memConn.delete(key)
	if err != nil {
		t.Error(err)
	}

	err = memConn.delete(key)
	if err == nil {
		t.Errorf("got %v want %v", nil, err)
	}
}
