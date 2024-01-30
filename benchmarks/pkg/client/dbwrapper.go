package client

import (
	"benchmark/pkg/measurement"
	"benchmark/ycsb"
	"context"
	"fmt"
	"time"
)

// DbWrapper stores the pointer to a implementation of ycsb.DB.
type DbWrapper struct {
	DB ycsb.DB
}

func measure(start time.Time, op string, err error) {
	lan := time.Since(start)
	if err != nil {
		measurement.Measure(fmt.Sprintf("%s_ERROR", op), start, lan)
		return
	}

	measurement.Measure(op, start, lan)
	measurement.Measure("TOTAL", start, lan)
}

func (db DbWrapper) Close() error {
	return db.DB.Close()
}

func (db DbWrapper) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return db.DB.InitThread(ctx, threadID, threadCount)
}

func (db DbWrapper) CleanupThread(ctx context.Context) {
	db.DB.CleanupThread(ctx)
}

func (db DbWrapper) Read(ctx context.Context, table string, key string) (_ string, err error) {
	start := time.Now()
	defer func() {
		measure(start, "READ", err)
	}()

	return db.DB.Read(ctx, table, key)
}

func (db DbWrapper) BatchRead(ctx context.Context, table string, keys []string, fields []string) (_ []map[string][]byte, err error) {
	batchDB, ok := db.DB.(ycsb.BatchDB)
	if ok {
		start := time.Now()
		defer func() {
			measure(start, "BATCH_READ", err)
		}()
		return batchDB.BatchRead(ctx, table, keys, fields)
	}
	for _, key := range keys {
		_, err := db.DB.Read(ctx, table, key)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// func (db DbWrapper) Scan(ctx context.Context, table string, startKey string, count int, fields []string) (_ []map[string][]byte, err error) {
// 	start := time.Now()
// 	defer func() {
// 		measure(start, "SCAN", err)
// 	}()

// 	return db.DB.Scan(ctx, table, startKey, count, fields)
// }

func (db DbWrapper) Update(ctx context.Context, table string, key string, value string) (err error) {
	start := time.Now()
	defer func() {
		measure(start, "UPDATE", err)
	}()

	return db.DB.Update(ctx, table, key, value)
}

func (db DbWrapper) BatchUpdate(ctx context.Context, table string, keys []string, values []string) (err error) {
	// batchDB, ok := db.DB.(ycsb.BatchDB)
	// if ok {
	// 	start := time.Now()
	// 	defer func() {
	// 		measure(start, "BATCH_UPDATE", err)
	// 	}()
	// 	return batchDB.BatchUpdate(ctx, table, keys, values)
	// }
	for i := range keys {
		err := db.DB.Update(ctx, table, keys[i], values[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (db DbWrapper) Insert(ctx context.Context, table string, key string, value string) (err error) {
	start := time.Now()
	defer func() {
		measure(start, "INSERT", err)
	}()

	return db.DB.Insert(ctx, table, key, value)
}

func (db DbWrapper) BatchInsert(ctx context.Context, table string, keys []string, values []string) (err error) {
	// batchDB, ok := db.DB.(ycsb.BatchDB)
	// if ok {
	// 	start := time.Now()
	// 	defer func() {
	// 		measure(start, "BATCH_INSERT", err)
	// 	}()
	// 	return batchDB.BatchInsert(ctx, table, keys, values)
	// }
	for i := range keys {
		err := db.DB.Insert(ctx, table, keys[i], values[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (db DbWrapper) Delete(ctx context.Context, table string, key string) (err error) {
	start := time.Now()
	defer func() {
		measure(start, "DELETE", err)
	}()

	return db.DB.Delete(ctx, table, key)
}

func (db DbWrapper) BatchDelete(ctx context.Context, table string, keys []string) (err error) {
	batchDB, ok := db.DB.(ycsb.BatchDB)
	if ok {
		start := time.Now()
		defer func() {
			measure(start, "BATCH_DELETE", err)
		}()
		return batchDB.BatchDelete(ctx, table, keys)
	}
	for _, key := range keys {
		err := db.DB.Delete(ctx, table, key)
		if err != nil {
			return err
		}
	}
	return nil
}

// func (db DbWrapper) Analyze(ctx context.Context, table string) error {
// 	if analyzeDB, ok := db.DB.(ycsb.AnalyzeDB); ok {
// 		return analyzeDB.Analyze(ctx, table)
// 	}
// 	return nil
// }
