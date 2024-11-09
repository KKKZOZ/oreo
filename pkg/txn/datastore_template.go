// a pure interface version
package txn

import (
	"cmp"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"golang.org/x/sync/errgroup"
)

var _ Datastorer = (*Datastore)(nil)

const (
	EMPTY         string = ""
	RETRYINTERVAL        = 10 * time.Millisecond
)

type PredicateInfo struct {
	State     config.State
	ItemKey   string
	LeaseTime time.Time
}

// Datastore represents a datastorer implementation using the underlying connector.
type Datastore struct {

	// Name is the name of the datastore.
	Name string

	// Txn is the current transaction.
	Txn *Transaction

	// conn is the connector interface used by the datastore.
	conn Connector

	// readCache is the cache for read operations in Datastore.
	// readCache util.ConcurrentMap[string, DataItem]
	readCache map[string]DataItem

	// writeCache is the cache for write operations in Datastore.
	writeCache map[string]DataItem

	// writtenSet is the set of keys that have successfully been written to the DB.
	// writtenSet util.ConcurrentMap[string, bool]

	// invisibleSet is the set of keys that are not visible to the current transaction.
	invisibleSet map[string]bool

	// validationSet util.ConcurrentMap[string, PredicateInfo]
	validationSet map[string]PredicateInfo

	// se is the serializer used for serializing and deserializing data in Datastore.
	se serializer.Serializer

	// itemFactory is the factory used for creating DataItems.
	itemFactory DataItemFactory

	// mu is the mutex used for locking the Datastore.
	mu sync.Mutex
}

// NewDatastore creates a new instance of Datastore with the given name and connection.
// It initializes the read and write caches, as well as the serializer.
func NewDatastore(name string, conn Connector, factory DataItemFactory) *Datastore {
	return &Datastore{
		Name:       name,
		conn:       conn,
		readCache:  make(map[string]DataItem),
		writeCache: make(map[string]DataItem),
		// writtenSet:    util.NewConcurrentMap[bool](),
		invisibleSet:  make(map[string]bool),
		validationSet: make(map[string]PredicateInfo),
		se:            config.Config.Serializer,
		itemFactory:   factory,
	}
}

// Start starts the Datastore by establishing a connection to the underlying server.
// It returns an error if the connection fails.
func (r *Datastore) Start() error {
	return r.conn.Connect()
}

// Read reads a record from the Datastore.
func (r *Datastore) Read(key string, value any) error {
	// if the record is in the writeCache
	if item, ok := r.writeCache[key]; ok {
		// if the record is marked as deleted
		if item.IsDeleted() {
			return errors.New(KeyNotFound)
		}
		return r.getValue(item, value)
	}
	// if the record is in the readCache
	if item, ok := r.readCache[key]; ok {
		return r.getValue(item, value)
	}

	if r.Txn.isRemote {
		return r.readFromRemote(key, value)
	} else {
		return r.readFromConn(key, value)
	}
}

func (r *Datastore) readFromRemote(key string, value any) error {
	item, readStrategy, err := r.Txn.RemoteRead(r.Name, key)
	if err != nil {
		return err
	}
	// TODO: logic for AssumeCommit and AssumeAbort
	switch readStrategy {
	case AssumeCommit:
		r.validationSet[item.GroupKeyList()] = PredicateInfo{
			ItemKey: item.Key(),
			State:   config.COMMITTED,
		}
	case AssumeAbort:
		r.validationSet[item.GroupKeyList()] = PredicateInfo{
			ItemKey: item.Key(),
			State:   config.ABORTED,
		}
	}

	if item.IsDeleted() {
		return errors.New(KeyNotFound)
	}
	r.readCache[item.Key()] = item
	return r.getValue(item, value)
}

