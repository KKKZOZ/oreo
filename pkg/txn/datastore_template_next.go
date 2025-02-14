package txn

// import (
// 	"cmp"
// 	"fmt"
// 	"log"
// 	"slices"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/go-errors/errors"
// 	"github.com/oreo-dtx-lab/oreo/pkg/config"
// 	"github.com/oreo-dtx-lab/oreo/pkg/errs"
// 	"github.com/oreo-dtx-lab/oreo/pkg/logger"
// 	"golang.org/x/sync/errgroup"
// )

// var _ Datastorer = (*DatastoreNext)(nil)

// // DatastoreNext represents a DatastoreNextr implementation using the underlying connector.
// type DatastoreNext struct {

// 	// Name is the name of the DatastoreNext.
// 	Name string

// 	// Txn is the current transaction.
// 	Txn *Transaction

// 	// conn is the connector interface used by the DatastoreNext.
// 	conn Connector

// 	// readCache is the cache for read operations in DatastoreNext.
// 	// readCache util.ConcurrentMap[string, DataItem]
// 	readCache map[string]DataItemNext

// 	// writeCache is the cache for write operations in DatastoreNext.
// 	writeCache map[string]DataItemNext

// 	// writtenSet is the set of keys that have successfully been written to the DB.
// 	// writtenSet util.ConcurrentMap[string, bool]

// 	// invisibleSet is the set of keys that are not visible to the current transaction.
// 	invisibleSet map[string]bool

// 	// validationSet util.ConcurrentMap[string, PredicateInfo]
// 	validationSet map[string]PredicateInfo

// 	itemFactory DataItemFactoryNext
// 	mu          sync.Mutex
// }

// // NewDatastoreNext creates a new instance of DatastoreNext with the given name and connection.
// // It initializes the read and write caches, as well as the serializer.
// func NewDatastoreNext(name string, conn Connector, factory DataItemFactoryNext) *DatastoreNext {
// 	return &DatastoreNext{
// 		Name:       name,
// 		conn:       conn,
// 		readCache:  make(map[string]DataItemNext),
// 		writeCache: make(map[string]DataItemNext),
// 		// writtenSet:    util.NewConcurrentMap[bool](),
// 		invisibleSet:  make(map[string]bool),
// 		validationSet: make(map[string]PredicateInfo),
// 		itemFactory:   factory,
// 	}
// }

// // Start starts the DatastoreNext by establishing a connection to the underlying server.
// // It returns an error if the connection fails.
// func (r *DatastoreNext) Start() error {
// 	return r.conn.Connect()
// }

// // Read reads a record from the DatastoreNext.
// func (r *DatastoreNext) Read(key string, value any) error {

// 	// if the record is in the writeCache
// 	if item, ok := r.writeCache[key]; ok {
// 		// if the record is marked as deleted
// 		if item.IsDeleted() {
// 			return errors.New(KeyNotFound)
// 		}
// 		return item.ParseValue(value)
// 	}
// 	// if the record is in the readCache
// 	if item, ok := r.readCache[key]; ok {
// 		return item.ParseValue(value)
// 	}
// 	if r.Txn.isRemote {
// 		return r.readFromRemote(key, value)
// 	} else {
// 		return r.readFromConn(key, value)
// 	}

// }

// func (r *DatastoreNext) readFromRemote(key string, value any) error {
// 	item123, readStrategy, groupKeyList, err := r.Txn.RemoteRead(r.Name, key)
// 	// TODO: item type
// 	item := item123.(DataItemNext)
// 	if err != nil {
// 		return errors.Join(errors.New("Remote read failed"), err)
// 	}
// 	// fmt.Printf("item: %v\n readStrategy: %v\n error: %v", item, readStrategy, err)
// 	switch readStrategy {
// 	case AssumeCommit:
// 		r.validationSet[groupKeyList] = PredicateInfo{
// 			ItemKey: item.Key(),
// 			State:   config.COMMITTED,
// 		}
// 	case AssumeAbort:
// 		r.validationSet[groupKeyList] = PredicateInfo{
// 			ItemKey: item.Key(),
// 			State:   config.ABORTED,
// 		}
// 	}

