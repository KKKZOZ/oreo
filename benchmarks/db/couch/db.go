package couch

import (
	"benchmark/ycsb"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-kivik/kivik/v4"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
)

var _ ycsb.DBCreator = (*CouchDBCreator)(nil)

type CouchDBCreator struct {
	Client *kivik.Client
}

func (cc *CouchDBCreator) Create() (ycsb.DB, error) {
	return NewCouchDB(cc.Client), nil
}

var _ ycsb.DB = (*CouchDB)(nil)

type CouchDB struct {
	Client *kivik.Client
	db     *kivik.DB
}

func NewCouchDB(client *kivik.Client) *CouchDB {
	err := client.CreateDB(context.Background(), "oreo", nil)
	// if the error is not 'PreconditionFailed' which means the DB already exists, return the error.
	if err != nil && kivik.HTTPStatus(err) != http.StatusPreconditionFailed {
		panic("Panic: " + err.Error())
	}
	db := client.DB("oreo")
	return &CouchDB{
		Client: client,
		db:     db,
	}
}

func (r *CouchDB) Close() error {
	return nil
}

func (r *CouchDB) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return ctx
}

func (r *CouchDB) CleanupThread(ctx context.Context) {
}

func (r *CouchDB) Read(ctx context.Context, table string, key string) (string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	row := r.db.Get(ctx, key)
	var doc map[string]interface{}
	if err := row.ScanDoc(&doc); err != nil {
		return "", err
	}

	value, ok := doc["value"].(string)
	if !ok {
		return "", errors.New("value is not a string")
	}

	return value, nil
}

func (r *CouchDB) Update(ctx context.Context, table string, key string, value string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	// Retrieve the existing document to get the _rev field
	row := r.db.Get(ctx, key)
	var doc map[string]interface{}
	if err := row.ScanDoc(&doc); err != nil {
		return err
	}

	doc["value"] = value

	_, err := r.db.Put(ctx, key, doc)
	return err
}

func (r *CouchDB) Insert(ctx context.Context, table string, key string, value string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	doc := map[string]interface{}{
		"_id":   key,
		"value": value,
	}

	_, err := r.db.Put(ctx, key, doc)
	return err
}

func (r *CouchDB) Delete(ctx context.Context, table string, key string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	// Retrieve the existing document to get the _rev field
	row := r.db.Get(ctx, key)
	var doc map[string]interface{}
	if err := row.ScanDoc(&doc); err != nil {
		return err
	}

	rev, ok := doc["_rev"].(string)
	if !ok {
		return errors.New("_rev field is missing or not a string")
	}

	_, err := r.db.Delete(ctx, key, rev)
	return err
}
