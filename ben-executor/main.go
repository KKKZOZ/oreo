package main

import (
	"benchmark/pkg/benconfig"
	"bytes"         // Import bytes
	"context"       // Import context
	"encoding/json" // Standard JSON for registry communication
	"flag"
	"fmt"
	"io" // Import io
	"log"
	"net/http" // Import net/http for registry communication
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync" // Import sync
	"syscall"
	"time"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	jsoniter "github.com/json-iterator/go" // Keep for application logic if needed
	"github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/cassandra"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/couchdb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/dynamodb"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/mongo"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/redis"
	"github.com/oreo-dtx-lab/oreo/pkg/datastore/tikv"
	"github.com/oreo-dtx-lab/oreo/pkg/network" // Assuming client.go is in this package
	"github.com/oreo-dtx-lab/oreo/pkg/serializer"
	"github.com/oreo-dtx-lab/oreo/pkg/timesource"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Use standard json for registry, jsoniter for app data if performance matters there
var json2 = jsoniter.ConfigCompatibleWithStandardLibrary

var Banner = `
 ____  _        _       _               
/ ___|| |_ __ _| |_ ___| | ___  ___ ___ 
\___ \| __/ _| | __/ _ \ |/ _ \/ __/ __|
 ___) | || (_| | ||  __/ |  __/\__ \__ \
|____/ \__\__,_|\__\___|_|\___||___/___/
`

type Server struct {
	port            int
	advertiseAddr   string // Address to advertise to the registry
	registryAddr    string // Address of the registry service
	handledDsNames  []string
	reader          network.Reader
	committer       network.Committer
	heartbeatCtx    context.Context
	heartbeatCancel context.CancelFunc
	wg              sync.WaitGroup   // To wait for heartbeat goroutine
	fasthttpServer  *fasthttp.Server // Keep track for shutdown
}

// Registry communication payloads
type RegistryRequest struct {
	Address string   `json:"address"`
	DsNames []string `json:"dsNames,omitempty"` // Used for registration
}

// NewServer modified to accept registry info and dsNames
func NewServer(port int, advertiseAddr, registryAddr string, handledDsNames []string, connMap map[string]txn.Connector, factory txn.DataItemFactory, timeSource timesource.TimeSourcer) *Server {
	reader := *network.NewReader(connMap, factory, serializer.NewJSON2Serializer(), network.NewCacher())
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		port:            port,
		advertiseAddr:   advertiseAddr,
		registryAddr:    registryAddr,
		handledDsNames:  handledDsNames,
		reader:          reader,
		committer:       *network.NewCommitter(connMap, reader, serializer.NewJSON2Serializer(), factory, timeSource),
		heartbeatCtx:    ctx,
		heartbeatCancel: cancel,
	}
}

// --- Registry Interaction ---

const (
	registryTimeout   = 5 * time.Second
	heartbeatInterval = 10 * time.Second // Should be less than registry TTL
)

