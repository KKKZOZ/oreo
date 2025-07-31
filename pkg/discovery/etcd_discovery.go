package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdServiceDiscovery etcd-based service discovery implementation
type EtcdServiceDiscovery struct {
	client    *clientv3.Client
	keyPrefix string
	config    RegistryConfig

	// Local cache
	mutex       sync.RWMutex
	services    map[string]ServiceInfo    // key: address, value: ServiceInfo
	dsNameIndex map[string][]string       // key: dsName, value: []address
	curIndexMap map[string]*atomic.Uint64 // key: dsName, value: round-robin index

	// Watch control
	watchCtx    context.Context
	watchCancel context.CancelFunc
	watchWg     sync.WaitGroup
}

// NewEtcdServiceDiscovery creates a new etcd service discovery
func NewEtcdServiceDiscovery(
	endpoints []string,
	keyPrefix string,
	config *RegistryConfig,
) (*EtcdServiceDiscovery, error) {
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

	ctx, cancel := context.WithCancel(context.Background())

	d := &EtcdServiceDiscovery{
		client:      client,
		keyPrefix:   keyPrefix,
		config:      *config,
		services:    make(map[string]ServiceInfo),
		dsNameIndex: make(map[string][]string),
		curIndexMap: make(map[string]*atomic.Uint64),
		watchCtx:    ctx,
		watchCancel: cancel,
	}

	// Load existing services during initialization
	if err := d.loadServices(); err != nil {
		cancel()
		_ = client.Close()
		return nil, fmt.Errorf("failed to load initial services: %w", err)
	}

	// Start watching
	d.startWatch()

	return d, nil
}

// GetService gets a service instance for the specified datastore (using round-robin load balancing)
func (d *EtcdServiceDiscovery) GetService(dsName string) (string, error) {
	d.mutex.RLock()
	addresses, exists := d.dsNameIndex[dsName]
	if !exists || len(addresses) == 0 {
		d.mutex.RUnlock()
		return "", fmt.Errorf("no available service for dsName: %s", dsName)
	}

	// Get or create round-robin counter
	indexPtr, exists := d.curIndexMap[dsName]
	if !exists {
		d.mutex.RUnlock()
		d.mutex.Lock()
		// Double check
		indexPtr, exists = d.curIndexMap[dsName]
		if !exists {
			indexPtr = &atomic.Uint64{}
			d.curIndexMap[dsName] = indexPtr
		}
		d.mutex.Unlock()
		d.mutex.RLock()
		// Re-get addresses
		addresses, exists = d.dsNameIndex[dsName]
		if !exists || len(addresses) == 0 {
			d.mutex.RUnlock()
			return "", fmt.Errorf("no available service for dsName: %s", dsName)
		}
	}

	// Round-robin selection
	index := indexPtr.Add(1) % uint64(len(addresses))
	selectedAddr := addresses[index]
	d.mutex.RUnlock()

	return selectedAddr, nil
}

// GetAllServices gets all service instances for the specified datastore
func (d *EtcdServiceDiscovery) GetAllServices(dsName string) ([]string, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	addresses, exists := d.dsNameIndex[dsName]
	if !exists || len(addresses) == 0 {
		return nil, fmt.Errorf("no available service for dsName: %s", dsName)
	}

	// Return a copy to avoid concurrent modification
	result := make([]string, len(addresses))
	copy(result, addresses)
	return result, nil
}

// Watch watches for changes in service registrations for the given datastore name
func (d *EtcdServiceDiscovery) Watch(
	ctx context.Context,
	dsName string,
) (<-chan ServiceChangeEvent, error) {
	ch := make(chan ServiceChangeEvent, 100)

	// Start a goroutine to watch for changes
	go func() {
		defer close(ch)

		// Watch for changes in the etcd key prefix
		watchCh := d.client.Watch(ctx, d.keyPrefix, clientv3.WithPrefix())

		for watchResp := range watchCh {
			for _, event := range watchResp.Events {
				address := d.extractAddressFromKey(string(event.Kv.Key))

				switch event.Type {
				case clientv3.EventTypePut:
					// Service added or updated
					var service ServiceInfo
					if err := json.Unmarshal(event.Kv.Value, &service); err != nil {
						continue
					}

					// Check if this service handles the requested dsName
					handlesDsName := false
					for _, serviceDsName := range service.DsNames {
						if serviceDsName == dsName {
							handlesDsName = true
							break
						}
					}

					if handlesDsName {
						select {
						case ch <- ServiceChangeEvent{
							Type:    ServiceAdded,
							Service: service,
						}:
						case <-ctx.Done():
							return
						}
					}

				case clientv3.EventTypeDelete:
					// Service removed
					d.mutex.RLock()
					service, exists := d.services[address]
					d.mutex.RUnlock()

					if exists {
						// Check if this service handles the requested dsName
						handlesDsName := false
						for _, serviceDsName := range service.DsNames {
							if serviceDsName == dsName {
								handlesDsName = true
								break
							}
						}

						if handlesDsName {
							select {
							case ch <- ServiceChangeEvent{
								Type:    ServiceRemoved,
								Service: service,
							}:
							case <-ctx.Done():
								return
							}
						}
					}
				}
			}
		}
	}()

	return ch, nil
}

