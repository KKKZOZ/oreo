package main

import (
	mongoDB "benchmark/db/mongo"
	"benchmark/db/oreo"
	"benchmark/db/redis"
	"benchmark/pkg/workload"
	"benchmark/ycsb"
	"context"
	"fmt"
	"log"
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

func RedisCreator(addr string) (ycsb.DBCreator, error) {
	rdb1 := goredis.NewClient(&goredis.Options{
		Addr:     addr,
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

func MongoCreator(addr string) (ycsb.DBCreator, error) {
	clientOptions := options.Client().ApplyURI(addr)
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

func NativeCreator(pattern string) (ycsb.DBCreator, error) {
	fmt.Printf("Creating native workload with pattern: %v\n", pattern)
	if pattern == "mm" {
		mongoCreator1, err1 := MongoCreator(MongoDBAddr1)
		mongoCreator2, err2 := MongoCreator(MongoDBAddr2)
		if err1 != nil || err2 != nil {
			log.Fatalf("Error when creating client: %v %v\n", err1, err2)
			return nil, nil
		}
		dbSetCreator := workload.DBSetCreator{
			CreatorMap: map[string]ycsb.DBCreator{
				"mongo1": mongoCreator1,
				"mongo2": mongoCreator2,
			},
		}

		return &dbSetCreator, nil
	}

	if pattern == "rm" {
		redisCreator, err1 := RedisCreator(RedisDBAddr)
		mongoCreator1, err2 := MongoCreator(MongoDBAddr1)
		if err1 != nil || err2 != nil {
			log.Fatalf("Error when creating client: %v %v\n", err1, err2)
			return nil, nil
		}
		dbSetCreator := workload.DBSetCreator{
			CreatorMap: map[string]ycsb.DBCreator{
				"redis1": redisCreator,
				"mongo1": mongoCreator1,
			},
		}

		return &dbSetCreator, nil
	}
	fmt.Printf("Unknown pattern: %v\n", pattern)
	return nil, nil
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
		IsRemote: isRemote,
		ConnList: []*redisCo.RedisConnection{
			redisConn1},
	}, nil
}

func OreoMongoCreator(isRemote bool) (ycsb.DBCreator, error) {
	mongoConn1 := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
		Address:        OreoMongoDBAddr1,
		DBName:         "oreo",
		CollectionName: "benchmark",
		Username:       "admin",
		Password:       "admin",
	})
	mongoConn2 := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
		Address:        OreoMongoDBAddr1,
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
		IsRemote: isRemote,
		ConnList: []*mongoCo.MongoConnection{
			mongoConn1, mongoConn2}}, nil
}

// TODO: Add isRemote logic
func OreoCouchCreator(isRemote bool) (ycsb.DBCreator, error) {
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

func OreoCreator(pattern string, isRemote bool) (ycsb.DBCreator, error) {

	if pattern == "mm" {
		mongoConn1 := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
			Address:        OreoMongoDBAddr1,
			DBName:         "oreo",
			CollectionName: "benchmark",
			Username:       "admin",
			Password:       "admin",
		})
		mongoConn2 := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
			Address:        OreoMongoDBAddr2,
			DBName:         "oreo",
			CollectionName: "benchmark",
			Username:       "admin",
			Password:       "admin",
		})
		mongoConn1.Connect()
		mongoConn2.Connect()

		// try to warm up the connection
		var wg sync.WaitGroup
		for i := 1; i <= 15; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				mongoConn1.Get("1")
				mongoConn2.Get("1")
			}()
		}
		wg.Wait()

		connMap := map[string]txn.Connector{
			"mongo1": mongoConn1,
			"mongo2": mongoConn2,
		}
		return &oreo.OreoCreator{
			IsRemote:            isRemote,
			ConnMap:             connMap,
			GlobalDatastoreName: "mongo1",
		}, nil
	}

	if pattern == "rm" {
		redisConn1 := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
			Address:  OreoRedisAddr,
			Password: "@ljy123456",
		})

		mongoConn1 := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
			Address:        OreoMongoDBAddr1,
			DBName:         "oreo",
			CollectionName: "benchmark",
			Username:       "admin",
			Password:       "admin",
		})
		redisConn1.Connect()
		mongoConn1.Connect()

		// try to warm up the connection
		var wg sync.WaitGroup
		for i := 1; i <= 15; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				redisConn1.Get("1")
				mongoConn1.Get("1")
			}()
		}
		wg.Wait()

		connMap := map[string]txn.Connector{
			"redis1": redisConn1,
			"mongo1": mongoConn1,
		}
		return &oreo.OreoCreator{
			IsRemote:            isRemote,
			ConnMap:             connMap,
			GlobalDatastoreName: "redis1",
		}, nil
	}

	return nil, nil
}
