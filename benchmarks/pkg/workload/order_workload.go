package workload

import (
	"context"
	"fmt"
	"log"
	"sync"

	"benchmark/pkg/benconfig"
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"benchmark/ycsb"
)

type OrderWorkload struct {
	mu sync.Mutex

	Randomizer
	taskChooser        *generator.Discrete
	wp                 *WorkloadParameter
	MongoDBNamespace1  string
	MongoDBNamespace2  string
	CassandraNamespace string
	RedisNamespace     string
	KVRocksNamespace   string
	task1Count         int
	task2Count         int
	task3Count         int
	task4Count         int
	task5Count         int
}

var _ Workload = (*OrderWorkload)(nil)

func NewOrderWorkload(wp *WorkloadParameter) *OrderWorkload {
	return &OrderWorkload{
		mu:                 sync.Mutex{},
		Randomizer:         *NewRandomizer(wp),
		taskChooser:        createTaskGenerator(wp),
		wp:                 wp,
		MongoDBNamespace1:  "products",
		MongoDBNamespace2:  "inventory",
		CassandraNamespace: "orders",
		RedisNamespace:     "sessions",
		KVRocksNamespace:   "reviews",
	}
}

func (wl *OrderWorkload) ProductBrowsing(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	session_id := wl.NextKeyName()
	product_id := wl.NextKeyName()
	_, _ = db.Read(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespace, session_id))
	_, _ = db.Read(ctx, "MongoDB", fmt.Sprintf("%v:%v", wl.MongoDBNamespace1, product_id))
	_ = db.Update(
		ctx,
		"Redis",
		fmt.Sprintf("%v:%v", wl.RedisNamespace, session_id),
		wl.RandomValue(),
	)

	db.Commit()
}

func (wl *OrderWorkload) OrderPlacement(ctx context.Context, db ycsb.TransactionDB) {
	orderNum := 2
	session_id := wl.NextKeyName()
	db.Start()

	for i := 1; i <= orderNum; i++ {
		product_id := wl.NextKeyName()
		quantityStr, err := db.Read(
			ctx,
			"MongoDB",
			fmt.Sprintf("%v:%v", wl.MongoDBNamespace2, product_id),
		)
		if err != nil {
			return
		}
		quantity := util.ToInt(quantityStr)
		if quantity > 0 {
			quantity--
			db.Update(
				ctx,
				"MongoDB",
				fmt.Sprintf("%v:%v", wl.MongoDBNamespace2, product_id),
				util.ToString(quantity),
			)

			order_id := wl.NextKeyName()
			db.Insert(
				ctx,
				"Cassandra",
				fmt.Sprintf("%v:%v", wl.CassandraNamespace, order_id),
				wl.RandomValue(),
			)
		}
		db.Update(
			ctx,
			"Redis",
			fmt.Sprintf("%v:%v", wl.RedisNamespace, session_id),
			wl.RandomValue(),
		)
	}

	db.Commit()
}

func (wl *OrderWorkload) InventoryRestocking(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()
	product_id := wl.NextKeyName()
	quantityStr, err := db.Read(
		ctx,
		"MongoDB",
		fmt.Sprintf("%v:%v", wl.MongoDBNamespace2, product_id),
	)
	if err != nil {
		return
	}
	quantity := util.ToInt(quantityStr)
	quantity += 10
	db.Update(
		ctx,
		"MongoDB",
		fmt.Sprintf("%v:%v", wl.MongoDBNamespace2, product_id),
		util.ToString(quantity),
	)
	db.Update(
		ctx,
		"MongoDB",
		fmt.Sprintf("%v:%v", wl.MongoDBNamespace1, product_id),
		wl.RandomValue(),
	)
	db.Commit()
}

func (wl *OrderWorkload) OrderTracking(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	// 获取订单信息
	order_id := wl.NextKeyName()
	session_id := wl.NextKeyName()

	// 从Cassandra读取订单详情
	_, _ = db.Read(ctx, "Cassandra", fmt.Sprintf("%v:%v", wl.CassandraNamespace, order_id))

	// 更新用户会话信息
	db.Update(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespace, session_id), wl.RandomValue())

	db.Commit()
}

func (wl *OrderWorkload) CustomerReview(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()

	order_id := wl.NextKeyName()
	product_id := wl.NextKeyName()
	session_id := wl.NextKeyName()

	// 检查订单是否存在
	_, _ = db.Read(ctx, "Cassandra", fmt.Sprintf("%v:%v", wl.CassandraNamespace, order_id))

	// 更新产品评价信息
	db.Update(
		ctx,
		"KVRocks",
		fmt.Sprintf("%v:%v", wl.KVRocksNamespace, product_id),
		wl.RandomValue(),
	)

	// 更新会话信息
	db.Update(ctx, "Redis", fmt.Sprintf("%v:%v", wl.RedisNamespace, session_id), wl.RandomValue())

	db.Commit()
}

func (wl *OrderWorkload) Load(ctx context.Context, opCount int,
	db ycsb.DB,
) {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}

	if opCount%benconfig.MaxLoadBatchSize != 0 {
		log.Fatalf(
			"opCount should be a multiple of MaxLoadBatchSize, opCount: %d, MaxLoadBatchSize: %d",
			opCount,
			benconfig.MaxLoadBatchSize,
		)
	}

	round := opCount / benconfig.MaxLoadBatchSize
	var aErr error
	for i := 0; i < round; i++ {
		txnDB.Start()
		for j := 0; j < benconfig.MaxLoadBatchSize; j++ {
			key := wl.NextKeyNameFromSequence()
			txnDB.Insert(
				ctx,
				"Redis",
				fmt.Sprintf("%v:%v", wl.RedisNamespace, key),
				wl.RandomValue(),
			)
			txnDB.Insert(
				ctx,
				"MongoDB",
				fmt.Sprintf("%v:%v", wl.MongoDBNamespace1, key),
				wl.RandomValue(),
			)
			txnDB.Insert(
				ctx,
				"MongoDB",
				fmt.Sprintf("%v:%v", wl.MongoDBNamespace2, key),
				wl.RandomValue(),
			)
		}
		err := txnDB.Commit()
		if err != nil {
			aErr = err
			fmt.Printf("Error when committing transaction: %v\n", err)
		}
	}
	if aErr != nil {
		fmt.Printf("Error in Oreo YCSB Load: %v\n", aErr)
	}
}

func (wl *OrderWorkload) Run(ctx context.Context, opCount int,
	db ycsb.DB,
) {
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
		case 4:
			wl.OrderTracking(ctx, txnDB)
		case 5:
			wl.CustomerReview(ctx, txnDB)
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
	fmt.Printf("Task 4 count: %v\n", wl.task4Count)
	fmt.Printf("Task 5 count: %v\n", wl.task5Count)
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
	case 4:
		wl.task4Count++
	case 5:
		wl.task5Count++
	default:
		panic("Invalid task")
	}
	return idx
}

func (wl *OrderWorkload) RandomValue() string {
	value := wl.r.Intn(10000)
	return util.ToString(value)
}
