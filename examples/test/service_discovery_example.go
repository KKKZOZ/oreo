package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kkkzoz/oreo/pkg/discovery"
)

func main() {
	// Complete Etcd service registration and discovery example
	fmt.Println("=== Complete Etcd Service Registration and Discovery Example ===")
	etcdCompleteExample()
}

// etcdCompleteExample Complete etcd service registration and discovery example
func etcdCompleteExample() {
	// Create configuration
	config := discovery.DefaultRegistryConfig()
	config.HeartbeatInterval = 2 * time.Second
	config.InstanceTTL = 30 * time.Second

	// Create EtcdServiceRegistry instance
	registry, err := discovery.NewEtcdServiceRegistry([]string{"localhost:2379"}, "/oreo/services", config)
	if err != nil {
		log.Printf("Failed to create EtcdServiceRegistry: %v", err)
		return
	}
	defer registry.Close()

	ctx := context.Background()

	// Register service
	err = registry.Register(ctx, "localhost:8002", []string{"Redis", "MongoDB"}, nil)
	if err != nil {
		log.Printf("Registration failed: %v", err)
		return
	}
	fmt.Println("Service registered to etcd successfully")

	// Send heartbeat
	err = registry.Heartbeat(ctx, "localhost:8002")
	if err != nil {
		log.Printf("Heartbeat failed: %v", err)
	}
	fmt.Println("Etcd heartbeat sent successfully")

	// Get service list (through registry)
	services, err := registry.GetServices(ctx, "Redis")
	if err != nil {
		log.Printf("Failed to get services: %v", err)
	} else {
		fmt.Printf("Etcd Redis service list: %v\n", services)
	}

	// Wait a moment to ensure service registration is complete
	time.Sleep(2 * time.Second)

	// Now use EtcdServiceDiscovery to discover services
	fmt.Println("\n--- Using EtcdServiceDiscovery to discover services ---")
	serviceDiscovery, err := discovery.NewEtcdServiceDiscovery([]string{"localhost:2379"}, "/oreo/services", config)
	if err != nil {
		log.Printf("Failed to create EtcdServiceDiscovery: %v", err)
		return
	}
	defer serviceDiscovery.Close()

	// Wait for etcd connection establishment and data synchronization
	time.Sleep(3 * time.Second)

	// Get service (from etcd)
	address, err := serviceDiscovery.GetService("Redis")
	if err != nil {
		log.Printf("Failed to get service from etcd: %v", err)
	} else {
		fmt.Printf("Got Redis service from etcd: %s\n", address)
	}

	// Get all services
	allServices, err := serviceDiscovery.GetAllServices("Redis")
	if err != nil {
		log.Printf("Failed to get services: %v", err)
	} else {
		fmt.Printf("Got all Redis services from etcd: %v\n", allServices)
	}

	// Finally deregister service
	err = registry.Deregister(ctx, "localhost:8002")
	if err != nil {
		log.Printf("Deregistration failed: %v", err)
	} else {
		fmt.Println("Service deregistered from etcd successfully")
	}
}