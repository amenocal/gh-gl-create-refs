package gitlab

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Client wraps the GitLab client with additional functionality
type Client struct {
	client          *gitlab.Client
	lastRequestTime time.Time
	minInterval     time.Duration
}

// MergeRequestRef represents a merge request reference
type MergeRequestRef struct {
	ID      int
	IID     int
	HeadSHA string
}

// MergeRequestProcessor is a callback function that processes each merge request as it's fetched
type MergeRequestProcessor func(MergeRequestRef) error

// NewClient creates a new GitLab client
func NewClient(token, baseURL string) (*Client, error) {
	var client *gitlab.Client
	var err error

	if baseURL != "" {
		log.Println("Using custom GitLab base URL:", baseURL)
		client, err = gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
	} else {
		client, err = gitlab.NewClient(token)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &Client{
		client:      client,
		minInterval: 100 * time.Millisecond, // Conservative rate limit: max 10 requests/second
	}, nil
}

// rateLimitWait ensures we don't exceed rate limits by waiting if necessary
func (c *Client) rateLimitWait() {
	now := time.Now()
	if !c.lastRequestTime.IsZero() {
		elapsed := now.Sub(c.lastRequestTime)
		if elapsed < c.minInterval {
			sleepDuration := c.minInterval - elapsed
			fmt.Printf("⏳ Respecting GitLab API rate limits, waiting %v before next request...\n", sleepDuration.Round(time.Millisecond))
			time.Sleep(sleepDuration)
		}
	}
	c.lastRequestTime = time.Now()
}

// checkRateLimitHeaders examines GitLab's rate limit headers and adjusts behavior accordingly
func (c *Client) checkRateLimitHeaders(resp *http.Response) {
	if resp == nil {
		return
	}

	// GitLab.com rate limit headers
	rateLimitRemaining := resp.Header.Get("RateLimit-Remaining")
	rateLimitReset := resp.Header.Get("RateLimit-ResetTime")

	// Alternative headers that might be present
	if rateLimitRemaining == "" {
		rateLimitRemaining = resp.Header.Get("X-RateLimit-Remaining")
	}
	if rateLimitReset == "" {
		rateLimitReset = resp.Header.Get("X-RateLimit-Reset")
	}

	if rateLimitRemaining != "" {
		if remaining, err := strconv.Atoi(rateLimitRemaining); err == nil {
			if remaining <= 10 { // If we're getting close to the limit
				fmt.Printf("⚠️  Rate limit warning: only %d requests remaining, slowing down requests\n", remaining)
				// Increase our conservative interval
				c.minInterval = 1 * time.Second
			} else if remaining <= 5 {
				fmt.Printf("🚨 Rate limit critical: only %d requests remaining, significantly slowing down\n", remaining)
				c.minInterval = 5 * time.Second
			}
		}
	}

	// Check if we hit the rate limit (status 429)
	if resp.StatusCode == 429 {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				sleepDuration := time.Duration(seconds) * time.Second
				fmt.Printf("🛑 GitLab API rate limit exceeded! Waiting %v as requested by server...\n", sleepDuration)
				fmt.Printf("   This is normal and helps ensure fair API usage. Please wait...\n")
				time.Sleep(sleepDuration)
				return
			}
		}
		// Fallback if no Retry-After header
		fmt.Printf("🛑 GitLab API rate limit exceeded! Waiting 60 seconds to retry...\n")
		fmt.Printf("   This is normal and helps ensure fair API usage. Please wait...\n")
		time.Sleep(60 * time.Second)
	}
}

// ParseRepoPath parses various GitLab repository path formats
// returns the base URL (if any) and the project path (example: group/subgroup/repo)
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

// FetchMergeRequestRefs fetches all merge request references for a given repository and processes them via callback
func (c *Client) FetchMergeRequestRefs(projectPath string, processor MergeRequestProcessor) error {
	// List all merge requests for the project
	opts := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100, // GitLab API max per page
		},
		State: gitlab.Ptr("all"), // Get both open and closed MRs
	}

	pageCount := 0

	for {
		pageCount++
		// Apply rate limiting before making the list request
		c.rateLimitWait()

		mrs, resp, err := c.client.MergeRequests.ListProjectMergeRequests(projectPath, opts)
		if err != nil {
			return fmt.Errorf("failed to fetch merge requests: %w", err)
		}

		fmt.Printf("📋 Processing page %d: found %d merge requests...\n", pageCount, len(mrs))

		// Check rate limit headers from the response
		c.checkRateLimitHeaders(resp.Response)

		for _, mr := range mrs {
			// Apply rate limiting before each detailed request
			c.rateLimitWait()

			// Fetch detailed merge request to get diff_refs
			detailedMR, detailResp, err := c.client.MergeRequests.GetMergeRequest(projectPath, mr.IID, nil)
			if err != nil {
				return fmt.Errorf("failed to fetch merge request %d: %w", mr.IID, err)
			}

			// Check rate limit headers from the detailed request response
			c.checkRateLimitHeaders(detailResp.Response)

			if detailedMR.DiffRefs.HeadSha != "" {
				ref := MergeRequestRef{
					ID:      mr.ID,
					IID:     mr.IID,
					HeadSHA: detailedMR.DiffRefs.HeadSha,
				}

				// Process the merge request via callback
				if err := processor(ref); err != nil {
					return fmt.Errorf("failed to process merge request %d: %w", mr.IID, err)
				}
			}
		}

		// Check if there are more pages
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

// FetchMergeRequestRefsFromRepo processes merge request references using a callback
func (c *Client) FetchMergeRequestRefsFromRepo(repoPath string, baseURLOverride string, processor MergeRequestProcessor) (string, error) {
	// Parse repository path and determine base URL
	baseURL, projectPath, err := ParseRepoPath(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse repository path: %w", err)
	}

	// Note: We don't use baseURL from parsing if baseURLOverride is provided
	// This is handled during client creation in NewClientFromFlags
	_ = baseURL

	// Fetch merge request references using callback
	err = c.FetchMergeRequestRefs(projectPath, processor)
	if err != nil {
		return "", c.wrapFetchError(err, projectPath)
	}

	return projectPath, nil
}

// wrapFetchError provides more helpful error messages for common GitLab API issues
func (c *Client) wrapFetchError(err error, projectPath string) error {
	errMsg := err.Error()

	if strings.Contains(errMsg, "404") {
		return fmt.Errorf("repository not found: %s. Please check the repository path and your access permissions", projectPath)
	}

	if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") {
		return fmt.Errorf("authentication failed: please check your GitLab token has access to repository %s", projectPath)
	}

	return fmt.Errorf("failed to fetch merge request references from %s: %w", projectPath, err)
}