func (r *Datastore) readFromConn(key string, value any) error {
	item, err := r.conn.GetItem(key)
	if err != nil {
		errMsg := err.Error() + " at GetItem in " + r.Name
		return errors.New(errMsg)
	}

	item, err = r.dirtyReadChecker(item)
	if err != nil {
		return err
	}

	resItem, err := r.basicVisibilityProcessor(item)
	if err != nil {
		return err
	}

	if config.Debug.NativeMode {
		return r.getValue(resItem, value)
	}

	logicFunc := func(curItem DataItem, isFound bool) error {
		// if the record has been deleted
		if !isFound || curItem.IsDeleted() {
			if curItem.IsDeleted() {
				// put into cache anyway
				//
				// This is a special case for getting the corresponding version
				// Consider the case where the record is deleted by another transaction
				// and the current transaction tries to read it to get version info
				// If it is not put into the cache, the record is invisible for prepare phase
				// so the code will regard it as a new record
				// and create a new record in the prepare phase (set `doCreate` to true)
				r.readCache[curItem.Key()] = curItem
				errMsg := "key not found because item is already deleted in " + r.Name
				return errors.New(errMsg)
			}
			return errors.New(KeyNotFound)
		}
		if value != nil {
			err := r.getValue(curItem, value)
			if err != nil {
				return err
			}
		}
		r.readCache[curItem.Key()] = curItem
		return nil
	}

	return r.treatAsCommitted(resItem, logicFunc)
}

// dirtyReadChecker will drop an item if it violates repeatable read rules.
func (r *Datastore) dirtyReadChecker(item DataItem) (DataItem, error) {
	if _, ok := r.invisibleSet[item.Key()]; ok {
		return nil, errors.New(KeyNotFound)
	} else {
		return item, nil
	}
}

// basicVisibilityProcessor performs basic visibility processing on a DataItem.
// It tries to bring the item to the COMMITTED state by performing rollback or rollforward operations.
func (r *Datastore) basicVisibilityProcessor(item DataItem) (DataItem, error) {
	// function to perform the rollback operation
	rollbackFunc := func() (DataItem, error) {
		item, err := r.rollback(item)
		if err != nil {
			return nil, err
		}

		if item.Empty() {
			return nil, errors.New(KeyNotFound)
		}
		return item, err
	}

	// function to perform the rollforward operation
	rollforwardFunc := func() (DataItem, error) {
		item, err := r.rollForward(item)
		if err != nil {
			return nil, err
		}
		return item, nil
	}

	if config.Debug.NativeMode {
		return item, nil
	}

	if item.TxnState() == config.COMMITTED {
		return item, nil
	}

	if item.TxnState() == config.PREPARED {
		groupKeyList, err := r.Txn.GetGroupKeyFromItem(item)
		if err == nil {
			if CommittedForAll(groupKeyList) {
				// if all the group keys are in COMMITTED state
				// Items in PREPARED state do not contain a valid TValid
				if item.TValid() == 0 {
					tCommit := int64(0)
					for _, gk := range groupKeyList {
						tCommit = max(tCommit, gk.TCommit)
					}
					item.SetTValid(tCommit)
				}
				return rollforwardFunc()
			} else {
				// or at least one of the group keys is in ABORTED state
				return rollbackFunc()
			}
		}
		// if at least one of the group key is not found
		// and if t_lease has expired
		// that is, item's TLease < current time
		// we should roll back the record
		if item.TLease().Before(time.Now()) {
			successNum := r.Txn.CreateGroupKeyFromItem(item, config.ABORTED)
			if successNum == 0 {
				return nil, fmt.Errorf("failed to rollback the record because none of the group keys are created")
				// if err.Error() == "key exists" {
				// 	if curState == config.COMMITTED {
				// 		return nil, errors.New("rollback failed because the corresponding transaction has committed")
				// 	}
				// 	// if curState == config.ABORTED
				// 	// it means the transaction has been rolled back
				// 	// so we can safely rollback the record
				// } else {
				// 	return nil, err
				// }
			}
			return rollbackFunc()
		}

		// if TSR does not exist
		// and if the corresponding transaction is a concurrent transaction
		// that is, txn's TStart < item's TValid < item's TLease
		// we should try check the previous record
		if r.Txn.TxnStartTime < item.TValid() {
			// Origin Cherry Garcia would do
			if config.Debug.CherryGarciaMode {
				return nil, errors.New(ReadFailed)
			}

			// a little trick here:
			// if the record is not found in the treatAsCommitted,
			// we should add it to the invisibleSet.
			// if the record can be found in the treatAsCommitted,
			// it will be stored in the readCache,
			// so we don't bother dirtyReadChecker anymore.
			r.invisibleSet[item.Key()] = true
			// if prev is empty
			if item.Prev() == "" {
				return nil, errors.New(KeyNotFound)
			}
			return r.getPrevItem(item)
		}

		if config.Config.ReadStrategy == config.Pessimistic {
			return nil, errors.New(ReadFailed)
		} else {
			switch config.Config.ReadStrategy {
			case config.AssumeCommit:
				r.validationSet[item.GroupKeyList()] = PredicateInfo{
					ItemKey: item.Key(),
					State:   config.COMMITTED,
				}
				return item, nil
			case config.AssumeAbort:
				r.validationSet[item.GroupKeyList()] = PredicateInfo{
					State:     config.ABORTED,
					ItemKey:   item.Key(),
					LeaseTime: item.TLease(),
				}
				if item.Prev() == "" {
					return nil, errors.New("key not found in AssumeAbort")
				}
				return r.getPrevItem(item)
			}
		}
	}
	return nil, errors.New(KeyNotFound)
}