// 	if item.IsDeleted() {
// 		return errors.New(KeyNotFound)
// 	}
// 	r.readCache[item.Key()] = item
// 	return item.ParseValue(value)
// }

// func (r *DatastoreNext) readFromConn(key string, value any) error {
// 	item123, err := r.conn.GetItem(key)
// 	item := item123.(DataItemNext)
// 	if err != nil {
// 		errMsg := err.Error() + " at GetItem in " + r.Name
// 		return errors.New(errMsg)
// 	}

// 	item, err = r.dirtyReadChecker(item)
// 	if err != nil {
// 		return err
// 	}

// 	upItem, err := r.basicVisibilityProcessor(item)
// 	if err != nil {
// 		return err
// 	}

// 	if config.Debug.NativeMode {
// 		return item.ParseValue(value)
// 	}

// 	curItem, isFound := upItem.GetValidItem(r.Txn.TxnStartTime)
// 	return r.endCheck(curItem, isFound, value)
// 	// return r.treatAsCommitted(resItem, logicFunc)
// }

// func (r *DatastoreNext) endCheck(curItem DataItemNext, isFound bool, valuePtr any) error {
// 	if !isFound {
// 		return errs.NewKeyNotFoundError(curItem.Key(), errs.NotFoundInAVC)
// 	}

// 	if curItem.IsDeleted() {
// 		r.readCache[curItem.Key()] = curItem
// 		return errs.NewKeyNotFoundError(curItem.Key(), errs.Deleted)
// 	}

// 	if valuePtr != nil {
// 		err := curItem.ParseValue(valuePtr)
// 		if err != nil {
// 			return fmt.Errorf("endCheck failed: %w", err)
// 		}
// 	}
// 	r.readCache[curItem.Key()] = curItem
// 	return nil
// }

// // dirtyReadChecker will drop an item if it violates repeatable read rules.
// func (r *DatastoreNext) dirtyReadChecker(item DataItemNext) (DataItemNext, error) {
// 	if _, ok := r.invisibleSet[item.Key()]; ok {
// 		return nil, errors.New(KeyNotFound)
// 	} else {
// 		return item, nil
// 	}
// }

// // basicVisibilityProcessor performs basic visibility processing on a DataItem.
// // It tries to bring the item to the COMMITTED state by performing rollback or rollforward operations.
// func (r *DatastoreNext) basicVisibilityProcessor(item DataItemNext) (DataItemNext, error) {
// 	// function to perform the rollback operation
// 	rollbackFunc := func() (DataItemNext, error) {
// 		item, err := r.rollback(item)
// 		if err != nil {
// 			return nil, err
// 		}

// 		if item.Empty() {
// 			return nil, errors.New(KeyNotFound)
// 		}
// 		return item, err
// 	}

// 	// function to perform the rollforward operation
// 	rollforwardFunc := func() (DataItemNext, error) {
// 		item, err := r.rollForward(item)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return item, nil
// 	}

// 	if config.Debug.NativeMode {
// 		return item, nil
// 	}

// 	if item.TxnState() == config.COMMITTED {
// 		return item, nil
// 	}

