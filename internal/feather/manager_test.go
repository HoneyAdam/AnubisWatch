package feather

import (
	"testing"
	"time"

	"github.com/AnubisWatch/anubiswatch/internal/core"
)

func TestNewManager(t *testing.T) {
	m := NewManager(nil)
	if m == nil {
		t.Fatal("Expected manager to be created")
	}
}

func TestAddFeather(t *testing.T) {
	m := NewManager(nil)

	cfg := core.FeatherConfig{
		Name:  "test-feather",
		Scope: "all",
		Rules: core.FeatherRules{
			P50: core.Duration{Duration: 100 * time.Millisecond},
			P95: core.Duration{Duration: 500 * time.Millisecond},
		},
		Window:             core.Duration{Duration: 1 * time.Hour},
		ViolationThreshold: 3,
	}

	m.AddFeather(cfg)

	feathers := m.GetFeathers()
	if len(feathers) != 1 {
		t.Errorf("Expected 1 feather, got %d", len(feathers))
	}
	if feathers["test-feather"].Name != "test-feather" {
		t.Errorf("Expected feather name 'test-feather', got %s", feathers["test-feather"].Name)
	}
}

func TestRemoveFeather(t *testing.T) {
	m := NewManager(nil)

	m.AddFeather(core.FeatherConfig{Name: "f1"})
	m.AddFeather(core.FeatherConfig{Name: "f2"})
	m.RemoveFeather("f1")

	feathers := m.GetFeathers()
	if _, exists := feathers["f1"]; exists {
		t.Error("Expected f1 to be removed")
	}
	if _, exists := feathers["f2"]; !exists {
		t.Error("Expected f2 to still exist")
	}
}

func TestRecordLatency(t *testing.T) {
	m := NewManager(nil)

	m.RecordLatency("soul-1", 50*time.Millisecond, core.SoulAlive)
	m.RecordLatency("soul-1", 75*time.Millisecond, core.SoulAlive)

	if len(m.latencyDB["soul-1"]) != 2 {
		t.Errorf("Expected 2 samples, got %d", len(m.latencyDB["soul-1"]))
	}
}

func TestEvaluate_NoFeathers(t *testing.T) {
	m := NewManager(nil)

	m.RecordLatency("soul-1", 50*time.Millisecond, core.SoulAlive)

	violations := m.Evaluate()
	if len(violations) != 0 {
		t.Errorf("Expected no violations with no feathers, got %d", len(violations))
	}
}

func TestEvaluate_WithinBudget(t *testing.T) {
	m := NewManager(nil)

	m.AddFeather(core.FeatherConfig{
		Name:  "strict-feather",
		Scope: "all",
		Rules: core.FeatherRules{
			P95: core.Duration{Duration: 5 * time.Second},
		},
		Window: core.Duration{Duration: 1 * time.Hour},
	})

	// Record samples well within budget
	for i := 0; i < 10; i++ {
		m.RecordLatency("soul-1", time.Duration(100+i*10)*time.Millisecond, core.SoulAlive)
	}

	violations := m.Evaluate()
	if len(violations) != 0 {
		t.Errorf("Expected no violations within budget, got %d", len(violations))
	}
}

func TestEvaluate_ExceedsP50(t *testing.T) {
	m := NewManager(nil)

	m.AddFeather(core.FeatherConfig{
		Name:  "p50-feather",
		Scope: "all",
		Rules: core.FeatherRules{
			P50: core.Duration{Duration: 50 * time.Millisecond},
		},
		Window: core.Duration{Duration: 1 * time.Hour},
	})

	// Record samples exceeding P50 budget
	for i := 0; i < 10; i++ {
		m.RecordLatency("soul-1", 100*time.Millisecond, core.SoulAlive)
	}

	violations := m.Evaluate()
	if len(violations) == 0 {
		t.Error("Expected P50 violation")
	}
	if len(violations) > 0 && violations[0].Metric != "p50" {
		t.Errorf("Expected p50 violation, got %s", violations[0].Metric)
	}
}

func TestEvaluate_ExceedsMax(t *testing.T) {
	m := NewManager(nil)

	m.AddFeather(core.FeatherConfig{
		Name:  "max-feather",
		Scope: "all",
		Rules: core.FeatherRules{
			Max: core.Duration{Duration: 100 * time.Millisecond},
		},
		Window: core.Duration{Duration: 1 * time.Hour},
	})

	m.RecordLatency("soul-1", 50*time.Millisecond, core.SoulAlive)
	m.RecordLatency("soul-1", 75*time.Millisecond, core.SoulAlive)
	m.RecordLatency("soul-1", 150*time.Millisecond, core.SoulAlive)

	violations := m.Evaluate()
	if len(violations) == 0 {
		t.Error("Expected max violation")
	}
}

func TestEvaluate_ViolationCallback(t *testing.T) {
	var callbackFired bool
	var callbackSoulID string

	m := NewManager(func(soulID, featherName string, v FeatherViolation) {
		callbackFired = true
		callbackSoulID = soulID
	})

	m.AddFeather(core.FeatherConfig{
		Name:  "callback-feather",
		Scope: "all",
		Rules: core.FeatherRules{
			P95: core.Duration{Duration: 10 * time.Millisecond},
		},
		Window:             core.Duration{Duration: 1 * time.Hour},
		ViolationThreshold: 1, // Fire after 1 violation
	})

	// Record high-latency samples
	for i := 0; i < 10; i++ {
		m.RecordLatency("soul-callback", 100*time.Millisecond, core.SoulAlive)
	}

	m.Evaluate() // Should trigger callback

	// Give goroutine time to fire
	time.Sleep(50 * time.Millisecond)

	if !callbackFired {
		t.Error("Expected violation callback to fire")
	}
	if callbackSoulID != "soul-callback" {
		t.Errorf("Expected soul-callback, got %s", callbackSoulID)
	}
}

