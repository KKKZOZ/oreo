package client

import (
	mongoDB "benchmark/db/mongo"
	"benchmark/db/oreo"
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
	c            *Client
	wp           *ycsb.WorkloadParameter
	threadID     int
	wrappedDBMap map[string]ycsb.DB
	originDBMap  map[string]ycsb.DB

	opCount int
}

func newWorker(c *Client,
	wp *ycsb.WorkloadParameter,
	threadID int, threadCount int,
	workDBMap map[string]ycsb.DB) *worker {
	w := new(worker)
	w.c = c
	w.wp = wp
	w.threadID = threadID
	w.originDBMap = workDBMap
	w.wrappedDBMap = make(map[string]ycsb.DB)

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

	if txnDB, ok := w.wrappedDBMap[dbName].(ycsb.TransactionDB); ok {
		txnDB.Start()
	}
	recordList := make([]string, 0)
	for i := 0; i < w.opCount; i++ {
		dbKey := w.c.NextKeyNameFromSequence()
		recordList = append(recordList, dbKey)
		var value string
		if w.wp.DataConsistencyTest || w.wp.AcrossDatastoreTest {
			value = util.ToString(w.wp.InitialAmountPerKey)
		} else {
			value = w.c.BuildRandomValue()
		}
		err := w.wrappedDBMap[dbName].Insert(context.Background(), w.wp.TableName, dbKey, value)
		if err != nil {
			fmt.Printf("Error when loading data: %v\n", err)
		}
	}
	if txnDB, ok := w.wrappedDBMap[dbName].(ycsb.TransactionDB); ok {
		err := txnDB.Commit()
		if err != nil {
			fmt.Printf("Error when committing transaction: %v\n", err)
		}
	}
	// fmt.Printf("recordList: %v\n", recordList)
}

func (w *worker) RunBenchmark(ctx context.Context, dbName string) {

	if w.wp.DataConsistencyTest {
		for i := 0; i <= w.opCount; i++ {
			_ = w.doAccountTransaction(ctx, w.wrappedDBMap[dbName])
		}
		return
	}

	if w.wp.TxnPerformanceTest {
		for i := 0; i < w.opCount; i++ {
			_ = w.doTxnPerformanceTest(ctx, w.originDBMap[dbName])
		}
		return
	}

	if w.wp.AcrossDatastoreTest {
		for i := 0; i < w.opCount; i++ {
			_ = w.doAccountTransactionAcrossDatastores(ctx)
		}
		return
	}

	// normal ycsb workload benchmark
	for i := 0; i <= w.opCount; i++ {
		if txnDB, ok := w.wrappedDBMap[dbName].(ycsb.TransactionDB); ok {
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
			_ = w.doRead(ctx, w.wrappedDBMap[dbName])
		case update:
			_ = w.doUpdate(ctx, w.wrappedDBMap[dbName])
		case insert:
			_ = w.doInsert(ctx, w.wrappedDBMap[dbName])
		case scan:
			continue
		default:
			_ = w.doReadModifyWrite(ctx, w.wrappedDBMap[dbName])
		}
	}
}

