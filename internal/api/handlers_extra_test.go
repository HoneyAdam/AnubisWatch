package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AnubisWatch/anubiswatch/internal/core"
)

// Helper to create a test REST server
func newTestServerWithStorage(store Storage) *RESTServer {
	config := core.ServerConfig{Port: 8080}
	logger := newTestLogger()
	return NewRESTServer(config, core.AuthConfig{Enabled: true}, store, &mockProbeEngine{}, &mockAlertManager{}, &mockAuthenticator{}, &mockClusterManager{}, nil, nil, nil, logger)
}

// TestHandleListJourneys tests handleListJourneys
func TestHandleListJourneys(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:   httptest.NewRequest("GET", "/api/v1/journeys", nil),
		Response:  rec,
		Workspace: "default",
	}

	err := server.handleListJourneys(ctx)
	if err != nil {
		t.Fatalf("handleListJourneys failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// TestHandleCreateJourney tests handleCreateJourney
func TestHandleCreateJourney(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	journey := core.JourneyConfig{
		Name: "Test Journey",
		Steps: []core.JourneyStep{
			{Name: "Step 1", Target: "http://example.com"},
		},
	}
	body, _ := json.Marshal(journey)

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:   httptest.NewRequest("POST", "/api/v1/journeys", bytes.NewReader(body)),
		Response:  rec,
		Workspace: "default",
	}

	err := server.handleCreateJourney(ctx)
	if err != nil {
		t.Fatalf("handleCreateJourney failed: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var result core.JourneyConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.Name != journey.Name {
		t.Errorf("Expected name %s, got %s", journey.Name, result.Name)
	}

	if result.ID == "" {
		t.Error("Expected journey ID to be generated")
	}
}

// TestHandleCreateJourney_InvalidData tests handleCreateJourney with invalid data
func TestHandleCreateJourney_InvalidData(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/journeys", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := &Context{
		Request:   req,
		Response:  rec,
		Workspace: "default",
	}

	err := server.handleCreateJourney(ctx)
	// Error may be returned or set in context, check both
	if err == nil && rec.Code != http.StatusBadRequest {
		t.Errorf("Expected error or bad request status, got status %d", rec.Code)
	}
}

// TestHandleUpdateJourney tests handleUpdateJourney
func TestHandleUpdateJourney(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	updated := core.JourneyConfig{
		Name: "Updated Name",
		Steps: []core.JourneyStep{
			{Name: "Updated Step", Target: "http://updated.com"},
		},
	}
	body, _ := json.Marshal(updated)

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:   httptest.NewRequest("PUT", "/api/v1/journeys/journey-1", bytes.NewReader(body)),
		Response:  rec,
		Params:    map[string]string{"id": "journey-1"},
		Workspace: "default",
	}

	err := server.handleUpdateJourney(ctx)
	if err != nil {
		t.Fatalf("handleUpdateJourney failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// TestHandleUpdateJourney_InvalidData tests handleUpdateJourney with invalid data
func TestHandleUpdateJourney_InvalidData(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/v1/journeys/journey-1", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	ctx := &Context{
		Request:  req,
		Response: rec,
		Params:   map[string]string{"id": "journey-1"},
	}

	err := server.handleUpdateJourney(ctx)
	if err == nil && rec.Code != http.StatusBadRequest {
		t.Errorf("Expected error or bad request status, got status %d", rec.Code)
	}
}

// TestHandleDeleteJourney tests handleDeleteJourney
func TestHandleDeleteJourney(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:  httptest.NewRequest("DELETE", "/api/v1/journeys/journey-1", nil),
		Response: rec,
		Params:   map[string]string{"id": "journey-1"},
	}

	// Just verify it doesn't panic
	_ = server.handleDeleteJourney(ctx)
}

// TestHandleGetJourney tests handleGetJourney
func TestHandleGetJourney(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	// Create a journey first
	store.SaveJourneyNoCtx(&core.JourneyConfig{
		ID:   "journey-1",
		Name: "Test Journey",
	})

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:  httptest.NewRequest("GET", "/api/v1/journeys/journey-1", nil),
		Response: rec,
		Params:   map[string]string{"id": "journey-1"},
		Workspace: "default",
	}

	err := server.handleGetJourney(ctx)
	if err != nil {
		t.Fatalf("handleGetJourney failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// TestHandleMCPTools tests handleMCPTools
func TestHandleMCPTools(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:  httptest.NewRequest("GET", "/api/v1/mcp/tools", nil),
		Response: rec,
	}

	err := server.handleMCPTools(ctx)
	if err != nil {
		t.Fatalf("handleMCPTools failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected at least one tool")
	}
}

// TestHandleRunJourney tests handleRunJourney
func TestHandleRunJourney(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	// Create a journey first
	store.SaveJourneyNoCtx(&core.JourneyConfig{
		ID:   "journey-1",
		Name: "Test Journey",
		Steps: []core.JourneyStep{
			{Name: "Step 1", Target: "http://example.com"},
		},
	})

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:  httptest.NewRequest("POST", "/api/v1/journeys/journey-1/run", nil),
		Response: rec,
		Params:   map[string]string{"id": "journey-1"},
	}

	err := server.handleRunJourney(ctx)
	if err != nil {
		t.Fatalf("handleRunJourney failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["journey_id"] != "journey-1" {
		t.Errorf("Expected journey_id journey-1, got %s", result["journey_id"])
	}

	if result["status"] != "execution_requested" {
		t.Errorf("Expected status execution_requested, got %s", result["status"])
	}
}

// TestHandleRunJourney_NotFound tests handleRunJourney with non-existent journey
func TestHandleRunJourney_NotFound(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:  httptest.NewRequest("POST", "/api/v1/journeys/nonexistent/run", nil),
		Response: rec,
		Params:   map[string]string{"id": "nonexistent"},
	}

	err := server.handleRunJourney(ctx)
	// Error may be returned or set in context, check both
	if err == nil && rec.Code != http.StatusNotFound {
		t.Error("Expected error for non-existent journey")
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

// TestHandleGetJourney_NotFound tests handleGetJourney with non-existent journey
func TestHandleGetJourney_NotFound(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:  httptest.NewRequest("GET", "/api/v1/journeys/nonexistent", nil),
		Response: rec,
		Params:   map[string]string{"id": "nonexistent"},
		Workspace: "default",
	}

	err := server.handleGetJourney(ctx)
	// Error may be returned or set in context, check both
	if err == nil && rec.Code != http.StatusNotFound {
		t.Errorf("Expected error or not found status, got status %d", rec.Code)
	}

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

// TestHandleDeleteJourney_NotFound tests handleDeleteJourney with non-existent journey
func TestHandleDeleteJourney_NotFound(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:  httptest.NewRequest("DELETE", "/api/v1/journeys/nonexistent", nil),
		Response: rec,
		Params:   map[string]string{"id": "nonexistent"},
	}

	err := server.handleDeleteJourney(ctx)
	// Error may be returned or set in context
	if err == nil && rec.Code != http.StatusNotFound {
		t.Errorf("Expected error or not found status, got status %d", rec.Code)
	}
}

// TestHandleSoulLogs tests handleSoulLogs
func TestHandleSoulLogs(t *testing.T) {
	store := newMockStorage()
	server := newTestServerWithStorage(store)

	rec := httptest.NewRecorder()
	ctx := &Context{
		Request:  httptest.NewRequest("GET", "/api/v1/souls/soul-1/logs", nil),
		Response: rec,
		Params:   map[string]string{"id": "soul-1"},
	}

	err := server.handleSoulLogs(ctx)
	if err != nil {
		t.Fatalf("handleSoulLogs failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(result) == 0 {
		t.Error("Expected at least one log entry")
	}
}

// TestHandleUpdateSoul_InvalidJSON tests update with invalid JSON
func TestHandleUpdateSoul_InvalidJSON(t *testing.T) {
	storage := newMockStorage()
	storage.SaveSoul(nil, &core.Soul{ID: "soul-1", Name: "Test Soul", Type: core.CheckHTTP})

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	// Invalid JSON body
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/souls/soul-1", bytes.NewBufferString("invalid json"))
	ctx := &Context{
		Request:   req,
		Response:  rec,
		Params:    map[string]string{"id": "soul-1"},
		Workspace: "default",
	}

	server.handleUpdateSoul(ctx)
	// Check response code - handlers may set error in context rather than return
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestHandleUpdateSoul_StorageError tests update when storage fails
func TestHandleUpdateSoul_StorageError(t *testing.T) {
	storage := &failingMockStorage{}

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	soul := core.Soul{Name: "Updated Soul", Type: core.CheckHTTP}
	body, _ := json.Marshal(soul)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/souls/soul-1", bytes.NewBuffer(body))
	ctx := &Context{
		Request:   req,
		Response:  rec,
		Params:    map[string]string{"id": "soul-1"},
		Workspace: "default",
	}

	server.handleUpdateSoul(ctx)
	// Check response code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// TestHandleForceCheck_ProbeError tests force check when probe fails
func TestHandleForceCheck_ProbeError(t *testing.T) {
	storage := newMockStorage()
	storage.SaveSoul(nil, &core.Soul{ID: "soul-1", Name: "Test Soul", Type: core.CheckHTTP})

	probe := &mockProbeEngine{forceCheckError: true}

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		probe:   probe,
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/souls/soul-1/check", nil)
	ctx := &Context{
		Request:  req,
		Response: rec,
		Params:   map[string]string{"id": "soul-1"},
	}

	server.handleForceCheck(ctx)
	// Check response code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// TestHandleListJudgments_StorageError tests listing judgments when storage fails
func TestHandleListJudgments_StorageError(t *testing.T) {
	storage := &failingMockStorage{}

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/souls/soul-1/judgments", nil)
	ctx := &Context{
		Request:  req,
		Response: rec,
		Params:   map[string]string{"id": "soul-1"},
	}

	server.handleListJudgments(ctx)
	// Check response code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// TestHandleCreateChannel_InvalidJSON tests creating channel with invalid JSON
func TestHandleCreateChannel_InvalidJSON(t *testing.T) {
	storage := newMockStorage()

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/channels", bytes.NewBufferString("invalid json"))
	ctx := &Context{
		Request:   req,
		Response:  rec,
		Workspace: "default",
	}

	server.handleCreateChannel(ctx)
	// Check response code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestHandleUpdateChannel_InvalidJSON tests updating channel with invalid JSON
func TestHandleUpdateChannel_InvalidJSON(t *testing.T) {
	storage := newMockStorage()
	storage.SaveChannelNoCtx(&core.AlertChannel{ID: "channel-1", Name: "Test Channel", Type: "email"})

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/channels/channel-1", bytes.NewBufferString("invalid json"))
	ctx := &Context{
		Request:   req,
		Response:  rec,
		Params:    map[string]string{"id": "channel-1"},
		Workspace: "default",
	}

	server.handleUpdateChannel(ctx)
	// Check response code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestHandleCreateRule_InvalidJSON tests creating rule with invalid JSON
func TestHandleCreateRule_InvalidJSON(t *testing.T) {
	storage := newMockStorage()

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/rules", bytes.NewBufferString("invalid json"))
	ctx := &Context{
		Request:   req,
		Response:  rec,
		Workspace: "default",
	}

	server.handleCreateRule(ctx)
	// Check response code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestHandleUpdateRule_InvalidJSON tests updating rule with invalid JSON
func TestHandleUpdateRule_InvalidJSON(t *testing.T) {
	storage := newMockStorage()
	storage.SaveRuleNoCtx(&core.AlertRule{ID: "rule-1", Name: "Test Rule", Channels: []string{"channel-1"}})

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/rules/rule-1", bytes.NewBufferString("invalid json"))
	ctx := &Context{
		Request:   req,
		Response:  rec,
		Params:    map[string]string{"id": "rule-1"},
		Workspace: "default",
	}

	server.handleUpdateRule(ctx)
	// Check response code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestHandleCreateWorkspace_InvalidJSON tests creating workspace with invalid JSON
func TestHandleCreateWorkspace_InvalidJSON(t *testing.T) {
	storage := newMockStorage()

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/workspaces", bytes.NewBufferString("invalid json"))
	ctx := &Context{
		Request:  req,
		Response: rec,
	}

	server.handleCreateWorkspace(ctx)
	// Check response code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestHandleCreateStatusPage_InvalidJSON tests creating status page with invalid JSON
func TestHandleCreateStatusPage_InvalidJSON(t *testing.T) {
	storage := newMockStorage()

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/status-pages", bytes.NewBufferString("invalid json"))
	ctx := &Context{
		Request:   req,
		Response:  rec,
		Workspace: "default",
	}

	server.handleCreateStatusPage(ctx)
	// Check response code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestHandleUpdateStatusPage_InvalidJSON tests updating status page with invalid JSON
func TestHandleUpdateStatusPage_InvalidJSON(t *testing.T) {
	storage := newMockStorage()

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/status-pages/page-1", bytes.NewBufferString("invalid json"))
	ctx := &Context{
		Request:   req,
		Response:  rec,
		Params:    map[string]string{"id": "page-1"},
		Workspace: "default",
	}

	server.handleUpdateStatusPage(ctx)
	// Check response code
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

// TestHandleClusterStatus_NotLeader tests cluster status when not leader
func TestHandleClusterStatus_NotLeader(t *testing.T) {
	storage := newMockStorage()

	cluster := &mockClusterManager{}
	cluster.isLeader = false

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: cluster,
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/cluster/status", nil)
	ctx := &Context{
		Request:  req,
		Response: rec,
	}

	err := server.handleClusterStatus(ctx)
	if err != nil {
		t.Fatalf("handleClusterStatus failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// TestHandleClusterPeers_NotLeader tests cluster peers when not leader
func TestHandleClusterPeers_NotLeader(t *testing.T) {
	storage := newMockStorage()

	cluster := &mockClusterManager{}
	cluster.isLeader = false

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: cluster,
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/cluster/peers", nil)
	ctx := &Context{
		Request:  req,
		Response: rec,
	}

	server.handleClusterPeers(ctx)
	// This handler doesn't check for leader status, it just returns peer info
	// The actual API might require leader in real implementation
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

// TestHandleListChannels_StorageError tests listing channels when storage fails
func TestHandleListChannels_StorageError(t *testing.T) {
	// Note: This handler uses s.alert.ListChannels(), not store
	// Skipping as it requires alert manager setup
	t.Skip("Skipped - handler uses alert manager, not storage")
}

// TestHandleListRules_StorageError tests listing rules when storage fails
func TestHandleListRules_StorageError(t *testing.T) {
	// Note: This handler uses s.alert.ListRules(), not store
	// Skipping as it requires alert manager setup
	t.Skip("Skipped - handler uses alert manager, not storage")
}

// TestHandleListWorkspaces_StorageError tests listing workspaces when storage fails
func TestHandleListWorkspaces_StorageError(t *testing.T) {
	storage := &failingMockStorage{}

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/workspaces", nil)
	ctx := &Context{
		Request:  req,
		Response: rec,
	}

	server.handleListWorkspaces(ctx)
	// Check response code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

// TestHandleListStatusPages_StorageError tests listing status pages when storage fails
func TestHandleListStatusPages_StorageError(t *testing.T) {
	storage := &failingMockStorage{}

	router := &Router{routes: make(map[string]map[string]Handler)}
	server := &RESTServer{
		store:   storage,
		router:  router,
		logger:  newTestLogger(),
		cluster: &mockClusterManager{},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/status-pages", nil)
	ctx := &Context{
		Request:  req,
		Response: rec,
	}

	server.handleListStatusPages(ctx)
	// Check response code
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}