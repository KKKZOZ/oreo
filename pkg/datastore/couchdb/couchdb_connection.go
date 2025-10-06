package couchdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/txn"
)

var _ txn.Connector = (*CouchDBConnection)(nil)

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        6000,
		MaxIdleConnsPerHost: 1000,
		MaxConnsPerHost:     1000,
		IdleConnTimeout:     90 * time.Second,
	},
}

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

var defaultOptions = ConnectionOptions{
	Address:  "http://admin:password@localhost:5984",
	DBName:   "oreo",
	Username: "admin",
	Password: "password",
}

// NewCouchDBConnection creates a new CouchDB connection.
func NewCouchDBConnection(config *ConnectionOptions) *CouchDBConnection {
	// Start with a copy of the default configuration.
	finalConfig := defaultOptions

	// If the user provided a configuration, layer their values on top.
	if config != nil {
		if config.Address != "" {
			finalConfig.Address = config.Address
		}
		if config.DBName != "" {
			finalConfig.DBName = config.DBName
		}
		if config.Username != "" {
			finalConfig.Username = config.Username
		}
		if config.Password != "" {
			finalConfig.Password = config.Password
		}
	}

	dsn := fmt.Sprintf(
		"http://%s:%s@%s/",
		finalConfig.Username,
		finalConfig.Password,
		finalConfig.Address,
	)

	client, _ := kivik.New("couch", dsn, couchdb.OptionHTTPClient(
		httpClient,
	))

	return &CouchDBConnection{
		client:       client,
		Address:      finalConfig.Address,
		config:       finalConfig,
		hasConnected: false,
	}
}

// Connect establishes a connection to the CouchDB server and selects the database.
func (r *CouchDBConnection) Connect() error {
	if r.hasConnected {
		return nil
	}

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

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			_ = r.db.Get(context.Background(), "test")
		}()
	}
	wg.Wait()

	return nil
}

// GetItem retrieves a structured transaction item from CouchDB.
// It returns a txn.DataItem, which represents a full document with transaction metadata.
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

// PutItem inserts or updates a transaction item in CouchDB.
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

// ConditionalUpdate atomically updates an item if the revision matches.
// If doCreate is true, it will create the item if it does not exist.
func (r *CouchDBConnection) ConditionalUpdate(
	key string,
	value txn.DataItem,
	doCreate bool,
) (string, error) {
	if !r.hasConnected {
		return "", fmt.Errorf("not connected to CouchDB")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	if doCreate {
		if value.Version() != "" {
			// 如果是创建模式但已有版本号，说明文档已存在
			return "", errors.New(txn.VersionMismatch)
		}
		// 创建模式，直接尝试创建文档
		newVer, err := r.db.Put(context.Background(), key, value)
		if err != nil {
			if kivik.HTTPStatus(err) == http.StatusConflict {
				return "", errors.New("key exists")
			}
			return "", err
		}
		return newVer, nil
	}

	// Update the document
	newVer, err := r.db.Put(context.Background(), key, value)
	if err != nil {
		if kivik.HTTPStatus(err) == http.StatusConflict {
			return "", errors.New(txn.VersionMismatch)
		}
		return "", err
	}

	return newVer, nil
}

// ConditionalCommit atomically commits a transaction if the revision matches.
func (r *CouchDBConnection) ConditionalCommit(
	key string,
	version string,
	tCommit int64,
) (string, error) {
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
	existing.SetTValid(tCommit)
	// Update the document
	newVer, err := r.db.Put(context.Background(), key, existing)
	if err != nil {
		return "", txn.VersionMismatch
	}

	return newVer, nil
}

// AtomicCreate creates a key-value pair if the key does not already exist.
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

// Get retrieves a simple string value from a document.
// This is a general-purpose getter, distinct from GetItem, which retrieves a structured txn.DataItem.
func (r *CouchDBConnection) Get(name string) (string, error) {
	if !r.hasConnected {
		return "", fmt.Errorf("not connected to CouchDB")
	}
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	row := r.db.Get(context.Background(), name)
	var value map[string]interface{}
	if err := row.ScanDoc(&value); err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			return "", errors.New(txn.KeyNotFound)
		}
		return "", err
	}

	// Check if it's wrapped in a "value" field (for simple string storage)
	if val, exists := value["value"]; exists {
		if strVal, ok := val.(string); ok {
			return strVal, nil
		}
		return fmt.Sprintf("%v", val), nil
	}

	// If not wrapped, it means the JSON was directly parsed by CouchDB
	// We need to reconstruct the original JSON by removing CouchDB metadata
	delete(value, "_id")
	delete(value, "_rev")

	// Re-serialize to JSON string
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// Put sets the value for a given key, creating or updating the document.
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

// Delete removes a document from CouchDB.
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
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			return nil
		}
		return err
	}
	_, err := r.db.Delete(context.Background(), name, rev.Rev)
	if err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			return nil
		}
		return err
	}
	return nil
}