// 	if item.TxnState() == config.PREPARED {
// 		groupKeyList, err := r.Txn.GetGroupKeyFromItem(item)
// 		if err == nil {
// 			if CommittedForAll(groupKeyList) {
// 				// if all the group keys are in COMMITTED state
// 				// Items in PREPARED state do not contain a valid TValid
// 				// update its TCommit first
// 				tCommit := int64(0)
// 				for _, gk := range groupKeyList {
// 					tCommit = max(tCommit, gk.TCommit)
// 				}
// 				item.SetTValid(tCommit)
// 				return rollforwardFunc()
// 			} else {
// 				// or at least one of the group keys is in ABORTED state
// 				return rollbackFunc()
// 			}
// 		}
// 		// if at least one of the group key is not found
// 		// and if t_lease has expired
// 		// that is, item's TLease < current time
// 		// we should roll back the record
// 		if item.TLease().Before(time.Now()) {
// 			successNum := r.Txn.CreateGroupKeyFromItem(item, config.ABORTED)
// 			if successNum == 0 {
// 				return nil, fmt.Errorf("failed to rollback the record because none of the group keys are created")
// 				// if err.Error() == "key exists" {
// 				// 	if curState == config.COMMITTED {
// 				// 		return nil, errors.New("rollback failed because the corresponding transaction has committed")
// 				// 	}
// 				// 	// if curState == config.ABORTED
// 				// 	// it means the transaction has been rolled back
// 				// 	// so we can safely rollback the record
// 				// } else {
// 				// 	return nil, err
// 				// }
// 			}
// 			return rollbackFunc()
// 		}

// 		// if TSR does not exist
// 		// and if the corresponding transaction is a concurrent transaction
// 		// that is,
// 		// txn's TStart < item's TValid < current time <item's TLease
// 		// we should try check the previous record
// 		if r.Txn.TxnStartTime < item.TValid() {
// 			// Origin Cherry Garcia would do
// 			if config.Debug.CherryGarciaMode {
// 				return nil, errors.New(ReadFailed)
// 			}

// 			// a little trick here:
// 			// if the record is not found in the treatAsCommitted,
// 			// we should add it to the invisibleSet.
// 			// if the record can be found in the treatAsCommitted,
// 			// it will be stored in the readCache,
// 			// so we don't bother dirtyReadChecker anymore.
// 			r.invisibleSet[item.Key()] = true
// 			// if prev is empty

// 			return item.GetPrevItem()

// 			// if item.Prev() == "" {
// 			// 	return nil, errors.New(KeyNotFound)
// 			// }
// 			// return r.getPrevItem(item)
// 		}

// 		if config.Config.ReadStrategy == config.Pessimistic {
// 			return nil, errors.New(ReadFailed)
// 		} else {
// 			switch config.Config.ReadStrategy {
// 			case config.AssumeCommit:
// 				r.validationSet[item.GroupKeyList()] = PredicateInfo{
// 					ItemKey: item.Key(),
// 					State:   config.COMMITTED,
// 				}
// 				return item, nil
// 			case config.AssumeAbort:
// 				r.validationSet[item.GroupKeyList()] = PredicateInfo{
// 					State:     config.ABORTED,
// 					ItemKey:   item.Key(),
// 					LeaseTime: item.TLease(),
// 				}

// 				return item.GetPrevItem()
// 				// if item.Prev() == "" {
// 				// 	return nil, errors.New("key not found in AssumeAbort")
// 				// }
// 				// return r.getPrevItem(item)
// 			}
// 		}
// 	}
// 	return nil, errors.New(KeyNotFound)
// }

// // Write writes a record to the cache.
// // It will serialize the value using the DatastoreNext's serializer.
// func (r *DatastoreNext) Write(key string, value any) error {

// 	// if the record is in the writeCache
// 	if item, ok := r.writeCache[key]; ok {
// 		item.SetValue(value)
// 		item.SetIsDeleted(false)
// 		r.writeCache[key] = item
// 		return nil
// 	}

// 	// else write a new record to the cache
// 	cacheItem := r.itemFactory.NewDataItem(ItemOptionsNext{
// 		Key:   key,
// 		Value: value,
// 		// GroupKeyList: strings.Join(r.Txn.GroupKeyUrls, ","),
// 	})
// 	return r.writeToCache(cacheItem)
// }

// // writeToCache writes the given DataItem to the cache.
// // It will find the corresponding version of the item.
// //   - If the item already exists in the read cache, it follows the read-modified-commit pattern
// //   - If this is a direct write, it will set the version to ""
// //
// // The item is then added to the write cache.
// func (r *DatastoreNext) writeToCache(cacheItem DataItemNext) error {

