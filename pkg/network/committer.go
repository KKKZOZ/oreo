package network

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alitto/pond/v2"
	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"golang.org/x/sync/errgroup"
)

type Committer struct {
	connMap     map[string]txn.Connector
	reader      Reader
	se          serializer.Serializer
	itemFactory txn.DataItemFactory
	timeSource  timesource.TimeSourcer
	pool        pond.Pool
}

func NewCommitter(connMap map[string]txn.Connector, reader Reader, se serializer.Serializer, itemFactory txn.DataItemFactory, timeSource timesource.TimeSourcer) *Committer {

	pool := pond.NewPool(200)

	// conn.Connect()
	return &Committer{
		connMap:     connMap,
		reader:      reader,
		se:          se,
		itemFactory: itemFactory,
		timeSource:  timeSource,
		pool:        pool,
	}
}

func (c *Committer) validate(dsName string, cfg txn.RecordConfig,
	validationMap map[string]txn.PredicateInfo) error {
	if cfg.ReadStrategy == config.Pessimistic {
		return nil
	}

	var eg errgroup.Group
	for gkl, predicate := range validationMap {
		gk := gkl
		pred := predicate
		if pred.ItemKey == "" {
			return errors.New("validation failed due to predicate item's empty key")
		}
		eg.Go(func() error {
			urlList := strings.Split(gk, ",")
			groupKey, err := c.reader.getGroupKey(urlList)
			if err != nil {
				// For AssumeAbort
				if cfg.ReadStrategy == config.AssumeAbort {
					if pred.LeaseTime.Before(time.Now()) {
						key := pred.ItemKey
						err := c.rollbackFromConn(dsName, key)
						if err != nil {
							return errors.Join(errors.New("validation failed in fine-AA mode"), err)
						} else {
							return nil
						}
					} else {
						errMsg := fmt.Sprintf("[getDSR err: %v] validation failed due to unknown status", err)
						return errors.New(errMsg)
					}
				}

				// For AssumeCommit
				errMsg := fmt.Sprintf("[getDSR err: %v] validation failed due to unknown status", err)
				return errors.New(errMsg)
			}

			// all group keys are found
			// AssumeCommit: all group keys are committed
			if cfg.ReadStrategy == config.AssumeCommit {
				if txn.CommittedForAll(groupKey) {
					fmt.Printf("all group keys are committed, key: %v\n", pred.ItemKey)
					return nil
				} else {
					return errors.New("validation failed due to false assumption")
				}
			}

			// AssumeAbort: at least one group key is aborted
			if cfg.ReadStrategy == config.AssumeAbort {
				if txn.AtLeastOneAborted(groupKey) {
					return nil
				} else {
					return errors.New("validation failed due to false assumption")
				}
			}

			return errors.New("validation failed due to unknown read strategy")
		})
	}

	return eg.Wait()
}

func (c *Committer) Prepare(dsName string, itemList []txn.DataItem,
	startTime int64, cfg txn.RecordConfig,
	validateMap map[string]txn.PredicateInfo) (map[string]string, int64, error) {

	debugStart := time.Now()

	err := c.validate(dsName, cfg, validateMap)
	logger.Log.Debugw("After validation", "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint", "cfg.ConcurrentOptimizationLevel", cfg.ConcurrentOptimizationLevel)
	if err != nil {
		return nil, 0, err
	}

	var tCommit int64

	if cfg.AblationLevel >= 3 {
		tCommit, err = c.timeSource.GetTime("commit")
		if err != nil {
			return nil, 0, errors.New("GetTime error: " + err.Error())
		}
	}

	var mu sync.Mutex
	// var eg errgroup.Group
	versionMap := make(map[string]string)

	subPool := c.pool.NewSubpool(5)
	taskGroup := subPool.NewGroup()

	for _, it := range itemList {
		item := it
		taskGroup.SubmitErr(func() error {
			var doCreate bool
			// if this item follows the read-modify-write pattern
			if item.Version() != "" {
				doCreate = false
			} else {
				// else we do a txn Read to determine its version
				dbItem, _, _, err := c.reader.Read(dsName, item.Key(), startTime, cfg, false)
				if err != nil && err.Error() != "key not found" {
					logger.Log.Errorw("Read error", "error", err)
					return err
				}
				if dbItem == nil {
					doCreate = true
				} else {
					doCreate = false
				}
				// logger.Log.Debugw("do a txn Read to determine the record version", "dbItem", dbItem)
				item, _ = c.updateMetadata(item, dbItem, 0, cfg)
			}

			// add TCommit to the item
			item.SetTValid(tCommit)
			ver, err := c.connMap[dsName].ConditionalUpdate(item.Key(), item, doCreate)

			mu.Lock()
			defer mu.Unlock()
			versionMap[item.Key()] = ver
			return err
		})
	}
	err = taskGroup.Wait()
	if err != nil {
		if cfg.AblationLevel >= 4 {
			_ = c.createGroupKey(dsName, itemList[0], config.ABORTED, tCommit)
		}
		return nil, 0, err
	}
	logger.Log.Debugw("After eg.Wait()", "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")

	if cfg.AblationLevel >= 4 {
		// create the corresponding group key
		if len(itemList) > 0 {
			err = c.createGroupKey(dsName, itemList[0], config.COMMITTED, tCommit)
			if err != nil {
				return nil, tCommit, fmt.Errorf("failed to create the group key: %v", err)
			}
			return versionMap, tCommit, nil
		}
	}

	return versionMap, tCommit, nil
}

