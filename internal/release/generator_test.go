package release

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewGenerator(t *testing.T) {
	logger := newTestLogger()
	gen := NewGenerator("1.0.0", "abc123", "test-output", logger)

	if gen == nil {
		t.Fatal("Expected generator to be created")
	}

	if gen.version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", gen.version)
	}
}

func TestCalculateChecksum(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Hello, AnubisWatch!")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	logger := newTestLogger()
	gen := NewGenerator("1.0.0", "abc123", tmpDir, logger)

	checksum, err := gen.calculateChecksum(testFile)
	if err != nil {
		t.Fatalf("Failed to calculate checksum: %v", err)
	}

	if len(checksum) != 64 {
		t.Errorf("Expected SHA256 checksum (64 hex chars), got %d chars", len(checksum))
	}

	// Verify deterministic
	checksum2, _ := gen.calculateChecksum(testFile)
	if checksum != checksum2 {
		t.Error("Checksum should be deterministic")
	}
}

func TestFindArtifacts(t *testing.T) {
	tmpDir := t.TempDir()
	logger := newTestLogger()
	gen := NewGenerator("1.0.0", "abc123", tmpDir, logger)

	// Create test artifacts
	artifacts := []string{
		"anubis-linux-amd64",
		"anubis-linux-amd64.tar.gz",
		"anubis-windows-amd64.exe",
		"checksums.txt", // Should be excluded
		"readme.md",     // Should be excluded
	}

	for _, name := range artifacts {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create artifact: %v", err)
		}
	}

	found, err := gen.findArtifacts(tmpDir)
	if err != nil {
		t.Fatalf("Failed to find artifacts: %v", err)
	}

	// Should find 3 artifacts (excluding checksums.txt and readme.md)
	if len(found) != 3 {
		t.Errorf("Expected 3 artifacts, found %d", len(found))
	}
}

func TestIsArtifact(t *testing.T) {
	logger := newTestLogger()
	gen := NewGenerator("1.0.0", "abc123", "output", logger)

	tests := []struct {
		name     string
		expected bool
	}{
		{"anubis-linux-amd64", true},
		{"anubis-linux-amd64.tar.gz", true},
		{"anubis-windows-amd64.exe", true},
		{"anubis-1.0.0.deb", true},
		{"anubis-1.0.0.rpm", true},
		{"checksums.txt", false},
		{"readme.md", false},
		{"main.go", false},
	}

	for _, tt := range tests {
		result := gen.isArtifact(tt.name)
		if result != tt.expected {
			t.Errorf("isArtifact(%q) = %v, expected %v", tt.name, result, tt.expected)
		}
	}
}

func TestCategorizeChange(t *testing.T) {
	logger := newTestLogger()
	gen := NewGenerator("1.0.0", "abc123", "output", logger)

	tests := []struct {
		message  string
		expected string
	}{
		{"feat: add new feature", "Added"},
		{"add something", "Added"},
		{"fix: bug fix", "Fixed"},
		{"bugfix: something", "Fixed"},
		{"update config", "Changed"},
		{"change: behavior", "Changed"},
		{"deprecate: old api", "Deprecated"},
		{"remove: unused code", "Removed"},
		{"delete file", "Removed"},
		{"security: fix CVE", "Security"},
		{"fix CVE-2024-1234", "Security"},
		{"other change", "Changed"},
	}

	for _, tt := range tests {
		change := Change{Message: tt.message}
		result := gen.categorizeChange(change)
		if result != tt.expected {
			t.Errorf("categorizeChange(%q) = %v, expected %v", tt.message, result, tt.expected)
		}
	}
}

func TestGenerateInstallScript(t *testing.T) {
	tmpDir := t.TempDir()
	logger := newTestLogger()
	gen := NewGenerator("v1.0.0", "abc123", tmpDir, logger)

	err := gen.GenerateInstallScript()
	if err != nil {
		t.Fatalf("Failed to generate install script: %v", err)
	}

	scriptPath := filepath.Join(tmpDir, "install.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read install script: %v", err)
	}

	// Check script contains expected content
	if !strings.Contains(string(content), "AnubisWatch") {
		t.Error("Install script should mention AnubisWatch")
	}

	if !strings.Contains(string(content), "v1.0.0") {
		t.Error("Install script should contain version")
	}

	// Check it's executable
	info, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("Failed to stat script: %v", err)
	}

	// On Unix systems, check if executable
	if info.Mode()&0111 != 0 {
		t.Log("Script is executable")
	}
}

func TestGeneratePackageScripts(t *testing.T) {
	tmpDir := t.TempDir()
	logger := newTestLogger()
	gen := NewGenerator("1.0.0", "abc123", tmpDir, logger)

	err := gen.GeneratePackageScripts()
	if err != nil {
		t.Fatalf("Failed to generate package scripts: %v", err)
	}

	// Check systemd service file
	servicePath := filepath.Join(tmpDir, "anubis.service")
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		t.Error("Systemd service file should be created")
	}

	// Check postinstall script
	postinstPath := filepath.Join(tmpDir, "postinstall.sh")
	if _, err := os.Stat(postinstPath); os.IsNotExist(err) {
		t.Error("Postinstall script should be created")
	}
}

func TestGenerateChecksums(t *testing.T) {
	tmpDir := t.TempDir()
	logger := newTestLogger()
	gen := NewGenerator("1.0.0", "abc123", tmpDir, logger)

	// Create test artifacts
	for _, name := range []string{"anubis-linux-amd64", "anubis-darwin-amd64"} {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(name+"-content"), 0644); err != nil {
			t.Fatalf("Failed to create artifact: %v", err)
		}
	}

	err := gen.GenerateChecksums(tmpDir)
	if err != nil {
		t.Fatalf("Failed to generate checksums: %v", err)
	}

	checksumPath := filepath.Join(tmpDir, "checksums.txt")
	content, err := os.ReadFile(checksumPath)
	if err != nil {
		t.Fatalf("Failed to read checksums: %v", err)
	}

	// Should have 2 checksum entries
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 checksum lines, got %d", len(lines))
	}
}

func TestWriteReleaseNotes(t *testing.T) {
	logger := newTestLogger()
	gen := NewGenerator("1.0.0", "abc123def456", "output", logger)

	notes := &ReleaseNotes{
		Version: "1.0.0",
		Commit:  "abc123def456",
		Date:    time.Now(), // Add this
		Sections: map[string][]Change{
			"Added": {
				{Message: "New feature 1", Commit: "abc123"},
				{Message: "New feature 2", Commit: "def456"},
			},
			"Fixed": {
				{Message: "Bug fix", Commit: "ghi789"},
			},
		},
	}

	var buf strings.Builder
	gen.writeReleaseNotes(&buf, notes)

	output := buf.String()

	if !strings.Contains(output, "1.0.0") {
		t.Error("Release notes should contain version")
	}

	if !strings.Contains(output, "Added") {
		t.Error("Release notes should contain Added section")
	}

	if !strings.Contains(output, "Fixed") {
		t.Error("Release notes should contain Fixed section")
	}

	if !strings.Contains(output, "New feature 1") {
		t.Error("Release notes should contain change messages")
	}
}

func TestChange(t *testing.T) {
	change := Change{
		Type:     "feat",
		Scope:    "api",
		Message:  "add new endpoint",
		Commit:   "abc123",
		Author:   "test@example.com",
		Breaking: false,
	}

	if change.Type != "feat" {
		t.Error("Change type not set correctly")
	}

	if change.Message != "add new endpoint" {
		t.Error("Change message not set correctly")
	}
}
