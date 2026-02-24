package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/TomasBack2Future/Kinetik/automation/pkg/config"
	"github.com/TomasBack2Future/Kinetik/automation/pkg/logger"
)

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		token:      cfg.GitHub.PersonalAccessToken,
		httpClient: &http.Client{},
	}
}

// CreateIssueComment posts a comment on an issue
func (c *Client) CreateIssueComment(owner, repo string, issueNumber int, body string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, issueNumber)

	payload := map[string]string{
		"body": body,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create comment: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	logger.WithField("issue_number", issueNumber).Info("Posted comment to issue")
	return nil
}

// CreateBranch creates a new branch from base branch
func (c *Client) CreateBranch(owner, repo, branchName, baseBranch string) error {
	// Get base branch SHA
	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/heads/%s", owner, repo, baseBranch)

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get base branch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to get base branch: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var refData struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&refData); err != nil {
		return fmt.Errorf("failed to decode base branch response: %w", err)
	}

	// Create new branch
	createURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", owner, repo)

	payload := map[string]string{
		"ref": "refs/heads/" + branchName,
		"sha": refData.Object.SHA,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal branch payload: %w", err)
	}

	req, err = http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create branch request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create branch: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	logger.WithField("branch_name", branchName).Info("Created branch")
	return nil
}

// AddIssueLabel adds a label to an issue
func (c *Client) AddIssueLabel(owner, repo string, issueNumber int, label string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/labels", owner, repo, issueNumber)

	payload := map[string][]string{
		"labels": {label},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal label: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add label: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add label: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	logger.WithField("label", label).Info("Added label to issue")
	return nil
}

// ParseRepoOwner splits "owner/repo" into owner and repo
func ParseRepoOwner(fullName string) (owner, repo string) {
	parts := strings.Split(fullName, "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}
