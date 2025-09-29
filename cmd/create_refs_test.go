package cmd

import (
	"testing"
)

func TestGenerateBranchName(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		expected string
	}{
		{
			name:     "single digit PR",
			prNumber: 1,
			expected: "migration-pr-1",
		},
		{
			name:     "double digit PR",
			prNumber: 16,
			expected: "migration-pr-16",
		},
		{
			name:     "triple digit PR",
			prNumber: 123,
			expected: "migration-pr-123",
		},
		{
			name:     "four digit PR",
			prNumber: 1234,
			expected: "migration-pr-1234",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateBranchName(tt.prNumber)
			if result != tt.expected {
				t.Errorf("generateBranchName(%d) = %s, want %s", tt.prNumber, result, tt.expected)
			}
		})
	}
}