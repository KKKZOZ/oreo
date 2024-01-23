package couchdb

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"
)

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
			Address: "http://localhost:5984",
			DBName:  "dbname",
		}
	}
	if config.Address == "" {
		config.Address = "http://localhost:5984"
	}

	if config.se == nil {
		config.se = serializer.NewJSONSerializer()
	}

	client, _ := kivik.New("couch", config.Address)
	return &CouchDBConnection{
		client:  client,
		Address: config.Address,
		se:      config.se,
	}
}

// Connect establishes a connection to the CouchDB server and selects database
func (r *CouchDBConnection) Connect() error {
	db := r.client.DB(r.config.DBName)
	if db.Err() != nil {
		return db.Err()
	}
	r.db = db
	return nil
}

func (r *CouchDBConnection) GetItem(key string) (txn.DataItem2, error) {
	row := r.db.Get(context.Background(), key)
	var value txn.DataItem2
	err := row.ScanDoc(&value)
	if err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			return txn.DataItem2{}, txn.KeyNotFound
		}
		// For all other errors, return as is
		return txn.DataItem2{}, err
	}
	return value, nil
}

func (r *CouchDBConnection) PutItem(key string, value txn.DataItem2) error {
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
func (r *CouchDBConnection) ConditionalUpdate(key string, value txn.DataItem2) error {
	// Dummy function, needs to be appropriately implemented
	return nil
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
