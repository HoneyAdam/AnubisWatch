package quota

import (
	"testing"

	"github.com/AnubisWatch/anubiswatch/internal/core"
)

func TestNewManager_NoQuota(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	if err := m.CheckSoulLimit("ws1"); err != nil {
		t.Errorf("Expected no error with unlimited quota, got %v", err)
	}
}

func TestCheckSoulLimit_WithinQuota(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxSouls: 10})
	m.counts["ws1"] = &UsageCounts{Souls: 5}

	if err := m.CheckSoulLimit("ws1"); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCheckSoulLimit_AtLimit(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxSouls: 10})
	m.counts["ws1"] = &UsageCounts{Souls: 10}

	err := m.CheckSoulLimit("ws1")
	if err == nil {
		t.Error("Expected quota exceeded error")
	}
	if !IsQuotaExceeded(err) {
		t.Error("Expected IsQuotaExceeded to return true")
	}
}

func TestCheckJourneyLimit_WithinQuota(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxJourneys: 5})
	m.counts["ws1"] = &UsageCounts{Journeys: 3}

	if err := m.CheckJourneyLimit("ws1"); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCheckJourneyLimit_Exceeded(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxJourneys: 5})
	m.counts["ws1"] = &UsageCounts{Journeys: 5}

	err := m.CheckJourneyLimit("ws1")
	if err == nil {
		t.Error("Expected quota exceeded error")
	}
}

func TestCheckAlertChannelLimit(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxAlertChannels: 3})
	m.counts["ws1"] = &UsageCounts{AlertChannels: 3}

	err := m.CheckAlertChannelLimit("ws1")
	if err == nil {
		t.Error("Expected quota exceeded error")
	}
}

func TestCheckTeamMemberLimit(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxTeamMembers: 10})
	m.counts["ws1"] = &UsageCounts{TeamMembers: 10}

	err := m.CheckTeamMemberLimit("ws1")
	if err == nil {
		t.Error("Expected quota exceeded error")
	}
}

func TestIncrementDecrementSoul(t *testing.T) {
	m := NewManager(core.QuotaConfig{})

	m.IncrementSoul("ws1")
	if m.counts["ws1"].Souls != 1 {
		t.Errorf("Expected 1 soul, got %d", m.counts["ws1"].Souls)
	}

	m.IncrementSoul("ws1")
	if m.counts["ws1"].Souls != 2 {
		t.Errorf("Expected 2 souls, got %d", m.counts["ws1"].Souls)
	}

	m.DecrementSoul("ws1")
	if m.counts["ws1"].Souls != 1 {
		t.Errorf("Expected 1 soul after decrement, got %d", m.counts["ws1"].Souls)
	}

	m.DecrementSoul("ws1")
	m.DecrementSoul("ws1") // Should not go below 0
	if m.counts["ws1"].Souls != 0 {
		t.Errorf("Expected 0 souls, got %d", m.counts["ws1"].Souls)
	}
}

func TestIncrementDecrementJourney(t *testing.T) {
	m := NewManager(core.QuotaConfig{})

	m.IncrementJourney("ws1")
	m.IncrementJourney("ws1")
	if m.counts["ws1"].Journeys != 2 {
		t.Errorf("Expected 2 journeys, got %d", m.counts["ws1"].Journeys)
	}

	m.DecrementJourney("ws1")
	if m.counts["ws1"].Journeys != 1 {
		t.Errorf("Expected 1 journey, got %d", m.counts["ws1"].Journeys)
	}
}

func TestGetQuota_DefaultFallback(t *testing.T) {
	m := NewManager(core.QuotaConfig{MaxSouls: 50})

	// Workspace without explicit quota should use defaults
	quota := m.GetQuota("unknown-ws")
	if quota.MaxSouls != 50 {
		t.Errorf("Expected default MaxSouls=50, got %d", quota.MaxSouls)
	}
}

func TestGetUsage_EmptyWorkspace(t *testing.T) {
	m := NewManager(core.QuotaConfig{})

	usage := m.GetUsage("new-ws")
	if usage.Souls != 0 || usage.Journeys != 0 {
		t.Errorf("Expected zero usage, got %+v", usage)
	}
}

func TestStats(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxSouls: 10, MaxJourneys: 5})
	m.counts["ws1"] = &UsageCounts{Souls: 3, Journeys: 2}

	stats := m.Stats("ws1")
	quota := stats["quota"].(map[string]interface{})
	usage := stats["usage"].(map[string]interface{})

	if quota["max_souls"] != 10 {
		t.Errorf("Expected max_souls=10, got %v", quota["max_souls"])
	}
	if usage["souls"] != 3 {
		t.Errorf("Expected souls=3, got %v", usage["souls"])
	}
}

func TestQuotaExceededError_Message(t *testing.T) {
	err := &QuotaExceededError{
		Workspace: "ws1",
		Resource:  "souls",
		Limit:     100,
		Current:   100,
	}

	expected := "quota exceeded: workspace ws1 has 100/100 souls"
	if err.Error() != expected {
		t.Errorf("Expected error message %q, got %q", expected, err.Error())
	}
}

func TestIncrementAlertChannel(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.IncrementAlertChannel("ws1")
	m.IncrementAlertChannel("ws1")
	if m.counts["ws1"].AlertChannels != 2 {
		t.Errorf("Expected 2 alert channels, got %d", m.counts["ws1"].AlertChannels)
	}

	m.DecrementAlertChannel("ws1")
	if m.counts["ws1"].AlertChannels != 1 {
		t.Errorf("Expected 1 alert channel, got %d", m.counts["ws1"].AlertChannels)
	}
}

