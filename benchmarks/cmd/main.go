package main

import (
	mongoDB "benchmark/db/mongo"
	"benchmark/db/oreo"
	"benchmark/db/redis"
	"benchmark/pkg/client"
	"benchmark/pkg/measurement"
	"benchmark/pkg/workload"
	"benchmark/ycsb"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/trace"
	"sync"
	"time"

	"github.com/kkkzoz/oreo/pkg/config"
	mongoCo "github.com/kkkzoz/oreo/pkg/datastore/mongo"
	redisCo "github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/txn"
	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	RedisDBAddr      = "43.139.62.221:6371"
	MongoDBAddr      = "mongodb://43.139.62.221:27017"
	MongoDBAddr2     = "mongodb://43.139.62.221:27021"
	MongoDBGroupAddr = "mongodb://43.139.62.221:27021,43.139.62.221:27022,43.139.62.221:27023/?replicaSet=dbrs"
	OreoRedisAddr    = "43.139.62.221:6380"
)

// const (
// 	RedisDBAddr   = "localhost:6379"
// 	MongoDBAddr   = "mongodb://43.139.62.221:27017"
// 	OreoRedisAddr = "localhost:6380"
// )

var help = flag.Bool("help", false, "Show help")
var dbType = ""
var mode = "load"
var workloadType = ""
var threadNum = 1
var traceFlag = false

// fmt.Println("Usage: main [DBType] [load|run] [ThreadNum] [TestTypeFlag]")
func main() {

	flag.StringVar(&dbType, "d", "", "DB type")
	flag.StringVar(&mode, "m", "load", "Mode: load or run")
	flag.StringVar(&workloadType, "wl", "", "Workload type")
	flag.IntVar(&threadNum, "t", 1, "Thread number")
	flag.BoolVar(&traceFlag, "trace", false, "Enable trace")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if threadNum <= 0 {
		fmt.Println("ThreadNum should be a positive integer")
		return
	}

	if traceFlag {
		f, err := os.Create("trace.out")
		if err != nil {
			panic(err)
		}
		err = trace.Start(f)
		if err != nil {
			panic(err)
		}
		defer trace.Stop()
	}

	// TODO: Read it from file
	wp := &workload.WorkloadParameter{
		RecordCount:               100,
		OperationCount:            100,
		TxnOperationGroup:         10,
		ReadProportion:            0,
		UpdateProportion:          0,
		InsertProportion:          0,
		ScanProportion:            0,
		ReadModifyWriteProportion: 1.0,

		InitialAmountPerKey:   1000,
		TransferAmountPerTxn:  5,
		PostCheckWorkerThread: 50,

		RedisProportion: 0.5,
		MongoProportion: 0.5,
	}
	wp.TotalAmount = wp.InitialAmountPerKey * wp.RecordCount

	config.Config.ConcurrentOptimizationLevel = 0
	config.Config.AsyncLevel = 2
	config.Config.MaxOutstandingRequest = 3
	// config.Config.MaxRecordLength = 2

	measurement.InitMeasure()
	measurement.EnableWarmUp(true)

	wp.ThreadCount = threadNum
	wl := createWorkload(wp)
	client := generateClient(&wl, wp, dbType)

	displayBenchmarkInfo()

	switch mode {
	case "load":
		config.Config.ConcurrentOptimizationLevel = config.DEFAULT
		wp.DoBenchmark = false
		fmt.Println("Start to load data")
		client.RunLoad()
		fmt.Println("Load finished")
	case "run":
		wp.DoBenchmark = true
		fmt.Println("Start to run benchmark")
		measurement.EnableWarmUp(false)
		client.RunBenchmark()
	default:
		panic("Invalid mode")
	}
}