// treatAsCommitted treats a DataItem as committed, finds a corresponding version
// according to its timestamp, and performs the given logic function on it.
func (r *Datastore) treatAsCommitted(item DataItem, logicFunc func(DataItem, bool) error) error {
	curItem := item
	for i := 1; i <= config.Config.MaxRecordLength; i++ {

		if curItem.TValid() < r.Txn.TxnStartTime {
			// find the corresponding version,
			// do some business logic.
			return logicFunc(curItem, true)
		}
		if i == config.Config.MaxRecordLength {
			break
		}
		// if prev is empty
		if curItem.Prev() == "" {
			return logicFunc(curItem, false)
		}

		// get the previous record
		preItem, err := r.getPrevItem(curItem)
		if err != nil {
			return err
		}
		curItem = preItem
	}
	return errors.New(KeyNotFound)
}

// Write writes a record to the cache.
// It will serialize the value using the Datastore's serializer.
func (r *Datastore) Write(key string, value any) error {
	bs, err := r.se.Serialize(value)
	if err != nil {
		return err
	}
	str := string(bs)
	// if the record is in the writeCache
	if item, ok := r.writeCache[key]; ok {
		item.SetValue(str)
		item.SetIsDeleted(false)
		r.writeCache[key] = item
		return nil
	}

	// else write a new record to the cache
	cacheItem := r.itemFactory.NewDataItem(ItemOptions{
		Key:          key,
		Value:        str,
		GroupKeyList: strings.Join(r.Txn.GroupKeyUrls, ","),
	})
	return r.writeToCache(cacheItem)
}

// writeToCache writes the given DataItem to the cache.
// It will find the corresponding version of the item.
//   - If the item already exists in the read cache, it follows the read-modified-commit pattern
//   - If this is a direct write, it will set the version to ""
//
// The item is then added to the write cache.
func (r *Datastore) writeToCache(cacheItem DataItem) error {

	// check if it follows read-modified-commit pattern
	if oldItem, ok := r.readCache[cacheItem.Key()]; ok {
		cacheItem.SetVersion(oldItem.Version())
	} else {
		// else we set it to empty, indicating this is a direct write
		cacheItem.SetVersion("")
	}

	r.writeCache[cacheItem.Key()] = cacheItem
	return nil
}

// Delete deletes a record from the Datastore.
// It will return an error if the record is not found.
func (r *Datastore) Delete(key string) error {
	// if the record is in the writeCache
	if item, ok := r.writeCache[key]; ok {
		if item.IsDeleted() {
			return errors.New(KeyNotFound)
		}
		item.SetIsDeleted(true)
		r.writeCache[key] = item
		return nil
	}

	// else write a new record to the cache
	cacheItem := r.itemFactory.NewDataItem(ItemOptions{
		Key:          key,
		GroupKeyList: strings.Join(r.Txn.GroupKeyUrls, ","),
		IsDeleted:    true,
	})
	return r.writeToCache(cacheItem)
}

// doConditionalUpdate performs the real conditonal update according to item's state
func (r *Datastore) doConditionalUpdate(cacheItem DataItem, dbItem DataItem) error {

	newItem, err := r.updateMetadata(cacheItem, dbItem)
	if err != nil {
		return err
	}
	doCreate := false
	if dbItem == nil || dbItem.Empty() {
		doCreate = true
	}
	newVer, err := r.conn.ConditionalUpdate(newItem.Key(), newItem, doCreate)
	if err != nil {
		return err
	}
	newItem.SetVersion(newVer)

	r.mu.Lock()
	r.writeCache[newItem.Key()] = newItem
	r.mu.Unlock()
	return nil
}

