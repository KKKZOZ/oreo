package main

import (
	mongoDB "benchmark/db/mongo"
	"benchmark/db/oreo"
	"benchmark/db/redis"
	"benchmark/ycsb"
	"context"
	"sync"
	"time"

	"github.com/kkkzoz/oreo/pkg/datastore/couchdb"
	mongoCo "github.com/kkkzoz/oreo/pkg/datastore/mongo"
	redisCo "github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/txn"
	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RedisCreator() (ycsb.DBCreator, error) {
	rdb1 := goredis.NewClient(&goredis.Options{
		Addr:     RedisDBAddr,
		Password: "@ljy123456",
	})

	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 15; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rdb1.Get(context.Background(), "1")
		}()
	}
	wg.Wait()

	return &redis.RedisCreator{RdbList: []*goredis.Client{rdb1}}, nil
}

func MongoCreator() (ycsb.DBCreator, error) {
	clientOptions := options.Client().ApplyURI(MongoDBAddr)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "admin",
	})
	context1, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(context1, clientOptions)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &mongoDB.MongoCreator{Client: client}, nil
}

func OreoRedisCreator(isRemote bool) (ycsb.DBCreator, error) {
	redisConn1 := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address:  OreoRedisAddr,
		Password: "@ljy123456",
		PoolSize: 100,
	})

	redisConn1.Connect()
	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			redisConn1.Get("1")
		}()
	}
	wg.Wait()
	return &oreo.OreoRedisCreator{
		ConnList: []*redisCo.RedisConnection{
			redisConn1},
		IsRemote: isRemote,
	}, nil
}

func OreoMongoCreator() (ycsb.DBCreator, error) {
	mongoConn1 := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
		Address:        MongoDBAddr,
		DBName:         "oreo",
		CollectionName: "benchmark",
		Username:       "admin",
		Password:       "admin",
	})
	mongoConn2 := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
		Address:        MongoDBAddr,
		DBName:         "oreo",
		CollectionName: "benchmark",
		Username:       "admin",
		Password:       "admin",
	})

	mongoConn1.Connect()
	mongoConn2.Connect()
	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mongoConn1.Get("1")
			mongoConn2.Get("1")
		}()
	}
	wg.Wait()
	return &oreo.OreoMongoCreator{
		ConnList: []*mongoCo.MongoConnection{
			mongoConn1, mongoConn2}}, nil
}

func OreoCouchCreator() (ycsb.DBCreator, error) {
	couchConn1 := couchdb.NewCouchDBConnection(&couchdb.ConnectionOptions{
		Address:  OreoCouchDBAddr,
		DBName:   "oreo",
		Username: "admin",
		Password: "password",
	})
	couchConn1.Connect()

	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			couchConn1.Get("1")
		}()
	}
	wg.Wait()

	return &oreo.OreoCouchCreator{
		ConnList: []*couchdb.CouchDBConnection{
			couchConn1}}, nil

}

func OreoCreator() (ycsb.DBCreator, error) {
	redisConn := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address: OreoRedisAddr,
	})
	mongoConn := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
		Address:        MongoDBAddr,
		DBName:         "oreo",
		CollectionName: "benchmark",
		Username:       "admin",
		Password:       "admin",
	})
	mongoConn.Connect()

	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 15; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			redisConn.Get("1")
			mongoConn.Get("1")
		}()
	}
	wg.Wait()

	connMap := map[string]txn.Connector{
		"redis": redisConn,
		"mongo": mongoConn,
	}
	return &oreo.OreoCreator{
		ConnMap:             connMap,
		GlobalDatastoreName: "redis",
	}, nil
}
