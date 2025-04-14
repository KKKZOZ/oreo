package network

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdlog "log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/logger" // Use provided logger for client ops
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/valyala/fasthttp"
	// Using json-iterator for potential performance benefits in client methods
)

// Ensure RegistryClient implements the RemoteClient interface
var _ txn.RemoteClient = (*Client)(nil)

// --- Registry Data Structures ---

// InstanceInfo holds info about a registered executor instance within the registry.
type InstanceInfo struct {
	Address       string    // The advertised network address (e.g., "1.2.3.4:8000")
	LastHeartbeat time.Time // Timestamp of the last successful heartbeat
	DsNames       []string  // List of datastore names this instance handles (e.g., ["Redis", "MongoDB1"])
}

// RegistryRequest is the payload structure used for communication
// between executors and the registry server (/register, /deregister, /heartbeat).
type RegistryRequest struct {
	Address string   `json:"address"`           // Address of the executor instance
	DsNames []string `json:"dsNames,omitempty"` // Datastore names (primarily for /register)
}

// --- Combined Client and Registry ---

// Client embeds both the registry server functionality and the client
// logic for discovering and communicating with executor instances.
type Client struct {
	// Registry Server Part
	registryMutex    sync.RWMutex            // Protects access to instances and dsNameIndex
	instances        map[string]InstanceInfo // Key: Instance Address (e.g., "1.2.3.4:8000"), Value: InstanceInfo
	dsNameIndex      map[string][]string     // Key: DsName (e.g., "Redis"), Value: List of active instance addresses handling it
	instanceTTL      time.Duration           // Time after last heartbeat before an instance is considered stale
	registryListener *http.Server            // The HTTP server instance for the registry endpoints
	shutdownCtx      context.Context         // Context to signal shutdown for background tasks
	shutdownCancel   context.CancelFunc      // Function to trigger the shutdown context
	wg               sync.WaitGroup          // Waits for background goroutines (listener, cleanup) to finish

	// Client Consumer Part
	clientMutex sync.Mutex                // Protects access to curIndexMap initialization
	curIndexMap map[string]*atomic.Uint64 // Key: DsName, Value: Atomic counter for round-robin index
	httpClient  *fasthttp.Client          // Reusable fasthttp client for executor communication
}

// Constants
const (
	ALL = "ALL" // Special DsName indicating an instance handles all datastores
	// Default TTL if not provided (should be longer than executor's heartbeatInterval)
	defaultInstanceTTL = 6 * time.Second
	// How often the registry checks for stale instances
	cleanupInterval = 3 * time.Second
	// Default timeout for requests to executors if not configured otherwise
	defaultRequestTimeout = 30000 * time.Millisecond
)

