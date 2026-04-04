package core

import (
	"os"
	"testing"
	"time"
)

func TestExpandEnvVars(t *testing.T) {
	// Set test env vars
	os.Setenv("TEST_VAR", "hello")
	os.Setenv("TEST_DEFAULT", "")
	defer os.Unsetenv("TEST_VAR")
	defer os.Unsetenv("TEST_DEFAULT")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple variable",
			input:    "value: ${TEST_VAR}",
			expected: "value: hello",
		},
		{
			name:     "variable with default",
			input:    "value: ${TEST_DEFAULT:-world}",
			expected: "value: world",
		},
		{
			name:     "variable without default",
			input:    "value: ${NONEXISTENT:-fallback}",
			expected: "value: fallback",
		},
		{
			name:     "no variables",
			input:    "value: static",
			expected: "value: static",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDurationMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{"seconds", "30s", 30 * time.Second},
		{"minutes", "5m", 5 * time.Minute},
		{"hours", "1h", time.Hour},
		{"complex", "1h30m", 90 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur, err := time.ParseDuration(tt.input)
			if err != nil {
				t.Fatalf("ParseDuration failed: %v", err)
			}

			if dur != tt.expected {
				t.Errorf("Duration(%q) = %v, want %v", tt.input, dur, tt.expected)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	config := &Config{}
	config.setDefaults()

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want %q", config.Server.Host, "0.0.0.0")
	}

	if config.Server.Port != 8443 {
		t.Errorf("Server.Port = %d, want %d", config.Server.Port, 8443)
	}

	if config.Storage.Path != "/var/lib/anubis/data" {
		t.Errorf("Storage.Path = %q, want %q", config.Storage.Path, "/var/lib/anubis/data")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Souls: []Soul{
					{
						Name:   "Test Soul",
						Type:   CheckHTTP,
						Target: "https://example.com",
					},
				},
				Channels: []ChannelConfig{},
				Verdicts: VerdictsConfig{
					Rules: []AlertRule{
						{
							Name:       "Test Rule",
							Conditions: []AlertCondition{{Type: "consecutive_failures", Threshold: 3}},
							Channels:   []string{"channel-1"},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "missing soul name",
			config: &Config{
				Souls: []Soul{
					{
						Type:   CheckHTTP,
						Target: "https://example.com",
					},
				},
			},
			wantError: true,
		},
		{
			name: "missing soul target",
			config: &Config{
				Souls: []Soul{
					{Name: "Test", Type: CheckHTTP},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if (err != nil) != tt.wantError {
				t.Errorf("validate() error = %v, wantError = %v", err, tt.wantError)
			}
		})
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create temp file
	tmpfile, err := os.CreateTemp("", "anubis-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	config := &Config{}
	config.setDefaults()
	config.Server.Host = "localhost"
	config.Server.Port = 9090

	// Save config
	if err := SaveConfig(tmpfile.Name(), config); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Load config
	loaded, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.Server.Host != "localhost" {
		t.Errorf("Loaded Server.Host = %q, want %q", loaded.Server.Host, "localhost")
	}

	if loaded.Server.Port != 9090 {
		t.Errorf("Loaded Server.Port = %d, want %d", loaded.Server.Port, 9090)
	}
}

func TestGenerateDefaultConfig(t *testing.T) {
	config := GenerateDefaultConfig()

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected Server.Host = 0.0.0.0, got %q", config.Server.Host)
	}

	if config.Server.Port != 8443 {
		t.Errorf("Expected Server.Port = 8443, got %d", config.Server.Port)
	}

	if !config.Server.TLS.Enabled {
		t.Error("Expected Server.TLS.Enabled to be true")
	}

	if config.Storage.Path == "" {
		t.Error("Expected Storage.Path to be set")
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"42", 42, false},
		{"0", 0, false},
		{"123456", 123456, false},
		{"", 0, false},
		{"abc", 0, true},
		{"12a34", 0, true},
	}

	for _, tt := range tests {
		result, err := parseInt(tt.input)
		if (err != nil) != tt.hasError {
			t.Errorf("parseInt(%q) error = %v, hasError = %v", tt.input, err, tt.hasError)
		}
		if !tt.hasError && result != tt.expected {
			t.Errorf("parseInt(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestRaftConfig_Validate(t *testing.T) {
	// Valid config
	cfg := &RaftConfig{
		NodeID:        "node-1",
		BindAddr:      "127.0.0.1:7000",
		AdvertiseAddr: "127.0.0.1:7000",
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}

	// Missing NodeID
	cfg = &RaftConfig{
		BindAddr: "127.0.0.1:7000",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for missing NodeID")
	}

	// Missing BindAddr
	cfg = &RaftConfig{
		NodeID: "node-1",
	}
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for missing BindAddr")
	}

	// Empty AdvertiseAddr should default to BindAddr
	cfg = &RaftConfig{
		NodeID:   "node-1",
		BindAddr: "127.0.0.1:7000",
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
	if cfg.AdvertiseAddr != "127.0.0.1:7000" {
		t.Errorf("Expected AdvertiseAddr to default to BindAddr")
	}
}
