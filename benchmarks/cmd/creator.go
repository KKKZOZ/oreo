package main

import (
	"benchmark/db/couch"
	mongoDB "benchmark/db/mongo"
	"benchmark/db/oreo"
	"benchmark/db/redis"
	"benchmark/pkg/workload"
	"benchmark/ycsb"
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-kivik/kivik/v4"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	mongoCo "github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	redisCo "github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NativeRealisticCreator(workloadType string) (ycsb.DBCreator, error) {
	if workloadType == "iot" {
		rdb1, err := NewRedisClient(OreoKVRocksAddr)
		if err != nil {
			return nil, err
		}
		kvDB := redis.NewRedis(rdb1)

		mongoClient1, err := NewMongoDBClient(OreoMongoDBAddr1)
		if err != nil {
			return nil, err
		}
		mongoDB1 := mongoDB.NewMongo(mongoClient1)

		return &oreo.NativeRealisticCreator{
			ConnMap: map[string]ycsb.DB{
				"KVRocks": kvDB,
				"MongoDB": mongoDB1,
			},
		}, nil
	}
	if workloadType == "social" {
		rdb1, err := NewRedisClient(OreoRedisAddr)
		if err != nil {
			return nil, err
		}
		redisDB := redis.NewRedis(rdb1)

		mongoClient1, err := NewMongoDBClient(OreoMongoDBAddr1)
		if err != nil {
			return nil, err
		}
		mongoDB := mongoDB.NewMongo(mongoClient1)

		couchClient, err := NewCouchClient(OreoCouchDBAddr)
		if err != nil {
			fmt.Printf("Error when creating couch client: %v\n", err)
			return nil, err
		}
		couchDB := couch.NewCouchDB(couchClient)

		return &oreo.NativeRealisticCreator{
			ConnMap: map[string]ycsb.DB{
				"Redis":   redisDB,
				"MongoDB": mongoDB,
				"CouchDB": couchDB,
			},
		}, nil
	}
	if workloadType == "order" {
		rdb1, err := NewRedisClient(OreoRedisAddr)
		if err != nil {
			return nil, err
		}
		redisDB := redis.NewRedis(rdb1)
		rdb2, err := NewRedisClient(OreoKVRocksAddr)
		if err != nil {
			return nil, err
		}
		kvDB := redis.NewRedis(rdb2)

		mongoClient1, err := NewMongoDBClient(OreoMongoDBAddr1)
		if err != nil {
			return nil, err
		}
		mongoDB := mongoDB.NewMongo(mongoClient1)

		couchClient, err := NewCouchClient(OreoCouchDBAddr)
		if err != nil {
			fmt.Printf("Error when creating couch client: %v\n", err)
			return nil, err
		}
		couchDB := couch.NewCouchDB(couchClient)

		return &oreo.NativeRealisticCreator{
			ConnMap: map[string]ycsb.DB{
				"Redis":   redisDB,
				"KVRocks": kvDB,
				"MongoDB": mongoDB,
				"CouchDB": couchDB,
			},
		}, nil
	}

	panic("Unknown pattern " + workloadType)
}

func OreoRealisticCreator(workloadType string, isRemote bool) (ycsb.DBCreator, error) {
	if workloadType == "iot" {
		kvConn := NewKVRocksConn()
		mongoConn := NewMongoDBConn()

		connMap := map[string]txn.Connector{
			"KVRocks": kvConn,
			"MongoDB": mongoConn,
		}
		return &oreo.OreoRealisticCreator{
			IsRemote:            isRemote,
			ConnMap:             connMap,
			GlobalDatastoreName: "KVRocks",
		}, nil
	}
	if workloadType == "social" {
		redisConn := NewRedisConn()
		mongoConn := NewMongoDBConn()
		couchConn := NewCouchDBConn()

		connMap := map[string]txn.Connector{
			"Redis":   redisConn,
			"MongoDB": mongoConn,
			"CouchDB": couchConn,
		}
		return &oreo.OreoRealisticCreator{
			IsRemote:            isRemote,
			ConnMap:             connMap,
			GlobalDatastoreName: "Redis",
		}, nil
	}
	if workloadType == "order" {
		redisConn := NewRedisConn()
		kvrocksConn := NewKVRocksConn()
		mongoConn := NewMongoDBConn()
		couchConn := NewCouchDBConn()

		connMap := map[string]txn.Connector{
			"Redis":   redisConn,
			"KVRocks": kvrocksConn,
			"MongoDB": mongoConn,
			"CouchDB": couchConn,
		}
		return &oreo.OreoRealisticCreator{
			IsRemote:            isRemote,
			ConnMap:             connMap,
			GlobalDatastoreName: "Redis",
		}, nil
	}
	panic("Unknown pattern " + workloadType)
}

