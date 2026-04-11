package grpcapi

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	v1 "github.com/AnubisWatch/anubiswatch/internal/grpcapi/v1"
)

// mockGRPCStore implements Store with in-memory data
type mockGRPCStore struct {
	souls      map[string]interface{}
	judgments  []interface{}
	channels   map[string]interface{}
	rules      map[string]interface{}
	journeys   map[string]interface{}
}

func newMockGRPCStore() *mockGRPCStore {
	return &mockGRPCStore{
		souls:      make(map[string]interface{}),
		channels:   make(map[string]interface{}),
		rules:      make(map[string]interface{}),
		journeys:   make(map[string]interface{}),
	}
}

func (m *mockGRPCStore) GetSoulNoCtx(id string) (interface{}, error)         { return m.souls[id], nil }
func (m *mockGRPCStore) ListSoulsNoCtx(ws string, o, l int) ([]interface{}, error) {
	result := make([]interface{}, 0, len(m.souls))
	for _, s := range m.souls {
		result = append(result, s)
	}
	return result, nil
}
func (m *mockGRPCStore) SaveSoulNoCtx(s interface{}) error    { return nil }
func (m *mockGRPCStore) DeleteSoulNoCtx(id string) error      { delete(m.souls, id); return nil }
func (m *mockGRPCStore) ListJudgmentsNoCtx(soulID string, start, end time.Time, limit int) ([]interface{}, error) {
	return m.judgments, nil
}
func (m *mockGRPCStore) GetChannelNoCtx(id string) (interface{}, error) { return m.channels[id], nil }
func (m *mockGRPCStore) ListChannelsNoCtx(ws string) ([]interface{}, error) {
	result := make([]interface{}, 0, len(m.channels))
	for _, c := range m.channels {
		result = append(result, c)
	}
	return result, nil
}
func (m *mockGRPCStore) SaveChannelNoCtx(ch interface{}) error { return nil }
func (m *mockGRPCStore) DeleteChannelNoCtx(id string) error    { delete(m.channels, id); return nil }
func (m *mockGRPCStore) GetRuleNoCtx(id string) (interface{}, error) { return m.rules[id], nil }
func (m *mockGRPCStore) ListRulesNoCtx(ws string) ([]interface{}, error) {
	result := make([]interface{}, 0, len(m.rules))
	for _, r := range m.rules {
		result = append(result, r)
	}
	return result, nil
}
func (m *mockGRPCStore) SaveRuleNoCtx(rule interface{}) error { return nil }
func (m *mockGRPCStore) DeleteRuleNoCtx(id string) error      { delete(m.rules, id); return nil }
func (m *mockGRPCStore) GetJourneyNoCtx(id string) (interface{}, error) { return m.journeys[id], nil }
func (m *mockGRPCStore) ListJourneysNoCtx(ws string, o, l int) ([]interface{}, error) {
	result := make([]interface{}, 0, len(m.journeys))
	for _, j := range m.journeys {
		result = append(result, j)
	}
	return result, nil
}
func (m *mockGRPCStore) SaveJourneyNoCtx(j interface{}) error { return nil }
func (m *mockGRPCStore) DeleteJourneyNoCtx(id string) error   { delete(m.journeys, id); return nil }

type mockGRPCProbe struct{}

func (m *mockGRPCProbe) ForceCheck(soulID string) (interface{}, error) {
	return nil, nil
}

func TestNewServer(t *testing.T) {
	store := newMockGRPCStore()
	srv := NewServer(":0", store, &mockGRPCProbe{}, nil)
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
	if srv.grpc == nil {
		t.Fatal("gRPC server not initialized")
	}
}

func TestServer_ListSouls(t *testing.T) {
	store := newMockGRPCStore()
	srv := NewServer(":0", store, &mockGRPCProbe{}, nil)

	resp, err := srv.ListSouls(context.Background(), &v1.ListSoulsRequest{
		Offset: 0,
		Limit:  20,
	})
	if err != nil {
		t.Fatalf("ListSouls failed: %v", err)
	}
	if resp.Pagination.Total != 0 {
		t.Errorf("Expected 0 souls, got %d", resp.Pagination.Total)
	}
}

