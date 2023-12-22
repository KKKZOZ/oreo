package timeoracle

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/kkkzoz/oreo/pkg/locker"
)

// TestSimpleTimeOracle_GetTime ensures that the GetTime function behaves correctly
func TestSimpleTimeOracle_GetTime(t *testing.T) {
	oracle := NewSimpleTimeOracle("localhost", 8080, locker.NewMemoryLocker())
	gotTime := oracle.GetTime()
	expectedTime := time.Now()

	duration := expectedTime.Sub(gotTime)
	if duration > time.Millisecond {
		t.Errorf("Duration between GetTime call and now is more than 1ms: %s", duration)
	}
}

// TestSimpleTimeOracle_ServeTime ensures that the serveTime handler returns the current time and correct HTTP status code
func TestSimpleTimeOracle_ServeTime(t *testing.T) {
	oracle := NewSimpleTimeOracle("localhost", 8080, locker.NewMemoryLocker())
	req, err := http.NewRequest("GET", oracle.baseURL+"/time", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(oracle.serveTime)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Status code differ. Expected %d .\n Got %d", http.StatusOK, rr.Code)
	}
}

// TestSimpleTimeOracle_ServeLock ensures that the serveLock handler locks correctly and returns correct HTTP status code
func TestSimpleTimeOracle_ServeLock(t *testing.T) {
	oracle := NewSimpleTimeOracle("localhost", 8080, locker.NewMemoryLocker())
	data := url.Values{}
	data.Set("key", "key")
	data.Set("id", "id1")
	data.Set("duration", strconv.Itoa(int(time.Second)))

	req, err := http.NewRequest("GET", oracle.baseURL+"/lock?"+data.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(oracle.serveLock)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || rr.Body.String() != "OK" {
		t.Errorf("Status code differ or Body not OK. Expected %d OK.\n Got %d %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

// TestSimpleTimeOracle_ServeUnlock ensures that the serveUnlock handler unlocks correctly and returns correct HTTP status code
func TestSimpleTimeOracle_ServeUnlock(t *testing.T) {
	oracle := NewSimpleTimeOracle("localhost", 8080, locker.NewMemoryLocker())
	data := url.Values{}
	data.Set("key", "key")
	data.Set("id", "id1")

	// Lock the key before unlocking
	err := oracle.locker.Lock("key", "id1", 2*time.Second)
	if err != nil {
		t.Fatal(fmt.Sprintf("Lock failed: %v", err))
	}

	req, err := http.NewRequest("GET", oracle.baseURL+"/unlock?"+data.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(oracle.serveUnlock)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || rr.Body.String() != "OK" {
		t.Errorf("Status code differ or Body not OK. Expected %d OK.\n Got %d %s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

// TestSimpleTimeOracle_ServeLockValidation checks if serveLock handles incorrect or missing parameters properly
func TestSimpleTimeOracle_ServeLockValidation(t *testing.T) {
	oracle := NewSimpleTimeOracle("localhost", 8081, locker.NewMemoryLocker())
	data := url.Values{}
	data.Set("key", "")
	data.Set("id", "id1")
	data.Set("duration", strconv.Itoa(int(time.Second)))

	req, err := http.NewRequest("GET", oracle.baseURL+"/lock?"+data.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(oracle.serveLock)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest || rr.Body.String() != "Missing parameters" {
		t.Errorf("Status code differ or Body not Lock failed. Expected %d Lock failed.\n Got %d %s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

// TestSimpleTimeOracle_ServeUnlockValidation checks if serveUnlock handles incorrect or missing parameters properly
func TestSimpleTimeOracle_ServeUnlockValidation(t *testing.T) {
	oracle := NewSimpleTimeOracle("localhost", 8081, locker.NewMemoryLocker())
	data := url.Values{}
	data.Set("key", "")
	data.Set("id", "id1")

	req, err := http.NewRequest("GET", oracle.baseURL+"/unlock?"+data.Encode(), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(oracle.serveUnlock)

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Status code differ. Expected %d.\n Got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestSimpleTimeOracle_LockUnlock checks if locking and unlocking functionality works as expected
func TestSimpleTimeOracle_LockUnlock(t *testing.T) {
	oracle := NewSimpleTimeOracle("localhost", 8081, locker.NewMemoryLocker())
	err := oracle.locker.Lock("key2", "id1", 2*time.Second)
	if err != nil {
		t.Fatalf("Lock failed: %v", err)
	}

	err = oracle.locker.Unlock("key2", "id1")
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	err = oracle.locker.Lock("key2", "id2", 2*time.Second)
	if err != nil {
		t.Fatalf("Reacquiring lock failed: %v", err)
	}
}

// TestSimpleTimeOracle_BadParameterValidation tests whether the bad requests are handled properly
func TestSimpleTimeOracle_BadParameterValidation(t *testing.T) {
	oracle := NewSimpleTimeOracle("localhost", 8081, locker.NewMemoryLocker())

	cases := []struct {
		name      string
		form      url.Values
		path      string
		wantCode  int
		wantError string
	}{
		{
			name:      "missing key in lock API",
			form:      url.Values{"id": {"id1"}, "duration": {"10"}},
			path:      oracle.baseURL + "/lock",
			wantCode:  http.StatusBadRequest,
			wantError: "Missing parameters",
		},
		{
			name:      "invalid duration in lock API",
			form:      url.Values{"key": {"key"}, "id": {"id1"}, "duration": {"something"}},
			path:      oracle.baseURL + "/lock",
			wantCode:  http.StatusBadRequest,
			wantError: "Invalid duration",
		},
		{
			name:      "missing id in unlock API",
			form:      url.Values{"key": {"key"}},
			path:      oracle.baseURL + "/unlock",
			wantCode:  http.StatusBadRequest,
			wantError: "Missing parameters",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tc.path+"?"+tc.form.Encode(), nil)
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/lock", oracle.serveLock).Methods("GET")
			router.HandleFunc("/unlock", oracle.serveUnlock).Methods("GET")

			router.ServeHTTP(recorder, req)

			resp := recorder.Result()

			// What we do here is verify the HTTP status and the error message received in case of bad or missing parameters
			if resp.StatusCode != tc.wantCode {
				t.Errorf("Got status %d, wanted status %d", resp.StatusCode, tc.wantCode)
			}

			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)

			if !strings.Contains(bodyString, tc.wantError) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tc.wantError, bodyString)
			}
		})
	}
}
