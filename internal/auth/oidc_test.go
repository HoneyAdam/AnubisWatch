package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AnubisWatch/anubiswatch/internal/core"
)

func TestOIDCAuthenticator_NewOIDCAuthenticator(t *testing.T) {
	cfg := core.OIDCAuth{
		Issuer:       "https://accounts.google.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/api/v1/auth/oidc/callback",
	}

	auth := NewOIDCAuthenticator(cfg, "", "admin@test.com", "admin123")
	defer auth.Shutdown()

	if auth == nil {
		t.Fatal("NewOIDCAuthenticator returned nil")
	}

	if auth.config.ClientID != cfg.ClientID {
		t.Errorf("Expected ClientID %s, got %s", cfg.ClientID, auth.config.ClientID)
	}
}

func TestOIDCAuthenticator_LocalFallback(t *testing.T) {
	cfg := core.OIDCAuth{
		Issuer:       "https://example.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	auth := NewOIDCAuthenticator(cfg, "", "admin@test.com", "admin123")
	defer auth.Shutdown()

	// Test local fallback login
	user, token, err := auth.Login("admin@test.com", "admin123")
	if err != nil {
		t.Fatalf("Local fallback Login failed: %v", err)
	}

	if user.Email != "admin@test.com" {
		t.Errorf("Expected email admin@test.com, got %s", user.Email)
	}

	if token == "" {
		t.Fatal("Token should not be empty")
	}

	// Test authenticate with the token
	authUser, err := auth.Authenticate(token)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	if authUser.Email != "admin@test.com" {
		t.Errorf("Expected authenticated email admin@test.com, got %s", authUser.Email)
	}
}

func TestOIDCAuthenticator_AddUser(t *testing.T) {
	cfg := core.OIDCAuth{
		Issuer:       "https://example.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	auth := NewOIDCAuthenticator(cfg, "", "admin@test.com", "admin123")
	defer auth.Shutdown()

	user := auth.AddUser("user@example.com", "Test User", "editor")
	if user == nil {
		t.Fatal("AddUser returned nil")
	}

	if user.Email != "user@example.com" {
		t.Errorf("Expected email user@example.com, got %s", user.Email)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got %s", user.Name)
	}

	if user.Role != "editor" {
		t.Errorf("Expected role 'editor', got %s", user.Role)
	}

	// Verify user is in the list
	users := auth.GetUsers()
	if len(users) == 0 {
		t.Error("Expected at least one user")
	}
}

func TestOIDCAuthenticator_OIDCCallback_InvalidState(t *testing.T) {
	cfg := core.OIDCAuth{
		Issuer:       "https://example.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	auth := NewOIDCAuthenticator(cfg, "", "admin@test.com", "admin123")
	defer auth.Shutdown()

	// Test callback with invalid state (no matching state was generated)
	_, _, err := auth.OIDCCallback("test-code", "invalid-state")
	if err == nil {
		t.Error("Expected error for invalid state")
	}
}

func TestOIDCAuthenticator_OIDCLoginURL_ConfigError(t *testing.T) {
	cfg := core.OIDCAuth{
		Issuer:       "https://nonexistent.invalid.domain",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	auth := NewOIDCAuthenticator(cfg, "", "admin@test.com", "admin123")
	defer auth.Shutdown()

	// Should fail because the issuer domain doesn't exist
	_, _, err := auth.OIDCLoginURL()
	if err == nil {
		t.Error("Expected error for invalid OIDC issuer")
	}
}

func TestOIDCAuthenticator_TokenExpiration(t *testing.T) {
	cfg := core.OIDCAuth{
		Issuer:       "https://example.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	auth := NewOIDCAuthenticator(cfg, "", "admin@test.com", "admin123")
	defer auth.Shutdown()

	// Login and get token
	_, token, err := auth.Login("admin@test.com", "admin123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Token should be valid
	_, err = auth.Authenticate(token)
	if err != nil {
		t.Fatalf("Token should be valid: %v", err)
	}

	// Logout
	err = auth.Logout(token)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Token should be invalid after logout
	_, err = auth.Authenticate(token)
	if err == nil {
		t.Error("Expected error for logged out token")
	}
}

func TestOIDCAuthenticator_ParseIDToken(t *testing.T) {
	auth := &OIDCAuthenticator{}

	// Valid JWT-like payload (base64url encoded JSON)
	// {"sub":"1234567890","email":"test@example.com","name":"Test User","email_verified":true}
	// base64url: eyJzdWIiOiIxMjM0NTY3ODkwIiwiZW1haWwiOiJ0ZXN0QGV4YW1wbGUuY29tIiwibmFtZSI6IlRlc3QgVXNlciIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlfQ
	validToken := "header.eyJzdWIiOiIxMjM0NTY3ODkwIiwiZW1haWwiOiJ0ZXN0QGV4YW1wbGUuY29tIiwibmFtZSI6IlRlc3QgVXNlciIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlfQ.sig"

	userInfo, err := auth.parseIDToken(validToken)
	if err != nil {
		t.Fatalf("parseIDToken failed: %v", err)
	}

	if userInfo.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", userInfo.Email)
	}

	if userInfo.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got %s", userInfo.Name)
	}

	if userInfo.Sub != "1234567890" {
		t.Errorf("Expected sub '1234567890', got %s", userInfo.Sub)
	}

	if !userInfo.EmailVerified {
		t.Error("Expected email_verified to be true")
	}
}

func TestOIDCAuthenticator_ParseIDToken_Invalid(t *testing.T) {
	auth := &OIDCAuthenticator{}

	// Empty token
	_, err := auth.parseIDToken("")
	if err == nil {
		t.Error("Expected error for empty token")
	}

	// Invalid format (not 3 parts)
	_, err = auth.parseIDToken("not.a.valid.jwt")
	if err == nil {
		t.Error("Expected error for invalid JWT format")
	}

	// Invalid base64
	_, err = auth.parseIDToken("header.!!!invalid!!!.sig")
	if err == nil {
		t.Error("Expected error for invalid base64")
	}
}

func TestOIDCAuthenticator_GetUsers(t *testing.T) {
	cfg := core.OIDCAuth{
		Issuer:       "https://example.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	auth := NewOIDCAuthenticator(cfg, "", "admin@test.com", "admin123")
	defer auth.Shutdown()

	// Add some users
	auth.AddUser("user1@example.com", "User One", "viewer")
	auth.AddUser("user2@example.com", "User Two", "editor")
	auth.AddUser("user3@example.com", "User Three", "admin")

	users := auth.GetUsers()
	if len(users) < 3 {
		t.Errorf("Expected at least 3 users, got %d", len(users))
	}
}

func TestOIDCAuthenticator_OIDCCallback_StateExpiration(t *testing.T) {
	cfg := core.OIDCAuth{
		Issuer:       "https://example.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	auth := NewOIDCAuthenticator(cfg, "", "admin@test.com", "admin123")
	defer auth.Shutdown()

	// Manually add an expired state
	auth.mu.Lock()
	auth.state["expired-state"] = &oidcState{
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	auth.mu.Unlock()

	_, _, err := auth.OIDCCallback("test-code", "expired-state")
	if err == nil {
		t.Error("Expected error for expired state")
	}
}

func TestOIDCAuthenticator_OIDCCallback_OIDCError(t *testing.T) {
	cfg := core.OIDCAuth{
		Issuer:       "https://example.com",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
	}

	// Create a mock server that returns an OIDC error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_grant","error_description":"code expired"}`))
	}))
	defer server.Close()

	// Use the mock server as issuer (it won't work but tests the error path)
	cfg.Issuer = server.URL
	auth := NewOIDCAuthenticator(cfg, "", "admin@test.com", "admin123")
	defer auth.Shutdown()

	// Generate a valid state
	loginURL, state, _ := auth.OIDCLoginURL()
	if loginURL == "" {
		// OIDC config fetch failed, which is expected for the mock server
		t.Log("OIDC login URL fetch failed as expected for mock server")
		return
	}

	// Try callback with valid state (the token exchange will fail)
	_, _, err := auth.OIDCCallback("test-code", state)
	if err == nil {
		t.Error("Expected error from failed token exchange")
	}
}
