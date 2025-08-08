package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/kkkzoz/oreo/pkg/timesource"
	"go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	port         int
	oracleType   string
	etcdEndpoints string
	electionName string
	maxSkewStr   string
	Log          *zap.SugaredLogger

	isActive    bool = false
	activeMutex sync.RWMutex

	globalOracle timesource.TimeSourcer
)

func handleTimestamp(w http.ResponseWriter, r *http.Request) {
	activeMutex.RLock()
	currentIsActive := isActive
	activeMutex.RUnlock()

	if !currentIsActive {
		Log.Warnw("Received timestamp request while inactive", "remoteAddr", r.RemoteAddr)
		http.Error(w, "Service not active", http.StatusServiceUnavailable)
		return
	}

	startTime := time.Now()
	timestamp, err := globalOracle.GetTime("pattern")
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

func handleHealth(w http.ResponseWriter, r *http.Request) {
	activeMutex.RLock()
	currentIsActive := isActive
	activeMutex.RUnlock()

	if currentIsActive {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("Service not active"))
	}
}

func main() {
	flag.IntVar(&port, "p", 8010, "HTTP server port number")
	flag.StringVar(&oracleType, "type", "hybrid", "Time Oracle Implementation Type (hybrid, simple, counter)")
	flag.StringVar(&etcdEndpoints, "etcd-endpoints", "localhost:2379", "Comma-separated etcd endpoints")
	flag.StringVar(&electionName, "election-name", "/timeoracle-leader", "etcd election name")
	flag.StringVar(&maxSkewStr, "max-skew", "50ms", "Maximum clock skew delta (e.g., 50ms, 100ms)")
	flag.Parse()

	newLogger()
	Log.Infow("Starting Time Oracle Node", "version", "1.0", "pid", os.Getpid())

	_, err := time.ParseDuration(maxSkewStr)
	if err != nil {
		Log.Fatalw("Invalid max-skew duration", "value", maxSkewStr, "error", err)
	}

	switch oracleType {
	case "hybrid":
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

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/timestamp/", handleTimestamp)

	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdEndpoints},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		Log.Fatalw("Failed to connect to etcd", "error", err)
	}
	defer etcdClient.Close()

	nodeID := uuid.New().String()
	leaseManager, err := NewLeaseManager(etcdClient, electionName, log.New(os.Stdout, "", log.LstdFlags))
	if err != nil {
		Log.Fatalw("Failed to create lease manager", "error", err)
	}
	defer leaseManager.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := leaseManager.Campaign(ctx, nodeID); err != nil {
			Log.Errorw("Failed during leader election campaign", "error", err)
			return
		}
		setActive(true)
	}()

	leaseManager.WatchLeader(ctx, func(leader string) {
		Log.Infow("Leader changed", "newLeader", leader)
		if leader != nodeID {
			setActive(false)
		}
	})

	serverAddress := fmt.Sprintf(":%d", port)
	httpServer := &http.Server{
		Addr:    serverAddress,
		Handler: mux,
	}

	go func() {
		Log.Infof("HTTP server listening on %s", serverAddress)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Log.Fatalw("HTTP server failed", "error", err)
		}
	}()

	// Graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	Log.Infow("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		Log.Warnw("Server shutdown failed", "error", err)
	}

	if err := leaseManager.Resign(context.Background()); err != nil {
		Log.Warnw("Failed to resign leadership", "error", err)
	}

	Log.Info("Server gracefully stopped")
}

func setActive(active bool) {
	activeMutex.Lock()
	defer activeMutex.Unlock()
	isActive = active
	Log.Infow("Node active state changed", "isActive", active)
}

func newLogger() {
	conf := zap.NewDevelopmentConfig()
	logLevelEnv := os.Getenv("LOG_LEVEL")
	level := zapcore.InfoLevel
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
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	conf.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	conf.EncoderConfig.MessageKey = "message"
	conf.EncoderConfig.TimeKey = "timestamp"
	conf.EncoderConfig.CallerKey = "caller"
	logger, err := conf.Build(zap.AddCallerSkip(1))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	Log = logger.Sugar()
}