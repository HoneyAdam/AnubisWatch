package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AnubisWatch/anubiswatch/internal/api"
	"github.com/AnubisWatch/anubiswatch/internal/core"
)

// OIDCAuthenticator implements OIDC authentication with local fallback
type OIDCAuthenticator struct {
	mu      sync.RWMutex
	config  core.OIDCAuth
	local   *LocalAuthenticator
	state   map[string]*oidcState // CSRF state tracking
	users   map[string]*api.User  // email -> user
	tokens  map[string]*session   // token -> session
	stopCh  chan struct{}
}

type oidcState struct {
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}

type oidcConfig struct {
	Issuer        string   `json:"issuer"`
	AuthURL       string   `json:"authorization_endpoint"`
	TokenURL      string   `json:"token_endpoint"`
	UserInfoURL   string   `json:"userinfo_endpoint"`
	JWKSURI       string   `json:"jwks_uri"`
	Scopes        []string `json:"scopes_supported"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type userInfoResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
}

// NewOIDCAuthenticator creates a new OIDC authenticator with local fallback
func NewOIDCAuthenticator(cfg core.OIDCAuth, localPath, adminEmail, adminPassword string) *OIDCAuthenticator {
	localAuth := NewLocalAuthenticator(localPath, adminEmail, adminPassword)
	return &OIDCAuthenticator{
		config: cfg,
		local:  localAuth,
		state:  make(map[string]*oidcState),
		users:  make(map[string]*api.User),
		tokens: make(map[string]*session),
		stopCh: make(chan struct{}),
	}
}

// OIDCLoginURL returns the OIDC authorization URL for redirect
func (o *OIDCAuthenticator) OIDCLoginURL() (string, string, error) {
	// Fetch OIDC configuration from issuer
	cfg, err := o.fetchOIDCConfig()
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch OIDC config: %w", err)
	}

	// Generate CSRF state
	state := generateToken()
	expiresAt := time.Now().Add(10 * time.Minute)

	o.mu.Lock()
	o.state[state] = &oidcState{ExpiresAt: expiresAt}
	o.mu.Unlock()

	// Build authorization URL
	redirectURL := cfg.AuthURL + "?" +
		"response_type=code" +
		"&client_id=" + o.config.ClientID +
		"&redirect_uri=" + o.config.RedirectURL +
		"&scope=openid+profile+email" +
		"&state=" + state

	return redirectURL, state, nil
}

// OIDCCallback handles the OIDC callback
func (o *OIDCAuthenticator) OIDCCallback(code, state string) (*api.User, string, error) {
	// Verify state (CSRF protection)
	o.mu.Lock()
	csrfState, exists := o.state[state]
	if !exists {
		o.mu.Unlock()
		return nil, "", fmt.Errorf("invalid state")
	}
	if time.Now().After(csrfState.ExpiresAt) {
		delete(o.state, state)
		o.mu.Unlock()
		return nil, "", fmt.Errorf("state expired")
	}
	delete(o.state, state)
	o.mu.Unlock()

	// Exchange code for token
	tokenResp, err := o.exchangeCode(code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info
	userInfo, err := o.getUserInfo(tokenResp.AccessToken)
	if err != nil {
		// Try ID token claims as fallback
		userInfo, err = o.parseIDToken(tokenResp.IDToken)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get user info: %w", err)
		}
	}

	if userInfo.Email == "" {
		return nil, "", fmt.Errorf("no email in OIDC response")
	}

	// Create or find user
	o.mu.Lock()
	user, exists := o.users[userInfo.Email]
	if !exists {
		user = &api.User{
			ID:        generateID(),
			Email:     userInfo.Email,
			Name:      userInfo.Name,
			Role:      "viewer", // Default role for OIDC users
			Workspace: "default",
			CreatedAt: time.Now(),
		}
		o.users[user.ID] = user
		o.users[user.Email] = user // Index by email for lookup
	}
	o.mu.Unlock()

	// Generate session token
	token := generateToken()
	o.mu.Lock()
	o.tokens[token] = &session{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	o.mu.Unlock()

	return user, token, nil
}

// fetchOIDCConfig fetches the OIDC discovery document
func (o *OIDCAuthenticator) fetchOIDCConfig() (*oidcConfig, error) {
	// Normalize issuer URL (remove trailing slash)
	issuer := strings.TrimSuffix(o.config.Issuer, "/")

	// Try well-known endpoint
	wellKnownURL := issuer + "/.well-known/openid-configuration"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(wellKnownURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OIDC config from %s: %w", wellKnownURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC config endpoint returned %d", resp.StatusCode)
	}

	var cfg oidcConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse OIDC config: %w", err)
	}

	if cfg.AuthURL == "" || cfg.TokenURL == "" {
		return nil, fmt.Errorf("OIDC config missing authorization or token endpoint")
	}

	return &cfg, nil
}

// exchangeCode exchanges an authorization code for tokens
func (o *OIDCAuthenticator) exchangeCode(code string) (*tokenResponse, error) {
	cfg, err := o.fetchOIDCConfig()
	if err != nil {
		return nil, err
	}

	// Build token request
	data := fmt.Sprintf(
		"grant_type=authorization_code&code=%s&redirect_uri=%s&client_id=%s&client_secret=%s",
		code, o.config.RedirectURL, o.config.ClientID, o.config.ClientSecret,
	)

	resp, err := http.Post(cfg.TokenURL, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// getUserInfo fetches user info from OIDC provider
func (o *OIDCAuthenticator) getUserInfo(accessToken string) (*userInfoResponse, error) {
	cfg, err := o.fetchOIDCConfig()
	if err != nil {
		return nil, err
	}

	if cfg.UserInfoURL == "" {
		return nil, fmt.Errorf("no userinfo endpoint in OIDC config")
	}

	req, err := http.NewRequest("GET", cfg.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned %d", resp.StatusCode)
	}

	var userInfo userInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo: %w", err)
	}

	return &userInfo, nil
}

// parseIDToken parses a JWT ID token to extract claims (simplified, no signature verification)
func (o *OIDCAuthenticator) parseIDToken(idToken string) (*userInfoResponse, error) {
	if idToken == "" {
		return nil, fmt.Errorf("empty ID token")
	}

	// JWT format: header.payload.signature
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode payload (base64url)
	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// Try with padding
		padded := parts[1] + strings.Repeat("=", (4-len(parts[1])%4)%4)
		decoded, err = base64.RawURLEncoding.DecodeString(padded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
		}
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims JSON: %w", err)
	}

	userInfo := &userInfoResponse{}
	if sub, ok := claims["sub"].(string); ok {
		userInfo.Sub = sub
	}
	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}
	if name, ok := claims["name"].(string); ok {
		userInfo.Name = name
	}
	if verified, ok := claims["email_verified"].(bool); ok {
		userInfo.EmailVerified = verified
	}

	return userInfo, nil
}

// Authenticate validates a token and returns the user
func (o *OIDCAuthenticator) Authenticate(token string) (*api.User, error) {
	// Check OIDC tokens first
	o.mu.RLock()
	sess, ok := o.tokens[token]
	if ok {
		if time.Now().After(sess.ExpiresAt) {
			o.mu.RUnlock()
			return nil, fmt.Errorf("token expired")
		}
		user := o.users[sess.UserID]
		o.mu.RUnlock()
		if user == nil {
			return nil, fmt.Errorf("user not found")
		}
		return user, nil
	}
	o.mu.RUnlock()

	// Fallback to local auth
	return o.local.Authenticate(token)
}

// Login validates credentials and creates a session (local fallback)
func (o *OIDCAuthenticator) Login(email, password string) (*api.User, string, error) {
	return o.local.Login(email, password)
}

// Logout invalidates a token
func (o *OIDCAuthenticator) Logout(token string) error {
	o.mu.Lock()
	delete(o.tokens, token)
	o.mu.Unlock()
	return o.local.Logout(token)
}

// Shutdown gracefully stops the authenticator
func (o *OIDCAuthenticator) Shutdown() {
	o.local.Shutdown()
	close(o.stopCh)
}

// GetUsers returns all OIDC users
func (o *OIDCAuthenticator) GetUsers() []*api.User {
	o.mu.RLock()
	defer o.mu.RUnlock()

	users := make([]*api.User, 0, len(o.users))
	for _, user := range o.users {
		users = append(users, user)
	}
	return users
}

// AddUser adds a user (for admin management)
func (o *OIDCAuthenticator) AddUser(email, name, role string) *api.User {
	o.mu.Lock()
	defer o.mu.Unlock()

	user := &api.User{
		ID:        generateID(),
		Email:     email,
		Name:      name,
		Role:      role,
		Workspace: "default",
		CreatedAt: time.Now(),
	}
	o.users[user.ID] = user
	o.users[user.Email] = user
	return user
}
