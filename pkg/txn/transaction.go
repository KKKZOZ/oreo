package txn

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/locker"
)

type SourceType string

const (
	// EMPTY  SourceType = "EMPTY"
	LOCAL  SourceType = "LOCAL"
	GLOBAL SourceType = "GLOBAL"
)

// Transaction represents a transaction in the system.
// It contains information such as the transaction ID, state, timestamps,
// datastores, time source, oracle URL, and locker.
type Transaction struct {
	// TxnId is the unique identifier for the transaction.
	TxnId string
	// TxnStartTime is the timestamp when the transaction started.
	TxnStartTime time.Time
	// TxnCommitTime is the timestamp when the transaction was committed.
	TxnCommitTime time.Time

	// tsrMaintainer is used to maintain the TSR (Transaction Status Record)
	// tsrMaintainer is responsible for handling and updating the status of transactions.
	tsrMaintainer TSRMaintainer
	// dataStoreMap is a map of transaction-specific datastores.
	dataStoreMap map[string]Datastore
	// timeSource represents the source of time for the transaction.
	timeSource SourceType
	// oracleURL is the URL of the oracle service used by the transaction.
	oracleURL string
	// locker is used for transaction-level locking.
	locker locker.Locker

	*StateMachine
}

// NewTransaction creates a new Transaction object.
// It initializes the Transaction with default values and returns a pointer to the newly created object.
func NewTransaction() *Transaction {
	return &Transaction{
		TxnId:        "",
		dataStoreMap: make(map[string]Datastore),
		timeSource:   LOCAL,
		locker:       locker.AMemoryLocker,
		StateMachine: NewStateMachine(),
	}
}

// Start begins the transaction.
// It checks if the transaction is already started and returns an error if so.
// It also checks if the necessary datastores are added and returns an error if not.
// It sets the transaction state to STARTED and generates a unique transaction ID.
// It starts each datastore associated with the transaction.
// Returns an error if any of the above steps fail, otherwise returns nil.
func (t *Transaction) Start() error {
	err := t.SetState(config.STARTED)
	if err != nil {
		return err
	}

	// check nessary datastores are added
	if t.tsrMaintainer == nil {
		return errors.New("global datastore not set")
	}
	if len(t.dataStoreMap) == 0 {
		return errors.New("no datastores added")
	}
	t.TxnId = uuid.NewString()
	t.TxnStartTime, err = t.getTime()
	if err != nil {
		return err
	}
	for _, ds := range t.dataStoreMap {
		err := ds.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

// AddDatastore adds a datastore to the transaction.
// It checks if the datastore name is duplicated and returns an error if it is.
// Otherwise, it sets the transaction for the datastore and adds it to the transaction's datastore map.
func (t *Transaction) AddDatastore(ds Datastore) error {
	// if name is duplicated
	if _, ok := t.dataStoreMap[ds.GetName()]; ok {
		return errors.New("duplicated datastore name")
	}
	ds.SetTxn(t)
	t.dataStoreMap[ds.GetName()] = ds
	return nil
}

// SetGlobalDatastore sets the global datastore for the transaction.
// It takes a Datastore parameter and assigns it to the globalDataStore field of the Transaction struct.
func (t *Transaction) SetGlobalDatastore(ds Datastore) {
	t.tsrMaintainer = ds.(TSRMaintainer)
}

// Read reads the value associated with the given key from the specified datastore.
// It returns an error if the transaction is not in the STARTED state or if the datastore is not found.
func (t *Transaction) Read(dsName string, key string, value any) error {
	err := t.CheckState(config.STARTED)
	if err != nil {
		return err
	}

	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Read(key, value)
	}
	return errors.New("datastore not found")
}

// Write writes the given key-value pair to the specified datastore in the transaction.
// It returns an error if the transaction is not in the STARTED state or if the datastore is not found.
func (t *Transaction) Write(dsName string, key string, value any) error {
	err := t.CheckState(config.STARTED)
	if err != nil {
		return err
	}

	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Write(key, value)
	}
	return errors.New("datastore not found")
}

// Delete deletes a key from the specified datastore in the transaction.
// It returns an error if the transaction is not in the STARTED state or if the datastore is not found.
func (t *Transaction) Delete(dsName string, key string) error {
	err := t.CheckState(config.STARTED)
	if err != nil {
		return err
	}

	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Delete(key)
	}
	return errors.New("datastore not found")
}

