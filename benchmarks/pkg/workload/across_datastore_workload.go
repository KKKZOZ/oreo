package workload

import (
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"benchmark/ycsb"
	"context"
	"fmt"
	"sync"
)

type AcrossDatastoreWorkload struct {
	mu                 sync.Mutex
	currentTotalAmount map[string]int

	Randomizer
	wp               *WorkloadParameter
	datastoreChooser *generator.Discrete
	internalDBName   []string
}

var _ Workload = (*AcrossDatastoreWorkload)(nil)

func NewAcrossDatastoreWorkload(wp *WorkloadParameter) *AcrossDatastoreWorkload {

	amountMap := make(map[string]int)
	amountMap["redis"] = 0
	amountMap["mongo"] = 0

	return &AcrossDatastoreWorkload{
		mu:                 sync.Mutex{},
		currentTotalAmount: amountMap,
		Randomizer:         *NewRandomizer(wp),
		datastoreChooser:   createDatastoreGenerator(wp),
		wp:                 wp,
		internalDBName:     []string{"redis", "mongo"},
	}
}

// DBType:
// + Oreo
// + Redis-Mongo
func (wl *AcrossDatastoreWorkload) Load(ctx context.Context, opCount int,
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
		value := util.ToString(wl.wp.InitialAmountPerKey)

		for _, dbName := range wl.internalDBName {
			err := db.Insert(ctx, dbName, dbKey, value)
			if err != nil {
				fmt.Printf("Error when loading data: %v\n", err)
			}
		}

	}
	if txnDB, ok := db.(ycsb.TransactionDB); ok {
		err := txnDB.Commit()
		if err != nil {
			fmt.Printf("Error when committing transaction: %v\n", err)
		}
	}
}

func (wl *AcrossDatastoreWorkload) Run(ctx context.Context, opCount int,
	db ycsb.DB) {
	for i := 0; i < opCount; i++ {
		if wl.wp.DBName == "oreo" {
			_ = wl.doInOreo(ctx, db)
		} else {
			_ = wl.doInOthers(ctx, db)
		}
	}
}

func (wl *AcrossDatastoreWorkload) NeedPostCheck() bool {
	return true
}

func (wl *AcrossDatastoreWorkload) NeedRawDB() bool {
	return false
}

func (wl *AcrossDatastoreWorkload) PostCheck(ctx context.Context, db ycsb.DB,
	resChan chan int) {
	redisAmount := 0
	mongoAmount := 0
	if txnDB, ok := db.(ycsb.TransactionDB); ok {
		txnDB.Start()
	}
	for i := 0; i < wl.wp.RecordCount/wl.wp.PostCheckWorkerThread; i++ {
		dbKey := wl.NextKeyNameFromSequence()

		for _, dbName := range wl.internalDBName {
			valueStr, err := db.Read(ctx, dbName, dbKey)
			if err != nil {
				fmt.Printf("Error when reading data: %v\n", err)
			}
			value := util.ToInt(valueStr)
			if dbName == "redis" {
				redisAmount += int(value)
			} else if dbName == "mongo" {
				mongoAmount += int(value)
			}
		}
	}
	if txnDB, ok := db.(ycsb.TransactionDB); ok {
		txnDB.Commit()
	}

	wl.mu.Lock()
	wl.currentTotalAmount["redis"] += redisAmount
	wl.currentTotalAmount["mongo"] += mongoAmount
	wl.mu.Unlock()
}

func (wl *AcrossDatastoreWorkload) DisplayCheckResult() {
	fmt.Println("---------------")
	total := 0
	for dbName, amount := range wl.currentTotalAmount {
		total += amount
		fmt.Printf("%s:\nCurrent  Amount: %v\n",
			dbName, amount)
	}
	fmt.Printf("Total Amount: %v\n", total)
}

func (wl *AcrossDatastoreWorkload) doInOreo(ctx context.Context, db ycsb.DB) error {
	transferAmount := wl.wp.TransferAmountPerTxn
	key1 := wl.NextKeyName()
	db1 := wl.NextDatastore()
	key2 := wl.NextKeyName()
	db2 := wl.NextDatastore()
	if key1 == key2 {
		return nil
	}
	txnDB, ok := db.(ycsb.TransactionDB)
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
	v1Num := util.ToInt(v1)
	v2Num := util.ToInt(v2)
	if v1Num < v2Num {
		v1Num -= int64(transferAmount)
		v2Num += int64(transferAmount)
	} else {
		v1Num += int64(transferAmount)
		v2Num -= int64(transferAmount)
	}
	v1 = fmt.Sprintf("%d", v1Num)
	v2 = fmt.Sprintf("%d", v2Num)

	_ = txnDB.Update(ctx, db1, key1, v1)
	_ = txnDB.Update(ctx, db2, key2, v2)
	return txnDB.Commit()
}

func (wl *AcrossDatastoreWorkload) doInOthers(ctx context.Context, db ycsb.DB) error {

	dbSet := db

	transferAmount := wl.wp.TransferAmountPerTxn
	key1 := wl.NextKeyName()
	db1 := wl.NextDatastore()
	key2 := wl.NextKeyName()
	db2 := wl.NextDatastore()
	if key1 == key2 {
		return nil
	}

	v1, err := dbSet.Read(ctx, db1, key1)
	if err != nil {
		return err
	}
	v2, err := dbSet.Read(ctx, db2, key2)
	if err != nil {
		return err
	}

	v1Num := util.ToInt(v1)
	v2Num := util.ToInt(v2)
	if v1Num < v2Num {
		v1Num -= int64(transferAmount)
		v2Num += int64(transferAmount)
	} else {
		v1Num += int64(transferAmount)
		v2Num -= int64(transferAmount)
	}
	v1 = fmt.Sprintf("%d", v1Num)
	v2 = fmt.Sprintf("%d", v2Num)
	// fmt.Printf("key1: %s v1: %s\nkey2: %s v2: %s\n", key1, v1, key2, v2)
	// v1 = fmt.Sprintf("%d", util.ToInt(v1)-int64(transferAmount))
	// v2 = fmt.Sprintf("%d", util.ToInt(v2)+int64(transferAmount))
	// fmt.Printf("[Updated]key1: %s v1: %s\n[Updated]key2: %s v2: %s\n", key1, v1, key2, v2)

	err = dbSet.Update(ctx, db1, key1, v1)
	if err != nil {
		return err
	}
	err = dbSet.Update(ctx, db2, key2, v2)
	if err != nil {
		return err
	}
	return nil
}

func createDatastoreGenerator(wp *WorkloadParameter) *generator.Discrete {
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

func (wl *AcrossDatastoreWorkload) NextDatastore() string {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	switch datastoreType(wl.datastoreChooser.Next(wl.r)) {
	case redisDatastore:
		return "redis"
	case mongoDatastore:
		return "mongo"
	default:
		return "unknown"
	}
}
