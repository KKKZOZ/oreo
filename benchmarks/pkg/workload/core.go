package workload

import (
	"benchmark/pkg/generator"
	"benchmark/ycsb"
	"context"
	"math/rand"
)

type operationType int64

const (
	read operationType = iota + 1
	update
	insert
	scan
	readModifyWrite
)

type contextKey string

const stateKey = contextKey("core")

type coreState struct {
	r *rand.Rand
	// fieldNames is a copy of core.fieldNames to be goroutine-local
	fieldNames []string
}

type core struct {
	wp *ycsb.WorkloadParameter

	table string

	keySequence                  ycsb.Generator
	operationChooser             *generator.Discrete
	keyChooser                   ycsb.Generator
	transactionInsertKeySequence *generator.AcknowledgedCounter
	orderedInserts               bool
	recordCount                  int64
	zeroPadding                  int64
	insertionRetryLimit          int64
	insertionRetryInterval       int64
}

func createOperationGenerator(wp *ycsb.WorkloadParameter) *generator.Discrete {
	readProportion := wp.ReadProportion
	updateProportion := wp.UpdateProportion
	insertProportion := wp.InsertProportion
	scanProportion := wp.ScanProportion
	readModifyWriteProportion := wp.ReadModifyWriteProportion

	operationChooser := generator.NewDiscrete()
	if readProportion > 0 {
		operationChooser.Add(readProportion, int64(read))
	}

	if updateProportion > 0 {
		operationChooser.Add(updateProportion, int64(update))
	}

	if insertProportion > 0 {
		operationChooser.Add(insertProportion, int64(insert))
	}

	if scanProportion > 0 {
		operationChooser.Add(scanProportion, int64(scan))
	}

	if readModifyWriteProportion > 0 {
		operationChooser.Add(readModifyWriteProportion, int64(readModifyWrite))
	}

	return operationChooser
}

// DoTransaction implements the Workload DoTransaction interface.
func (c *core) DoTransaction(ctx context.Context, db ycsb.DB) error {
	// state := ctx.Value(stateKey).(*coreState)
	// r := state.r

	// operation := operationType(c.operationChooser.Next(r))
	// switch operation {
	// case read:
	// 	return c.doTransactionRead(ctx, db, state)
	// case update:
	// 	return c.doTransactionUpdate(ctx, db, state)
	// case insert:
	// 	return c.doTransactionInsert(ctx, db, state)
	// case scan:
	// 	return c.doTransactionScan(ctx, db, state)
	// default:
	// 	return c.doTransactionReadModifyWrite(ctx, db, state)
	// }
	return nil
}
