package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	stdlog "log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kkkzoz/oreo/pkg/logger"
)

// HTTPServiceDiscovery implements the discovery.ServiceDiscovery interface.
var _ ServiceDiscovery = (*HTTPServiceDiscovery)(nil)

// HTTPServiceDiscovery implements ServiceDiscovery interface for HTTP-based service discovery
type HTTPServiceDiscovery struct {
	registryMutex    sync.RWMutex
	instances        map[string]InstanceInfo
	dsNameIndex      map[string][]string
	instanceTTL      time.Duration
	registryListener *http.Server
	shutdownCtx      context.Context
	shutdownCancel   context.CancelFunc
	wg               sync.WaitGroup

	curIndexMap map[string]*atomic.Uint64
}

// InstanceInfo holds info about a registered executor instance within the registry.
type InstanceInfo struct {
	Address       string    // The advertised network address (e.g., "1.2.3.4:8000")
	LastHeartbeat time.Time // Timestamp of the last successful heartbeat
	DsNames       []string  // List of datastore names this instance handles (e.g., ["Redis", "MongoDB1"])
}

// Constants
const (
	ALL = "ALL" // Special DsName indicating an instance handles all datastores
	// Default TTL if not provided (should be longer than executor's heartbeatInterval)
	defaultInstanceTTL = 6 * time.Second
	// How often the registry checks for stale instances
	cleanupInterval = 3 * time.Second
)

// NewHTTPServiceDiscovery creates a new HTTP-based service discovery instance
func NewHTTPServiceDiscovery(
	registryListenAddr string,
	instanceTTL ...time.Duration,
) (*HTTPServiceDiscovery, error) {
	ttl := defaultInstanceTTL
	if len(instanceTTL) > 0 {
		ttl = instanceTTL[0]
	}

	ctx, cancel := context.WithCancel(context.Background())

	hsd := &HTTPServiceDiscovery{
		instances:      make(map[string]InstanceInfo),
		dsNameIndex:    make(map[string][]string),
		instanceTTL:    ttl,
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
		curIndexMap:    make(map[string]*atomic.Uint64),
	}

	// Set up HTTP server for registry endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/register", hsd.handleRegister)
	mux.HandleFunc("/deregister", hsd.handleDeregister)
	mux.HandleFunc("/heartbeat", hsd.handleHeartbeat)
	mux.HandleFunc("/services", hsd.handleGetServices)

	hsd.registryListener = &http.Server{
		Addr:    registryListenAddr,
		Handler: mux,
	}

	// Start the registry server
	hsd.wg.Add(1)
	go func() {
		defer hsd.wg.Done()
		if err := hsd.registryListener.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			stdlog.Printf("Registry server error: %v", err)
		}
	}()

	// Start cleanup routine
	hsd.wg.Add(1)
	go func() {
		defer hsd.wg.Done()
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				hsd.cleanupStaleInstances()
			case <-hsd.shutdownCtx.Done():
				return
			}
		}
	}()

	return hsd, nil
}

