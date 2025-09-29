package cmd

import (
	"fmt"

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

	// Fetch merge request references using the client
	refs, projectPath, err := client.FetchMergeRequestRefsFromRepo(repoPath, gitlabBaseURL)
	if err != nil {
		return err
	}

	if len(refs) == 0 {
		fmt.Printf("No merge requests found in %s\n", projectPath)
		return nil
	}

	fmt.Printf("Found %d merge requests from %s\n", len(refs), projectPath)

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
