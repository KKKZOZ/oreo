package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"syscall"
	"time"

	"benchmark/pkg/benconfig" // nolint
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	jsoniter "github.com/json-iterator/go" // Keep for application logic if needed
	"github.com/kkkzoz/oreo/pkg/config"
	"github.com/kkkzoz/oreo/pkg/datastore/redis"
	"github.com/kkkzoz/oreo/pkg/discovery"
	"github.com/kkkzoz/oreo/pkg/logger"
	"github.com/kkkzoz/oreo/pkg/network"
	"github.com/kkkzoz/oreo/pkg/serializer"
	"github.com/kkkzoz/oreo/pkg/timesource"
	"github.com/kkkzoz/oreo/pkg/txn"
	"github.com/valyala/fasthttp"
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
	port           int
	advertiseAddr  string   // Address to advertise to the registry
	registryAddrs  []string // Addresses of the registry service
	handledDsNames []string
	reader         network.Reader
	committer      network.Committer
	fasthttpServer *fasthttp.Server // Keep track for shutdown

	// Service registry interface
	registry discovery.ServiceRegistry
}

// NewServer modified to accept registry type
func NewServer(
	port int,
	advertiseAddr string,
	registryAddrs []string,
	handledDsNames []string,
	connMap map[string]txn.Connector,
	factory txn.DataItemFactory,
	timeSource timesource.TimeSourcer,
	registryType string,
) *Server {
	reader := *network.NewReader(connMap, factory, serializer.NewJSON2Serializer(), network.NewCacher())

	// Extract database connection addresses from connMap
	dbConnections := make(map[string]string)
	for dsName, connector := range connMap {
		// Get real address based on connector type
		switch conn := connector.(type) {
		case interface{ GetAddress() string }:
			dbConnections[dsName] = conn.GetAddress()
		default:
			dbConnections[dsName] = advertiseAddr // Fallback to executor address
		}
	}

	// Initialize the appropriate registry based on type
	var registry discovery.ServiceRegistry
	switch registryType {
	case "etcd":
		endpoints := make([]string, 0, len(registryAddrs))
		for _, addr := range registryAddrs {
			for _, endpoint := range strings.Split(addr, ",") {
				endpoint = strings.TrimSpace(endpoint)
				if endpoint != "" {
					endpoints = append(endpoints, endpoint)
				}
			}
		}
		if len(endpoints) == 0 {
			logger.Fatal("No etcd registry endpoints provided")
		}
		config := discovery.DefaultRegistryConfig()
		etcdReg, err := discovery.NewEtcdServiceRegistry(endpoints, "/oreo/services", config)
		if err != nil {
			logger.Fatalw("Failed to create etcd registry", "error", err)
		}
		registry = etcdReg
	case "http":
		registry = discovery.NewHTTPServiceRegistry(registryAddrs, advertiseAddr, handledDsNames)
	default:
		panic("unknown registry type")
	}

	return &Server{
		port:           port,
		advertiseAddr:  advertiseAddr,
		registryAddrs:  registryAddrs,
		handledDsNames: handledDsNames,
		reader:         reader,
		committer:      *network.NewCommitter(connMap, reader, serializer.NewJSON2Serializer(), factory, timeSource),
		registry:       registry,
	}
}

// --- Registry Interaction ---

const (
	defaultRegistryTimeout = 5 * time.Second
)