// Close closes the service discovery
func (d *EtcdServiceDiscovery) Close() error {
	// Stop watching
	d.watchCancel()
	d.watchWg.Wait()

	// Close etcd client
	return d.client.Close()
}

// loadServices loads service list from etcd
func (d *EtcdServiceDiscovery) loadServices() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := d.client.Get(ctx, d.keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("failed to get services from etcd: %w", err)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Clear existing data
	d.services = make(map[string]ServiceInfo)
	d.dsNameIndex = make(map[string][]string)

	// Load services from etcd
	for _, kv := range resp.Kvs {
		var serviceInfo ServiceInfo
		if err := json.Unmarshal(kv.Value, &serviceInfo); err != nil {
			continue // Skip services that cannot be parsed
		}

		// Update service cache
		d.services[serviceInfo.Address] = serviceInfo

		// Update dsName index
		for _, dsName := range serviceInfo.DsNames {
			d.dsNameIndex[dsName] = append(d.dsNameIndex[dsName], serviceInfo.Address)
		}
	}

	return nil
}

// startWatch starts etcd watching
func (d *EtcdServiceDiscovery) startWatch() {
	d.watchWg.Add(1)
	go func() {
		defer d.watchWg.Done()

		watchCh := d.client.Watch(d.watchCtx, d.keyPrefix, clientv3.WithPrefix())
		for {
			select {
			case <-d.watchCtx.Done():
				return
			case watchResp := <-watchCh:
				if watchResp.Err() != nil {
					continue
				}

				// Handle change events
				for _, event := range watchResp.Events {
					switch event.Type {
					case clientv3.EventTypePut:
						// Service added or updated
						var serviceInfo ServiceInfo
						if err := json.Unmarshal(event.Kv.Value, &serviceInfo); err == nil {
							d.addService(serviceInfo)
						}
					case clientv3.EventTypeDelete:
						// Service deleted
						// Extract address from key
						address := d.extractAddressFromKey(string(event.Kv.Key))
						if address != "" {
							d.removeService(address)
						}
					}
				}
			}
		}
	}()
}

// addService adds service to local cache
func (d *EtcdServiceDiscovery) addService(service ServiceInfo) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// If service already exists, remove old index first
	if oldService, exists := d.services[service.Address]; exists {
		d.removeServiceFromIndexLocked(oldService.Address, oldService.DsNames)
	}

	// Add new service
	d.services[service.Address] = service
	d.addServiceToIndexLocked(service.Address, service.DsNames)
}

// removeService removes service from local cache
func (d *EtcdServiceDiscovery) removeService(address string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if service, exists := d.services[address]; exists {
		d.removeServiceFromIndexLocked(address, service.DsNames)
		delete(d.services, address)
	}
}

// addServiceToIndexLocked adds service to dsName index (requires holding lock)
func (d *EtcdServiceDiscovery) addServiceToIndexLocked(address string, dsNames []string) {
	for _, dsName := range dsNames {
		// Check if already exists to avoid duplication
		addresses := d.dsNameIndex[dsName]
		for _, addr := range addresses {
			if addr == address {
				continue // Already exists, skip
			}
		}
		// Does not exist, add
		d.dsNameIndex[dsName] = append(d.dsNameIndex[dsName], address)
	}
}

// removeServiceFromIndexLocked removes service from dsName index (requires holding lock)
func (d *EtcdServiceDiscovery) removeServiceFromIndexLocked(address string, dsNames []string) {
	for _, dsName := range dsNames {
		addresses := d.dsNameIndex[dsName]
		for i, addr := range addresses {
			if addr == address {
				// Remove this address
				d.dsNameIndex[dsName] = append(addresses[:i], addresses[i+1:]...)
				break
			}
		}
		// If there are no services under this dsName, delete the key
		if len(d.dsNameIndex[dsName]) == 0 {
			delete(d.dsNameIndex, dsName)
		}
	}
}

// extractAddressFromKey extracts service address from etcd key
func (d *EtcdServiceDiscovery) extractAddressFromKey(key string) string {
	// key format: /oreo/services/192-168-1-100_8000
	// Extract the last part and convert back to address format
	parts := strings.Split(key, "/")
	if len(parts) == 0 {
		return ""
	}

	safeAddr := parts[len(parts)-1]
	// Convert safe characters back to original address
	address := strings.ReplaceAll(safeAddr, "_", ":")
	address = strings.ReplaceAll(address, "-", ".")

	return address
}
