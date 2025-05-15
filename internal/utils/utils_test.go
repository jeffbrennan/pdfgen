package utils

import (
	"testing"
)

func TestValidateRepo(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected string
	}{
		{
			"SUPER_SECRET should return",
			"SUPER_SECRET",
			"SHHH"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			value, _ := LoadSecret(tt.input)
			if value != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, value)
			}
		})
	}
}
