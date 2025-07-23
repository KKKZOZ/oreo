package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	// Assuming timesource is in the correct relative path or GOPATH
	// Adjust the import path if necessary, e.g., "your_module/pkg/timesource"
	"github.com/kkkzoz/oreo/pkg/timesource"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	port                   int
	oracleType             string
	role                   string
	primaryAddr            string
	maxSkewStr             string
	healthCheckIntervalStr string
	healthCheckTimeoutStr  string
	failureThreshold       int
	Log                    *zap.SugaredLogger

	// Channel to signal when the backup should become active
	becomeActive chan struct{} = make(chan struct{})
	// Flag to indicate if the server is currently active (serving timestamps)
	isActive bool = false
	// Mutex to protect access to isActive flag
	activeMutex sync.RWMutex

	// Store the oracle instance globally for handlers
	globalOracle timesource.TimeSourcer
)

// handleTimestamp serves timestamps only if the instance is active.
func handleTimestamp(w http.ResponseWriter, r *http.Request) {
	activeMutex.RLock()
	currentIsActive := isActive
	activeMutex.RUnlock()

	if !currentIsActive {
		Log.Warnw("Received timestamp request while inactive", "remoteAddr", r.RemoteAddr)
		http.Error(w, "Service not active (backup node?)", http.StatusServiceUnavailable) // 503
		return
	}

	startTime := time.Now()
	// Use the globally initialized oracle
	timestamp, err := globalOracle.GetTime(
		"pattern",
	) // "pattern" might need adjustment based on your timesource usage
	if err != nil {
		Log.Errorw("Failed to get time from oracle", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	latency := time.Since(startTime).Microseconds()
	Log.Debugw(
		"handleTimestamp",
		"LatencyInFunction",
		latency,
		"Topic",
		"CheckPoint",
		"Timestamp",
		timestamp,
	)

	w.Header().Set("Content-Type", "text/plain")
	_, writeErr := fmt.Fprintf(w, "%d", timestamp)
	if writeErr != nil {
		Log.Errorw("Failed to write timestamp response", "error", writeErr)
	}
}

// handleHealth reports status based on the isActive flag. Crucial for HAProxy.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	activeMutex.RLock()
	currentIsActive := isActive
	activeMutex.RUnlock()

	if currentIsActive {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK) // 200 OK - Ready
		_, _ = w.Write([]byte("OK"))
	} else {
		w.Header().Set("Content-Type", "text/plain")
		// 503 Service Unavailable tells HAProxy this node is not ready
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("Service not active (backup node?)"))
	}
}

func main() {
	// --- Flag Definitions ---
	flag.IntVar(&port, "p", 8010, "HTTP server port number")
	flag.StringVar(
		&oracleType,
		"type",
		"hybrid",
		"Time Oracle Implementation Type (hybrid, simple, counter)",
	)
	flag.StringVar(&role, "role", "primary", "Node role: primary or backup")
	flag.StringVar(
		&primaryAddr,
		"primary-addr",
		"",
		"HTTP address of the primary node (e.g., http://192.168.1.100:8010) (required for backup role)",
	)
	flag.StringVar(&maxSkewStr, "max-skew", "50ms", "Maximum clock skew delta (e.g., 50ms, 100ms)")
	flag.StringVar(
		&healthCheckIntervalStr,
		"health-check-interval",
		"2s",
		"Interval for backup health checks (e.g., 1s, 2s)",
	)
	flag.StringVar(
		&healthCheckTimeoutStr,
		"health-check-timeout",
		"1s",
		"Timeout for health check request (e.g., 500ms, 1s)",
	)
	flag.IntVar(
		&failureThreshold,
		"failure-threshold",
		3,
		"Consecutive failures to declare primary down",
	)
	flag.Parse()

	// --- Logger Setup ---
	newLogger()
	Log.Infow("Starting Time Oracle Node", "version", "1.0", "pid", os.Getpid()) // Example version

	// --- Parse Durations ---
	maxSkew, err := time.ParseDuration(maxSkewStr)
	if err != nil {
		Log.Fatalw("Invalid max-skew duration", "value", maxSkewStr, "error", err)
	}
	healthCheckInterval, err := time.ParseDuration(healthCheckIntervalStr)
	if err != nil {
		Log.Fatalw(
			"Invalid health-check-interval duration",
			"value",
			healthCheckIntervalStr,
			"error",
			err,
		)
	}
	healthCheckTimeout, err := time.ParseDuration(healthCheckTimeoutStr)
	if err != nil {
		Log.Fatalw(
			"Invalid health-check-timeout duration",
			"value",
			healthCheckTimeoutStr,
			"error",
			err,
		)
	}
	Log.Infow(
		"Configuration",
		"port",
		port,
		"oracleType",
		oracleType,
		"role",
		role,
		"primaryAddr",
		primaryAddr,
		"maxSkew",
		maxSkew,
		"healthCheckInterval",
		healthCheckInterval,
		"healthCheckTimeout",
		healthCheckTimeout,
		"failureThreshold",
		failureThreshold,
	)

	// --- Initialize Oracle ---
	switch oracleType {
	case "hybrid":
		// TODO: Make HybridTimeSource params configurable if needed
		globalOracle = timesource.NewHybridTimeSource(10, 6)
		Log.Info("Using Hybrid TimeSource")
	case "simple":
		globalOracle = timesource.NewSimpleTimeSource()
		Log.Info("Using Simple TimeSource")
	case "counter":
		globalOracle = timesource.NewCounterTimeSource()
		Log.Info("Using Counter TimeSource")
	default:
		Log.Fatalw("Invalid oracle type specified", "type", oracleType)
	}

	// --- Role-Specific Logic ---
	mux := http.NewServeMux() // Use a mux for clarity
	mux.HandleFunc("/health", handleHealth)

	switch role {
	case "primary":
		Log.Info("Starting as PRIMARY node.")
		setActive(true)                                // Primary starts active
		mux.HandleFunc("/timestamp/", handleTimestamp) // Register timestamp handler immediately
	case "backup":
		Log.Info("Starting as BACKUP node.")
		if primaryAddr == "" {
			Log.Fatal("Backup role requires --primary-addr flag to be set")
		}
		setActive(false) // Backup starts inactive

		// Start monitoring in the background
		go monitorPrimary(
			primaryAddr,
			maxSkew,
			healthCheckInterval,
			healthCheckTimeout,
			failureThreshold,
		)

		// Register timestamp handler, but it will return 503 until isActive is true
		mux.HandleFunc("/timestamp/", handleTimestamp)

		// Backup: Wait until signaled to become active
		Log.Info("Waiting for signal to become active...")
		<-becomeActive // Block here
		Log.Info("Received signal. Promoting to ACTIVE.")
		setActive(true)
		Log.Info("Backup node is now ACTIVE and serving timestamps.")

	default:
		Log.Fatalw("Invalid role specified. Use 'primary' or 'backup'.", "role", role)
	}

	// --- Start HTTP Server ---
	serverAddress := fmt.Sprintf(":%d", port)
	Log.Infof("HTTP server listening on %s", serverAddress)

	// Create the server explicitly to set timeouts potentially
	httpServer := &http.Server{
		Addr:    serverAddress,
		Handler: mux,
		// Add ReadTimeout, WriteTimeout etc. for production robustness
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		// IdleTimeout:  120 * time.Second,
	}

	err = httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		Log.Fatalw("HTTP server failed", "error", err)
	}
	Log.Info("HTTP server shut down.")
}

