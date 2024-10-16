package network

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"golang.org/x/sync/errgroup"
)

type Committer struct {
	connMap     map[string]txn.Connector
	reader      Reader
	se          serializer.Serializer
	itemFactory txn.DataItemFactory
}

func NewCommitter(connMap map[string]txn.Connector, se serializer.Serializer, itemFactory txn.DataItemFactory) *Committer {
	// conn.Connect()
	return &Committer{
		connMap:     connMap,
		reader:      *NewReader(connMap, itemFactory, se),
		se:          se,
		itemFactory: itemFactory,
	}
}

func (c *Committer) validate(dsName string, cfg txn.RecordConfig,
	validationMap map[string]txn.PredicateInfo) error {
	if cfg.ReadStrategy == config.Pessimistic {
		return nil
	}

	var eg errgroup.Group
	for tId, predicate := range validationMap {
		txnId := tId
		pred := predicate
		if pred.ItemKey == "" {
			return errors.New("validation failed due to predicate item's empty key")
		}
		eg.Go(func() error {
			curState, err := c.reader.readTSR(cfg.GlobalName, txnId)
			// fmt.Printf("curState: %v, err: %v\n",curState, err)
			if err != nil {
				if cfg.ReadStrategy == config.AssumeAbort {
					if pred.LeaseTime.Before(time.Now()) {
						// fmt.Printf("Leasetime has expired\n")
						key := pred.ItemKey
						err := c.rollbackFromConn(dsName, key, txnId, cfg.GlobalName)
						if err != nil {
							return errors.Join(errors.New("validation failed in AA mode"), err)
						} else {
							return nil
						}
					}
				}
				return errors.New("validation failed due to unknown status")
			}
			if curState != pred.State {
				return errors.New("validation failed due to false assumption")
			} else {
				return nil
			}
		})
	}

	return eg.Wait()

}

func (c *Committer) Prepare(dsName string, itemList []txn.DataItem,
	startTime int64, commitTime int64,
	cfg txn.RecordConfig, validateMap map[string]txn.PredicateInfo) (map[string]string, error) {

	debugStart := time.Now()

	err := c.validate(dsName, cfg, validateMap)
	logger.Log.Debugw("After validation", "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint", "cfg.ConcurrentOptimizationLevel", cfg.ConcurrentOptimizationLevel)
	if err != nil {
		return nil, err
	}

	var mu sync.Mutex
	var eg errgroup.Group
	versionMap := make(map[string]string)

	if cfg.ConcurrentOptimizationLevel == 0 {
		eg.SetLimit(1)
	}

	for _, it := range itemList {
		item := it
		eg.Go(func() error {
			var doCreate bool
			// if this item follows the read-modify-write pattern
			if item.Version() != "" {
				doCreate = false
				// fmt.Printf("TCommit: %v  TLease: %v\n",item.TValid(),item.TLease())
			} else {
				// fmt.Printf("do a txn Read\n")
				// else we do a txn Read to determine its version
				dbItem, _, err := c.reader.Read(dsName, item.Key(), startTime, cfg, false)
				if err != nil && err.Error() != "key not found" {
					return err
				}
				logger.Log.Debugw("error?", "err", err)
				if dbItem == nil {
					doCreate = true
				} else {
					doCreate = false
				}
				logger.Log.Debugw("do a txn Read to determine the record version", "dbItem", dbItem)
				item, _ = c.updateMetadata(item, dbItem, commitTime, cfg)
			}
			logger.Log.Debugw("Before c.connMap[dsName].ConditionalUpdate", "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")
			ver, err := c.connMap[dsName].ConditionalUpdate(item.Key(), item, doCreate)
			logger.Log.Debugw("After c.connMap[dsName].ConditionalUpdate", "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")
			// if err != nil {
			// 	fmt.Printf("ConditionalUpdate error: %v in %v, item: %v\n", err, dsName, item)
			// }

			mu.Lock()
			defer mu.Unlock()
			versionMap[item.Key()] = ver
			return err
		})
	}
	logger.Log.Debugw("Before eg.Wait()", "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")
	err = eg.Wait()
	logger.Log.Debugw("After eg.Wait()", "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")

	return versionMap, err
}

func (c *Committer) Abort(dsName string, keyList []string, txnId string) error {
	var eg errgroup.Group
	for _, k := range keyList {
		key := k
		eg.Go(func() error {
			item, err := c.connMap[dsName].GetItem(key)
			if err != nil {
				return err
			}
			if item.TxnId() == txnId {
				_, err = c.rollback(dsName, item)
				return err
			} else {
				return nil
			}
		})
	}
	return eg.Wait()
}

func (c *Committer) Commit(dsName string, infoList []txn.CommitInfo) error {
	var eg errgroup.Group
	for _, info := range infoList {
		item := info
		eg.Go(func() error {
			_, err := c.connMap[dsName].ConditionalCommit(item.Key, item.Version)
			return err
		})
	}
	return eg.Wait()
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

func (c *Committer) rollbackFromConn(dsName string, key string, txnId string, globalName string) error {

	item, err := c.connMap[dsName].GetItem(key)
	if err != nil {
		return err
	}

	if item.TxnState() != config.PREPARED {
		return errors.New("rollback failed due to wrong state")
	}
	if item.TxnId() != txnId {
		// fmt.Printf("item: %v   txnId: %v\n",item,txnId)
		return errors.New("rollback failed due to wrong txnId")
	}

	if item.TLease().Before(time.Now()) {
		curState, err := c.reader.createTSR(globalName, item.TxnId(), config.ABORTED)
		if err != nil {
			if err.Error() == "key exists" {
				if curState == config.COMMITTED {
					return errors.New("rollback failed because the corresponding transaction has committed")
				}
				// if curState == config.ABORTED
				// it means the transaction has been rolled back
				// so we can safely rollback the record
			} else {
				return err
			}
		}

		// err := r.Txn.WriteTSR(item.TxnId(), config.ABORTED)
		// if err != nil {
		// 	return err
		// }
		_, err = c.rollback(dsName, item)
		return err
	}
	fmt.Printf("rollbackFromConn OK\n")
	return nil
}
