package main

import (
	"benchmark/db/oreo"
	"benchmark/db/redis"
	"benchmark/pkg/client"
	"benchmark/pkg/measurement"
	"benchmark/ycsb"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/kkkzoz/oreo/pkg/config"
	redisCo "github.com/kkkzoz/oreo/pkg/datastore/redis"
)

// 43.139.62.221

func RedisCreator() ycsb.DBCreator {
	redisConn := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address: "43.139.62.221:6379",
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
	return &redis.RedisCreator{Conn: redisConn}
}

func OreoDBCreator() ycsb.DBCreator {
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
	return &oreo.OreoRedisCreator{Conn: redisConn}
}

func main() {

	args := os.Args

	argsLen := len(args)
	if argsLen < 4 {

		fmt.Println("Usage: main [redis|oreo] [load|run] [ThreadNum]")
		return
	}

	wp := &ycsb.WorkloadParameter{
		RecordCount:               1000,
		OperationCount:            100,
		TxnOperationGroup:         10,
		ReadProportion:            0.5,
		UpdateProportion:          0.5,
		InsertProportion:          0,
		ScanProportion:            0,
		ReadModifyWriteProportion: 0,
	}

	// ignore INFO level messages
	config.Config.ConcurrentOptimizationLevel = config.PARALLELIZE_ON_UPDATE
	config.Config.AsyncLevel = config.AsyncLevelZero

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

	if args[1] == "redis" {
		fmt.Println("Pure redis test")
		c = client.NewClient(wp, RedisCreator())
	} else {
		fmt.Println("Oreo test")
		c = client.NewClient(wp, OreoDBCreator())
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

	if mode == "load" {
		wp.DoBenchmark = false
		fmt.Println("Start to load data")
		c.RunLoad()
		fmt.Println("Load finished")
		return
	} else {
		wp.DoBenchmark = true
		fmt.Println("Start to run benchmark")

		measurement.EnableWarmUp(false)
		start := time.Now()
		c.RunBenchmark()
		fmt.Println("**********************************************************")
		fmt.Printf("Run finished, takes %s\n", time.Since(start))

		measurement.Output()
	}

}