// setActive safely updates the isActive flag.
func setActive(active bool) {
	activeMutex.Lock()
	defer activeMutex.Unlock()
	isActive = active
	Log.Infow("Node active state changed", "isActive", active)
}

// monitorPrimary runs on the backup node to check the primary's health.
func monitorPrimary(pAddr string, maxSkew, interval, timeout time.Duration, threshold int) {
	Log.Infow(
		"Starting primary monitoring routine",
		"primary",
		pAddr,
		"checkInterval",
		interval,
		"checkTimeout",
		timeout,
		"failureThreshold",
		threshold,
	)
	client := http.Client{
		Timeout: timeout,
	}
	// Ensure the primary address includes the scheme (http://)
	healthURL := pAddr + "/health" // Assume primary exposes /health endpoint
	Log.Infof("Monitoring primary health at: %s", healthURL)

	consecutiveFailures := 0
	waitDuration := 2 * maxSkew

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		resp, err := client.Get(healthURL)
		if err != nil {
			consecutiveFailures++
			Log.Warnw(
				"Health check failed (network/timeout)",
				"target",
				healthURL,
				"error",
				err,
				"failures",
				consecutiveFailures,
				"threshold",
				threshold,
			)
		} else {
			// Best practice: always read and close body
			// _, _ = io.Copy(io.Discard, resp.Body) // Read body fully
			_ = resp.Body.Close() // Close body

			if resp.StatusCode == http.StatusOK {
				if consecutiveFailures > 0 {
					Log.Info("Health check succeeded after previous failures. Resetting failure count.")
				}
				consecutiveFailures = 0 // Reset on success
			} else {
				consecutiveFailures++
				Log.Warnw("Health check failed (non-200 status)", "target", healthURL, "status", resp.StatusCode, "failures", consecutiveFailures, "threshold", threshold)
			}
		}

		if consecutiveFailures >= threshold {
			Log.Warnw("Primary health check failure threshold reached.", "threshold", threshold)
			Log.Infof(
				"Primary node declared DOWN. Waiting for safety period: %v (2 * maxSkew)",
				waitDuration,
			)

			// --- THE CRITICAL WAIT ---
			time.Sleep(waitDuration)
			// --- END CRITICAL WAIT ---

			Log.Info("Safety wait period finished. Attempting to take over.")
			close(becomeActive) // Signal main goroutine to become active
			Log.Info("Monitoring routine finished. Signaled main goroutine.")
			return // Stop monitoring
		}
		// Potentially add a quit channel here if graceful shutdown is needed
	}
}

// newLogger initializes the Zap sugared logger.
func newLogger() {
	conf := zap.NewDevelopmentConfig() // Provides human-readable output

	// Allow log level override via environment variable
	logLevelEnv := os.Getenv("LOG_LEVEL")
	level := zapcore.InfoLevel // Default level
	switch logLevelEnv {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	case "FATAL":
		level = zapcore.FatalLevel
	default:
		if logLevelEnv != "" {
			fmt.Fprintf(os.Stderr, "Invalid LOG_LEVEL '%s'. Defaulting to INFO.\n", logLevelEnv)
		}
	}
	conf.Level = zap.NewAtomicLevelAt(level)

	// Customize encoder settings
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Colorized level
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Standard time format
	conf.EncoderConfig.MessageKey = "message"                         // Rename 'msg' key
	conf.EncoderConfig.TimeKey = "timestamp"                          // Rename 'ts' key
	conf.EncoderConfig.CallerKey = "caller"

	// Build the logger
	logger, err := conf.Build(
		zap.AddCallerSkip(1),
	) // AddCallerSkip helps show correct caller file/line
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	Log = logger.Sugar()
}