// NewRegistryClient initializes the combined client, starts the registry HTTP server,
// and launches the background cleanup routine.
// registryListenAddr: The network address for the registry server to listen on (e.g., ":9000").
// instanceTTL: Optional duration after which an instance is removed if no heartbeat is received. Defaults to defaultInstanceTTL.
func NewClient(registryListenAddr string, instanceTTL ...time.Duration) (*Client, error) {
	ttl := defaultInstanceTTL
	if len(instanceTTL) > 0 && instanceTTL[0] > 0 {
		ttl = instanceTTL[0]
	}
	// Warn if TTL is potentially too short compared to cleanup frequency
	if ttl < cleanupInterval*2 {
		stdlog.Printf("WARN: RegistryClient configured instance TTL (%v) is potentially too short compared to cleanup interval (%v)", ttl, cleanupInterval)
	}

	// Create cancellable context for managing background goroutine lifecycle
	ctx, cancel := context.WithCancel(context.Background())

	rc := &Client{
		instances:      make(map[string]InstanceInfo),
		dsNameIndex:    make(map[string][]string),
		instanceTTL:    ttl,
		curIndexMap:    make(map[string]*atomic.Uint64),
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
		httpClient:     &fasthttp.Client{ // Initialize the fasthttp client
			// Set reasonable default timeouts for reading/writing - DoTimeout overrides these per call
			// ReadTimeout:  defaultRequestTimeout,
			// WriteTimeout: defaultRequestTimeout,
			// MaxIdleConnDuration: 1 * time.Minute, // Keep idle connections open
			// MaxConnsPerHost:     100,             // Limit connections per executor host
			// Consider other options like Name for identification in logs if needed
		},
	}

	// Setup Registry HTTP Server
	mux := http.NewServeMux()
	mux.HandleFunc("/register", rc.handleRegister)
	mux.HandleFunc("/deregister", rc.handleDeregister)
	mux.HandleFunc("/heartbeat", rc.handleHeartbeat)
	mux.HandleFunc("/services", rc.handleGetServices) // Optional: endpoint for viewing registered services

	rc.registryListener = &http.Server{
		Addr:    registryListenAddr,
		Handler: mux,
		// Consider adding ReadTimeout, WriteTimeout, IdleTimeout for production robustness
		// ReadTimeout:  10 * time.Second,
		// WriteTimeout: 10 * time.Second,
		// IdleTimeout:  60 * time.Second,
	}

	// Start the registry HTTP listener in a goroutine
	rc.wg.Add(1) // Increment counter for the listener goroutine
	go func() {
		defer rc.wg.Done() // Decrement counter when goroutine finishes
		stdlog.Printf("Registry server starting on %s (TTL: %v)", registryListenAddr, rc.instanceTTL)
		// ListenAndServe blocks until the server is shut down
		if err := rc.registryListener.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// Use standard log for fatal errors during startup
			stdlog.Fatalf("FATAL: Registry server failed to listen on %s: %v", registryListenAddr, err)
		}
		// This log message is reached when ListenAndServe returns after Shutdown() is called
		stdlog.Printf("Registry server on %s stopped listening.", registryListenAddr)
	}()

	// Start the background cleanup goroutine
	rc.wg.Add(1) // Increment counter for the cleanup goroutine
	go rc.cleanupStaleInstances()

	stdlog.Printf("RegistryClient initialized successfully.")
	return rc, nil
}

// Shutdown gracefully stops the registry HTTP server and the cleanup routine.
// It waits for background tasks to complete or until the provided context times out.
func (rc *Client) Shutdown(ctx context.Context) error {
	stdlog.Println("RegistryClient shutting down...")

	// 1. Signal background goroutines to stop
	rc.shutdownCancel()

	// 2. Shutdown the HTTP registry listener
	shutdownStart := time.Now()
	var listenerErr error
	if rc.registryListener != nil {
		stdlog.Println("Shutting down registry HTTP listener...")
		// Use the provided context for listener shutdown timeout
		if err := rc.registryListener.Shutdown(ctx); err != nil {
			stdlog.Printf("Error shutting down registry listener: %v", err)
			listenerErr = err // Store error but continue
		}
	}

	// 3. Wait for background goroutines (listener + cleanup) to finish
	stdlog.Println("Waiting for background routines to stop...")
	waitChan := make(chan struct{})
	go func() {
		rc.wg.Wait() // This blocks until wg counter is zero
		close(waitChan)
	}()

	select {
	case <-waitChan:
		stdlog.Printf("All background routines stopped gracefully (waited %v).", time.Since(shutdownStart))
	case <-ctx.Done():
		stdlog.Printf("Shutdown context timed out after %v waiting for background routines.", time.Since(shutdownStart))
		return fmt.Errorf("registry client shutdown timed out: %w", ctx.Err())
	}

	stdlog.Println("RegistryClient shutdown complete.")
	// Return listener error if it occurred
	return listenerErr
}

// --- Registry Server Handlers (Unchanged) ---

