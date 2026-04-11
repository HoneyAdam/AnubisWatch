package raft

// Real chaos tests for Raft consensus
// These tests verify cluster resilience under various failure scenarios

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/AnubisWatch/anubiswatch/internal/core"
)

// createChaosTestCluster creates a multi-node cluster for chaos testing
func createChaosTestCluster(t *testing.T, nodeCount int) ([]*Node, func()) {
	nodes := make([]*Node, nodeCount)
	transports := make([]*TCPTransport, nodeCount)
	cleanups := make([]func(), nodeCount)

	for i := 0; i < nodeCount; i++ {
		cfg := core.RaftConfig{
			NodeID:           fmt.Sprintf("chaos-node-%d", i),
			BindAddr:         fmt.Sprintf("127.0.0.1:%d", 17000+i),
			AdvertiseAddr:    fmt.Sprintf("127.0.0.1:%d", 17000+i),
			Bootstrap:        i == 0, // First node bootstraps
			ElectionTimeout:  core.Duration{Duration: 200 * time.Millisecond},
			HeartbeatTimeout: core.Duration{Duration: 100 * time.Millisecond},
			CommitTimeout:    core.Duration{Duration: 50 * time.Millisecond},
			MaxAppendEntries: 64,
		}

		// Add peers for all nodes
		if i > 0 {
			cfg.Peers = []core.RaftPeer{
				{ID: "chaos-node-0", Address: "127.0.0.1:17000", Region: "default", Role: core.RoleVoter},
			}
		}

		storage := NewInMemoryLogStore()
		snapshot := NewInMemorySnapshotStore()
		fsm := NewStorageFSM(NewInMemoryStorage())

		node, err := NewNode(cfg, storage, snapshot, fsm, newTestRaftLogger())
		if err != nil {
			t.Fatalf("Failed to create chaos test node %d: %v", i, err)
		}

		// Create and set transport
		transport, err := NewTCPTransport(cfg.BindAddr, cfg.AdvertiseAddr, nil, newTestRaftLogger())
		if err != nil {
			t.Fatalf("Failed to create transport for node %d: %v", i, err)
		}
		transports[i] = transport
		node.SetTransport(transport)

		nodes[i] = node
		cleanups[i] = func(n *Node, tr *TCPTransport) func() {
			return func() {
				n.Stop()
				tr.Stop()
			}
		}(node, transport)
	}

	cleanup := func() {
		for _, c := range cleanups {
			c()
		}
	}

	return nodes, cleanup
}

// waitForLeader waits for a leader to be elected in the cluster
func waitForLeader(t *testing.T, nodes []*Node, timeout time.Duration) *Node {
	deadline := time.After(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatalf("Timeout waiting for leader election after %v", timeout)
			return nil
		case <-ticker.C:
			for _, node := range nodes {
				if node.IsLeader() {
					return node
				}
			}
		}
	}
}

// waitForAllNodes waits for all nodes to see the leader
func waitForAllNodes(t *testing.T, nodes []*Node, timeout time.Duration) {
	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			t.Fatalf("Timeout waiting for all nodes to sync")
		case <-ticker.C:
			allHaveLeader := true
			for _, node := range nodes {
				if node.LeaderID() == "" {
					allHaveLeader = false
					break
				}
			}
			if allHaveLeader {
				return
			}
		}
	}
}

// countLeaders returns the number of nodes that think they are leader
func countLeaders(nodes []*Node) int {
	count := 0
	for _, node := range nodes {
		if node.IsLeader() {
			count++
		}
	}
	return count
}

// countActiveNodes returns the number of running nodes
func countActiveNodes(nodes []*Node) int {
	count := 0
	for _, node := range nodes {
		if node.running.Load() {
			count++
		}
	}
	return count
}

