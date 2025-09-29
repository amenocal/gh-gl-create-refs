package gitlab

import (
	"testing"
)

func TestParseRepoPath(t *testing.T) {
	tests := []struct {
		name         string
		repoPath     string
		expectedBase string
		expectedPath string
		expectError  bool
	}{
		{
			name:         "simple group/repo",
			repoPath:     "group/repo",
			expectedBase: "",
			expectedPath: "group/repo",
			expectError:  false,
		},
		{
			name:         "nested subgroups",
			repoPath:     "group/subgroup/repo",
			expectedBase: "",
			expectedPath: "group/subgroup/repo",
			expectError:  false,
		},
		{
			name:         "multiple nested subgroups",
			repoPath:     "group/sub1/sub2/sub3/repo",
			expectedBase: "",
			expectedPath: "group/sub1/sub2/sub3/repo",
			expectError:  false,
		},
		{
			name:         "https URL",
			repoPath:     "https://gitlab.com/group/repo",
			expectedBase: "https://gitlab.com",
			expectedPath: "group/repo",
			expectError:  false,
		},
		{
			name:         "https URL with .git",
			repoPath:     "https://gitlab.com/group/repo.git",
			expectedBase: "https://gitlab.com",
			expectedPath: "group/repo",
			expectError:  false,
		},
		{
			name:         "custom gitlab instance",
			repoPath:     "https://gitlab.example.com/group/subgroup/repo",
			expectedBase: "https://gitlab.example.com",
			expectedPath: "group/subgroup/repo",
			expectError:  false,
		},
		{
			name:         "http URL",
			repoPath:     "http://gitlab.com/group/repo",
			expectedBase: "http://gitlab.com",
			expectedPath: "group/repo",
			expectError:  false,
		},
		{
			name:         "http URL with .git",
			repoPath:     "http://gitlab.com/group/repo.git",
			expectedBase: "http://gitlab.com",
			expectedPath: "group/repo",
			expectError:  false,
		},
		{
			name:         "http custom instance with nested groups",
			repoPath:     "http://gitlab.internal.com/group/subgroup/project",
			expectedBase: "http://gitlab.internal.com",
			expectedPath: "group/subgroup/project",
			expectError:  false,
		},
		{
			name:         "http with port number",
			repoPath:     "http://gitlab.local:8080/group/repo",
			expectedBase: "http://gitlab.local:8080",
			expectedPath: "group/repo",
			expectError:  false,
		},
		{
			name:        "invalid format - single word",
			repoPath:    "invalidrepo",
			expectError: true,
		},
		{
			name:        "invalid format - empty",
			repoPath:    "",
			expectError: true,
		},
		{
			name:        "invalid URL",
			repoPath:    "http://[invalid-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, path, err := ParseRepoPath(tt.repoPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("ParseRepoPath(%s) expected error, but got none", tt.repoPath)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseRepoPath(%s) unexpected error: %v", tt.repoPath, err)
				return
			}

			if baseURL != tt.expectedBase {
				t.Errorf("ParseRepoPath(%s) baseURL = %s, want %s", tt.repoPath, baseURL, tt.expectedBase)
			}

			if path != tt.expectedPath {
				t.Errorf("ParseRepoPath(%s) path = %s, want %s", tt.repoPath, path, tt.expectedPath)
			}
		})
	}
}

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
			result := GenerateBranchName(tt.prNumber)
			if result != tt.expected {
				t.Errorf("GenerateBranchName(%d) = %s, want %s", tt.prNumber, result, tt.expected)
			}
		})
	}
}