func (w *worker) RunPostCheck(ctx context.Context, dbName string, resChan chan int) {

	recordList := make([]string, 0)
	totalAmount := 0
	if txnDB, ok := w.wrappedDBMap[dbName].(ycsb.TransactionDB); ok {
		txnDB.Start()
	}
	for i := 0; i < w.wp.RecordCount/w.wp.PostCheckWorkerThread; i++ {
		dbKey := w.c.NextKeyNameFromSequence()
		recordList = append(recordList, dbKey)
		valueStr, err := w.wrappedDBMap[dbName].Read(ctx, w.wp.TableName, dbKey)
		if err != nil {
			fmt.Printf("Error when reading data: %v\n", err)
		}
		value := util.ToInt(valueStr)
		totalAmount += int(value)
	}
	if txnDB, ok := w.wrappedDBMap[dbName].(ycsb.TransactionDB); ok {
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
	dbName := w.wp.DBName
	key1 := w.c.NextKeyName()
	key2 := w.c.NextKeyName()

	if key1 == key2 {
		return nil
	}

	if txnDB, ok := w.wrappedDBMap[dbName].(ycsb.TransactionDB); ok {
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

	if txnDB, ok := w.wrappedDBMap[dbName].(ycsb.TransactionDB); ok {
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
		m, ok := w.originDBMap[w.wp.DBName].(*mongoDB.Mongo)
		if !ok {
			panic("Unsupport db type")
		}
		db := m.Client.Database("oreo")
		coll := db.Collection("benchmark")

		session, err := m.Client.StartSession()
		if err != nil {
			fmt.Printf("StartSession error: %v\n", err)
			return err
		}
		defer session.EndSession(ctx)

		callback := func(sessCtx mongo.SessionContext) (interface{}, error) {

			keys := make([]string, opCount)
			values := make([]string, opCount)
			for i := 0; i < opCount; i++ {
				keys[i] = "/" + w.c.NextKeyName()
				values[i] = w.c.BuildRandomValue()

				// Read
				var result mongoDB.MyDocument
				err := coll.FindOne(sessCtx, bson.M{"_id": keys[i]}).Decode(&result)
				if err != nil {
					fmt.Printf("FindOne error: %v\n", err)
					return nil, err
				}
			}
			fmt.Printf("Read done: %v\n", time.Since(start))
			// mid := time.Now()

			for i := 0; i < opCount; i++ {
				_, err = coll.UpdateOne(sessCtx,
					bson.M{"_id": keys[i]},
					bson.M{"$set": bson.M{"valueField": values[i]}})
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
		txnDB, ok := w.wrappedDBMap[w.wp.DBName].(ycsb.TransactionDB)
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

func (w *worker) doAccountTransactionAcrossDatastores(ctx context.Context) error {
	key1 := w.c.NextKeyName()
	db1 := w.c.NextDatastore()
	key2 := w.c.NextKeyName()
	db2 := w.c.NextDatastore()

	if key1 == key2 {
		return nil
	}

	fmt.Printf("key1: %s db1: %s\nkey2: %s db2: %s\n", key1, db1, key2, db2)

	// special case
	if db, ok := w.originDBMap["oreo"]; ok {
		txnDB, ok := db.(*oreo.OreoDatastore)
		if !ok {
			panic("Unsupport db type")
		}
		txnDB.Start()
		v1, err := txnDB.Read(ctx, db1, key1)
		if err != nil {
			return err
		}
		v2, err := txnDB.Read(ctx, db2, key2)
		if err != nil {
			return err
		}
		v1 = fmt.Sprintf("%d", util.ToInt(v1)-1)
		v2 = fmt.Sprintf("%d", util.ToInt(v2)+1)
		_ = txnDB.Update(ctx, db1, key1, v1)
		_ = txnDB.Update(ctx, db2, key2, v2)
		return txnDB.Commit()
	}

	v1, err := w.wrappedDBMap[db1].Read(ctx, w.wp.TableName, key1)
	if err != nil {
		return err
	}
	v2, err := w.wrappedDBMap[db2].Read(ctx, w.wp.TableName, key2)
	if err != nil {
		return err
	}
	// fmt.Printf("key1: %s v1: %s\nkey2: %s v2: %s\n", key1, v1, key2, v2)
	v1 = fmt.Sprintf("%d", util.ToInt(v1)-1)
	v2 = fmt.Sprintf("%d", util.ToInt(v2)+1)
	// fmt.Printf("[Updated]key1: %s v1: %s\n[Updated]key2: %s v2: %s\n", key1, v1, key2, v2)

	err = w.wrappedDBMap[db1].Update(ctx, w.wp.TableName, key1, v1)
	if err != nil {
		return err
	}
	err = w.wrappedDBMap[db1].Update(ctx, w.wp.TableName, key2, v2)
	if err != nil {
		return err
	}
	return nil
}
