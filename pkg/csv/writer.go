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

// StreamWriter handles incremental writing of merge request references to CSV
type StreamWriter struct {
	file   *os.File
	writer *csv.Writer
}

// NewStreamWriter creates a new CSV stream writer for incremental writing
func NewStreamWriter(filename string) (*StreamWriter, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %s: %w", filename, err)
	}

	writer := csv.NewWriter(file)

	return &StreamWriter{
		file:   file,
		writer: writer,
	}, nil
}

// WriteRef writes a single merge request reference to the CSV file
func (sw *StreamWriter) WriteRef(ref gitlab.MergeRequestRef) error {
	record := []string{
		strconv.Itoa(ref.IID), // Use IID (internal ID) which is the MR number shown in GitLab UI
		ref.HeadSHA,
	}

	if err := sw.writer.Write(record); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}

	// Flush after each write to ensure data is written immediately
	sw.writer.Flush()

	return sw.writer.Error()
}

// Close closes the CSV writer and file
func (sw *StreamWriter) Close() error {
	sw.writer.Flush()
	if err := sw.writer.Error(); err != nil {
		sw.file.Close() // Still close the file even if flush fails
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}

	if err := sw.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// ReadRefsFromFile reads merge request references from a CSV file
func ReadRefsFromFile(filename string) ([]gitlab.MergeRequestRef, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV file: %w", err)
	}

	var refs []gitlab.MergeRequestRef
	for i, record := range records {
		if len(record) != 2 {
			return nil, fmt.Errorf("invalid CSV format at line %d: expected 2 columns, got %d", i+1, len(record))
		}

		iid, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("invalid merge request IID at line %d: %w", i+1, err)
		}

		refs = append(refs, gitlab.MergeRequestRef{
			IID:     iid,
			HeadSHA: record[1],
		})
	}

	return refs, nil
}
