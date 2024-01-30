package client

import (
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"benchmark/ycsb"
	"context"
	"fmt"
	"math/rand"
	"time"
)

type operationType int64

const (
	read operationType = iota + 1
	update
	insert
	scan
	readModifyWrite
)

const (
	MAX_VALUE_LENGTH = 100
)

type Client struct {
	db ycsb.DB
	wp *ycsb.WorkloadParameter

	table string

	r                *rand.Rand
	operationChooser *generator.Discrete
	keyChooser       ycsb.Generator
	keySequence      ycsb.Generator

	zeroPadding int64
}

func NewClient(db ycsb.DB, wp *ycsb.WorkloadParameter) *Client {
	c := &Client{
		db:    db,
		wp:    wp,
		table: wp.TableName,
		r:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	c.operationChooser = createOperationGenerator(wp)

	insertStart := int64(0)
	insertCount := int64(wp.RecordCount) - insertStart

	c.keySequence = generator.NewCounter(insertStart)

	var keyrangeLowerBound int64 = insertStart
	var keyrangeUpperBound int64 = insertStart + insertCount - 1

	insertProportion := wp.InsertProportion
	opCount := wp.OperationCount
	expectedNewKeys := int64(float64(opCount) * insertProportion * 2.0)
	keyrangeUpperBound = insertStart + insertCount + expectedNewKeys
	c.keyChooser = generator.NewScrambledZipfian(keyrangeLowerBound, keyrangeUpperBound, generator.ZipfianConstant)

	return c
}

func (c *Client) RunLoad() {

	for i := 1; i <= c.wp.RecordCount; i++ {
		keyNum := c.keySequence.Next(c.r)
		dbKey := c.buildKeyName(keyNum)
		value := c.buildRandomValue()
		_ = c.db.Insert(context.Background(), c.table, dbKey, value)
	}

}

func (c *Client) RunBenchmark() {

	ctx := context.Background()

	// startTime := time.Now()

	for i := 1; i <= c.wp.OperationCount; i++ {

		// var err error

		operation := operationType(c.operationChooser.Next(c.r))
		switch operation {
		case read:
			_ = c.doRead(ctx, c.db)
		case update:
			_ = c.doUpdate(ctx, c.db)
		case insert:
			_ = c.doInsert(ctx, c.db)
		case scan:
			continue
		default:
			_ = c.doReadModifyWrite(ctx, c.db)
		}
	}

	// endTime := time.Now()

	// elapsedTime := endTime.Sub(startTime)

	// fmt.Printf(" elapsed time: %dμs\n OPS: %f\n",
	// 	elapsedTime.Microseconds(), float64(c.wp.OperationCount)/elapsedTime.Seconds())

}

func (c *Client) doRead(ctx context.Context, db ycsb.DB) error {
	keyNum := c.keyChooser.Next(c.r)
	keyName := c.buildKeyName(keyNum)

	_, err := db.Read(ctx, c.table, keyName)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) doUpdate(ctx context.Context, db ycsb.DB) error {
	keyNum := c.keyChooser.Next(c.r)
	keyName := c.buildKeyName(keyNum)

	value := c.buildRandomValue()

	err := db.Update(ctx, c.table, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) doInsert(ctx context.Context, db ycsb.DB) error {
	keyNum := c.keyChooser.Next(c.r)
	keyName := c.buildKeyName(keyNum)

	value := c.buildRandomValue()

	err := db.Insert(ctx, c.table, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) doReadModifyWrite(ctx context.Context, db ycsb.DB) error {
	keyNum := c.keyChooser.Next(c.r)
	keyName := c.buildKeyName(keyNum)

	value, err := db.Read(ctx, c.table, keyName)
	if err != nil {
		return err
	}

	value = c.buildRandomValue()

	err = db.Update(ctx, c.table, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) buildRandomValue() string {
	len := c.r.Intn(MAX_VALUE_LENGTH) + 1
	buf := make([]byte, len)
	util.RandBytes(c.r, buf)
	return string(buf)
}

func (c *Client) buildKeyName(keyNum int64) string {

	keyNum = util.Hash64(keyNum)

	prefix := "benchmark"
	return fmt.Sprintf("%s%0[3]*[2]d", prefix, keyNum, c.zeroPadding)
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