func (s *Server) registerWithRegistry() error {
	if s.registryAddr == "" {
		Log.Warnw("Registry address not set, skipping registration")
		return nil // Or return an error if registration is mandatory
	}
	if s.advertiseAddr == "" {
		return fmt.Errorf("advertise address not set, cannot register")
	}

	Log.Infow("Attempting to register with registry", "registry", s.registryAddr, "advertise", s.advertiseAddr, "dsNames", s.handledDsNames)

	reqBody := RegistryRequest{
		Address: s.advertiseAddr,
		DsNames: s.handledDsNames,
	}
	jsonData, err := json.Marshal(reqBody) // Use standard JSON for registry comms
	if err != nil {
		return fmt.Errorf("failed to marshal register request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), registryTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.registryAddr+"/register", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create register request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send register request to %s: %w", s.registryAddr, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Read body for detailed error
		return fmt.Errorf("registry at %s returned non-OK status for register: %s Body: %s", s.registryAddr, resp.Status, string(bodyBytes))
	}
	Log.Infow("Successfully registered with registry", "registry", s.registryAddr)
	return nil
}

func (s *Server) deregisterFromRegistry() error {
	if s.registryAddr == "" || s.advertiseAddr == "" {
		Log.Warnw("Registry or advertise address not set, skipping deregistration")
		return nil
	}

	Log.Infow("Attempting to deregister from registry", "registry", s.registryAddr, "advertise", s.advertiseAddr)

	reqBody := RegistryRequest{Address: s.advertiseAddr}
	jsonData, err := json.Marshal(reqBody) // Use standard JSON
	if err != nil {
		Log.Errorw("Failed to marshal deregister request", "error", err)
		return err // Log but don't block shutdown further
	}

	ctx, cancel := context.WithTimeout(context.Background(), registryTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.registryAddr+"/deregister", bytes.NewBuffer(jsonData))
	if err != nil {
		Log.Errorw("Failed to create deregister request", "error", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Network errors during shutdown are warnings
		Log.Warnw("Failed to send deregister request", "registry", s.registryAddr, "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Read body
		Log.Warnw("Registry returned non-OK status for deregister", "registry", s.registryAddr, "status", resp.Status, "body", string(bodyBytes))
		return fmt.Errorf("registry returned non-OK status for deregister: %s", resp.Status)
	}
	Log.Infow("Successfully deregistered from registry", "registry", s.registryAddr)
	return nil
}

func (s *Server) startHeartbeat() {
	if s.registryAddr == "" || s.advertiseAddr == "" {
		Log.Warnw("Registry or advertise address not set, skipping heartbeat")
		return
	}

	s.wg.Add(1) // Increment wait group counter
	go func() {
		defer s.wg.Done() // Decrement counter when goroutine exits
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()

		Log.Infow("Starting heartbeat ticker", "interval", heartbeatInterval, "registry", s.registryAddr, "advertise", s.advertiseAddr)

		reqBody := RegistryRequest{Address: s.advertiseAddr}
		jsonData, err := json.Marshal(reqBody) // Use standard JSON
		// Handle initial marshalling error - likely fatal if it persists
		if err != nil {
			Log.Errorw("CRITICAL: Failed to marshal heartbeat request on startup, heartbeat disabled", "error", err)
			return // Stop the heartbeat goroutine
		}

		for {
			select {
			case <-ticker.C:
				// Create a new context for each request attempt
				reqCtx, reqCancel := context.WithTimeout(s.heartbeatCtx, registryTimeout) // Child of main heartbeat context

				// Create request using standard net/http
				req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, s.registryAddr+"/heartbeat", bytes.NewBuffer(jsonData))
				if err != nil {
					Log.Errorw("Failed to create heartbeat request (will retry)", "error", err)
					reqCancel() // Must cancel context if request creation fails
					continue    // Skip this tick
				}
				req.Header.Set("Content-Type", "application/json")

				// Send request
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					// Network errors are expected sometimes, just log warning
					Log.Warnw("Failed to send heartbeat (will retry)", "error", err)
					// Ensure context is cancelled even on error
					reqCancel()
					continue
				}

				// Process response
				if resp.StatusCode != http.StatusOK {
					// If registry doesn't know us (e.g., registry restarted), try re-registering
					bodyBytes, _ := io.ReadAll(resp.Body) // Read body for logging
					Log.Warnw("Registry returned non-OK for heartbeat, attempting re-registration", "status", resp.Status, "body", string(bodyBytes))
					// Close body *before* potential re-register call
					resp.Body.Close()
					if regErr := s.registerWithRegistry(); regErr != nil {
						Log.Errorw("Failed to re-register after failed heartbeat", "error", regErr)
					}
				} else {
					// Success, ensure body is read and closed to reuse connection
					_, _ = io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}

				// Cancel context after request is done (success or handled error)
				reqCancel()

			case <-s.heartbeatCtx.Done():
				Log.Infow("Heartbeat ticker stopping due to context cancellation.")
				return // Exit goroutine
			}
		}
	}()
}

func (s *Server) stopHeartbeat() {
	Log.Info("Stopping heartbeat...")
	s.heartbeatCancel() // Signal the heartbeat goroutine to stop via context cancellation
	s.wg.Wait()         // Wait for the heartbeat goroutine to finish (calls Done())
	Log.Info("Heartbeat stopped.")
}

// --- Server Execution and Handlers ---

func (s *Server) RunAndBlock() {
	// 1. Register first
	if err := s.registerWithRegistry(); err != nil {
		// Decide if failure is fatal based on application requirements
		Log.Errorw("Failed to register with registry on startup, continuing without registration/heartbeat", "error", err)
		// To make it fatal: Log.Fatalw("Failed to register with registry on startup", "error", err)
	} else {
		// 2. Start heartbeat only if registration was attempted/successful (or if optional)
		s.startHeartbeat()
	}

	// 3. Setup fasthttp router
	router := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/ping":
			s.pingHandler(ctx)
		case "/read":
			s.readHandler(ctx)
		case "/prepare":
			s.prepareHandler(ctx)
		case "/commit":
			s.commitHandler(ctx)
		case "/abort":
			s.abortHandler(ctx)
		case "/cache":
			s.cacheHandler(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	address := fmt.Sprintf(":%d", s.port)
	s.fasthttpServer = &fasthttp.Server{Handler: router} // Store server instance

	// Channel to signal fasthttp server startup errors
	serverErrChan := make(chan error, 1)

	Log.Infow("Executor server starting", "address", address, "advertise", s.advertiseAddr)

	// 4. Start fasthttp server in a goroutine
	go func() {
		err := s.fasthttpServer.ListenAndServe(address)
		// Send error (or nil if shutdown gracefully) to the channel
		serverErrChan <- err
	}()

	// 5. Wait for shutdown signal OR server error
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		// Server failed to start or stopped unexpectedly
		if err != nil {
			Log.Errorw("Executor fasthttp server failed", "error", err)
			// If server fails, maybe try to deregister? Might not be possible.
		} else {
			Log.Info("Executor fasthttp server stopped gracefully (unexpectedly).")
		}
		// Proceed to stop heartbeat if it was running
		if s.registryAddr != "" && s.advertiseAddr != "" {
			s.stopHeartbeat()
		}

	case sig := <-sigs:
		// Received OS signal for shutdown
		Log.Infow("Shutdown signal received.", "signal", sig)

		// --- Graceful Shutdown Sequence ---
		// a. Stop Heartbeat
		if s.registryAddr != "" && s.advertiseAddr != "" {
			s.stopHeartbeat()
		}

		// b. Deregister from Registry
		if err := s.deregisterFromRegistry(); err != nil {
			Log.Warnw("Deregistration failed during shutdown", "error", err)
			// Continue shutdown anyway
		}

		// c. Shutdown fasthttp server
		Log.Info("Shutting down fasthttp server...")
		if err := s.fasthttpServer.Shutdown(); err != nil {
			Log.Warnw("Error during fasthttp server shutdown", "error", err)
		} else {
			Log.Info("Executor fasthttp server stopped.")
		}
		// Wait for the server goroutine to finish after calling Shutdown
		<-serverErrChan // This will receive the nil error from ListenAndServe after Shutdown completes
	}

	Log.Info("Executor shutdown process complete.")
	fmt.Printf("Final Cache Stats: %v\n", s.reader.GetCacheStatistic())
}

// --- fasthttp Handlers ---

func (s *Server) pingHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("pong")
}

