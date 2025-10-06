package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kkkzoz/oreo/pkg/logger"
)

// HTTPServiceRegistry implements the discovery.ServiceRegistry interface.
var _ ServiceRegistry = (*HTTPServiceRegistry)(nil)

// TODO: `address` is unused in the Register and Deregister methods.
type HTTPServiceRegistry struct {
	registryAddrs  []string
	advertiseAddr  string
	handledDsNames []string
	config         *RegistryConfig

	client *http.Client

	mu              sync.Mutex // Protects all fields below this line
	isRegistered    bool
	activeAddrs     map[string]struct{}
	heartbeatCtx    context.Context
	heartbeatCancel context.CancelFunc
	heartbeatTicker *time.Ticker
}

// NewHTTPServiceRegistry creates a new HTTP registry instance.
func NewHTTPServiceRegistry(
	registryAddrs []string,
	advertiseAddr string,
	handledDsNames []string,
) *HTTPServiceRegistry {
	logger.Infow("Creating new HTTP registry",
		"registryAddrs", registryAddrs,
		"advertiseAddr", advertiseAddr,
	)
	ctx, cancel := context.WithCancel(context.Background())
	config := DefaultRegistryConfig()

	return &HTTPServiceRegistry{
		registryAddrs:  append([]string(nil), registryAddrs...),
		advertiseAddr:  advertiseAddr,
		handledDsNames: handledDsNames,
		config:         config,
		client: &http.Client{
			Timeout: config.RequestTimeout,
		},
		activeAddrs:     make(map[string]struct{}),
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

	if h.advertiseAddr == "" {
		logger.Warnw("Advertise address not set, skipping deregistration")
		return nil
	}

	targets := h.collectRegistryTargets()
	if len(targets) == 0 {
		logger.Warnw("No registry addresses configured, skipping deregistration")
		return nil
	}

	reqBody := RegistryRequest{Address: h.advertiseAddr}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal deregister request: %w", err)
	}

	var errs []error
	for _, registryAddr := range targets {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			registryAddr+"/deregister",
			bytes.NewReader(jsonData),
		)
		if err != nil {
			logger.Warnw(
				"Failed to build HTTP registry deregister request",
				"registry",
				registryAddr,
				"error",
				err,
			)
			errs = append(errs, fmt.Errorf("%s: %w", registryAddr, err))
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := h.client.Do(req)
		if err != nil {
			logger.Warnw(
				"HTTP registry deregister request failed",
				"registry",
				registryAddr,
				"error",
				err,
			)
			errs = append(errs, fmt.Errorf("%s: %w", registryAddr, err))
			continue
		}

		func() {
			defer func() {
				_ = resp.Body.Close()
			}()
			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				logger.Warnw(
					"HTTP registry rejected deregistration",
					"registry",
					registryAddr,
					"status",
					resp.Status,
					"body",
					string(bodyBytes),
				)
				errs = append(
					errs,
					fmt.Errorf(
						"%s: deregistration failed with status %s, Body: %s",
						registryAddr,
						resp.Status,
						string(bodyBytes),
					),
				)
				return
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			logger.Infow("Successfully deregistered from registry", "registry", registryAddr)
		}()
	}

	h.mu.Lock()
	h.isRegistered = false
	h.activeAddrs = make(map[string]struct{})
	h.mu.Unlock()

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

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
			// isRegistered := h.isRegistered
			allRegistered := len(h.activeAddrs) == len(h.registryAddrs)
			h.mu.Unlock()

			if !allRegistered {
				logger.Debugw("Service is not fully registered. Attempting to register...")
				ctx, cancel := context.WithTimeout(h.heartbeatCtx, h.config.RequestTimeout)
				err := h.performRegistration(ctx)
				cancel() // Release context resources

				if err == nil {
					logger.Infow("Background registration successful.")
					h.mu.Lock()
					h.isRegistered = true
					h.mu.Unlock()
				} else {
					// logger.Warnw("Background registration attempt failed, will retry.", "error", err)
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

	if len(h.registryAddrs) == 0 {
		return fmt.Errorf("no http registry addresses configured")
	}

	successes := make([]string, 0, len(h.registryAddrs))
	attempted := 0
	var errs []error

	for _, registryAddr := range h.registryAddrs {
		trimmedAddr := strings.TrimSpace(registryAddr)
		if trimmedAddr == "" {
			continue
		}
		attempted++
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			trimmedAddr+"/register",
			bytes.NewReader(jsonData),
		)
		if err != nil {
			logger.Warnw(
				"Failed to build HTTP registry register request",
				"registry",
				trimmedAddr,
				"error",
				err,
			)
			errs = append(errs, fmt.Errorf("%s: %w", trimmedAddr, err))
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := h.client.Do(req)
		if err != nil {
			// logger.Warnw(
			// 	"HTTP registry register request failed",
			// 	"registry",
			// 	trimmedAddr,
			// 	"error",
			// 	err,
			// )
			errs = append(errs, fmt.Errorf("%s: %w", trimmedAddr, err))
			continue
		}

		func() {
			defer func() {
				_ = resp.Body.Close()
			}()
			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				logger.Warnw(
					"HTTP registry rejected register request",
					"registry",
					trimmedAddr,
					"status",
					resp.Status,
					"body",
					string(bodyBytes),
				)
				errs = append(
					errs,
					fmt.Errorf(
						"%s: register returned status %s, Body: %s",
						trimmedAddr,
						resp.Status,
						string(bodyBytes),
					),
				)
				return
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			successes = append(successes, trimmedAddr)
		}()
	}

	h.mu.Lock()
	newActive := make(map[string]struct{}, len(successes))
	for _, addr := range successes {
		newActive[addr] = struct{}{}
	}
	h.activeAddrs = newActive
	h.mu.Unlock()

	if len(successes) == 0 {
		joined := errors.Join(errs...)
		if joined == nil {
			joined = fmt.Errorf("no registry addresses available")
		}
		return fmt.Errorf("failed to register with any HTTP registry: %w", joined)
	}

	if len(successes) < attempted {
		logger.Warnw(
			"Registration failed for some HTTP registries",
			"successful",
			successes,
			"failed",
			attempted-len(successes),
		)
		joined := errors.Join(errs...)
		if joined == nil {
			joined = fmt.Errorf(
				"registered successfully with %d of %d registries",
				len(successes),
				attempted,
			)
		}
		return fmt.Errorf("failed to register with all HTTP registries: %w", joined)
	}

	logger.Infow("Registration succeeded for all HTTP registries", "registries", successes)

	return nil
}

// performHeartbeat sends a single heartbeat request to the registry.
func (h *HTTPServiceRegistry) performHeartbeat(ctx context.Context) error {
	reqBody := RegistryRequest{Address: h.advertiseAddr}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat request: %w", err)
	}

	activeAddrs := h.snapshotActiveAddrs()
	if len(activeAddrs) == 0 {
		return fmt.Errorf("no active HTTP registry connections available for heartbeat")
	}

	survivors := make(map[string]struct{}, len(activeAddrs))
	var errs []error

	for _, registryAddr := range activeAddrs {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			registryAddr+"/heartbeat",
			bytes.NewReader(jsonData),
		)
		if err != nil {
			logger.Warnw(
				"Failed to build HTTP registry heartbeat request",
				"registry",
				registryAddr,
				"error",
				err,
			)
			errs = append(errs, fmt.Errorf("%s: %w", registryAddr, err))
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := h.client.Do(req)
		if err != nil {
			logger.Warnw(
				"HTTP registry heartbeat request failed",
				"registry",
				registryAddr,
				"error",
				err,
			)
			errs = append(errs, fmt.Errorf("%s: %w", registryAddr, err))
			continue
		}

		func() {
			defer func() {
				_ = resp.Body.Close()
			}()
			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				logger.Warnw(
					"HTTP registry rejected heartbeat",
					"registry",
					registryAddr,
					"status",
					resp.Status,
					"body",
					string(bodyBytes),
				)
				errs = append(
					errs,
					fmt.Errorf(
						"%s: heartbeat returned status %s, Body: %s",
						registryAddr,
						resp.Status,
						string(bodyBytes),
					),
				)
				return
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			survivors[registryAddr] = struct{}{}
		}()
	}

	h.mu.Lock()
	h.activeAddrs = survivors
	h.mu.Unlock()

	if len(survivors) == 0 {
		joined := errors.Join(errs...)
		if joined == nil {
			joined = fmt.Errorf("no heartbeat responses received")
		}
		return fmt.Errorf("heartbeat failed for all HTTP registries: %w", joined)
	}

	if len(errs) > 0 {
		logger.Warnw(
			"Heartbeat succeeded for a subset of HTTP registries",
			"healthy",
			survivors,
			"failed",
			len(errs),
		)
	}

	return nil
}

func (h *HTTPServiceRegistry) snapshotActiveAddrs() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	addrs := make([]string, 0, len(h.activeAddrs))
	for addr := range h.activeAddrs {
		addrs = append(addrs, addr)
	}
	return addrs
}

func (h *HTTPServiceRegistry) collectRegistryTargets() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	seen := make(map[string]struct{}, len(h.registryAddrs)+len(h.activeAddrs))
	targets := make([]string, 0, len(seen))

	for _, addr := range h.registryAddrs {
		trimmed := strings.TrimSpace(addr)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		targets = append(targets, trimmed)
	}
	for addr := range h.activeAddrs {
		trimmed := strings.TrimSpace(addr)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		targets = append(targets, trimmed)
	}
	return targets
}