// conditionalUpdate performs a conditional update operation on the Datastore.
// It retrieves the corresponding item from the Redis connection and applies basic visibility processing.
// If the item is not found, it performs a conditional update with an empty DataItem.
// If there is an error during the retrieval or processing, it handles the error accordingly.
// Finally, it performs the conditional update operation with the cacheItem and the processed dbItem.
func (r *Datastore) conditionalUpdate(cacheItem DataItem) error {
	debugStart := time.Now()
	defer func() {
		logger.Log.Debugw("Datastore.conditionalUpdate() finishes", "LatencyInFunc", time.Since(debugStart))
	}()

	// if the cacheItem follows read-modified-write pattern,
	// it already has a valid version, we can skip the read step.
	if cacheItem.Version() != "" {
		dbItem, _ := r.readCache[cacheItem.Key()]
		return r.doConditionalUpdate(cacheItem, dbItem)
	}

	// else we read from connection
	err := r.readFromConn(cacheItem.Key(), nil)
	if err != nil {
		if !strings.Contains(err.Error(), "key not found") {
			return err
		}
	}
	dbItem, _ := r.readCache[cacheItem.Key()]
	// if the record is dropped by the repeatable read rule
	if res, ok := r.invisibleSet[cacheItem.Key()]; ok && res {
		dbItem = nil
	}
	return r.doConditionalUpdate(cacheItem, dbItem)
}

// truncate truncates the linked list of DataItems
// if the length exceeds the maximum record length defined in the configuration.
//
// It takes a pointer to a DataItem as input and returns the truncated DataItem and an error, if any.
// If the length of the linked list is greater than the maximum record length, it creates a stack of DataItems and pops the items from the stack until the length is reduced to the maximum record length.
//
// It then updates the Prev and LinkedLen fields of the DataItems in the stack accordingly.
// Finally, it returns the last popped DataItem as the truncated DataItem.
//
// If the length of the linked list is less than or equal to the maximum record length, it returns the input DataItem as is.
func (r *Datastore) truncate(newItem DataItem) (DataItem, error) {
	maxLen := config.Config.MaxRecordLength

	if newItem.LinkedLen() > maxLen {
		stack := util.NewStack[DataItem]()
		stack.Push(newItem)
		curItem := &newItem
		for i := 1; i <= maxLen-1; i++ {
			preItem, err := r.getPrevItem(*curItem)
			if err != nil {
				return nil, errors.New("Unmarshal error: " + err.Error())
			}
			curItem = &preItem
			stack.Push(*curItem)
		}

		tarItem, err := stack.Pop()
		if err != nil {
			return nil, errors.New("Pop error: " + err.Error())
		}
		tarItem.SetPrev("")
		tarItem.SetLinkedLen(1)

		for !stack.IsEmpty() {
			item, err := stack.Pop()
			if err != nil {
				return nil, errors.New("Pop error: " + err.Error())
			}
			bs, err := r.se.Serialize(tarItem)
			if err != nil {
				return nil, errors.New("Serialize error: " + err.Error())
			}
			item.SetPrev(string(bs))
			item.SetLinkedLen(tarItem.LinkedLen() + 1)
			tarItem = item
		}
		return tarItem, nil
	} else {
		return newItem, nil
	}
}

// updateMetadata updates the metadata of a DataItem by comparing it with the oldItem.
//
// If the oldItem is empty, it sets the LinkedLen of the newItem to 1.
// Otherwise, it increments the LinkedLen of the newItem by 1 and sets the Prev and Version fields based on the oldItem.
//
// It then truncates the record using the truncate method and sets the TxnState, TValid, and TLease fields of the newItem.
// Finally, it returns the updated newItem and any error that occurred during the process.
func (r *Datastore) updateMetadata(newItem DataItem, oldItem DataItem) (DataItem, error) {
	if oldItem == nil {
		newItem.SetLinkedLen(1)
	} else {
		newItem.SetLinkedLen(oldItem.LinkedLen() + 1)
		bs, err := r.se.Serialize(oldItem)
		if err != nil {
			return nil, err
		}
		newItem.SetPrev(string(bs))
		newItem.SetVersion(oldItem.Version())
	}

	// truncate the record
	newItem, err := r.truncate(newItem)
	if err != nil {
		return nil, err
	}

	newItem.SetTxnState(config.PREPARED)
	newItem.SetTValid(r.Txn.TxnCommitTime)
	// TODO: time.Now() is temporary
	newItem.SetTLease(time.Now().Add(config.Config.LeaseTime))
	return newItem, nil
}

