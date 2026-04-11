package quota

import (
	"fmt"
	"sync"

	"github.com/AnubisWatch/anubiswatch/internal/core"
)

// Manager enforces resource quotas per workspace
type Manager struct {
	mu           sync.RWMutex
	quotas       map[string]core.QuotaConfig // workspaceID -> quota
	counts       map[string]*UsageCounts     // workspaceID -> current usage
	defaultQuota core.QuotaConfig            // default quota for any workspace
}

// UsageCounts tracks current resource usage for a workspace
type UsageCounts struct {
	Souls         int `json:"souls"`
	Journeys      int `json:"journeys"`
	AlertChannels int `json:"alert_channels"`
	TeamMembers   int `json:"team_members"`
}

// NewManager creates a quota manager with optional per-workspace quotas
func NewManager(defaultQuota core.QuotaConfig) *Manager {
	return &Manager{
		quotas:       make(map[string]core.QuotaConfig),
		counts:       make(map[string]*UsageCounts),
		defaultQuota: defaultQuota,
	}
}

// SetQuota sets the quota for a specific workspace
func (m *Manager) SetQuota(workspaceID string, quota core.QuotaConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.quotas[workspaceID] = quota
}

// GetQuota returns the effective quota for a workspace
func (m *Manager) GetQuota(workspaceID string) core.QuotaConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if q, ok := m.quotas[workspaceID]; ok {
		return q
	}
	return m.defaultQuota
}

// GetUsage returns current usage counts
func (m *Manager) GetUsage(workspaceID string) UsageCounts {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if c, ok := m.counts[workspaceID]; ok {
		return *c
	}
	return UsageCounts{}
}

// CheckSoulLimit checks if a workspace can add another soul
func (m *Manager) CheckSoulLimit(workspaceID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	quota := m.quotas[workspaceID]
	if quota.MaxSouls == 0 {
		return nil // unlimited
	}

	usage := m.counts[workspaceID]
	if usage == nil {
		return nil
	}
	if usage.Souls >= quota.MaxSouls {
		return &QuotaExceededError{
			Workspace: workspaceID,
			Resource:  "souls",
			Limit:     quota.MaxSouls,
			Current:   usage.Souls,
		}
	}
	return nil
}

// CheckJourneyLimit checks if a workspace can add another journey
func (m *Manager) CheckJourneyLimit(workspaceID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	quota := m.quotas[workspaceID]
	if quota.MaxJourneys == 0 {
		return nil
	}

	usage := m.counts[workspaceID]
	if usage == nil {
		return nil
	}
	if usage.Journeys >= quota.MaxJourneys {
		return &QuotaExceededError{
			Workspace: workspaceID,
			Resource:  "journeys",
			Limit:     quota.MaxJourneys,
			Current:   usage.Journeys,
		}
	}
	return nil
}

// CheckAlertChannelLimit checks if a workspace can add another alert channel
func (m *Manager) CheckAlertChannelLimit(workspaceID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	quota := m.quotas[workspaceID]
	if quota.MaxAlertChannels == 0 {
		return nil
	}

	usage := m.counts[workspaceID]
	if usage == nil {
		return nil
	}
	if usage.AlertChannels >= quota.MaxAlertChannels {
		return &QuotaExceededError{
			Workspace: workspaceID,
			Resource:  "alert_channels",
			Limit:     quota.MaxAlertChannels,
			Current:   usage.AlertChannels,
		}
	}
	return nil
}

// CheckTeamMemberLimit checks if a workspace can add another team member
func (m *Manager) CheckTeamMemberLimit(workspaceID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	quota := m.quotas[workspaceID]
	if quota.MaxTeamMembers == 0 {
		return nil
	}

	usage := m.counts[workspaceID]
	if usage == nil {
		return nil
	}
	if usage.TeamMembers >= quota.MaxTeamMembers {
		return &QuotaExceededError{
			Workspace: workspaceID,
			Resource:  "team_members",
			Limit:     quota.MaxTeamMembers,
			Current:   usage.TeamMembers,
		}
	}
	return nil
}

// IncrementSoul increments the soul count for a workspace
func (m *Manager) IncrementSoul(workspaceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.counts[workspaceID] == nil {
		m.counts[workspaceID] = &UsageCounts{}
	}
	m.counts[workspaceID].Souls++
}

// DecrementSoul decrements the soul count
func (m *Manager) DecrementSoul(workspaceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.counts[workspaceID]; ok {
		if c.Souls > 0 {
			c.Souls--
		}
	}
}

// IncrementJourney increments the journey count
func (m *Manager) IncrementJourney(workspaceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.counts[workspaceID] == nil {
		m.counts[workspaceID] = &UsageCounts{}
	}
	m.counts[workspaceID].Journeys++
}

// DecrementJourney decrements the journey count
func (m *Manager) DecrementJourney(workspaceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.counts[workspaceID]; ok {
		if c.Journeys > 0 {
			c.Journeys--
		}
	}
}

// IncrementAlertChannel increments the alert channel count
func (m *Manager) IncrementAlertChannel(workspaceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.counts[workspaceID] == nil {
		m.counts[workspaceID] = &UsageCounts{}
	}
	m.counts[workspaceID].AlertChannels++
}

// DecrementAlertChannel decrements the alert channel count
func (m *Manager) DecrementAlertChannel(workspaceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.counts[workspaceID]; ok {
		if c.AlertChannels > 0 {
			c.AlertChannels--
		}
	}
}

// IncrementTeamMember increments the team member count
func (m *Manager) IncrementTeamMember(workspaceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.counts[workspaceID] == nil {
		m.counts[workspaceID] = &UsageCounts{}
	}
	m.counts[workspaceID].TeamMembers++
}

// DecrementTeamMember decrements the team member count
func (m *Manager) DecrementTeamMember(workspaceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.counts[workspaceID]; ok {
		if c.TeamMembers > 0 {
			c.TeamMembers--
		}
	}
}

// QuotaExceededError is returned when a resource limit is exceeded
type QuotaExceededError struct {
	Workspace string
	Resource  string
	Limit     int
	Current   int
}

func (e *QuotaExceededError) Error() string {
	return fmt.Sprintf("quota exceeded: workspace %s has %d/%d %s",
		e.Workspace, e.Current, e.Limit, e.Resource)
}

// IsQuotaExceeded checks if an error is a quota exceeded error
func IsQuotaExceeded(err error) bool {
	_, ok := err.(*QuotaExceededError)
	return ok
}

// Stats returns quota statistics for a workspace
func (m *Manager) Stats(workspaceID string) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	quota := m.quotas[workspaceID]
	usage := m.counts[workspaceID]
	if usage == nil {
		usage = &UsageCounts{}
	}

	return map[string]interface{}{
		"workspace": workspaceID,
		"quota": map[string]interface{}{
			"max_souls":          quota.MaxSouls,
			"max_journeys":       quota.MaxJourneys,
			"max_alert_channels": quota.MaxAlertChannels,
			"max_team_members":   quota.MaxTeamMembers,
			"retention_days":     quota.RetentionDays,
		},
		"usage": map[string]interface{}{
			"souls":          usage.Souls,
			"journeys":       usage.Journeys,
			"alert_channels": usage.AlertChannels,
			"team_members":   usage.TeamMembers,
		},
	}
}
