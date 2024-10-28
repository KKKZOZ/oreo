package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"golang.org/x/sync/errgroup"
)

const (
	ReadFailed = "read failed due to unknown txn status"
)

type Reader struct {
	connMap     map[string]txn.Connector
	itemFactory txn.DataItemFactory
	se          serializer.Serializer
}

func NewReader(connMap map[string]txn.Connector, itemFactory txn.DataItemFactory, se serializer.Serializer) *Reader {
	return &Reader{
		connMap:     connMap,
		itemFactory: itemFactory,
		se:          se,
	}
}

// If the record is marked as IsDeleted, this function will return it.
//
// Let the upper layer decide what to do with it
func (r *Reader) Read(dsName string, key string, ts int64, cfg txn.RecordConfig,
	isRemoteCall bool) (txn.DataItem, txn.RemoteDataStrategy, error) {
	dataType := txn.Normal

	item, err := r.connMap[dsName].GetItem(key)
	if err != nil {
		return nil, dataType, err
	}

	resItem, dataType, err := r.basicVisibilityProcessor(dsName, item, ts, cfg)
	if err != nil {
		return nil, dataType, err
	}
	logicFunc := func(curItem txn.DataItem, isFound bool) (txn.DataItem, error) {
		if !isFound {
			if curItem.IsDeleted() {
				return nil, errors.New("key not found, item is deleted")
			}
			return nil, errors.New("key not found")
		}
		if isRemoteCall && cfg.MaxRecordLen > 2 {
			curItem.SetPrev("")
			curItem.SetVersion("")
		}
		return curItem, nil
	}

	item, err = r.treatAsCommitted(resItem, ts, logicFunc, cfg)
	return item, dataType, err
	// return r.treatAsCommitted(resItem, ts, logicFunc, cfg)
}