func (c *Committer) createGroupKey(dsName string, item txn.DataItem, state config.State, tCommit int64) error {
	singleGK := strings.Split(item.GroupKeyList(), ",")[0]
	txnId := strings.Split(singleGK, ":")[1]
	url := dsName + ":" + txnId
	return c.reader.createSingleGroupKey(url, state, tCommit)
}

func (c *Committer) Abort(dsName string, keyList []string, groupKeyList string) error {
	// var eg errgroup.Group
	subPool := c.pool.NewSubpool(5)
	taskGroup := subPool.NewGroup()
	for _, k := range keyList {
		key := k
		taskGroup.SubmitErr(func() error {
			item, err := c.connMap[dsName].GetItem(key)
			if err != nil {
				return err
			}
			if item.GroupKeyList() == groupKeyList {
				_, err = c.rollback(dsName, item)
				return err
			} else {
				return nil
			}
		})
	}
	return taskGroup.Wait()
}

func (c *Committer) Commit(dsName string, infoList []txn.CommitInfo, tCommit int64) error {
	// var eg errgroup.Group
	subPool := c.pool.NewSubpool(5)
	taskGroup := subPool.NewGroup()
	for _, info := range infoList {
		item := info
		taskGroup.SubmitErr(func() error {
			_, err := c.connMap[dsName].ConditionalCommit(item.Key, item.Version, tCommit)
			return err
		})
	}
	return taskGroup.Wait()
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
func (c *Committer) truncate(newItem txn.DataItem, cfg txn.RecordConfig) (txn.DataItem, error) {
	maxLen := cfg.MaxRecordLen

	if newItem.LinkedLen() > maxLen {
		stack := util.NewStack[txn.DataItem]()
		stack.Push(newItem)
		curItem := &newItem
		for i := 1; i <= maxLen-1; i++ {
			preItem, err := c.getPrevItem(*curItem)
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
			bs, err := c.se.Serialize(tarItem)
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
func (c *Committer) updateMetadata(newItem txn.DataItem,
	oldItem txn.DataItem, commitTime int64, cfg txn.RecordConfig) (txn.DataItem, error) {
	if oldItem == nil {
		newItem.SetLinkedLen(1)
	} else {
		newItem.SetLinkedLen(oldItem.LinkedLen() + 1)
		bs, err := c.se.Serialize(oldItem)
		if err != nil {
			return nil, err
		}
		newItem.SetPrev(string(bs))
		newItem.SetVersion(oldItem.Version())
	}

	// truncate the record
	newItem, err := c.truncate(newItem, cfg)
	if err != nil {
		return nil, err
	}

	newItem.SetTxnState(config.PREPARED)
	newItem.SetTValid(commitTime)
	// TODO: time.Now() is temporary
	newItem.SetTLease(time.Now().Add(config.Config.LeaseTime))
	return newItem, nil
}

func (c *Committer) getPrevItem(item txn.DataItem) (txn.DataItem, error) {
	preItem := c.itemFactory.NewDataItem(txn.ItemOptions{})
	err := c.se.Deserialize([]byte(item.Prev()), &preItem)
	if err != nil {
		return nil, err
	}
	return preItem, nil
}

// rollback overwrites the record with the application data
// and metadata that found in field Prev.
// if the `Prev` is empty, it simply deletes the record
func (c *Committer) rollback(dsName string, item txn.DataItem) (txn.DataItem, error) {

	if item.Prev() == "" {
		item.SetIsDeleted(true)
		item.SetTxnState(config.COMMITTED)
		newVer, err := c.connMap[dsName].ConditionalUpdate(item.Key(), item, false)
		if err != nil {
			return nil, errors.Join(errors.New("rollback failed"), err)
		}
		item.SetVersion(newVer)
		return item, err
	}

	newItem, err := c.getPrevItem(item)
	if err != nil {
		return nil, errors.Join(errors.New("rollback failed"), err)
	}
	// try to rollback through ConditionalUpdate
	newItem.SetVersion(item.Version())
	newVer, err := c.connMap[dsName].ConditionalUpdate(item.Key(), newItem, false)
	// err = r.conn.PutItem(item.Key, newItem)
	if err != nil {
		return nil, errors.Join(errors.New("rollback failed"), err)
	}
	// update the version
	newItem.SetVersion(newVer)

	return newItem, err
}

func (c *Committer) rollbackFromConn(dsName string, key string) error {

	item, err := c.connMap[dsName].GetItem(key)
	if err != nil {
		return err
	}

	if item.TxnState() != config.PREPARED {
		return errors.New("rollback failed due to wrong state")
	}
	// if item.TxnId() != txnId {
	// 	// fmt.Printf("item: %v   txnId: %v\n",item,txnId)
	// 	return errors.New("rollback failed due to wrong txnId")
	// }

	if item.TLease().Before(time.Now()) {
		successNum := c.reader.createGroupKey(strings.Split(item.GroupKeyList(), ","), config.ABORTED, 0)
		if successNum == 0 {
			return fmt.Errorf("failed to rollback the record because none of the group keys are created")
		}
		_, err = c.rollback(dsName, item)
		return err
	}
	fmt.Printf("rollbackFromConn OK\n")
	return nil
}