func TestCleanOld(t *testing.T) {
	m := NewManager(nil)

	now := time.Now().UTC()
	samples := []latencySample{
		{duration: 100 * time.Millisecond, timestamp: now.Add(-2 * time.Hour)},
		{duration: 200 * time.Millisecond, timestamp: now.Add(-10 * time.Minute)},
		{duration: 300 * time.Millisecond, timestamp: now.Add(-1 * time.Minute)},
	}

	cleaned := m.cleanOld(samples, 1*time.Hour)
	if len(cleaned) != 2 {
		t.Errorf("Expected 2 samples after cleaning, got %d", len(cleaned))
	}
}

func TestFeatherMatches_Scope(t *testing.T) {
	m := NewManager(nil)

	m.AddFeather(core.FeatherConfig{Name: "soul-specific", Scope: "soul-1"})
	m.AddFeather(core.FeatherConfig{Name: "global", Scope: "all"})

	matches := m.matchingFeathers("soul-1")
	if len(matches) != 2 {
		t.Errorf("Expected 2 matching feathers for soul-1, got %d", len(matches))
	}

	matches = m.matchingFeathers("soul-2")
	if len(matches) != 1 {
		t.Errorf("Expected 1 matching feather for soul-2, got %d", len(matches))
	}
}

func TestFeatherMatches_EmptyScope(t *testing.T) {
	m := NewManager(nil)

	m.AddFeather(core.FeatherConfig{Name: "no-scope", Scope: ""})

	matches := m.matchingFeathers("any-soul")
	if len(matches) != 1 {
		t.Errorf("Expected 1 match for empty scope, got %d", len(matches))
	}
}

func TestStats(t *testing.T) {
	m := NewManager(nil)

	m.AddFeather(core.FeatherConfig{Name: "f1"})
	m.RecordLatency("soul-1", 100*time.Millisecond, core.SoulAlive)

	stats := m.Stats()

	if stats["total_feathers"] != 1 {
		t.Errorf("Expected 1 feather in stats, got %v", stats["total_feathers"])
	}
	if stats["tracked_souls"] != 1 {
		t.Errorf("Expected 1 tracked soul in stats, got %v", stats["tracked_souls"])
	}
}

func TestPercentile(t *testing.T) {
	sorted := make([]time.Duration, 100)
	for i := 0; i < 100; i++ {
		sorted[i] = time.Duration(i+1) * time.Millisecond
	}

	p50 := percentile(sorted, 50)
	// Linear interpolation: rank = 0.50 * 99 = 49.5 -> 50 + 0.5*(51-50) = 50.5ms
	expectedP50 := 50500000 * time.Nanosecond
	if p50 != expectedP50 {
		t.Errorf("Expected P50 50.5ms, got %v", p50)
	}

	p95 := percentile(sorted, 95)
	// rank = 0.95 * 99 = 94.05 -> 95 + 0.05*(96-95) = 95.05ms
	expectedP95 := 95050000 * time.Nanosecond
	if p95 != expectedP95 {
		t.Errorf("Expected P95 95.05ms, got %v", p95)
	}
}

func TestPercentile_SingleElement(t *testing.T) {
	sorted := []time.Duration{42 * time.Millisecond}

	result := percentile(sorted, 99)
	if result != 42*time.Millisecond {
		t.Errorf("Expected 42ms, got %v", result)
	}
}

func TestPercentile_EmptySlice(t *testing.T) {
	result := percentile(nil, 50)
	if result != 0 {
		t.Errorf("Expected 0 for empty slice, got %v", result)
	}
}

func TestEvaluate_P95Violation(t *testing.T) {
	m := NewManager(nil)

	m.AddFeather(core.FeatherConfig{
		Name:  "p95-feather",
		Scope: "all",
		Rules: core.FeatherRules{
			P95: core.Duration{Duration: 200 * time.Millisecond},
		},
		Window: core.Duration{Duration: 1 * time.Hour},
	})

	// Record mostly fast, some slow samples
	for i := 0; i < 18; i++ {
		m.RecordLatency("soul-p95", 100*time.Millisecond, core.SoulAlive)
	}
	for i := 0; i < 2; i++ {
		m.RecordLatency("soul-p95", 500*time.Millisecond, core.SoulAlive)
	}

	violations := m.Evaluate()
	if len(violations) == 0 {
		t.Error("Expected P95 violation")
	}
}

func TestEvaluate_P99Violation(t *testing.T) {
	m := NewManager(nil)

	m.AddFeather(core.FeatherConfig{
		Name:  "p99-feather",
		Scope: "all",
		Rules: core.FeatherRules{
			P99: core.Duration{Duration: 100 * time.Millisecond},
		},
		Window: core.Duration{Duration: 1 * time.Hour},
	})

	// Record 100 samples, all exceeding budget
	for i := 0; i < 100; i++ {
		m.RecordLatency("soul-p99", 200*time.Millisecond, core.SoulAlive)
	}

	violations := m.Evaluate()
	if len(violations) == 0 {
		t.Error("Expected P99 violation")
	}
}