// handleRegister handles POST requests from executors wishing to join the registry.
func (rc *Client) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RegistryRequest
	// Use standard JSON decoder for HTTP handler
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		stdlog.Printf("Error decoding register request: %v", err)
		http.Error(w, fmt.Sprintf("Bad Request: %v", err), http.StatusBadRequest)
		return
	}
	if req.Address == "" {
		http.Error(w, "Bad Request: Instance 'address' is required", http.StatusBadRequest)
		return
	}
	if len(req.DsNames) == 0 {
		stdlog.Printf("Warning: Registering instance %s without specific DsNames. Assuming it handles 'ALL'.", req.Address)
		req.DsNames = []string{ALL} // Default to ALL if none provided
	}

	rc.registryMutex.Lock()
	defer rc.registryMutex.Unlock()

	now := time.Now()
	existing, exists := rc.instances[req.Address]

	if exists {
		stdlog.Printf("Re-registering/updating instance: %s (Old DsNames: %v, New DsNames: %v)", req.Address, existing.DsNames, req.DsNames)
		rc.removeInstanceFromDsNameIndexLocked(req.Address, existing.DsNames)
		existing.LastHeartbeat = now
		existing.DsNames = req.DsNames
		rc.instances[req.Address] = existing
	} else {
		stdlog.Printf("Registering new instance: %s for DsNames: %v", req.Address, req.DsNames)
		rc.instances[req.Address] = InstanceInfo{
			Address:       req.Address,
			LastHeartbeat: now,
			DsNames:       req.DsNames,
		}
	}

	rc.addInstanceToDsNameIndexLocked(req.Address, req.DsNames)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Registered successfully")
}

// handleDeregister handles POST requests from executors gracefully leaving the pool.
func (rc *Client) handleDeregister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RegistryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		stdlog.Printf("Error decoding deregister request: %v", err)
		http.Error(w, fmt.Sprintf("Bad Request: %v", err), http.StatusBadRequest)
		return
	}
	if req.Address == "" {
		http.Error(w, "Bad Request: Instance 'address' is required", http.StatusBadRequest)
		return
	}

	rc.registryMutex.Lock()
	defer rc.registryMutex.Unlock()

	instance, ok := rc.instances[req.Address]
	if ok {
		stdlog.Printf("Deregistering instance: %s (handled DsNames: %v)", req.Address, instance.DsNames)
		delete(rc.instances, req.Address)
		rc.removeInstanceFromDsNameIndexLocked(req.Address, instance.DsNames)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Deregistered successfully")
	} else {
		stdlog.Printf("Deregister attempt for unknown or already removed instance: %s", req.Address)
		http.Error(w, "Instance not found", http.StatusNotFound)
	}
}

// handleHeartbeat handles periodic POST requests from executors to signal they are alive.
func (rc *Client) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RegistryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		stdlog.Printf("Error decoding heartbeat request: %v", err)
		http.Error(w, fmt.Sprintf("Bad Request: %v", err), http.StatusBadRequest)
		return
	}
	if req.Address == "" {
		http.Error(w, "Bad Request: Instance 'address' is required", http.StatusBadRequest)
		return
	}

	rc.registryMutex.Lock()
	defer rc.registryMutex.Unlock()

	instance, ok := rc.instances[req.Address]
	if !ok {
		stdlog.Printf("Heartbeat received from unknown instance: %s. Instance should re-register.", req.Address)
		http.Error(w, "Instance not registered; please re-register", http.StatusNotFound)
		return
	}

	instance.LastHeartbeat = time.Now()
	rc.instances[req.Address] = instance

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Heartbeat received")
}

// handleGetServices provides an optional diagnostic endpoint (GET /services)
func (rc *Client) handleGetServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	rc.registryMutex.RLock() // Use read lock for viewing data
	defer rc.registryMutex.RUnlock()

	output := make(map[string][]string)
	now := time.Now()
	activeCount := 0
	staleCount := 0

	for addr, info := range rc.instances {
		if now.Sub(info.LastHeartbeat) <= rc.instanceTTL {
			activeCount++
			for _, dsName := range info.DsNames {
				output[dsName] = append(output[dsName], addr)
			}
		} else {
			staleCount++
		}
	}

	output["_summary_"] = []string{
		fmt.Sprintf("active_instances: %d", activeCount),
		fmt.Sprintf("stale_instances (pending cleanup): %d", staleCount),
		fmt.Sprintf("total_instances (in map): %d", len(rc.instances)),
		fmt.Sprintf("instance_ttl: %v", rc.instanceTTL),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(output); err != nil {
		stdlog.Printf("Error encoding service list response: %v", err)
	}
}

