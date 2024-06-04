package workload

import (
	"benchmark/ycsb"
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type MultiYCSBWorkload struct {
	Randomizer
	wp *WorkloadParameter

	mu        sync.Mutex
	recordMap map[string]int
}

var _ Workload = (*YCSBWorkload)(nil)

func NewMultiYCSBWorkload(wp *WorkloadParameter) *MultiYCSBWorkload {

	return &MultiYCSBWorkload{
		Randomizer: *NewRandomizer(wp),
		wp:         wp,
		recordMap:  make(map[string]int),
	}
}

func (wl *MultiYCSBWorkload) Load(ctx context.Context, opCount int,
	db ycsb.DB) {
	panic("please load data to each datastore separately")
}

func (wl *MultiYCSBWorkload) Run(ctx context.Context, opCount int, db ycsb.DB) {
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

		dsType := wl.NextDatastore()
		dsName := wl.datastoreTypeToName(dsType)
		operation := wl.NextOperation()
		switch operation {
		case read:
			_ = wl.doRead(ctx, db, dsName)
		case update:
			_ = wl.doUpdate(ctx, db, dsName)
		case insert:
			_ = wl.doInsert(ctx, db, dsName)
		case scan:
			continue
		case readModifyWrite:
			_ = wl.doReadModifyWrite(ctx, db, dsName)
		default:
			panic("Unknown operation")
		}
	}
}

func (wl *MultiYCSBWorkload) NeedPostCheck() bool {
	return false
}

func (wl *MultiYCSBWorkload) NeedRawDB() bool {
	return false
}

func (wl *MultiYCSBWorkload) PostCheck(ctx context.Context, db ycsb.DB,
	resChan chan int) {
}

func (wl *MultiYCSBWorkload) DisplayCheckResult() {
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

func (wl *MultiYCSBWorkload) doRead(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName()

	_, err := db.Read(ctx, dsName, keyName)
	if err != nil {
		return err
	}
	return nil
}

func (wl *MultiYCSBWorkload) doUpdate(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue()

	err := db.Update(ctx, dsName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (wl *MultiYCSBWorkload) doInsert(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue()

	err := db.Insert(ctx, dsName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (wl *MultiYCSBWorkload) doReadModifyWrite(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue()

	// fmt.Printf("Key: %v dsName: %v\n", keyName, dsName)

	_, err := db.Read(ctx, dsName, keyName)
	if err != nil {
		return err
	}

	err = db.Update(ctx, dsName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (wl *MultiYCSBWorkload) doDoubleSeqCommit(ctx context.Context, db ycsb.DB) error {
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

func (wl *MultiYCSBWorkload) NextKeyName() string {
	keyName := wl.Randomizer.NextKeyName()

	wl.mu.Lock()
	defer wl.mu.Unlock()
	wl.recordMap[keyName]++
	return keyName
}

func (wl *MultiYCSBWorkload) datastoreTypeToName(dsType datastoreType) string {
	switch dsType {
	case redisDatastore1:
		return "redis1"
	case mongoDatastore1:
		return "mongo1"
	case mongoDatastore2:
		return "mongo2"
	case couchDatastore1:
		return "couchdb"
	default:
		return ""
	}
}
