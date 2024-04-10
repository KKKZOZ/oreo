package client

import (
	"benchmark/pkg/generator"
	"benchmark/pkg/measurement"
	"benchmark/pkg/util"
	"benchmark/ycsb"
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type operationType int64

const (
	read operationType = iota + 1
	update
	insert
	scan
	readModifyWrite
)

type datastoreType int64

const (
	redisDatastore datastoreType = iota + 1
	mongoDatastore
)

// type PostCheckResult struct {
// 	DBName string
// 	Amount int
// }

const (
	MAX_VALUE_LENGTH = 100
)

type Client struct {
	dbCreatorMap map[string]ycsb.DBCreator
	wp           *ycsb.WorkloadParameter

	table string

	mu sync.Mutex

	r                *rand.Rand
	operationChooser *generator.Discrete
	datastoreChooser *generator.Discrete
	keyChooser       ycsb.Generator
	keySequence      ycsb.Generator

	zeroPadding int64
}

func NewClient(wp *ycsb.WorkloadParameter, dbCreatorMap map[string]ycsb.DBCreator) *Client {
	c := &Client{
		dbCreatorMap: dbCreatorMap,
		wp:           wp,
		table:        wp.TableName,
		r:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	c.operationChooser = createOperationGenerator(wp)
	c.datastoreChooser = createDatastoreGenerator(wp)

	insertStart := int64(0)
	insertCount := int64(wp.RecordCount) - insertStart

	c.keySequence = generator.NewCounter(insertStart)

	var keyrangeLowerBound int64 = insertStart
	var keyrangeUpperBound int64 = insertStart + insertCount - 1

	// insertProportion := wp.InsertProportion
	// opCount := wp.OperationCount
	// // expectedNewKeys := int64(float64(opCount) * insertProportion * 2.0)
	// // keyrangeUpperBound = insertStart + insertCount + expectedNewKeys
	c.keyChooser = generator.NewScrambledZipfian(keyrangeLowerBound, keyrangeUpperBound, generator.ZipfianConstant)

	return c
}

func (c *Client) NextOperation() operationType {
	c.mu.Lock()
	defer c.mu.Unlock()
	return operationType(c.operationChooser.Next(c.r))
}

func (c *Client) NextDatastore() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch datastoreType(c.datastoreChooser.Next(c.r)) {
	case redisDatastore:
		return "redis"
	case mongoDatastore:
		return "mongo"
	default:
		return "unknown"
	}
}

func (c *Client) NextKeyName() string {
	c.mu.Lock()
	keyNum := c.keyChooser.Next(c.r)
	c.mu.Unlock()
	return c.buildKeyName(keyNum)
}

func (c *Client) NextKeyNameFromSequence() string {
	c.mu.Lock()
	keyNum := c.keySequence.Next(c.r)
	c.mu.Unlock()
	return c.buildKeyName(keyNum)
}

func (c *Client) RunLoad() {

	ctx := context.Background()
	// we need to load data to all the datastores
	if c.wp.DBName == "oreo" {
		var wg sync.WaitGroup

		fmt.Printf("Loading data to oreo-redis\n")
		c.keySequence = generator.NewCounter(0)
		wg.Add(c.wp.ThreadCount)
		c.wp.TableName = "redis"
		for i := 0; i < c.wp.ThreadCount; i++ {
			go func(threadID int) {
				defer wg.Done()
				dbMap := c.genDBmap()
				w := newWorker(c, c.wp, threadID, c.wp.ThreadCount, dbMap)
				w.RunLoad(ctx, "oreo")
			}(i)
		}
		wg.Wait()

		fmt.Printf("Loading data to oreo-mongo\n")
		c.keySequence = generator.NewCounter(0)
		wg.Add(c.wp.ThreadCount)
		c.wp.TableName = "mongo"
		for i := 0; i < c.wp.ThreadCount; i++ {
			go func(threadID int) {
				defer wg.Done()
				dbMap := c.genDBmap()
				w := newWorker(c, c.wp, threadID, c.wp.ThreadCount, dbMap)
				w.RunLoad(ctx, "oreo")
			}(i)
		}
		wg.Wait()
		return
	}

	for dbName, creator := range c.dbCreatorMap {
		fmt.Printf("Loading data to %s\n", dbName)
		// reset the key sequence to load data
		c.keySequence = generator.NewCounter(0)
		var wg sync.WaitGroup
		wg.Add(c.wp.ThreadCount)

		for i := 0; i < c.wp.ThreadCount; i++ {
			go func(threadID int) {
				defer wg.Done()
				db, _ := creator.Create()
				dbMap := map[string]ycsb.DB{
					dbName: db,
				}
				w := newWorker(c, c.wp, threadID, c.wp.ThreadCount, dbMap)
				w.RunLoad(ctx, dbName)
			}(i)
		}

		wg.Wait()
	}

}

func (c *Client) RunBenchmark() {
	start := time.Now()
	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(c.wp.ThreadCount)

	for i := 0; i < c.wp.ThreadCount; i++ {
		go func(threadID int) {
			defer wg.Done()
			dbMap := c.genDBmap()
			w := newWorker(c, c.wp, threadID, c.wp.ThreadCount, dbMap)
			w.RunBenchmark(ctx, c.wp.DBName)
		}(i)
	}

	wg.Wait()

	fmt.Println("**********************************************************")
	fmt.Printf("Run finished, takes %s\n", time.Since(start))
	measurement.Output()

	if c.wp.DataConsistencyTest || c.wp.AcrossDatastoreTest {
		time.Sleep(5 * time.Second)

		amountMap := make(map[string]int)

		if c.wp.DBName == "oreo" {
			// check Redis
			// reset the key sequence to scan the whole datastore
			fmt.Printf("Start to check oreo-redis\n")
			c.keySequence = generator.NewCounter(0)
			resChan := make(chan int, c.wp.PostCheckWorkerThread)
			c.wp.TableName = "redis"
			for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
				go func(threadID int) {
					dbMap := c.genDBmap()
					w := newWorker(c, c.wp, threadID, c.wp.PostCheckWorkerThread, dbMap)
					w.RunPostCheck(ctx, "oreo", resChan)
				}(i)
			}
			curTotalAmount := 0
			for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
				curTotalAmount += <-resChan
				fmt.Printf("Progress: %v/%v\n", i+1, c.wp.PostCheckWorkerThread)
			}
			amountMap["oreo-redis"] = curTotalAmount

			// check Redis
			// reset the key sequence to scan the whole datastore
			fmt.Printf("Start to check oreo-mongo\n")
			c.keySequence = generator.NewCounter(0)
			c.wp.TableName = "mongo"
			for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
				go func(threadID int) {
					dbMap := c.genDBmap()
					w := newWorker(c, c.wp, threadID, c.wp.PostCheckWorkerThread, dbMap)
					w.RunPostCheck(ctx, "oreo", resChan)
				}(i)
			}
			curTotalAmount = 0
			for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
				curTotalAmount += <-resChan
				fmt.Printf("Progress: %v/%v\n", i+1, c.wp.PostCheckWorkerThread)
			}
			amountMap["oreo-mongo"] = curTotalAmount

		} else {

			for dbName, creator := range c.dbCreatorMap {
				// reset the key sequence to scan the whole datastore
				c.keySequence = generator.NewCounter(0)
				resChan := make(chan int, c.wp.PostCheckWorkerThread)

				for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
					go func(threadID int) {
						db, _ := creator.Create()
						dbMap := map[string]ycsb.DB{
							dbName: db,
						}
						w := newWorker(c, c.wp, threadID, c.wp.PostCheckWorkerThread, dbMap)
						w.RunPostCheck(ctx, dbName, resChan)
					}(i)
				}
				curTotalAmount := 0
				for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
					curTotalAmount += <-resChan
				}
				amountMap[dbName] = curTotalAmount
			}
		}

		// iterate all the datastores
		// since they all will be scanned

		fmt.Println("**********************************************************")
		fmt.Printf("Data consistency check finished.\n")
		total := 0
		for dbName, amount := range amountMap {
			total += amount
			fmt.Printf("%s:\nExpected Amount: %v\nCurrent  Amount: %v\n",
				dbName, c.wp.TotalAmount, amount)
		}
		fmt.Printf("Total: Current  Amount: %v\n", total)
	}
}

