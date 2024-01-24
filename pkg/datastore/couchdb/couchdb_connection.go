package couchdb

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"

	_ "github.com/go-kivik/kivik/v4/couchdb"
)

var _ txn.Connector = (*CouchDBConnection)(nil)

type CouchDBConnection struct {
	client  *kivik.Client
	db      *kivik.DB
	Address string
	config  ConnectionOptions
	se      serializer.Serializer
}

type ConnectionOptions struct {
	Address  string
	Username string
	Password string
	se       serializer.Serializer
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
		config.Address = "http://localhost:5984"
	}

	if config.se == nil {
		config.se = serializer.NewJSONSerializer()
	}

	client, _ := kivik.New("couch", config.Address)

	// Set the basic authorization header

	return &CouchDBConnection{
		client:  client,
		Address: config.Address,
		config:  *config,
		se:      config.se,
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
	return nil
}

func (r *CouchDBConnection) GetItem(key string) (txn.DataItem, error) {
	row := r.db.Get(context.Background(), key)
	var value CouchDBItem
	err := row.ScanDoc(&value)
	if err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			return &CouchDBItem{}, txn.KeyNotFound
		}
		// For all other errors, return as is
		return &CouchDBItem{}, err
	}
	return &value, nil
}

func (r *CouchDBConnection) PutItem(key string, value txn.DataItem) error {
	_, err := r.db.Put(context.Background(), key, value)
	if err != nil {
		return err
	}
	return nil
}

/*
	Conditional update in this style isn't the same as in CouchDB due to their different concurrency controls.

Consider using CouchDB's conflict resolution system by checking for conflicts on Get(), or apply updates using previous _rev as a reference to ensure atomicity
*/
func (r *CouchDBConnection) ConditionalUpdate(key string, value txn.DataItem, doCreate bool) (string, error) {

	if doCreate {

	}

	// Get the existing document
	row := r.db.Get(context.Background(), key)
	var existing Item
	if err = row.ScanDoc(&existing); err != nil {
		return err
	}

	// Check if the version is the same
	if existing.Version != item.Version {
		return errors.New("version mismatch")
	}

	// Use the existing document revision
	item.Rev = existing.Rev

	// Update the document
	_, err = r.db.Put(context.Background(), key, item)

	return err
}

// Retrieve the value associated with the given key
func (r *CouchDBConnection) Get(name string) (string, error) {
	row := r.db.Get(context.Background(), name)
	var value string
	if err := row.ScanDoc(&value); err != nil {
		return "", err
	}
	return value, nil
}

// Store the given value with the specified name (key)
func (r *CouchDBConnection) Put(name string, value interface{}) error {
	_, err := r.db.Put(context.Background(), name, value)
	if err != nil {
		return err
	}
	return nil
}

// Delete the specified key
func (r *CouchDBConnection) Delete(name string) error {
	row := r.db.Get(context.Background(), name)
	var rev string

	if err := row.ScanDoc(&rev); err != nil {
		return err
	}
	_, err := r.db.Delete(context.Background(), name, rev)
	if err != nil {
		return err
	}
	return nil
}
