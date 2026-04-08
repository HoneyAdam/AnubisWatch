package backup

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AnubisWatch/anubiswatch/internal/core"
)

// mockStorage implements BackupStorage interface for testing
type mockStorage struct {
	workspaces    []*core.Workspace
	souls         []*core.Soul
	alertChannels []*core.AlertChannel
	alertRules    []*core.AlertRule
	statusPages   []*core.StatusPage
	journeys      []*core.JourneyConfig
	systemConfig  map[string][]byte
}

func (m *mockStorage) ListWorkspaces(ctx context.Context) ([]*core.Workspace, error) {
	return m.workspaces, nil
}

func (m *mockStorage) ListSouls(ctx context.Context, workspaceID string, offset, limit int) ([]*core.Soul, error) {
	var result []*core.Soul
	for _, s := range m.souls {
		if s.WorkspaceID == workspaceID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockStorage) ListAlertChannels() ([]*core.AlertChannel, error) {
	return m.alertChannels, nil
}

func (m *mockStorage) ListAlertRules() ([]*core.AlertRule, error) {
	return m.alertRules, nil
}

func (m *mockStorage) ListStatusPages() ([]*core.StatusPage, error) {
	return m.statusPages, nil
}

func (m *mockStorage) ListJourneys(ctx context.Context, workspaceID string) ([]*core.JourneyConfig, error) {
	var result []*core.JourneyConfig
	for _, j := range m.journeys {
		if j.WorkspaceID == workspaceID {
			result = append(result, j)
		}
	}
	return result, nil
}

func (m *mockStorage) ListJudgments(ctx context.Context, soulID string, start, end time.Time, limit int) ([]*core.Judgment, error) {
	return nil, nil
}

func (m *mockStorage) GetSystemConfig(ctx context.Context, key string) ([]byte, error) {
	return m.systemConfig[key], nil
}

// mockRestoreStorage implements RestoreStorage interface for testing
type mockRestoreStorage struct {
	souls         []*core.Soul
	workspaces    []*core.Workspace
	alertChannels []*core.AlertChannel
	alertRules    []*core.AlertRule
	statusPages   []*core.StatusPage
	journeys      []*core.JourneyConfig
	systemConfig  map[string][]byte
}

func (m *mockRestoreStorage) SaveSoul(ctx context.Context, soul *core.Soul) error {
	m.souls = append(m.souls, soul)
	return nil
}

func (m *mockRestoreStorage) SaveWorkspace(ctx context.Context, ws *core.Workspace) error {
	m.workspaces = append(m.workspaces, ws)
	return nil
}

func (m *mockRestoreStorage) SaveAlertChannel(ch *core.AlertChannel) error {
	m.alertChannels = append(m.alertChannels, ch)
	return nil
}

func (m *mockRestoreStorage) SaveAlertRule(rule *core.AlertRule) error {
	m.alertRules = append(m.alertRules, rule)
	return nil
}

func (m *mockRestoreStorage) SaveStatusPage(page *core.StatusPage) error {
	m.statusPages = append(m.statusPages, page)
	return nil
}

func (m *mockRestoreStorage) SaveJourney(ctx context.Context, j *core.JourneyConfig) error {
	m.journeys = append(m.journeys, j)
	return nil
}

func (m *mockRestoreStorage) SaveSystemConfig(ctx context.Context, key string, value []byte) error {
	if m.systemConfig == nil {
		m.systemConfig = make(map[string][]byte)
	}
	m.systemConfig[key] = value
	return nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestManager_Init(t *testing.T) {
	tempDir := t.TempDir()
	mgr := NewManager(nil, tempDir, newTestLogger())

	err := mgr.Init()
	if err != nil {
		t.Errorf("Init() error = %v", err)
	}

	// Check that backups directory was created
	backupsDir := filepath.Join(tempDir, "backups")
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		t.Error("backups directory was not created")
	}
}

func TestManager_Create(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock storage with test data
	storage := &mockStorage{
		workspaces: []*core.Workspace{
			{ID: "ws1", Name: "Workspace 1"},
			{ID: "ws2", Name: "Workspace 2"},
		},
		souls: []*core.Soul{
			{ID: "soul1", Name: "Soul 1", WorkspaceID: "ws1"},
			{ID: "soul2", Name: "Soul 2", WorkspaceID: "ws1"},
			{ID: "soul3", Name: "Soul 3", WorkspaceID: "ws2"},
		},
		alertChannels: []*core.AlertChannel{
			{ID: "ch1", Name: "Channel 1"},
		},
		alertRules: []*core.AlertRule{
			{ID: "rule1", Name: "Rule 1"},
		},
		statusPages: []*core.StatusPage{
			{ID: "page1", Name: "Status Page 1"},
		},
		journeys: []*core.JourneyConfig{
			{ID: "journey1", Name: "Journey 1", WorkspaceID: "ws1"},
		},
	}

	mgr := NewManager(storage, tempDir, newTestLogger())
	if err := mgr.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	opts := DefaultOptions()
	ctx := context.Background()

	backup, path, err := mgr.Create(ctx, opts)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify backup contents
	if backup.Version != "1.0" {
		t.Errorf("backup.Version = %s, want 1.0", backup.Version)
	}

	if backup.BackupType != "full" {
		t.Errorf("backup.BackupType = %s, want full", backup.BackupType)
	}

	if backup.Metadata.Workspaces != 2 {
		t.Errorf("backup.Metadata.Workspaces = %d, want 2", backup.Metadata.Workspaces)
	}

	if backup.Metadata.Souls != 3 {
		t.Errorf("backup.Metadata.Souls = %d, want 3", backup.Metadata.Souls)
	}

	if backup.Metadata.AlertChannels != 1 {
		t.Errorf("backup.Metadata.AlertChannels = %d, want 1", backup.Metadata.AlertChannels)
	}

	if backup.Checksum == "" {
		t.Error("backup.Checksum is empty")
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("backup file was not created")
	}

	// Cleanup
	os.Remove(path)
}

func TestManager_CreateAndRestore(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock storage with test data
	storage := &mockStorage{
		workspaces: []*core.Workspace{
			{ID: "ws1", Name: "Workspace 1"},
		},
		souls: []*core.Soul{
			{ID: "soul1", Name: "Soul 1", WorkspaceID: "ws1", Target: "https://example.com"},
		},
		alertChannels: []*core.AlertChannel{
			{ID: "ch1", Name: "Channel 1", Type: core.ChannelSlack},
		},
		alertRules: []*core.AlertRule{
			{ID: "rule1", Name: "Rule 1", Enabled: true},
		},
		statusPages: []*core.StatusPage{
			{ID: "page1", Name: "Status Page 1", Slug: "status"},
		},
		journeys: []*core.JourneyConfig{
			{ID: "journey1", Name: "Journey 1", WorkspaceID: "ws1"},
		},
	}

	mgr := NewManager(storage, tempDir, newTestLogger())
	if err := mgr.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Create backup
	opts := DefaultOptions()
	ctx := context.Background()

	_, path, err := mgr.Create(ctx, opts)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer os.Remove(path)

	// Restore to new storage
	restoreStorage := &mockRestoreStorage{}
	restoreOpts := DefaultRestoreOptions()

	err = mgr.Restore(ctx, restoreStorage, path, restoreOpts)
	if err != nil {
		t.Fatalf("Restore() error = %v", err)
	}

	// Verify restored data
	if len(restoreStorage.workspaces) != 1 {
		t.Errorf("restored workspaces = %d, want 1", len(restoreStorage.workspaces))
	}

	if len(restoreStorage.souls) != 1 {
		t.Errorf("restored souls = %d, want 1", len(restoreStorage.souls))
	}

	if restoreStorage.souls[0].Target != "https://example.com" {
		t.Errorf("restored soul target = %s, want https://example.com", restoreStorage.souls[0].Target)
	}

	if len(restoreStorage.alertChannels) != 1 {
		t.Errorf("restored alert channels = %d, want 1", len(restoreStorage.alertChannels))
	}

	if len(restoreStorage.alertRules) != 1 {
		t.Errorf("restored alert rules = %d, want 1", len(restoreStorage.alertRules))
	}

	if len(restoreStorage.statusPages) != 1 {
		t.Errorf("restored status pages = %d, want 1", len(restoreStorage.statusPages))
	}

	if len(restoreStorage.journeys) != 1 {
		t.Errorf("restored journeys = %d, want 1", len(restoreStorage.journeys))
	}
}

func TestManager_List(t *testing.T) {
	tempDir := t.TempDir()

	storage := &mockStorage{
		workspaces: []*core.Workspace{{ID: "ws1", Name: "Workspace 1"}},
		souls:      []*core.Soul{{ID: "soul1", Name: "Soul 1", WorkspaceID: "ws1"}},
	}

	mgr := NewManager(storage, tempDir, newTestLogger())
	if err := mgr.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Create a backup first
	ctx := context.Background()
	_, path, err := mgr.Create(ctx, DefaultOptions())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer os.Remove(path)

	// List backups
	backups, err := mgr.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(backups) != 1 {
		t.Errorf("List() returned %d backups, want 1", len(backups))
	}

	if backups[0].Filename == "" {
		t.Error("backup.Filename is empty")
	}
}

func TestManager_Delete(t *testing.T) {
	tempDir := t.TempDir()

	storage := &mockStorage{
		workspaces: []*core.Workspace{{ID: "ws1", Name: "Workspace 1"}},
		souls:      []*core.Soul{{ID: "soul1", Name: "Soul 1", WorkspaceID: "ws1"}},
	}

	mgr := NewManager(storage, tempDir, newTestLogger())
	if err := mgr.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Create a backup
	ctx := context.Background()
	_, path, err := mgr.Create(ctx, DefaultOptions())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	filename := filepath.Base(path)

	// Delete the backup
	err = mgr.Delete(filename)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("backup file still exists after delete")
	}
}

