package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/pkg/discovery"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testEtcdEndpoints []string

func TestMain(m *testing.M) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "quay.io/coreos/etcd:v3.5.9",
		ExposedPorts: []string{"2379/tcp"},
		Env: map[string]string{
			"ETCD_NAME":                        "etcd0",
			"ETCD_ADVERTISE_CLIENT_URLS":        "http://0.0.0.0:2379",
			"ETCD_LISTEN_CLIENT_URLS":           "http://0.0.0.0:2379",
			"ETCD_INITIAL_ADVERTISE_PEER_URLS":  "http://0.0.0.0:2380",
			"ETCD_LISTEN_PEER_URLS":             "http://0.0.0.0:2380",
			"ETCD_INITIAL_CLUSTER_TOKEN":        "etcd-cluster-1",
			"ETCD_INITIAL_CLUSTER":              "etcd0=http://0.0.0.0:2380",
			"ETCD_INITIAL_CLUSTER_STATE":        "new",
			"ALLOW_NONE_AUTHENTICATION":        "yes",
		},
		WaitingFor: wait.ForListeningPort("2379/tcp"),
	}
	etcdContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)
	if err != nil {
		log.Fatalf("Could not start etcd container: %s", err)
	}

	defer func() {
		if err := etcdContainer.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop etcd container: %s", err)
		}
	}()

	mappedPort, _ := etcdContainer.MappedPort(ctx, "2379")
	host, _ := etcdContainer.Host(ctx)
	testEtcdEndpoints = []string{fmt.Sprintf("%s:%s", host, mappedPort.Port())}

	// Wait for etcd to be ready
	time.Sleep(2 * time.Second)

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestEtcdServiceDiscovery(t *testing.T) {
	ctx := context.Background()

	// Create configuration
	config := discovery.DefaultRegistryConfig()
	config.HeartbeatInterval = 2 * time.Second
	config.InstanceTTL = 30 * time.Second

	// Create EtcdServiceRegistry instance
	registry, err := discovery.NewEtcdServiceRegistry(
		testEtcdEndpoints,
		"/oreo/services",
		config,
	)
	if err != nil {
		t.Fatalf("Failed to create EtcdServiceRegistry: %v", err)
	}
	defer func() { _ = registry.Close() }()

	// Register service
	err = registry.Register(ctx, "localhost:8002", []string{"Redis", "MongoDB"}, nil)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}
	t.Log("Service registered to etcd successfully")

	// Send heartbeat
	err = registry.Heartbeat(ctx, "localhost:8002")
	if err != nil {
		t.Fatalf("Heartbeat failed: %v", err)
	}
	t.Log("Etcd heartbeat sent successfully")

	// Wait a moment to ensure service registration is complete
	time.Sleep(2 * time.Second)

	// Now use EtcdServiceDiscovery to discover services
	t.Log("Using EtcdServiceDiscovery to discover services")
	serviceDiscovery, err := discovery.NewEtcdServiceDiscovery(
		testEtcdEndpoints,
		"/oreo/services",
		config,
	)
	if err != nil {
		t.Fatalf("Failed to create EtcdServiceDiscovery: %v", err)
	}
	defer func() { _ = serviceDiscovery.Close() }()

	// Wait for etcd connection establishment and data synchronization
	time.Sleep(3 * time.Second)

	// Get service (from etcd)
	address, err := serviceDiscovery.GetService("Redis")
	if err != nil {
		t.Fatalf("Failed to get service from etcd: %v", err)
	}
	t.Logf("Got Redis service from etcd: %s", address)

	// Test multiple calls to GetService (round-robin)
	t.Log("Testing round-robin load balancing")
	for i := 0; i < 3; i++ {
		addr, err := serviceDiscovery.GetService("Redis")
		if err != nil {
			t.Fatalf("Failed to get service from etcd (call %d): %v", i+1, err)
		}
		t.Logf("Round-robin call %d: %s", i+1, addr)
	}

	// Register another service to test load balancing
	err = registry.Register(ctx, "localhost:8003", []string{"Redis"}, nil)
	if err != nil {
		t.Logf("Failed to register second service: %v", err)
	} else {
		t.Log("Second Redis service registered")
		// Wait for discovery to pick up the new service
		time.Sleep(2 * time.Second)
		
		// Test load balancing with multiple services
		addresses := make(map[string]int)
		for i := 0; i < 10; i++ {
			addr, err := serviceDiscovery.GetService("Redis")
			if err != nil {
				t.Fatalf("Failed to get service: %v", err)
			}
			addresses[addr]++
		}
		t.Logf("Load balancing results: %v", addresses)
	}

	// Finally deregister services
	err = registry.Deregister(ctx, "localhost:8002")
	if err != nil {
		t.Fatalf("Deregistration failed: %v", err)
	}
	t.Log("Service deregistered from etcd successfully")

	err = registry.Deregister(ctx, "localhost:8003")
	if err != nil {
		t.Logf("Failed to deregister second service: %v", err)
	}
}

func TestEtcdServiceRegistryBasic(t *testing.T) {
	ctx := context.Background()

	// Create configuration
	config := discovery.DefaultRegistryConfig()
	config.HeartbeatInterval = 1 * time.Second
	config.InstanceTTL = 10 * time.Second

	// Create EtcdServiceRegistry instance
	registry, err := discovery.NewEtcdServiceRegistry(
		testEtcdEndpoints,
		"/oreo/test",
		config,
	)
	if err != nil {
		t.Fatalf("Failed to create EtcdServiceRegistry: %v", err)
	}
	defer func() { _ = registry.Close() }()

	// Test service registration
	serviceAddress := "localhost:9001"
	serviceTypes := []string{"TestService"}

	err = registry.Register(ctx, serviceAddress, serviceTypes, nil)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Create a discovery client to verify registration
	discovery, err := discovery.NewEtcdServiceDiscovery(
		testEtcdEndpoints,
		"/oreo/test",
		config,
	)
	if err != nil {
		t.Fatalf("Failed to create EtcdServiceDiscovery: %v", err)
	}
	defer func() { _ = discovery.Close() }()

	// Wait for service to be available
	time.Sleep(2 * time.Second)

	// Test service retrieval
	addr, err := discovery.GetService("TestService")
	if err != nil {
		t.Fatalf("Failed to get service: %v", err)
	}

	if addr != serviceAddress {
		t.Fatalf("Expected service address %s, got %s", serviceAddress, addr)
	}

	// Test heartbeat
	err = registry.Heartbeat(ctx, serviceAddress)
	if err != nil {
		t.Fatalf("Failed to send heartbeat: %v", err)
	}

	// Test service deregistration
	err = registry.Deregister(ctx, serviceAddress)
	if err != nil {
		t.Fatalf("Failed to deregister service: %v", err)
	}

	// Verify service is removed
	time.Sleep(2 * time.Second)
	_, err = discovery.GetService("TestService")
	if err == nil {
		t.Fatal("Service still found after deregistration")
	}
	t.Log("Service successfully removed after deregistration")
}