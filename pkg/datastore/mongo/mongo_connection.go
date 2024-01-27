package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/go-errors/errors"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/txn"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ txn.Connector = (*MongoConnection)(nil)

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

// NewMongoConnection creates a new MongoDB connection using the provided configuration options.
// If the config parameter is nil, default values will be used.
// The MongoDB connection is established using the specified address, username, password, and database name.
// The address format should be in the form "mongodb://host:port".
// The se parameter is used for data serialization and deserialization.
// If se is nil, a default JSON serializer will be used.
// Returns a pointer to the created MongoConnection.
func NewMongoConnection(config *ConnectionOptions) *MongoConnection {
	if config == nil {
		config = &ConnectionOptions{
			Address:        "mongodb://localhost:27017",
			Username:       "",
			Password:       "",
			DBName:         "oreo",
			CollectionName: "records",
		}
	}
	if config.Address == "" {
		config.Address = "mongodb://localhost:27017"
	}

	conn := &MongoConnection{
		Address:      config.Address,
		config:       *config,
		hasConnected: false,
	}

	return conn
}

// Connect establishes a connection to the MongoDB server.
// It returns an error if the connection cannot be established.
func (m *MongoConnection) Connect() error {
	clientOptions := options.Client().ApplyURI(m.Address)
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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
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

// Close closes the MongoDB connection.
// It's important to defer this function after creating a new connection.
func (m *MongoConnection) Close() error {
	if !m.hasConnected {
		return nil
	}
	return m.client.Disconnect(context.Background())
}

// GetItem retrieves a txn.DataItem from the MongoDB database based on the specified key.
// If the key is not found, it returns an empty txn.DataItem and an error.
func (m *MongoConnection) GetItem(key string) (txn.DataItem, error) {
	if !m.hasConnected {
		return &MongoItem{}, errors.Errorf("not connected to MongoDB")
	}
	var item MongoItem
	err := m.coll.FindOne(context.Background(), bson.M{"_id": key}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &MongoItem{}, errors.New(txn.KeyNotFound)
		}
		return &MongoItem{}, err
	}
	return &item, nil
}

// PutItem puts an item into the MongoDB database with the specified key and value.
// The function returns an error if there was a problem executing the MongoDB commands.
func (m *MongoConnection) PutItem(key string, value txn.DataItem) (string, error) {
	if !m.hasConnected {
		return "", errors.Errorf("not connected to MongoDB")
	}

	_, err := m.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": key},
		bson.D{
			{Key: "$set", Value: value},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return "", err
	}
	return "", nil
}

// ConditionalUpdate updates the value of a Mongo item if the version matches the provided value.
// It takes a key string and a txn.DataItem value as parameters.
// If the item's version does not match, it returns a version mismatch error.
// Note: if the previous version of the item is not found, it will return a key not found error.
// Otherwise, it updates the item with the provided values and returns the updated item.
func (m *MongoConnection) ConditionalUpdate(key string, value txn.DataItem, doCreat bool) (string, error) {
	if !m.hasConnected {
		return "", errors.Errorf("not connected to MongoDB")
	}

	if doCreat {
		return m.atomicCreate(key, value)
	}

	newVer := util.AddToString(value.Version(), 1)

	filter := bson.M{"_id": key, "Version": value.Version()}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "Value", Value: value.Value()},
			{Key: "TxnId", Value: value.TxnId()},
			{Key: "TxnState", Value: value.TxnState()},
			{Key: "TValid", Value: value.TValid().Format(time.RFC3339Nano)},
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
	err := m.coll.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedItem)
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

func (m *MongoConnection) atomicCreate(key string, value txn.DataItem) (string, error) {
	filter := bson.M{"_id": key}

	var result MongoItem
	err := m.coll.FindOne(context.Background(), filter).Decode(&result)

	newVer := util.AddToString(value.Version(), 1)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			_, err := m.coll.InsertOne(context.Background(), bson.D{
				{Key: "_id", Value: key},
				{Key: "Value", Value: value.Value()},
				{Key: "TxnId", Value: value.TxnId()},
				{Key: "TxnState", Value: value.TxnState()},
				{Key: "TValid", Value: value.TValid().Format(time.RFC3339Nano)},
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

// Get retrieves the value associated with the given key from the MongoDB database.
// If the key is not found, it returns an empty string and an error indicating the key was not found.
// If an error occurs during the retrieval, it returns an empty string and the error.
// Otherwise, it returns the retrieved value and nil error.
func (m *MongoConnection) Get(key string) (string, error) {
	if !m.hasConnected {
		return "", fmt.Errorf("not connected to MongoDB")
	}
	var result struct {
		Key   string `bson:"_id"`
		Value string `bson:"Value"`
	}
	err := m.coll.FindOne(context.Background(), bson.M{"_id": key}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New(txn.KeyNotFound)
		}
		return "", err
	}
	return result.Value, nil
}

// Put stores the given value with the specified key in the MongoDB database.
// It will overwrite the value if the key already exists.
// It returns an error if the operation fails.
func (m *MongoConnection) Put(key string, value any) error {
	if !m.hasConnected {
		return fmt.Errorf("not connected to MongoDB")
	}

	str := util.ToString(value)

	_, err := m.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": key},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "Value", Value: str},
			}},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes the specified key from the MongoDB database.
// It allows for the deletion of a key that does not exist.
func (m *MongoConnection) Delete(key string) error {
	if !m.hasConnected {
		return fmt.Errorf("not connected to MongoDB")
	}
	_, err := m.coll.DeleteOne(context.Background(), bson.M{"_id": key})
	if err != nil {
		return err
	}
	return nil
}
