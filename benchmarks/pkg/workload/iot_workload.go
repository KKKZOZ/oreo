package workload

import (
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"benchmark/ycsb"
	"context"
	"fmt"
	"sync"
)

type IotWorkload struct {
	mu sync.Mutex

	Randomizer
	taskChooser      *generator.Discrete
	wp               *WorkloadParameter
	seriesCnt        int
	KVRocksNamespace string
	MongoDBNamespace string
	task1Count       int
	task2Count       int
	task3Count       int
}

var _ Workload = (*IotWorkload)(nil)

func NewIotWorkload(wp *WorkloadParameter) *IotWorkload {
	return &IotWorkload{
		mu:               sync.Mutex{},
		Randomizer:       *NewRandomizer(wp),
		taskChooser:      createTaskGenerator(wp),
		wp:               wp,
		seriesCnt:        3,
		KVRocksNamespace: "sensor_data",
		MongoDBNamespace: "processed_data",
	}
}

func (wl *IotWorkload) DataIngestion(ctx context.Context, db ycsb.TransactionDB) {
	sensor_id := wl.NextKeyName()

	db.Start()
	for i := 1; i <= wl.seriesCnt; i++ {
		db.Update(ctx, "KVRocks", fmt.Sprintf("%v:%v:%d", wl.KVRocksNamespace, sensor_id, i), wl.RandomValue())
	}
	db.Commit()
}

func (wl *IotWorkload) DataProcessing(ctx context.Context, db ycsb.TransactionDB) {
	sensor_id := wl.NextKeyName()

	db.Start()

	sum := int64(0)
	for i := 1; i <= wl.seriesCnt; i++ {
		value, err := db.Read(ctx, "KVRocks", fmt.Sprintf("%v:%v:%d", wl.KVRocksNamespace, sensor_id, i))
		if err != nil {
			continue
		}
		sum += util.ToInt(value)
	}
	db.Update(ctx, "MongoDB", fmt.Sprintf("%v:%v", wl.MongoDBNamespace, sensor_id), util.ToString(sum))
	db.Commit()
}

func (wl *IotWorkload) DataQuery(ctx context.Context, db ycsb.TransactionDB) {

	db.Start()
	for i := 1; i <= wl.seriesCnt; i++ {
		sensor_id := wl.NextKeyName()
		db.Read(ctx, "MongoDB", fmt.Sprintf("%v:%v", wl.MongoDBNamespace, sensor_id))
	}
	db.Commit()
}

func (wl *IotWorkload) Load(ctx context.Context, opCount int,
	db ycsb.DB) {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}
	txnDB.Start()
	for i := 0; i < opCount; i++ {
		keyPrefix := fmt.Sprintf("%v:%v", wl.KVRocksNamespace, wl.NextKeyNameFromSequence())
		for j := 1; j <= wl.seriesCnt; j++ {
			key := fmt.Sprintf("%v:%d", keyPrefix, j)
			txnDB.Update(ctx, "KVRocks", key, wl.RandomValue())
		}

	}
	err := txnDB.Commit()
	if err != nil {
		fmt.Printf("Error when committing transaction: %v\n", err)
	}

}

func (wl *IotWorkload) Run(ctx context.Context, opCount int,
	db ycsb.DB) {

	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}
	for i := 0; i < opCount; i++ {
		switch wl.NextTask() {
		case 1:
			wl.DataIngestion(ctx, txnDB)
		case 2:
			wl.DataProcessing(ctx, txnDB)
		case 3:
			wl.DataQuery(ctx, txnDB)
		default:
			panic("Invalid task")
		}
	}

}

func (wl *IotWorkload) Cleanup() {}

func (wl *IotWorkload) NeedPostCheck() bool {
	return true
}

func (wl *IotWorkload) NeedRawDB() bool {
	return false
}

func (wl *IotWorkload) PostCheck(context.Context, ycsb.DB, chan int) {}

func (wl *IotWorkload) DisplayCheckResult() {
	fmt.Printf("Task 1 count: %v\n", wl.task1Count)
	fmt.Printf("Task 2 count: %v\n", wl.task2Count)
	fmt.Printf("Task 3 count: %v\n", wl.task3Count)
}

func (wl *IotWorkload) NextTask() int64 {
	wl.mu.Lock()
	defer wl.mu.Unlock()
	idx := wl.taskChooser.Next(wl.r)
	switch idx {
	case 1:
		wl.task1Count++
	case 2:
		wl.task2Count++
	case 3:
		wl.task3Count++
	default:
		panic("Invalid task")
	}
	return idx
}

func (wl *IotWorkload) RandomValue() string {
	value := wl.r.Intn(10000)
	return util.ToString(value)
}