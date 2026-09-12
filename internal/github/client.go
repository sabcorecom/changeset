// Package github is a minimal hand-rolled REST client for the GitHub
// Pulls API, using only the stdlib net/http and encoding/json — no
// third-party GitHub SDK, consistent with the rest of this codebase.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	defaultBaseURL = "https://api.github.com"
	apiVersion     = "2022-11-28"
)

// PullRequest is the subset of the GitHub pull request resource this
// package needs.
type PullRequest struct {
	Number  int    `json:"number"`
	HTMLURL string `json:"html_url"`
}

// Client is a minimal GitHub REST API client scoped to one repository.
type Client struct {
	token      string
	owner      string
	repo       string
	baseURL    string
	httpClient *http.Client
}

// NewClient builds a Client for owner/repo, authenticating with token
// against baseURL (pass defaultBaseURL's value, "https://api.github.com",
// for production use; tests can point this at an httptest.Server).
func NewClient(token, owner, repo, baseURL string) *Client {
	return &Client{
		token:      token,
		owner:      owner,
		repo:       repo,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: http.DefaultClient,
	}
}

// NewClientFromEnv builds a Client from the standard GitHub Actions
// environment: GITHUB_TOKEN for auth and GITHUB_REPOSITORY (format
// "owner/repo") for the target repository. Both are required; an
// unset or malformed value is a fail-fast error rather than a silent
// unauthenticated or misdirected client. GITHUB_API_URL, another
// GitHub Actions default env var (used for GitHub Enterprise Server),
// overrides the API base URL when set.
func NewClientFromEnv() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN is not set")
	}

	repository := os.Getenv("GITHUB_REPOSITORY")
	owner, repo, ok := strings.Cut(repository, "/")
	if !ok || owner == "" || repo == "" {
		return nil, fmt.Errorf("GITHUB_REPOSITORY must be in the form owner/repo, got %q", repository)
	}

	baseURL := os.Getenv("GITHUB_API_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	return NewClient(token, owner, repo, baseURL), nil
}

func (c *Client) do(method, path string, body any, out any) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("build request %s %s: %w", method, path, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", apiVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body for %s %s: %w", method, path, err)
	}

	if resp.StatusCode >= 300 {
		return resp, fmt.Errorf("%s %s: unexpected status %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return resp, fmt.Errorf("decode response body for %s %s: %w", method, path, err)
		}
	}

	return resp, nil
}

// FindOpenPullRequest returns the single open pull request from head to
// base, or nil if none exists. It errors out rather than silently picking
// one if the API reports more than one match — that would mean the
// "never more than one open Version PR" invariant has already been
// violated and needs surfacing, not papering over.
func (c *Client) FindOpenPullRequest(head, base string) (*PullRequest, error) {
	query := url.Values{
		"head":  {c.owner + ":" + head},
		"base":  {base},
		"state": {"open"},
	}
	path := fmt.Sprintf("/repos/%s/%s/pulls?%s", c.owner, c.repo, query.Encode())

	var prs []PullRequest
	if _, err := c.do(http.MethodGet, path, nil, &prs); err != nil {
		return nil, err
	}

	switch len(prs) {
	case 0:
		return nil, nil
	case 1:
		return &prs[0], nil
	default:
		return nil, fmt.Errorf("found %d open pull requests for head %s, base %s: expected at most one", len(prs), head, base)
	}
}

// CreatePullRequest opens a new pull request from head to base.
func (c *Client) CreatePullRequest(head, base, title, body string) (*PullRequest, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls", c.owner, c.repo)

	reqBody := map[string]string{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	}

	var pr PullRequest
	if _, err := c.do(http.MethodPost, path, reqBody, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// UpdatePullRequest updates the title and body of the pull request
// identified by number.
func (c *Client) UpdatePullRequest(number int, title, body string) error {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", c.owner, c.repo, number)

	reqBody := map[string]string{
		"title": title,
		"body":  body,
	}

	_, err := c.do(http.MethodPatch, path, reqBody, nil)
	return err
}
