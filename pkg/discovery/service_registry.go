package discovery

import (
	"context"
	"time"
)

// ServiceInfo represents information about a registered service instance

type ServiceInfo struct {
	ID string // Unique identifier for the service instance (e.g., UUID)

	Address string // The advertised network address (e.g., "1.2.3.4:8000")

	LastHeartbeat time.Time // Timestamp of the last successful heartbeat

	DsNames []string // List of datastore names this instance handles (e.g., ["Redis", "MongoDB1"])

	Metadata map[string]string // Additional metadata for the service
}

// ServiceChangeType represents the type of service change

type ServiceChangeType int

const (
	ServiceAdded ServiceChangeType = iota

	ServiceRemoved

	ServiceUpdated
)

// ServiceChangeEvent represents a change in service registration

type ServiceChangeEvent struct {
	Type ServiceChangeType

	Service ServiceInfo

	Services []ServiceInfo // For batch updates
}

// ServiceRegistry defines the interface for service registration and discovery

type ServiceRegistry interface {
	// Register registers a service instance with the registry

	Register(
		ctx context.Context,

		address string,

		dsNames []string,

		metadata map[string]string,
	) error

	// Deregister removes a service instance from the registry

	Deregister(ctx context.Context, address string) error

	// Heartbeat sends a heartbeat to maintain the service registration

	Heartbeat(ctx context.Context, address string) error

	// GetServices retrieves all registered service instances for given datastore name

	GetServices(ctx context.Context, dsName string) ([]ServiceInfo, error)

	// GetAllServices retrieves all registered service instances

	GetAllServices(ctx context.Context) ([]ServiceInfo, error)

	// Watch watches for changes in service registrations

	Watch(ctx context.Context, dsName string) (<-chan ServiceChangeEvent, error)

	// Close closes the registry and cleans up resources

	Close() error
}

// ServiceDiscovery defines the interface for service discovery (client-side)

type ServiceDiscovery interface {
	// GetService returns a service instance address for the given datastore name

	GetService(dsName string) (string, error)

	// GetAllServices returns all available service instances for the given datastore name

	GetAllServices(dsName string) ([]string, error)

	// Watch watches for changes in service registrations for the given datastore name

	Watch(ctx context.Context, dsName string) (<-chan ServiceChangeEvent, error)

	// Close closes the discovery client

	Close() error
}

// RegistryConfig holds configuration for service registry

type RegistryConfig struct {
	InstanceTTL time.Duration // Time after last heartbeat before an instance is considered stale

	HeartbeatInterval time.Duration // Interval for sending heartbeats

	CleanupInterval time.Duration // Interval for cleaning up stale instances

	RegistryAddr string // Address of the registry service

	RequestTimeout time.Duration // Timeout for HTTP requests
}

// DefaultRegistryConfig returns default configuration for service registry

func DefaultRegistryConfig() *RegistryConfig {
	return &RegistryConfig{
		InstanceTTL: 6 * time.Second,

		HeartbeatInterval: 2 * time.Second,

		CleanupInterval: 3 * time.Second,

		RequestTimeout: 1 * time.Second,
	}
}

// RegistryRequest is the payload structure used for communication

// between executors and the registry server

type RegistryRequest struct {
	Address string `json:"address"` // Address of the executor instance

	DsNames []string `json:"dsNames,omitempty"` // Datastore names (primarily for /register)
}