// GetService returns a service instance address for the given datastore name
func (hsd *HTTPServiceDiscovery) GetService(dsName string) (string, error) {
	hsd.registryMutex.RLock()
	defer hsd.registryMutex.RUnlock()

	// Try specific dsName first (case-insensitive), then fallback to ALL
	for _, name := range []string{dsName, ALL} {
		// For case-insensitive matching, check all keys
		for indexKey, instances := range hsd.dsNameIndex {
			if strings.EqualFold(indexKey, name) && len(instances) > 0 {
				index := hsd.curIndexMap[indexKey].Add(1) - 1
				address := instances[index%uint64(len(instances))]
				if address != "" {
					return address, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no available instances for datastore: %s", dsName)
}

// Close closes the HTTP service discovery and cleans up resources
func (hsd *HTTPServiceDiscovery) Close() error {
	logger.Info("Shutting down HTTP service discovery...")

	// Cancel the shutdown context to stop background goroutines
	hsd.shutdownCancel()

	// Shutdown the HTTP server with a timeout
	if hsd.registryListener != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := hsd.registryListener.Shutdown(shutdownCtx); err != nil {
			logger.Errorf("Error shutting down registry server: %v", err)
			// Don't return immediately, still wait for other goroutines
		}
	}

	// Wait for all background goroutines to finish
	hsd.wg.Wait()

	logger.Info("HTTP service discovery shutdown complete")
	return nil
}

// HTTP handlers
func (hsd *HTTPServiceDiscovery) handleRegister(w http.ResponseWriter, r *http.Request) {
	req, err := hsd.validatePostRequest(w, r)
	if err != nil {
		return
	}

	hsd.registryMutex.Lock()
	defer hsd.registryMutex.Unlock()

	// Check if this is a new instance or an update
	existingInstance, exists := hsd.instances[req.Address]
	isNewInstance := !exists
	if exists {
		// Remove from old dsName index
		for _, dsName := range existingInstance.DsNames {
			hsd.removeInstanceFromDsName(req.Address, dsName)
		}
	}

	// Add/update the instance
	hsd.instances[req.Address] = InstanceInfo{
		Address:       req.Address,
		LastHeartbeat: time.Now(),
		DsNames:       req.DsNames,
	}

	// Add to dsName index and log only for Redis services
	for _, dsName := range req.DsNames {
		if hsd.dsNameIndex[dsName] == nil {
			hsd.dsNameIndex[dsName] = make([]string, 0)
			hsd.curIndexMap[dsName] = &atomic.Uint64{}
		}
		hsd.dsNameIndex[dsName] = append(hsd.dsNameIndex[dsName], req.Address)

		// Only log for Redis services
		if strings.ToLower(dsName) == "redis" && isNewInstance {
			logger.Infof("Redis service online: %s", req.Address)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (hsd *HTTPServiceDiscovery) handleDeregister(w http.ResponseWriter, r *http.Request) {
	req, err := hsd.validatePostRequest(w, r)
	if err != nil {
		return
	}

	hsd.registryMutex.Lock()
	defer hsd.registryMutex.Unlock()

	// Service deregistration without logging
	hsd.removeInstanceLocked(req.Address, "deregister")
	w.WriteHeader(http.StatusOK)
}

func (hsd *HTTPServiceDiscovery) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	req, err := hsd.validatePostRequest(w, r)
	if err != nil {
		return
	}

	hsd.registryMutex.Lock()
	defer hsd.registryMutex.Unlock()

	if instance, exists := hsd.instances[req.Address]; exists {
		instance.LastHeartbeat = time.Now()
		hsd.instances[req.Address] = instance
		w.WriteHeader(http.StatusOK)
		return
	}

	// If instance doesn't exist, force re-registration
	w.WriteHeader(http.StatusNotFound)
}

func (hsd *HTTPServiceDiscovery) handleGetServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if dsName query parameter is provided
	dsName := r.URL.Query().Get("dsName")
	if dsName != "" {
		// Return a single service address for the specified dsName
		hsd.registryMutex.RLock()
		var address string
		var found bool
		// Try specific dsName first (case-insensitive), then fallback to ALL
		for _, name := range []string{dsName, ALL} {
			// For case-insensitive matching, check all keys
			for indexKey, instances := range hsd.dsNameIndex {
				if strings.EqualFold(indexKey, name) && len(instances) > 0 {
					index := hsd.curIndexMap[indexKey].Add(1) - 1
					address = instances[index%uint64(len(instances))]
					if address != "" {
						found = true
						break
					}
				}
			}
			if found {
				break
			}
		}
		hsd.registryMutex.RUnlock()
		if !found {
			http.Error(w, "Service not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(address)); err != nil {
			logger.Warnf("Failed to write response: %v", err)
		}
		return
	}

	// Return all services as JSON (original behavior)
	hsd.registryMutex.RLock()
	response := make(map[string]InstanceInfo, len(hsd.instances))
	for addr, instance := range hsd.instances {
		response[addr] = instance
	}
	hsd.registryMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Helper methods
// validatePostRequest validates POST requests and decodes JSON payload
func (hsd *HTTPServiceDiscovery) validatePostRequest(
	w http.ResponseWriter,
	r *http.Request,
) (*RegistryRequest, error) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return nil, fmt.Errorf("method not allowed")
	}

	var req RegistryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	if req.Address == "" {
		http.Error(w, "Address is required", http.StatusBadRequest)
		return nil, fmt.Errorf("address is required")
	}

	return &req, nil
}

// removeInstanceLocked removes an instance from both instances map and dsName index
func (hsd *HTTPServiceDiscovery) removeInstanceLocked(instanceAddr, reason string) {
	instance, exists := hsd.instances[instanceAddr]
	if !exists {
		return
	}

	// Check if this is a Redis service for logging
	isRedisService := false
	for _, dsName := range instance.DsNames {
		if strings.ToLower(dsName) == "redis" {
			isRedisService = true
			break
		}
	}

	// Remove from dsName index
	for _, dsName := range instance.DsNames {
		hsd.removeInstanceFromDsName(instanceAddr, dsName)
	}
	delete(hsd.instances, instanceAddr)

	// Only log for Redis services
	if isRedisService {
		if reason == "stale" {
			logger.Warnf("Redis service timeout: %s (last heartbeat: %v)",
				instanceAddr, instance.LastHeartbeat.Format("15:04:05"))
		} else {
			logger.Infof("Redis service offline: %s", instanceAddr)
		}
	}
}

// removeInstanceFromDsName removes an instance from a specific dsName index
func (hsd *HTTPServiceDiscovery) removeInstanceFromDsName(instanceAddr, dsName string) {
	instances, exists := hsd.dsNameIndex[dsName]
	if !exists {
		return
	}

	for i, addr := range instances {
		if addr == instanceAddr {
			hsd.dsNameIndex[dsName] = append(instances[:i], instances[i+1:]...)
			break
		}
	}

	// Clean up empty dsName index
	if len(hsd.dsNameIndex[dsName]) == 0 {
		delete(hsd.dsNameIndex, dsName)
		delete(hsd.curIndexMap, dsName)
	}
}

func (hsd *HTTPServiceDiscovery) cleanupStaleInstances() {
	staleThreshold := time.Now().Add(-hsd.instanceTTL)

	hsd.registryMutex.Lock()
	defer hsd.registryMutex.Unlock()

	// Find stale instances
	var staleInstances []string
	for addr, instance := range hsd.instances {
		if instance.LastHeartbeat.Before(staleThreshold) {
			staleInstances = append(staleInstances, addr)
		}
	}

	if len(staleInstances) == 0 {
		return
	}

	// Remove stale instances (logging is handled in removeInstanceLocked)
	for _, addr := range staleInstances {
		hsd.removeInstanceLocked(addr, "stale")
	}
}
