package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://gitlab.com"

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type Project struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
	DefaultBranch     string `json:"default_branch"`
}

type Branch struct {
	Name   string `json:"name"`
	WebURL string `json:"web_url"`
	Commit struct {
		ID string `json:"id"`
	} `json:"commit"`
}

type MergeRequest struct {
	IID          int64  `json:"iid"`
	Title        string `json:"title"`
	State        string `json:"state"`
	WebURL       string `json:"web_url"`
	SourceBranch string `json:"source_branch"`
	Author       struct {
		Name string `json:"name"`
	} `json:"author"`
}

type Pipeline struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
	Ref    string `json:"ref"`
	SHA    string `json:"sha"`
	WebURL string `json:"web_url"`
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("gitlab api error: status %d: %s", e.StatusCode, e.Body)
}

func (c *Client) GetProject(ctx context.Context, projectID int64) (*Project, error) {
	var p Project
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("/projects/%d", projectID), nil, &p)
	return &p, err
}

func (c *Client) CreateBranch(ctx context.Context, projectID int64, branch, ref string) (*Branch, error) {
	path := fmt.Sprintf("/projects/%d/repository/branches?branch=%s&ref=%s",
		projectID, url.QueryEscape(branch), url.QueryEscape(ref))

	var b Branch
	err := c.do(ctx, http.MethodPost, path, nil, &b)
	return &b, err
}

func (c *Client) ListMergeRequests(ctx context.Context, projectID int64, sourceBranch string) ([]MergeRequest, error) {
	path := fmt.Sprintf("/projects/%d/merge_requests?source_branch=%s&state=all",
		projectID, url.QueryEscape(sourceBranch))

	var mrs []MergeRequest
	err := c.do(ctx, http.MethodGet, path, nil, &mrs)
	return mrs, err
}

func (c *Client) CreateMergeRequestNote(ctx context.Context, projectID, mergeRequestIID int64, body string) error {
	path := fmt.Sprintf("/projects/%d/merge_requests/%d/notes", projectID, mergeRequestIID)
	return c.do(ctx, http.MethodPost, path, map[string]string{"body": body}, nil)
}

func (c *Client) RetryPipeline(ctx context.Context, projectID, pipelineID int64) (*Pipeline, error) {
	path := fmt.Sprintf("/projects/%d/pipelines/%d/retry", projectID, pipelineID)

	var p Pipeline
	err := c.do(ctx, http.MethodPost, path, nil, &p)
	return &p, err
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+"/api/v4"+path, reader)
	if err != nil {
		return err
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(payload))}
	}

	if out == nil {
		return nil
	}
	return json.Unmarshal(payload, out)
}