// 	// check if it follows read-modified-commit pattern
// 	if oldItem, ok := r.readCache[cacheItem.Key()]; ok {
// 		cacheItem.SetVersion(oldItem.Version())
// 	} else {
// 		// else we set it to empty, indicating this is a direct write
// 		cacheItem.SetVersion("")
// 	}

// 	r.writeCache[cacheItem.Key()] = cacheItem
// 	return nil
// }

// // Delete deletes a record from the DatastoreNext.
// // It will return an error if the record is not found.
// func (r *DatastoreNext) Delete(key string) error {
// 	// if the record is in the writeCache
// 	if item, ok := r.writeCache[key]; ok {
// 		if item.IsDeleted() {
// 			return errors.New(KeyNotFound)
// 		}
// 		item.SetIsDeleted(true)
// 		r.writeCache[key] = item
// 		return nil
// 	}

// 	// else write a new record to the cache
// 	cacheItem := r.itemFactory.NewDataItem(ItemOptionsNext{
// 		Key:          key,
// 		GroupKeyList: strings.Join(r.Txn.GroupKeyUrls, ","),
// 		IsDeleted:    true,
// 	})
// 	return r.writeToCache(cacheItem)
// }

// // doConditionalUpdate performs the real conditonal update according to item's state
// func (r *DatastoreNext) doConditionalUpdate(cacheItem DataItemNext, dbItem DataItemNext) error {

// 	err := cacheItem.UpdateMetadata(dbItem)
// 	// newItem, err := r.updateMetadata(cacheItem, dbItem)
// 	if err != nil {
// 		return err
// 	}
// 	doCreate := false
// 	if dbItem == nil || dbItem.Empty() {
// 		doCreate = true
// 	}
// 	newVer, err := r.conn.ConditionalUpdate(cacheItem.Key(), cacheItem, doCreate)
// 	if err != nil {
// 		return err
// 	}
// 	cacheItem.SetVersion(newVer)

// 	r.mu.Lock()
// 	r.writeCache[cacheItem.Key()] = cacheItem
// 	r.mu.Unlock()
// 	return nil
// }

// // conditionalUpdate performs a conditional update operation on the DatastoreNext.
// // It retrieves the corresponding item from the Redis connection and applies basic visibility processing.
// // If the item is not found, it performs a conditional update with an empty DataItem.
// // If there is an error during the retrieval or processing, it handles the error accordingly.
// // Finally, it performs the conditional update operation with the cacheItem and the processed dbItem.
// func (r *DatastoreNext) conditionalUpdate(cacheItem DataItemNext) error {
// 	debugStart := time.Now()
// 	defer func() {
// 		logger.Log.Debugw("DatastoreNext.conditionalUpdate() finishes", "LatencyInFunc", time.Since(debugStart))
// 	}()

// 	// if the cacheItem follows read-modified-write pattern,
// 	// it already has a valid version, we can skip the read step.
// 	if cacheItem.Version() != "" {
// 		dbItem, _ := r.readCache[cacheItem.Key()]
// 		return r.doConditionalUpdate(cacheItem, dbItem)
// 	}

// 	// else we read from connection
// 	err := r.readFromConn(cacheItem.Key(), nil)
// 	if err != nil {
// 		if !strings.Contains(err.Error(), "key not found") {
// 			return err
// 		}
// 	}
// 	dbItem, _ := r.readCache[cacheItem.Key()]
// 	// if the record is dropped by the repeatable read rule
// 	if res, ok := r.invisibleSet[cacheItem.Key()]; ok && res {
// 		dbItem = nil
// 	}
// 	return r.doConditionalUpdate(cacheItem, dbItem)
// }

// func (r *DatastoreNext) rollbackFromConn(key string) error {

// 	item, err := r.conn.GetItem(key)
// 	if err != nil {
// 		return err
// 	}

// 	if item.TxnState() != config.PREPARED {
// 		return errors.New("rollback failed due to wrong state")
// 	}
// 	// if item.TxnId() != txnId {
// 	// 	return errors.New("rollback failed due to wrong txnId")
// 	// }

