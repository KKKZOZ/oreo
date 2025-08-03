package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HttpRegistry wraps the existing HTTP-based registry logic
type HttpRegistry struct {
	registryAddr    string
	advertiseAddr   string
	handledDsNames  []string
	heartbeatCtx    context.Context
	heartbeatCancel context.CancelFunc
	heartbeatTicker *time.Ticker
	config          *RegistryConfig
}

// NewHttpRegistry creates a new HTTP registry instance
func NewHttpRegistry(registryAddr, advertiseAddr string, handledDsNames []string) *HttpRegistry {
	fmt.Printf("[HTTP] Connecting to registry at %s\n", registryAddr)
	ctx, cancel := context.WithCancel(context.Background())
	return &HttpRegistry{
		registryAddr:    registryAddr,
		advertiseAddr:   advertiseAddr,
		handledDsNames:  handledDsNames,
		heartbeatCtx:    ctx,
		heartbeatCancel: cancel,
		config:          DefaultRegistryConfig(),
	}
}

// Register implements ServiceRegistry interface
func (h *HttpRegistry) Register(
	ctx context.Context,
	address string,
	dsNames []string,
	metadata map[string]string,
) error {
	fmt.Printf("[HTTP] Registering service with address %s\n", address)

	if h.registryAddr == "" {
		fmt.Println("Registry address not set, skipping registration")
		return nil
	}
	if h.advertiseAddr == "" {
		return fmt.Errorf("advertise address not set, cannot register")
	}

	fmt.Printf("Attempting to register with registry: %s, advertise: %s, dsNames: %v\n",
		h.registryAddr, h.advertiseAddr, h.handledDsNames)

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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send register request to %s: %w", h.registryAddr, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(
			"registry at %s returned non-OK status for register: %s Body: %s",
			h.registryAddr,
			resp.Status,
			string(bodyBytes),
		)
	}
	fmt.Printf("Successfully registered with registry: %s\n", h.registryAddr)

	// Start heartbeat after successful registration
	h.startHeartbeat()
	return nil
}

// Deregister implements ServiceRegistry interface
func (h *HttpRegistry) Deregister(ctx context.Context, address string) error {
	fmt.Printf("[HTTP] Deregistering service from %s\n", address)

	// Stop heartbeat first
	h.stopHeartbeat()

	if h.registryAddr == "" || h.advertiseAddr == "" {
		fmt.Println("Registry or advertise address not set, skipping deregistration")
		return nil
	}

	fmt.Printf("Attempting to deregister from registry: %s, advertise: %s\n",
		h.registryAddr, h.advertiseAddr)

	reqBody := RegistryRequest{Address: h.advertiseAddr}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Failed to marshal deregister request: %v\n", err)
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		h.registryAddr+"/deregister",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		fmt.Printf("Failed to create deregister request: %v\n", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send deregister request to %s: %v\n", h.registryAddr, err)
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("Registry returned non-OK status for deregister: %s Body: %s\n",
			resp.Status, string(bodyBytes))
	}
	fmt.Printf("Successfully deregistered from registry: %s\n", h.registryAddr)
	return nil
}

// Close implements ServiceRegistry interface
func (h *HttpRegistry) Close() error {
	fmt.Println("[HTTP] Registry connection closed.")
	h.stopHeartbeat()
	return nil
}

// startHeartbeat starts the heartbeat mechanism
func (h *HttpRegistry) startHeartbeat() {
	if h.registryAddr == "" || h.advertiseAddr == "" {
		fmt.Println("Registry or advertise address not set, skipping heartbeat")
		return
	}

	fmt.Printf(
		"Starting heartbeat for %s with interval %v\n",
		h.advertiseAddr,
		h.config.HeartbeatInterval,
	)
	h.heartbeatTicker = time.NewTicker(h.config.HeartbeatInterval)

	go func() {
		reqBody := RegistryRequest{Address: h.advertiseAddr}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			fmt.Printf("CRITICAL: Failed to marshal heartbeat request: %v\n", err)
			return
		}

		for {
			select {
			case <-h.heartbeatTicker.C:
				ctx, cancel := context.WithTimeout(h.heartbeatCtx, h.config.RequestTimeout)
				req, err := http.NewRequestWithContext(
					ctx,
					http.MethodPost,
					h.registryAddr+"/heartbeat",
					bytes.NewBuffer(jsonData),
				)
				if err != nil {
					fmt.Printf("Failed to create heartbeat request: %v\n", err)
					cancel()
					continue
				}
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					// Network errors are expected sometimes, just continue
					cancel()
					continue
				}

				if resp.StatusCode != http.StatusOK {
					bodyBytes, _ := io.ReadAll(resp.Body)
					fmt.Printf("Registry returned non-OK for heartbeat: %s Body: %s\n",
						resp.Status, string(bodyBytes))
					// Try to re-register
					if regErr := h.reregister(); regErr != nil {
						fmt.Printf("Failed to re-register after failed heartbeat: %v\n", regErr)
					}
				} else {
					_, _ = io.Copy(io.Discard, resp.Body)
				}
				_ = resp.Body.Close()
				cancel()

			case <-h.heartbeatCtx.Done():
				fmt.Println("Heartbeat stopping due to context cancellation")
				return
			}
		}
	}()
}

// stopHeartbeat stops the heartbeat mechanism
func (h *HttpRegistry) stopHeartbeat() {
	if h.heartbeatTicker != nil {
		h.heartbeatTicker.Stop()
		h.heartbeatTicker = nil
	}
	if h.heartbeatCancel != nil {
		h.heartbeatCancel()
	}
	fmt.Println("Heartbeat stopped")
}

// reregister attempts to re-register with the registry
func (h *HttpRegistry) reregister() error {
	ctx, cancel := context.WithTimeout(context.Background(), h.config.RequestTimeout)
	defer cancel()
	return h.Register(ctx, h.advertiseAddr, h.handledDsNames, nil)
}