func (s *Server) cacheHandler(ctx *fasthttp.RequestCtx) {
	method := string(ctx.Method())

	if method == http.MethodGet { // Use http constants for methods
		status := s.reader.GetCacheStatistic()
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("text/plain; charset=utf-8") // Be explicit about content type
		ctx.SetBodyString(status)
		return
	}

	if method == http.MethodPost { // Use POST for actions like clearing
		s.reader.ClearCache()
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("text/plain; charset=utf-8")
		ctx.SetBodyString("Cache cleared successfully")
		return
	}

	// Handle unsupported request methods
	ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
	ctx.SetBodyString("Method not allowed")
}

func (s *Server) readHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		// Use structured logging
		Log.Debugw("Read request processing finished", "latency_ms", time.Since(startTime).Milliseconds())
	}()

	var req network.ReadRequest
	// Use jsoniter for application data unmarshalling
	if err := json2.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid read request body: %s", err.Error())
		Log.Errorw(errMsg, "body", string(ctx.PostBody()))
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	Log.Infow("Read request received", "dsName", req.DsName, "key", req.Key, "startTime", req.StartTime, "config", req.Config)

	item, dataType, gk, err := s.reader.Read(req.DsName, req.Key, req.StartTime, req.Config, true)

	var response network.ReadResponse
	if err != nil {
		Log.Warnw("Read operation failed", "dsName", req.DsName, "key", req.Key, "error", err)
		response = network.ReadResponse{
			Status: "Error",
			ErrMsg: err.Error(),
		}
		ctx.SetStatusCode(fasthttp.StatusInternalServerError) // Or map specific errors
	} else {
		response = network.ReadResponse{
			Status:       "OK",
			DataStrategy: dataType,
			Data:         item,
			GroupKey:     gk,
			ItemType:     network.GetItemType(req.DsName),
		}
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	// Use jsoniter for response marshalling
	respBytes, marshalErr := json2.Marshal(response)
	if marshalErr != nil {
		Log.Errorw("Failed to marshal read response", "error", marshalErr)
		// Send a generic error if marshalling fails
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	ctx.Write(respBytes)
}

func (s *Server) prepareHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Debugw("Prepare request processing finished", "latency_ms", time.Since(startTime).Milliseconds(), "Topic", "CheckPoint")
	}()

	var req network.PrepareRequest
	// Use jsoniter for application data
	if err := json2.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid prepare request body: %s", err.Error())
		Log.Errorw(errMsg, "body", string(ctx.PostBody()))
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	Log.Infow("Prepare request received", "dsName", req.DsName, "itemCount", len(req.ItemList), "startTime", req.StartTime, "config", req.Config, "validationMapKeys", getMapKeys(req.ValidationMap))

	verMap, tCommit, err := s.committer.Prepare(req.DsName, req.ItemList,
		req.StartTime, req.Config, req.ValidationMap)
	var resp network.PrepareResponse
	if err != nil {
		Log.Warnw("Prepare operation failed", "dsName", req.DsName, "startTime", req.StartTime, "error", err)
		resp = network.PrepareResponse{
			Status: "Error",
			ErrMsg: err.Error(),
		}
		ctx.SetStatusCode(fasthttp.StatusInternalServerError) // Or map specific errors (e.g., Conflict)
	} else {
		resp = network.PrepareResponse{
			Status:  "OK",
			VerMap:  verMap,
			TCommit: tCommit,
		}
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	// Use jsoniter for response
	respBytes, marshalErr := json2.Marshal(resp)
	if marshalErr != nil {
		Log.Errorw("Failed to marshal prepare response", "error", marshalErr)
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	ctx.Write(respBytes)
}

func (s *Server) commitHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Debugw("Commit request processing finished", "latency_ms", time.Since(startTime).Milliseconds())
	}()

	var req network.CommitRequest
	// Use jsoniter for application data
	if err := json2.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid commit request body: %s", err.Error())
		Log.Errorw(errMsg, "body", string(ctx.PostBody()))
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	Log.Infow("Commit request received", "dsName", req.DsName, "commitInfoCount", len(req.List), "tCommit", req.TCommit)

	err := s.committer.Commit(req.DsName, req.List, req.TCommit)
	var resp network.Response[string] // Generic response type
	if err != nil {
		Log.Warnw("Commit operation failed", "dsName", req.DsName, "tCommit", req.TCommit, "error", err)
		resp = network.Response[string]{
			Status: "Error",
			ErrMsg: err.Error(),
		}
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	} else {
		resp = network.Response[string]{
			Status: "OK",
		}
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	// Use jsoniter for response
	respBytes, marshalErr := json2.Marshal(resp)
	if marshalErr != nil {
		Log.Errorw("Failed to marshal commit response", "error", marshalErr)
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	ctx.Write(respBytes)
}

func (s *Server) abortHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		Log.Debugw("Abort request processing finished", "latency_ms", time.Since(startTime).Milliseconds())
	}()

	var req network.AbortRequest
	// Use jsoniter for application data
	if err := json2.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid abort request body: %s", err.Error())
		Log.Errorw(errMsg, "body", string(ctx.PostBody()))
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	Log.Infow("Abort request received", "dsName", req.DsName, "keyListCount", len(req.KeyList), "groupKey", req.GroupKeyList)

	err := s.committer.Abort(req.DsName, req.KeyList, req.GroupKeyList)
	var resp network.Response[string] // Generic response type
	if err != nil {
		// Abort failing is usually just a warning unless it leaks resources
		Log.Warnw("Abort operation failed", "dsName", req.DsName, "groupKey", req.GroupKeyList, "error", err)
		resp = network.Response[string]{
			Status: "Error",
			ErrMsg: err.Error(),
		}
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	} else {
		resp = network.Response[string]{
			Status: "OK",
		}
		ctx.SetStatusCode(fasthttp.StatusOK)
	}

	// Use jsoniter for response
	respBytes, marshalErr := json2.Marshal(resp)
	if marshalErr != nil {
		Log.Errorw("Failed to marshal abort response", "error", marshalErr)
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	ctx.Write(respBytes)
}

