package network

import (
	"errors"
	"time"

	"github.com/oreo-dtx-lab/oreo/internal/util"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
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
	// conn.Connect()
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
		state, err := r.readTSR(cfg.GlobalName, item.TxnId())
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
		// that is, item's TLease < current time
		// we should roll back the record
		if item.TLease().Before(time.Now()) {
			// the corresponding transaction is considered ABORTED
			// TODO: we can retry here
			err := r.writeTSR(cfg.GlobalName, item.TxnId(), config.ABORTED)
			if err != nil {
				return nil, txn.Normal, err
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

func (r *Reader) readTSR(dsName string, txnId string) (config.State, error) {
	var txnState config.State
	state, err := r.connMap[dsName].Get(txnId)
	if err != nil {
		return txnState, err
	}
	txnState = config.State(util.ToInt(state))
	return txnState, nil
}

func (r *Reader) writeTSR(dsName string, txnId string, txnState config.State) error {
	return r.connMap[dsName].Put(txnId, util.ToString(txnState))
}

func (r *Reader) createTSR(dsName string, txnId string, txnState config.State) (config.State, error) {
	oldValue, err := r.connMap[dsName].AtomicCreate(txnId, util.ToString(txnState))
	if err != nil {
		if err.Error() == "key exists" {
			oldState := config.State(util.ToInt(oldValue))
			return oldState, errors.New("key exists")
		} else {
			return -1, err
		}
	}
	return -1, nil
}

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