// --- Registry Helper Functions (Unchanged) ---

// addInstanceToDsNameIndexLocked updates the dsName -> [address] lookup map.
// It MUST be called while holding the registryMutex write lock.
func (rc *Client) addInstanceToDsNameIndexLocked(instanceAddr string, dsNames []string) {
	for _, dsName := range dsNames {
		list, _ := rc.dsNameIndex[dsName]
		found := false
		for _, addr := range list {
			if addr == instanceAddr {
				found = true
				break
			}
		}
		if !found {
			rc.dsNameIndex[dsName] = append(list, instanceAddr)
		}
	}
}

// removeInstanceFromDsNameIndexLocked removes an instance's address from the
// dsName -> [address] lookup map for all DsNames it handled.
// It MUST be called while holding the registryMutex write lock.
func (rc *Client) removeInstanceFromDsNameIndexLocked(instanceAddr string, dsNames []string) {
	for _, dsName := range dsNames {
		list, ok := rc.dsNameIndex[dsName]
		if !ok {
			continue
		}
		newList := make([]string, 0, len(list))
		for _, addr := range list {
			if addr != instanceAddr {
				newList = append(newList, addr)
			}
		}
		if len(newList) == 0 {
			delete(rc.dsNameIndex, dsName)
		} else {
			rc.dsNameIndex[dsName] = newList
		}
	}
}

// cleanupStaleInstances runs as a background goroutine, periodically checking for
// instances that have exceeded their TTL and removing them from the registry.
func (rc *Client) cleanupStaleInstances() {
	defer rc.wg.Done()

	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	stdlog.Printf("Starting stale instance cleanup routine (TTL: %v, Interval: %v)", rc.instanceTTL, cleanupInterval)

	for {
		select {
		case <-ticker.C:
			rc.registryMutex.Lock()

			now := time.Now()
			cleanedCount := 0
			staleInstances := make(map[string][]string)

			for addr, info := range rc.instances {
				if now.Sub(info.LastHeartbeat) > rc.instanceTTL {
					stdlog.Printf("Cleanup: Stale instance found: %s (last heartbeat: %v ago)", addr, now.Sub(info.LastHeartbeat).Round(time.Second))
					staleInstances[addr] = info.DsNames
					delete(rc.instances, addr)
					cleanedCount++
				}
			}

			for addr, dsNames := range staleInstances {
				rc.removeInstanceFromDsNameIndexLocked(addr, dsNames)
			}

			rc.registryMutex.Unlock()

			if cleanedCount > 0 {
				stdlog.Printf("Cleanup: Removed %d stale instance(s)", cleanedCount)
			}

		case <-rc.shutdownCtx.Done():
			stdlog.Println("Stale instance cleanup routine stopping due to shutdown signal.")
			return
		}
	}
}

// --- Client/Consumer Methods (Modified) ---

// GetServerAddr selects an active executor address for the given dsName using round-robin.
// (Unchanged - Round Robin Logic is separate from HTTP execution)
func (rc *Client) GetServerAddr(dsName string) (string, error) {
	rc.registryMutex.RLock() // Use read lock to access the dsNameIndex

	addrList, ok := rc.dsNameIndex[dsName]
	if !ok || len(addrList) == 0 {
		addrList, ok = rc.dsNameIndex[ALL]
		if !ok || len(addrList) == 0 {
			rc.registryMutex.RUnlock()
			logger.Log.Errorw("No active executor instance found for dsName", "dsName", dsName)
			return "", fmt.Errorf("no active executor instance available for dsName %s (or ALL)", dsName)
		}
		dsName = ALL // Use ALL key for counter
	}

	numInstances := len(addrList)
	addresses := make([]string, numInstances)
	copy(addresses, addrList)
	rc.registryMutex.RUnlock()

	rc.clientMutex.Lock()
	counter, exists := rc.curIndexMap[dsName]
	if !exists {
		var newAtomicCounter atomic.Uint64
		rc.curIndexMap[dsName] = &newAtomicCounter
		counter = &newAtomicCounter
	}
	rc.clientMutex.Unlock()

	idx := counter.Add(1) - 1
	instanceIndex := int(idx % uint64(numInstances))

	selectedAddr := addresses[instanceIndex]

	return selectedAddr, nil
}