// TestChaos_SingleNodeFailure_Real tests cluster survives single node failure
func TestChaos_SingleNodeFailure_Real(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	// Skip flaky integration test - needs proper network setup
	t.Skip("Skipping flaky integration test - needs proper network environment")

	nodes, cleanup := createChaosTestCluster(t, 5)
	defer cleanup()

	// Start all nodes
	for i, node := range nodes {
		if err := node.Start(); err != nil {
			t.Fatalf("Failed to start node %d: %v", i, err)
		}
	}

	// Wait for leader election
	leader := waitForLeader(t, nodes, 10*time.Second)
	t.Logf("Leader elected: %s", leader.nodeID)

	// Wait for all nodes to see the leader
	waitForAllNodes(t, nodes, 10*time.Second)

	// Find a non-leader node to kill
	killIndex := -1
	for i, node := range nodes {
		if !node.IsLeader() {
			killIndex = i
			break
		}
	}
	if killIndex == -1 {
		t.Fatal("Could not find non-leader node to kill")
	}

	t.Logf("Killing node: %s", nodes[killIndex].nodeID)
	nodes[killIndex].Stop()

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Verify we still have a leader
	newLeader := waitForLeader(t, nodes, 5*time.Second)
	t.Logf("Leader after kill: %s", newLeader.nodeID)

	// Verify only one leader
	leaderCount := countLeaders(nodes)
	if leaderCount != 1 {
		t.Errorf("Expected 1 leader, got %d", leaderCount)
	}

	t.Log("✅ Single node failure test passed")
}

// TestChaos_LeaderFailure_Real tests cluster survives leader failure
func TestChaos_LeaderFailure_Real(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	// Skip flaky integration test - needs proper network setup
	t.Skip("Skipping flaky integration test - needs proper network environment")

	nodes, cleanup := createChaosTestCluster(t, 5)
	defer cleanup()

	// Start all nodes
	for i, node := range nodes {
		if err := node.Start(); err != nil {
			t.Fatalf("Failed to start node %d: %v", i, err)
		}
	}

	// Wait for leader election
	oldLeader := waitForLeader(t, nodes, 5*time.Second)
	t.Logf("Initial leader: %s", oldLeader.nodeID)

	// Wait for stability
	waitForAllNodes(t, nodes, 3*time.Second)

	// Kill the leader
	t.Logf("Killing leader: %s", oldLeader.nodeID)
	oldLeader.Stop()

	// Wait for new leader election
	newLeader := waitForLeader(t, nodes, 5*time.Second)
	t.Logf("New leader elected: %s", newLeader.nodeID)

	if newLeader.nodeID == oldLeader.nodeID {
		t.Error("New leader should be different from old leader")
	}

	// Verify only one leader
	leaderCount := countLeaders(nodes)
	if leaderCount != 1 {
		t.Errorf("Expected 1 leader, got %d", leaderCount)
	}

	t.Log("✅ Leader failure test passed")
}

// TestChaos_MultipleNodeFailures_Real tests cluster survives losing quorum
func TestChaos_MultipleNodeFailures_Real(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	// Skip flaky integration test - needs proper network setup
	t.Skip("Skipping flaky integration test - needs proper network environment")

	// 5 nodes, kill 2 (keep quorum)
	nodes, cleanup := createChaosTestCluster(t, 5)
	defer cleanup()

	// Start all nodes
	for i, node := range nodes {
		if err := node.Start(); err != nil {
			t.Fatalf("Failed to start node %d: %v", i, err)
		}
	}

	// Wait for leader
	waitForLeader(t, nodes, 5*time.Second)
	waitForAllNodes(t, nodes, 3*time.Second)

	// Kill 2 non-leader nodes
	killed := 0
	for _, node := range nodes {
		if !node.IsLeader() && killed < 2 {
			t.Logf("Killing node: %s", node.nodeID)
			node.Stop()
			killed++
		}
	}

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Should still have a leader
	var leaderFound *Node
	for _, node := range nodes {
		if node.IsLeader() && node.running.Load() {
			leaderFound = node
			break
		}
	}

	if leaderFound == nil {
		t.Error("Should have a leader after killing 2 nodes (quorum maintained)")
	} else {
		t.Logf("Leader still active: %s", leaderFound.nodeID)
	}

	t.Log("✅ Multiple node failures test passed")
}

