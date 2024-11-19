package workload

import (
	"benchmark/ycsb"
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type OreoYCSBWorkload struct {
	Randomizer
	wp *WorkloadParameter

	mu        sync.Mutex
	recordMap map[string]int
}

var _ Workload = (*YCSBWorkload)(nil)

func NewOreoYCSBWorkload(wp *WorkloadParameter) *OreoYCSBWorkload {

	return &OreoYCSBWorkload{
		Randomizer: *NewRandomizer(wp),
		wp:         wp,
		recordMap:  make(map[string]int),
	}
}

func (wl *OreoYCSBWorkload) Load(ctx context.Context, opCount int,
	db ycsb.DB) {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}

	// dbList := make([]string, 0)
	// if wl.wp.KVRocksProportion > 0 {
	// 	dbList = append(dbList, "KVRocks")
	// }
	// if wl.wp.Redis1Proportion > 0 {
	// 	dbList = append(dbList, "Redis")
	// }
	// if wl.wp.Mongo1Proportion > 0 {
	// 	dbList = append(dbList, "MongoDB")
	// }
	// if wl.wp.CouchDBProportion > 0 {
	// 	dbList = append(dbList, "CouchDB")
	// }
	// if wl.wp.CassandraProportion > 0 {
	// 	dbList = append(dbList, "Cassandra")
	// }
	// if wl.wp.DynamoDBProportion > 0 {
	// 	dbList = append(dbList, "DynamoDB")
	// }
	dbList := getDatabases(*wl.wp)
	err := wl.doLoad(ctx, txnDB, dbList, opCount)
	if err != nil {
		fmt.Printf("Error in Oreo YCSB Load: %v\n", err)
	}
}

func getDatabases(wp WorkloadParameter) []string {
	dbList := make([]string, 0)
	value := reflect.ValueOf(wp)
	typ := reflect.TypeOf(wp)

	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)

		// 只处理以 Proportion 结尾的字段
		if strings.HasSuffix(field.Name, "Proportion") {
			// 确保字段是 float64 类型
			if field.Type.Kind() == reflect.Float64 {
				fieldValue := value.Field(i).Float()
				if fieldValue > 0 {
					if dbName, ok := field.Tag.Lookup("oreo"); ok && dbName != "" {
						dbList = append(dbList, dbName)
					}
				}
			}
		}
	}

	return dbList
}

func (wl *OreoYCSBWorkload) doLoad(ctx context.Context, db ycsb.TransactionDB, dbList []string, opCount int) error {
	db.Start()
	for i := 0; i < opCount; i++ {
		keyName := wl.NextKeyNameFromSequence()
		value := wl.BuildRandomValue()

		for _, dsName := range dbList {
			err := db.Insert(ctx, dsName, keyName, value)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	err := db.Commit()
	if err != nil {
		fmt.Printf("Error in Oreo YCSB Load: %v\n", err)
	}
	return err
}

func (wl *OreoYCSBWorkload) Run(ctx context.Context, opCount int, db ycsb.DB) {

	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
	}

	for i := 0; i < opCount; i++ {
		wl.doTxn(ctx, txnDB)
	}
}

func (wl *OreoYCSBWorkload) doTxn(ctx context.Context, db ycsb.TransactionDB) {
	db.Start()
	for i := 0; i < wl.wp.TxnOperationGroup; i++ {
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
		case readModifyWrite:
			_ = wl.doReadModifyWrite(ctx, db, dsName)
		case scan:
			continue
		default:
			panic("Unknown operation")
		}
	}
	db.Commit()
}

func (wl *OreoYCSBWorkload) NeedPostCheck() bool {
	return false
}

func (wl *OreoYCSBWorkload) NeedRawDB() bool {
	return false
}

func (wl *OreoYCSBWorkload) PostCheck(ctx context.Context, db ycsb.DB,
	resChan chan int) {
}

func (wl *OreoYCSBWorkload) DisplayCheckResult() {

}

func (wl *OreoYCSBWorkload) doRead(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName()

	_, err := db.Read(ctx, dsName, keyName)
	if err != nil {
		return err
	}
	return nil
}

func (wl *OreoYCSBWorkload) doUpdate(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue()

	err := db.Update(ctx, dsName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (wl *OreoYCSBWorkload) doInsert(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue()

	err := db.Insert(ctx, dsName, keyName, value)
	if err != nil {
		return err
	}
	return nil
}

func (wl *OreoYCSBWorkload) doReadModifyWrite(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue()

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

func (wl *OreoYCSBWorkload) NextKeyName() string {
	keyName := wl.Randomizer.NextKeyName()

	wl.mu.Lock()
	defer wl.mu.Unlock()
	wl.recordMap[keyName]++
	return keyName
}

func (wl *OreoYCSBWorkload) datastoreTypeToName(dsType datastoreType) string {
	switch dsType {
	case redisDatastore1:
		return "Redis"
	case mongoDatastore1:
		return "MongoDB"
	case mongoDatastore2:
		return "MongoDB"
	case couchDatastore1:
		return "CouchDB"
	case kvrocksDatastore1:
		return "KVRocks"
	case cassandraDatastore1:
		return "Cassandra"
	case dynamodbDatastore1:
		return "DynamoDB"
	case tikvDatastore1:
		return "TiKV"
	default:
		return ""
	}
}