func RedisCreator(addr string) (ycsb.DBCreator, error) {
	rdb1 := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: RedisPassword,
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
		Username: MongoUsername,
		Password: MongoPassword,
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
		Password: RedisPassword,
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
		Username:       MongoUsername,
		Password:       MongoPassword,
	})
	mongoConn2 := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
		Address:        OreoMongoDBAddr1,
		DBName:         "oreo",
		CollectionName: "benchmark",
		Username:       MongoUsername,
		Password:       MongoPassword,
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
		Username: CouchUsername,
		Password: CouchPassword,
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
			Username:       MongoUsername,
			Password:       MongoPassword,
		})
		mongoConn2 := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
			Address:        OreoMongoDBAddr2,
			DBName:         "oreo",
			CollectionName: "benchmark",
			Username:       MongoUsername,
			Password:       MongoPassword,
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
			Password: RedisPassword,
		})

		mongoConn1 := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
			Address:        OreoMongoDBAddr1,
			DBName:         "oreo",
			CollectionName: "benchmark",
			Username:       MongoUsername,
			Password:       MongoPassword,
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

func NewRedisConn() *redisCo.RedisConnection {
	redisConn := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address:  OreoRedisAddr,
		Password: RedisPassword,
		PoolSize: 100,
	})
	redisConn.Connect()
	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			redisConn.Get("1")
		}()
	}
	wg.Wait()

	return redisConn
}

// NewKVRocksConn initializes a new Redis connection using the KVRocks configuration.
// The function also attempts to warm up the connection by performing multiple concurrent GET operations.
//
// Returns:
//
//	*redisCo.RedisConnection: A pointer to the initialized Redis connection.
func NewKVRocksConn() *redisCo.RedisConnection {
	kvConn := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address:  OreoKVRocksAddr,
		Password: KVRocksPassword,
		PoolSize: 100,
	})
	kvConn.Connect()
	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			kvConn.Get("1")
		}()
	}
	wg.Wait()

	return kvConn
}

// NewMongoDBConn initializes a new MongoDB connection using the provided
// connection options, attempts to warm up the connection by performing
// multiple concurrent read operations, and returns the established connection.
// //
// Returns:
//
//	*mongoCo.MongoConnection: A pointer to the established MongoDB connection.
func NewMongoDBConn() *mongoCo.MongoConnection {
	mongoConn := mongoCo.NewMongoConnection(&mongoCo.ConnectionOptions{
		Address:        OreoMongoDBAddr1,
		DBName:         "oreo",
		CollectionName: "benchmark",
		Username:       MongoUsername,
		Password:       MongoPassword,
	})
	mongoConn.Connect()
	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mongoConn.Get("1")
		}()
	}
	wg.Wait()

	return mongoConn
}

func NewCouchDBConn() *couchdb.CouchDBConnection {
	couchConn := couchdb.NewCouchDBConnection(&couchdb.ConnectionOptions{
		Address:  OreoCouchDBAddr,
		DBName:   "oreo",
		Username: CouchUsername,
		Password: CouchPassword,
	})
	couchConn.Connect()
	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			couchConn.Get("1")
		}()
	}
	wg.Wait()

	return couchConn
}

func NewRedisClient(addr string) (*goredis.Client, error) {
	rdb1 := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: RedisPassword,
		PoolSize: 100,
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

	return rdb1, nil
}

func NewMongoDBClient(addr string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(addr)
	clientOptions.SetAuth(options.Credential{
		Username: MongoUsername,
		Password: MongoPassword,
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

	return client, nil
}

func NewCouchClient(addr string) (*kivik.Client, error) {
	client, err := kivik.New("couch", OreoCouchDBAddr)
	if err != nil {
		return nil, err
	}
	err = client.CreateDB(context.Background(), "oreo", nil)

	if err != nil && kivik.HTTPStatus(err) != http.StatusPreconditionFailed {
		return nil, err
	}

	return client, nil
}
