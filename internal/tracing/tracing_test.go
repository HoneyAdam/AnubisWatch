package tracing

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewTracer(t *testing.T) {
	exporter := &ConsoleExporter{}
	tracer := NewTracer(true, exporter)

	if tracer == nil {
		t.Fatal("Expected tracer to be created")
	}
	if !tracer.enabled {
		t.Error("Expected tracer to be enabled")
	}
	if tracer.exporter != exporter {
		t.Error("Expected exporter to be set")
	}
}

func TestNewTracerDisabled(t *testing.T) {
	tracer := NewTracer(false, nil)

	if tracer == nil {
		t.Fatal("Expected tracer to be created")
	}
	if tracer.enabled {
		t.Error("Expected tracer to be disabled")
	}
}

func TestTracerStartSpan(t *testing.T) {
	exporter := &ConsoleExporter{}
	tracer := NewTracer(true, exporter)
	ctx := context.Background()

	ctx, span := tracer.StartSpan(ctx, "test-operation")
	if span == nil {
		t.Fatal("Expected span to be created")
	}

	if span.Name != "test-operation" {
		t.Errorf("Expected name 'test-operation', got %s", span.Name)
	}
	if span.TraceID == "" {
		t.Error("Expected TraceID to be set")
	}
	if span.SpanID == "" {
		t.Error("Expected SpanID to be set")
	}
	if span.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}
	if span.Status != StatusOK {
		t.Errorf("Expected status OK, got %v", span.Status)
	}

	// Check context contains span
	ctxSpan := SpanFromContext(ctx)
	if ctxSpan == nil {
		t.Error("Expected span to be in context")
	}
	if ctxSpan != span {
		t.Error("Expected context span to match returned span")
	}
}

func TestTracerStartSpanDisabled(t *testing.T) {
	tracer := NewTracer(false, nil)
	ctx := context.Background()

	ctx, span := tracer.StartSpan(ctx, "test-operation")
	if span != nil {
		t.Error("Expected no span when tracer is disabled")
	}

	// Check context does not contain span
	ctxSpan := SpanFromContext(ctx)
	if ctxSpan != nil {
		t.Error("Expected no span in context when tracer is disabled")
	}
}

func TestTracerStartSpanWithParent(t *testing.T) {
	exporter := &ConsoleExporter{}
	tracer := NewTracer(true, exporter)
	ctx := context.Background()

	// Create parent span
	ctx, parentSpan := tracer.StartSpan(ctx, "parent-operation")

	// Create child span
	_, childSpan := tracer.StartSpan(ctx, "child-operation")

	if childSpan.ParentID != parentSpan.SpanID {
		t.Errorf("Expected ParentID to be %s, got %s", parentSpan.SpanID, childSpan.ParentID)
	}

	if childSpan.TraceID != parentSpan.TraceID {
		t.Errorf("Expected TraceID to match parent, got %s", childSpan.TraceID)
	}
}

func TestTracerFinish(t *testing.T) {
	exporter := NewMemoryExporter()
	tracer := NewTracer(true, exporter)
	ctx := context.Background()

	_, span := tracer.StartSpan(ctx, "test-operation")
	time.Sleep(10 * time.Millisecond)
	tracer.Finish(span)

	if span.EndTime.IsZero() {
		t.Error("Expected EndTime to be set")
	}
	if span.Duration == 0 {
		t.Error("Expected Duration to be set")
	}
}

func TestSpanSetAttribute(t *testing.T) {
	span := &Span{
		Attributes: make(map[string]any),
	}

	span.SetAttribute("key1", "value1")
	span.SetAttribute("key2", 123)

	if span.Attributes["key1"] != "value1" {
		t.Errorf("Expected attribute 'key1' to be 'value1', got %v", span.Attributes["key1"])
	}
	if span.Attributes["key2"] != 123 {
		t.Errorf("Expected attribute 'key2' to be 123, got %v", span.Attributes["key2"])
	}
}

func TestSpanSetAttributeNil(t *testing.T) {
	var span *Span
	span.SetAttribute("key", "value") // Should not panic
}

func TestSpanLog(t *testing.T) {
	span := &Span{
		Logs: make([]LogEntry, 0),
	}

	fields := map[string]any{
		"event": "test-event",
		"data":  "test-data",
	}
	span.Log(fields)

	if len(span.Logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(span.Logs))
	}

	entry := span.Logs[0]
	if entry.Timestamp.IsZero() {
		t.Error("Expected log timestamp to be set")
	}
	if entry.Fields["event"] != "test-event" {
		t.Errorf("Expected event 'test-event', got %v", entry.Fields["event"])
	}
}

func TestSpanLogNil(t *testing.T) {
	var span *Span
	span.Log(map[string]any{"event": "test"}) // Should not panic
}

func TestSpanSetError(t *testing.T) {
	span := &Span{
		Attributes: make(map[string]any),
		Status:     StatusOK,
	}

	testErr := errors.New("test error")
	span.SetError(testErr)

	if span.Status != StatusError {
		t.Errorf("Expected status Error, got %v", span.Status)
	}
	if span.Attributes["error"] != "test error" {
		t.Errorf("Expected error attribute 'test error', got %v", span.Attributes["error"])
	}
}

func TestSpanSetErrorNil(t *testing.T) {
	var span *Span
	span.SetError(errors.New("test")) // Should not panic
}

