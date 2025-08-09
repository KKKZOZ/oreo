package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"benchmark/pkg/benconfig"
	"benchmark/pkg/errrecord"
	"benchmark/pkg/measurement"
	"benchmark/pkg/workload"
	"benchmark/ycsb"
	"github.com/kkkzoz/oreo/pkg/config"
)

type Client struct {
	mu           sync.Mutex
	dbCreatorMap map[string]ycsb.DBCreator
	wp           *workload.WorkloadParameter
	table        string

	wl workload.Workload
}

func NewClient(
	workload *workload.Workload,
	wp *workload.WorkloadParameter,
	dbCreatorMap map[string]ycsb.DBCreator,
) *Client {
	return &Client{
		mu:           sync.Mutex{},
		wl:           *workload,
		dbCreatorMap: dbCreatorMap,
		wp:           wp,
		table:        wp.TableName,
	}
}

func (c *Client) RunLoad() {
	ctx := context.Background()

	for dbName, creator := range c.dbCreatorMap {
		fmt.Printf("Loading data to %s\n", dbName)
		c.wl.ResetKeySequence()
		var wg sync.WaitGroup
		wg.Add(c.wp.ThreadCount)

		for i := 0; i < c.wp.ThreadCount; i++ {
			go func(threadID int) {
				defer wg.Done()
				db, _ := creator.Create()
				dbMap := map[string]ycsb.DB{
					dbName: db,
				}
				w := newWorker(c.wl, c.wp, threadID, c.wp.ThreadCount, dbMap)
				w.RunLoad(ctx, dbName)
			}(i)
		}
		wg.Wait()
	}

	// we need to load data to all the datastores
	// if c.wp.DBName == "oreo" {
	// 	var wg sync.WaitGroup

	// 	fmt.Printf("Loading data to oreo-redis\n")
	// 	c.wl.ResetKeySequence()
	// 	wg.Add(c.wp.ThreadCount)
	// 	c.wp.TableName = "redis"
	// 	for i := 0; i < c.wp.ThreadCount; i++ {
	// 		go func(threadID int) {
	// 			defer wg.Done()
	// 			dbMap := c.genDBmap()
	// 			w := newWorker(c.wl, c.wp, threadID, c.wp.ThreadCount, dbMap)
	// 			w.RunLoad(ctx, "oreo")
	// 		}(i)
	// 	}
	// 	wg.Wait()

	// 	fmt.Printf("Loading data to oreo-mongo\n")
	// 	c.wl.ResetKeySequence()
	// 	wg.Add(c.wp.ThreadCount)
	// 	c.wp.TableName = "mongo"
	// 	for i := 0; i < c.wp.ThreadCount; i++ {
	// 		go func(threadID int) {
	// 			defer wg.Done()
	// 			dbMap := c.genDBmap()
	// 			w := newWorker(c.wl, c.wp, threadID, c.wp.ThreadCount, dbMap)
	// 			w.RunLoad(ctx, "oreo")
	// 		}(i)
	// 	}
	// 	wg.Wait()
	// 	return
	// }
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
			w := newWorker(c.wl, c.wp, threadID, c.wp.ThreadCount, dbMap)
			w.RunBenchmark(ctx, c.wp.DBName)
		}(i)
	}
	wg.Wait()

	fmt.Println("----------------------------------")
	fmt.Printf("Run finished, takes %.8fs\n", time.Since(start).Seconds())
	measurement.Output()
	errrecord.Summary()
	fmt.Println("----------------------------------")
	fmt.Printf("AssumptionCount: %v\n", config.Debug.AssumptionCount)

	// if c.wp.WorkloadName == "ycsb" {
	// 	fmt.Printf("Check record distribution\n")
	// 	c.wl.DisplayCheckResult()
	// }
	c.getCacheState()

	if !c.wl.NeedPostCheck() {
		return
	}

	time.Sleep(time.Duration(c.wp.PostCheckInterval) * time.Millisecond)
	// amountMap := make(map[string]int)

	for dbName, creator := range c.dbCreatorMap {
		// reset the key sequence to scan the whole datastore
		c.wl.ResetKeySequence()
		resChan := make(chan int, c.wp.PostCheckWorkerThread)
		var wg sync.WaitGroup
		wg.Add(c.wp.PostCheckWorkerThread)
		for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
			go func(threadID int) {
				defer wg.Done()
				db, _ := creator.Create()
				dbMap := map[string]ycsb.DB{
					dbName: db,
				}
				w := newWorker(c.wl, c.wp, threadID, c.wp.PostCheckWorkerThread, dbMap)
				w.RunPostCheck(ctx, dbName, resChan)
			}(i)
		}
		wg.Wait()
	}

	c.wl.DisplayCheckResult()

	// time.Sleep(5 * time.Second)
	// if c.wp.DBName == "oreo" {
	// 	// check Redis
	// 	// reset the key sequence to scan the whole datastore
	// 	fmt.Printf("Start to check oreo-redis\n")
	// 	c.wl.ResetKeySequence()
	// 	resChan := make(chan int, c.wp.PostCheckWorkerThread)
	// 	c.wp.TableName = "redis"
	// 	for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
	// 		go func(threadID int) {
	// 			dbMap := c.genDBmap()
	// 			w := newWorker(c.wl, c.wp, threadID, c.wp.PostCheckWorkerThread, dbMap)
	// 			w.RunPostCheck(ctx, "oreo", resChan)
	// 		}(i)
	// 	}
	// 	curTotalAmount := 0
	// 	for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
	// 		curTotalAmount += <-resChan
	// 		fmt.Printf("Progress: %v/%v\n", i+1, c.wp.PostCheckWorkerThread)
	// 	}
	// 	amountMap["oreo-redis"] = curTotalAmount

	// 	// check Redis
	// 	// reset the key sequence to scan the whole datastore
	// 	fmt.Printf("Start to check oreo-mongo\n")
	// 	c.wl.ResetKeySequence()
	// 	c.wp.TableName = "mongo"
	// 	for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
	// 		go func(threadID int) {
	// 			dbMap := c.genDBmap()
	// 			w := newWorker(c.wl, c.wp, threadID, c.wp.PostCheckWorkerThread, dbMap)
	// 			w.RunPostCheck(ctx, "oreo", resChan)
	// 		}(i)
	// 	}
	// 	curTotalAmount = 0
	// 	for i := 0; i < c.wp.PostCheckWorkerThread; i++ {
	// 		curTotalAmount += <-resChan
	// 		fmt.Printf("Progress: %v/%v\n", i+1, c.wp.PostCheckWorkerThread)
	// 	}
	// 	amountMap["oreo-mongo"] = curTotalAmount
	// }
}

func (c *Client) getCacheState() {
	fmt.Println("----------------------------------")

	executorUrlList := benconfig.ExecutorAddressMap["ALL"]

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	for _, url := range executorUrlList {
		tar_url := "http://" + url + "/cache"
		resp, err := client.Get(tar_url)
		if err != nil {
			fmt.Printf("Failed to send GET request to %s: %v\n", tar_url, err)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body from %s: %v\n", tar_url, err)
			continue
		}

		fmt.Printf("[%s]: %s\n", tar_url, string(body))
		_, _ = http.Post(tar_url, "application/json", nil)
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
