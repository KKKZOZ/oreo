package main

import (
	"benchmark/pkg/client"
	"benchmark/pkg/config"
	"benchmark/pkg/measurement"
	"benchmark/pkg/workload"
	"benchmark/ycsb"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"sync"

	cfg "github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/network"
)

const (
	RedisDBAddr = "43.139.62.221:6371"
	// MongoDBAddr      = "mongodb://43.139.62.221:27017"
	MongoDBAddr      = "mongodb://localhost:27017"
	MongoDBAddr2     = "mongodb://43.139.62.221:27021"
	MongoDBGroupAddr = "mongodb://43.139.62.221:27021,43.139.62.221:27022,43.139.62.221:27023/?replicaSet=dbrs"
	OreoRedisAddr    = "43.139.62.221:6379"
	// OreoRedisAddr = "localhost:6380"

	OreoCouchDBAddr = "http://admin:password@43.139.62.221:5984"
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
var isRemote = false

// fmt.Println("Usage: main [DBType] [load|run] [ThreadNum] [TestTypeFlag]")
func main() {

	flag.StringVar(&dbType, "d", "", "DB type")
	flag.StringVar(&mode, "m", "load", "Mode: load or run")
	flag.StringVar(&workloadType, "wl", "", "Workload type")
	flag.IntVar(&threadNum, "t", 1, "Thread number")
	flag.BoolVar(&traceFlag, "trace", false, "Enable trace")
	flag.BoolVar(&isRemote, "remote", false, "Run in remote mode (for Oreo series)")
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
		RecordCount:               1000,
		OperationCount:            5,
		TxnOperationGroup:         5,
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

	cfg.Config.ConcurrentOptimizationLevel = 0
	cfg.Config.AsyncLevel = 1
	cfg.Config.MaxOutstandingRequest = 5
	cfg.Config.MaxRecordLength = 2

	measurement.InitMeasure()
	measurement.EnableWarmUp(true)

	wp.ThreadCount = threadNum
	wl := createWorkload(wp)
	client := generateClient(&wl, wp, dbType)

	if isRemote {
		cfg.Config.AsyncLevel = 2
		warmUpHttpClient()
	}
	displayBenchmarkInfo()

	switch mode {
	case "load":
		cfg.Config.ConcurrentOptimizationLevel = cfg.DEFAULT
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
	fmt.Printf("Remote Mode: %v\n", isRemote)
	fmt.Printf("ConcurrentOptimizationLevel: %d\nAsyncLevel: %d\nMaxOutstandingRequest: %d\nMaxRecordLength: %d\n",
		cfg.Config.ConcurrentOptimizationLevel, cfg.Config.AsyncLevel,
		cfg.Config.MaxOutstandingRequest, cfg.Config.MaxRecordLength)
	fmt.Printf("-----------------\n")
}

func warmUpHttpClient() {
	url := fmt.Sprintf("http://%s/ping", config.RemoteAddress)
	num := min(300, threadNum)

	var wg sync.WaitGroup
	wg.Add(num)

	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()
			resp, err := network.HttpClient.Get(url)
			if err != nil {
				fmt.Printf("Error when warming up http client: %v\n", err)
			}
			defer func() {
				_, _ = io.CopyN(io.Discard, resp.Body, 1024*4)
				_ = resp.Body.Close()
			}()
		}()
	}
	wg.Wait()
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
		creator, err := OreoRedisCreator(isRemote)
		if err != nil {
			fmt.Printf("Error when creating oreo-redis client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "oreo-mongo":
		wp.DBName = "oreo-mongo"
		creator, err := OreoMongoCreator()
		if err != nil {
			fmt.Printf("Error when creating oreo-mongo client: %v\n", err)
			return nil
		}
		creatorMap := map[string]ycsb.DBCreator{
			dbName: creator,
		}
		c = client.NewClient(wl, wp, creatorMap)
	case "oreo-couch":
		wp.DBName = "oreo-couch"
		creator, err := OreoCouchCreator()
		if err != nil {
			fmt.Printf("Error when creating oreo-couch client: %v\n", err)
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
