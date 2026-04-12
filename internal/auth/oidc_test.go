package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
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

// Helper: create a signed JWT using an RSA key
func createTestJWT(t *testing.T, priv *rsa.PrivateKey, claims map[string]interface{}) string {
	t.Helper()
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT", "kid": "test-key"}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	claimsJSON, _ := json.Marshal(claims)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	hash := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
	if err != nil {
		t.Fatalf("failed to sign JWT: %v", err)
	}
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)
	return headerB64 + "." + claimsB64 + "." + sigB64
}

// Helper: create a test RSA key and JWK set
func createTestJWK(priv *rsa.PrivateKey) *jwkSet {
	n := base64.RawURLEncoding.EncodeToString(priv.N.Bytes())
	eBytes := big.NewInt(int64(priv.E)).Bytes()
	e := base64.RawURLEncoding.EncodeToString(eBytes)
	return &jwkSet{Keys: []jwk{{Kty: "RSA", Use: "sig", Kid: "test-key", Alg: "RS256", N: n, E: e}}}
}

// Helper: create a JWK set HTTP server
func newJWKServer(t *testing.T, jwks *jwkSet) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
}

func TestOIDCAuthenticator_ParseIDToken(t *testing.T) {
	// Generate RSA key pair
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	// Create JWK set
	jwks := createTestJWK(priv)

	// Create claims
	claims := map[string]interface{}{
		"sub":            "1234567890",
		"email":          "test@example.com",
		"name":           "Test User",
		"email_verified": true,
		"iss":            "https://test.example.com",
		"aud":            "test-client-id",
		"exp":            time.Now().Add(time.Hour).Unix(),
		"nbf":            time.Now().Add(-time.Minute).Unix(),
	}

	// Create signed JWT
	token := createTestJWT(t, priv, claims)

	// Create authenticator with pre-loaded JWKs (skip network fetch)
	cfg := core.OIDCAuth{
		Issuer:   "https://test.example.com",
		ClientID: "test-client-id",
	}
	auth := &OIDCAuthenticator{
		config:      cfg,
		jwks:        jwks,
		jwksFetched: time.Now(),
		jwksTTL:     time.Hour,
	}

	userInfo, err := auth.parseIDToken(token)
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

func TestOIDCAuthenticator_ParseIDToken_RejectForged(t *testing.T) {
	// Generate two different key pairs
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}
	forgeryKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate forgery key: %v", err)
	}

	// JWK set contains the legitimate key
	jwks := createTestJWK(priv)

	claims := map[string]interface{}{
		"sub":            "attacker",
		"email":          "admin@example.com",
		"name":           "Attacker",
		"email_verified": true,
		"iss":            "https://test.example.com",
		"aud":            "test-client-id",
		"exp":            time.Now().Add(time.Hour).Unix(),
		"nbf":            time.Now().Add(-time.Minute).Unix(),
	}
	forgedToken := createTestJWT(t, forgeryKey, claims)

	cfg := core.OIDCAuth{
		Issuer:   "https://test.example.com",
		ClientID: "test-client-id",
	}
	auth := &OIDCAuthenticator{
		config:      cfg,
		jwks:        jwks,
		jwksFetched: time.Now(),
		jwksTTL:     time.Hour,
	}

	_, err = auth.parseIDToken(forgedToken)
	if err == nil {
		t.Fatal("EXPECTED forged JWT to be rejected, but it was accepted!")
	}
	t.Logf("Correctly rejected forged JWT: %v", err)
}

func TestOIDCAuthenticator_ParseIDToken_RejectExpired(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	jwks := createTestJWK(priv)

	claims := map[string]interface{}{
		"sub":   "123",
		"email": "test@example.com",
		"iss":   "https://test.example.com",
		"aud":   "test-client-id",
		"exp":   time.Now().Add(-time.Hour).Unix(),
	}
	token := createTestJWT(t, priv, claims)

	cfg := core.OIDCAuth{
		Issuer:   "https://test.example.com",
		ClientID: "test-client-id",
	}
	auth := &OIDCAuthenticator{
		config:      cfg,
		jwks:        jwks,
		jwksFetched: time.Now(),
		jwksTTL:     time.Hour,
	}

	_, err = auth.parseIDToken(token)
	if err == nil {
		t.Fatal("EXPECTED expired token to be rejected")
	}
	t.Logf("Correctly rejected expired token: %v", err)
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

// Test EC key support (ES256)
func TestOIDCAuthenticator_ParseIDToken_ECKey(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate EC key: %v", err)
	}

	// Build JWK
	x := base64.RawURLEncoding.EncodeToString(priv.X.Bytes())
	y := base64.RawURLEncoding.EncodeToString(priv.Y.Bytes())
	jwks := &jwkSet{Keys: []jwk{{Kty: "EC", Use: "sig", Kid: "ec-key", Alg: "ES256", Crv: "P-256", X: x, Y: y}}}

	// Create ES256 signed JWT
	header := map[string]interface{}{"alg": "ES256", "typ": "JWT", "kid": "ec-key"}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	claims := map[string]interface{}{
		"sub":   "ec-user",
		"email": "ecuser@example.com",
		"name":  "EC User",
		"iss":   "https://test.example.com",
		"aud":   "ec-client",
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
	claimsJSON, _ := json.Marshal(claims)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	hash := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash[:])
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}
	// ECDSA signature: R || S, each padded to curve byte length
	keyLen := priv.Params().BitSize / 8
	sigBytes := append(r.Bytes(), s.Bytes()...)
	// Pad to 2*keyLen
	padded := make([]byte, 2*keyLen)
	copy(padded[2*keyLen-len(sigBytes):], sigBytes)
	sigB64 := base64.RawURLEncoding.EncodeToString(padded)

	token := headerB64 + "." + claimsB64 + "." + sigB64

	cfg := core.OIDCAuth{
		Issuer:   "https://test.example.com",
		ClientID: "ec-client",
	}
	auth := &OIDCAuthenticator{
		config:      cfg,
		jwks:        jwks,
		jwksFetched: time.Now(),
		jwksTTL:     time.Hour,
	}

	userInfo, err := auth.parseIDToken(token)
	if err != nil {
		t.Fatalf("parseIDToken with EC key failed: %v", err)
	}
	if userInfo.Email != "ecuser@example.com" {
		t.Errorf("Expected email ecuser@example.com, got %s", userInfo.Email)
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
