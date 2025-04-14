package client

import (
	"benchmark/pkg/workload"
	"benchmark/ycsb"
	"context"
	"fmt"
	"os"
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
	workDBMap map[string]ycsb.DB) *worker {

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

	w.opCount = totalOpCount / threadCount
	if threadID < totalOpCount%threadCount {
		w.opCount++
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

	ctxKV := context.WithValue(ctx, "threadID", w.threadID)
	w.wl.Run(ctxKV, w.opCount, db)
}

func (w *worker) RunPostCheck(ctx context.Context, dbName string, resChan chan int) {
	w.wl.PostCheck(ctx, w.wrappedDBMap[dbName], resChan)
}
