package mongo

import (
	"context"
	"time"

	"benchmark/ycsb"
	"github.com/kkkzoz/oreo/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ ycsb.DBCreator = (*MongoCreator)(nil)

type MongoCreator struct {
	Client *mongo.Client
}

func (mc *MongoCreator) Create() (ycsb.DB, error) {
	return NewMongo(mc.Client), nil
}

var _ ycsb.DB = (*Mongo)(nil)

type MyDocument struct {
	Key   string `bson:"_id,omitempty"`
	Value string `bson:"value"`
}

type Mongo struct {
	Client *mongo.Client
	coll   *mongo.Collection
}

func NewMongo(client *mongo.Client) *Mongo {
	db := client.Database("oreo")
	coll := db.Collection("benchmark")
	return &Mongo{
		Client: client,
		coll:   coll,
	}
}

func (r *Mongo) Close() error {
	return nil
}

func (r *Mongo) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return ctx
}

func (r *Mongo) CleanupThread(ctx context.Context) {
}

func (r *Mongo) Read(ctx context.Context, table string, key string) (string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	var doc MyDocument
	err := r.coll.FindOne(ctx, bson.M{"_id": key}).Decode(&doc)
	if err != nil {
		return "", err
	}

	return doc.Value, nil
}

func (r *Mongo) Update(ctx context.Context, table string, key string, value string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	_, err := r.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": key},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "Value", Value: value},
			}},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *Mongo) Insert(ctx context.Context, table string, key string, value string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	_, err := r.coll.UpdateOne(
		context.Background(),
		bson.M{"_id": key},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "Value", Value: value},
			}},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *Mongo) Delete(ctx context.Context, table string, key string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.ConnAdditionalLatency)
	}

	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": key})
	return err
}
