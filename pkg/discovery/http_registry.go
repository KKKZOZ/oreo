package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/kkkzoz/oreo/pkg/logger"
)

// HTTPServiceRegistry implements the discovery.ServiceRegistry interface.
var _ ServiceRegistry = (*HTTPServiceRegistry)(nil)

// TODO: `address` is unused in the Register and Deregister methods.
type HTTPServiceRegistry struct {
	registryAddr   string
	advertiseAddr  string
	handledDsNames []string
	config         *RegistryConfig

	client *http.Client

	mu              sync.Mutex // Protects all fields below this line
	isRegistered    bool
	heartbeatCtx    context.Context
	heartbeatCancel context.CancelFunc
	heartbeatTicker *time.Ticker
}

// NewHTTPServiceRegistry creates a new HTTP registry instance.
func NewHTTPServiceRegistry(
	registryAddr, advertiseAddr string,
	handledDsNames []string,
) *HTTPServiceRegistry {
	logger.Infow("Creating new HTTP registry",
		"registryAddr", registryAddr,
		"advertiseAddr", advertiseAddr,
	)
	ctx, cancel := context.WithCancel(context.Background())
	config := DefaultRegistryConfig()

	return &HTTPServiceRegistry{
		registryAddr:   registryAddr,
		advertiseAddr:  advertiseAddr,
		handledDsNames: handledDsNames,
		config:         config,
		client: &http.Client{
			Timeout: config.RequestTimeout,
		},
		heartbeatCtx:    ctx,
		heartbeatCancel: cancel,
	}
}

// Register registers the service with the central registry.
// It will start a background loop to maintain the registration (via heartbeats or retries).
// Per request, the API signature is kept the same; the `address`, `dsNames`, and `metadata`
// parameters are unused in this implementation.
func (h *HTTPServiceRegistry) Register(
	ctx context.Context,
	address string,
	dsNames []string,
	metadata map[string]string,
) error {
	logger.Infow("Registering service...",
		"advertiseAddr", h.advertiseAddr,
		// Note: The parameters are ignored, using values from the constructor instead.
		"ignored_address_param", address,
		"ignored_dsNames_param", dsNames,
	)

	// Always start the background management loop. This allows retrying if the
	// discovery service is not yet available.
	h.startManagementLoop()

	// Make a single, initial attempt to register immediately.
	// The background loop will handle retries if this fails.
	err := h.performRegistration(ctx)
	if err != nil {
		logger.Warnw(
			"Initial registration attempt failed, will retry in background",
			"error",
			err,
		)
		return err // Return error to inform the caller, but the retry process is now running.
	}

	// If the first attempt is successful, update the state immediately.
	h.mu.Lock()
	h.isRegistered = true
	h.mu.Unlock()
	logger.Infow("Initial registration successful")

	return nil
}

// Deregister deregisters the service from the central registry.
// Per request, the API signature is kept the same; the `address` parameter is unused.
func (h *HTTPServiceRegistry) Deregister(ctx context.Context, address string) error {
	logger.Infow("Deregistering service", "address", h.advertiseAddr)
	// Stop the background loop first to prevent further heartbeats or registration attempts.
	h.stopManagementLoop()

	if h.registryAddr == "" || h.advertiseAddr == "" {
		logger.Warnw("Registry or advertise address not set, skipping deregistration")
		return nil
	}

	// Set state to unregistered.
	h.mu.Lock()
	h.isRegistered = false
	h.mu.Unlock()

	reqBody := RegistryRequest{Address: h.advertiseAddr}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal deregister request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		h.registryAddr+"/deregister",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create deregister request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send deregister request to %s: %w", h.registryAddr, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf(
			"deregistration failed with status %s, Body: %s",
			resp.Status,
			string(bodyBytes),
		)
		logger.Warnw("Deregistration failed", "error", err)
		return err
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	logger.Infow("Successfully deregistered from registry", "registry", h.registryAddr)
	return nil
}

// Close stops the background management loop.
func (h *HTTPServiceRegistry) Close() error {
	logger.Infow("Closing registry connection")
	h.stopManagementLoop()
	return nil
}

// startManagementLoop starts the background loop for registration and heartbeats.
func (h *HTTPServiceRegistry) startManagementLoop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.heartbeatTicker != nil {
		logger.Debugw("Management loop already running, skipping start")
		return
	}

	logger.Infow("Starting background management loop", "interval", h.config.HeartbeatInterval)
	h.heartbeatTicker = time.NewTicker(h.config.HeartbeatInterval)
	go h.manageConnectionLoop()
}

// stopManagementLoop stops the background loop.
func (h *HTTPServiceRegistry) stopManagementLoop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.heartbeatTicker != nil {
		h.heartbeatTicker.Stop()
		h.heartbeatTicker = nil
		logger.Infow("Management loop stopped")
	}
	if h.heartbeatCancel != nil {
		h.heartbeatCancel()
	}
}

// manageConnectionLoop is the core loop that handles the service's registration state.
// It will periodically try to register if not already registered, or send a heartbeat if it is.
func (h *HTTPServiceRegistry) manageConnectionLoop() {
	for {
		select {
		case <-h.heartbeatTicker.C:
			h.mu.Lock()
			isRegistered := h.isRegistered
			h.mu.Unlock()

			if !isRegistered {
				logger.Infow("Service is not registered. Attempting to register...")
				ctx, cancel := context.WithTimeout(h.heartbeatCtx, h.config.RequestTimeout)
				err := h.performRegistration(ctx)
				cancel() // Release context resources

				if err == nil {
					logger.Infow("Background registration successful.")
					h.mu.Lock()
					h.isRegistered = true
					h.mu.Unlock()
				} else {
					logger.Warnw("Background registration attempt failed, will retry.", "error", err)
				}
			} else {
				logger.Debugw("Service is registered. Sending heartbeat...")
				ctx, cancel := context.WithTimeout(h.heartbeatCtx, h.config.RequestTimeout)
				err := h.performHeartbeat(ctx)
				cancel() // Release context resources

				if err != nil {
					logger.Warnw("Heartbeat failed. Marking as unregistered to trigger re-registration.", "error", err)
					h.mu.Lock()
					h.isRegistered = false
					h.mu.Unlock()
				} else {
					logger.Debugw("Heartbeat successful.")
				}
			}
		case <-h.heartbeatCtx.Done():
			logger.Infow("Management loop stopping due to context cancellation.")
			return
		}
	}
}

// performRegistration sends a single registration request.
func (h *HTTPServiceRegistry) performRegistration(ctx context.Context) error {
	reqBody := RegistryRequest{
		Address: h.advertiseAddr,
		DsNames: h.handledDsNames,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal register request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		h.registryAddr+"/register",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create register request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send register request to %s: %w", h.registryAddr, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry returned non-OK status for register: %s, Body: %s",
			resp.Status, string(bodyBytes))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// performHeartbeat sends a single heartbeat request to the registry.
func (h *HTTPServiceRegistry) performHeartbeat(ctx context.Context) error {
	reqBody := RegistryRequest{Address: h.advertiseAddr}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		h.registryAddr+"/heartbeat",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registry returned non-OK status for heartbeat: %s, Body: %s",
			resp.Status, string(bodyBytes))
	}

	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
