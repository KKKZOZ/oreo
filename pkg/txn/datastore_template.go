// a pure interface version
package txn

import (
	"cmp"
	"slices"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/logger"
	"github.com/kkkzoz/oreo/pkg/serializer"
)

var _ Datastorer = (*Datastore)(nil)
var _ TSRMaintainer = (*Datastore)(nil)

const (
	EMPTY         string = ""
	RETRYINTERVAL        = 10 * time.Millisecond
)

// Datastore represents a datastorer implementation using the underlying connector.
type Datastore struct {

	// Name is the name of the datastore.
	Name string

	// Txn is the current transaction.
	Txn *Transaction

	// conn is the connector interface used by the datastore.
	conn Connector

	// readCache is the cache for read operations in Datastore.
	readCache map[string]DataItem

	// writeCache is the cache for write operations in Datastore.
	writeCache map[string]DataItem

	// invisibleSet is the set of keys that are not visible to the current transaction.
	invisibleSet map[string]bool

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
		Name:         name,
		conn:         conn,
		readCache:    make(map[string]DataItem),
		writeCache:   make(map[string]DataItem),
		invisibleSet: make(map[string]bool),
		se:           config.Config.Serializer,
		itemFactory:  factory,
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

	// else get if from connection
	return r.readFromConn(key, value)
}

