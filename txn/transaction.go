package txn

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kkkzoz/vanilla-icecream/config"
	"github.com/kkkzoz/vanilla-icecream/locker"
)

type SourceType string

const (
	// EMPTY  SourceType = "EMPTY"
	LOCAL  SourceType = "LOCAL"
	GLOBAL SourceType = "GLOBAL"
)

type Transaction struct {
	TxnId           string
	TxnState        config.State
	TxnStartTime    time.Time
	TxnCommitTime   time.Time
	globalDataStore Datastore
	dataStoreMap    map[string]Datastore
	timeSource      SourceType
	oracleURL       string
	locker          locker.Locker
}

// NewTransaction creates a new Transaction object.
// It initializes the Transaction with default values and returns a pointer to the newly created object.
func NewTransaction() *Transaction {
	return &Transaction{
		TxnId:        "",
		TxnState:     config.EMPTY,
		dataStoreMap: make(map[string]Datastore),
		timeSource:   LOCAL,
		locker:       locker.NewMemoryLocker(),
	}
}

// Start begins the transaction.
// It checks if the transaction is already started and returns an error if so.
// It also checks if the necessary datastores are added and returns an error if not.
// It sets the transaction state to STARTED and generates a unique transaction ID.
// It starts each datastore associated with the transaction.
// Returns an error if any of the above steps fail, otherwise returns nil.
func (t *Transaction) Start() error {
	if t.TxnState != config.EMPTY {
		return errors.New("transaction is already started")
	}
	t.TxnState = config.STARTED

	// check nessary datastores are added
	if t.globalDataStore == nil {
		return errors.New("global datastore not set")
	}
	if len(t.dataStoreMap) == 0 {
		return errors.New("no datastores added")
	}
	t.TxnId = uuid.NewString()
	var err error
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
	t.globalDataStore = ds
}

// Read reads the value associated with the given key from the specified datastore.
// It returns an error if the transaction is not in the STARTED state or if the datastore is not found.
func (t *Transaction) Read(dsName string, key string, value any) error {
	if t.TxnState != config.STARTED {
		return errors.New("transaction is not in STARTED state")
	}

	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Read(key, value)
	}
	return errors.New("datastore not found")
}

// Write writes the given key-value pair to the specified datastore in the transaction.
// It returns an error if the transaction is not in the STARTED state or if the datastore is not found.
func (t *Transaction) Write(dsName string, key string, value any) error {
	if t.TxnState != config.STARTED {
		return errors.New("transaction is not in STARTED state")
	}

	if ds, ok := t.dataStoreMap[dsName]; ok {
		return ds.Write(key, value)
	}
	return errors.New("datastore not found")
}

// Delete deletes a key from the specified datastore in the transaction.
// It returns an error if the transaction is not in the STARTED state or if the datastore is not found.
func (t *Transaction) Delete(dsName string, key string) error {
	if t.TxnState != config.STARTED {
		return errors.New("transaction is not in STARTED state")
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
	if t.TxnState == config.COMMITTED {
		return errors.New("transaction is already being committed")
	}
	if t.TxnState == config.ABORTED {
		return errors.New("transaction is already aborted")
	}
	if t.TxnState == config.EMPTY {
		return errors.New("transaction is not started")
	}
	t.TxnState = config.COMMITTED

	// Prepare phase
	var err error
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
		for _, ds := range t.dataStoreMap {
			ds.Abort(true)
		}
		return errors.New("prepare phase failed: " + cause.Error())
	}

	// Commit phase
	// The sync point

	txnState, err := t.GetTSRState()
	if err == nil && txnState == config.ABORTED {
		t.Abort()
		return errors.New("transaction is aborted by other transaction")
	}

	err = t.WriteTSR(config.COMMITTED)
	if err != nil {
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
	if t.TxnState == config.COMMITTED {
		return errors.New("transaction is already committed")
	}
	if t.TxnState == config.ABORTED {
		return errors.New("transaction is already aborted")
	}
	if t.TxnState == config.EMPTY {
		return errors.New("transaction is not started")
	}

	t.TxnState = config.ABORTED
	for _, ds := range t.dataStoreMap {
		ds.Abort(false)
	}
	return nil
}

// WriteTSR writes the transaction state record (TSR) for the given transaction.
// It takes the transaction state as an argument and returns an error if any.
// The TSR is written to the global data store using the transaction's ID.
func (t *Transaction) WriteTSR(txnState config.State) error {
	return t.globalDataStore.WriteTSR(t.TxnId, txnState)
}

// DeleteTSR deletes the Transaction Status Record (TSR) associated with the Transaction.
// It calls the DeleteTSR method of the globalDataStore to perform the deletion.
// It returns an error if the deletion operation fails.
func (t *Transaction) DeleteTSR() error {
	return t.globalDataStore.DeleteTSR(t.TxnId)
}

// GetTSRState returns the current state of the transaction.
// It retrieves the state from the global data store using the transaction ID.
// If successful, it returns the state and nil error.
// If an error occurs during the retrieval, it returns an empty state and the error.
func (t *Transaction) GetTSRState() (config.State, error) {
	var state config.State
	err := t.globalDataStore.Read(t.TxnId, &state)
	return state, err
}

// SetGlobalTimeSource sets the global time source for the transaction.
// It takes the address and port of the time source server as parameters.
// The time source is set to GLOBAL and the time URL is constructed using the provided address and port.
func (t *Transaction) SetGlobalTimeSource(address string, port int) {
	t.timeSource = GLOBAL
	t.oracleURL = fmt.Sprintf("http://%s:%d", address, port)
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

func (t *Transaction) Lock(key string, id string, duration time.Duration) error {
	data := url.Values{}
	data.Set("key", key)
	data.Set("id", id)
	data.Set("duration", strconv.Itoa(int(duration)))

	_, err := http.Get(t.oracleURL + "/lock?" + data.Encode())
	if err != nil {
		return errors.New("failed to lock")
	}
	return nil
}

func (t *Transaction) Unlock(key string, id string) error {
	data := url.Values{}
	data.Set("key", key)
	data.Set("id", id)

	_, err := http.Get(t.oracleURL + "/unlock?" + data.Encode())
	if err != nil {
		return err
	}
	return nil
}
