package network

import (
	"errors"
	"sync"
	"time"

	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"
	"golang.org/x/sync/errgroup"
)

type Committer struct {
	conn        txn.Connector
	reader      Reader
	se          serializer.Serializer
	itemFactory txn.DataItemFactory
}

func NewCommitter(conn txn.Connector, se serializer.Serializer, itemFactory txn.DataItemFactory) *Committer {
	conn.Connect()
	return &Committer{
		conn:        conn,
		reader:      *NewReader(conn, itemFactory, se),
		se:          se,
		itemFactory: itemFactory,
	}
}

func (c *Committer) Prepare(itemList []txn.DataItem,
	startTime time.Time, commitTime time.Time) (map[string]string, error) {

	var mu sync.Mutex
	var eg errgroup.Group
	versionMap := make(map[string]string)

	for _, it := range itemList {
		item := it
		eg.Go(func() error {
			var doCreate bool
			// if this item follows the read-modify-write pattern
			if item.Version() != "" {
				doCreate = false
			} else {
				// else we do a txn Read to determine its version
				dbItem, err := c.reader.Read(item.Key(), startTime, false)
				if err != nil && err.Error() != "key not found" {
					return err
				}
				if dbItem == nil {
					doCreate = true
				} else {
					doCreate = false
				}
				item, _ = c.updateMetadata(item, dbItem, commitTime)
			}
			ver, err := c.conn.ConditionalUpdate(item.Key(), item, doCreate)
			mu.Lock()
			defer mu.Unlock()
			versionMap[item.Key()] = ver
			return err
		})
	}
	err := eg.Wait()
	return versionMap, err
}

func (c *Committer) Abort(keyList []string, txnId string) error {
	var eg errgroup.Group
	for _, k := range keyList {
		key := k
		eg.Go(func() error {
			item, err := c.conn.GetItem(key)
			if err != nil {
				return err
			}
			if item.TxnId() == txnId {
				_, err = c.rollback(item)
				return err
			} else {
				return nil
			}
		})
	}
	return eg.Wait()
}

func (c *Committer) Commit(infoList []txn.CommitInfo) error {
	var eg errgroup.Group
	for _, info := range infoList {
		item := info
		eg.Go(func() error {
			_, err := c.conn.ConditionalCommit(item.Key, item.Version)
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
func (c *Committer) truncate(newItem txn.DataItem) (txn.DataItem, error) {
	maxLen := config.Config.MaxRecordLength

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
	oldItem txn.DataItem, commitTime time.Time) (txn.DataItem, error) {
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
	newItem, err := c.truncate(newItem)
	if err != nil {
		return nil, err
	}

	newItem.SetTxnState(config.PREPARED)
	newItem.SetTValid(commitTime)
	newItem.SetTLease(commitTime.Add(config.Config.LeaseTime))
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
func (c *Committer) rollback(item txn.DataItem) (txn.DataItem, error) {

	if item.Prev() == "" {
		item.SetIsDeleted(true)
		item.SetTxnState(config.COMMITTED)
		newVer, err := c.conn.ConditionalUpdate(item.Key(), item, false)
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
	newVer, err := c.conn.ConditionalUpdate(item.Key(), newItem, false)
	// err = r.conn.PutItem(item.Key, newItem)
	if err != nil {
		return nil, errors.Join(errors.New("rollback failed"), err)
	}
	// update the version
	newItem.SetVersion(newVer)

	return newItem, err
}
