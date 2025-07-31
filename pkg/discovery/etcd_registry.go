package discovery

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdServiceRegistry etcd-based service registration implementation
type EtcdServiceRegistry struct {
	client           *clientv3.Client
	config           RegistryConfig
	keyPrefix        string // Key prefix for service registration in etcd
	leaseID          clientv3.LeaseID
	leaseGranted     bool
	keepAliveRunning bool
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewEtcdServiceRegistry creates a new etcd service registry
func NewEtcdServiceRegistry(
	endpoints []string,
	keyPrefix string,
	config *RegistryConfig,
) (*EtcdServiceRegistry, error) {
	if config == nil {
		config = DefaultRegistryConfig()
	}

	if keyPrefix == "" {
		keyPrefix = "/oreo/services"
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &EtcdServiceRegistry{
		client:    client,
		config:    *config,
		keyPrefix: keyPrefix,
	}, nil
}

// Register registers a service instance
func (e *EtcdServiceRegistry) Register(
	ctx context.Context,
	address string,
	dsNames []string,
	metadata map[string]string,
) error {
	// Create lease
	if !e.leaseGranted {
		leaseResp, err := e.client.Grant(ctx, int64(e.config.InstanceTTL.Seconds()))
		if err != nil {
			return fmt.Errorf("failed to grant lease: %w", err)
		}
		e.leaseID = leaseResp.ID
		e.leaseGranted = true
	}

	// Generate unique ID for the service instance
	serviceID, err := generateUUID()
	if err != nil {
		return fmt.Errorf("failed to generate service ID: %w", err)
	}

	// Construct service information
	serviceInfo := ServiceInfo{
		ID:            serviceID,
		Address:       address,
		DsNames:       dsNames,
		LastHeartbeat: time.Now(),
		Metadata:      metadata,
	}

	serviceData, err := json.Marshal(serviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	// Register service to etcd
	key := e.getServiceKey(address)
	_, err = e.client.Put(ctx, key, string(serviceData), clientv3.WithLease(e.leaseID))
	if err != nil {
		return fmt.Errorf("failed to register service in etcd: %w", err)
	}

	// Start lease renewal
	if !e.keepAliveRunning {
		e.keepAliveRunning = true
		go func() {
			e.keepAlive()
			e.keepAliveRunning = false
		}()
	}

	return nil
}

// Deregister deregisters a service instance
func (e *EtcdServiceRegistry) Deregister(ctx context.Context, address string) error {
	key := e.getServiceKey(address)
	_, err := e.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to deregister service from etcd: %w", err)
	}

	// Revoke lease
	if e.leaseGranted {
		_, _ = e.client.Revoke(ctx, e.leaseID)
		e.leaseGranted = false
	}

	return nil
}

// Heartbeat sends heartbeat (implemented through lease renewal)
func (e *EtcdServiceRegistry) Heartbeat(ctx context.Context, address string) error {
	if !e.leaseGranted {
		return fmt.Errorf("lease not granted, cannot send heartbeat")
	}

	// etcd heartbeat is automatically handled through lease keepalive
	// Here we can choose to update LastHeartbeat time
	key := e.getServiceKey(address)
	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get service info for heartbeat: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return fmt.Errorf("service not found in etcd")
	}

	var serviceInfo ServiceInfo
	if err := json.Unmarshal(resp.Kvs[0].Value, &serviceInfo); err != nil {
		return fmt.Errorf("failed to unmarshal service info: %w", err)
	}

	// Update heartbeat time
	serviceInfo.LastHeartbeat = time.Now()
	serviceData, err := json.Marshal(serviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal updated service info: %w", err)
	}

	_, err = e.client.Put(ctx, key, string(serviceData), clientv3.WithLease(e.leaseID))
	if err != nil {
		return fmt.Errorf("failed to update heartbeat in etcd: %w", err)
	}

	return nil
}

// GetServices gets services for the specified dsName
func (e *EtcdServiceRegistry) GetServices(
	ctx context.Context,
	dsName string,
) ([]ServiceInfo, error) {
	resp, err := e.client.Get(ctx, e.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get services from etcd: %w", err)
	}

	var services []ServiceInfo
	for _, kv := range resp.Kvs {
		var serviceInfo ServiceInfo
		if err := json.Unmarshal(kv.Value, &serviceInfo); err != nil {
			continue // Skip services that cannot be parsed
		}

		// Check if it matches the specified dsName
		if dsName == "" {
			services = append(services, serviceInfo)
		} else {
			for _, serviceDsName := range serviceInfo.DsNames {
				if serviceDsName == dsName || serviceDsName == "ALL" {
					services = append(services, serviceInfo)
					break
				}
			}
		}
	}

	return services, nil
}

// GetAllServices gets all services
func (e *EtcdServiceRegistry) GetAllServices(ctx context.Context) ([]ServiceInfo, error) {
	return e.GetServices(ctx, "")
}

// Watch watches for service changes
func (e *EtcdServiceRegistry) Watch(
	ctx context.Context,
	dsName string,
) (<-chan ServiceChangeEvent, error) {
	ch := make(chan ServiceChangeEvent, 1)

	// First send the current service list
	go func() {
		defer close(ch)

		// Send initial service list
		services, err := e.GetServices(ctx, dsName)
		if err == nil {
			select {
			case ch <- ServiceChangeEvent{Type: ServiceUpdated, Services: services}:
			case <-ctx.Done():
				return
			}
		}

		// Watch for changes
		watchCh := e.client.Watch(ctx, e.keyPrefix, clientv3.WithPrefix())
		for {
			select {
			case <-ctx.Done():
				return
			case watchResp := <-watchCh:
				if watchResp.Err() != nil {
					continue
				}

				// Re-get service list when there are changes
				services, err := e.GetServices(ctx, dsName)
				if err == nil {
					select {
					case ch <- ServiceChangeEvent{Type: ServiceUpdated, Services: services}:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// Close closes the etcd client
func (e *EtcdServiceRegistry) Close() error {
	// Cancel keepAlive context to stop the goroutine
	if e.cancel != nil {
		e.cancel()
	}
	
	if e.leaseGranted {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = e.client.Revoke(ctx, e.leaseID)
	}
	return e.client.Close()
}

// keepAlive keeps the lease alive
func (e *EtcdServiceRegistry) keepAlive() {
	if !e.leaseGranted {
		return
	}

	if e.ctx == nil || e.cancel == nil {
		e.ctx, e.cancel = context.WithCancel(context.Background())
	}

	ch, kaerr := e.client.KeepAlive(e.ctx, e.leaseID)
	if kaerr != nil {
		return
	}

	// Consume keepalive responses
	for ka := range ch {
		_ = ka // Ignore response, as long as channel is not closed, it means lease is still alive
	}
}

// getServiceKey gets the key of the service in etcd
func (e *EtcdServiceRegistry) getServiceKey(address string) string {
	// Replace special characters in address with safe characters
	safeAddr := strings.ReplaceAll(address, ":", "_")
	safeAddr = strings.ReplaceAll(safeAddr, ".", "-")
	return path.Join(e.keyPrefix, safeAddr)
}

// generateUUID generates a simple UUID-like string
func generateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