// --- Main Function and Setup ---

// Command-line flags
var (
	port              = flag.Int("p", 8000, "Port to listen on")
	poolSize          = flag.Int("s", 60, "Database connection pool size")
	traceFlag         = flag.Bool("trace", false, "Enable execution tracing to trace.out")
	pprofFlag         = flag.Bool("pprof", false, "Enable CPU profiling to executor_cpu_profile.prof")
	workloadType      = flag.String("w", "", "Workload type (e.g., iot, social, order, ycsb)")
	db_combination    = flag.String("db", "", "Database combination for YCSB workload (comma-separated, e.g., Redis,MongoDB1)")
	benConfigPath     = flag.String("bc", "", "Path to benchmark configuration YAML file (required)")
	cg                = flag.Bool("cg", false, "Enable Cherry Garcia Mode (config.Debug.CherryGarciaMode)")
	advertiseAddrFlag = flag.String("advertise-addr", "", "Address (host:port) to advertise to the registry (e.g., 1.2.3.4:8000, defaults to localhost:port)")
	registryAddrFlag  = flag.String("registry-addr", "", "HTTP address of the registry service (e.g., http://localhost:9000)")
)

var Log *zap.SugaredLogger

// Global benchmark config loaded from YAML
var benConfig = benconfig.BenchmarkConfig{}

