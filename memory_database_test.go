package main

import (
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestServerStartAndStop(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go func() {
		err := memoryDatabase.start()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Memory database Start failed: %v\n", err)
		}
	}()
	defer memoryDatabase.stop()
	time.Sleep(100 * time.Millisecond)
}

func TestGetNormal(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	memoryDatabase.records["1"] = "hello"
	go func() {
		err := memoryDatabase.start()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Memory database Start failed: %v\n", err)
		}
	}()
	defer memoryDatabase.stop()
	time.Sleep(100 * time.Millisecond)

	req, _ := http.NewRequest("GET", "http://localhost:8321/get/1", nil)
	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body := getBodyString(response)

	expected := "hello"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
}

func TestGetNotFound(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go func() {
		err := memoryDatabase.start()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Memory database Start failed: %v\n", err)
		}
	}()
	defer memoryDatabase.stop()
	time.Sleep(100 * time.Millisecond)

	req, _ := http.NewRequest("GET", "http://localhost:8321/get/1", nil)
	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusNotFound {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusNotFound,
		)
	}
	body := getBodyString(response)

	expected := "Key not found"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
}

func TestPutNormal(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go func() {
		err := memoryDatabase.start()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Memory database Start failed: %v\n", err)
		}
	}()
	defer memoryDatabase.stop()
	time.Sleep(100 * time.Millisecond)

	baseURL, err := url.Parse("http://localhost:8321/put/1")
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("value", "hello")
	baseURL.RawQuery = params.Encode()
	req, _ := http.NewRequest("POST", baseURL.String(), nil)

	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body := getBodyString(response)

	expected := "OK"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
	if memoryDatabase.records["1"] != "hello" {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			memoryDatabase.records["1"], "hello",
		)
	}
}

func TestPutEmpty(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go func() {
		err := memoryDatabase.start()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Memory database Start failed: %v\n", err)
		}
	}()
	defer memoryDatabase.stop()
	time.Sleep(100 * time.Millisecond)

	baseURL, err := url.Parse("http://localhost:8321/put/1")
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("value", "")
	baseURL.RawQuery = params.Encode()
	req, _ := http.NewRequest("POST", baseURL.String(), nil)

	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusBadRequest {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusBadRequest,
		)
	}
	body := getBodyString(response)

	expected := "Value is empty"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
}

func TestPutAndGet(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go func() {
		err := memoryDatabase.start()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Memory database Start failed: %v\n", err)
		}
	}()
	defer memoryDatabase.stop()
	time.Sleep(100 * time.Millisecond)

	baseURL, err := url.Parse("http://localhost:8321/put/1")
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("value", "hello")
	baseURL.RawQuery = params.Encode()
	req, _ := http.NewRequest("POST", baseURL.String(), nil)

	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body := getBodyString(response)

	expected := "OK"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}

	req, _ = http.NewRequest("GET", "http://localhost:8321/get/1", nil)
	response, _ = client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body = getBodyString(response)

	expected = "hello"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
}

func TestReplaceAndGet(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	baseURL, err := url.Parse("http://localhost:8321/put/1")
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("value", "hello")
	baseURL.RawQuery = params.Encode()
	req, _ := http.NewRequest("POST", baseURL.String(), nil)

	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body := getBodyString(response)

	expected := "OK"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}

	baseURL, err = url.Parse("http://localhost:8321/put/1")
	if err != nil {
		log.Fatal(err)
	}

	params = url.Values{}
	params.Add("value", "world")
	baseURL.RawQuery = params.Encode()
	req, _ = http.NewRequest("POST", baseURL.String(), nil)

	response, _ = client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body = getBodyString(response)

	expected = "OK"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}

	req, _ = http.NewRequest("GET", "http://localhost:8321/get/1", nil)
	response, _ = client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body = getBodyString(response)

	expected = "world"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
}

func TestGetAndDelete(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	baseURL, err := url.Parse("http://localhost:8321/put/1")
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("value", "hello")
	baseURL.RawQuery = params.Encode()
	req, _ := http.NewRequest("POST", baseURL.String(), nil)

	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body := getBodyString(response)

	expected := "OK"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}

	req, _ = http.NewRequest("DELETE", "http://localhost:8321/delete/1", nil)
	response, _ = client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body = getBodyString(response)

	expected = "OK"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}

	req, _ = http.NewRequest("GET", "http://localhost:8321/get/1", nil)
	response, _ = client.Do(req)

	if response.StatusCode != http.StatusNotFound {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusNotFound,
		)
	}
	body = getBodyString(response)

	expected = "Key not found"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
}

func TestDeleteNotFound(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)

	req, _ := http.NewRequest("DELETE", "http://localhost:8321/delete/1", nil)
	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusNotFound {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusNotFound,
		)
	}
	body := getBodyString(response)

	expected := "Key not found"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
}

func TestDeleteTwice(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.start()

	time.Sleep(100 * time.Millisecond)
	baseURL, err := url.Parse("http://localhost:8321/put/1")
	if err != nil {
		log.Fatal(err)
	}

	params := url.Values{}
	params.Add("value", "hello")
	baseURL.RawQuery = params.Encode()
	req, _ := http.NewRequest("POST", baseURL.String(), nil)

	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body := getBodyString(response)

	expected := "OK"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}

	req, _ = http.NewRequest("DELETE", "http://localhost:8321/delete/1", nil)
	response, _ = client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body = getBodyString(response)

	expected = "OK"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}

	req, _ = http.NewRequest("DELETE", "http://localhost:8321/delete/1", nil)
	response, _ = client.Do(req)

	if response.StatusCode != http.StatusNotFound {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusNotFound,
		)
	}
	body = getBodyString(response)

	expected = "Key not found"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
}
