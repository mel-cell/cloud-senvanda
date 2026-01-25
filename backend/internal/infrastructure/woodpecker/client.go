package woodpecker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
}

// NewClient initializes a new Woodpecker API client
func NewClient(baseURL, apiToken string) *Client {
	return &Client{
		baseURL:  baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LookupRepoID finds the internal Woodpecker ID for a given repo slug (owner/repo)
func (c *Client) LookupRepoID(owner, repoName string) (int64, error) {
	// fetching all user repos is safer for JSON response in v2
	url := fmt.Sprintf("%s/api/user/repos", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("user repos lookup failed: %s - %s", resp.Status, string(body))
	}

	var repos []struct {
		ID       int64  `json:"id"`
		Owner    string `json:"owner"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	}
	if err := json.Unmarshal(body, &repos); err != nil {
		return 0, fmt.Errorf("failed to parse repos list: %w", err)
	}

	target := fmt.Sprintf("%s/%s", owner, repoName)
	for _, r := range repos {
		if r.FullName == target || (r.Owner == owner && r.Name == repoName) {
			return r.ID, nil
		}
	}

	return 0, fmt.Errorf("repository %s not found in Woodpecker", target)
}

// TriggerBuild triggers a new pipeline build for the specified repo ID
// It sends POST to /api/repos/{repo_id}/pipelines
func (c *Client) TriggerBuild(repoID int64, branch string) (int64, error) {
	url := fmt.Sprintf("%s/api/repos/%d/pipelines", c.baseURL, repoID)

	// Woodpecker v2 preferred trigger is via POST to /pipelines
	// Often requires the branch in the body
	payload := map[string]string{
		"branch": branch,
	}
	bodyBytes, _ := json.Marshal(payload)

	// Construct request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to call woodpecker api: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	// log.Printf("[Woodpecker Debug] URL: %s, Status: %s, Body Snippet: %s", url, resp.Status, string(respBody))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("woodpecker api error: %s - %s", resp.Status, string(respBody))
	}

	// Woodpecker v2 might return empty body on success trigger
	if len(respBody) == 0 || string(respBody) == "" {
		log.Printf("⚠️ Woodpecker returned empty body, but status is %s. Assuming success.", resp.Status)
		return 0, nil
	}

	// Parse Response to get Pipeline Number
	var pipelineResp struct {
		Number int64 `json:"number"`
	}
	if err := json.Unmarshal(respBody, &pipelineResp); err != nil {
		// If fails to parse but body not empty, it's a real error
		return 0, fmt.Errorf("failed to parse pipeline response: %w (body: %s)", err, string(respBody))
	}

	return pipelineResp.Number, nil
}