func main() {
	flag.Parse() // Parse flags early
	newLogger()  // Setup logger immediately after parsing

	// Validate required flags
	if *benConfigPath == "" {
		Log.Fatal("Benchmark Configuration Path (--bc) must be specified")
	}
	if *workloadType == "" {
		Log.Fatal("Workload Type (--w) must be specified")
	}
	if *workloadType == "ycsb" && *db_combination == "" {
		Log.Fatal("Database Combination (--db) must be specified for YCSB workload")
	}

	// Load benchmark configuration from YAML
	err := loadConfig(*benConfigPath)
	if err != nil {
		Log.Fatalw("Failed to load benchmark configuration", "path", *benConfigPath, "error", err)
	}

	// Setup profiling and tracing if enabled
	if *pprofFlag {
		startCPUProfile()
		// Optionally add memory profiling setup here
	}
	if *traceFlag {
		stopTrace := startTrace()
		defer stopTrace() // Ensure trace stops on exit
	}

	// Apply specific debug configurations
	if *cg {
		Log.Info("Running under Cherry Garcia Mode")
		config.Debug.CherryGarciaMode = true
	}
	config.Debug.DebugMode = false // Ensure standard debug mode is off unless explicitly enabled

	// Establish database connections based on workload
	connMap := getConnMap(*workloadType, *db_combination)
	if len(connMap) == 0 {
		Log.Fatalw("No database connections established for workload", "workload", *workloadType, "db_combination", *db_combination)
	}

	// Determine which datastore names this executor handles
	handledDsNames := make([]string, 0, len(connMap))
	for dsName := range connMap {
		handledDsNames = append(handledDsNames, dsName)
	}
	Log.Infow("Executor configured to handle datastores", "dsNames", handledDsNames)

	// Determine the address to advertise to the registry
	finalAdvertiseAddr := *advertiseAddrFlag
	if finalAdvertiseAddr == "" {
		// Default to localhost:port - may not be reachable externally!
		finalAdvertiseAddr = fmt.Sprintf("localhost:%d", *port)
		Log.Warnw("Advertise address (--advertise-addr) not specified, defaulting to loopback address", "address", finalAdvertiseAddr)
	}
	// Optional: Add logic here to determine public IP if needed

	// Validate registry address format
	if *registryAddrFlag != "" && !strings.HasPrefix(*registryAddrFlag, "http://") && !strings.HasPrefix(*registryAddrFlag, "https://") {
		Log.Warnw("Registry address might be missing scheme (http:// or https://), assuming http://", "registry-addr", *registryAddrFlag)
		*registryAddrFlag = "http://" + *registryAddrFlag // Default to http if scheme missing
	}

	// Create the time source (Oracle)
	oracle := timesource.NewGlobalTimeSource(benConfig.TimeOracleUrl)

	// Create the main Server instance
	server := NewServer(*port, finalAdvertiseAddr, *registryAddrFlag, handledDsNames, connMap, &redis.RedisItemFactory{}, oracle)

	// Run the server (this includes registration, heartbeat, fasthttp server, and blocks until shutdown)
	server.RunAndBlock()

	Log.Info("Main function finished.") // Should be reached after RunAndBlock completes
}

// --- Configuration Loading ---

func loadConfig(configPath string) error {
	Log.Infow("Loading benchmark configuration", "path", configPath)
	loader := aconfig.LoaderFor(&benConfig, aconfig.Config{
		SkipDefaults: true,
		SkipFiles:    false,
		SkipEnv:      true,
		SkipFlags:    true, // Flags are parsed separately
		Files:        []string{configPath},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
			".yml":  aconfigyaml.New(), // Support .yml too
		},
	})

	if err := loader.Load(); err != nil {
		return fmt.Errorf("error loading config file '%s': %w", configPath, err)
	}

	// Validate essential config values
	if benConfig.TimeOracleUrl == "" {
		return fmt.Errorf("timeOracleUrl must be specified in the benchmark configuration")
	}
	Log.Infow("Benchmark configuration loaded successfully", "timeOracleUrl", benConfig.TimeOracleUrl)
	// Add more validation as needed (e.g., check required DB addresses based on workload)

	return nil
}

