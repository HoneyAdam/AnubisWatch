package metrics

import (
	"strings"
	"testing"
	"time"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry("test")
	if r == nil {
		t.Fatal("Expected registry to be created")
	}
	if r.prefix != "test" {
		t.Errorf("Expected prefix 'test', got %s", r.prefix)
	}
}

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry("test")
	counter := NewCounter("requests_total", "Total requests")

	err := r.Register(counter)
	if err != nil {
		t.Fatalf("Failed to register metric: %v", err)
	}

	// Try to register duplicate
	err = r.Register(counter)
	if err == nil {
		t.Error("Expected error when registering duplicate metric")
	}
}

func TestRegistryGet(t *testing.T) {
	r := NewRegistry("test")
	counter := NewCounter("requests_total", "Total requests")
	r.Register(counter)

	metric, exists := r.Get("requests_total")
	if !exists {
		t.Error("Expected metric to exist")
	}
	if metric == nil {
		t.Error("Expected metric to be returned")
	}

	_, exists = r.Get("non_existent")
	if exists {
		t.Error("Expected metric to not exist")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry("test")
	r.Register(NewCounter("counter1", "First counter"))
	r.Register(NewCounter("counter2", "Second counter"))

	metrics := r.List()
	if len(metrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(metrics))
	}
}

func TestNewCounter(t *testing.T) {
	c := NewCounter("requests_total", "Total requests")
	if c == nil {
		t.Fatal("Expected counter to be created")
	}
	if c.Name() != "requests_total" {
		t.Errorf("Expected name 'requests_total', got %s", c.Name())
	}
	if c.Type() != TypeCounter {
		t.Errorf("Expected type '%s', got %s", TypeCounter, c.Type())
	}
}

func TestCounterInc(t *testing.T) {
	c := NewCounter("requests_total", "Total requests")

	c.Inc()
	if c.Get() != 1 {
		t.Errorf("Expected value 1, got %d", c.Get())
	}

	c.Inc()
	if c.Get() != 2 {
		t.Errorf("Expected value 2, got %d", c.Get())
	}
}

func TestCounterAdd(t *testing.T) {
	c := NewCounter("requests_total", "Total requests")

	c.Add(5)
	if c.Get() != 5 {
		t.Errorf("Expected value 5, got %d", c.Get())
	}

	c.Add(3)
	if c.Get() != 8 {
		t.Errorf("Expected value 8, got %d", c.Get())
	}
}

func TestCounterReset(t *testing.T) {
	c := NewCounter("requests_total", "Total requests")
	c.Add(10)
	c.Reset()

	if c.Get() != 0 {
		t.Errorf("Expected value 0 after reset, got %d", c.Get())
	}
}

func TestNewGauge(t *testing.T) {
	g := NewGauge("temperature", "Current temperature")
	if g == nil {
		t.Fatal("Expected gauge to be created")
	}
	if g.Name() != "temperature" {
		t.Errorf("Expected name 'temperature', got %s", g.Name())
	}
	if g.Type() != TypeGauge {
		t.Errorf("Expected type '%s', got %s", TypeGauge, g.Type())
	}
}

func TestGaugeSet(t *testing.T) {
	g := NewGauge("temperature", "Current temperature")

	g.Set(25.5)
	if g.Get() != 25.5 {
		t.Errorf("Expected value 25.5, got %f", g.Get())
	}

	g.Set(30.0)
	if g.Get() != 30.0 {
		t.Errorf("Expected value 30.0, got %f", g.Get())
	}
}

func TestGaugeIncDec(t *testing.T) {
	g := NewGauge("value", "Test value")
	g.Set(10)

	g.Inc()
	if g.Get() != 11.0 {
		t.Errorf("Expected value 11, got %f", g.Get())
	}

	g.Dec()
	if g.Get() != 10.0 {
		t.Errorf("Expected value 10, got %f", g.Get())
	}
}

func TestGaugeAdd(t *testing.T) {
	g := NewGauge("value", "Test value")
	g.Set(10)

	g.Add(5.5)
	if g.Get() != 15.5 {
		t.Errorf("Expected value 15.5, got %f", g.Get())
	}
}

func TestNewHistogram(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0, 5.0}
	h := NewHistogram("request_duration", "Request duration", buckets)

	if h == nil {
		t.Fatal("Expected histogram to be created")
	}
	if h.Name() != "request_duration" {
		t.Errorf("Expected name 'request_duration', got %s", h.Name())
	}
	if h.Type() != TypeHistogram {
		t.Errorf("Expected type '%s', got %s", TypeHistogram, h.Type())
	}
}

func TestNewHistogramDefaultBuckets(t *testing.T) {
	// Create histogram with empty buckets (should use defaults)
	h := NewHistogram("request_duration", "Request duration", nil)

	buckets, _ := h.GetBuckets()
	if len(buckets) == 0 {
		t.Error("Expected default buckets to be set")
	}
}