// Helper to get timeout value (can be extended to read from config)
func getRequestTimeout() time.Duration {
	// TODO: Read this value from config.System.NetworkRequestTimeout if available
	// For now, use the default.
	// Example:
	// if config.System.NetworkRequestTimeout > 0 {
	//     return config.System.NetworkRequestTimeout
	// }
	return defaultRequestTimeout
}

// Read sends a read request with timeout.
func (rc *Client) Read(dsName string, key string, ts int64, cfg txn.RecordConfig) (txn.DataItem, txn.RemoteDataStrategy, string, error) {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	addr, err := rc.GetServerAddr(dsName)
	if err != nil {
		return nil, txn.Normal, "", fmt.Errorf("failed to get executor address for read dsName '%s': %w", dsName, err)
	}
	reqUrl := "http://" + addr + "/read"
	logger.Log.Debugw("Executing Read request", "url", reqUrl, "dsName", dsName, "key", key)

	reqData := ReadRequest{DsName: dsName, Key: key, StartTime: ts, Config: cfg}
	jsonData, err := json2.Marshal(reqData)
	if err != nil {
		logger.Log.Errorw("Failed to marshal Read request body", "error", err)
		return nil, txn.Normal, "", fmt.Errorf("failed to marshal read request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	// Execute with timeout
	timeout := getRequestTimeout()
	err = rc.httpClient.DoTimeout(req, resp, timeout)
	if err != nil {
		// Check specifically for timeout error
		if errors.Is(err, fasthttp.ErrTimeout) {
			logger.Log.Errorw("Read HTTP request timed out", "url", reqUrl, "timeout", timeout, "error", err)
			return nil, txn.Normal, "", fmt.Errorf("request to executor %s timed out after %v: %w", reqUrl, timeout, err)
		}
		// Handle other potential errors (connection refused, DNS error, etc.)
		logger.Log.Errorw("Failed to execute Read HTTP request", "url", reqUrl, "error", err)
		return nil, txn.Normal, "", fmt.Errorf("http request to executor %s failed: %w", reqUrl, err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		errMsg := fmt.Sprintf("executor %s returned status %d for read", addr, resp.StatusCode())
		logger.Log.Warnw(errMsg, "url", reqUrl, "responseBody", string(resp.Body()))
		// return nil, txn.Normal, "", errors.New(errMsg)
	}

	var response ReadResponse
	err = json2.Unmarshal(resp.Body(), &response)
	if err != nil {
		logger.Log.Errorw("Failed to unmarshal Read response body", "url", reqUrl, "body", string(resp.Body()), "error", err)
		return nil, txn.Normal, "", fmt.Errorf("unmarshal read response error: %w", err)
	}

	if response.Status == "OK" {
		return response.Data, response.DataStrategy, response.GroupKey, nil
	} else {
		errMsg := response.ErrMsg
		logger.Log.Warnw("Read operation failed on executor (application error)", "url", reqUrl, "error", errMsg)
		return nil, txn.Normal, "", errors.New(errMsg)
	}
}

// Prepare sends a prepare request with timeout.
func (rc *Client) Prepare(dsName string, itemList []txn.DataItem,
	startTime int64, cfg txn.RecordConfig,
	validationMap map[string]txn.PredicateInfo) (map[string]string, int64, error) {

	debugStart := time.Now()
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	addr, err := rc.GetServerAddr(dsName)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get executor address for prepare dsName '%s': %w", dsName, err)
	}
	reqUrl := "http://" + addr + "/prepare"
	logger.Log.Debugw("Executing Prepare request", "url", reqUrl, "dsName", dsName, "itemCount", len(itemList))

	reqData := PrepareRequest{DsName: dsName, ItemType: GetItemType(dsName), ItemList: itemList, StartTime: startTime, Config: cfg, ValidationMap: validationMap}
	jsonData, err := json2.Marshal(reqData)
	if err != nil {
		logger.Log.Errorw("Failed to marshal Prepare request body", "error", err)
		return nil, 0, fmt.Errorf("failed to marshal prepare request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	// Execute with timeout
	timeout := getRequestTimeout()
	debugMsg := fmt.Sprintf("HttpClient.DoTimeout(Prepare) to %s", reqUrl)
	logger.Log.Debugw("Before "+debugMsg, "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint", "Timeout", timeout)
	err = rc.httpClient.DoTimeout(req, resp, timeout)
	logger.Log.Debugw("After "+debugMsg, "LatencyInFunc", time.Since(debugStart), "Topic", "CheckPoint")
	if err != nil {
		if errors.Is(err, fasthttp.ErrTimeout) {
			logger.Log.Errorw("Prepare HTTP request timed out", "url", reqUrl, "timeout", timeout, "error", err)
			return nil, 0, fmt.Errorf("request to executor %s timed out after %v: %w", reqUrl, timeout, err)
		}
		logger.Log.Errorw("Failed to execute Prepare HTTP request", "url", reqUrl, "error", err)
		return nil, 0, fmt.Errorf("http request to executor %s failed: %w", reqUrl, err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		errMsg := fmt.Sprintf("executor %s returned status %d for prepare", addr, resp.StatusCode())
		logger.Log.Warnw(errMsg, "url", reqUrl, "responseBody", string(resp.Body()))
		return nil, 0, errors.New(errMsg)
	}

	var response PrepareResponse
	err = json2.Unmarshal(resp.Body(), &response)
	if err != nil {
		logger.Log.Errorw("Failed to unmarshal Prepare response body", "url", reqUrl, "body", string(resp.Body()), "error", err)
		return nil, 0, fmt.Errorf("unmarshal prepare response error: %w", err)
	}

	if response.Status == "OK" {
		return response.VerMap, response.TCommit, nil
	} else {
		errMsg := response.ErrMsg
		logger.Log.Warnw("Prepare operation failed on executor (application error)", "url", reqUrl, "error", errMsg)
		return nil, 0, errors.New(errMsg)
	}
}

// Commit sends a commit request with timeout.
func (rc *Client) Commit(dsName string, infoList []txn.CommitInfo, tCommit int64) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	addr, err := rc.GetServerAddr(dsName)
	if err != nil {
		return fmt.Errorf("failed to get executor address for commit dsName '%s': %w", dsName, err)
	}
	reqUrl := "http://" + addr + "/commit"
	logger.Log.Debugw("Executing Commit request", "url", reqUrl, "dsName", dsName, "infoCount", len(infoList))

	reqData := CommitRequest{DsName: dsName, List: infoList, TCommit: tCommit}
	jsonData, err := json2.Marshal(reqData)
	if err != nil {
		logger.Log.Errorw("Failed to marshal Commit request body", "error", err)
		return fmt.Errorf("failed to marshal commit request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	// Execute with timeout
	timeout := getRequestTimeout()
	err = rc.httpClient.DoTimeout(req, resp, timeout)
	if err != nil {
		if errors.Is(err, fasthttp.ErrTimeout) {
			logger.Log.Errorw("Commit HTTP request timed out", "url", reqUrl, "timeout", timeout, "error", err)
			return fmt.Errorf("request to executor %s timed out after %v: %w", reqUrl, timeout, err)
		}
		logger.Log.Errorw("Failed to execute Commit HTTP request", "url", reqUrl, "error", err)
		return fmt.Errorf("http request to executor %s failed: %w", reqUrl, err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		errMsg := fmt.Sprintf("executor %s returned status %d for commit", addr, resp.StatusCode())
		logger.Log.Warnw(errMsg, "url", reqUrl, "responseBody", string(resp.Body()))
		return errors.New(errMsg)
	}

	var response Response[string] // Generic response structure
	err = json2.Unmarshal(resp.Body(), &response)
	if err != nil {
		logger.Log.Errorw("Failed to unmarshal Commit response body", "url", reqUrl, "body", string(resp.Body()), "error", err)
		return fmt.Errorf("unmarshal commit response error: %w", err)
	}

	if response.Status == "OK" {
		return nil
	} else {
		errMsg := response.ErrMsg
		logger.Log.Warnw("Commit operation failed on executor (application error)", "url", reqUrl, "error", errMsg)
		return errors.New(errMsg)
	}
}

// Abort sends an abort request with timeout.
func (rc *Client) Abort(dsName string, keyList []string, groupKeyList string) error {
	if config.Debug.DebugMode {
		time.Sleep(config.Debug.HTTPAdditionalLatency)
	}

	addr, err := rc.GetServerAddr(dsName)
	if err != nil {
		return fmt.Errorf("failed to get executor address for abort dsName '%s': %w", dsName, err)
	}
	reqUrl := "http://" + addr + "/abort"
	logger.Log.Debugw("Executing Abort request", "url", reqUrl, "dsName", dsName, "keyCount", len(keyList))

	reqData := AbortRequest{DsName: dsName, KeyList: keyList, GroupKeyList: groupKeyList}
	jsonData, err := json2.Marshal(reqData)
	if err != nil {
		logger.Log.Errorw("Failed to marshal Abort request body", "error", err)
		return fmt.Errorf("failed to marshal abort request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqUrl)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.SetBody(jsonData)

	// Execute with timeout
	timeout := getRequestTimeout()
	err = rc.httpClient.DoTimeout(req, resp, timeout)
	if err != nil {
		if errors.Is(err, fasthttp.ErrTimeout) {
			logger.Log.Errorw("Abort HTTP request timed out", "url", reqUrl, "timeout", timeout, "error", err)
			return fmt.Errorf("request to executor %s timed out after %v: %w", reqUrl, timeout, err)
		}
		logger.Log.Errorw("Failed to execute Abort HTTP request", "url", reqUrl, "error", err)
		return fmt.Errorf("http request to executor %s failed: %w", reqUrl, err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		errMsg := fmt.Sprintf("executor %s returned status %d for abort", addr, resp.StatusCode())
		logger.Log.Warnw(errMsg, "url", reqUrl, "responseBody", string(resp.Body()))
		return errors.New(errMsg)
	}

	var response Response[string] // Generic response structure
	err = json2.Unmarshal(resp.Body(), &response)
	if err != nil {
		logger.Log.Errorw("Failed to unmarshal Abort response body", "url", reqUrl, "body", string(resp.Body()), "error", err)
		return fmt.Errorf("unmarshal abort response error: %w", err)
	}

	if response.Status == "OK" {
		return nil
	} else {
		errMsg := response.ErrMsg
		logger.Log.Warnw("Abort operation failed on executor (application error)", "url", reqUrl, "error", errMsg)
		return errors.New(errMsg)
	}
}

// GetExecutorAddrList retrieves a list of currently known executor addresses.
// (Unchanged)
func (rc *Client) GetExecutorAddrList() []string {
	rc.registryMutex.RLock()
	defer rc.registryMutex.RUnlock()

	addressSet := make(map[string]struct{})
	for _, instance := range rc.instances {
		// Consider filtering only active instances based on TTL if needed
		// now := time.Now()
		// if now.Sub(instance.LastHeartbeat) <= rc.instanceTTL {
		addressSet[instance.Address] = struct{}{}
		// }
	}

	addressList := make([]string, 0, len(addressSet))
	for addr := range addressSet {
		addressList = append(addressList, addr)
	}

	return addressList
}
