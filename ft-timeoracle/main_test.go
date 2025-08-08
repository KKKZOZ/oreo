
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kkkzoz/oreo/pkg/timesource"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func setupEtcd(t *testing.T) (*embed.Etcd, *clientv3.Client, func()) {
	dir, err := ioutil.TempDir("", "etcd-")
	if err != nil {
		t.Fatal(err)
	}

	cfg := embed.NewConfig()
	cfg.Dir = dir
	cfg.LogLevel = "error"
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case <-e.Server.ReadyNotify():
		// Ready
	case <-time.After(10 * time.Second):
		e.Server.Stop() // trigger a shutdown
		t.Fatal("Etcd server took too long to start")
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{e.Clients[0].Addr().String()},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		e.Close()
		os.RemoveAll(dir)
		client.Close()
	}

	return e, client, cleanup
}

func TestLeaderElection(t *testing.T) {
	_, client, cleanup := setupEtcd(t)
	defer cleanup()

	electionName := "/test-leader"
	node1ID := "node1"
	node2ID := "node2"

	lm1, err := NewLeaseManager(client, electionName, log.New(os.Stdout, "node1: ", log.LstdFlags))
	if err != nil {
		t.Fatalf("Failed to create lease manager for node1: %v", err)
	}
	defer lm1.Close()

	lm2, err := NewLeaseManager(client, electionName, log.New(os.Stdout, "node2: ", log.LstdFlags))
	if err != nil {
		t.Fatalf("Failed to create lease manager for node2: %v", err)
	}
	defer lm2.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Node 1 becomes leader
	go func() {
		if err := lm1.Campaign(ctx, node1ID); err != nil {
			t.Errorf("Node1 campaign failed: %v", err)
		}
	}()

	time.Sleep(2 * time.Second) // Give time for node1 to become leader

	// Check leader
	resp, err := lm1.election.Leader(ctx)
	if err != nil {
		t.Fatalf("Failed to get leader: %v", err)
	}
	if string(resp.Kvs[0].Value) != node1ID {
		t.Fatalf("Expected leader to be %s, but got %s", node1ID, string(resp.Kvs[0].Value))
	}

	// Node 2 tries to become leader, should block
	node2Campaigned := make(chan struct{})
	go func() {
		if err := lm2.Campaign(ctx, node2ID); err != nil {
			t.Errorf("Node2 campaign failed: %v", err)
		}
		close(node2Campaigned)
	}()

	// Resign leadership from node 1
	if err := lm1.Resign(ctx); err != nil {
		t.Fatalf("Node1 failed to resign: %v", err)
	}

	// Node 2 should now become leader
	<-node2Campaigned

	resp, err = lm2.election.Leader(ctx)
	if err != nil {
		t.Fatalf("Failed to get leader: %v", err)
	}
	if string(resp.Kvs[0].Value) != node2ID {
		t.Fatalf("Expected leader to be %s, but got %s", node2ID, string(resp.Kvs[0].Value))
	}
}

func TestMainFunctionality(t *testing.T) {
	etcd, client, cleanup := setupEtcd(t)
	defer cleanup()

	// Mock the main function's core logic
	etcdEndpoints = etcd.Clients[0].Addr().String()
	port = 8081 // Use a different port for testing
	oracleType = "simple"
	electionName = "/test-main-leader"
	maxSkewStr = "50ms"

	// Setup logger
	newLogger()

	// Setup oracle
	globalOracle = timesource.NewSimpleTimeSource()

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/timestamp/", handleTimestamp)
	serverAddress := fmt.Sprintf(":%d", port)
	httpServer := &http.Server{
		Addr:    serverAddress,
		Handler: mux,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Test server closed: %v", err)
		}
	}()
	defer httpServer.Shutdown(context.Background())

	// Test inactive node
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/timestamp", port))
	if err != nil {
		t.Fatalf("Failed to send request to inactive node: %v", err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d for inactive node, got %d", http.StatusServiceUnavailable, resp.StatusCode)
	}
	resp.Body.Close()

	// Start leader election
	nodeID := "test-node"
	lm, err := NewLeaseManager(client, electionName, log.New(os.Stdout, "", log.LstdFlags))
	if err != nil {
		t.Fatalf("Failed to create lease manager: %v", err)
	}
	defer lm.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := lm.Campaign(ctx, nodeID); err != nil {
			t.Errorf("Campaign failed: %v", err)
		}
		setActive(true)
	}()

	time.Sleep(2 * time.Second) // Give time for election

	// Test active node
	resp, err = http.Get(fmt.Sprintf("http://localhost:%d/timestamp", port))
	if err != nil {
		t.Fatalf("Failed to send request to active node: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d for active node, got %d", http.StatusOK, resp.StatusCode)
	}
	resp.Body.Close()
}
