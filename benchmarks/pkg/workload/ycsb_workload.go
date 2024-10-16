package workload

import (
	"benchmark/pkg/measurement"
	"benchmark/ycsb"
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type YCSBWorkload struct {
	Randomizer
	wp *WorkloadParameter

	mu        sync.Mutex
	recordMap map[string]int
}

var _ Workload = (*YCSBWorkload)(nil)

func NewYCSBWorkload(wp *WorkloadParameter) *YCSBWorkload {

	return &YCSBWorkload{
		Randomizer: *NewRandomizer(wp),
		wp:         wp,
		recordMap:  make(map[string]int),
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
	var startTime time.Time
	for i := 0; i <= opCount; i++ {
		if i%wl.wp.TxnOperationGroup == 0 {
			if i != 0 {
				// txnDB.Commit()
				measure(startTime, "TxnGroup", nil)
			}
			if i != opCount {
				// txnDB.Start()
				startTime = time.Now()
			}
		}
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
		case readModifyWrite:
			_ = wl.doReadModifyWrite(ctx, db)
		case doubleSeqCommit:
			_ = wl.doDoubleSeqCommit(ctx, db)
		default:
			panic("Unknown operation")
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
	type pair struct {
		key   string
		value int
	}
	pairs := make([]pair, 0)
	for k, v := range wl.recordMap {
		pairs = append(pairs, pair{k, v})
	}
	// sort pairs by value in descending order
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value > pairs[j].value
	})

	topN := 20
	for i := 0; i < topN && i < len(pairs); i++ {
		fmt.Printf("Key: %s, Count: %d\n", pairs[i].key, pairs[i].value)
	}

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

func (wl *YCSBWorkload) doDoubleSeqCommit(ctx context.Context, db ycsb.DB) error {
	if txnDB, ok := db.(ycsb.TransactionDB); ok {
		newDB := txnDB.NewTransaction()
		keyName := "benchmark211"
		value := wl.BuildRandomValue()

		err := newDB.Start()
		if err != nil {
			return err
		}

		_, err = db.Read(ctx, wl.wp.TableName, keyName)
		if err != nil {
			return err
		}

		err = db.Update(ctx, wl.wp.TableName, keyName, value)
		if err != nil {
			return err
		}

		err = newDB.Commit()
		if err != nil {
			return err
		}

		// Second commit
		err = newDB.Start()
		if err != nil {
			return err
		}

		_, err = db.Read(ctx, wl.wp.TableName, keyName)
		if err != nil {
			return err
		}

		err = db.Update(ctx, wl.wp.TableName, keyName, value)
		if err != nil {
			return err
		}

		err = newDB.Commit()
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("DB does not support transaction")
}

func (wl *YCSBWorkload) NextKeyName() string {
	keyName := wl.Randomizer.NextKeyName()

	wl.mu.Lock()
	defer wl.mu.Unlock()
	wl.recordMap[keyName]++
	return keyName
}

func measure(start time.Time, op string, err error) {
	lan := time.Since(start)
	if err != nil {
		measurement.Measure(fmt.Sprintf("%s_ERROR", op), start, lan)
		return
	}

	measurement.Measure(op, start, lan)
	measurement.Measure("TOTAL", start, lan)
}