// Modified registry methods to use the interface
func (s *Server) registerWithRegistry() error {
	if s.registry == nil {
		return fmt.Errorf("registry not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultRegistryTimeout)
	defer cancel()
	return s.registry.Register(ctx, s.advertiseAddr, s.handledDsNames, nil)
}

func (s *Server) deregisterFromRegistry() error {
	if s.registry == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultRegistryTimeout)
	defer cancel()
	return s.registry.Deregister(ctx, s.advertiseAddr)
}

// --- Server Execution and Handlers ---

func (s *Server) RunAndBlock() {
	// 1. Register first
	if err := s.registerWithRegistry(); err != nil {
		logger.Warnw(
			"Failed to register with registry on startup, continuing without registration",
			"error",
			err,
		)
	}
	logger.Infow("Registered with registry",
		"address", s.advertiseAddr,
		"handledDsNames", s.handledDsNames)

	// 2. Setup fasthttp router
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
	s.fasthttpServer = &fasthttp.Server{Handler: router}

	// 3.Channel to signal fasthttp server startup errors
	serverErrChan := make(chan error, 1)

	logger.Infow("Executor server starting", "address", address, "advertise", s.advertiseAddr)

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
			logger.Errorw("Executor fasthttp server failed", "error", err)
			// If server fails, maybe try to deregister? Might not be possible.
		} else {
			logger.Info("Executor fasthttp server stopped gracefully (unexpectedly).")
		}
		// Registry will handle its own cleanup

	case sig := <-sigs:
		// Received OS signal for shutdown
		logger.Infow("Shutdown signal received.", "signal", sig)

		// --- Graceful Shutdown Sequence ---
		// a. Deregister from Registry (this will also stop heartbeat)
		if err := s.deregisterFromRegistry(); err != nil {
			logger.Warnw("Deregistration failed during shutdown", "error", err)
			// Continue shutdown anyway
		}

		// c. Shutdown fasthttp server
		logger.Info("Shutting down fasthttp server...")
		if err := s.fasthttpServer.Shutdown(); err != nil {
			logger.Warnw("Error during fasthttp server shutdown", "error", err)
		} else {
			logger.Info("Executor fasthttp server stopped.")
		}
		// Wait for the server goroutine to finish after calling Shutdown
		<-serverErrChan // This will receive the nil error from ListenAndServe after Shutdown completes
	}

	logger.Info("Executor shutdown process complete.")
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
		logger.Debugw(
			"Read request processing finished",
			"latency_ms",
			time.Since(startTime).Milliseconds(),
		)
	}()

	var req network.ReadRequest
	// Use jsoniter for application data unmarshalling
	if err := json2.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid read request body: %s", err.Error())
		logger.Errorw(errMsg, "body", string(ctx.PostBody()))
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	logger.Infow(
		"Read request received",
		"dsName",
		req.DsName,
		"key",
		req.Key,
		"startTime",
		req.StartTime,
		"config",
		req.Config,
	)

	item, dataType, gk, err := s.reader.Read(req.DsName, req.Key, req.StartTime, req.Config, true)

	var response network.ReadResponse
	if err != nil {
		logger.Warnw("Read operation failed", "dsName", req.DsName, "key", req.Key, "error", err)
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
		logger.Errorw("Failed to marshal read response", "error", marshalErr)
		// Send a generic error if marshalling fails
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	_, err = ctx.Write(respBytes)
	if err != nil {
		logger.Errorw("Failed to write response", "error", err)
	}
}

func (s *Server) prepareHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		logger.Debugw(
			"Prepare request processing finished",
			"latency_ms",
			time.Since(startTime).Milliseconds(),
			"Topic",
			"CheckPoint",
		)
	}()

	var req network.PrepareRequest
	// Use jsoniter for application data
	if err := json2.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid prepare request body: %s", err.Error())
		logger.Errorw(errMsg, "body", string(ctx.PostBody()))
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	logger.Infow(
		"Prepare request received",
		"dsName",
		req.DsName,
		"itemCount",
		len(req.ItemList),
		"startTime",
		req.StartTime,
		"config",
		req.Config,
		"validationMapKeys",
		getMapKeys(req.ValidationMap),
	)

	verMap, tCommit, err := s.committer.Prepare(req.DsName, req.ItemList,
		req.StartTime, req.Config, req.ValidationMap)
	var resp network.PrepareResponse
	if err != nil {
		logger.Warnw(
			"Prepare operation failed",
			"dsName",
			req.DsName,
			"startTime",
			req.StartTime,
			"error",
			err,
		)
		resp = network.PrepareResponse{
			Status: "Error",
			ErrMsg: err.Error(),
		}
		ctx.SetStatusCode(
			fasthttp.StatusInternalServerError,
		) // Or map specific errors (e.g., Conflict)
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
		logger.Errorw("Failed to marshal prepare response", "error", marshalErr)
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	_, err = ctx.Write(respBytes)
	if err != nil {
		logger.Errorw("Failed to write response", "error", err)
	}
}

