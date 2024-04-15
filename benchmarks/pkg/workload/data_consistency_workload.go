package workload

import (
	"benchmark/pkg/util"
	"benchmark/ycsb"
	"context"
	"fmt"
	"sync"
)

type DataConsistencyWorkload struct {
	mu                  sync.Mutex
	currentTotalAmount  int
	expectedTotalAmount int

	Randomizer
	wp *WorkloadParameter
}

var _ Workload = (*DataConsistencyWorkload)(nil)

func NewDataConsistencyWorkload(wp *WorkloadParameter) *DataConsistencyWorkload {
	return &DataConsistencyWorkload{
		mu:                  sync.Mutex{},
		expectedTotalAmount: wp.TotalAmount,
		Randomizer:          *NewRandomizer(wp),
		wp:                  wp,
	}
}

func (wl *DataConsistencyWorkload) Load(ctx context.Context, opCount int,
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

func (wl *DataConsistencyWorkload) Run(ctx context.Context, opCount int,
	db ycsb.DB) {
	for i := 0; i <= opCount; i++ {
		_ = wl.doAccountTransaction(ctx, db)
	}
}

func (wl *DataConsistencyWorkload) NeedPostCheck() bool {
	return true
}

func (wl *DataConsistencyWorkload) NeedRawDB() bool {
	return false
}

func (wl *DataConsistencyWorkload) PostCheck(ctx context.Context, db ycsb.DB,
	resChan chan int) {
	totalAmount := 0
	if txnDB, ok := db.(ycsb.TransactionDB); ok {
		txnDB.Start()
	}
	for i := 0; i < wl.wp.RecordCount/wl.wp.PostCheckWorkerThread; i++ {
		dbKey := wl.NextKeyNameFromSequence()
		valueStr, err := db.Read(ctx, wl.wp.TableName, dbKey)
		if err != nil {
			fmt.Printf("Error when reading data: %v\n", err)
		}
		value := util.ToInt(valueStr)
		totalAmount += int(value)
	}
	if txnDB, ok := db.(ycsb.TransactionDB); ok {
		txnDB.Commit()
	}

	wl.mu.Lock()
	wl.currentTotalAmount += totalAmount
	wl.mu.Unlock()
}

func (wl *DataConsistencyWorkload) DisplayCheckResult() {
	fmt.Println("---------------")
	fmt.Printf("%s:\nExpected Amount: %v\nCurrent  Amount: %v\n",
		wl.wp.DBName, wl.expectedTotalAmount, wl.currentTotalAmount)
}

func (wl *DataConsistencyWorkload) doAccountTransaction(ctx context.Context, db ycsb.DB) error {
	transferAmount := wl.wp.TransferAmountPerTxn
	key1 := wl.NextKeyName()
	key2 := wl.NextKeyName()

	if key1 == key2 {
		return nil
	}

	if txnDB, ok := db.(ycsb.TransactionDB); ok {
		txnDB.Start()
	}
	v1, err := db.Read(ctx, wl.wp.TableName, key1)
	if err != nil {
		return err
	}
	v2, err := db.Read(ctx, wl.wp.TableName, key2)
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

	err = db.Update(ctx, wl.wp.TableName, key1, v1)
	if err != nil {
		return err
	}
	err = db.Update(ctx, wl.wp.TableName, key2, v2)
	if err != nil {
		return err
	}

	if txnDB, ok := db.(ycsb.TransactionDB); ok {
		return txnDB.Commit()
	}
	return nil
}