func (r *Datastore) rollbackFromConn(key string) error {

	item, err := r.conn.GetItem(key)
	if err != nil {
		return err
	}

	if item.TxnState() != config.PREPARED {
		return errors.New("rollback failed due to wrong state")
	}
	// if item.TxnId() != txnId {
	// 	return errors.New("rollback failed due to wrong txnId")
	// }

	if item.TLease().Before(time.Now()) {
		successNum := r.Txn.CreateGroupKeyFromItem(item, config.ABORTED)
		if successNum == 0 {
			return fmt.Errorf("failed to rollback the record because none of the group keys are created")
		}
		_, err = r.rollback(item)
		return err
	}
	return nil
}

func (r *Datastore) validate() error {

	if config.Config.ReadStrategy == config.Pessimistic {
		return nil
	}

	var eg errgroup.Group
	for gkk, predd := range r.validationSet {
		gk := gkk
		pred := predd
		if pred.ItemKey == "" {
			log.Fatalf("item's key is empty")
			continue
		}
		eg.Go(func() error {
			urlList := strings.Split(gk, ",")
			groupKey, err := r.Txn.GetGroupKeyFromUrls(urlList)
			// curState, err := r.Txn.tsrMaintainer.ReadTSR(txnId)
			if err != nil {
				if config.Config.ReadStrategy == config.AssumeAbort {
					if pred.LeaseTime.Before(time.Now()) {
						key := pred.ItemKey
						err := r.rollbackFromConn(key)
						if err != nil {
							return errors.Join(errors.New("validation failed in AA mode"), err)
						} else {
							return nil
						}
					}
				}
				return errors.New("validation failed due to unknown status")
			}
			if AtLeastOneAborted(groupKey) {
				return nil
			} else {
				return errors.New("validation failed due to false assumption")
			}
			// if groupKey.TxnState != pred.State {
			// 	return errors.New("validation failed due to false assumption")
			// } else {
			// 	return nil
			// }
		})
	}

	return eg.Wait()
}

// Prepare prepares the Datastore for commit.
func (r *Datastore) Prepare() (int64, error) {

	items := make([]DataItem, 0, len(r.writeCache))
	for _, v := range r.writeCache {
		items = append(items, v)
	}

	if len(items) == 0 {
		return 0, nil
	}

	if config.Debug.NativeMode {
		return r.prepareInNative(items)
	}

	// Only return TCommit for remote mode
	if r.Txn.isRemote {
		return r.prepareInRemote(items)
	}

	err := r.validate()
	if err != nil {
		return 0, err
	}

	if config.Config.ConcurrentOptimizationLevel < config.PARALLELIZE_ON_UPDATE {
		// sort records by key
		slices.SortFunc(
			items, func(i, j DataItem) int {
				return cmp.Compare(i.Key(), j.Key())
			},
		)
		for _, item := range items {
			if err := r.conditionalUpdate(item); err != nil {
				return 0, err
			}
		}
		return 0, nil
	}

	var eg errgroup.Group
	// eg.SetLimit(config.Config.MaxOutstandingRequest)
	for _, item := range items {
		it := item
		eg.Go(func() error {
			return r.conditionalUpdate(it)
		})
	}
	return 0, eg.Wait()
}

func (r *Datastore) prepareInNative(items []DataItem) (int64, error) {
	var err error
	for _, itemm := range items {
		item := itemm
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			cacheItem, _ := r.readCache[item.Key()]
			item, aErr := r.updateMetadata(item, cacheItem)
			if aErr != nil {
				err = aErr
			}
			r.conn.PutItem(item.Key(), item)
		}()
		wg.Wait()
	}

	return 0, err
}

