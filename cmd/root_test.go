package cmd

import (
	"testing"
)

func TestDerefStr(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{"nil pointer returns dash", nil, "-"},
		{"non-nil pointer returns value", strPtr("hello"), "hello"},
		{"empty string returns empty", strPtr(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := derefStr(tt.input)
			if result != tt.expected {
				t.Errorf("derefStr() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDerefInt32(t *testing.T) {
	tests := []struct {
		name     string
		input    *int32
		expected int32
	}{
		{"nil pointer returns 0", nil, 0},
		{"non-nil pointer returns value", int32Ptr(512), 512},
		{"zero value returns 0", int32Ptr(0), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := derefInt32(tt.input)
			if result != tt.expected {
				t.Errorf("derefInt32() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestParseGenericFilters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{"empty string", "", map[string]string{}},
		{"single filter", "engine=postgres", map[string]string{"engine": "postgres"}},
		{"multiple filters", "engine=postgres,status=available", map[string]string{"engine": "postgres", "status": "available"}},
		{"value with special chars", "prefix=prod/db-", map[string]string{"prefix": "prod/db-"}},
		{"invalid pair ignored", "noequals", map[string]string{}},
		{"mixed valid and invalid", "engine=mysql,bad,status=running", map[string]string{"engine": "mysql", "status": "running"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGenericFilters(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseGenericFilters(%q) returned %d items, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("parseGenericFilters(%q)[%q] = %q, want %q", tt.input, k, result[k], v)
				}
			}
		})
	}
}

// Helper functions for creating pointers in tests
func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