// basicVisibilityProcessor performs basic visibility processing on a DataItem.
// It tries to bring the item to the COMMITTED state by performing rollback or rollforward operations.
func (r *Reader) basicVisibilityProcessor(dsName string, item txn.DataItem,
	startTime int64, cfg txn.RecordConfig) (txn.DataItem, txn.RemoteDataStrategy, error) {
	// function to perform the rollback operation
	rollbackFunc := func() (txn.DataItem, txn.RemoteDataStrategy, error) {
		item, err := r.rollback(dsName, item)
		if err != nil {
			return nil, txn.Normal, err
		}

		if item.Empty() {
			return nil, txn.Normal, errors.New("key not found")
		}
		return item, txn.Normal, err
	}

	// function to perform the rollforward operation
	rollforwardFunc := func() (txn.DataItem, txn.RemoteDataStrategy, error) {
		item, err := r.rollForward(dsName, item)
		if err != nil {
			return nil, txn.Normal, err
		}
		return item, txn.Normal, nil
	}

	if item.TxnState() == config.COMMITTED {
		return item, txn.Normal, nil
	}
	if item.TxnState() == config.PREPARED {
		groupKeyList, err := r.getGroupKey(strings.Split(item.GroupKeyList(), ","))
		if err == nil {
			if txn.CommittedForAll(groupKeyList) {
				// if all the group keys are in COMMITTED state
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
		// if TSR does not exist
		// and if t_lease has expired
		// that is, item's TLease < current time
		// we should roll back the record
		if item.TLease().Before(time.Now()) {
			successNum := r.createGroupKey(strings.Split(item.GroupKeyList(), ","), config.ABORTED)
			if successNum == 0 {
				return nil, txn.Normal, errors.New("failed to rollback the record because none of the group keys are created")
			}
			return rollbackFunc()
		}
		// if TSR does not exist
		// and if the corresponding transaction is a concurrent transaction
		// that is, txn's TStart < item's TValid < item's TLease
		// we should try check the previous record
		if startTime < item.TValid() {

			// Origin Cherry Garcia would do
			if config.Debug.CherryGarciaMode {
				return nil, txn.Normal, errors.New(ReadFailed)
			}

			// a little trick here:
			// if the record is not found in the treatAsCommitted,
			// we should add it to the invisibleSet.
			// if the record can be found in the treatAsCommitted,
			// it will be stored in the readCache,
			// so we don't bother dirtyReadChecker anymore.

			// r.invisibleSet[item.Key()] = true

			// if prev is empty
			if item.Prev() == "" {
				return nil, txn.Normal, errors.New("key not found")
			}
			// get the previous record
			preItem, err := r.getPrevItem(item)
			return preItem, txn.Normal, err
		}

		if cfg.ReadStrategy == config.Pessimistic {
			return nil, txn.Normal, errors.New(ReadFailed)
		} else {
			switch cfg.ReadStrategy {
			case config.AssumeCommit:
				return item, txn.AssumeCommit, nil
			case config.AssumeAbort:
				if item.Prev() == "" {
					return nil, txn.Normal, errors.New("key not found in AssumeAbort")
				}
				preItem, err := r.getPrevItem(item)
				return preItem, txn.AssumeAbort, err
			}
		}

	}
	return nil, txn.Normal, errors.New("key not found")
}

// rollback overwrites the record with the application data
// and metadata that found in field Prev.
// if the `Prev` is empty, it simply deletes the record
func (r *Reader) rollback(dsName string, item txn.DataItem) (txn.DataItem, error) {

	if item.Prev() == "" {
		item.SetIsDeleted(true)
		item.SetTxnState(config.COMMITTED)
		newVer, err := r.connMap[dsName].ConditionalUpdate(item.Key(), item, false)
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
	newVer, err := r.connMap[dsName].ConditionalUpdate(item.Key(), newItem, false)
	// err = r.conn.PutItem(item.Key, newItem)
	if err != nil {
		return nil, errors.Join(errors.New("rollback failed"), err)
	}
	// update the version
	newItem.SetVersion(newVer)

	return newItem, err
}

// rollForward makes the record metadata with COMMITTED state
func (r *Reader) rollForward(dsName string, item txn.DataItem) (txn.DataItem, error) {

	item.SetTxnState(config.COMMITTED)
	newVer, err := r.connMap[dsName].ConditionalUpdate(item.Key(), item, false)
	if err != nil {
		return nil, errors.Join(errors.New("rollForward failed"), err)
	}
	item.SetVersion(newVer)
	return item, err
}

func (r *Reader) getPrevItem(item txn.DataItem) (txn.DataItem, error) {
	preItem := r.itemFactory.NewDataItem(txn.ItemOptions{})
	err := r.se.Deserialize([]byte(item.Prev()), &preItem)
	if err != nil {
		return nil, err
	}
	return preItem, nil
}

// func (r *Reader) readGroupKey(dsName string, txnId string) (txn.GroupKey, error) {

// 	groupKeyStr, err := r.connMap[dsName].Get(txnId)
// 	if err != nil {
// 		return txn.GroupKey{}, err
// 	}
// 	var keyItem txn.GroupKeyItem
// 	err = json.Unmarshal([]byte(groupKeyStr), &keyItem)

// 	if err != nil {
// 		return txn.GroupKey{}, err
// 	}
// 	groupKey := txn.GroupKey{
// 		Key:          txnId,
// 		GroupKeyItem: keyItem,
// 	}
// 	return groupKey, nil

// }

func (r *Reader) getGroupKey(urls []string) ([]txn.GroupKey, error) {
	groupKeys := make([]txn.GroupKey, len(urls))
	var mu sync.Mutex
	var eg errgroup.Group
	for _, urll := range urls {
		url := urll
		eg.Go(func() error {
			groupKey, err := r.getSingleGroupKey(url)
			if err != nil {
				return err
			}
			mu.Lock()
			groupKeys = append(groupKeys, groupKey)
			mu.Unlock()
			return nil
		})
	}
	err := eg.Wait()
	// This means at least one of the group key is not found
	if err != nil {
		return nil, err
	}

	return groupKeys, nil
}

func (r *Reader) getSingleGroupKey(url string) (txn.GroupKey, error) {
	tokens := strings.Split(url, ":")
	conn, ok := r.connMap[tokens[0]]
	if !ok {
		return txn.GroupKey{}, fmt.Errorf("connector to %s is not found", tokens[0])
	}
	groupKeyStr, err := conn.Get(tokens[1])
	if err != nil {
		return txn.GroupKey{}, err
	}
	var keyItem txn.GroupKeyItem
	err = json.Unmarshal([]byte(groupKeyStr), &keyItem)
	if err != nil {
		return txn.GroupKey{}, fmt.Errorf("failed to unmarshal group key item %s", groupKeyStr)
	}
	return *txn.NewGroupKey(url, keyItem.TxnState, keyItem.TCommit), nil
}

func (r *Reader) createGroupKey(urls []string, state config.State) int {

	resChan := make(chan error, len(urls))
	for _, urll := range urls {
		url := urll
		go func() {
			tokens := strings.Split(url, ":")
			conn, ok := r.connMap[tokens[0]]
			if !ok {
				resChan <- fmt.Errorf("connector to %s is not found", tokens[0])
				return
			}

			groupKey := txn.NewGroupKey(url, state, 0)
			groupKeyStr, err := json.Marshal(groupKey)
			if err != nil {
				resChan <- fmt.Errorf("failed to marshal group key item %s", groupKey)
				return
			}
			// CHECK: we do not need the returned value?
			_, err = conn.AtomicCreate(url, util.ToString(groupKeyStr))
			if err != nil {
				resChan <- err
				return
			}
			resChan <- nil
		}()
	}
	okInTotal := len(urls)
	for i := 0; i < len(urls); i++ {
		err := <-resChan
		if err != nil {
			okInTotal--
		}
	}
	return okInTotal
}

func (r *Reader) createSingleGroupKey(url string, state config.State) error {
	tokens := strings.Split(url, ":")
	conn, ok := r.connMap[tokens[0]]
	if !ok {
		return fmt.Errorf("connector to %s is not found", tokens[0])
	}

	groupKey := txn.NewGroupKey(url, state, 0)
	groupKeyStr, err := json.Marshal(groupKey)
	if err != nil {
		return fmt.Errorf("failed to marshal group key item %s", groupKey)
	}
	_, err = conn.AtomicCreate(url, util.ToString(groupKeyStr))
	if err != nil {
		return err
	}
	return nil
}

// func (r *Reader) createGroupKey(dsName string, txnId string, txnState config.State, tCommit int64) (config.State, error) {
// 	groupKeyItem := txn.GroupKeyItem{
// 		TxnState: txnState,
// 		TCommit:  tCommit,
// 	}
// 	groupKeyStr, err := json.Marshal(groupKeyItem)
// 	if err != nil {
// 		return -1, errors.Join(errors.New("GroupKey marshal failed"), err)
// 	}

// 	existValue, err := r.connMap[dsName].AtomicCreate(txnId, util.ToString(groupKeyStr))
// 	if err != nil {
// 		if err.Error() == "key exists" {
// 			existKeyItem := txn.GroupKeyItem{}
// 			err := json.Unmarshal([]byte(existValue), &existKeyItem)
// 			if err != nil {
// 				return -1, err
// 			}
// 			oldState := config.State(existKeyItem.TxnState)
// 			return oldState, errors.New("key exists")
// 		} else {
// 			return -1, err
// 		}
// 	}
// 	return -1, nil
// }

// treatAsCommitted treats a DataItem as committed, finds a corresponding version
// according to its timestamp, and performs the given logic function on it.
func (r *Reader) treatAsCommitted(item txn.DataItem,
	startTime int64, logicFunc func(txn.DataItem, bool) (txn.DataItem, error),
	cfg txn.RecordConfig) (txn.DataItem, error) {
	curItem := item
	for i := 1; i <= cfg.MaxRecordLen; i++ {

		if curItem.TValid() < startTime {
			// find the corresponding version,
			// do some business logic.
			return logicFunc(curItem, true)
		}
		if i == cfg.MaxRecordLen {
			break
		}
		// if prev is empty
		if curItem.Prev() == "" {
			return logicFunc(curItem, false)
		}

		// get the previous record
		preItem, err := r.getPrevItem(curItem)
		if err != nil {
			return nil, err
		}
		curItem = preItem
	}
	return nil, errors.New("key not found")
}