func (r *Datastore) prepareInRemote(items []DataItem) (int64, error) {
	// for those whose version is clear, update their metadata
	for _, item := range items {
		if item.Version() != "" {
			dbItem, _ := r.readCache[item.Key()]
			newItem, err := r.updateMetadata(item, dbItem)
			if err != nil {
				return 0, errors.Errorf("UpdateMetadata failed: %v", err)
			}
			r.writeCache[item.Key()] = newItem
		}
	}

	verMap, tCommit, err := r.Txn.RemotePrepare(r.Name, items, r.validationSet)
	logger.Log.Debugw("Remote prepare Result",
		"TxnId", r.Txn.TxnId, "verMap", verMap, "err", err, "Latency", time.Since(r.Txn.debugStart), "Topic", "CheckPoint")
	if err != nil {
		return 0, errors.Join(errors.New("Remote prepare failed"), err)
	}
	for k, v := range verMap {
		r.writeCache[k].SetVersion(v)
	}
	return tCommit, nil
}

// Commit updates the state of records in the data store to COMMITTED.
//
// It iterates over the write cache and updates each record's state to COMMITTED.
//
// After updating the records, it clears the write cache.
// Returns an error if there is any issue updating the records.
func (r *Datastore) Commit() error {
	logger.Log.Debugw("Datastore.Commit() starts", "r.Txn.isRemote", r.Txn.isRemote)

	if r.Txn.isRemote {
		return r.commitInRemote()
	}

	// update record's state to the COMMITTED state in the data store
	var eg errgroup.Group
	// eg.SetLimit(config.Config.MaxOutstandingRequest)
	for _, item := range r.writeCache {
		it := item
		eg.Go(func() error {
			it.SetTxnState(config.COMMITTED)

			_, err := r.conn.ConditionalUpdate(it.Key(), it, false)
			if errors.Is(err, VersionMismatch) {
				// this indicates that the record has been rolled forward
				// by another transaction.
				return nil
			}
			return err
		})
	}
	eg.Wait()
	logger.Log.Debugw("Datastore.Commit() finishes", "TxnId", r.Txn.TxnId)
	r.clear()
	return nil
}

func (r *Datastore) commitInRemote() error {
	infoList := make([]CommitInfo, 0, len(r.writeCache))
	for _, item := range r.writeCache {
		infoList = append(infoList, CommitInfo{Key: item.Key(), Version: item.Version()})
	}

	err := r.Txn.RemoteCommit(r.Name, infoList)
	if err != nil {
		logger.Log.Infow("Remote commit failed", "TxnId", r.Txn.TxnId)
		return err
	}
	return err
}

// Abort discards the changes made in the current transaction.
//
//   - If hasCommitted is false, it clears the write cache.
//   - If hasCommitted is true, it rolls back the changes made by the current transaction.
//
// It returns an error if there is any issue during the rollback process.
func (r *Datastore) Abort(hasCommitted bool) error {

	if !hasCommitted {
		r.clear()
		return nil
	}

	if r.Txn.isRemote {
		keyList := make([]string, 0, len(r.writeCache))
		for _, item := range r.writeCache {
			keyList = append(keyList, item.Key())
		}
		return r.Txn.RemoteAbort(r.Name, keyList)
	}

	for _, v := range r.writeCache {
		item, err := r.conn.GetItem(v.Key())
		if err != nil {
			return err
		}
		// if the record has been modified by this transaction
		curGroupKeyList := strings.Join(r.Txn.GroupKeyUrls, ",")
		if item.GroupKeyList() == curGroupKeyList {
			r.rollback(item)
		}
	}
	r.clear()
	return nil
}

func (r *Datastore) OnePhaseCommit() error {
	if len(r.writeCache) == 0 {
		return nil
	}
	// there is only one record in the writeCache
	for _, item := range r.writeCache {
		return r.conditionalUpdate(item)
	}
	return nil
}

// rollback overwrites the record with the application data
// and metadata that found in field Prev.
// if the `Prev` is empty, it simply deletes the record
func (r *Datastore) rollback(item DataItem) (DataItem, error) {

	if item.Prev() == "" {
		item.SetIsDeleted(true)
		item.SetTxnState(config.COMMITTED)
		newVer, err := r.conn.ConditionalUpdate(item.Key(), item, false)
		if err != nil {
			return nil, errors.Join(errors.New("rollback failed"), err)
		}
		item.SetVersion(newVer)
		return item, err
	}

	newItem, err := r.getPrevItem(item)
	if err != nil {
		return nil, errors.Join(errors.New("rollback failed"), err)
	}
	// try to rollback through ConditionalUpdate
	newItem.SetVersion(item.Version())
	newVer, err := r.conn.ConditionalUpdate(item.Key(), newItem, false)
	// err = r.conn.PutItem(item.Key, newItem)
	if err != nil {
		return nil, errors.Join(errors.New("rollback failed"), err)
	}
	// update the version
	newItem.SetVersion(newVer)

	return newItem, err
}