// genDBmap generates a map of database instances
// based on the registered database creators.
func (c *Client) genDBmap() map[string]ycsb.DB {
	c.mu.Lock()
	defer c.mu.Unlock()
	dbMap := make(map[string]ycsb.DB)
	for dbName, creator := range c.dbCreatorMap {
		db, _ := creator.Create()
		dbMap[dbName] = db
	}
	return dbMap
}

func (c *Client) BuildRandomValue() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	len := c.r.Intn(MAX_VALUE_LENGTH) + 1
	buf := make([]byte, len)
	util.RandBytes(c.r, buf)
	return string(buf)
}

func (c *Client) buildKeyName(keyNum int64) string {

	// keyNum = util.Hash64(keyNum)
	prefix := "benchmark"
	return fmt.Sprintf("%s%0[3]*[2]d", prefix, keyNum, c.zeroPadding)
}

func createOperationGenerator(wp *ycsb.WorkloadParameter) *generator.Discrete {
	readProportion := wp.ReadProportion
	updateProportion := wp.UpdateProportion
	insertProportion := wp.InsertProportion
	scanProportion := wp.ScanProportion
	readModifyWriteProportion := wp.ReadModifyWriteProportion

	operationChooser := generator.NewDiscrete()
	if readProportion > 0 {
		operationChooser.Add(readProportion, int64(read))
	}

	if updateProportion > 0 {
		operationChooser.Add(updateProportion, int64(update))
	}

	if insertProportion > 0 {
		operationChooser.Add(insertProportion, int64(insert))
	}

	if scanProportion > 0 {
		operationChooser.Add(scanProportion, int64(scan))
	}

	if readModifyWriteProportion > 0 {
		operationChooser.Add(readModifyWriteProportion, int64(readModifyWrite))
	}

	return operationChooser
}

func createDatastoreGenerator(wp *ycsb.WorkloadParameter) *generator.Discrete {
	redisProportion := wp.RedisProportion
	mongoProportion := wp.MongoProportion

	datastoreChooser := generator.NewDiscrete()
	if redisProportion > 0 {
		datastoreChooser.Add(redisProportion, int64(redisDatastore))
	}

	if mongoProportion > 0 {
		datastoreChooser.Add(mongoProportion, int64(mongoDatastore))
	}

	return datastoreChooser
}
