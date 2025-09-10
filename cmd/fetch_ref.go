package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/amenocal/gh-gl-create-refs/pkg/csv"
	"github.com/amenocal/gh-gl-create-refs/pkg/gitlab"
	"github.com/spf13/cobra"
)

var fetchRefCmd = &cobra.Command{
	Use:   "fetch-ref [repository]",
	Short: "Fetch merge request references from a GitLab repository",
	Long: `Fetch all merge request references from a GitLab repository and output them to a CSV file.

The repository can be specified in various formats:
- Full URL: https://gitlab.com/group/project
- Group/project: group/project
- Nested groups: group/subgroup/project or group/subgroup/subgroup/project

The output CSV file will contain two columns:
1. Merge request number (IID)
2. Head SHA from diff_refs

Examples:
  gh gl-create-refs fetch-ref group/project
  gh gl-create-refs fetch-ref https://gitlab.example.com/group/subgroup/project
  gh gl-create-refs fetch-ref group/subgroup/subgroup/project`,
	Args: cobra.ExactArgs(1),
	RunE: runFetchRef,
}

var (
	gitlabToken   string
	gitlabBaseURL string
	outputFile    string
)

func init() {
	rootCmd.AddCommand(fetchRefCmd)

	fetchRefCmd.Flags().StringVarP(&gitlabToken, "token", "t", "", "GitLab access token (can also use GITLAB_TOKEN environment variable)")
	fetchRefCmd.Flags().StringVarP(&gitlabBaseURL, "base-url", "b", "", "GitLab base URL (default: https://gitlab.com)")
	fetchRefCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output CSV file path (default: auto-generated from repository name)")
}

func runFetchRef(cmd *cobra.Command, args []string) error {
	repoPath := args[0]

	// Get token from flag or environment variable
	token := gitlabToken
	if token == "" {
		token = os.Getenv("GITLAB_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("GitLab token is required. Use --token flag or set GITLAB_TOKEN environment variable")
	}

	// Parse repository path and determine base URL
	baseURL, projectPath, err := gitlab.ParseRepoPath(repoPath)
	if err != nil {
		return fmt.Errorf("failed to parse repository path: %w", err)
	}

	// Use provided base URL or the one parsed from the repo path
	if gitlabBaseURL != "" {
		baseURL = gitlabBaseURL
	}

	// Create GitLab client
	client, err := gitlab.NewClient(token, baseURL)
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}

	fmt.Printf("Fetching merge requests from %s...\n", projectPath)

	// Fetch merge request references
	refs, err := client.FetchMergeRequestRefs(projectPath)
	if err != nil {
		// Provide more helpful error messages for common issues
		errMsg := err.Error()
		if strings.Contains(errMsg, "404") {
			return fmt.Errorf("repository not found: %s. Please check the repository path and your access permissions", projectPath)
		}
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") {
			return fmt.Errorf("authentication failed: please check your GitLab token has access to repository %s", projectPath)
		}
		return fmt.Errorf("failed to fetch merge request references from %s: %w", projectPath, err)
	}

	if len(refs) == 0 {
		fmt.Println("No merge requests found in the repository")
		return nil
	}

	fmt.Printf("Found %d merge requests\n", len(refs))

	// Generate output file path
	var outputPath string
	if outputFile != "" {
		outputPath = outputFile
	} else {
		outputPath, err = csv.WriteRefsToCSV(refs, repoPath)
		if err != nil {
			return fmt.Errorf("failed to write CSV file: %w", err)
		}
	}

	// If custom output file is specified, write to it
	if outputFile != "" {
		err = csv.WriteRefsToFile(refs, outputFile)
		if err != nil {
			return fmt.Errorf("failed to write CSV file: %w", err)
		}
		outputPath = outputFile
	}

	fmt.Printf("Successfully exported merge request references to: %s\n", outputPath)

	return nil
}