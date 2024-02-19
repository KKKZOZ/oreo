package client

import (
	mongoDB "benchmark/db/mongo"
	rds "benchmark/db/redis"
	"benchmark/pkg/util"
	"benchmark/ycsb"
	"context"
	"fmt"
	"os"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type worker struct {
	c         *Client
	wp        *ycsb.WorkloadParameter
	threadID  int
	wrappedDB ycsb.DB
	originDB  ycsb.DB

	opCount int
}

func newWorker(c *Client, wp *ycsb.WorkloadParameter, threadID int, threadCount int, workDB ycsb.DB) *worker {
	w := new(worker)
	w.c = c
	w.wp = wp
	w.threadID = threadID
	w.originDB = workDB

	switch db := workDB.(type) {
	case ycsb.TransactionDB:
		w.wrappedDB = &TxnDbWrapper{DB: db}
	case ycsb.DB:
		w.wrappedDB = &DbWrapper{DB: db}
	default:
		fmt.Printf("unknown db type: %T", workDB)
		os.Exit(-1)
	}

	var totalOpCount int
	if w.wp.DoBenchmark {
		totalOpCount = w.wp.OperationCount
	} else {
		totalOpCount = w.wp.RecordCount
	}

	if totalOpCount < threadCount {
		fmt.Printf("totalOpCount(%d/%d): %d should be bigger than threadCount: %d",
			wp.OperationCount,
			wp.RecordCount,
			totalOpCount,
			threadCount)

		os.Exit(-1)
	}

	w.opCount = totalOpCount / threadCount
	if threadID < totalOpCount%threadCount {
		w.opCount++
	}

	return w
}

func (w *worker) RunLoad(ctx context.Context) {
	if txnDB, ok := w.wrappedDB.(ycsb.TransactionDB); ok {
		txnDB.Start()
	}
	recordList := make([]string, 0)
	for i := 0; i < w.opCount; i++ {
		dbKey := w.c.NextKeyNameFromSequence()
		recordList = append(recordList, dbKey)
		var value string
		if w.wp.DataConsistencyTest {
			value = util.ToString(w.wp.InitialAmountPerKey)
		} else {
			value = w.c.BuildRandomValue()
		}
		_ = w.wrappedDB.Insert(context.Background(), w.wp.TableName, dbKey, value)
	}
	if txnDB, ok := w.wrappedDB.(ycsb.TransactionDB); ok {
		txnDB.Commit()
	}
	// fmt.Printf("recordList: %v\n", recordList)
}

func (w *worker) RunBenchmark(ctx context.Context) {

	if w.wp.DataConsistencyTest {
		for i := 0; i <= w.opCount; i++ {
			_ = w.doAccountTransaction(ctx, w.wrappedDB)
		}
		return
	}

	if w.wp.TxnPerformanceTest {
		for i := 0; i < w.opCount; i++ {
			_ = w.doTxnPerformanceTest(ctx, w.originDB)
		}
		return
	}

	for i := 0; i <= w.opCount; i++ {
		if txnDB, ok := w.wrappedDB.(ycsb.TransactionDB); ok {
			if i%w.wp.TxnOperationGroup == 0 {
				if i != 0 {
					txnDB.Commit()
				}
				if i != w.opCount {
					txnDB.Start()
				}
			}
		}
		if i == w.opCount {
			break
		}

		operation := w.c.NextOperation()
		switch operation {
		case read:
			_ = w.doRead(ctx, w.wrappedDB)
		case update:
			_ = w.doUpdate(ctx, w.wrappedDB)
		case insert:
			_ = w.doInsert(ctx, w.wrappedDB)
		case scan:
			continue
		default:
			_ = w.doReadModifyWrite(ctx, w.wrappedDB)
		}
	}
}

func (w *worker) RunPostCheck(ctx context.Context, resChan chan int) {
	recordList := make([]string, 0)
	totalAmount := 0
	if txnDB, ok := w.wrappedDB.(ycsb.TransactionDB); ok {
		txnDB.Start()
	}
	for i := 0; i < w.wp.RecordCount/w.wp.PostCheckWorkerThread; i++ {
		dbKey := w.c.NextKeyNameFromSequence()
		recordList = append(recordList, dbKey)
		valueStr, _ := w.wrappedDB.Read(ctx, w.wp.TableName, dbKey)
		value := util.ToInt(valueStr)
		totalAmount += int(value)
	}
	if txnDB, ok := w.wrappedDB.(ycsb.TransactionDB); ok {
		txnDB.Commit()
	}
	// fmt.Printf("recordList: %v\n", recordList)
	resChan <- totalAmount
}

func (w *worker) doRead(ctx context.Context, db ycsb.DB) error {
	keyName := w.c.NextKeyName()

	_, err := db.Read(ctx, w.wp.TableName, keyName)
	if err != nil {
		return err
	}
	return nil
}

func (w *worker) doUpdate(ctx context.Context, db ycsb.DB) error {
	keyName := w.c.NextKeyName()
	value := w.c.BuildRandomValue()

	err := db.Update(ctx, w.wp.TableName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (w *worker) doInsert(ctx context.Context, db ycsb.DB) error {
	keyName := w.c.NextKeyName()
	value := w.c.BuildRandomValue()

	err := db.Insert(ctx, w.wp.TableName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (w *worker) doReadModifyWrite(ctx context.Context, db ycsb.DB) error {
	keyName := w.c.NextKeyName()
	value := w.c.BuildRandomValue()

	_, err := db.Read(ctx, w.wp.TableName, keyName)
	if err != nil {
		return err
	}

	err = db.Update(ctx, w.wp.TableName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (w *worker) doAccountTransaction(ctx context.Context, db ycsb.DB) error {
	key1 := w.c.NextKeyName()
	key2 := w.c.NextKeyName()

	if key1 == key2 {
		return nil
	}

	if txnDB, ok := w.wrappedDB.(ycsb.TransactionDB); ok {
		txnDB.Start()
	}
	v1, err := db.Read(ctx, w.wp.TableName, key1)
	if err != nil {
		return err
	}
	v2, err := db.Read(ctx, w.wp.TableName, key2)
	if err != nil {
		return err
	}
	// fmt.Printf("key1: %s v1: %s\nkey2: %s v2: %s\n", key1, v1, key2, v2)
	v1 = fmt.Sprintf("%d", util.ToInt(v1)-1)
	v2 = fmt.Sprintf("%d", util.ToInt(v2)+1)
	// fmt.Printf("[Updated]key1: %s v1: %s\n[Updated]key2: %s v2: %s\n", key1, v1, key2, v2)

	err = db.Update(ctx, w.wp.TableName, key1, v1)
	if err != nil {
		return err
	}
	err = db.Update(ctx, w.wp.TableName, key2, v2)
	if err != nil {
		return err
	}

	if txnDB, ok := w.wrappedDB.(ycsb.TransactionDB); ok {
		return txnDB.Commit()
	}
	return nil
}

// Only supports Read-Modify-Write pattern
func (w *worker) doTxnPerformanceTest(ctx context.Context, db ycsb.DB) error {

	opCount := 10
	start := time.Now()
	defer func() {
		fmt.Printf("TxnPerformanceTest: %v\n", time.Since(start))
	}()

	if w.wp.DBName == "redis" {
		r, ok := db.(*rds.Redis)
		if !ok {
			panic("Unsupport db type")
		}
		rdb := r.Rdb
		keys := make([]string, opCount)
		values := make([]string, opCount)

		// Transactional function.
		txnf := func(tx *goredis.Tx) error {
			for i := 0; i < opCount; i++ {
				keys[i] = "/" + w.c.NextKeyName()
				values[i] = w.c.BuildRandomValue()

				_, err := tx.Get(ctx, keys[i]).Result()
				if err != nil {
					fmt.Printf("Get error: %v\n", err)
					return err
				}
			}

			fmt.Printf("Read done: %v\n", time.Since(start))
			mid := time.Now()

			// Operation is commited only if the watched keys remain unchanged.
			_, err := tx.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
				for i := 0; i < opCount; i++ {
					pipe.Set(ctx, keys[i], values[i], 0)
				}
				return nil
			})

			fmt.Printf("TxPipelined done: %v\n", time.Since(mid))
			if err != nil {
				fmt.Printf("TxPipelined error: %v\n", err)
			}
			return err
		}

		err := rdb.Watch(ctx, txnf, keys...)
		return err
	}

	if w.wp.DBName == "mongo" {
		m, ok := w.originDB.(*mongoDB.Mongo)
		if !ok {
			panic("Unsupport db type")
		}
		db := m.Client.Database("oreo")
		coll := db.Collection("benchmark")

		session, err := m.Client.StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
			for i := 0; i < opCount; i++ {
				key := w.c.NextKeyName()
				value := w.c.BuildRandomValue()

				// Read
				var result string
				err := coll.FindOne(sessCtx, bson.M{"_id": key}).Decode(&result)
				if err != nil {
					return nil, err
				}

				// Modify - Implement logic based on your application's needs

				// Write
				_, err = coll.UpdateOne(sessCtx, bson.M{"_id": key}, bson.M{"$set": bson.M{"valueField": value}})
				if err != nil {
					return nil, err
				}
			}
			return nil, nil
		}

		_, err = session.WithTransaction(ctx, callback)
		return err

	}

	if w.wp.DBName == "oreo-redis" {
		txnDB, ok := w.wrappedDB.(ycsb.TransactionDB)
		if !ok {
			panic("Unsupport db type")
		}

		txnDB.Start()
		for i := 0; i < 5; i++ {
			keyName := w.c.NextKeyName()
			value := w.c.BuildRandomValue()

			_, err := txnDB.Read(ctx, w.wp.TableName, keyName)
			if err != nil {
				return err
			}

			err = txnDB.Update(ctx, w.wp.TableName, keyName, value)
			if err != nil {
				return err
			}
		}
		return txnDB.Commit()
	}
	return nil
}
