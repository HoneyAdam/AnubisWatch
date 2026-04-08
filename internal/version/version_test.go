package version

import (
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	if info.Platform == "" {
		t.Error("Platform should not be empty")
	}

	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
}

func TestGetExtended(t *testing.T) {
	info := GetExtended()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	// NumCPU and NumGoroutine should be populated
	if info.NumCPU <= 0 {
		t.Error("NumCPU should be positive")
	}
}

func TestInfoString(t *testing.T) {
	info := Info{
		Version:   "1.0.0",
		Commit:    "abc123def456",
		BuildDate: "2026-01-01",
		GoVersion: "go1.26",
		Platform:  "linux/amd64",
	}

	s := info.String()

	if !strings.Contains(s, "1.0.0") {
		t.Error("String should contain version")
	}

	if !strings.Contains(s, "linux/amd64") {
		t.Error("String should contain platform")
	}
}

func TestInfoShort(t *testing.T) {
	info := Info{
		Version: "1.0.0",
		Commit:  "abc123def456",
	}

	short := info.Short()
	expected := "1.0.0-abc123d"

	if short != expected {
		t.Errorf("Expected short version %s, got %s", expected, short)
	}
}

func TestIsDev(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"dev", true},
		{"1.0.0-dirty", true},
		{"1.0.0", false},
		{"v1.2.3", false},
	}

	for _, tt := range tests {
		info := Info{Version: tt.version}
		if got := info.IsDev(); got != tt.want {
			t.Errorf("IsDev() for version %q = %v, want %v", tt.version, got, tt.want)
		}
	}
}

func TestIsRelease(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"1.0.0", true},
		{"v1.2.3", true},
		{"dev", false},
		{"1.0.0-alpha", false},
		{"1.0.0-dirty", false},
	}

	for _, tt := range tests {
		info := Info{Version: tt.version}
		if got := info.IsRelease(); got != tt.want {
			t.Errorf("IsRelease() for version %q = %v, want %v", tt.version, got, tt.want)
		}
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.1.0", "1.0.0", 1},
		{"2.0.0", "1.9.9", 1},
		{"v1.0.0", "1.0.0", 0},
		{"1.0.0-alpha", "1.0.0", 0}, // Same version parts
	}

	for _, tt := range tests {
		got := Compare(tt.v1, tt.v2)
		if got != tt.want {
			t.Errorf("Compare(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
		}
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		version string
		want    [3]int
	}{
		{"1.2.3", [3]int{1, 2, 3}},
		{"v1.2.3", [3]int{1, 2, 3}},
		{"1.2", [3]int{1, 2, 0}},
		{"1.2.3-alpha", [3]int{1, 2, 3}},
		{"invalid", [3]int{0, 0, 0}},
	}

	for _, tt := range tests {
		got := parseVersion(tt.version)
		if got != tt.want {
			t.Errorf("parseVersion(%q) = %v, want %v", tt.version, got, tt.want)
		}
	}
}
