package cache

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New(100, time.Minute)
	if c == nil {
		t.Fatal("Expected cache to be created")
	}
	if c.maxSize != 100 {
		t.Errorf("Expected maxSize 100, got %d", c.maxSize)
	}
	if c.defaultTTL != time.Minute {
		t.Errorf("Expected defaultTTL 1m, got %v", c.defaultTTL)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	c := New(100, time.Minute)

	// Set a value
	c.Set("key1", "value1", time.Minute)

	// Get the value
	val, found := c.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}
}

func TestCacheGetNonExistent(t *testing.T) {
	c := New(100, time.Minute)

	val, found := c.Get("non-existent")
	if found {
		t.Error("Expected not to find non-existent key")
	}
	if val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}

func TestCacheExpiration(t *testing.T) {
	c := New(100, time.Hour)

	// Set with very short TTL
	c.Set("key1", "value1", 1*time.Millisecond)

	// Should exist immediately
	_, found := c.Get("key1")
	if !found {
		t.Error("Expected to find key1 immediately after set")
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Should be expired now
	_, found = c.Get("key1")
	if found {
		t.Error("Expected key1 to be expired")
	}
}

func TestCacheGetString(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("str-key", "string-value", time.Minute)

	val, found := c.GetString("str-key")
	if !found {
		t.Error("Expected to find str-key")
	}
	if val != "string-value" {
		t.Errorf("Expected 'string-value', got %s", val)
	}

	// Set non-string value
	c.Set("int-key", 123, time.Minute)
	_, found = c.GetString("int-key")
	if found {
		t.Error("Expected not to find int-key as string")
	}
}

func TestCacheGetBytes(t *testing.T) {
	c := New(100, time.Minute)

	data := []byte("byte-data")
	c.Set("bytes-key", data, time.Minute)

	val, found := c.GetBytes("bytes-key")
	if !found {
		t.Error("Expected to find bytes-key")
	}
	if string(val) != "byte-data" {
		t.Errorf("Expected 'byte-data', got %s", string(val))
	}

	// Set non-bytes value
	c.Set("str-key", "string", time.Minute)
	_, found = c.GetBytes("str-key")
	if found {
		t.Error("Expected not to find str-key as bytes")
	}
}

func TestCacheSetString(t *testing.T) {
	c := New(100, time.Minute)

	c.SetString("key1", "value1", time.Minute)

	val, found := c.GetString("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %s", val)
	}
}

func TestCacheSetBytes(t *testing.T) {
	c := New(100, time.Minute)

	c.SetBytes("key1", []byte("value1"), time.Minute)

	val, found := c.GetBytes("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if string(val) != "value1" {
		t.Errorf("Expected value1, got %s", string(val))
	}
}

