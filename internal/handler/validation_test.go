package handler

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateMonthYear(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid date", "01-2025", true},
		{"Valid date with single digit month", "1-2025", false}, // должен быть 01
		{"Invalid month", "13-2025", false},
		{"Invalid month", "00-2025", false},
		{"Invalid year", "01-20", false},
		{"Invalid format", "2025-01", false},
		{"Invalid format", "01/2025", false},
		{"Invalid characters", "MM-YYYY", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMonthYear(tt.input)
			assert.Equal(t, tt.expected, result, "For input: %s", tt.input)
		})
	}
}

func TestParseMonthYear(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{"Valid date", "01-2025", false},
		{"Valid date December", "12-2025", false},
		{"Invalid month", "13-2025", true},
		{"Invalid format", "2025-01", true},
		{"Invalid characters", "MM-YYYY", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseMonthYear(tt.input)
			if tt.shouldError {
				assert.Error(t, err, "Expected error for input: %s", tt.input)
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", tt.input)
			}
		})
	}
}
