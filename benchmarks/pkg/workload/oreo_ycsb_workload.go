package workload

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"benchmark/pkg/benconfig" // Added for random selection within combination
	"benchmark/ycsb"          // Added for error handling
)

type OreoYCSBWorkload struct {
	Randomizer
	wp *WorkloadParameter

	mu             sync.Mutex
	recordMap      map[string]int
	dbCombinations [][]string // Stores pre-calculated combinations of database names
	involvedDBNum  int
	activeDBs      []string // Stores active database names
}

// DisplayCheckResult implements Workload.
func (wl *OreoYCSBWorkload) DisplayCheckResult() {}

// NeedPostCheck implements Workload.
func (wl *OreoYCSBWorkload) NeedPostCheck() bool {
	return true
}

// NeedRawDB implements Workload.
func (wl *OreoYCSBWorkload) NeedRawDB() bool {
	return false
}

// PostCheck implements Workload.
func (wl *OreoYCSBWorkload) PostCheck(ctx context.Context, db ycsb.DB, resChan chan int) {}

// ResetKeySequence implements Workload.
// Subtle: this method shadows the method (Randomizer).ResetKeySequence of OreoYCSBWorkload.Randomizer.
func (wl *OreoYCSBWorkload) ResetKeySequence() {}

var _ Workload = (*OreoYCSBWorkload)(nil)

func NewOreoYCSBWorkload(wp *WorkloadParameter) *OreoYCSBWorkload {
	activeDBs := getDatabases(*wp)
	if len(activeDBs) == 0 {
		log.Fatal("Oreo YCSB Workload: No databases found with non-zero proportions.")
		return nil // Or handle error differently
	}
	if wp.InvolvedDBNum == 0 {
		log.Printf("WARN: Oreo YCSB Workload: InvolvedDBNum (%d), default to 2", wp.InvolvedDBNum)
		wp.InvolvedDBNum = 2
		// return nil
	}
	if len(activeDBs) < wp.InvolvedDBNum {
		log.Fatalf(
			"Oreo YCSB Workload: Not enough active databases (%d) to satisfy InvolvedDBNum (%d). Active DBs: %v",
			len(activeDBs),
			wp.InvolvedDBNum,
			activeDBs,
		)
		return nil
	}
	if wp.TxnOperationGroup > 0 && wp.TxnOperationGroup < wp.InvolvedDBNum {
		log.Fatalf(
			"TxnOperationGroup (%d) is less than InvolvedDBNum (%d). Not all selected databases might be used in every transaction.",
			wp.TxnOperationGroup,
			wp.InvolvedDBNum,
		)
	}

	combinations := generateCombinations(activeDBs, wp.InvolvedDBNum)
	if len(combinations) == 0 {
		// This should theoretically not happen if len(activeDBs) >= InvolvedDBNum > 0
		log.Fatalf(
			"Oreo YCSB Workload: Failed to generate database combinations for InvolvedDBNum=%d from %v",
			wp.InvolvedDBNum,
			activeDBs,
		)
		return nil
	}

	// Debug log
	// log.Printf("Generated %d combinations for InvolvedDBNum=%d: %v", len(combinations), InvolvedDBNum, combinations)

	return &OreoYCSBWorkload{
		Randomizer:     *NewRandomizer(wp),
		wp:             wp,
		recordMap:      make(map[string]int),
		dbCombinations: combinations,
		involvedDBNum:  wp.InvolvedDBNum,
		activeDBs:      activeDBs, // Store active DBs for Load phase if needed
	}
}

// --- Load function remains largely the same, uses activeDBs ---

func (wl *OreoYCSBWorkload) Load(ctx context.Context, opCount int,
	db ycsb.DB,
) {
	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions")
		return
		// fmt.Println("The DB does not support transactions for Load")
		// // Decide if loading should proceed without transactions or stop
		// // If single inserts are okay for load:
		// // wl.doLoadSimple(ctx, db, wl.activeDBs, opCount)
		// // return
		// log.Println("Warning: DB does not support TransactionDB interface, attempting load with standard DB interface (no batching)")
		// wl.doLoadSimple(ctx, db, wl.activeDBs, opCount) // Use a non-transactional load
		// return
	}

	// Use activeDBs obtained during initialization
	err := wl.doLoad(ctx, txnDB, wl.activeDBs, opCount)
	if err != nil {
		fmt.Printf("Error in Oreo YCSB Load: %v\n", err)
	}
}