func TestCacheDelete(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("key1", "value1", time.Minute)
	c.Delete("key1")

	_, found := c.Get("key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}
}

func TestCacheClear(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("key1", "value1", time.Minute)
	c.Set("key2", "value2", time.Minute)
	c.Clear()

	_, found1 := c.Get("key1")
	_, found2 := c.Get("key2")

	if found1 || found2 {
		t.Error("Expected cache to be cleared")
	}
}

func TestCacheSize(t *testing.T) {
	c := New(100, time.Minute)

	if c.Size() != 0 {
		t.Errorf("Expected size 0, got %d", c.Size())
	}

	c.Set("key1", "value1", time.Minute)
	c.Set("key2", "value2", time.Minute)

	if c.Size() != 2 {
		t.Errorf("Expected size 2, got %d", c.Size())
	}
}

func TestCacheExists(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("key1", "value1", time.Minute)

	if !c.Exists("key1") {
		t.Error("Expected key1 to exist")
	}

	if c.Exists("non-existent") {
		t.Error("Expected non-existent key to not exist")
	}
}

func TestCacheEviction(t *testing.T) {
	// Small cache to trigger eviction
	c := New(2, time.Minute)

	c.Set("key1", "value1", time.Minute)
	c.Set("key2", "value2", time.Minute)
	c.Set("key3", "value3", time.Minute) // Should evict key1

	// Check that oldest item was evicted
	_, found := c.Get("key1")
	if found {
		t.Error("Expected key1 to be evicted")
	}

	// Newer items should still exist
	_, found2 := c.Get("key2")
	_, found3 := c.Get("key3")
	if !found2 || !found3 {
		t.Error("Expected key2 and key3 to exist")
	}
}

func TestCacheDefaultTTL(t *testing.T) {
	c := New(100, 10*time.Millisecond)

	// Set without specifying TTL (should use default)
	c.Set("key1", "value1", 0)

	// Should exist immediately
	_, found := c.Get("key1")
	if !found {
		t.Error("Expected to find key1 immediately")
	}

	// Wait for default TTL to expire
	time.Sleep(20 * time.Millisecond)

	_, found = c.Get("key1")
	if found {
		t.Error("Expected key1 to be expired after default TTL")
	}
}

func TestCacheUpdateExisting(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("key1", "value1", time.Minute)
	c.Set("key1", "value2", time.Minute)

	val, found := c.Get("key1")
	if !found {
		t.Fatal("Expected to find key1")
	}
	if val != "value2" {
		t.Errorf("Expected value2, got %v", val)
	}
}

func TestCacheWithLoader(t *testing.T) {
	loader := func(key string) (any, error) {
		return "loaded-" + key, nil
	}

	c := NewWithLoader(100, time.Minute, loader)

	// Get should load the value
	val, err := c.GetOrLoad("key1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val != "loaded-key1" {
		t.Errorf("Expected 'loaded-key1', got %v", val)
	}

	// Second call should return cached value
	val2, err := c.GetOrLoad("key1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val2 != "loaded-key1" {
		t.Errorf("Expected 'loaded-key1', got %v", val2)
	}
}

func TestCacheStats(t *testing.T) {
	c := New(100, time.Minute)

	c.Set("key1", "value1", time.Minute)

	stats := c.Stats()
	if stats.Size != 1 {
		t.Errorf("Expected size 1, got %d", stats.Size)
	}
	if stats.MaxSize != 100 {
		t.Errorf("Expected maxSize 100, got %d", stats.MaxSize)
	}
}

func TestCacheExistsAfterExpiration(t *testing.T) {
	c := New(100, time.Hour)

	c.Set("key1", "value1", 1*time.Millisecond)

	// Should exist immediately
	if !c.Exists("key1") {
		t.Error("Expected key1 to exist immediately")
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Should not exist after expiration
	if c.Exists("key1") {
		t.Error("Expected key1 to not exist after expiration")
	}
}

// TestCacheCleanup tests manual cleanup of expired items
func TestCacheCleanup(t *testing.T) {
	c := New(100, time.Hour)

	// Add items with different TTLs
	c.Set("key1", "value1", 1*time.Millisecond)
	c.Set("key2", "value2", time.Hour)
	c.Set("key3", "value3", 1*time.Millisecond)

	// Verify all exist initially
	if !c.Exists("key1") || !c.Exists("key2") || !c.Exists("key3") {
		t.Error("Expected all keys to exist initially")
	}

	// Wait for short TTL items to expire
	time.Sleep(10 * time.Millisecond)

	// Trigger cleanup
	c.cleanup()

	// Verify expired items are removed
	if c.Exists("key1") {
		t.Error("Expected key1 to be cleaned up")
	}
	if c.Exists("key3") {
		t.Error("Expected key3 to be cleaned up")
	}
	if !c.Exists("key2") {
		t.Error("Expected key2 to still exist")
	}

	// Verify size
	if c.Size() != 1 {
		t.Errorf("Expected size 1 after cleanup, got %d", c.Size())
	}
}

// TestCacheCleanup_NoExpired tests cleanup with no expired items
func TestCacheCleanup_NoExpired(t *testing.T) {
	c := New(100, time.Hour)

	// Add items with long TTLs
	c.Set("key1", "value1", time.Hour)
	c.Set("key2", "value2", time.Hour)

	// Trigger cleanup
	c.cleanup()

	// Verify all items still exist
	if !c.Exists("key1") || !c.Exists("key2") {
		t.Error("Expected all keys to still exist")
	}

	if c.Size() != 2 {
		t.Errorf("Expected size 2, got %d", c.Size())
	}
}

// TestCacheCleanup_EmptyCache tests cleanup on empty cache
func TestCacheCleanup_EmptyCache(t *testing.T) {
	c := New(100, time.Hour)

	// Trigger cleanup on empty cache - should not panic
	c.cleanup()

	if c.Size() != 0 {
		t.Errorf("Expected size 0, got %d", c.Size())
	}
}

// TestNewTyped tests creating a typed cache
func TestNewTyped(t *testing.T) {
	// Create a typed cache for strings
	c := NewTyped[string](100, time.Minute)

	if c == nil {
		t.Fatal("Expected typed cache to be created")
	}

	if c.maxSize != 100 {
		t.Errorf("Expected maxSize 100, got %d", c.maxSize)
	}
}

// TestTypedCache_Get tests getting from typed cache
func TestTypedCache_Get(t *testing.T) {
	c := NewTyped[string](100, time.Minute)

	// Set a value using the underlying cache
	c.Cache.Set("key1", "value1", time.Minute)

	// Get using typed method
	val, found := c.Get("key1")
	if !found {
		t.Error("Expected key1 to be found")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	// Try to get non-existent key
	_, found = c.Get("nonexistent")
	if found {
		t.Error("Expected nonexistent key to not be found")
	}
}

// TestTypedCache_Get_WrongType tests getting wrong type from typed cache
func TestTypedCache_Get_WrongType(t *testing.T) {
	c := NewTyped[string](100, time.Minute)

	// Set an integer value directly (bypassing type safety)
	c.Cache.Set("key1", 123, time.Minute)

	// Get using typed method - should return empty value and false
	val, found := c.Get("key1")
	if found {
		t.Error("Expected type mismatch to return not found")
	}
	if val != "" {
		t.Errorf("Expected empty string, got %v", val)
	}
}

// TestTypedCache_Set tests setting in typed cache
func TestTypedCache_Set(t *testing.T) {
	c := NewTyped[string](100, time.Minute)

	// Set using typed method
	c.Set("key1", "value1", time.Minute)

	// Verify via underlying cache
	val, found := c.Cache.Get("key1")
	if !found {
		t.Error("Expected key1 to be found in underlying cache")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	// Verify via typed Get
	val2, found2 := c.Get("key1")
	if !found2 {
		t.Error("Expected key1 to be found via typed Get")
	}
	if val2 != "value1" {
		t.Errorf("Expected 'value1', got %v", val2)
	}
}

// TestTypedCache_Set_Overwrite tests overwriting in typed cache
func TestTypedCache_Set_Overwrite(t *testing.T) {
	c := NewTyped[string](100, time.Minute)

	c.Set("key1", "value1", time.Minute)
	c.Set("key1", "value2", time.Minute)

	val, found := c.Get("key1")
	if !found {
		t.Fatal("Expected key1 to be found")
	}
	if val != "value2" {
		t.Errorf("Expected 'value2', got %v", val)
	}
}

// TestTypedCache_Expiration tests typed cache expiration
func TestTypedCache_Expiration(t *testing.T) {
	c := NewTyped[string](100, time.Hour)

	c.Set("key1", "value1", 1*time.Millisecond)

	// Should exist immediately
	_, found := c.Get("key1")
	if !found {
		t.Error("Expected key1 to exist immediately")
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Should not exist after expiration
	_, found = c.Get("key1")
	if found {
		t.Error("Expected key1 to not exist after expiration")
	}
}

// TestTypedCache_IntType tests typed cache with int type
func TestTypedCache_IntType(t *testing.T) {
	c := NewTyped[int](100, time.Minute)

	c.Set("count", 42, time.Minute)

	val, found := c.Get("count")
	if !found {
		t.Fatal("Expected count to be found")
	}
	if val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}
}

// TestTypedCache_StructType tests typed cache with struct type
func TestTypedCache_StructType(t *testing.T) {
	type TestStruct struct {
		Name  string
		Value int
	}

	c := NewTyped[TestStruct](100, time.Minute)

	original := TestStruct{Name: "test", Value: 123}
	c.Set("struct1", original, time.Minute)

	val, found := c.Get("struct1")
	if !found {
		t.Fatal("Expected struct1 to be found")
	}
	if val.Name != "test" || val.Value != 123 {
		t.Errorf("Expected {test, 123}, got %v", val)
	}
}
