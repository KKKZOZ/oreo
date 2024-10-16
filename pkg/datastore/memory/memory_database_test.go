package memory

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/oreo-dtx-lab/oreo/internal/util"
)

func TestServerStartAndStop(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

}

func TestGetNormal(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	memoryDatabase.records["1"] = "hello"
	req, _ := http.NewRequest("GET", "http://localhost:8321/get/1", nil)
	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusOK {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusOK,
		)
	}
	body := util.GetBodyString(response)

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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
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
	body := util.GetBodyString(response)

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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
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
	body := util.GetBodyString(response)

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

func TestPutInvalidForm(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
	time.Sleep(100 * time.Millisecond)

	baseURL, err := url.Parse("http://localhost:8321/put/1")
	if err != nil {
		log.Fatal(err)
	}
	baseURL.RawQuery = "invalid;;;"

	req, _ := http.NewRequest("POST", baseURL.String(), nil)

	client := &http.Client{}
	response, _ := client.Do(req)

	if response.StatusCode != http.StatusBadRequest {
		t.Errorf(
			"Handler returned wrong status code: got %v want %v",
			response.StatusCode, http.StatusBadRequest,
		)
	}
	body := util.GetBodyString(response)

	expected := "Bad request"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
}

func TestPutEmpty(t *testing.T) {
	memoryDatabase := NewMemoryDatabase("localhost", 8321)
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
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
	body := util.GetBodyString(response)

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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
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
	body := util.GetBodyString(response)

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
	body = util.GetBodyString(response)

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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
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
	body := util.GetBodyString(response)

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
	body = util.GetBodyString(response)

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
	body = util.GetBodyString(response)

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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
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
	body := util.GetBodyString(response)

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
	body = util.GetBodyString(response)

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
	body = util.GetBodyString(response)

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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
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
	body := util.GetBodyString(response)

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
	go memoryDatabase.Start()
	defer func() { <-memoryDatabase.MsgChan }()
	defer func() { go memoryDatabase.Stop() }()
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
	body := util.GetBodyString(response)

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
	body = util.GetBodyString(response)

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
	body = util.GetBodyString(response)

	expected = "Key not found"
	if body != expected {
		t.Errorf(
			"Handler returned unexpected body: got %v want %v",
			body, expected,
		)
	}
}

func TestMemoryDatabase_BadParameterValidation(t *testing.T) {
	db := NewMemoryDatabase("localhost", 8321)
	cases := []struct {
		name      string
		method    string
		path      string
		form      url.Values
		wantCode  int
		wantError string
	}{
		{
			name:      "missing value in PUT API",
			method:    "POST",
			path:      "/put/test_key",
			form:      url.Values{},
			wantCode:  http.StatusBadRequest,
			wantError: "Value is empty",
		},
		{
			name:      "non-existed key in GET",
			method:    "GET",
			path:      "/get/test_key",
			form:      nil,
			wantCode:  http.StatusNotFound,
			wantError: "Key not found",
		},
		{
			name:      "non-existed key in DELETE",
			method:    "DELETE",
			path:      "/delete/test_key",
			form:      nil,
			wantCode:  http.StatusNotFound,
			wantError: "Key not found",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, tc.path, strings.NewReader(tc.form.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/get/{key}", db.serveGet).Methods("GET")
			router.HandleFunc("/put/{key}", db.servePut).Methods("POST")
			router.HandleFunc("/delete/{key}", db.serveDelete).Methods("DELETE")

			router.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Errorf("expected code %d, got %d", tc.wantCode, rec.Code)
			}

			if body, _ := io.ReadAll(rec.Body); strings.TrimSpace(string(body)) != tc.wantError {
				t.Errorf("expected error '%s', got '%s'", tc.wantError, body)
			}
		})
	}
}
