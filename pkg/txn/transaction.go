package txn

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/oreo-dtx-lab/oreo/internal/testutil"
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/locker"
	. "github.com/oreo-dtx-lab/oreo/pkg/logger"
	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
)

type SourceType string

type TxnError string

func (e TxnError) Error() string {
	return string(e)
}

var (
	KeyNotFound      = errors.Errorf("key not found")
	DirtyRead        = errors.Errorf("dirty read")
	DeserializeError = errors.Errorf("deserialize error")
	VersionMismatch  = errors.Errorf("version mismatch")
	KeyExists        = errors.Errorf("key exists")
	ReadFailed       = errors.Errorf("read failed due to unknown txn status")
)

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
	TxnStartTime int64
	// TxnCommitTime is the timestamp when the transaction was committed.
	TxnCommitTime int64

	// groupKeyMaintainer is used to maintain the Group Key
	// groupKeyMaintainer is responsible for handling and updating the status of transactions.
	groupKeyMaintainer GroupKeyMaintainer
	// dataStoreMap is a map of transaction-specific datastores.
	dataStoreMap map[string]Datastorer
	// timeSource represents the source of time for the transaction.
	timeSource timesource.TimeSourcer
	// oracleURL is the URL of the oracle service used by the transaction.
	// oracleURL string
	// locker is used for transaction-level locking.
	locker locker.Locker

	// isReadOnly indicates whether the transaction is read-only.
	isReadOnly bool

	// writeCount is the number of write operations performed by the transaction.
	writeCount int

	// client is the network client used by the transaction.
	client RemoteClient

	// isRemote indicates whether the transaction is remote.
	isRemote bool

	*StateMachine

	debugStart time.Time
}

// NewTransaction creates a new Transaction object.
// It initializes the Transaction with default values and returns a pointer to the newly created object.
func NewTransaction() *Transaction {
	return &Transaction{
		dataStoreMap: make(map[string]Datastorer),
		timeSource:   timesource.NewSimpleTimeSource(),
		locker:       locker.AMemoryLocker,
		isReadOnly:   true,
		StateMachine: NewStateMachine(),
		isRemote:     false,
	}
}

func NewTransactionWithOracle(oracle timesource.TimeSourcer) *Transaction {
	return &Transaction{
		dataStoreMap: make(map[string]Datastorer),
		timeSource:   oracle,
		locker:       locker.AMemoryLocker,
		isReadOnly:   true,
		StateMachine: NewStateMachine(),
		isRemote:     false,
	}
}

func NewTransactionWithRemote(client RemoteClient, oracle timesource.TimeSourcer) *Transaction {
	return &Transaction{
		dataStoreMap: make(map[string]Datastorer),
		timeSource:   oracle,
		locker:       locker.AMemoryLocker,
		isReadOnly:   true,
		StateMachine: NewStateMachine(),
		client:       client,
		isRemote:     true,
	}
}

