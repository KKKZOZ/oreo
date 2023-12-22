package locker

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// HttpLocker represents a locker that uses HTTP requests to interact with an Oracle service.
type HttpLocker struct {
	oracleURL string
}

// NewHttpLocker creates a new instance of HttpLocker with the specified oracleURL.
// The oracleURL parameter is the URL of the Oracle server used for locking.
func NewHttpLocker(oracleURL string) *HttpLocker {
	return &HttpLocker{
		oracleURL: oracleURL,
	}
}

// Lock locks the specified key with the given ID for the specified duration.
// It sends an HTTP GET request to the oracleURL with the key, ID, and duration as query parameters.
// If the lock request fails, it returns an error indicating the failure.
func (l *HttpLocker) Lock(key string, id string, holdDuration time.Duration) error {
	data := url.Values{}
	data.Set("key", key)
	data.Set("id", id)
	data.Set("duration", strconv.Itoa(int(holdDuration)))

	_, err := http.Get(l.oracleURL + "/lock?" + data.Encode())
	if err != nil {
		return errors.New("failed to lock")
	}
	return nil
}

// Unlock unlocks the specified key with the given ID using HTTP.
// It sends a GET request to the oracleURL with the key and ID as query parameters.
// If the request is successful, it returns nil. Otherwise, it returns an error.
func (l *HttpLocker) Unlock(key string, id string) error {
	data := url.Values{}
	data.Set("key", key)
	data.Set("id", id)

	_, err := http.Get(l.oracleURL + "/unlock?" + data.Encode())
	if err != nil {
		return err
	}
	return nil
}