// 	if item.TLease().Before(time.Now()) {
// 		successNum := r.Txn.CreateGroupKeyFromItem(item, config.ABORTED)
// 		if successNum == 0 {
// 			return fmt.Errorf("failed to rollback the record because none of the group keys are created")
// 		}
// 		_, err = r.rollback(item)
// 		return err
// 	}
// 	return nil
// }

// func (r *DatastoreNext) validate() error {

// 	if config.Config.ReadStrategy == config.Pessimistic {
// 		return nil
// 	}

// 	var eg errgroup.Group
// 	for gkk, predd := range r.validationSet {
// 		gk := gkk
// 		pred := predd
// 		if pred.ItemKey == "" {
// 			log.Fatalf("item's key is empty")
// 			continue
// 		}
// 		eg.Go(func() error {
// 			urlList := strings.Split(gk, ",")
// 			groupKey, err := r.Txn.GetGroupKeyFromUrls(urlList)
// 			// curState, err := r.Txn.tsrMaintainer.ReadTSR(txnId)
// 			if err != nil {
// 				if config.Config.ReadStrategy == config.AssumeAbort {
// 					if pred.LeaseTime.Before(time.Now()) {
// 						key := pred.ItemKey
// 						err := r.rollbackFromConn(key)
// 						if err != nil {
// 							return errors.Join(errors.New("validation failed in AA mode"), err)
// 						} else {
// 							return nil
// 						}
// 					}
// 				}
// 				return errors.New("validation failed due to unknown status")
// 			}
// 			if AtLeastOneAborted(groupKey) {
// 				return nil
// 			} else {
// 				return errors.New("validation failed due to false assumption")
// 			}
// 			// if groupKey.TxnState != pred.State {
// 			// 	return errors.New("validation failed due to false assumption")
// 			// } else {
// 			// 	return nil
// 			// }
// 		})
// 	}

// 	return eg.Wait()
// }

// // Prepare prepares the DatastoreNext for commit.
// func (r *DatastoreNext) Prepare() (int64, error) {

// 	items := make([]DataItemNext, 0, len(r.writeCache))
// 	for _, v := range r.writeCache {
// 		v.SetGroupKeyList(strings.Join(r.Txn.GroupKeyUrls, ","))
// 		items = append(items, v)
// 	}

// 	if len(items) == 0 {
// 		return 0, nil
// 	}

// 	if config.Debug.NativeMode {
// 		return r.prepareInNative(items)
// 	}

// 	// Only return TCommit for remote mode
// 	if r.Txn.isRemote {
// 		return r.prepareInRemote(items)
// 	}

// 	err := r.validate()
// 	if err != nil {
// 		return 0, err
// 	}

// 	if config.Config.ConcurrentOptimizationLevel < config.PARALLELIZE_ON_UPDATE {
// 		// sort records by key
// 		slices.SortFunc(
// 			items, func(i, j DataItemNext) int {
// 				return cmp.Compare(i.Key(), j.Key())
// 			},
// 		)
// 		for _, item := range items {
// 			if err := r.conditionalUpdate(item); err != nil {
// 				return 0, err
// 			}
// 		}
// 		return 0, nil
// 	}

// 	var eg errgroup.Group
// 	// eg.SetLimit(config.Config.MaxOutstandingRequest)
// 	for _, item := range items {
// 		it := item
// 		eg.Go(func() error {
// 			return r.conditionalUpdate(it)
// 		})
// 	}
// 	return 0, eg.Wait()
// }

// func (r *DatastoreNext) prepareInNative(items []DataItemNext) (int64, error) {
// 	var err error
// 	for _, itemm := range items {
// 		item := itemm
// 		var wg sync.WaitGroup
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			cacheItem, _ := r.readCache[item.Key()]
// 			aErr := item.UpdateMetadata(cacheItem)
// 			// item, aErr := r.updateMetadata(item, cacheItem)
// 			if aErr != nil {
// 				err = aErr
// 			}
// 			_, aErr = r.conn.PutItem(item.Key(), item)
// 			if aErr != nil {
// 				err = aErr
// 			}
// 		}()
// 		wg.Wait()
// 	}

