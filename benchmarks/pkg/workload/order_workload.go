package workload

import (
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"benchmark/ycsb"
	"context"
	"fmt"
	"sync"
)

type OrderWorkload struct {
	mu sync.Mutex

	Randomizer
	taskChooser      *generator.Discrete
	wp               *WorkloadParameter
	MongoDBNamespace string
	CouchDBNamespace string
	RedisNamespace   string
	KVRocksNamespace string
	task1Count       int
	task2Count       int
	task3Count       int
}

var _ Workload = (*OrderWorkload)(nil)

func NewOrderWorkload(wp *WorkloadParameter) *OrderWorkload {
	return &OrderWorkload{
		mu:               sync.Mutex{},
		Randomizer:       *NewRandomizer(wp),
		taskChooser:      createTaskGenerator(wp),
		wp:               wp,
		MongoDBNamespace: "products",
		CouchDBNamespace: "orders",
		RedisNamespace:   "sessions",
		KVRocksNamespace: "inventory",
	}
}

func (wl *OrderWorkload) ProductBrowsing(ctx context.Context, db ycsb.TransactionDB) {

	db.Start()

	session_id := wl.NextKeyName()
	product_id := wl.NextKeyName()
	_, _ = db.Read(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespace, session_id))
	_, _ = db.Read(ctx, "MongoDB", fmt.Sprintf("%v:%v", wl.MongoDBNamespace, product_id))
	_ = db.Update(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespace, session_id), wl.RandomValue())

	db.Commit()
}

func (wl *OrderWorkload) OrderPlacement(ctx context.Context, db ycsb.TransactionDB) {

	orderNum := 1
	session_id := wl.NextKeyName()
	db.Start()

	for i := 1; i <= orderNum; i++ {
		product_id := wl.NextKeyName()
		quantityStr, err := db.Read(ctx, "KVRocks", fmt.Sprintf("%v:%v", wl.KVRocksNamespace, product_id))
		if err != nil {
			return
		}
		quantity := util.ToInt(quantityStr)
		if quantity > 0 {
			quantity--
			db.Update(ctx, "KVRocks", fmt.Sprintf("%v:%v", wl.KVRocksNamespace, product_id), util.ToString(quantity))

			order_id := wl.NextKeyName()
			db.Insert(ctx, "CouchDB", fmt.Sprintf("%v:%v", wl.CouchDBNamespace, order_id), wl.RandomValue())
		}
		db.Update(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespace, session_id), wl.RandomValue())
	}

	db.Commit()

}

func (wl *OrderWorkload) InventoryRestocking(ctx context.Context, db ycsb.TransactionDB) {

	db.Start()
	product_id := wl.NextKeyName()
	quantityStr, err := db.Read(ctx, "KVRocks", fmt.Sprintf("%v:%v", wl.KVRocksNamespace, product_id))
	if err != nil {
		return
	}
	quantity := util.ToInt(quantityStr)
	quantity += 10
	db.Update(ctx, "KVRocks", fmt.Sprintf("%v:%v", wl.KVRocksNamespace, product_id), util.ToString(quantity))
	db.Update(ctx, "MongoDB", fmt.Sprintf("%v:%v", wl.MongoDBNamespace, product_id), wl.RandomValue())
	db.Commit()
}

func (wl *OrderWorkload) Load(ctx context.Context, opCount int,
	db ycsb.DB) {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}
	txnDB.Start()
	for i := 0; i < opCount; i++ {
		key := wl.NextKeyNameFromSequence()
		txnDB.Insert(ctx, "MongoDB", fmt.Sprintf("%v:%v", wl.MongoDBNamespace, key), wl.RandomValue())
		txnDB.Insert(ctx, "KVRocks", fmt.Sprintf("%v:%v", wl.KVRocksNamespace, key), wl.RandomValue())
	}
	err := txnDB.Commit()
	if err != nil {
		fmt.Printf("Error when committing transaction: %v\n", err)
	}

}

func (wl *OrderWorkload) Run(ctx context.Context, opCount int,
	db ycsb.DB) {

	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}
	for i := 0; i < opCount; i++ {
		switch wl.NextTask() {
		case 1:
			wl.ProductBrowsing(ctx, txnDB)
		case 2:
			wl.OrderPlacement(ctx, txnDB)
		case 3:
			wl.InventoryRestocking(ctx, txnDB)
		default:
			panic("Invalid task")
		}
	}

}

func (wl *OrderWorkload) Cleanup() {}

func (wl *OrderWorkload) NeedPostCheck() bool {
	return true
}

func (wl *OrderWorkload) NeedRawDB() bool {
	return false
}

func (wl *OrderWorkload) PostCheck(context.Context, ycsb.DB, chan int) {
}

func (wl *OrderWorkload) DisplayCheckResult() {
	fmt.Printf("Task 1 count: %v\n", wl.task1Count)
	fmt.Printf("Task 2 count: %v\n", wl.task2Count)
	fmt.Printf("Task 3 count: %v\n", wl.task3Count)
}

func (wl *OrderWorkload) NextTask() int64 {
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

func (wl *OrderWorkload) RandomValue() string {
	value := wl.r.Intn(10000)
	return util.ToString(value)
}
