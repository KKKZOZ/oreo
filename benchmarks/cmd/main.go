package main

import (
	"benchmark/db/oreo"
	"benchmark/db/redis"
	"benchmark/pkg/client"
	"benchmark/pkg/measurement"
	"benchmark/ycsb"
	"fmt"
	"os"
	"runtime/trace"
	"sync"
	"time"

	"github.com/kkkzoz/oreo/pkg/config"
	redisCo "github.com/kkkzoz/oreo/pkg/datastore/redis"
)

// 43.139.62.221

func PureRedisDB() client.DbWrapper {
	redisConn := redisCo.NewRedisConnection(&redisCo.ConnectionOptions{
		Address: "43.139.62.221:6379",
	})

	// try to warm up the connection
	// var wg sync.WaitGroup
	// for i := 1; i <= 15; i++ {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		redisConn.Get("1")
	// 	}()
	// }
	// wg.Wait()
	return client.DbWrapper{DB: redis.NewRedis(redisConn)}
}

func OreoDB() *client.TxnDbWrapper {
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
	return &client.TxnDbWrapper{DB: oreo.NewRedisDatastore(redisConn)}
}

func main() {

	args := os.Args

	wp := &ycsb.WorkloadParameter{
		RecordCount:               100,
		OperationCount:            10,
		TxnOperationGroup:         10,
		ReadProportion:            0,
		UpdateProportion:          0,
		InsertProportion:          0,
		ScanProportion:            0,
		ReadModifyWriteProportion: 1.0,
	}

	// ignore INFO level messages
	config.Config.ConcurrentOptimizationLevel = config.DEFAULT

	// f, err := os.Create("trace.out")
	// if err != nil {
	// 	panic(err)
	// }
	// err = trace.Start(f)
	// if err != nil {
	// 	panic(err)
	// }
	defer trace.Stop()

	var c *client.Client

	if args[1] == "redis" {
		fmt.Println("Pure redis test")
		c = client.NewClient(PureRedisDB(), wp)
	} else {
		fmt.Println("Oreo test")
		c = client.NewClient(OreoDB(), wp)
	}

	measurement.InitMeasure()
	measurement.EnableWarmUp(true)

	if args[2] == "load" {
		fmt.Println("Start to load data")
		c.RunLoad()
		fmt.Println("Load finished")
		return
	} else {
		fmt.Println("Start to run benchmark")

		measurement.EnableWarmUp(false)
		start := time.Now()
		c.RunBenchmark()
		fmt.Println("**********************************************************")
		fmt.Printf("Run finished, takes %s\n", time.Since(start))

		measurement.Output()
	}

}