// rollForward makes the record metadata with COMMITTED state
func (r *Datastore) rollForward(item DataItem) (DataItem, error) {

	item.SetTxnState(config.COMMITTED)
	newVer, err := r.conn.ConditionalUpdate(item.Key(), item, false)
	if err != nil {
		return nil, errors.Join(errors.New("rollForward failed"), err)
	}
	item.SetVersion(newVer)
	return item, err
}

// getPrevItem retrieves the previous item of the given DataItem.
// It deserializes the "Prev" field of the item and returns the deserialized DataItem.
// If there is an error during deserialization, it returns an empty DataItem and the error.
func (r *Datastore) getPrevItem(item DataItem) (DataItem, error) {
	preItem := r.itemFactory.NewDataItem(ItemOptions{})
	err := r.se.Deserialize([]byte(item.Prev()), &preItem)
	if err != nil {
		return nil, err
	}
	return preItem, nil
}

// getValue retrieves the value of a DataItem from the Datastore and deserializes it into the provided value.
// It uses the Datastore's serializer to deserialize the value.
// If an error occurs during deserialization, it is returned.
func (r *Datastore) getValue(item DataItem, value any) error {
	return r.se.Deserialize([]byte(item.Value()), value)
}

// GetName returns the name of the Datastore.
func (r *Datastore) GetName() string {
	return r.Name
}

// SetTxn sets the transaction for the MemoryDatastore.
// It takes a pointer to a Transaction as input and assigns it to the Txn field of the MemoryDatastore.
func (r *Datastore) SetTxn(txn *Transaction) {
	r.Txn = txn
}

// SetSerializer sets the serializer for the Datastore.
// The serializer is used to serialize and deserialize data when storing and retrieving it from Redis.
func (r *Datastore) SetSerializer(se serializer.Serializer) {
	r.se = se
}

// func (r *Datastore) CreateGroupKeyList(key string, txnState config.State, tCommit int64) (config.State, error) {

// 	if config.Debug.DebugMode {
// 		time.Sleep(config.Debug.HTTPAdditionalLatency)
// 	}
// 	groupKeyItem := GroupKeyItem{
// 		TxnState: txnState,
// 		TCommit:  tCommit,
// 	}
// 	groupKeyStr, err := json.Marshal(groupKeyItem)

// 	if err != nil {
// 		return -1, errors.Join(errors.New("GroupKey marshal failed"), err)
// 	}

// 	existValue, err := r.conn.AtomicCreate(key, util.ToString(groupKeyStr))
// 	if err != nil {
// 		if err.Error() == "key exists" {
// 			existKeyItem := GroupKeyItem{}
// 			err := json.Unmarshal([]byte(existValue), &existKeyItem)
// 			if err != nil {
// 				return -1, err
// 			}
// 			oldState := config.State(existKeyItem.TxnState)
// 			return oldState, errors.New(KeyExists)
// 		} else {
// 			return -1, err
// 		}
// 	}
// 	return -1, nil
// }

// func (r *Datastore) DeleteGroupKeyList(key string) error {
// 	return r.conn.Delete(key)
// }

// Copy returns a new instance of Datastore with the same name and connection.
// It is used to create a copy of the Datastore object.
func (r *Datastore) Copy() Datastorer {
	return NewDatastore(r.Name, r.conn, r.itemFactory)
}

func (r *Datastore) GetConn() Connector {
	return r.conn
}

func (r *Datastore) GetWriteCacheSize() int {
	return len(r.writeCache)
}

func (r *Datastore) clear() {
	r.readCache = make(map[string]DataItem)
	r.writeCache = make(map[string]DataItem)
	// r.writtenSet = util.NewConcurrentMap[bool]()
	r.invisibleSet = make(map[string]bool)
}
