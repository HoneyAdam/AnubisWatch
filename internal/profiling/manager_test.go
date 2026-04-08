package profiling

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestNewManager(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DataDir = t.TempDir()

	mgr := NewManager(cfg, newTestLogger())
	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}

	if mgr.dataDir != cfg.DataDir {
		t.Errorf("dataDir = %s, want %s", mgr.dataDir, cfg.DataDir)
	}

	if mgr.enabled != cfg.Enabled {
		t.Errorf("enabled = %v, want %v", mgr.enabled, cfg.Enabled)
	}
}

func TestManager_StartStop(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.DataDir = tempDir

	mgr := NewManager(cfg, newTestLogger())

	err := mgr.Start()
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Give it time to collect some stats
	time.Sleep(100 * time.Millisecond)

	err = mgr.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Verify profiles directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("profiles directory was not created")
	}
}

func TestManager_StartNotEnabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = false

	mgr := NewManager(cfg, newTestLogger())

	err := mgr.Start()
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestCollectStats(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = tempDir

	mgr := NewManager(cfg, newTestLogger())
	mgr.collectStats()

	mem := mgr.GetMemProfile()
	if mem.Timestamp.IsZero() {
		t.Error("MemProfile timestamp is zero")
	}

	cpu := mgr.GetCPUProfile()
	if cpu.NumCPU == 0 {
		t.Error("CPUProfile NumCPU is zero")
	}

	goros := mgr.GetGoroutineStats()
	if goros.Count == 0 {
		t.Error("GoroutineStats Count is zero")
	}
}

func TestWriteHeapProfile(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = tempDir

	mgr := NewManager(cfg, newTestLogger())

	filename, err := mgr.WriteHeapProfile()
	if err != nil {
		t.Fatalf("WriteHeapProfile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("heap profile file was not created")
	}

	// Verify it's in the right directory
	if filepath.Dir(filename) != tempDir {
		t.Errorf("heap profile in wrong directory: %s", filepath.Dir(filename))
	}
}

func TestWriteGoroutineProfile(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = tempDir

	mgr := NewManager(cfg, newTestLogger())

	filename, err := mgr.WriteGoroutineProfile()
	if err != nil {
		t.Fatalf("WriteGoroutineProfile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("goroutine profile file was not created")
	}
}

func TestStartCPUProfile(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = tempDir

	mgr := NewManager(cfg, newTestLogger())

	filename, err := mgr.StartCPUProfile(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("StartCPUProfile() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("CPU profile file was not created")
	}

	// Wait for profile to complete
	time.Sleep(200 * time.Millisecond)

	// Verify file has content
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to stat CPU profile: %v", err)
	}
	if info.Size() == 0 {
		t.Error("CPU profile file is empty")
	}
}

func TestStartCPUProfile_AlreadyActive(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = tempDir

	mgr := NewManager(cfg, newTestLogger())

	// Start first profile
	_, err := mgr.StartCPUProfile(1 * time.Second)
	if err != nil {
		t.Fatalf("First StartCPUProfile() error = %v", err)
	}

	// Try to start second profile
	_, err = mgr.StartCPUProfile(1 * time.Second)
	if err == nil {
		t.Error("Second StartCPUProfile() should fail when already active")
	}

	// Cleanup
	mgr.StopCPUProfile()
}

func TestGenerateReport(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = tempDir

	mgr := NewManager(cfg, newTestLogger())
	mgr.collectStats()

	report, err := mgr.GenerateReport()
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}

	if report.Timestamp.IsZero() {
		t.Error("report.Timestamp is zero")
	}

	if report.MemStats.Timestamp.IsZero() {
		t.Error("report.MemStats.Timestamp is zero")
	}
}

func TestReport_WriteToFile(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = tempDir

	mgr := NewManager(cfg, newTestLogger())
	mgr.collectStats()

	report, err := mgr.GenerateReport()
	if err != nil {
		t.Fatalf("GenerateReport() error = %v", err)
	}

	path := filepath.Join(tempDir, "report.json")
	err = report.WriteToFile(path)
	if err != nil {
		t.Fatalf("WriteToFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("report file was not created")
	}

	// Verify it's valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	if len(data) == 0 {
		t.Error("report file is empty")
	}
}

func TestGetAllStats(t *testing.T) {
	tempDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.DataDir = tempDir

	mgr := NewManager(cfg, newTestLogger())
	mgr.collectStats()

	stats := mgr.GetAllStats()
	if stats == nil {
		t.Fatal("GetAllStats() returned nil")
	}

	requiredKeys := []string{"memory", "cpu", "goroutines", "gc", "timestamp"}
	for _, key := range requiredKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("GetAllStats() missing key: %s", key)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("DefaultConfig().Enabled should be false")
	}

	if cfg.DataDir != "./profiles" {
		t.Errorf("DefaultConfig().DataDir = %s, want ./profiles", cfg.DataDir)
	}

	if cfg.Interval != 30*time.Second {
		t.Errorf("DefaultConfig().Interval = %v, want 30s", cfg.Interval)
	}
}