// Transactional Load
func (wl *OreoYCSBWorkload) doLoad(
	ctx context.Context,
	db ycsb.TransactionDB,
	dbList []string,
	opCount int,
) error {
	if opCount == 0 {
		return nil
	}
	if benconfig.MaxLoadBatchSize <= 0 {
		return errors.New("MaxLoadBatchSize must be positive for transactional load")
	}
	// Allow opCount not divisible by batch size, handle remainder
	// if opCount%benconfig.MaxLoadBatchSize != 0 {
	//  log.Printf("Warning: opCount (%d) is not a multiple of MaxLoadBatchSize (%d). The last batch will be smaller.", opCount, benconfig.MaxLoadBatchSize)
	// }

	numBatches := opCount / benconfig.MaxLoadBatchSize
	remainder := opCount % benconfig.MaxLoadBatchSize
	var aErr error

	loadBatch := func(batchSize int) error {
		if batchSize == 0 {
			return nil
		}
		var batchErr error
		db.Start()
		for j := 0; j < batchSize; j++ {
			keyName := wl.NextKeyNameFromSequence()
			value := wl.BuildRandomValue()

			for _, dsName := range dbList {
				err := db.Insert(ctx, dsName, keyName, value)
				if err != nil {
					fmt.Printf(
						"Error inserting key %s into %s during load: %v\n",
						keyName,
						dsName,
						err,
					)
				}
			}
		}
		commitErr := db.Commit()
		if commitErr != nil {
			batchErr = fmt.Errorf("error committing load batch: %w", commitErr)
			fmt.Printf("%v\n", batchErr)
		}
		return batchErr
	}

	for i := 0; i < numBatches; i++ {
		err := loadBatch(benconfig.MaxLoadBatchSize)
		if err != nil {
			aErr = err // Keep track of the last error
		}
	}

	// Handle the remainder
	if remainder > 0 {
		err := loadBatch(remainder)
		if err != nil {
			aErr = err // Keep track of the last error
		}
	}

	return aErr
}

// --- getDatabases remains the same ---
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
				// Ensure tag exists and proportion is positive
				if fieldValue > 0 {
					if dbName, ok := field.Tag.Lookup("oreo"); ok && dbName != "" {
						dbList = append(dbList, dbName)
					}
				}
			}
		}
	}
	// Sort the list to ensure consistent combination generation order
	// Although combinations themselves are order-independent, sorting the input helps.
	// sort.Strings(dbList) // If consistency is needed across runs with same config but different map iteration order
	return dbList
}

// --- Run and doTxn are modified ---

func (wl *OreoYCSBWorkload) Run(ctx context.Context, opCount int, db ycsb.DB) {
	if len(wl.dbCombinations) == 0 {
		fmt.Println(
			"Error: No database combinations available for Run phase. Check configuration and InvolvedDBNum.",
		)
		return
	}

	txnDB, ok := db.(ycsb.TransactionDB)
	if !ok {
		fmt.Println("The DB does not support transactions for Run phase, cannot execute workload.")
		return
	}

	for i := 0; i < opCount; i++ {
		if benconfig.GlobalIsFaultTolerance {
			interval := rand.Intn(
				int(benconfig.GlobalFaultToleranceRequestInterval.Milliseconds())+1,
			) + 1
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
		wl.doTxn(ctx, txnDB)
		restTime := rand.Intn(5) + 5
		time.Sleep(time.Duration(restTime) * time.Millisecond)
	}
}

func (wl *OreoYCSBWorkload) doTxn(ctx context.Context, db ycsb.TransactionDB) {
	threadID := ctx.Value("threadID").(int)
	// 1. Select a combination of databases for this transaction uniformly.
	// Use the workload's random source for reproducibility if seeded.
	// selectedCombination := wl.dbCombinations[wl.Intn(len(wl.dbCombinations))]

	// selectedCombination := []string{"Redis", "Cassandra"}
	selectedCombination := wl.dbCombinations[threadID%len(wl.dbCombinations)] // Use threadID for reproducibility

	// log.Printf("Thread %d selected combination: %v", threadID, selectedCombination)

	err := db.Start()
	if err != nil {
		log.Printf("Error starting transaction: %v\n", err)
		return
	}
	for i := 0; i < wl.wp.TxnOperationGroup; i++ {
		// 2. Select an operation type based on its proportion.
		operation := wl.NextOperation()

		// 3. Select a database *from the chosen combination* for this specific operation.
		// We select uniformly from the combination to ensure fair distribution among them within the transaction.
		dsName := selectedCombination[wl.Intn(wl.involvedDBNum)] // Or len(selectedCombination) which is InvolvedDBNum

		// 4. Execute the operation
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
			// Scan operation might need special handling if it applies to one DB
			// Or if it should scan across the combination? Current logic skips it.
			// If scan should operate on one DB from the combo:
			// _ = wl.doScan(ctx, db, dsName) // Assuming doScan exists
			continue // Keep skipping as per original code
		default:
			// Use log.Fatalf for unrecoverable errors
			log.Fatalf("Unknown operation encountered in doTxn: %v", operation)
		}
	}
	err = db.Commit()
	if err != nil {
		// Decide how to handle commit errors, e.g., log, metric, etc.
		// fmt.Printf("Error committing transaction: %v\n", err)
	}
}

