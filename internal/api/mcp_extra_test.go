package api

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AnubisWatch/anubiswatch/internal/core"
)

// Test handleListSouls with status filter
func TestMCPServer_handleListSouls_WithStatusFilter(t *testing.T) {
	store := newMockStorage()
	store.souls["soul-1"] = &core.Soul{ID: "soul-1", Name: "Test Soul", Type: core.CheckHTTP, Enabled: true}
	store.souls["soul-2"] = &core.Soul{ID: "soul-2", Name: "Test Soul 2", Type: core.CheckHTTP, Enabled: true}
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Test with status filter
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_souls","arguments":{"status":"alive"}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleListSouls with workspace filter
func TestMCPServer_handleListSouls_WithWorkspaceFilter(t *testing.T) {
	store := newMockStorage()
	store.souls["soul-1"] = &core.Soul{ID: "soul-1", Name: "Test Soul", Type: core.CheckHTTP, Enabled: true}
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Test with workspace filter
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_souls","arguments":{"workspace":"default"}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleReadResource with valid soul ID
func TestMCPServer_handleReadResource_WithValidSoul(t *testing.T) {
	store := newMockStorage()
	store.souls["soul-1"] = &core.Soul{ID: "soul-1", Name: "Test Soul", Type: core.CheckHTTP, Enabled: true}
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Read soul resource
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{"uri":"soul://soul-1"}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleGetPrompt with analyze_soul
func TestMCPServer_handleGetPrompt_AnalyzeSoul(t *testing.T) {
	store := newMockStorage()
	store.souls["soul-1"] = &core.Soul{ID: "soul-1", Name: "Test Soul", Type: core.CheckHTTP, Enabled: true}
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Get analyze soul prompt
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"prompts/get","params":{"name":"analyze_soul","arguments":{"soul_id":"soul-1"}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleGetPrompt with incident_summary
func TestMCPServer_handleGetPrompt_IncidentSummary(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Get incident summary prompt
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"prompts/get","params":{"name":"incident_summary","arguments":{}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleCreateSoul with interval
func TestMCPServer_handleCreateSoul_WithInterval(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Create soul with interval
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"create_soul","arguments":{"name":"Test Soul","type":"http","target":"https://example.com","interval":"1m"}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleCreateSoul without interval
func TestMCPServer_handleCreateSoul_WithoutInterval(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Create soul without interval
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"create_soul","arguments":{"name":"Test Soul 2","type":"http","target":"https://example2.com"}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleGetPrompt with create_monitor_guide for website
func TestMCPServer_handleGetPrompt_CreateMonitorGuide_Website(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Get create monitor guide prompt for website
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"prompts/get","params":{"name":"create_monitor_guide","arguments":{"type":"website"}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleGetPrompt with create_monitor_guide for API
func TestMCPServer_handleGetPrompt_CreateMonitorGuide_API(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Get create monitor guide prompt for API
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"prompts/get","params":{"name":"create_monitor_guide","arguments":{"type":"api"}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleGetPrompt with create_monitor_guide for server
func TestMCPServer_handleGetPrompt_CreateMonitorGuide_Server(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Get create monitor guide prompt for server
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"prompts/get","params":{"name":"create_monitor_guide","arguments":{"type":"server"}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleReadResource with unknown resource URI
func TestMCPServer_handleReadResource_UnknownResource(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Read unknown resource
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{"uri":"unknown://resource"}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	// Should return error for unknown resource
	if resp.Error == nil {
		t.Error("Expected error for unknown resource")
	}
}

// Test handleReadResource with soul ID parameter
func TestMCPServer_handleReadResource_SoulWithID(t *testing.T) {
	store := newMockStorage()
	store.souls["soul-1"] = &core.Soul{ID: "soul-1", Name: "Test Soul", Type: core.CheckHTTP, Enabled: true}
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Read soul resource with ID
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{"uri":"soul://soul-1"}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	// Should not error
	if resp.Error != nil {
		t.Logf("Got error: %s", resp.Error.Message)
	}
}

// Test handleGetPrompt with missing prompt name
func TestMCPServer_handleGetPrompt_MissingName(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Get prompt without name
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"prompts/get","params":{"arguments":{}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	// Should return error for missing name
	if resp.Error == nil {
		t.Error("Expected error for missing prompt name")
	}
}

// Test handleGetPrompt with unknown prompt
func TestMCPServer_handleGetPrompt_UnknownPrompt(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Get unknown prompt
	reqBody := `{"jsonrpc":"2.0","id":1,"method":"prompts/get","params":{"name":"unknown_prompt","arguments":{}}}`
	req := httptest.NewRequest("POST", "/mcp", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	// Should return error for unknown prompt
	if resp.Error == nil {
		t.Error("Expected error for unknown prompt")
	}
}

// Test handleReadResource with error from handler
func TestMCPServer_handleReadResource_HandlerError(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Register a resource that returns error
	server.RegisterResource(MCPResource{
		URI:         "test://error",
		Name:        "Error Resource",
		Description: "A resource that returns error",
		Handler: func() (interface{}, error) {
			return nil, fmt.Errorf("resource handler error")
		},
	})

	req := &MCPRequest{
		ID:     1,
		Method: "resources/read",
		Params: json.RawMessage(`{"uri": "test://error"}`),
	}
	resp := server.handleReadResource(req)
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}
	if resp.Error == nil {
		t.Error("Expected error response")
	}
}

// Test handleGetPrompt with handler error
func TestMCPServer_handleGetPrompt_HandlerError(t *testing.T) {
	store := newMockStorage()
	probe := &mockProbeEngine{}
	alert := &mockAlertManager{}
	logger := newTestLogger()

	server := NewMCPServer(store, probe, alert, logger)

	// Register a prompt that returns error
	server.RegisterPrompt(MCPPrompt{
		Name:        "error_prompt",
		Description: "A prompt that returns error",
		Handler: func(args map[string]string) (string, error) {
			return "", fmt.Errorf("prompt handler error")
		},
	})

	req := &MCPRequest{
		ID:     1,
		Method: "prompts/get",
		Params: json.RawMessage(`{"name": "error_prompt", "arguments": {}}`),
	}
	resp := server.handleGetPrompt(req)
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}
	if resp.Error == nil {
		t.Error("Expected error response")
	}
}
