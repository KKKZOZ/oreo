package workload

import (
	"benchmark/pkg/config"
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"benchmark/ycsb"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Randomizer struct {
	mu sync.Mutex

	r                *rand.Rand
	operationChooser *generator.Discrete
	datastoreChooser *generator.Discrete
	keyChooser       ycsb.Generator
	keySequence      ycsb.Generator

	zeroPadding int64
}

func NewRandomizer(wp *WorkloadParameter) *Randomizer {
	insertStart := int64(0)
	insertCount := int64(wp.RecordCount) - insertStart

	var keyrangeLowerBound int64 = insertStart
	var keyrangeUpperBound int64 = insertStart + insertCount - 1

	// fmt.Println("Start NewRandomizer")
	r := &Randomizer{
		mu:               sync.Mutex{},
		r:                rand.New(rand.NewSource(time.Now().UnixNano())),
		operationChooser: createOperationGenerator(wp),
		datastoreChooser: createDatastoreGenerator(wp),
		keySequence:      generator.NewCounter(insertStart),
		keyChooser: generator.NewScrambledZipfian(
			keyrangeLowerBound,
			keyrangeUpperBound,
			config.ZipfianConstant),
	}
	// fmt.Println("NewRandomizer")
	return r
}

func (r *Randomizer) ResetKeySequence() {
	r.mu.Lock()
	r.keySequence = generator.NewCounter(0)
	r.mu.Unlock()
}

func createDatastoreGenerator(wp *WorkloadParameter) *generator.Discrete {
	redis1Proportion := wp.Redis1Proportion
	mongo1Proportion := wp.Mongo1Proportion
	mongo2Proportion := wp.Mongo2Proportion
	couch1Proportion := wp.CouchDBProportion

	datastoreChooser := generator.NewDiscrete()
	if redis1Proportion > 0 {
		datastoreChooser.Add(redis1Proportion, int64(redisDatastore1))
	}

	if mongo1Proportion > 0 {
		datastoreChooser.Add(mongo1Proportion, int64(mongoDatastore1))
	}

	if mongo2Proportion > 0 {
		datastoreChooser.Add(mongo2Proportion, int64(mongoDatastore2))
	}

	if couch1Proportion > 0 {
		datastoreChooser.Add(couch1Proportion, int64(couchDatastore1))
	}

	return datastoreChooser
}

func createTaskGenerator(wp *WorkloadParameter) *generator.Discrete {

	taskChooser := generator.NewDiscrete()
	if wp.Task1Proportion > 0 {
		taskChooser.Add(wp.Task1Proportion, 1)
	}

	if wp.Task2Proportion > 0 {
		taskChooser.Add(wp.Task2Proportion, 2)
	}

	if wp.Task3Proportion > 0 {
		taskChooser.Add(wp.Task3Proportion, 3)
	}

	return taskChooser
}

func createOperationGenerator(wp *WorkloadParameter) *generator.Discrete {
	readProportion := wp.ReadProportion
	updateProportion := wp.UpdateProportion
	insertProportion := wp.InsertProportion
	scanProportion := wp.ScanProportion
	readModifyWriteProportion := wp.ReadModifyWriteProportion
	doubleSeqCommitProportion := wp.DoubleSeqCommitProportion

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

	if doubleSeqCommitProportion > 0 {
		operationChooser.Add(doubleSeqCommitProportion, int64(doubleSeqCommit))
	}

	return operationChooser
}

func (r *Randomizer) NextOperation() operationType {
	r.mu.Lock()
	defer r.mu.Unlock()
	return operationType(r.operationChooser.Next(r.r))
}

func (r *Randomizer) NextDatastore() datastoreType {
	r.mu.Lock()
	defer r.mu.Unlock()
	return datastoreType(r.datastoreChooser.Next(r.r))
}

func (r *Randomizer) NextKeyName() string {
	r.mu.Lock()
	keyNum := r.keyChooser.Next(r.r)
	r.mu.Unlock()
	return r.buildKeyName(keyNum)
}

func (r *Randomizer) NextKeyNameFromSequence() string {
	r.mu.Lock()
	keyNum := r.keySequence.Next(r.r)
	r.mu.Unlock()
	return r.buildKeyName(keyNum)
}

func (r *Randomizer) buildKeyName(keyNum int64) string {

	prefix := "benchmark"
	return fmt.Sprintf("%s%0[3]*[2]d", prefix, keyNum, r.zeroPadding)
}
func (r *Randomizer) BuildRandomValue() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	len := MAX_VALUE_LENGTH + 1
	buf := make([]byte, len)
	util.RandBytes(r.r, buf)
	return string(buf)
}