func TestHistogramObserve(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0, 5.0}
	h := NewHistogram("request_duration", "Request duration", buckets)

	h.Observe(0.05)
	h.Observe(0.3)
	h.Observe(2.0)

	count := h.GetCount()
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	sum := h.GetSum()
	if sum < 2.3 || sum > 2.4 {
		t.Errorf("Expected sum ~2.35, got %f", sum)
	}
}

func TestHistogramBuckets(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0}
	h := NewHistogram("request_duration", "Request duration", buckets)

	h.Observe(0.05) // First bucket
	h.Observe(0.3)  // Second bucket
	h.Observe(2.0)  // Third bucket (inf)

	buckets, counts := h.GetBuckets()
	if len(buckets) != 3 {
		t.Errorf("Expected 3 buckets, got %d", len(buckets))
	}
	if len(counts) != 4 { // 3 + inf
		t.Errorf("Expected 4 counts, got %d", len(counts))
	}
}

func TestNewTimer(t *testing.T) {
	h := NewHistogram("duration", "Duration", []float64{0.1, 0.5, 1.0})

	timer := NewTimer(h)
	time.Sleep(10 * time.Millisecond)
	duration := timer.ObserveDuration()

	if duration == 0 {
		t.Error("Expected duration to be recorded")
	}

	count := h.GetCount()
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestNewTimerNil(t *testing.T) {
	timer := NewTimer(nil)
	time.Sleep(5 * time.Millisecond)
	duration := timer.ObserveDuration()

	if duration == 0 {
		t.Error("Expected duration to be recorded even with nil histogram")
	}
}

func TestRegisterCounter(t *testing.T) {
	c := RegisterCounter("test_counter", "Test counter")
	if c == nil {
		t.Fatal("Expected counter to be registered")
	}

	metric, exists := Get("test_counter")
	if !exists {
		t.Error("Expected counter to exist in default registry")
	}
	if metric == nil {
		t.Error("Expected counter to be returned")
	}
}

func TestRegisterGauge(t *testing.T) {
	g := RegisterGauge("test_gauge", "Test gauge")
	if g == nil {
		t.Fatal("Expected gauge to be registered")
	}

	metric, exists := Get("test_gauge")
	if !exists {
		t.Error("Expected gauge to exist in default registry")
	}
	if metric == nil {
		t.Error("Expected gauge to be returned")
	}
}

func TestRegisterHistogram(t *testing.T) {
	h := RegisterHistogram("test_histogram", "Test histogram", []float64{0.1, 0.5})
	if h == nil {
		t.Fatal("Expected histogram to be registered")
	}

	metric, exists := Get("test_histogram")
	if !exists {
		t.Error("Expected histogram to exist in default registry")
	}
	if metric == nil {
		t.Error("Expected histogram to be returned")
	}
}

func TestPrometheusExporter(t *testing.T) {
	r := NewRegistry("test")
	r.Register(NewCounter("requests", "Total requests"))
	r.Register(NewGauge("temperature", "Current temp"))
	r.Register(NewHistogram("duration", "Duration", []float64{0.1, 0.5}))

	exporter := NewPrometheusExporter(r)
	output := exporter.Export()

	if !strings.Contains(output, "requests") {
		t.Error("Expected output to contain 'requests'")
	}
	if !strings.Contains(output, "temperature") {
		t.Error("Expected output to contain 'temperature'")
	}
	if !strings.Contains(output, "duration") {
		t.Error("Expected output to contain 'duration'")
	}
	if !strings.Contains(output, "# HELP") {
		t.Error("Expected output to contain '# HELP'")
	}
	if !strings.Contains(output, "# TYPE") {
		t.Error("Expected output to contain '# TYPE'")
	}
}

func TestCounterConcurrency(t *testing.T) {
	c := NewCounter("concurrent", "Concurrent counter")

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				c.Inc()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if c.Get() != 1000 {
		t.Errorf("Expected value 1000, got %d", c.Get())
	}
}

func TestGaugeConcurrency(t *testing.T) {
	g := NewGauge("concurrent", "Concurrent gauge")
	g.Set(0)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				g.Inc()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if g.Get() != 1000 {
		t.Errorf("Expected value 1000, got %f", g.Get())
	}
}

func TestHistogramConcurrency(t *testing.T) {
	h := NewHistogram("concurrent", "Concurrent histogram", []float64{1, 10, 100})

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				h.Observe(float64(id))
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	count := h.GetCount()
	if count != 1000 {
		t.Errorf("Expected count 1000, got %d", count)
	}
}

func TestCounterDescribe(t *testing.T) {
	c := NewCounter("test", "Test description")
	if c.Describe() != "Test description" {
		t.Errorf("Expected description 'Test description', got %s", c.Describe())
	}
}

func TestGaugeDescribe(t *testing.T) {
	g := NewGauge("test", "Gauge description")
	if g.Describe() != "Gauge description" {
		t.Errorf("Expected description 'Gauge description', got %s", g.Describe())
	}
}

func TestHistogramDescribe(t *testing.T) {
	h := NewHistogram("test", "Histogram description", []float64{0.1})
	if h.Describe() != "Histogram description" {
		t.Errorf("Expected description 'Histogram description', got %s", h.Describe())
	}
}
