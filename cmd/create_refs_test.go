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
		{
			name:     "zero PR number",
			prNumber: 0,
			expected: "migration-pr-0",
		},
		{
			name:     "large PR number",
			prNumber: 999999,
			expected: "migration-pr-999999",
		},
		{
			name:     "negative PR number",
			prNumber: -1,
			expected: "migration-pr--1",
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

func TestValidateCreateRefsFlags(t *testing.T) {
	tests := []struct {
		name       string
		repository string
		fetch      bool
		inputFile  string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid with input file",
			repository: "group/project",
			fetch:      false,
			inputFile:  "test.csv",
			wantErr:    false,
		},
		{
			name:       "valid with fetch mode",
			repository: "group/project",
			fetch:      true,
			inputFile:  "",
			wantErr:    false,
		},
		{
			name:       "valid with fetch mode and input file",
			repository: "group/project",
			fetch:      true,
			inputFile:  "test.csv",
			wantErr:    false,
		},
		{
			name:       "missing repository",
			repository: "",
			fetch:      false,
			inputFile:  "test.csv",
			wantErr:    true,
			errMsg:     "--repository is required",
		},
		{
			name:       "missing input file without fetch",
			repository: "group/project",
			fetch:      false,
			inputFile:  "",
			wantErr:    true,
			errMsg:     "--input is required unless --fetch is used",
		},
		{
			name:       "missing repository and input file",
			repository: "",
			fetch:      false,
			inputFile:  "",
			wantErr:    true,
			errMsg:     "--repository is required",
		},
		{
			name:       "whitespace repository",
			repository: "   ",
			fetch:      false,
			inputFile:  "test.csv",
			wantErr:    false, // Note: this validates as valid since we only check if it's empty string
		},
		{
			name:       "whitespace input file without fetch",
			repository: "group/project",
			fetch:      false,
			inputFile:  "   ",
			wantErr:    false, // Note: this validates as valid since we only check if it's empty string
		},
		{
			name:       "complex repository path",
			repository: "group/subgroup/project-name",
			fetch:      true,
			inputFile:  "",
			wantErr:    false,
		},
		{
			name:       "repository with special characters",
			repository: "group-1/project_name.test",
			fetch:      false,
			inputFile:  "data/test-file.csv",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateRefsFlags(tt.repository, tt.fetch, tt.inputFile)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateCreateRefsFlags() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("validateCreateRefsFlags() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateCreateRefsFlags() unexpected error = %v", err)
				}
			}
		})
	}
}
