package auth

import (
	"testing"
)

func TestLocalAuthenticator(t *testing.T) {
	auth := NewLocalAuthenticator()

	// Test login
	user, token, err := auth.Login("test@example.com", "password")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if user == nil {
		t.Fatal("User should not be nil")
	}

	if token == "" {
		t.Fatal("Token should not be empty")
	}

	// Test authenticate with valid token
	authUser, err := auth.Authenticate(token)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	if authUser == nil {
		t.Fatal("Authenticated user should not be nil")
	}

	// Test logout
	if err := auth.Logout(token); err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Token should be invalid after logout
	_, err = auth.Authenticate(token)
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestInvalidCredentials(t *testing.T) {
	auth := NewLocalAuthenticator()

	_, _, err := auth.Login("", "")
	if err == nil {
		t.Error("Expected error for empty credentials")
	}
}

// TestLogin_EmptyEmail tests login with empty email
func TestLogin_EmptyEmail(t *testing.T) {
	auth := NewLocalAuthenticator()

	_, _, err := auth.Login("", "password")
	if err == nil {
		t.Error("Expected error for empty email")
	}
}

// TestLogin_EmptyPassword tests login with empty password
func TestLogin_EmptyPassword(t *testing.T) {
	auth := NewLocalAuthenticator()

	_, _, err := auth.Login("test@example.com", "")
	if err == nil {
		t.Error("Expected error for empty password")
	}
}

// TestLogin_RepeatedLoginSameUser tests that repeated logins with same email return same user
func TestLogin_RepeatedLoginSameUser(t *testing.T) {
	auth := NewLocalAuthenticator()

	user1, token1, err := auth.Login("same@example.com", "password1")
	if err != nil {
		t.Fatalf("First login failed: %v", err)
	}

	user2, token2, err := auth.Login("same@example.com", "password2")
	if err != nil {
		t.Fatalf("Second login failed: %v", err)
	}

	if user1.ID != user2.ID {
		t.Error("Expected same user ID for repeated logins")
	}

	if token1 == token2 {
		t.Error("Expected different tokens for separate logins")
	}
}

// TestLogout_NonExistentToken tests logout with non-existent token
func TestLogout_NonExistentToken(t *testing.T) {
	auth := NewLocalAuthenticator()

	err := auth.Logout("non-existent-token")
	if err != nil {
		t.Errorf("Logout should not error for non-existent token: %v", err)
	}
}

// TestAuthenticate_NonExistentToken tests authenticate with non-existent token
func TestAuthenticate_NonExistentToken(t *testing.T) {
	auth := NewLocalAuthenticator()

	_, err := auth.Authenticate("non-existent-token")
	if err == nil {
		t.Error("Expected error for non-existent token")
	}
}
