package core

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Field:   "server.port",
		Message: "must be between 1 and 65535",
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
	if err.Code() != 400 {
		t.Errorf("Expected code 400, got %d", err.Code())
	}
	if err.Slug() != "config_error" {
		t.Errorf("Expected slug 'config_error', got %s", err.Slug())
	}
}

func TestNotFoundError(t *testing.T) {
	err := &NotFoundError{
		Entity: "soul",
		ID:     "test-123",
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
	if err.Code() != 404 {
		t.Errorf("Expected code 404, got %d", err.Code())
	}
	if err.Slug() != "not_found" {
		t.Errorf("Expected slug 'not_found', got %s", err.Slug())
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "name",
		Message: "is required",
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
	if err.Code() != 400 {
		t.Errorf("Expected code 400, got %d", err.Code())
	}
	if err.Slug() != "validation_error" {
		t.Errorf("Expected slug 'validation_error', got %s", err.Slug())
	}
}

func TestConflictError(t *testing.T) {
	err := &ConflictError{
		Message: "resource already exists",
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
	if err.Code() != 409 {
		t.Errorf("Expected code 409, got %d", err.Code())
	}
	if err.Slug() != "conflict" {
		t.Errorf("Expected slug 'conflict', got %s", err.Slug())
	}
}

func TestUnauthorizedError(t *testing.T) {
	err := &UnauthorizedError{
		Message: "invalid API key",
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
	if err.Code() != 401 {
		t.Errorf("Expected code 401, got %d", err.Code())
	}
	if err.Slug() != "unauthorized" {
		t.Errorf("Expected slug 'unauthorized', got %s", err.Slug())
	}

	// Test empty message
	err2 := &UnauthorizedError{}
	if err2.Error() != "unauthorized" {
		t.Errorf("Expected 'unauthorized' for empty message, got %s", err2.Error())
	}
}

func TestForbiddenError(t *testing.T) {
	err := &ForbiddenError{
		Message: "insufficient permissions",
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
	if err.Code() != 403 {
		t.Errorf("Expected code 403, got %d", err.Code())
	}
	if err.Slug() != "forbidden" {
		t.Errorf("Expected slug 'forbidden', got %s", err.Slug())
	}

	// Test empty message
	err2 := &ForbiddenError{}
	if err2.Error() != "forbidden" {
		t.Errorf("Expected 'forbidden' for empty message, got %s", err2.Error())
	}
}

func TestInternalError(t *testing.T) {
	cause := errors.New("database connection failed")
	err := &InternalError{
		Message: "failed to save record",
		Cause:   cause,
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
	if err.Code() != 500 {
		t.Errorf("Expected code 500, got %d", err.Code())
	}
	if err.Slug() != "internal_error" {
		t.Errorf("Expected slug 'internal_error', got %s", err.Slug())
	}

	// Test without cause
	err2 := &InternalError{
		Message: "something went wrong",
	}
	if err2.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestULID_GenerateAndParse(t *testing.T) {
	// Generate ULID
	ulid, err := GenerateULID()
	if err != nil {
		t.Fatalf("GenerateULID failed: %v", err)
	}

	// Convert to string
	str := ulid.String()
	if len(str) != 26 {
		t.Errorf("Expected ULID string length 26, got %d", len(str))
	}

	// Parse back
	parsed, err := ParseULID(str)
	if err != nil {
		t.Fatalf("ParseULID failed: %v", err)
	}

	// Compare
	if ulid.Compare(parsed) != 0 {
		t.Error("Parsed ULID should match original")
	}
}

func TestULID_MarshalJSON(t *testing.T) {
	ulid, err := GenerateULID()
	if err != nil {
		t.Fatalf("GenerateULID failed: %v", err)
	}

	data, err := json.Marshal(ulid)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Should be quoted string
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		t.Errorf("Expected quoted JSON, got %s", string(data))
	}
}

func TestULID_UnmarshalJSON(t *testing.T) {
	ulid, err := GenerateULID()
	if err != nil {
		t.Fatalf("GenerateULID failed: %v", err)
	}

	str := ulid.String()
	jsonData := []byte(`"` + str + `"`)

	var parsed ULID
	err = parsed.UnmarshalJSON(jsonData)
	if err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if ulid.Compare(parsed) != 0 {
		t.Error("Unmarshaled ULID should match original")
	}
}

func TestULID_UnmarshalJSON_Invalid(t *testing.T) {
	var ulid ULID
	err := ulid.UnmarshalJSON([]byte("invalid"))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	err = ulid.UnmarshalJSON([]byte(`"tooshort"`))
	if err == nil {
		t.Error("Expected error for invalid ULID string")
	}
}

func TestULID_Time(t *testing.T) {
	now := time.Now().UTC()
	ulid, err := GenerateULIDAt(now)
	if err != nil {
		t.Fatalf("GenerateULIDAt failed: %v", err)
	}

	timestamp := ulid.Time()
	// Allow 1 second tolerance for millisecond precision
	diff := timestamp.Sub(now).Abs()
	if diff > time.Second {
		t.Errorf("Timestamp difference too large: got %v, want < 1s", diff)
	}
}

func TestULID_Compare(t *testing.T) {
	ulid1, _ := GenerateULID()
	ulid2, _ := GenerateULID()

	// Same ULID should compare equal
	if ulid1.Compare(ulid1) != 0 {
		t.Error("Same ULID should compare equal")
	}

	// Different ULIDs should have non-zero comparison
	// (we can't predict which is greater due to randomness)
	if ulid1.Compare(ulid2) == 0 && ulid1 != ulid2 {
		t.Error("Different ULIDs should not compare equal")
	}
}

func TestMustGenerateULID(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustGenerateULID panicked: %v", r)
		}
	}()

	ulid := MustGenerateULID()
	if ulid == ZeroULID {
		t.Error("Expected non-zero ULID")
	}
}

func TestGenerateID(t *testing.T) {
	id := GenerateID()
	if len(id) != 26 {
		t.Errorf("Expected ID length 26, got %d", len(id))
	}

	// Should be parseable
	_, err := ParseULID(id)
	if err != nil {
		t.Errorf("Generated ID should be parseable: %v", err)
	}
}

func TestULID_MarshalText(t *testing.T) {
	ulid, _ := GenerateULID()

	data, err := ulid.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText failed: %v", err)
	}

	if string(data) != ulid.String() {
		t.Errorf("Marshaled text should match String()")
	}
}

func TestULID_UnmarshalText(t *testing.T) {
	ulid, _ := GenerateULID()
	str := ulid.String()

	var parsed ULID
	err := parsed.UnmarshalText([]byte(str))
	if err != nil {
		t.Fatalf("UnmarshalText failed: %v", err)
	}

	if ulid.Compare(parsed) != 0 {
		t.Error("Unmarshaled ULID should match original")
	}

	// Invalid text
	err = parsed.UnmarshalText([]byte("invalid"))
	if err == nil {
		t.Error("Expected error for invalid text")
	}
}