// --- Helper functions (doRead, doUpdate, etc.) remain the same ---
// --- NeedPostCheck, NeedRawDB, PostCheck, DisplayCheckResult remain the same ---
// --- NextKeyName remains the same ---
// --- datastoreTypeToName is now unused by doTxn, might be removable if not used elsewhere ---

func (wl *OreoYCSBWorkload) doRead(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName() // This still uses the proportion-based key selection

	_, err := db.Read(ctx, dsName, keyName)
	if err != nil {
		// Log or handle read errors appropriately
		// fmt.Printf("Read error on %s for key %s: %v\n", dsName, keyName, err)
		return err
	}
	return nil
}

func (wl *OreoYCSBWorkload) doUpdate(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue()

	err := db.Update(ctx, dsName, keyName, value)
	if err != nil {
		// Log or handle update errors appropriately
		// fmt.Printf("Update error on %s for key %s: %v\n", dsName, keyName, err)
		return err
	}
	return nil
}

func (wl *OreoYCSBWorkload) doInsert(ctx context.Context, db ycsb.DB, dsName string) error {
	keyName := wl.NextKeyName() // Consider if insert keys should be different (e.g., sequential)
	value := wl.BuildRandomValue()

	err := db.Insert(ctx, dsName, keyName, value)
	if err != nil {
		// Log or handle insert errors appropriately
		// fmt.Printf("Insert error on %s for key %s: %v\n", dsName, keyName, err)
		return err
	}
	return nil
}

func (wl *OreoYCSBWorkload) doReadModifyWrite(
	ctx context.Context,
	db ycsb.DB,
	dsName string,
) error {
	keyName := wl.NextKeyName()
	value := wl.BuildRandomValue() // Build value *before* read? Ok if value is independent

	_, err := db.Read(ctx, dsName, keyName)
	if err != nil {
		// Log or handle read part of RMW errors
		// fmt.Printf("RMW Read error on %s for key %s: %v\n", dsName, keyName, err)
		return err // Fail the RMW if read fails
	}

	// If read succeeds, attempt update
	err = db.Update(ctx, dsName, keyName, value)
	if err != nil {
		// Log or handle update part of RMW errors
		// fmt.Printf("RMW Update error on %s for key %s: %v\n", dsName, keyName, err)
		return err
	}
	return nil
}

func (wl *OreoYCSBWorkload) NextKeyName() string {
	keyName := wl.Randomizer.NextKeyName() // Uses Zipfian or Uniform based on Randomizer's setup

	wl.mu.Lock()
	defer wl.mu.Unlock()
	wl.recordMap[keyName]++ // Still useful for tracking key usage
	return keyName
}

// This function is no longer directly used by doTxn but might be useful elsewhere.
// Keep it or remove it based on overall usage.
func (wl *OreoYCSBWorkload) datastoreTypeToName(dsType datastoreType) string {
	switch dsType {
	case redisDatastore1:
		return "Redis"
	case mongoDatastore1:
		return "MongoDB1"
	case mongoDatastore2:
		return "MongoDB2"
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
		log.Printf("Warning: Unknown datastoreType encountered: %v", dsType)
		return ""
	}
}

// --- Helper function to generate combinations ---

// generateCombinations generates all unique combinations of size k from a slice of strings.
func generateCombinations(items []string, k int) [][]string {
	if k < 0 || k > len(items) {
		return [][]string{} // Return empty slice for invalid k
	}
	if k == 0 {
		return [][]string{{}} // One combination: the empty set
	}
	if k == len(items) {
		// Return a copy to avoid modifying the original slice later
		combination := make([]string, k)
		copy(combination, items)
		return [][]string{combination}
	}

	var combinations [][]string

	// Recursive function to build combinations
	var find func(start int, currentCombo []string)
	find = func(start int, currentCombo []string) {
		if len(currentCombo) == k {
			// Found a combination of size k, make a copy and add it
			comboCopy := make([]string, k)
			copy(comboCopy, currentCombo)
			combinations = append(combinations, comboCopy)
			return
		}

		// Optimization: if remaining items aren't enough to reach k, stop
		if len(items)-start < k-len(currentCombo) {
			return
		}

		// Iterate through remaining items
		for i := start; i < len(items); i++ {
			// Include items[i] in the current combination and recurse
			find(i+1, append(currentCombo, items[i]))
		}
	}

	find(0, []string{}) // Start the process

	return combinations
}