// Commit commits the transaction.
// It checks the transaction state and performs the prepare phase.
// If the prepare phase fails, it aborts the transaction and returns an error.
// Otherwise, it proceeds to the commit phase and commits the transaction in all data stores.
// Finally, it deletes the transaction state record.
// Returns an error if any operation fails.
func (t *Transaction) Commit() error {

	err := t.SetState(config.COMMITTED)
	if err != nil {
		return err
	}

	// Prepare phase
	t.TxnCommitTime, err = t.getTime()
	if err != nil {
		return err
	}

	success := true
	var cause error
	for _, ds := range t.dataStoreMap {
		err := ds.Prepare()
		if err != nil {
			success = false
			cause = err
			break
		}
	}
	if !success {
		t.Abort()
		return errors.New("prepare phase failed: " + cause.Error())
	}

	// Commit phase
	// The sync point
	txnState, err := t.GetTSRState(t.TxnId)
	if err == nil && txnState == config.ABORTED {
		t.Abort()
		return errors.New("transaction is aborted by other transaction")
	}

	err = t.WriteTSR(t.TxnId, config.COMMITTED)
	if err != nil {
		t.Abort()
		return err
	}

	for _, ds := range t.dataStoreMap {
		// TODO: do not allow abort after Commit
		// try indefinitely until success
		ds.Commit()
	}

	t.DeleteTSR()
	return nil
}

// Abort aborts the transaction.
// It checks the current state of the transaction and returns an error if the transaction is already committed, aborted, or not started.
// If the transaction is in a valid state, it sets the transaction state to ABORTED and calls the Abort method on each data store associated with the transaction.
// Returns an error if any of the data store's Abort method returns an error, otherwise returns nil.
func (t *Transaction) Abort() error {
	lastState := t.GetState()
	err := t.SetState(config.ABORTED)
	if err != nil {
		return err
	}

	hasCommitted := false
	if lastState == config.COMMITTED {
		hasCommitted = true
	}
	for _, ds := range t.dataStoreMap {
		ds.Abort(hasCommitted)
	}
	return nil
}

// WriteTSR writes the Transaction State Record (TSR) for the given transaction ID and state.
// It uses the global data store to persist the TSR.
// The txnId parameter specifies the ID of the transaction.
// The txnState parameter specifies the state of the transaction.
// Returns an error if there was a problem writing the TSR.
func (t *Transaction) WriteTSR(txnId string, txnState config.State) error {
	return t.tsrMaintainer.WriteTSR(txnId, txnState)
}

// DeleteTSR deletes the Transaction Status Record (TSR) associated with the Transaction.
// It calls the DeleteTSR method of the globalDataStore to perform the deletion.
// It returns an error if the deletion operation fails.
func (t *Transaction) DeleteTSR() error {
	return t.tsrMaintainer.DeleteTSR(t.TxnId)
}

func (t *Transaction) GetTSRState(txnId string) (config.State, error) {
	return t.tsrMaintainer.ReadTSR(txnId)
}

// SetGlobalTimeSource sets the global time source for the transaction.
// It takes a URL as a parameter and assigns it to the transaction's oracleURL field.
// The timeSource field is set to GLOBAL.
func (t *Transaction) SetGlobalTimeSource(url string) {
	t.timeSource = GLOBAL
	t.oracleURL = url
}

// getTime returns the current time based on the time source configured in the Transaction.
// If the time source is set to LOCAL, it returns the current local time.
// If the time source is set to GLOBAL, it retrieves the time from the specified time URL.
// It returns the parsed time value and any error encountered during the process.
func (t *Transaction) getTime() (time.Time, error) {
	if t.timeSource == LOCAL {
		return time.Now(), nil
	}
	if t.timeSource == GLOBAL {
		res, err := http.Get(t.oracleURL + "/time")
		if err != nil {
			return time.Now(), errors.New("failed to get time from global time source")
		}

		var timeString string
		fmt.Fscan(res.Body, &timeString)
		return time.Parse(time.RFC3339, timeString)
	}
	return time.Now(), nil
}

// SetLocker sets the locker for the transaction.
// The locker is responsible for managing the concurrency of the transaction.
// It ensures that only one goroutine can access the transaction at a time.
// The locker must implement the locker.Locker interface.
func (t *Transaction) SetLocker(locker locker.Locker) {
	t.locker = locker
}

// Lock locks the specified key with the given ID for the specified duration.
// If the locker is not set, it returns an error.
func (t *Transaction) Lock(key string, id string, duration time.Duration) error {
	if t.locker == nil {
		return errors.New("locker not set")
	}
	return t.locker.Lock(key, id, duration)
}

// Unlock unlocks the specified key with the given ID.
// It returns an error if the locker is not set or if unlocking fails.
func (t *Transaction) Unlock(key string, id string) error {
	if t.locker == nil {
		return errors.New("locker not set")
	}
	return t.locker.Unlock(key, id)
}
