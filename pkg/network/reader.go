package network

import (
	"errors"
	"time"

	"github.com/kkkzoz/oreo/internal/util"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/txn"
)

type Reader struct {
	conn        txn.Connector
	itemFactory txn.DataItemFactory
	se          serializer.Serializer
}

func NewReader(conn txn.Connector, itemFactory txn.DataItemFactory, se serializer.Serializer) *Reader {
	conn.Connect()
	return &Reader{
		conn:        conn,
		itemFactory: itemFactory,
		se:          se,
	}
}

// If the record is marked as IsDeleted, this function will return it.
//
// Let the upper layer decide what to do with it
func (r *Reader) Read(key string, ts time.Time) (txn.DataItem, error) {
	item, err := r.conn.GetItem(key)
	if err != nil {
		return nil, err
	}

	resItem, err := r.basicVisibilityProcessor(item, ts)
	if err != nil {
		return nil, err
	}
	logicFunc := func(curItem txn.DataItem, isFound bool) (txn.DataItem, error) {
		if !isFound {
			return nil, errors.New("key not found")
		}
		if config.Config.MaxRecordLength > 2 {
			curItem.SetPrev("")
			curItem.SetVersion("")
		}
		return curItem, nil
	}
	return r.treatAsCommitted(resItem, ts, logicFunc)
}

// basicVisibilityProcessor performs basic visibility processing on a DataItem.
// It tries to bring the item to the COMMITTED state by performing rollback or rollforward operations.
func (r *Reader) basicVisibilityProcessor(item txn.DataItem, startTime time.Time) (txn.DataItem, error) {
	// function to perform the rollback operation
	rollbackFunc := func() (txn.DataItem, error) {
		item, err := r.rollback(item)
		if err != nil {
			return nil, err
		}

		if item.Empty() {
			return nil, errors.New("key not found")
		}
		return item, err
	}

	// function to perform the rollforward operation
	rollforwardFunc := func() (txn.DataItem, error) {
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
		state, err := r.readTSR(item.TxnId())
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
		if item.TLease().Before(startTime) {
			// the corresponding transaction is considered ABORTED
			// TODO: we can retry here
			err := r.writeTSR(item.TxnId(), config.ABORTED)
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
		// r.invisibleSet[item.Key()] = true
		// if prev is empty
		if item.Prev() == "" {
			return nil, errors.New("key not found")
		}

		return r.getPrevItem(item)
		// return DataItem{}, DirtyRead
	}
	return nil, errors.New("key not found")
}

// rollback overwrites the record with the application data
// and metadata that found in field Prev.
// if the `Prev` is empty, it simply deletes the record
func (r *Reader) rollback(item txn.DataItem) (txn.DataItem, error) {

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
func (r *Reader) rollForward(item txn.DataItem) (txn.DataItem, error) {

	item.SetTxnState(config.COMMITTED)
	newVer, err := r.conn.ConditionalUpdate(item.Key(), item, false)
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

func (r *Reader) readTSR(txnId string) (config.State, error) {
	var txnState config.State
	state, err := r.conn.Get(txnId)
	if err != nil {
		return txnState, err
	}
	txnState = config.State(util.ToInt(state))
	return txnState, nil
}

func (r *Reader) writeTSR(txnId string, txnState config.State) error {
	return r.conn.Put(txnId, util.ToString(txnState))
}

// treatAsCommitted treats a DataItem as committed, finds a corresponding version
// according to its timestamp, and performs the given logic function on it.
func (r *Reader) treatAsCommitted(item txn.DataItem,
	startTime time.Time, logicFunc func(txn.DataItem, bool) (txn.DataItem, error)) (txn.DataItem, error) {
	curItem := item
	for i := 1; i <= config.Config.MaxRecordLength; i++ {

		if curItem.TValid().Before(startTime) {
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
			return nil, err
		}
		curItem = preItem
	}
	return nil, errors.New("key not found")
}
