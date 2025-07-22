package client

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"benchmark/pkg/workload"
	"benchmark/ycsb"
)

type worker struct {
	wl           workload.Workload
	wp           *workload.WorkloadParameter
	threadID     int
	wrappedDBMap map[string]ycsb.DB
	originDBMap  map[string]ycsb.DB

	opCount int
}

func newWorker(
	wl workload.Workload,
	wp *workload.WorkloadParameter,
	threadID int, threadCount int,
	workDBMap map[string]ycsb.DB,
) *worker {
	w := &worker{
		wl:           wl,
		wp:           wp,
		threadID:     threadID,
		originDBMap:  workDBMap,
		wrappedDBMap: make(map[string]ycsb.DB),
	}

	for name, workDB := range workDBMap {
		switch db := workDB.(type) {
		case ycsb.TransactionDB:
			w.wrappedDBMap[name] = &TxnDbWrapper{DB: db}
		case ycsb.DB:
			w.wrappedDBMap[name] = &DbWrapper{DB: db}
		default:
			fmt.Printf("unknown db type: %T", workDB)
			os.Exit(-1)
		}
	}

	var totalOpCount int
	if w.wp.DoBenchmark {
		totalOpCount = w.wp.OperationCount
	} else {
		totalOpCount = w.wp.RecordCount
	}

	// if totalOpCount < threadCount {
	// 	fmt.Printf("totalOpCount(%d/%d): %d should be bigger than threadCount: %d",
	// 		wp.OperationCount,
	// 		wp.RecordCount,
	// 		totalOpCount,
	// 		threadCount)

	// 	os.Exit(-1)
	// }

	// log.Printf("threadID: %d, totalOpCount: %d, threadCount: %d\n", threadID, totalOpCount, threadCount)

	w.opCount = totalOpCount / threadCount
	if threadID < totalOpCount%threadCount {
		w.opCount++
	}
	if w.opCount == 0 {
		log.Fatalf("opCount should be bigger than 0")
		os.Exit(-1)
	}

	return w
}

func (w *worker) RunLoad(ctx context.Context, dbName string) {
	w.wl.Load(ctx, w.opCount, w.wrappedDBMap[dbName])
}

func (w *worker) RunBenchmark(ctx context.Context, dbName string) {
	var db ycsb.DB
	if w.wl.NeedRawDB() {
		db = w.originDBMap[dbName]
	} else {
		db = w.wrappedDBMap[dbName]
	}
	interval := rand.Intn(200)
	time.Sleep(time.Duration(interval) * time.Millisecond)
	ctxKV := context.WithValue(ctx, "threadID", w.threadID)
	log.Printf("Worker %d: start benchmark for %s\n", w.threadID, dbName)
	w.wl.Run(ctxKV, w.opCount, db)
}

func (w *worker) RunPostCheck(ctx context.Context, dbName string, resChan chan int) {
	w.wl.PostCheck(ctx, w.wrappedDBMap[dbName], resChan)
}
