package workload

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"benchmark/pkg/benconfig"
	"benchmark/pkg/generator"
	"benchmark/pkg/util"
	"benchmark/ycsb"
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
			benconfig.ZipfianConstant),
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
	proportions := map[int64]float64{
		int64(kvrocksDatastore1):   wp.KVRocksProportion,
		int64(redisDatastore1):     wp.Redis1Proportion,
		int64(mongoDatastore1):     wp.Mongo1Proportion,
		int64(mongoDatastore2):     wp.Mongo2Proportion,
		int64(couchDatastore1):     wp.CouchDBProportion,
		int64(cassandraDatastore1): wp.CassandraProportion,
		int64(dynamodbDatastore1):  wp.DynamoDBProportion,
		int64(tikvDatastore1):      wp.TiKVProportion,
	}

	datastoreChooser := generator.NewDiscrete()
	for datastore, proportion := range proportions {
		if proportion > 0 {
			datastoreChooser.Add(proportion, datastore)
		}
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

	if wp.Task4Proportion > 0 {
		taskChooser.Add(wp.Task4Proportion, 4)
	}

	if wp.Task5Proportion > 0 {
		taskChooser.Add(wp.Task5Proportion, 5)
	}
	if wp.Task6Proportion > 0 {
		taskChooser.Add(wp.Task6Proportion, 6)
	}
	if wp.Task7Proportion > 0 {
		taskChooser.Add(wp.Task7Proportion, 7)
	}
	if wp.Task8Proportion > 0 {
		taskChooser.Add(wp.Task8Proportion, 8)
	}
	if wp.Task9Proportion > 0 {
		taskChooser.Add(wp.Task9Proportion, 9)
	}
	if wp.Task10Proportion > 0 {
		taskChooser.Add(wp.Task10Proportion, 10)
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

func (r *Randomizer) Intn(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.r.Intn(n)
}
