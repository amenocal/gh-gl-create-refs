package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/amenocal/gh-gl-create-refs/pkg/csv"
	"github.com/amenocal/gh-gl-create-refs/pkg/gitlab"
	"github.com/spf13/cobra"
)

var fetchRefCmd = &cobra.Command{
	Use:   "fetch-refs",
	Short: "Fetch merge request references from a GitLab repository",
	Long: `Fetch all merge request references from a GitLab repository and output them to a CSV file.

The repository can be specified using the --repository flag in various formats:
- Full URL: https://gitlab.com/group/project
- Group/project: group/project
- Nested groups: group/subgroup/project or group/subgroup/subgroup/project

The output CSV file will contain two columns:
1. Merge request number (IID)
2. Head SHA from diff_refs

Examples:
  gh gl-create-refs fetch-refs --repository group/project
  gh gl-create-refs fetch-refs --repository https://gitlab.example.com/group/subgroup/project
  gh gl-create-refs fetch-refs -r group/subgroup/subgroup/project`,
	Args: cobra.NoArgs,
	RunE: runFetchRef,
}

var (
	gitlabToken   string
	gitlabBaseURL string
	outputFile    string
	repository    string
)

func init() {
	rootCmd.AddCommand(fetchRefCmd)

	fetchRefCmd.Flags().StringVarP(&gitlabToken, "token", "t", "", "GitLab access token (can also use GITLAB_TOKEN environment variable)")
	fetchRefCmd.Flags().StringVarP(&gitlabBaseURL, "base-url", "b", "", "GitLab base URL (default: https://gitlab.com)")
	fetchRefCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output CSV file path (default: auto-generated from repository name)")
	fetchRefCmd.Flags().StringVarP(&repository, "repository", "r", "", "GitLab repository path (required)")

	// Mark the repository flag as required
	fetchRefCmd.MarkFlagRequired("repository")
}

func runFetchRef(cmd *cobra.Command, args []string) error {
	repoPath := repository

	// Create GitLab client from flags and environment
	client, err := gitlab.NewClient(gitlabToken, gitlabBaseURL)
	if err != nil {
		return err
	}

	fmt.Printf("Fetching merge requests from repository...\n")

	// Determine output file path
	var outputPath string
	if outputFile != "" {
		outputPath = outputFile
	} else {
		outputPath = csv.GenerateFilename(repoPath)
	}

	// Create CSV stream writer for incremental writing
	csvWriter, err := csv.NewStreamWriter(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV writer: %w", err)
	}
	defer csvWriter.Close()

	// Track progress
	refCount := 0

	// Create processor callback that writes each MR to CSV immediately
	processor := func(ref gitlab.MergeRequestRef) error {
		if err := csvWriter.WriteRef(ref); err != nil {
			return fmt.Errorf("failed to write merge request %d to CSV: %w", ref.IID, err)
		}
		refCount++
		return nil
	}

	// Fetch merge request references using the callback-based API
	projectPath, err := client.FetchMergeRequestRefsFromRepo(repoPath, gitlabBaseURL, processor)
	if err != nil {
		return err
	}

	if refCount == 0 {
		fmt.Printf("No merge requests found in %s\n", projectPath)
		return nil
	}

	fmt.Printf("Found %d merge requests from %s\n", refCount, projectPath)

	// Get absolute path for the output
	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath // Fallback to relative path
	}

	fmt.Printf("Successfully exported merge request references to: %s\n", absPath)

	return nil
}
