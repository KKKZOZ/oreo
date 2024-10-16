package couchdb

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-kivik/kivik/v4"
	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"

	_ "github.com/go-kivik/kivik/v4/couchdb"
)

var _ txn.Connector = (*CouchDBConnection)(nil)

type CouchDBConnection struct {
	client       *kivik.Client
	db           *kivik.DB
	Address      string
	config       ConnectionOptions
	hasConnected bool
}

type ConnectionOptions struct {
	Address  string
	Username string
	Password string
	DBName   string
}

func NewCouchDBConnection(config *ConnectionOptions) *CouchDBConnection {
	if config == nil {
		config = &ConnectionOptions{
			Address:  "http://admin:password@localhost:5984",
			DBName:   "oreo",
			Username: "admin",
			Password: "password",
		}
	}
	if config.Address == "" {
		config.Address = "http://admin:password@localhost:5984"
	}

	client, _ := kivik.New("couch", config.Address)

	// Set the basic authorization header

	return &CouchDBConnection{
		client:       client,
		Address:      config.Address,
		config:       *config,
		hasConnected: false,
	}
}

// Connect establishes a connection to the CouchDB server and selects database
func (r *CouchDBConnection) Connect() error {
	err := r.client.CreateDB(context.Background(), r.config.DBName, nil)
	// if the error is not 'PreconditionFailed' which means the DB already exists, return the error.
	if err != nil && kivik.HTTPStatus(err) != http.StatusPreconditionFailed {
		return err
	}

	db := r.client.DB(r.config.DBName)
	if dbErr := db.Err(); dbErr != nil {
		return dbErr
	}
	r.db = db
	r.hasConnected = true
	return nil
}

func (r *CouchDBConnection) GetItem(key string) (txn.DataItem, error) {
	if !r.hasConnected {
		return &CouchDBItem{}, fmt.Errorf("not connected to CouchDB")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	row := r.db.Get(context.Background(), key)
	var value CouchDBItem
	err := row.ScanDoc(&value)
	if err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			return &CouchDBItem{}, errors.New(txn.KeyNotFound)
		}
		// For all other errors, return as is
		return &CouchDBItem{}, err
	}
	return &value, nil
}

func (r *CouchDBConnection) PutItem(key string, value txn.DataItem) (string, error) {
	if !r.hasConnected {
		return "", fmt.Errorf("not connected to CouchDB")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	rev, err := r.db.Put(context.Background(), key, value, nil)
	if err != nil {
		return "", err
	}
	return rev, nil
}

func (r *CouchDBConnection) ConditionalUpdate(key string, value txn.DataItem, doCreate bool) (string, error) {
	if !r.hasConnected {
		return "", fmt.Errorf("not connected to CouchDB")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	var existing CouchDBItem
	err := r.db.Get(context.Background(), key).ScanDoc(&existing)

	if doCreate {
		if err != nil {
			if kivik.HTTPStatus(err) == http.StatusNotFound {
				// If the document doesn't exist, create it
				newVer, err := r.db.Put(context.Background(), key, value)
				if err != nil {
					return "", err
				}
				return newVer, nil
			}
			// For all other errors, return as is
			return "", err
		}
		return "", errors.New(txn.VersionMismatch)
	}

	if err != nil {
		return "", errors.New(txn.VersionMismatch)
	}

	// Update the document
	newVer, err := r.db.Put(context.Background(), key, value)
	if err != nil {
		return "", txn.VersionMismatch
	}

	return newVer, nil
}

func (r *CouchDBConnection) ConditionalCommit(key string, version string) (string, error) {
	if !r.hasConnected {
		return "", fmt.Errorf("not connected to CouchDB")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	var existing CouchDBItem
	err := r.db.Get(context.Background(), key).ScanDoc(&existing)

	if err != nil {
		return "", errors.New(txn.VersionMismatch)
	}

	existing.SetTxnState(config.COMMITTED)
	// Update the document
	newVer, err := r.db.Put(context.Background(), key, existing)
	if err != nil {
		return "", txn.VersionMismatch
	}

	return newVer, nil
}

func (r *CouchDBConnection) AtomicCreate(name string, value any) (string, error) {
	if !r.hasConnected {
		return "", fmt.Errorf("not connected to CouchDB")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	value = map[string]interface{}{
		"value": util.ToString(value),
	}

	_, err := r.db.Put(context.Background(), name, value)
	if err != nil {
		oldValue, _ := r.Get(name)
		return oldValue, errors.New(txn.KeyExists)
	}
	return "", nil
}

// Retrieve the value associated with the given key
func (r *CouchDBConnection) Get(name string) (string, error) {
	if !r.hasConnected {
		return "", fmt.Errorf("not connected to CouchDB")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	row := r.db.Get(context.Background(), name)
	var value map[string]string
	if err := row.ScanDoc(&value); err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			return "", errors.New(txn.KeyNotFound)
		}
		return "", err
	}
	return value["value"], nil
}

// Store the given value with the specified name (key)
func (r *CouchDBConnection) Put(name string, value interface{}) error {
	if !r.hasConnected {
		return fmt.Errorf("not connected to CouchDB")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	if _, ok := value.(string); ok {
		value = map[string]interface{}{
			"value": value,
		}
	}
	if _, ok := value.(config.State); ok {
		value = map[string]interface{}{
			"value": util.ToString(value),
		}
	}

	_, err := r.db.Put(context.Background(), name, value)
	if err != nil {
		return err
	}
	return nil
}

// Delete the specified key
func (r *CouchDBConnection) Delete(name string) error {
	if !r.hasConnected {
		return fmt.Errorf("not connected to CouchDB")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	type Item struct {
		Rev string `json:"_rev,omitempty"`
	}

	row := r.db.Get(context.Background(), name)
	var rev Item

	if err := row.ScanDoc(&rev); err != nil {
		return err
	}
	_, err := r.db.Delete(context.Background(), name, rev.Rev)
	if err != nil {
		return err
	}
	return nil
}