// Start begins the transaction.
// It checks if the transaction is already started and returns an error if so.
// It also checks if the necessary datastores are added and returns an error if not.
// It sets the transaction state to STARTED and generates a unique transaction ID.
// It starts each datastore associated with the transaction.
// Returns an error if any of the above steps fail, otherwise returns nil.
func (t *Transaction) Start() error {
	t.debugStart = time.Now()
	defer func() {
		Log.Debugw("txn.Start() ends", "latency", time.Since(t.debugStart), "Topic", "CheckPoint")
	}()

	err := t.SetState(config.STARTED)
	if err != nil {
		return err
	}

	// check nessary datastores are added
	if t.groupKeyMaintainer == nil {
		return errors.New("global datastore not set")
	}
	if len(t.dataStoreMap) == 0 {
		return errors.New("no datastores added")
	}
	t.TxnId = config.Config.IdGenerator.GenerateId()
	Log.Infow("starting transaction", "txnId", t.TxnId)
	t.TxnStartTime, err = t.getTime("start")
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
func (t *Transaction) AddDatastore(ds Datastorer) error {
	// if name is duplicated
	if _, ok := t.dataStoreMap[ds.GetName()]; ok {
		return errors.New("duplicated datastore name")
	}
	ds.SetTxn(t)
	t.dataStoreMap[ds.GetName()] = ds
	return nil
}

// AddDatastores adds multiple datastores to the transaction.
// It takes a variadic parameter `dss` of type `Datastorer` which represents the datastores to be added.
// It returns an error if any datastore fails to be added, otherwise it returns nil.
func (t *Transaction) AddDatastores(dss ...Datastorer) error {
	for _, ds := range dss {
		err := t.AddDatastore(ds)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetGlobalDatastore sets the global datastore for the transaction.
// It takes a Datastore parameter and assigns it to the globalDataStore field of the Transaction struct.
func (t *Transaction) SetGlobalDatastore(ds Datastorer) {
	t.groupKeyMaintainer = ds.(GroupKeyMaintainer)
}

// Read reads the value associated with the given key from the specified datastore.
// It returns an error if the transaction is not in the STARTED state or if the datastore is not found.
func (t *Transaction) Read(dsName string, key string, value any) error {
	err := t.CheckState(config.STARTED)
	if err != nil {
		return err
	}

	t.debug(testutil.DRead, "read in %v: [Key: %v]", dsName, key)
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
	t.isReadOnly = false
	t.writeCount++
	// msgStr := fmt.Sprintf("write in %v: [Key: %v]", dsName, key)
	// Log.Debugw(msgStr, "txnId", t.TxnId, "topic", testutil.DWrite)
	// t.debug(testutil.DWrite, "write in %v: [Key: %v]", dsName, key)
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
	t.isReadOnly = false
	msgStr := fmt.Sprintf("delete in %v: [Key: %v]", dsName, key)
	Log.Debugw(msgStr, "txnId", t.TxnId, "topic", testutil.DDelete)
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

	defer func() {
		Log.Debugw("txn.Commit() ends", "latency", time.Since(t.debugStart), "Topic", "CheckPoint")
	}()

	Log.Infow("Starts to txn.Commit()", "txnId", t.TxnId, "latency", time.Since(t.debugStart), "Topic", "CheckPoint")
	err := t.SetState(config.COMMITTED)
	if err != nil {
		return err
	}

	if t.isReadOnly {
		Log.Infow("transaction is read-only, Commit() complete", "txnId", t.TxnId)
		return nil
	}

	if config.Debug.CherryGarciaMode {
		return t.commitInCherryGarcia()
	} else {
		return t.commitInOreo()
	}
}

func (t *Transaction) commitInCherryGarcia() error {
	success := true
	var cause error
	mu := sync.Mutex{}

	prepareDatastoreFunc := func(ds Datastorer) {
		defer func() {
			msg := fmt.Sprintf("%s prepare phase ends", ds.GetName())
			Log.Debugw(msg, "Latency", time.Since(t.debugStart), "Topic", "CheckPoint")
		}()
		// Cherry Garcia's prepare stage will not return the TCommit
		_, err := ds.Prepare()
		if err != nil {
			mu.Lock()
			success, cause = false, err
			mu.Unlock()
			if stackError, ok := err.(*errors.Error); ok {
				errMsg := fmt.Sprintf("prepare phase failed: %v", stackError.ErrorStack())
				Log.Errorw(errMsg, "txnId", t.TxnId, "ds", ds.GetName())
			}
			Log.Errorw("prepare phase failed", "txnId", t.TxnId, "cause", err, "ds", ds.GetName())
		}
	}

	for _, ds := range t.dataStoreMap {
		prepareDatastoreFunc(ds)
	}

	if !success {
		t.Abort()
		return errors.New("prepare phase failed: " + cause.Error())
	}

	Log.Infow("finishes prepare phase", "txnId", t.TxnId, "latency", time.Since(t.debugStart), "Topic", "CheckPoint")

	_, err := t.CreateGroupKey(t.TxnId, config.COMMITTED, t.TxnCommitTime)
	if err != nil {
		t.Abort()
		return errors.New("transaction is aborted by other transaction")
	}
	Log.Debugw("GroypKey created", "Latency", time.Since(t.debugStart), "Topic", "CheckPoint")

	var wg = sync.WaitGroup{}
	for _, ds := range t.dataStoreMap {
		wg.Add(1)
		go func(ds Datastorer) {
			defer wg.Done()
			ds.Commit()
		}(ds)
	}
	wg.Wait()

	go func() {
		t.DeleteGroupKey()
	}()
	return nil
}

func (t *Transaction) commitInOreo() error {
	tCommit := int64(0)
	success := true
	var cause error
	mu := sync.Mutex{}
	prepareDatastoreFunc := func(ds Datastorer) {
		defer func() {
			msg := fmt.Sprintf("%s prepare phase ends", ds.GetName())
			Log.Debugw(msg, "Latency", time.Since(t.debugStart), "Topic", "CheckPoint")
		}()
		ts, err := ds.Prepare()
		if err != nil {
			mu.Lock()
			success, cause = false, err
			tCommit = max(tCommit, ts)
			mu.Unlock()
			if stackError, ok := err.(*errors.Error); ok {
				errMsg := fmt.Sprintf("prepare phase failed: %v", stackError.ErrorStack())
				Log.Errorw(errMsg, "txnId", t.TxnId, "ds", ds.GetName())
			}
			Log.Errorw("prepare phase failed", "txnId", t.TxnId, "cause", err, "ds", ds.GetName())
		}
	}

	Log.Infow("Starting to make ds.Prepare()", "txnId", t.TxnId, "Latency", time.Since(t.debugStart), "Topic", "CheckPoint")

	var wg = sync.WaitGroup{}
	for _, ds := range t.dataStoreMap {
		wg.Add(1)
		go func(ds Datastorer) {
			defer wg.Done()
			prepareDatastoreFunc(ds)
		}(ds)
	}
	wg.Wait()

	if !success {
		go t.Abort()
		return errors.New("prepare phase failed: " + cause.Error())
	}

	Log.Infow("finishes prepare phase", "txnId", t.TxnId, "latency", time.Since(t.debugStart), "Topic", "CheckPoint")

	var err error
	t.TxnCommitTime = tCommit

	_, err = t.CreateGroupKey(t.TxnId, config.COMMITTED, t.TxnCommitTime)
	if err != nil {
		go t.Abort()
		return errors.New("transaction is aborted by other transaction")
	}
	Log.Debugw("GroupKey created", "Latency", time.Since(t.debugStart), "Topic", "CheckPoint")

	go func() {
		Log.Infow("Starting to make ds.Commit()", "txnId", t.TxnId)
		var wg = sync.WaitGroup{}
		for _, ds := range t.dataStoreMap {
			wg.Add(1)
			go func(ds Datastorer) {
				defer wg.Done()
				ds.Commit()
			}(ds)
		}
		wg.Wait()
		t.DeleteGroupKey()
	}()
	return nil

}

func (t *Transaction) OnePhaseCommit() error {
	for _, ds := range t.dataStoreMap {
		err := ds.OnePhaseCommit()
		if err != nil {
			Log.Errorw("one phase commit failed", "txnId", t.TxnId, "ds", ds.GetName(), "cause", err)
			go t.Abort()
			return err
		}
	}
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
	Log.Infow("aborting transaction", "txnId", t.TxnId, "hasCommitted", hasCommitted)
	t.CreateGroupKey(t.TxnId, config.ABORTED, 0)
	for _, ds := range t.dataStoreMap {
		err := ds.Abort(hasCommitted)
		if err != nil {
			Log.Errorw("abort failed", "txnId", t.TxnId, "cause", err, "ds", ds.GetName())
		}
	}
	return nil
}

func (t *Transaction) WriteGroupKey(txnId string, txnState config.State, tCommit int64) error {
	return t.groupKeyMaintainer.WriteGroupKey(txnId, txnState, tCommit)
}

// SetTSR writes the Transaction State Record (TSR) for the given transaction ID and state.
// It uses the global data store to persist the TSR.
// The txnId parameter specifies the ID of the transaction.
// The txnState parameter specifies the state of the transaction.
// Returns an error if there was a problem writing the TSR.
func (t *Transaction) CreateGroupKey(txnId string, txnState config.State, tCommit int64) (config.State, error) {
	return t.groupKeyMaintainer.CreateGroupKey(txnId, txnState, tCommit)
}

// DeleteTSR deletes the Transaction Status Record (TSR) associated with the Transaction.
// It calls the DeleteTSR method of the tsrMaintainer to perform the deletion.
// It returns an error if the deletion operation fails.
func (t *Transaction) DeleteGroupKey() error {
	return t.groupKeyMaintainer.DeleteGroupKey(t.TxnId)
}

func (t *Transaction) GetGroupKey(txnId string) (GroupKey, error) {
	return t.groupKeyMaintainer.ReadGroupKey(txnId)
}

// getTime returns the current time based on the time source configured in the Transaction.
func (t *Transaction) getTime(mode string) (int64, error) {
	if config.Debug.DebugMode {
		// simulate the latency of the HTTP request
		// used in benchmark
		time.Sleep(config.Debug.TimeOracleAdditionalLatency)
	}
	return t.timeSource.GetTime(mode)
}

func (t *Transaction) RemoteRead(dsName string, key string) (DataItem, RemoteDataStrategy, error) {
	if !t.isRemote {
		return nil, Normal, errors.New("not a remote transaction")
	}

	globalName := t.groupKeyMaintainer.(Datastorer).GetName()

	return t.client.Read(dsName, key, t.TxnStartTime, RecordConfig{
		GlobalName:                  globalName,
		MaxRecordLen:                config.Config.MaxRecordLength,
		ReadStrategy:                config.Config.ReadStrategy,
		ConcurrentOptimizationLevel: config.Config.ConcurrentOptimizationLevel,
	})
}

func (t *Transaction) RemoteValidate(dsName string, key string, item DataItem) error {
	panic("not implemented")
}

func (t *Transaction) RemotePrepare(dsName string, itemList []DataItem, validationMap map[string]PredicateInfo) (map[string]string, int64, error) {
	if !t.isRemote {
		return nil, 0, errors.New("not a remote transaction")
	}
	globalName := t.groupKeyMaintainer.(Datastorer).GetName()

	cfg := RecordConfig{
		GlobalName:                  globalName,
		MaxRecordLen:                config.Config.MaxRecordLength,
		ReadStrategy:                config.Config.ReadStrategy,
		ConcurrentOptimizationLevel: config.Config.ConcurrentOptimizationLevel,
	}
	return t.client.Prepare(dsName, itemList, t.TxnStartTime,
		cfg, validationMap)
}

func (t *Transaction) RemoteCommit(dsName string, infoList []CommitInfo) error {
	if !t.isRemote {
		return errors.New("not a remote transaction")
	}
	return t.client.Commit(dsName, infoList, t.TxnCommitTime)
}

func (t *Transaction) RemoteAbort(dsName string, keyList []string) error {
	if !t.isRemote {
		return errors.New("not a remote transaction")
	}
	return t.client.Abort(dsName, keyList, t.TxnId)
}

func (t *Transaction) debug(topic testutil.TxnTopic, format string, a ...interface{}) {
	prefix := fmt.Sprintf("%v ", t.TxnId)
	testutil.Debug(topic, prefix+format, a...)
}