func (s *Server) commitHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		logger.Debugw(
			"Commit request processing finished",
			"latency_ms",
			time.Since(startTime).Milliseconds(),
		)
	}()

	var req network.CommitRequest
	// Use jsoniter for application data
	if err := json2.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid commit request body: %s", err.Error())
		logger.Errorw(errMsg, "body", string(ctx.PostBody()))
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	logger.Infow(
		"Commit request received",
		"dsName",
		req.DsName,
		"commitInfoCount",
		len(req.List),
		"tCommit",
		req.TCommit,
	)

	err := s.committer.Commit(req.DsName, req.List, req.TCommit)
	var resp network.Response[string] // Generic response type
	if err != nil {
		// logger.Warnw("Commit operation failed", "dsName", req.DsName, "tCommit", req.TCommit, "error", err)
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
		logger.Errorw("Failed to marshal commit response", "error", marshalErr)
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	_, err = ctx.Write(respBytes)
	if err != nil {
		logger.Errorw("Failed to write response", "error", err)
	}
}

func (s *Server) abortHandler(ctx *fasthttp.RequestCtx) {
	startTime := time.Now()
	defer func() {
		logger.Debugw(
			"Abort request processing finished",
			"latency_ms",
			time.Since(startTime).Milliseconds(),
		)
	}()

	var req network.AbortRequest
	// Use jsoniter for application data
	if err := json2.Unmarshal(ctx.PostBody(), &req); err != nil {
		errMsg := fmt.Sprintf("Invalid abort request body: %s", err.Error())
		logger.Errorw(errMsg, "body", string(ctx.PostBody()))
		ctx.Error(errMsg, fasthttp.StatusBadRequest)
		return
	}

	logger.Infow(
		"Abort request received",
		"dsName",
		req.DsName,
		"keyListCount",
		len(req.KeyList),
		"groupKey",
		req.GroupKeyList,
	)

	err := s.committer.Abort(req.DsName, req.KeyList, req.GroupKeyList)
	var resp network.Response[string] // Generic response type
	if err != nil {
		// Abort failing is usually just a warning unless it leaks resources
		logger.Warnw(
			"Abort operation failed",
			"dsName",
			req.DsName,
			"groupKey",
			req.GroupKeyList,
			"error",
			err,
		)
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
		logger.Errorw("Failed to marshal abort response", "error", marshalErr)
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	_, err = ctx.Write(respBytes)
	if err != nil {
		logger.Errorw("Failed to write response", "error", err)
	}
}

// --- Main Function and Setup ---

// Command-line flags
var (
	port      = flag.Int("p", 8000, "Port to listen on")
	poolSize  = flag.Int("s", 60, "Database connection pool size")
	traceFlag = flag.Bool("trace", false, "Enable execution tracing to trace.out")
	pprofFlag = flag.Bool(
		"pprof",
		false,
		"Enable CPU profiling to executor_cpu_profile.prof",
	)
	workloadType   = flag.String("w", "", "Workload type (e.g., iot, social, order, ycsb)")
	db_combination = flag.String(
		"db",
		"",
		"Database combination for YCSB workload (comma-separated, e.g., Redis,MongoDB1)",
	)
	benConfigPath = flag.String(
		"bc",
		"",
		"Path to benchmark configuration YAML file (required)",
	)
	cg = flag.Bool(
		"cg",
		false,
		"Enable Cherry Garcia Mode (config.Debug.CherryGarciaMode)",
	)
	advertiseAddrFlag = flag.String(
		"advertise-addr",
		"",
		"Address (host:port) to advertise to the registry (e.g., 1.2.3.4:8000, defaults to localhost:port)",
	)
	registryAddrFlag = flag.String(
		"registry-addr",
		"",
		"Address(es) of the registry service. Use a single address or comma-separated list (e.g., http://localhost:9000,https://backup:9000)",
	)
	// New flag for registry type
	registryType = flag.String(
		"registry",
		"http",
		"Type of registry to use: 'http' or 'etcd'",
	)
)

// Global benchmark config loaded from YAML
var benConfig = benconfig.BenchmarkConfig{}

func main() {
	parseFlags()
	// Load benchmark configuration from YAML
	err := loadConfig(*benConfigPath)
	if err != nil {
		logger.Fatalw(
			"Failed to load benchmark configuration",
			"path",
			*benConfigPath,
			"error",
			err,
		)
	}

	// Setup profiling and tracing if enabled
	if *pprofFlag {
		startCPUProfile()
	}
	if *traceFlag {
		stopTrace := startTrace()
		defer stopTrace() // Ensure trace stops on exit
	}

	// Apply specific debug configurations
	if *cg {
		logger.Info("Running under Cherry Garcia Mode")
		config.Debug.CherryGarciaMode = true
	}
	config.Debug.DebugMode = false // Ensure standard debug mode is off unless explicitly enabled

	registryAddrs, err := determineRegistryAddrs(*registryType, *registryAddrFlag, &benConfig)
	if err != nil {
		logger.Fatalw("Failed to resolve registry addresses", "error", err)
	}

	if *registryType == "http" {
		registryAddrs = normalizeHTTPRegistryAddrs(registryAddrs)
	}

	logger.Infow("Resolved registry addresses", "type", *registryType, "addrs", registryAddrs)

	// Establish database connections based on workload
	connMap := getConnMap(*workloadType, *db_combination)

	// if len(connMap) == 0 {
	// 	logger.Fatalw(
	// 		"No database connections established for workload",
	// 		"workload",
	// 		*workloadType,
	// 		"db_combination",
	// 		*db_combination,
	// 	)
	// }

	// Determine which datastore names this executor handles
	handledDsNames := make([]string, 0, len(connMap))
	for dsName := range connMap {
		handledDsNames = append(handledDsNames, dsName)
	}
	logger.Infow("Executor configured to handle datastores", "dsNames", handledDsNames)

	// Create the time source (Oracle)
	oracle := timesource.NewGlobalTimeSource(benConfig.TimeOracleUrl)

	// Create the main Server instance with registry type
	server := NewServer(
		*port,
		*advertiseAddrFlag,
		registryAddrs,
		handledDsNames,
		connMap,
		&redis.RedisItemFactory{},
		oracle,
		*registryType, // Pass the registry type
	)

	// Run the server (this includes registration, heartbeat, fasthttp server, and blocks until shutdown)
	server.RunAndBlock()

	logger.Info("Main function finished.") // Should be reached after RunAndBlock completes
}

// --- Configuration Loading ---

func parseFlags() {
	flag.Parse()
	// Validate required flags
	if *benConfigPath == "" {
		logger.Fatal("Benchmark Configuration Path (--bc) must be specified")
	}

	if *workloadType == "" {
		logger.Fatal("Workload Type (--w) must be specified")
	}

	if *workloadType == "ycsb" && *db_combination == "" {
		logger.Fatal("Database Combination (--db) must be specified for YCSB workload")
	}

	if *advertiseAddrFlag == "" {
		logger.Fatal("Advertise address (--advertise-addr) not specified")
	}

	// Validate registry type
	if *registryType != "http" && *registryType != "etcd" {
		logger.Fatalf("Invalid registry type '%s'. Please use 'http' or 'etcd'.", *registryType)
	}
}

func loadConfig(configPath string) error {
	logger.Infow("Loading benchmark configuration", "path", configPath)
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
	logger.Infow(
		"Benchmark configuration loaded successfully",
		"timeOracleUrl",
		benConfig.TimeOracleUrl,
	)
	// Add more validation as needed (e.g., check required DB addresses based on workload)

	return nil
}

func determineRegistryAddrs(registryType, cliValue string, cfg *benconfig.BenchmarkConfig) ([]string, error) {
	cliValue = strings.TrimSpace(cliValue)

	switch registryType {
	case "http":
		if cliValue != "" {
			addresses := parseAddressList([]string{cliValue})
			if len(addresses) == 0 {
				return nil, fmt.Errorf("no http registry addresses found in --registry-addr")
			}
			return addresses, nil
		}
		addrs := cfg.ResolveHTTPRegistryAddrs()
		if len(addrs) == 0 {
			return nil, fmt.Errorf("no http registry addresses provided via --registry-addr or config")
		}
		addresses := parseAddressList(addrs)
		if len(addresses) == 0 {
			return nil, fmt.Errorf("no http registry addresses remained after parsing config")
		}
		return addresses, nil
	case "etcd":
		if cliValue != "" {
			addresses := parseAddressList([]string{cliValue})
			if len(addresses) == 0 {
				return nil, fmt.Errorf("no etcd registry addresses found in --registry-addr")
			}
			return addresses, nil
		}
		sources := make([]string, 0, 1+len(cfg.RegistryAddrs))
		if cfg.RegistryAddr != "" {
			sources = append(sources, cfg.RegistryAddr)
		}
		if len(cfg.RegistryAddrs) > 0 {
			sources = append(sources, cfg.RegistryAddrs...)
		}
		if len(sources) == 0 {
			return nil, fmt.Errorf("no etcd registry addresses provided via --registry-addr or config")
		}
		addresses := parseAddressList(sources)
		if len(addresses) == 0 {
			return nil, fmt.Errorf("no etcd registry addresses remained after parsing config")
		}
		return addresses, nil
	default:
		return nil, fmt.Errorf("unsupported registry type '%s'", registryType)
	}
}

func parseAddressList(inputs []string) []string {
	var out []string
	for _, input := range inputs {
		for _, part := range strings.Split(input, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

func normalizeHTTPRegistryAddrs(addrs []string) []string {
	normalized := make([]string, 0, len(addrs))
	seen := make(map[string]struct{}, len(addrs))
	for _, addr := range addrs {
		trimmed := strings.TrimSpace(addr)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
			logger.Warnw(
				"HTTP registry address missing scheme, defaulting to http://",
				"address",
				trimmed,
			)
			trimmed = "http://" + trimmed
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

// --- Profiling and Tracing Setup ---

func startCPUProfile() {
	f, err := os.Create("executor_cpu_profile.prof")
	if err != nil {
		logger.Errorw("Cannot create CPU profile file", "error", err)
		return
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		logger.Errorw("Cannot start CPU profile", "error", err)
		_ = f.Close() // Close the file if starting profile fails
		return
	}
	logger.Info("CPU profiling enabled, writing to executor_cpu_profile.prof")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs // Wait for shutdown signal
		logger.Info("Stopping CPU profile due to shutdown signal...")
		pprof.StopCPUProfile()
		_ = f.Close() // Close the profile file
		logger.Info("CPU profile stopped.")
	}()
}

func startTrace() func() {
	f, err := os.Create("trace.out")
	if err != nil {
		logger.Errorw("Cannot create trace file", "error", err)
		return func() {} // Return no-op stop function
	}
	err = trace.Start(f)
	if err != nil {
		logger.Errorw("Cannot start trace", "error", err)
		_ = f.Close()
		return func() {} // Return no-op stop function
	}
	logger.Info("Execution tracing enabled, writing to trace.out")
	return func() {
		logger.Info("Stopping execution trace...")
		trace.Stop()
		_ = f.Close() // Close the trace file
		logger.Info("Execution trace stopped.")
	}
}
