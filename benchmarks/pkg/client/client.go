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

const (
	MAX_VALUE_LENGTH = 100
)

type Client struct {
	dbCreator ycsb.DBCreator
	wp        *ycsb.WorkloadParameter

	table string

	mu sync.Mutex

	r                *rand.Rand
	operationChooser *generator.Discrete
	keyChooser       ycsb.Generator
	keySequence      ycsb.Generator

	zeroPadding int64
}

func NewClient(wp *ycsb.WorkloadParameter, dbCreator ycsb.DBCreator) *Client {
	c := &Client{
		dbCreator: dbCreator,
		wp:        wp,
		table:     wp.TableName,
		r:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	c.operationChooser = createOperationGenerator(wp)

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
	var wg sync.WaitGroup
	wg.Add(c.wp.ThreadCount)

	for i := 0; i < c.wp.ThreadCount; i++ {

		go func(threadID int) {
			defer wg.Done()
			db, _ := c.dbCreator.Create()
			w := newWorker(c, c.wp, threadID, c.wp.ThreadCount, db)
			w.RunLoad(ctx)
		}(i)
	}

	wg.Wait()

}

func (c *Client) RunBenchmark() {
	start := time.Now()

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(c.wp.ThreadCount)

	for i := 0; i < c.wp.ThreadCount; i++ {

		go func(threadID int) {
			defer wg.Done()
			db, _ := c.dbCreator.Create()
			w := newWorker(c, c.wp, threadID, c.wp.ThreadCount, db)
			w.RunBenchmark(ctx)
		}(i)
	}

	wg.Wait()

	fmt.Println("**********************************************************")
	fmt.Printf("Run finished, takes %s\n", time.Since(start))
	measurement.Output()

	if c.wp.DataConsistencyTest {
		time.Sleep(5 * time.Second)
		// reset the key sequence
		c.keySequence = generator.NewCounter(0)
		resChan := make(chan int, c.wp.PostCheckWorkerThread)
		for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
			go func(threadID int) {
				db, _ := c.dbCreator.Create()
				w := newWorker(c, c.wp, threadID, c.wp.PostCheckWorkerThread, db)
				w.RunPostCheck(ctx, resChan)
			}(i)
		}

		curTotalAmount := 0
		for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
			curTotalAmount += <-resChan
		}
		fmt.Println("**********************************************************")
		fmt.Printf("Data consistency check finished.\n")
		fmt.Printf("Expected Amount: %v\nCurrent  Amount: %v\n",
			c.wp.TotalAmount, curTotalAmount)
	}
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