// 	return 0, err
// }

// func (r *DatastoreNext) prepareInRemote(items []DataItemNext) (int64, error) {
// 	// for those whose version is clear, update their metadata
// 	for _, item := range items {
// 		if item.Version() != "" {
// 			dbItem, _ := r.readCache[item.Key()]
// 			err := item.UpdateMetadata(dbItem)
// 			// newItem, err := r.updateMetadata(item, dbItem)
// 			if err != nil {
// 				return 0, errors.Errorf("UpdateMetadata failed: %v", err)
// 			}
// 			r.writeCache[item.Key()] = item
// 		}
// 	}

// 	if len(r.validationSet) != 0 {
// 		config.Debug.AssumptionCount++
// 	}

// 	verMap, tCommit, err := r.Txn.RemotePrepare(r.Name, items, r.validationSet)
// 	logger.Log.Debugw("Remote prepare Result",
// 		"TxnId", r.Txn.TxnId, "verMap", verMap, "err", err, "Latency", time.Since(r.Txn.debugStart), "Topic", "CheckPoint")
// 	if err != nil {
// 		return 0, errors.Join(errors.New("Remote prepare failed"), err)
// 	}
// 	for k, v := range verMap {
// 		r.writeCache[k].SetVersion(v)
// 	}
// 	return tCommit, nil
// }

// // Commit updates the state of records in the data store to COMMITTED.
// //
// // It iterates over the write cache and updates each record's state to COMMITTED.
// //
// // After updating the records, it clears the write cache.
// // Returns an error if there is any issue updating the records.
// func (r *DatastoreNext) Commit() error {
// 	logger.Log.Debugw("DatastoreNext.Commit() starts", "r.Txn.isRemote", r.Txn.isRemote)

// 	if r.Txn.isRemote {
// 		return r.commitInRemote()
// 	}

// 	// update record's state to the COMMITTED state in the data store
// 	var eg errgroup.Group
// 	// eg.SetLimit(config.Config.MaxOutstandingRequest)
// 	for _, item := range r.writeCache {
// 		it := item
// 		eg.Go(func() error {
// 			it.SetTxnState(config.COMMITTED)

// 			_, err := r.conn.ConditionalUpdate(it.Key(), it, false)
// 			if errors.Is(err, VersionMismatch) {
// 				// this indicates that the record has been rolled forward
// 				// by another transaction.
// 				return nil
// 			}
// 			return err
// 		})
// 	}
// 	eg.Wait()
// 	logger.Log.Debugw("DatastoreNext.Commit() finishes", "TxnId", r.Txn.TxnId)
// 	r.clear()
// 	return nil
// }

// func (r *DatastoreNext) commitInRemote() error {
// 	infoList := make([]CommitInfo, 0, len(r.writeCache))
// 	for _, item := range r.writeCache {
// 		infoList = append(infoList, CommitInfo{Key: item.Key(), Version: item.Version()})
// 	}

// 	err := r.Txn.RemoteCommit(r.Name, infoList)
// 	if err != nil {
// 		logger.Log.Infow("Remote commit failed", "TxnId", r.Txn.TxnId)
// 		return err
// 	}
// 	return err
// }

// // Abort discards the changes made in the current transaction.
// //
// //   - If hasCommitted is false, it clears the write cache.
// //   - If hasCommitted is true, it rolls back the changes made by the current transaction.
// //
// // It returns an error if there is any issue during the rollback process.
// func (r *DatastoreNext) Abort(hasCommitted bool) error {

// 	if !hasCommitted {
// 		r.clear()
// 		return nil
// 	}

