package workload

import (
	"benchmark/ycsb"
	"context"
	"fmt"

	mongoDB "benchmark/db/mongo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	rds "benchmark/db/redis"

	goredis "github.com/redis/go-redis/v9"
)

type TxnPerformanceWorkload struct {
	Randomizer
	wp *WorkloadParameter
}

var _ Workload = (*TxnPerformanceWorkload)(nil)

func NewTxnPerformanceWorkload(wp *WorkloadParameter) *TxnPerformanceWorkload {
	return &TxnPerformanceWorkload{
		Randomizer: *NewRandomizer(wp),
		wp:         wp,
	}
}

func (wl *TxnPerformanceWorkload) Load(ctx context.Context, opCount int,
	db ycsb.DB) {
	if txnDB, ok := db.(ycsb.TransactionDB); ok {
		err := txnDB.Start()
		if err != nil {
			fmt.Printf("Error when loading data: %v\n", err)
			return
		}
	}
	for i := 0; i < opCount; i++ {
		dbKey := wl.NextKeyNameFromSequence()
		value := wl.BuildRandomValue()

		err := db.Insert(context.Background(), wl.wp.TableName, dbKey, value)
		if err != nil {
			fmt.Printf("Error when loading data: %v\n", err)
		}
	}
	if txnDB, ok := db.(ycsb.TransactionDB); ok {
		err := txnDB.Commit()
		if err != nil {
			fmt.Printf("Error when committing transaction: %v\n", err)
		}
	}
}

func (wl *TxnPerformanceWorkload) Run(ctx context.Context, opCount int,
	db ycsb.DB) {
	for i := 0; i < opCount; i++ {
		_ = wl.doTxnPerformanceTest(ctx, db)
	}
}

func (wl *TxnPerformanceWorkload) NeedPostCheck() bool {
	return false
}

func (wl *TxnPerformanceWorkload) NeedRawDB() bool {
	return true
}

func (wl *TxnPerformanceWorkload) PostCheck(ctx context.Context, db ycsb.DB,
	resChan chan int) {
}

func (wl *TxnPerformanceWorkload) DisplayCheckResult() {
}

// Only supports Read-Modify-Write pattern
func (wl *TxnPerformanceWorkload) doTxnPerformanceTest(ctx context.Context, db ycsb.DB) error {
	opCount := 5

	switch wl.wp.DBName {
	case "redis":
		return wl.doInRedis(ctx, db, opCount)
	case "mongo":
		return wl.doInMongo(ctx, db, opCount)
	case "oreo-redis":
		return wl.doInOreoRedis(ctx, db, opCount)
	default:
		panic("Unsupport db type")
	}
}

func (wl *TxnPerformanceWorkload) doInOreoRedis(ctx context.Context, db ycsb.DB, opCount int) error {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		panic("Unsupport db type")
	}

	txnDB.Start()
	for i := 0; i < opCount; i++ {
		keyName := wl.NextKeyName()
		value := wl.BuildRandomValue()

		_, err := txnDB.Read(ctx, wl.wp.TableName, keyName)
		if err != nil {
			return err
		}

		err = txnDB.Update(ctx, wl.wp.TableName, keyName, value)
		if err != nil {
			return err
		}
	}
	return txnDB.Commit()
}

func (wl *TxnPerformanceWorkload) doInMongo(ctx context.Context, db ycsb.DB, opCount int) error {
	m, ok := db.(*mongoDB.Mongo)
	if !ok {
		panic("Unsupport db type")
	}
	mdb := m.Client.Database("oreo")
	coll := mdb.Collection("benchmark")

	sessionOpts := options.Session().
		SetDefaultReadConcern(readconcern.Linearizable()).
		SetDefaultWriteConcern(writeconcern.Majority())

	session, err := m.Client.StartSession(sessionOpts)

	if err != nil {
		fmt.Printf("StartSession error: %v\n", err)
		return err
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {

		keys := make([]string, opCount)
		values := make([]string, opCount)
		for i := 0; i < opCount; i++ {
			keys[i] = "/" + wl.NextKeyName()
			values[i] = wl.BuildRandomValue()

			// Read
			var result mongoDB.MyDocument
			err := coll.FindOne(sessCtx, bson.M{"_id": keys[i]}).Decode(&result)
			if err != nil {
				fmt.Printf("FindOne error: %v\n", err)
				return nil, err
			}
		}

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

func (wl *TxnPerformanceWorkload) doInRedis(ctx context.Context, db ycsb.DB, opCount int) error {
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
			keys[i] = "/" + wl.NextKeyName()
			values[i] = wl.BuildRandomValue()

			_, err := tx.Get(ctx, keys[i]).Result()
			if err != nil {
				fmt.Printf("Get error: %v\n", err)
				return err
			}
		}

		// fmt.Printf("Read done: %v\n", time.Since(start))
		// mid := time.Now()

		// Operation is commited only if the watched keys remain unchanged.
		_, err := tx.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
			for i := 0; i < opCount; i++ {
				pipe.Set(ctx, keys[i], values[i], 0)
			}
			return nil
		})

		// fmt.Printf("TxPipelined done: %v\n", time.Since(mid))
		if err != nil {
			fmt.Printf("TxPipelined error: %v\n", err)
		}
		return err
	}

	err := rdb.Watch(ctx, txnf, keys...)
	return err

}
