package main

import (
	"context"
	"fmt"
	"time"

	mong "github.com/kkkzoz/oreo/pkg/datastore/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const uri = "mongodb://localhost:27017"

func main() {
	opts := options.Client().ApplyURI(uri)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll := client.Database("oreo").Collection("test")

	now := time.Now()
	mongoItem := &mong.MongoItem{
		MKey:          "test_key",
		MValue:        "test_value",
		MGroupKeyList: "1",
		MTxnState:     1,
		MTValid:       now.Add(-3 * time.Second).UnixMicro(),
		MTLease:       now.Add(-2 * time.Second),
		MPrev:         "",
		MIsDeleted:    false,
		MVersion:      "2",
	}

	keyFilter := bson.D{{Key: "TxnId", Value: "1"}}
	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	update := bson.D{{Key: "$set", Value: mongoItem}}

	_, err = coll.UpdateOne(context.Background(), keyFilter, update, &opt)
	if err != nil {
		panic(err)
	}

	filter := bson.D{{Key: "TxnId", Value: "1"}}

	var result mong.MongoItem
	err = coll.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("Fail to find")
			return
		}
		panic(err)
	}

	if result.Equal(mongoItem) {
		fmt.Println("Success")
	} else {
		fmt.Println("Fail")
		fmt.Printf("got\n %v\n expected\n %v\n", result, mongoItem)
	}
}
