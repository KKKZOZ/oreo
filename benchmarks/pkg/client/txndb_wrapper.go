package client

import (
	"benchmark/ycsb"
	"context"
	"fmt"
	"time"
)

var _ ycsb.DB = (*TxnDbWrapper)(nil)
var _ ycsb.TransactionDB = (*TxnDbWrapper)(nil)

// TxnDbWrapper stores the pointer to a implementation of ycsb.TransactionDB
type TxnDbWrapper struct {
	DB       ycsb.TransactionDB
	TxnStart time.Time
}

func (db *TxnDbWrapper) Start() (err error) {
	db.TxnStart = time.Now()
	start := time.Now()
	defer func() {
		measure(start, "Start", err)
	}()
	return db.DB.Start()
}

func (db *TxnDbWrapper) Commit() (err error) {
	start := time.Now()
	defer func() {
		measure(start, "COMMIT", err)
		measure(db.TxnStart, "TXN", err)
	}()
	return db.DB.Commit()
}

func (db *TxnDbWrapper) Abort() error {
	return db.DB.Abort()
}

func (db *TxnDbWrapper) Close() error {
	return db.DB.Close()
}

func (db *TxnDbWrapper) InitThread(ctx context.Context, threadID int, threadCount int) context.Context {
	return db.DB.InitThread(ctx, threadID, threadCount)
}

func (db *TxnDbWrapper) CleanupThread(ctx context.Context) {
	db.DB.CleanupThread(ctx)
}

func (db *TxnDbWrapper) Read(ctx context.Context, table string, key string) (_ string, err error) {
	start := time.Now()
	defer func() {
		if err != nil {
			fmt.Println("Error in Read: ", err)
		}
		measure(start, "READ", err)
	}()

	return db.DB.Read(ctx, table, key)
}

func (db *TxnDbWrapper) BatchRead(ctx context.Context, table string, keys []string, fields []string) (_ []map[string][]byte, err error) {
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

func (db *TxnDbWrapper) Update(ctx context.Context, table string, key string, value string) (err error) {
	start := time.Now()
	defer func() {
		measure(start, "UPDATE", err)
	}()

	return db.DB.Update(ctx, table, key, value)
}

func (db *TxnDbWrapper) BatchUpdate(ctx context.Context, table string, keys []string, values []string) (err error) {
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

func (db *TxnDbWrapper) Insert(ctx context.Context, table string, key string, value string) (err error) {
	start := time.Now()
	defer func() {
		measure(start, "INSERT", err)
	}()

	return db.DB.Insert(ctx, table, key, value)
}

func (db *TxnDbWrapper) BatchInsert(ctx context.Context, table string, keys []string, values []string) (err error) {
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

func (db *TxnDbWrapper) Delete(ctx context.Context, table string, key string) (err error) {
	start := time.Now()
	defer func() {
		measure(start, "DELETE", err)
	}()

	return db.DB.Delete(ctx, table, key)
}

func (db *TxnDbWrapper) BatchDelete(ctx context.Context, table string, keys []string) (err error) {
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
