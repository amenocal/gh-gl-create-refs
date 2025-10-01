package csv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amenocal/gh-gl-create-refs/pkg/gitlab"
)

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		expected string
	}{
		{
			name:     "simple group/repo",
			repoPath: "group/repo",
			expected: "group-repo.csv",
		},
		{
			name:     "nested subgroups",
			repoPath: "group/subgroup/repo",
			expected: "group-subgroup-repo.csv",
		},
		{
			name:     "multiple nested subgroups",
			repoPath: "group/sub1/sub2/sub3/repo",
			expected: "group-sub1-sub2-sub3-repo.csv",
		},
		{
			name:     "URL with .git suffix",
			repoPath: "https://gitlab.com/group/repo.git",
			expected: "group-repo.csv",
		},
		{
			name:     "URL without .git suffix",
			repoPath: "https://gitlab.com/group/subgroup/repo",
			expected: "group-subgroup-repo.csv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFilename(tt.repoPath)
			if result != tt.expected {
				t.Errorf("GenerateFilename(%s) = %s, want %s", tt.repoPath, result, tt.expected)
			}
		})
	}
}

func TestWriteRefsToFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")

	// Test data
	refs := []gitlab.MergeRequestRef{
		{ID: 1, IID: 1, HeadSHA: "abc123"},
		{ID: 2, IID: 16, HeadSHA: "def456"},
		{ID: 3, IID: 17, HeadSHA: "ghi789"},
	}

	// Write refs to file
	err := WriteRefsToFile(refs, testFile)
	if err != nil {
		t.Fatalf("WriteRefsToFile failed: %v", err)
	}

	// Read the file and verify content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	expected := "1,abc123\n16,def456\n17,ghi789\n"
	if string(content) != expected {
		t.Errorf("File content = %q, want %q", string(content), expected)
	}
}

func TestReadRefsFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.csv")

	// Create test CSV content
	content := "1,abc123\n16,def456\n17,ghi789\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Read refs from file
	refs, err := ReadRefsFromFile(testFile)
	if err != nil {
		t.Fatalf("ReadRefsFromFile failed: %v", err)
	}

	// Verify the content
	expected := []gitlab.MergeRequestRef{
		{IID: 1, HeadSHA: "abc123"},
		{IID: 16, HeadSHA: "def456"},
		{IID: 17, HeadSHA: "ghi789"},
	}

	if len(refs) != len(expected) {
		t.Fatalf("Expected %d refs, got %d", len(expected), len(refs))
	}

	for i, ref := range refs {
		if ref.IID != expected[i].IID || ref.HeadSHA != expected[i].HeadSHA {
			t.Errorf("Ref %d: expected IID=%d SHA=%s, got IID=%d SHA=%s", 
				i, expected[i].IID, expected[i].HeadSHA, ref.IID, ref.HeadSHA)
		}
	}
}

func TestReadRefsFromFile_InvalidFormat(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid.csv")

	// Test with invalid number of columns
	content := "1,abc123,extra\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = ReadRefsFromFile(testFile)
	if err == nil {
		t.Fatal("Expected error for invalid CSV format, got nil")
	}
}

func TestReadRefsFromFile_InvalidIID(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "invalid_iid.csv")

	// Test with invalid IID
	content := "not_a_number,abc123\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = ReadRefsFromFile(testFile)
	if err == nil {
		t.Fatal("Expected error for invalid IID, got nil")
	}
}