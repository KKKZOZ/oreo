package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/go-errors/errors"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/txn"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ txn.Connector = (*MongoConnection)(nil)

const defaultMongoTimeout = 5000 * time.Millisecond

type KeyValueItem struct {
	Key   string `bson:"_id"`
	Value string `bson:"Value"`
}

type MongoConnection struct {
	client       *mongo.Client
	db           *mongo.Database
	coll         *mongo.Collection
	Address      string
	config       ConnectionOptions
	hasConnected bool
}

type ConnectionOptions struct {
	Address        string
	Username       string
	Password       string
	DBName         string
	CollectionName string
}

var defaultOptions = ConnectionOptions{
	Address:        "mongodb://localhost:27017",
	Username:       "",
	Password:       "",
	DBName:         "oreo",
	CollectionName: "records",
}

// NewMongoConnection creates a new MongoDB connection.
func NewMongoConnection(config *ConnectionOptions) *MongoConnection {
	// Start with a copy of the default configuration.
	finalConfig := defaultOptions

	// If the user provided a configuration, layer their values on top.
	// This avoids mutating the original 'config' object.
	if config != nil {
		if config.Address != "" {
			finalConfig.Address = config.Address
		}
		if config.Username != "" {
			finalConfig.Username = config.Username
		}
		if config.Password != "" {
			finalConfig.Password = config.Password
		}
		if config.DBName != "" {
			finalConfig.DBName = config.DBName
		}
		if config.CollectionName != "" {
			finalConfig.CollectionName = config.CollectionName
		}
	}

	conn := &MongoConnection{
		Address:      finalConfig.Address,
		config:       finalConfig,
		hasConnected: false,
	}

	return conn
}

// Connect establishes a connection to the MongoDB server.
func (m *MongoConnection) Connect() error {
	if m.hasConnected {
		return nil
	}

	clientOptions := options.Client().ApplyURI(m.Address)

	// clientOptions.SetConnectTimeout(1 * time.Second)

	if m.config.Username != "" && m.config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: m.config.Username,
			Password: m.config.Password,
		})
	}
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	m.db = client.Database(m.config.DBName)
	m.coll = m.db.Collection(m.config.CollectionName)
	m.hasConnected = true
	return nil
}

// Close disconnects from the MongoDB server.
func (m *MongoConnection) Close() error {
	if !m.hasConnected {
		return nil
	}
	return m.client.Disconnect(context.Background())
}

// GetItem retrieves a structured transaction item from MongoDB.
// It returns a txn.DataItem, which represents a full document with transaction metadata.
func (m *MongoConnection) GetItem(key string) (txn.DataItem, error) {
	if !m.hasConnected {
		return &MongoItem{}, errors.Errorf("not connected to MongoDB")
	}

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()

	var item MongoItem
	err := m.coll.FindOne(ctx, bson.M{"_id": key}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &MongoItem{}, errors.New(txn.KeyNotFound)
		}
		return &MongoItem{}, err
	}
	return &item, nil
}

// PutItem inserts or updates an item in MongoDB.
func (m *MongoConnection) PutItem(key string, value txn.DataItem) (string, error) {
	if !m.hasConnected {
		return "", errors.Errorf("not connected to MongoDB")
	}

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()

	_, err := m.coll.UpdateOne(
		ctx,
		bson.M{"_id": key},
		bson.D{
			{Key: "$set", Value: value},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return "", err
	}
	return value.Version(), nil
}

// ConditionalUpdate atomically updates an item if the version matches.
// If doCreat is true, it will create the item if it does not exist.
func (m *MongoConnection) ConditionalUpdate(
	key string,
	value txn.DataItem,
	doCreat bool,
) (string, error) {
	if !m.hasConnected {
		return "", errors.Errorf("not connected to MongoDB")
	}

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	if doCreat {
		return m.atomicCreateMongoItem(key, value)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()

	newVer := util.AddToString(value.Version(), 1)

	filter := bson.M{"_id": key, "Version": value.Version()}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "Value", Value: value.Value()},
			{Key: "GroupKeyList", Value: value.GroupKeyList()},
			{Key: "TxnState", Value: value.TxnState()},
			{Key: "TValid", Value: value.TValid()},
			{Key: "TLease", Value: value.TLease().Format(time.RFC3339Nano)},
			{Key: "Prev", Value: value.Prev()},
			{Key: "LinkedLen", Value: value.LinkedLen()},
			{Key: "IsDeleted", Value: value.IsDeleted()},
			{Key: "Version", Value: newVer},
		}},
	}
	after := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	var updatedItem MongoItem
	err := m.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedItem)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New(txn.VersionMismatch)
		}
		return "", err
	}

	if updatedItem.Version() != newVer {
		return "", errors.New(txn.VersionMismatch)
	}

	return newVer, nil
}