func displayBenchmarkInfo() {
	fmt.Printf("-----------------\n")
	fmt.Printf("DBType: %s\n", dbType)
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("WorkloadType: %s\n", workloadType)
	fmt.Printf("ThreadNum: %d\n", threadNum)
	fmt.Printf("ConcurrentOptimizationLevel: %d\nAsyncLevel: %d\nMaxOutstandingRequest: %d\n",
		config.Config.ConcurrentOptimizationLevel, config.Config.AsyncLevel,
		config.Config.MaxOutstandingRequest)
	fmt.Printf("-----------------\n")
}

func createWorkload(wp *workload.WorkloadParameter) workload.Workload {
	if workloadType != "" {
		switch workloadType {
		case "ycsb":
			fmt.Println("This is a YCSB benchmark")
			return workload.NewYCSBWorkload(wp)
		case "dc":
			fmt.Println("This is a data consistency test")
			return workload.NewDataConsistencyWorkload(wp)
		case "tp":
			fmt.Println("This is a transaction performance test")
			return workload.NewTxnPerformanceWorkload(wp)
		case "ad":
			fmt.Println("This is a across datastore test")
			return workload.NewAcrossDatastoreWorkload(wp)
		default:
			panic("Invalid workload type")
		}
	} else {
		panic("WorkloadType should be specified")
	}
}

func generateClient(wl *workload.Workload, wp *workload.WorkloadParameter, dbName string) *client.Client {
	if dbType == "" {
		panic("DBType should be specified")
	}

	var c *client.Client
	switch dbName {
	case "redis":
		wp.DBName = "redis"
		creator, err := RedisCreator()
		if err != nil {
			fmt.Printf("Error when creating redis client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "mongo":
		wp.DBName = "mongo"
		creator, err := MongoCreator()
		if err != nil {
			fmt.Printf("Error when creating mongo client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "redis-mongo":
		wp.DBName = "redis-mongo"
		redisCreator, err1 := RedisCreator()
		mongoCreator, err2 := MongoCreator()
		if err1 != nil || err2 != nil {
			fmt.Printf("Error when creating client: %v %v\n", err1, err2)
			return nil
		}
		dbSetCreator := workload.DBSetCreator{
			CreatorMap: map[string]ycsb.DBCreator{
				"redis": redisCreator,
				"mongo": mongoCreator,
			},
		}

		creatorMap := map[string]ycsb.DBCreator{
			"redis-mongo": &dbSetCreator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "oreo-redis":
		wp.DBName = "oreo-redis"
		creator, err := OreoRedisCreator()
		if err != nil {
			fmt.Printf("Error when creating oreo-redis client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "oreo":
		wp.DBName = "oreo"
		creator, err := OreoCreator()
		if err != nil {
			fmt.Printf("Error when creating oreo-redis client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	default:
		panic("Unsupport db type")
	}
	return c
}
func RedisCreator() (ycsb.DBCreator, error) {
	rdb1 := goredis.NewClient(&goredis.Options{
		Addr:     RedisDBAddr,
		Password: "@ljy123456",
	})

	rdb2 := goredis.NewClient(&goredis.Options{
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
			rdb2.Get(context.Background(), "1")
		}()
	}
	wg.Wait()

	return &redis.RedisCreator{RdbList: []*goredis.Client{rdb1, rdb2}}, nil
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

func OreoRedisCreator() (ycsb.DBCreator, error) {
	redisConn1 := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address:  OreoRedisAddr,
		Password: "@ljy123456",
	})
	redisConn2 := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address:  OreoRedisAddr,
		Password: "@ljy123456",
	})
	redisConn3 := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address:  OreoRedisAddr,
		Password: "@ljy123456",
	})
	redisConn4 := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address:  OreoRedisAddr,
		Password: "@ljy123456",
	})

	redisConn1.Connect()
	redisConn2.Connect()
	redisConn3.Connect()
	redisConn4.Connect()
	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			redisConn1.Get("1")
			redisConn2.Get("1")
			redisConn3.Get("1")
			redisConn4.Get("1")
		}()
	}
	wg.Wait()
	return &oreo.OreoRedisCreator{
		ConnList: []*redisCo.RedisConnection{
			redisConn1, redisConn2, redisConn3, redisConn4}}, nil
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