func TestManager_Delete_InvalidPath(t *testing.T) {
	tempDir := t.TempDir()
	mgr := NewManager(nil, tempDir, newTestLogger())

	// Try to delete a file outside backups directory
	err := mgr.Delete("../etc/passwd")
	if err == nil {
		t.Error("Delete() should fail for path outside backups directory")
	}
}

func TestManager_VerifyChecksum(t *testing.T) {
	tempDir := t.TempDir()

	storage := &mockStorage{
		workspaces: []*core.Workspace{{ID: "ws1", Name: "Workspace 1"}},
		souls:      []*core.Soul{{ID: "soul1", Name: "Soul 1", WorkspaceID: "ws1"}},
	}

	mgr := NewManager(storage, tempDir, newTestLogger())
	if err := mgr.Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Create a backup
	ctx := context.Background()
	backup, path, err := mgr.Create(ctx, DefaultOptions())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer os.Remove(path)

	// Verify checksum is valid
	if err := mgr.verifyChecksum(backup); err != nil {
		t.Errorf("verifyChecksum() error = %v", err)
	}

	// Modify checksum and verify it fails
	backup.Checksum = "invalid"
	if err := mgr.verifyChecksum(backup); err == nil {
		t.Error("verifyChecksum() should fail for invalid checksum")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if !opts.Compress {
		t.Error("DefaultOptions().Compress should be true")
	}

	if !opts.IncludeJudgments {
		t.Error("DefaultOptions().IncludeJudgments should be true")
	}

	if opts.JudgmentDays != 7 {
		t.Errorf("DefaultOptions().JudgmentDays = %d, want 7", opts.JudgmentDays)
	}
}

func TestDefaultRestoreOptions(t *testing.T) {
	opts := DefaultRestoreOptions()

	if !opts.IncludeWorkspaces {
		t.Error("DefaultRestoreOptions().IncludeWorkspaces should be true")
	}

	if !opts.IncludeSouls {
		t.Error("DefaultRestoreOptions().IncludeSouls should be true")
	}

	if !opts.IncludeAlerts {
		t.Error("DefaultRestoreOptions().IncludeAlerts should be true")
	}

	if !opts.IncludeStatusPages {
		t.Error("DefaultRestoreOptions().IncludeStatusPages should be true")
	}

	if !opts.IncludeJourneys {
		t.Error("DefaultRestoreOptions().IncludeJourneys should be true")
	}

	if !opts.IncludeSystemConfig {
		t.Error("DefaultRestoreOptions().IncludeSystemConfig should be true")
	}

	if opts.ContinueOnError {
		t.Error("DefaultRestoreOptions().ContinueOnError should be false")
	}
}

func TestIsWithinDirectory(t *testing.T) {
	tests := []struct {
		path     string
		dir      string
		expected bool
	}{
		{"/data/backups/file.json", "/data/backups", true},
		{"/data/backups/subdir/file.json", "/data/backups", true},
		{"/data/other/file.json", "/data/backups", false},
		{"/etc/passwd", "/data/backups", false},
	}

	for _, tt := range tests {
		result := IsWithinDirectory(tt.path, tt.dir)
		if result != tt.expected {
			t.Errorf("IsWithinDirectory(%q, %q) = %v, want %v",
				tt.path, tt.dir, result, tt.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}