// --- Profiling and Tracing Setup ---

func startCPUProfile() {
	f, err := os.Create("executor_cpu_profile.prof")
	if err != nil {
		Log.Errorw("Cannot create CPU profile file", "error", err)
		return
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		Log.Errorw("Cannot start CPU profile", "error", err)
		f.Close() // Close the file if starting profile fails
		return
	}
	Log.Info("CPU profiling enabled, writing to executor_cpu_profile.prof")
	// Schedule StopCPUProfile on shutdown (can be done via defer in main, but cleaner with explicit stop)
	// For simplicity, using defer in main might be okay, but this is more robust if main structure changes.
	// We'll rely on the OS signal handling to eventually stop it via process exit for now.
	// A more robust way would be to hook into the shutdown signal handler.
	// Let's add it to the shutdown sequence:
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // Wait for shutdown signal
		Log.Info("Stopping CPU profile due to shutdown signal...")
		pprof.StopCPUProfile()
		f.Close() // Close the profile file
		Log.Info("CPU profile stopped.")
	}()
}

func startTrace() func() {
	f, err := os.Create("trace.out")
	if err != nil {
		Log.Errorw("Cannot create trace file", "error", err)
		return func() {} // Return no-op stop function
	}
	err = trace.Start(f)
	if err != nil {
		Log.Errorw("Cannot start trace", "error", err)
		f.Close()
		return func() {} // Return no-op stop function
	}
	Log.Info("Execution tracing enabled, writing to trace.out")
	return func() {
		Log.Info("Stopping execution trace...")
		trace.Stop()
		f.Close() // Close the trace file
		Log.Info("Execution trace stopped.")
	}
}

// --- Database Connection Map ---

func getConnMap(wType string, dbComb string) map[string]txn.Connector {
	connMap := make(map[string]txn.Connector)
	Log.Infow("Setting up database connections", "workload", wType, "dbCombination", dbComb)

	switch wType {
	case "iot":
		if benConfig.MongoDBAddr1 != "" {
			connMap["MongoDB"] = getMongoConn(1)
		} else {
			Log.Warn("MongoDBAddr1 not configured for iot workload")
		}
		if benConfig.RedisAddr != "" {
			connMap["Redis"] = getRedisConn(1)
		} else {
			Log.Warn("RedisAddr not configured for iot workload")
		}
	case "social":
		if benConfig.MongoDBAddr1 != "" {
			connMap["MongoDB"] = getMongoConn(1)
		} else {
			Log.Warn("MongoDBAddr1 not configured for social workload")
		}
		if benConfig.RedisAddr != "" {
			connMap["Redis"] = getRedisConn(1)
		} else {
			Log.Warn("RedisAddr not configured for social workload")
		}
		if len(benConfig.CassandraAddr) > 0 {
			connMap["Cassandra"] = getCassandraConn()
		} else {
			Log.Warn("CassandraAddr not configured for social workload")
		}
	case "order":
		if benConfig.MongoDBAddr1 != "" {
			connMap["MongoDB"] = getMongoConn(1)
		} else {
			Log.Warn("MongoDBAddr1 not configured for order workload")
		}
		if benConfig.KVRocksAddr != "" {
			connMap["KVRocks"] = getKVRocksConn()
		} else {
			Log.Warn("KVRocksAddr not configured for order workload")
		}
		if benConfig.RedisAddr != "" {
			connMap["Redis"] = getRedisConn(1)
		} else {
			Log.Warn("RedisAddr not configured for order workload")
		}
		if len(benConfig.CassandraAddr) > 0 {
			connMap["Cassandra"] = getCassandraConn()
		} else {
			Log.Warn("CassandraAddr not configured for order workload")
		}
	case "ycsb":
		dbList := strings.Split(dbComb, ",")
		for _, db := range dbList {
			db = strings.TrimSpace(db) // Trim whitespace
			if db == "" {
				continue
			}
			Log.Infow("Configuring database for YCSB", "db", db)
			switch db {
			case "Redis":
				if benConfig.RedisAddr != "" {
					connMap["Redis"] = getRedisConn(1)
				} else {
					Log.Warn("RedisAddr not configured despite being requested in --db")
				}
			case "MongoDB1":
				if benConfig.MongoDBAddr1 != "" {
					connMap["MongoDB1"] = getMongoConn(1)
				} else {
					Log.Warn("MongoDBAddr1 not configured despite being requested in --db")
				}
			case "MongoDB2":
				if benConfig.MongoDBAddr2 != "" {
					connMap["MongoDB2"] = getMongoConn(2)
				} else {
					Log.Warn("MongoDBAddr2 not configured despite being requested in --db")
				}
			case "KVRocks":
				if benConfig.KVRocksAddr != "" {
					connMap["KVRocks"] = getKVRocksConn()
				} else {
					Log.Warn("KVRocksAddr not configured despite being requested in --db")
				}
			case "CouchDB":
				if benConfig.CouchDBAddr != "" {
					connMap["CouchDB"] = getCouchConn()
				} else {
					Log.Warn("CouchDBAddr not configured despite being requested in --db")
				}
			case "Cassandra":
				if len(benConfig.CassandraAddr) > 0 {
					connMap["Cassandra"] = getCassandraConn()
				} else {
					Log.Warn("CassandraAddr not configured despite being requested in --db")
				}
			case "DynamoDB":
				if benConfig.DynamoDBAddr != "" {
					connMap["DynamoDB"] = getDynamoConn()
				} else {
					Log.Warn("DynamoDBAddr not configured despite being requested in --db")
				}
			case "TiKV":
				if len(benConfig.TiKVAddr) > 0 {
					connMap["TiKV"] = getTiKVConn()
				} else {
					Log.Warn("TiKVAddr not configured despite being requested in --db")
				}
			default:
				Log.Errorf("Invalid database name '%s' in --db combination", db)
			}
		}
	default:
		Log.Fatalf("Unsupported workload type: %s", wType)
	}

	if len(connMap) == 0 {
		Log.Warnw("No database connections were successfully configured based on workload and config", "workload", wType, "dbCombination", dbComb)
	} else {
		dsNames := make([]string, 0, len(connMap))
		for name := range connMap {
			dsNames = append(dsNames, name)
		}
		Log.Infow("Database connections established", "datastores", dsNames)
	}
	return connMap
}

