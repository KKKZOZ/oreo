
package main

import (
	"context"
	"log"

	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// LeaseManager handles the leader election process using etcd.
type LeaseManager struct {
	client   *clientv3.Client
	session  *concurrency.Session
	election *concurrency.Election
	logger   *log.Logger
}

// NewLeaseManager creates a new LeaseManager.
func NewLeaseManager(client *clientv3.Client, electionName string, logger *log.Logger) (*LeaseManager, error) {
	session, err := concurrency.NewSession(client)
	if err != nil {
		return nil, err
	}

	election := concurrency.NewElection(session, electionName)

	return &LeaseManager{
		client:   client,
		session:  session,
		election: election,
		logger:   logger,
	}, nil
}

// Campaign starts the leader election campaign. It blocks until this node becomes the leader.
func (lm *LeaseManager) Campaign(ctx context.Context, nodeID string) error {
	lm.logger.Printf("Starting leader election campaign for node %s", nodeID)
	if err := lm.election.Campaign(ctx, nodeID); err != nil {
		return err
	}
	lm.logger.Printf("Node %s became the leader", nodeID)
	return nil
}

// Resign gives up the leadership.
func (lm *LeaseManager) Resign(ctx context.Context) error {
	lm.logger.Printf("Resigning leadership")
	return lm.election.Resign(ctx)
}

// Close closes the underlying session.
func (lm *LeaseManager) Close() error {
	if lm.session != nil {
		return lm.session.Close()
	}
	return nil
}

// WatchLeader watches for changes in the leadership and calls the provided function when a new leader is elected.
func (lm *LeaseManager) WatchLeader(ctx context.Context, onNewLeader func(leader string)) {
	ch := lm.election.Observe(ctx)
	go func() {
		for {
			select {
			case resp := <-ch:
				if len(resp.Kvs) > 0 {
					leader := string(resp.Kvs[0].Value)
					onNewLeader(leader)
				}
			case <-ctx.Done():
				lm.logger.Println("Leader watch stopped")
				return
			}
		}
	}()
}