func TestSpanSetTag(t *testing.T) {
	span := &Span{
		Tags: make(map[string]string),
	}

	span.SetTag("tag1", "value1")

	if span.Tags["tag1"] != "value1" {
		t.Errorf("Expected tag 'tag1' to be 'value1', got %v", span.Tags["tag1"])
	}
}

func TestSpanSetTagNil(t *testing.T) {
	var span *Span
	span.SetTag("key", "value") // Should not panic
}

func TestSpanStatusString(t *testing.T) {
	tests := []struct {
		status   SpanStatus
		expected string
	}{
		{StatusOK, "ok"},
		{StatusError, "error"},
		{StatusCancelled, "cancelled"},
		{SpanStatus(999), "unknown"},
	}

	for _, test := range tests {
		result := test.status.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestContextWithSpan(t *testing.T) {
	span := &Span{Name: "test-span"}
	ctx := ContextWithSpan(context.Background(), span)

	retrieved := SpanFromContext(ctx)
	if retrieved == nil {
		t.Fatal("Expected to retrieve span from context")
	}
	if retrieved != span {
		t.Error("Expected retrieved span to match original")
	}
}

func TestContextWithSpanNil(t *testing.T) {
	ctx := ContextWithSpan(context.Background(), nil)

	retrieved := SpanFromContext(ctx)
	if retrieved != nil {
		t.Error("Expected no span in context when passing nil")
	}
}

func TestSpanFromContextNil(t *testing.T) {
	span := SpanFromContext(nil)
	if span != nil {
		t.Error("Expected nil span from nil context")
	}
}

func TestSpanFromContextNoSpan(t *testing.T) {
	ctx := context.Background()
	span := SpanFromContext(ctx)
	if span != nil {
		t.Error("Expected nil span from context without span")
	}
}

func TestMemoryExporter(t *testing.T) {
	exporter := NewMemoryExporter()
	tracer := NewTracer(true, exporter)
	ctx := context.Background()

	_, span1 := tracer.StartSpan(ctx, "span1")
	tracer.Finish(span1)

	// Wait for async export
	time.Sleep(10 * time.Millisecond)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Errorf("Expected 1 span, got %d", len(spans))
	}

	exporter.Clear()
	spans = exporter.GetSpans()
	if len(spans) != 0 {
		t.Errorf("Expected 0 spans after clear, got %d", len(spans))
	}
}

func TestTracedFunction(t *testing.T) {
	exporter := NewMemoryExporter()
	tracer := NewTracer(true, exporter)

	called := false
	fn := TracedFunction(tracer, "test-fn", func(ctx context.Context) error {
		called = true
		return nil
	})

	err := fn(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !called {
		t.Error("Expected function to be called")
	}
}

func TestTracedFunctionWithError(t *testing.T) {
	exporter := NewMemoryExporter()
	tracer := NewTracer(true, exporter)

	testErr := errors.New("test error")
	fn := TracedFunction(tracer, "test-fn", func(ctx context.Context) error {
		return testErr
	})

	err := fn(context.Background())
	if err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, err)
	}
}

func TestTracedFunctionDisabled(t *testing.T) {
	tracer := NewTracer(false, nil)

	called := false
	fn := TracedFunction(tracer, "test-fn", func(ctx context.Context) error {
		called = true
		return nil
	})

	err := fn(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !called {
		t.Error("Expected function to be called even when tracer disabled")
	}
}

func TestNoopTracer(t *testing.T) {
	tracer := NoopTracer()
	ctx := context.Background()

	_, span := tracer.StartSpan(ctx, "test")
	if span != nil {
		t.Error("Expected nil span from noop tracer")
	}
}

func TestGetStats(t *testing.T) {
	exporter := NewMemoryExporter()
	tracer := NewTracer(true, exporter)
	ctx := context.Background()

	// Create some spans
	_, span1 := tracer.StartSpan(ctx, "span1")
	time.Sleep(5 * time.Millisecond)
	tracer.Finish(span1)

	_, span2 := tracer.StartSpan(ctx, "span2")
	span2.SetError(errors.New("test error"))
	time.Sleep(5 * time.Millisecond)
	tracer.Finish(span2)

	// Wait for async export
	time.Sleep(10 * time.Millisecond)

	stats := tracer.GetStats()
	if stats.TotalSpans != 2 {
		t.Errorf("Expected 2 total spans, got %d", stats.TotalSpans)
	}
	if stats.TotalErrors != 1 {
		t.Errorf("Expected 1 error span, got %d", stats.TotalErrors)
	}
	if stats.AvgDuration == 0 {
		t.Error("Expected non-zero average duration")
	}
}

func TestGetStatsEmpty(t *testing.T) {
	tracer := NewTracer(true, nil)

	stats := tracer.GetStats()
	if stats.TotalSpans != 0 {
		t.Errorf("Expected 0 total spans, got %d", stats.TotalSpans)
	}
	if stats.MinDuration != 0 {
		t.Errorf("Expected 0 min duration for empty spans, got %d", stats.MinDuration)
	}
}

func TestConsoleExporter(t *testing.T) {
	exporter := &ConsoleExporter{}
	span := &Span{
		TraceID:   "trace-12345678",
		SpanID:    "span-12345678",
		Name:      "test-span",
		Duration:  100,
		Status:    StatusOK,
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}

	err := exporter.Export(span)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTracerFinishNil(t *testing.T) {
	tracer := NewTracer(true, nil)
	tracer.Finish(nil) // Should not panic
}

func TestTracerFinishDisabled(t *testing.T) {
	tracer := NewTracer(false, nil)
	span := &Span{Name: "test"}
	tracer.Finish(span) // Should not panic
}