// --- Logger Setup ---

func newLogger() {
	logLevelEnv := os.Getenv("LOG")
	var level zapcore.Level

	switch strings.ToUpper(logLevelEnv) {
	case "DEBUG":
		level = zap.DebugLevel
	case "INFO":
		level = zap.InfoLevel
	case "WARN":
		level = zap.WarnLevel
	case "ERROR":
		level = zap.ErrorLevel
	case "FATAL":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel // Default to INFO if LOG env var is not set or invalid
	}

	// Use production config for better performance, but with development settings for readability
	conf := zap.NewProductionConfig()
	conf.Level = zap.NewAtomicLevelAt(level)
	conf.Encoding = "console"                                         // Human-readable console output
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Standard time format
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Colored level
	conf.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder      // Short caller path
	conf.OutputPaths = []string{"stdout"}
	conf.ErrorOutputPaths = []string{"stderr"}

	logger, err := conf.Build()
	if err != nil {
		// Fallback to standard logger if zap fails
		log.Printf("Failed to initialize zap logger: %v. Falling back to standard logger.", err)
		Log = zap.NewNop().Sugar() // Use a no-op logger if zap fails
		return
	}

	Log = logger.Sugar()
	Log.Infow("Logger initialized", "level", level.String())
}

// --- Individual Database Connection Helpers ---

func getKVRocksConn() *redis.RedisConnection {
	Log.Infow("Connecting to KVRocks", "address", benConfig.KVRocksAddr)
	kvConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  benConfig.KVRocksAddr,
		Password: benConfig.RedisPassword, // Assuming same password config as Redis
		PoolSize: *poolSize,
	})
	err := kvConn.Connect()
	if err != nil {
		Log.Fatalw("Failed to connect to KVRocks", "address", benConfig.KVRocksAddr, "error", err)
	}
	// Optional: Ping check
	// if _, err := kvConn.Get("ping"); err != nil { // Use GET for KVRocks? Or specific PING?
	// 	Log.Warnw("KVRocks ping failed", "address", benConfig.KVRocksAddr, "error", err)
	// }
	Log.Info("Connected to KVRocks successfully.")
	return kvConn
}

func getCouchConn() *couchdb.CouchDBConnection {
	Log.Infow("Connecting to CouchDB", "address", benConfig.CouchDBAddr, "dbName", "oreo")
	couchConn := couchdb.NewCouchDBConnection(&couchdb.ConnectionOptions{
		Address:  benConfig.CouchDBAddr,
		Username: benConfig.CouchDBUsername, // Use configured username/password
		Password: benConfig.CouchDBPassword,
		DBName:   "oreo", // Hardcoded DB name? Consider making configurable
	})
	err := couchConn.Connect()
	if err != nil {
		Log.Fatalw("Failed to connect to CouchDB", "address", benConfig.CouchDBAddr, "error", err)
	}
	Log.Info("Connected to CouchDB successfully.")
	return couchConn
}

