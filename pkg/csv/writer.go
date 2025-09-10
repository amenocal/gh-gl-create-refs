package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/amenocal/gh-gl-create-refs/pkg/gitlab"
)

// GenerateFilename creates a safe filename from repository path
func GenerateFilename(repoPath string) string {
	// Remove any URL prefixes and .git suffix
	name := repoPath
	if strings.Contains(name, "://") {
		parts := strings.Split(name, "/")
		if len(parts) >= 4 {
			name = strings.Join(parts[3:], "/")
		}
	}
	
	// Remove .git suffix if present
	name = strings.TrimSuffix(name, ".git")
	
	// Replace / with - to make it a valid filename
	name = strings.ReplaceAll(name, "/", "-")
	
	return fmt.Sprintf("%s.csv", name)
}

// WriteRefsToFile writes merge request references to a CSV file
func WriteRefsToFile(refs []gitlab.MergeRequestRef, filename string) error {
	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write each merge request reference
	for _, ref := range refs {
		record := []string{
			strconv.Itoa(ref.IID), // Use IID (internal ID) which is the MR number shown in GitLab UI
			ref.HeadSHA,
		}
		
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

// WriteRefsToCSV is a convenience function that generates filename and writes refs
func WriteRefsToCSV(refs []gitlab.MergeRequestRef, repoPath string) (string, error) {
	filename := GenerateFilename(repoPath)
	
	// Get absolute path for the output
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	err = WriteRefsToFile(refs, filename)
	if err != nil {
		return "", err
	}
	
	return absPath, nil
}