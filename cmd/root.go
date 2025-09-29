package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-gl-create-refs",
	Short: "A GitHub CLI extension to work with GitLab repository references",
	Long: `gh-gl-create-refs is a GitHub CLI extension that provides utilities to work with GitLab repository references.
It can fetch merge request references from GitLab and export them in various formats.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands here
}