func (r *Datastore) readFromConn(key string, value any) error {
	item, err := r.conn.GetItem(key)
	if err != nil {
		return err
	}

	item, err = r.dirtyReadChecker(item)
	if err != nil {
		return err
	}

	resItem, err := r.basicVisibilityProcessor(item)
	if err != nil {
		return err
	}

	logicFunc := func(curItem DataItem, isFound bool) error {
		r.mu.Lock()
		defer r.mu.Unlock()
		// if the record has been deleted
		if !isFound || curItem.IsDeleted() {
			if curItem.IsDeleted() {
				// put into cache anyway
				r.readCache[curItem.Key()] = curItem
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

	if item.TxnState() == config.COMMITTED {
		return item, nil
	}
	if item.TxnState() == config.PREPARED {
		state, err := r.Txn.GetTSRState(item.TxnId())
		if err == nil {
			switch state {
			// if TSR exists and the TSR is in COMMITTED state
			// roll forward the record
			case config.COMMITTED:
				return rollforwardFunc()
			// if TSR exists and the TSR is in ABORTED state
			// we should roll back the record
			// because the transaction that modified the record has been aborted
			case config.ABORTED:
				return rollbackFunc()
			}
		}
		// if TSR does not exist
		// and if t_lease has expired
		// we should roll back the record
		if item.TLease().Before(r.Txn.TxnStartTime) {
			// the corresponding transaction is considered ABORTED
			// TODO: we can retry here
			err := r.Txn.WriteTSR(item.TxnId(), config.ABORTED)
			if err != nil {
				return nil, err
			}
			return rollbackFunc()
		}
		// the corresponding transaction is running,
		// we should try previous record instead of raising an error

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
		// return DataItem{}, DirtyRead
	}
	return nil, errors.New(KeyNotFound)
}

// treatAsCommitted treats a DataItem as committed, finds a corresponding version
// according to its timestamp, and performs the given logic function on it.
func (r *Datastore) treatAsCommitted(item DataItem, logicFunc func(DataItem, bool) error) error {
	curItem := item
	for i := 1; i <= config.Config.MaxRecordLength; i++ {

		if curItem.TValid().Before(r.Txn.TxnStartTime) {
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
		Key:   key,
		Value: str,
		TxnId: r.Txn.TxnId,
	})
	return r.writeToCache(cacheItem)
}

// writeToCache writes the given DataItem to the cache.
// It will find the corresponding version of the item.
//   - If the item already exists in the read cache, it follows the read-modified-commit pattern
//   - If this is a direct write, it will set the version to -1
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
		Key:       key,
		TxnId:     r.Txn.TxnId,
		IsDeleted: true,
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

	// if the cacheItem follows read-modified-write pattern,
	// it already has a valid version, we can skip the read step.
	if cacheItem.Version() != "" {
		dbItem := r.readCache[cacheItem.Key()]
		return r.doConditionalUpdate(cacheItem, dbItem)
	}

	// else we read from connection
	err := r.readFromConn(cacheItem.Key(), nil)
	if err != nil && !errors.Is(err, KeyNotFound) {
		return err
	}
	dbItem := r.readCache[cacheItem.Key()]
	// if the record is dropped by the repeatable read rule
	if res, ok := r.invisibleSet[cacheItem.Key()]; ok && res {
		dbItem = nil
	}
	return r.doConditionalUpdate(cacheItem, dbItem)
}

// truncate truncates the linked list of DataItems if the length exceeds the maximum record length defined in the configuration.
// It takes a pointer to a DataItem as input and returns the truncated DataItem and an error, if any.
// If the length of the linked list is greater than the maximum record length, it creates a stack of DataItems and pops the items from the stack until the length is reduced to the maximum record length.
// It then updates the Prev and LinkedLen fields of the DataItems in the stack accordingly.
// Finally, it returns the last popped DataItem as the truncated DataItem.
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
// If the oldItem is empty, it sets the LinkedLen of the newItem to 1.
// Otherwise, it increments the LinkedLen of the newItem by 1 and sets the Prev and Version fields based on the oldItem.
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
	newItem.SetTLease(r.Txn.TxnCommitTime.Add(config.Config.LeaseTime))
	return newItem, nil
}

// Prepare prepares the Datastore for commit.
func (r *Datastore) Prepare() error {
	items := make([]DataItem, 0, len(r.writeCache))
	for _, v := range r.writeCache {
		items = append(items, v)
	}
	// sort records by key
	slices.SortFunc(
		items, func(i, j DataItem) int {
			return cmp.Compare(i.Key(), j.Key())
		},
	)
	if config.Config.ConcurrentOptimizationLevel < config.PARALLELIZE_ON_UPDATE {
		for _, item := range items {
			if err := r.conditionalUpdate(item); err != nil {
				return err
			}
		}
		return nil
	}

	resChan := make(chan error, len(items))
	for _, item := range items {
		go func(it DataItem) { resChan <- r.conditionalUpdate(it) }(item)
	}

	//TODO: replace it with errgroup
	success := true
	var cause error
	for i := 0; i < len(items); i++ {
		err := <-resChan
		if err != nil {
			success, cause = false, err
		}
	}
	if !success {
		return cause
	}
	return nil
}

// Commit updates the state of records in the data store to COMMITTED.
// It iterates over the write cache and updates each record's state to COMMITTED.
// After updating the records, it clears the write cache.
// Returns an error if there is any issue updating the records.
func (r *Datastore) Commit() error {

	logger.Log.Debugw("Datastore.Commit() starts", "TxnId", r.Txn.TxnId)
	// update record's state to the COMMITTED state in the data store
	var wg sync.WaitGroup
	for _, item := range r.writeCache {
		wg.Add(1)
		go func(item DataItem) {
			defer wg.Done()
			item.SetTxnState(config.COMMITTED)
			util.RetryHelper(3, RETRYINTERVAL, func() error {
				_, err := r.conn.ConditionalUpdate(item.Key(), item, false)
				if errors.Is(err, VersionMismatch) {
					// this indicates that the record has been rolled forward
					// by another transaction.
					return nil
				}
				return err
			})
		}(item)
	}
	wg.Wait()
	logger.Log.Debugw("Datastore.Commit() finishes", "TxnId", r.Txn.TxnId)
	// clear the cache
	r.writeCache = make(map[string]DataItem)
	r.readCache = make(map[string]DataItem)
	r.invisibleSet = make(map[string]bool)
	return nil
}

// Abort discards the changes made in the current transaction.
// If hasCommitted is false, it clears the write cache.
// If hasCommitted is true, it rolls back the changes made by the current transaction.
// It returns an error if there is any issue during the rollback process.
func (r *Datastore) Abort(hasCommitted bool) error {

	if !hasCommitted {
		r.writeCache = make(map[string]DataItem)
		r.invisibleSet = make(map[string]bool)
		return nil
	}

	for _, v := range r.writeCache {
		item, err := r.conn.GetItem(v.Key())
		if err != nil {
			return err
		}
		// if the record has been modified by this transaction
		curTxnId := r.Txn.TxnId
		if item.TxnId() == curTxnId {
			r.rollback(item)
		}
	}
	r.readCache = make(map[string]DataItem)
	r.writeCache = make(map[string]DataItem)
	r.invisibleSet = make(map[string]bool)
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

// ReadTSR reads the transaction state from the Redis datastore.
// It takes a transaction ID as input and returns the corresponding state and an error, if any.
func (r *Datastore) ReadTSR(txnId string) (config.State, error) {
	var txnState config.State
	state, err := r.conn.Get(txnId)
	if err != nil {
		return txnState, err
	}
	txnState = config.State(util.ToInt(state))
	return txnState, nil
}

// WriteTSR writes the transaction state (txnState) associated with the given transaction ID (txnId) to the Redis datastore.
// It returns an error if the write operation fails.
func (r *Datastore) WriteTSR(txnId string, txnState config.State) error {
	return r.conn.Put(txnId, util.ToString(txnState))
}

// DeleteTSR deletes a transaction with the given transaction ID from the Redis datastore.
// It returns an error if the deletion operation fails.
func (r *Datastore) DeleteTSR(txnId string) error {
	return r.conn.Delete(txnId)
}

// Copy returns a new instance of Datastore with the same name and connection.
// It is used to create a copy of the Datastore object.
func (r *Datastore) Copy() Datastorer {
	return NewDatastore(r.Name, r.conn, r.itemFactory)
}
