package gitlab

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Client wraps the GitLab client with additional functionality
type Client struct {
	client *gitlab.Client
}

// MergeRequestRef represents a merge request reference
type MergeRequestRef struct {
	ID      int
	IID     int
	HeadSHA string
}

// NewClient creates a new GitLab client
func NewClient(token, baseURL string) (*Client, error) {
	var client *gitlab.Client
	var err error

	if baseURL != "" {
		client, err = gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	} else {
		client, err = gitlab.NewClient(token)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &Client{client: client}, nil
}

// ParseRepoPath parses various GitLab repository path formats
func ParseRepoPath(repoPath string) (string, string, error) {
	// Handle full URLs
	if strings.HasPrefix(repoPath, "http") {
		u, err := url.Parse(repoPath)
		if err != nil {
			return "", "", fmt.Errorf("invalid URL: %w", err)
		}
		
		baseURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		path := strings.TrimPrefix(u.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		
		return baseURL, path, nil
	}

	// Handle group/repo or group/subgroup/repo formats
	// Validate the format
	if !regexp.MustCompile(`^[a-zA-Z0-9._-]+(/[a-zA-Z0-9._-]+)+$`).MatchString(repoPath) {
		return "", "", fmt.Errorf("invalid repository path format: %s", repoPath)
	}

	return "", repoPath, nil
}

// FetchMergeRequestRefs fetches all merge request references for a given repository
func (c *Client) FetchMergeRequestRefs(projectPath string) ([]MergeRequestRef, error) {
	// List all merge requests for the project
	opts := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100, // GitLab API max per page
		},
		State: gitlab.Ptr("all"), // Get both open and closed MRs
	}

	var allMRs []MergeRequestRef

	for {
		mrs, resp, err := c.client.MergeRequests.ListProjectMergeRequests(projectPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch merge requests: %w", err)
		}

		for _, mr := range mrs {
			// Fetch detailed merge request to get diff_refs
			detailedMR, _, err := c.client.MergeRequests.GetMergeRequest(projectPath, mr.IID, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch merge request %d: %w", mr.IID, err)
			}

			if detailedMR.DiffRefs.HeadSha != "" {
				allMRs = append(allMRs, MergeRequestRef{
					ID:      mr.ID,
					IID:     mr.IID,
					HeadSHA: detailedMR.DiffRefs.HeadSha,
				})
			}
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allMRs, nil
}