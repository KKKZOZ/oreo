package main

import (
	mongoDB "benchmark/db/mongo"
	"benchmark/db/oreo"
	"benchmark/db/redis"
	"benchmark/pkg/client"
	"benchmark/pkg/measurement"
	"benchmark/ycsb"
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/kkkzoz/oreo/pkg/config"
	redisCo "github.com/kkkzoz/oreo/pkg/datastore/redis"
	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 43.139.62.221

func RedisCreator() (ycsb.DBCreator, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr: "43.139.62.221:6371",
	})

	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 15; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rdb.Get(context.Background(), "1")
		}()
	}
	wg.Wait()
	return &redis.RedisCreator{Rdb: rdb}, nil
}

func MongoCreator() (ycsb.DBCreator, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
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

func OreoDBCreator() (ycsb.DBCreator, error) {
	// 43.139.62.221:6380
	redisConn := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address: "43.139.62.221:6380",
	})

	// try to warm up the connection
	var wg sync.WaitGroup
	for i := 1; i <= 15; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			redisConn.Get("1")
		}()
	}
	wg.Wait()
	return &oreo.OreoRedisCreator{Conn: redisConn}, nil
}

func main() {

	args := os.Args
	argsLen := len(args)
	if argsLen < 4 {
		fmt.Println("Usage: main [DBType] [load|run] [ThreadNum] [TestTypeFlag]")
		return
	}

	// TODO: Read it from file
	wp := &ycsb.WorkloadParameter{
		RecordCount:               100,
		OperationCount:            100,
		TxnOperationGroup:         10,
		ReadProportion:            0.5,
		UpdateProportion:          0.5,
		InsertProportion:          0,
		ScanProportion:            0,
		ReadModifyWriteProportion: 0,

		DataConsistencyTest:   false,
		InitialAmountPerKey:   1000,
		TransferAmountPerTxn:  1,
		PostCheckWorkerThread: 100,
	}
	wp.TotalAmount = wp.InitialAmountPerKey * wp.RecordCount

	config.Config.ConcurrentOptimizationLevel = 1
	config.Config.AsyncLevel = 2

	// f, err := os.Create("trace.out")
	// if err != nil {
	// 	panic(err)
	// }
	// err = trace.Start(f)
	// if err != nil {
	// 	panic(err)
	// }
	// defer trace.Stop()

	var c *client.Client

	switch args[1] {
	case "redis":
		fmt.Println("Pure redis test")
		wp.DBName = "redis"
		creator, err := RedisCreator()
		if err != nil {
			fmt.Printf("Error when creating redis client: %v\n", err)
			return
		}
		c = client.NewClient(wp, creator)
	case "mongo":
		fmt.Println("Mongo test")
		wp.DBName = "mongo"
		creator, err := MongoCreator()
		if err != nil {
			fmt.Printf("Error when creating mongo client: %v\n", err)
			return
		}
		c = client.NewClient(wp, creator)
	case "oreo-redis":
		fmt.Println("Oreo test")
		wp.DBName = "oreo-redis"
		creator, err := OreoDBCreator()
		if err != nil {
			fmt.Printf("Error when creating redis client: %v\n", err)
			return
		}
		c = client.NewClient(wp, creator)
	default:
		panic("Unsupport db type")
	}

	measurement.InitMeasure()
	measurement.EnableWarmUp(true)

	mode := args[2]
	threadNum, err := strconv.Atoi(args[3])
	if err != nil {
		fmt.Println("ThreadNum should be an integer")
		return
	}
	wp.ThreadCount = threadNum

	if argsLen == 5 {
		if args[4] == "-dc" {
			wp.DataConsistencyTest = true
			fmt.Println("This is a data consistency test")
		}
		if args[4] == "-tp" {
			wp.TxnPerformanceTest = true
			fmt.Println("This is a transaction performance test")
		}
	}

	if mode == "load" {
		// TODO:
		config.Config.ConcurrentOptimizationLevel = config.DEFAULT
		wp.DoBenchmark = false
		fmt.Println("Start to load data")
		c.RunLoad()
		fmt.Println("Load finished")
		return
	} else {
		wp.DoBenchmark = true
		fmt.Println("Start to run benchmark")
		measurement.EnableWarmUp(false)
		c.RunBenchmark()
	}

}