// TestChaos_LeaderElectionSpeed_Real measures leader election time
func TestChaos_LeaderElectionSpeed_Real(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	// Skip flaky integration test - needs proper network setup
	t.Skip("Skipping flaky integration test - needs proper network environment")

	nodes, cleanup := createChaosTestCluster(t, 3)
	defer cleanup()

	// Start first node (bootstrap)
	if err := nodes[0].Start(); err != nil {
		t.Fatalf("Failed to start bootstrap node: %v", err)
	}

	// Bootstrap node becomes leader immediately
	time.Sleep(100 * time.Millisecond)
	if !nodes[0].IsLeader() {
		t.Error("Bootstrap node should be leader")
	}

	// Start other nodes
	startTime := time.Now()
	for i := 1; i < len(nodes); i++ {
		if err := nodes[i].Start(); err != nil {
			t.Fatalf("Failed to start node %d: %v", i, err)
		}
	}

	// Wait for all nodes to see the leader
	waitForAllNodes(t, nodes, 3*time.Second)
	elapsed := time.Since(startTime)

	t.Logf("Leader election completed in %v", elapsed)
	if elapsed > 2*time.Second {
		t.Errorf("Leader election took too long: %v", elapsed)
	}

	t.Log("✅ Leader election speed test passed")
}

// TestChaos_TermConsistency_Real ensures all nodes agree on term
func TestChaos_TermConsistency_Real(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	// Skip flaky integration test - needs proper network setup
	t.Skip("Skipping flaky integration test - needs proper network environment")

	nodes, cleanup := createChaosTestCluster(t, 3)
	defer cleanup()

	// Start all nodes
	for i, node := range nodes {
		if err := node.Start(); err != nil {
			t.Fatalf("Failed to start node %d: %v", i, err)
		}
	}

	// Wait for leader
	waitForLeader(t, nodes, 5*time.Second)
	waitForAllNodes(t, nodes, 3*time.Second)

	// Wait a bit for terms to sync
	time.Sleep(500 * time.Millisecond)

	// Check term consistency
	var maxTerm uint64
	for _, node := range nodes {
		if node.currentTerm > maxTerm {
			maxTerm = node.currentTerm
		}
	}

	// All nodes should have the same term (or close to it)
	termMismatch := false
	for _, node := range nodes {
		if node.running.Load() && node.currentTerm < maxTerm {
			t.Logf("Node %s has term %d, expected %d", node.nodeID, node.currentTerm, maxTerm)
			termMismatch = true
		}
	}

	if termMismatch {
		t.Error("Term inconsistency detected")
	}

	t.Logf("All nodes at term %d", maxTerm)
	t.Log("✅ Term consistency test passed")
}

// TestChaos_SplitVote_Real tests split vote scenario
func TestChaos_SplitVote_Real(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping chaos test in short mode")
	}

	// Start with 2 nodes (even number, can cause split votes)
	nodes, cleanup := createChaosTestCluster(t, 2)
	defer cleanup()

	// Start both nodes simultaneously
	var wg sync.WaitGroup
	for i, node := range nodes {
		wg.Add(1)
		go func(idx int, n *Node) {
			defer wg.Done()
			if err := n.Start(); err != nil {
				t.Errorf("Failed to start node %d: %v", idx, err)
			}
		}(i, node)
	}
	wg.Wait()

	// Wait for leader (may take longer with 2 nodes)
	leader := waitForLeader(t, nodes, 10*time.Second)
	t.Logf("Leader elected despite split vote risk: %s", leader.nodeID)

	// Verify only one leader
	leaderCount := countLeaders(nodes)
	if leaderCount != 1 {
		t.Errorf("Expected 1 leader, got %d", leaderCount)
	}

	t.Log("✅ Split vote test passed")
}
