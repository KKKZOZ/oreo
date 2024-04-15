package workload

import (
	"benchmark/ycsb"
	"context"
	"fmt"
)

type YCSBWorkload struct {
	Randomizer
	wp *WorkloadParameter
}

var _ Workload = (*YCSBWorkload)(nil)

func NewYCSBWorkload(wp *WorkloadParameter) *YCSBWorkload {

	return &YCSBWorkload{
		Randomizer: *NewRandomizer(wp),
		wp:         wp,
	}
}

func (wl *YCSBWorkload) Load(ctx context.Context, opCount int,
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

func (wl *YCSBWorkload) Run(ctx context.Context, opCount int, db ycsb.DB) {
	for i := 0; i <= opCount; i++ {
		if txnDB, ok := db.(ycsb.TransactionDB); ok {
			if i%wl.wp.TxnOperationGroup == 0 {
				if i != 0 {
					txnDB.Commit()
				}
				if i != opCount {
					txnDB.Start()
				}
			}
		}
		if i == opCount {
			break
		}

		operation := wl.NextOperation()
		switch operation {
		case read:
			_ = wl.doRead(ctx, db)
		case update:
			_ = wl.doUpdate(ctx, db)
		case insert:
			_ = wl.doInsert(ctx, db)
		case scan:
			continue
		default:
			_ = wl.doReadModifyWrite(ctx, db)
		}
	}
}

func (wl *YCSBWorkload) NeedPostCheck() bool {
	return false
}

func (wl *YCSBWorkload) NeedRawDB() bool {
	return false
}

func (wl *YCSBWorkload) PostCheck(ctx context.Context, db ycsb.DB,
	resChan chan int) {
}

func (wl *YCSBWorkload) DisplayCheckResult() {
}

func (wl *YCSBWorkload) doRead(ctx context.Context, db ycsb.DB) error {
	keyName := wl.NextKeyName()

	_, err := db.Read(ctx, wl.wp.TableName, keyName)
	if err != nil {
		return err
	}
	return nil
}

func (wl *YCSBWorkload) doUpdate(ctx context.Context, db ycsb.DB) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue()

	err := db.Update(ctx, wl.wp.TableName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (wl *YCSBWorkload) doInsert(ctx context.Context, db ycsb.DB) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue()

	err := db.Insert(ctx, wl.wp.TableName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (wl *YCSBWorkload) doReadModifyWrite(ctx context.Context, db ycsb.DB) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue()

	_, err := db.Read(ctx, wl.wp.TableName, keyName)
	if err != nil {
		return err
	}

	err = db.Update(ctx, wl.wp.TableName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}