func getMongoConn(id int) *mongo.MongoConnection {
	var address string
	switch id {
	case 1:
		address = benConfig.MongoDBAddr1
	case 2:
		address = benConfig.MongoDBAddr2
	default:
		Log.Fatalf("Invalid MongoDB connection ID requested: %d", id)
		return nil // Should not be reached
	}
	Log.Infow("Connecting to MongoDB", "id", id, "address", address, "dbName", "oreo", "collection", "benchmark")
	mongoConn := mongo.NewMongoConnection(&mongo.ConnectionOptions{
		Address:        address,
		DBName:         "oreo",      // Hardcoded DB name?
		CollectionName: "benchmark", // Hardcoded collection?
		Username:       benConfig.MongoDBUsername,
		Password:       benConfig.MongoDBPassword,
		// PoolSize not directly configurable here, managed by driver
	})
	err := mongoConn.Connect()
	if err != nil {
		Log.Fatalw("Failed to connect to MongoDB", "id", id, "address", address, "error", err)
	}
	Log.Infof("Connected to MongoDB%d successfully.", id)
	return mongoConn
}

func getRedisConn(id int) *redis.RedisConnection {
	var address string
	switch id {
	case 1:
		address = benConfig.RedisAddr
	// Add cases for RedisAddr2, etc. if needed
	default:
		Log.Fatalf("Invalid Redis connection ID requested: %d", id)
		return nil // Should not be reached
	}
	Log.Infow("Connecting to Redis", "id", id, "address", address)
	redisConn := redis.NewRedisConnection(&redis.ConnectionOptions{
		Address:  address,
		Password: benConfig.RedisPassword,
		PoolSize: *poolSize,
	})
	err := redisConn.Connect() // Connect attempts to ping
	if err != nil {
		Log.Fatalw("Failed to connect to Redis", "id", id, "address", address, "error", err)
	}
	Log.Infof("Connected to Redis%d successfully.", id)
	return redisConn
}

func getCassandraConn() *cassandra.CassandraConnection {
	Log.Infow("Connecting to Cassandra", "hosts", benConfig.CassandraAddr, "keyspace", "oreo")
	cassConn := cassandra.NewCassandraConnection(&cassandra.ConnectionOptions{
		Hosts:    benConfig.CassandraAddr,
		Keyspace: "oreo", // Hardcoded keyspace?
		// PoolSize (NumConns) might be configurable via options if needed
	})
	err := cassConn.Connect()
	if err != nil {
		Log.Fatalw("Failed to connect to Cassandra", "hosts", benConfig.CassandraAddr, "error", err)
	}
	Log.Info("Connected to Cassandra successfully.")
	return cassConn
}

func getDynamoConn() *dynamodb.DynamoDBConnection {
	Log.Infow("Connecting to DynamoDB", "endpoint", benConfig.DynamoDBAddr, "tableName", "oreo")
	dynamoConn := dynamodb.NewDynamoDBConnection(&dynamodb.ConnectionOptions{
		TableName: "oreo",                 // Hardcoded table name?
		Endpoint:  benConfig.DynamoDBAddr, // Use Endpoint for local/mock, otherwise relies on AWS SDK defaults
		// Credentials handled by AWS SDK (env vars, instance profile, etc.)
	})
	err := dynamoConn.Connect() // Connect likely initializes the client
	if err != nil {
		Log.Fatalw("Failed to connect to DynamoDB", "endpoint", benConfig.DynamoDBAddr, "error", err)
	}
	Log.Info("Connected to DynamoDB successfully.")
	return dynamoConn
}

func getTiKVConn() *tikv.TiKVConnection {
	Log.Infow("Connecting to TiKV", "pdAddrs", benConfig.TiKVAddr)
	tikvConn := tikv.NewTiKVConnection(&tikv.ConnectionOptions{
		PDAddrs: benConfig.TiKVAddr,
		// Security options (TLS) might be needed here via config
	})
	err := tikvConn.Connect() // Connect initializes the client
	if err != nil {
		Log.Fatalw("Failed to connect to TiKV", "pdAddrs", benConfig.TiKVAddr, "error", err)
	}
	Log.Info("Connected to TiKV successfully.")
	return tikvConn
}

// --- Helper Functions ---

// getMapKeys is a small helper for logging map keys without values
func getMapKeys(m map[string]txn.PredicateInfo) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
