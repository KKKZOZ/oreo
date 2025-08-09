package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdServiceRegistry implements the discovery.ServiceRegistry interface.
var _ ServiceRegistry = (*EtcdServiceRegistry)(nil)

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

	// Generate unique ID for the service instance using google/uuid
	serviceID := uuid.NewString()

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