func TestServer_GetSoul_NotFound(t *testing.T) {
	store := newMockGRPCStore()
	srv := NewServer(":0", store, &mockGRPCProbe{}, nil)

	_, err := srv.GetSoul(context.Background(), &v1.GetSoulRequest{Id: "nonexistent"})
	if err == nil {
		t.Error("Expected error for nonexistent soul")
	}
}

func TestServer_GetClusterStatus(t *testing.T) {
	store := newMockGRPCStore()
	srv := NewServer(":0", store, &mockGRPCProbe{}, nil)

	resp, err := srv.GetClusterStatus(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetClusterStatus failed: %v", err)
	}
	if resp.NodeId != "single-node" {
		t.Errorf("Expected node ID 'single-node', got %s", resp.NodeId)
	}
	if !resp.IsLeader {
		t.Error("Expected IsLeader to be true")
	}
}

func TestServer_ListChannels(t *testing.T) {
	store := newMockGRPCStore()
	srv := NewServer(":0", store, &mockGRPCProbe{}, nil)

	resp, err := srv.ListChannels(context.Background(), &v1.ListChannelsRequest{})
	if err != nil {
		t.Fatalf("ListChannels failed: %v", err)
	}
	if len(resp.Channels) != 0 {
		t.Errorf("Expected 0 channels, got %d", len(resp.Channels))
	}
}

func TestServer_ListRules(t *testing.T) {
	store := newMockGRPCStore()
	srv := NewServer(":0", store, &mockGRPCProbe{}, nil)

	resp, err := srv.ListRules(context.Background(), &v1.ListRulesRequest{})
	if err != nil {
		t.Fatalf("ListRules failed: %v", err)
	}
	if len(resp.Rules) != 0 {
		t.Errorf("Expected 0 rules, got %d", len(resp.Rules))
	}
}

func TestServer_ListJourneys(t *testing.T) {
	store := newMockGRPCStore()
	srv := NewServer(":0", store, &mockGRPCProbe{}, nil)

	resp, err := srv.ListJourneys(context.Background(), &v1.ListJourneysRequest{})
	if err != nil {
		t.Fatalf("ListJourneys failed: %v", err)
	}
	if len(resp.Journeys) != 0 {
		t.Errorf("Expected 0 journeys, got %d", len(resp.Journeys))
	}
}

// TestGRPCServer_Listen tests that the server can actually listen and accept connections
func TestGRPCServer_Listen(t *testing.T) {
	store := newMockGRPCStore()
	srv := NewServer("127.0.0.1:0", store, &mockGRPCProbe{}, nil)

	if err := srv.Start(); err != nil {
		t.Fatalf("Failed to start gRPC server: %v", err)
	}
	defer srv.Stop()

	// Try to connect
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := srv.listener.Addr().String()
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := v1.NewAnubisWatchServiceClient(conn)
	status, err := client.GetClusterStatus(ctx, nil)
	if err != nil {
		t.Fatalf("GetClusterStatus RPC failed: %v", err)
	}
	if status.NodeId != "single-node" {
		t.Errorf("Expected node ID 'single-node', got %s", status.NodeId)
	}
}

// TestGRPCServer_Bufconn tests the server with an in-memory buffer connection
func TestGRPCServer_Bufconn(t *testing.T) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	store := newMockGRPCStore()
	srv := NewServer("bufconn", store, &mockGRPCProbe{}, nil)

	go func() {
		srv.grpc.Serve(lis)
	}()
	defer srv.grpc.GracefulStop()

	// Dial with bufconn
	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := v1.NewAnubisWatchServiceClient(conn)

	// Test ListSouls
	resp, err := client.ListSouls(ctx, &v1.ListSoulsRequest{Limit: 10})
	if err != nil {
		t.Fatalf("ListSouls failed: %v", err)
	}
	if resp == nil {
		t.Fatal("ListSouls returned nil response")
	}

	// Test GetClusterStatus
	status, err := client.GetClusterStatus(ctx, nil)
	if err != nil {
		t.Fatalf("GetClusterStatus failed: %v", err)
	}
	if status.NodeCount != 1 {
		t.Errorf("Expected 1 node, got %d", status.NodeCount)
	}
}