// ConditionalCommit atomically commits a transaction if the version matches.
func (m *MongoConnection) ConditionalCommit(
	key string,
	version string,
	tCommit int64,
) (string, error) {
	if !m.hasConnected {
		return "", errors.Errorf("not connected to MongoDB")
	}

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()

	newVer := util.AddToString(version, 1)

	filter := bson.M{"_id": key, "Version": version}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "TxnState", Value: config.COMMITTED},
			{Key: "Version", Value: newVer},
			{Key: "TValid", Value: tCommit},
		}},
	}
	after := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	var updatedItem MongoItem
	err := m.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedItem)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New(txn.VersionMismatch)
		}
		return "", err
	}
	if updatedItem.Version() != newVer {
		return "", errors.New(txn.VersionMismatch)
	}
	return newVer, nil
}

// AtomicCreate creates a key-value pair if the key does not already exist.
func (m *MongoConnection) AtomicCreate(key string, value any) (string, error) {
	if !m.hasConnected {
		return "", errors.Errorf("not connected to MongoDB")
	}

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()

	filter := bson.M{"_id": key}
	var result KeyValueItem
	err := m.coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// we can safely create the item
			str := util.ToString(value)
			_, err := m.coll.InsertOne(ctx, bson.D{
				{Key: "_id", Value: key},
				{Key: "Value", Value: str},
			})
			if err != nil {
				return "", err
			}
			return "", nil
		}
		return "", err
	}
	// the key already exists, return an error and the old state
	return result.Value, errors.New(txn.KeyExists)
}

func (m *MongoConnection) atomicCreateMongoItem(key string, value txn.DataItem) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()

	filter := bson.M{"_id": key}
	var result MongoItem
	err := m.coll.FindOne(ctx, filter).Decode(&result)

	newVer := util.AddToString(value.Version(), 1)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			_, err := m.coll.InsertOne(ctx, bson.D{
				{Key: "_id", Value: key},
				{Key: "Value", Value: value.Value()},
				{Key: "GroupKeyList", Value: value.GroupKeyList()},
				{Key: "TxnState", Value: value.TxnState()},
				{Key: "TValid", Value: value.TValid()},
				{Key: "TLease", Value: value.TLease().Format(time.RFC3339Nano)},
				{Key: "Prev", Value: value.Prev()},
				{Key: "LinkedLen", Value: value.LinkedLen()},
				{Key: "IsDeleted", Value: value.IsDeleted()},
				{Key: "Version", Value: newVer},
			})
			if err != nil {
				return "", err
			}
			return newVer, nil
		}
		return "", err
	}

	return "", errors.New(txn.VersionMismatch)
}

// Get retrieves a simple string value from a key-value pair document.
// This is a general-purpose getter, distinct from GetItem, which retrieves a structured txn.DataItem.
func (m *MongoConnection) Get(key string) (string, error) {
	if !m.hasConnected {
		return "", fmt.Errorf("not connected to MongoDB")
	}

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()

	var result KeyValueItem
	err := m.coll.FindOne(ctx, bson.M{"_id": key}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New(txn.KeyNotFound)
		}
		return "", err
	}
	return result.Value, nil
}

// Put sets the value for a given key.
func (m *MongoConnection) Put(key string, value any) error {
	if !m.hasConnected {
		return fmt.Errorf("not connected to MongoDB")
	}

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()

	// str := util.ToString(value)

	_, err := m.coll.UpdateOne(
		ctx,
		bson.M{"_id": key},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "Value", Value: value},
			}},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a key-value pair.
func (m *MongoConnection) Delete(key string) error {
	if !m.hasConnected {
		return fmt.Errorf("not connected to MongoDB")
	}

	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultMongoTimeout)
	defer cancel()

	_, err := m.coll.DeleteOne(ctx, bson.M{"_id": key})
	if err != nil {
		return err
	}
	return nil
}
