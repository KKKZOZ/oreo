package main

import (
	"fmt"
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/network"
	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

func main() {
	client, err := network.NewClient(":9000")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Waiting for the connection of executors...\n")
	time.Sleep(3 * time.Second)

	oracle := timesource.NewGlobalTimeSource("http://localhost:8010")

	redis_conn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  "localhost:6379",
		Password: "kkkzoz",
	})

	mongo1_conn := mongo.NewMongoConnection(&mongo.ConnectionOptions{
		Address:        "mongodb://localhost:27017",
		Username:       "admin",
		Password:       "password",
		DBName:         "oreo",
		CollectionName: "records",
	})

	redis_datastore := redis.NewRedisDatastore("Redis", redis_conn)
	mongo1_datastore := mongo.NewMongoDatastore("MongoDB1", mongo1_conn)

	write_txn := txn.NewTransactionWithRemote(client, oracle)
	_ = write_txn.AddDatastores(redis_datastore, mongo1_datastore)

	_ = write_txn.Start()
	_ = write_txn.Write("Redis", "key1", "value1")
	_ = write_txn.Write("MongoDB1", "key2", "value2")

	err = write_txn.Commit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Write transaction committed successfully.\n")

	read_txn := txn.NewTransactionWithRemote(client, oracle)
	_ = read_txn.AddDatastores(redis_datastore, mongo1_datastore)

	_ = read_txn.Start()

	var value1, value2 string

	err = read_txn.Read("Redis", "key1", &value1)
	if err != nil {
		panic(err)
	}
	err = read_txn.Read("MongoDB1", "key2", &value2)
	if err != nil {
		panic(err)
	}
	err = read_txn.Commit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Read values: %s, %s\n", value1, value2)
}
