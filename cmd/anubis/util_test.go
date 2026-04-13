package main

import (
	"testing"
)

func TestParseConfigValue_Float(t *testing.T) {
	// Float value
	result := parseConfigValue("3.14")
	if f, ok := result.(float64); !ok || f != 3.14 {
		t.Errorf("Expected float64 3.14, got %v (%T)", result, result)
	}
}