// 	if r.Txn.isRemote {
// 		keyList := make([]string, 0, len(r.writeCache))
// 		for _, item := range r.writeCache {
// 			keyList = append(keyList, item.Key())
// 		}
// 		return r.Txn.RemoteAbort(r.Name, keyList)
// 	}

// 	for _, v := range r.writeCache {
// 		item, err := r.conn.GetItem(v.Key())
// 		if err != nil {
// 			return err
// 		}
// 		// if the record has been modified by this transaction
// 		curGroupKeyList := strings.Join(r.Txn.GroupKeyUrls, ",")
// 		if item.GroupKeyList() == curGroupKeyList {
// 			r.rollback(item)
// 		}
// 	}
// 	r.clear()
// 	return nil
// }

// func (r *DatastoreNext) OnePhaseCommit() error {
// 	if len(r.writeCache) == 0 {
// 		return nil
// 	}
// 	// there is only one record in the writeCache
// 	for _, item := range r.writeCache {
// 		return r.conditionalUpdate(item)
// 	}
// 	return nil
// }

// // rollback overwrites the record with the application data
// // and metadata that found in field Prev.
// // if the `Prev` is empty, it simply deletes the record
// func (r *DatastoreNext) rollback(item DataItemNext) (DataItemNext, error) {

// 	prevItem, err := item.GetPrevItem()
// 	if err != nil && err.Error() == "key not found" {
// 		item.SetIsDeleted(true)
// 		item.SetTxnState(config.COMMITTED)
// 		newVer, err := r.conn.ConditionalUpdate(item.Key(), item, false)
// 		if err != nil {
// 			return nil, errors.Join(errors.New("rollback failed"), err)
// 		}
// 		item.SetVersion(newVer)
// 		return item, err
// 	}
// 	if err != nil {
// 		return nil, errors.Join(errors.New("rollback failed"), err)
// 	}

// 	// try to rollback through ConditionalUpdate
// 	prevItem.SetVersion(item.Version())
// 	newVer, err := r.conn.ConditionalUpdate(item.Key(), prevItem, false)
// 	// err = r.conn.PutItem(item.Key, newItem)
// 	if err != nil {
// 		return nil, errors.Join(errors.New("rollback failed"), err)
// 	}
// 	// update the version
// 	prevItem.SetVersion(newVer)

// 	return prevItem, err
// }

// // rollForward makes the record metadata with COMMITTED state
// func (r *DatastoreNext) rollForward(item DataItemNext) (DataItemNext, error) {

// 	item.SetTxnState(config.COMMITTED)
// 	newVer, err := r.conn.ConditionalUpdate(item.Key(), item, false)
// 	if err != nil {
// 		return nil, errors.Join(errors.New("rollForward failed"), err)
// 	}
// 	item.SetVersion(newVer)
// 	return item, err
// }

// // GetName returns the name of the DatastoreNext.
// func (r *DatastoreNext) GetName() string {
// 	return r.Name
// }

// // SetTxn sets the transaction for the MemoryDatastoreNext.
// // It takes a pointer to a Transaction as input and assigns it to the Txn field of the MemoryDatastoreNext.
// func (r *DatastoreNext) SetTxn(txn *Transaction) {
// 	r.Txn = txn
// }

// // Copy returns a new instance of DatastoreNext with the same name and connection.
// // It is used to create a copy of the DatastoreNext object.
// func (r *DatastoreNext) Copy() Datastorer {
// 	return NewDatastoreNext(r.Name, r.conn, r.itemFactory)
// }

// func (r *DatastoreNext) GetConn() Connector {
// 	return r.conn
// }

// func (r *DatastoreNext) GetWriteCacheSize() int {
// 	return len(r.writeCache)
// }

// func (r *DatastoreNext) clear() {
// 	r.readCache = make(map[string]DataItemNext)
// 	r.writeCache = make(map[string]DataItemNext)
// 	// r.writtenSet = util.NewConcurrentMap[bool]()
// 	r.invisibleSet = make(map[string]bool)
// 	r.validationSet = make(map[string]PredicateInfo)
// }