func TestIncrementTeamMember(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.IncrementTeamMember("ws1")
	if m.counts["ws1"].TeamMembers != 1 {
		t.Errorf("Expected 1 team member, got %d", m.counts["ws1"].TeamMembers)
	}

	m.DecrementTeamMember("ws1")
	if m.counts["ws1"].TeamMembers != 0 {
		t.Errorf("Expected 0 team members, got %d", m.counts["ws1"].TeamMembers)
	}
}

func TestNoUsageCounts_CreatesOnIncrement(t *testing.T) {
	m := NewManager(core.QuotaConfig{})

	m.IncrementSoul("new-workspace")
	if m.counts["new-workspace"].Souls != 1 {
		t.Error("Expected auto-creation of usage counts on increment")
	}
}

func TestDecrementNonExistent(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	// Should not panic
	m.DecrementSoul("nonexistent")
	m.DecrementJourney("nonexistent")
	m.DecrementAlertChannel("nonexistent")
	m.DecrementTeamMember("nonexistent")
}

// TestCheckSoulLimit_NoUsage tests that nil usage allows adding souls
func TestCheckSoulLimit_NoUsage(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxSouls: 10})
	// No usage counts set - should allow

	if err := m.CheckSoulLimit("ws1"); err != nil {
		t.Errorf("Expected no error with nil usage, got %v", err)
	}
}

// TestCheckAlertChannelLimit_Within tests alert channel within quota
func TestCheckAlertChannelLimit_Within(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxAlertChannels: 5})
	m.counts["ws1"] = &UsageCounts{AlertChannels: 3}

	if err := m.CheckAlertChannelLimit("ws1"); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestCheckAlertChannelLimit_NoUsage tests nil usage path
func TestCheckAlertChannelLimit_NoUsage(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxAlertChannels: 5})

	if err := m.CheckAlertChannelLimit("ws1"); err != nil {
		t.Errorf("Expected no error with nil usage, got %v", err)
	}
}

// TestCheckAlertChannelLimit_Unlimited tests zero limit (unlimited)
func TestCheckAlertChannelLimit_Unlimited(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.counts["ws1"] = &UsageCounts{AlertChannels: 100}

	if err := m.CheckAlertChannelLimit("ws1"); err != nil {
		t.Errorf("Expected no error with unlimited quota, got %v", err)
	}
}

// TestCheckTeamMemberLimit_Within tests team member within quota
func TestCheckTeamMemberLimit_Within(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxTeamMembers: 20})
	m.counts["ws1"] = &UsageCounts{TeamMembers: 5}

	if err := m.CheckTeamMemberLimit("ws1"); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestCheckTeamMemberLimit_NoUsage tests nil usage path
func TestCheckTeamMemberLimit_NoUsage(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxTeamMembers: 10})

	if err := m.CheckTeamMemberLimit("ws1"); err != nil {
		t.Errorf("Expected no error with nil usage, got %v", err)
	}
}

// TestCheckTeamMemberLimit_Unlimited tests zero limit
func TestCheckTeamMemberLimit_Unlimited(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.counts["ws1"] = &UsageCounts{TeamMembers: 100}

	if err := m.CheckTeamMemberLimit("ws1"); err != nil {
		t.Errorf("Expected no error with unlimited quota, got %v", err)
	}
}

// TestCheckJourneyLimit_NoUsage tests nil usage path
func TestCheckJourneyLimit_NoUsage(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.SetQuota("ws1", core.QuotaConfig{MaxJourneys: 5})

	if err := m.CheckJourneyLimit("ws1"); err != nil {
		t.Errorf("Expected no error with nil usage, got %v", err)
	}
}

// TestGetQuota_ExplicitQuota tests fetching a workspace with explicit quota
func TestGetQuota_ExplicitQuota(t *testing.T) {
	m := NewManager(core.QuotaConfig{MaxSouls: 50})
	m.SetQuota("ws1", core.QuotaConfig{MaxSouls: 100, MaxJourneys: 10})

	quota := m.GetQuota("ws1")
	if quota.MaxSouls != 100 {
		t.Errorf("Expected MaxSouls=100, got %d", quota.MaxSouls)
	}
	if quota.MaxJourneys != 10 {
		t.Errorf("Expected MaxJourneys=10, got %d", quota.MaxJourneys)
	}
}

// TestGetUsage_PopulatedWorkspace tests GetUsage with existing counts
func TestGetUsage_PopulatedWorkspace(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.counts["ws1"] = &UsageCounts{Souls: 7, Journeys: 3, AlertChannels: 2, TeamMembers: 5}

	usage := m.GetUsage("ws1")
	if usage.Souls != 7 || usage.Journeys != 3 || usage.AlertChannels != 2 || usage.TeamMembers != 5 {
		t.Errorf("Expected full usage counts, got %+v", usage)
	}
}

// TestStats_UnknownWorkspace tests Stats for unknown workspace
func TestStats_UnknownWorkspace(t *testing.T) {
	m := NewManager(core.QuotaConfig{MaxSouls: 50})

	stats := m.Stats("unknown-ws")
	quota := stats["quota"].(map[string]interface{})
	// Stats uses m.quotas directly (not GetQuota), so unknown workspace gets zero values
	if quota["max_souls"] != 0 {
		t.Errorf("Expected max_souls=0 for unknown workspace, got %v", quota["max_souls"])
	}
	usage := stats["usage"].(map[string]interface{})
	if usage["souls"] != 0 {
		t.Errorf("Expected souls=0, got %v", usage["souls"])
	}
}

// TestDecrementTeamMember_NonExistent tests decrement team member for workspace with no counts
func TestDecrementTeamMember_NonExistent(t *testing.T) {
	m := NewManager(core.QuotaConfig{})
	m.DecrementTeamMember("nonexistent")
	// Should not panic
}
