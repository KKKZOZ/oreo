package client

import (
	"benchmark/ycsb"
	"context"
	"fmt"
	"os"
)

type worker struct {
	c        *Client
	wp       *ycsb.WorkloadParameter
	threadID int
	db       ycsb.DB

	opCount int
}

func newWorker(c *Client, wp *ycsb.WorkloadParameter, threadID int, threadCount int, workDB ycsb.DB) *worker {
	w := new(worker)
	w.c = c
	w.wp = wp
	w.threadID = threadID

	switch db := workDB.(type) {
	case ycsb.TransactionDB:
		w.db = &TxnDbWrapper{DB: db}
	case ycsb.DB:
		w.db = &DbWrapper{DB: db}
	default:
		fmt.Printf("unknown db type: %T", workDB)
		os.Exit(-1)
	}

	var totalOpCount int
	if w.wp.DoBenchmark {
		totalOpCount = w.wp.OperationCount
	} else {
		totalOpCount = w.wp.RecordCount
	}

	if totalOpCount < threadCount {
		fmt.Printf("totalOpCount(%d/%d): %d should be bigger than threadCount: %d",
			wp.OperationCount,
			wp.RecordCount,
			totalOpCount,
			threadCount)

		os.Exit(-1)
	}

	w.opCount = totalOpCount / threadCount
	if threadID < totalOpCount%threadCount {
		w.opCount++
	}

	return w
}

func (w *worker) RunLoad(ctx context.Context) {

	for i := 0; i <= w.opCount; i++ {

		if txnDB, ok := w.db.(ycsb.TransactionDB); ok {
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
		dbKey := w.c.NextKeyNameFromSequence()
		value := w.c.buildRandomValue()
		_ = w.db.Insert(context.Background(), w.wp.TableName, dbKey, value)
	}
}

func (w *worker) RunBenchmark(ctx context.Context) {

	for i := 0; i <= w.opCount; i++ {

		if txnDB, ok := w.db.(ycsb.TransactionDB); ok {
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
			_ = w.doRead(ctx, w.db)
		case update:
			_ = w.doUpdate(ctx, w.db)
		case insert:
			_ = w.doInsert(ctx, w.db)
		case scan:
			continue
		default:
			_ = w.doReadModifyWrite(ctx, w.db)
		}
	}

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
	value := w.c.buildRandomValue()

	err := db.Update(ctx, w.wp.TableName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (w *worker) doInsert(ctx context.Context, db ycsb.DB) error {
	keyName := w.c.NextKeyName()
	value := w.c.buildRandomValue()

	err := db.Insert(ctx, w.wp.TableName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (w *worker) doReadModifyWrite(ctx context.Context, db ycsb.DB) error {
	keyName := w.c.NextKeyName()
	value := w.c.buildRandomValue()

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
