package main

import (
	"benchmark/db/redis"
	"benchmark/pkg/client"
	"benchmark/pkg/measurement"
	"benchmark/ycsb"
	"fmt"
	"time"

	redisCo "github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/logger"
)

func main() {
	redisConn := redisCo.NewRedisConnection(nil)
	redis := redis.NewRedis(redisConn)
	redisDB := client.DbWrapper{DB: redis}

	wp := &ycsb.WorkloadParameter{
		RecordCount:               1000,
		OperationCount:            1000,
		ReadProportion:            0.5,
		UpdateProportion:          0.5,
		InsertProportion:          0,
		ScanProportion:            0,
		ReadModifyWriteProportion: 0,
	}
	client := client.NewClient(redisDB, wp)

	measurement.InitMeasure()
	measurement.EnableWarmUp(true)

	logger.Log.Infof("Start benchmark")
	logger.Log.Infof("Start to load data")
	client.RunLoad()
	logger.Log.Infof("Start to run benchmark")

	measurement.EnableWarmUp(false)
	start := time.Now()
	client.RunBenchmark()
	fmt.Println("**********************************************************")
	fmt.Printf("Run finished, takes %s\n", time.Since(start))

	measurement.Output()

}
