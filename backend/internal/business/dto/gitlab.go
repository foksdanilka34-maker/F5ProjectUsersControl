package dto

import "time"

const (
	GitLinkBranch       = "BRANCH"
	GitLinkMergeRequest = "MERGE_REQUEST"
	GitLinkCommit       = "COMMIT"
)

type GitLabIntegration struct {
	ProjectID          int64     `json:"project_id"`
	BaseURL            string    `json:"base_url"`
	GitLabProjectID    int64     `json:"gitlab_project_id"`
	AccessTokenEnc     []byte    `json:"-"`
	WebhookSecret      string    `json:"-"`
	DefaultBranch      string    `json:"default_branch"`
	TaskKeyPrefix      string    `json:"task_key_prefix"`
	AutoMoveInProgress bool      `json:"auto_move_in_progress"`
	AutoMoveReview     bool      `json:"auto_move_review"`
	AutoCloseOnMerge   bool      `json:"auto_close_on_merge"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type GitLabIntegrationResponse struct {
	GitLabIntegration
	TokenSet      bool   `json:"token_set"`
	WebhookURL    string `json:"webhook_url"`
	WebhookSecret string `json:"webhook_secret"`
}

type SaveGitLabIntegrationRequest struct {
	BaseURL            string `json:"base_url"`
	GitLabProjectID    int64  `json:"gitlab_project_id"`
	AccessToken        string `json:"access_token,omitempty"`
	DefaultBranch      string `json:"default_branch"`
	TaskKeyPrefix      string `json:"task_key_prefix"`
	AutoMoveInProgress *bool  `json:"auto_move_in_progress,omitempty"`
	AutoMoveReview     *bool  `json:"auto_move_review,omitempty"`
	AutoCloseOnMerge   *bool  `json:"auto_close_on_merge,omitempty"`
	IsActive           *bool  `json:"is_active,omitempty"`
}

type TaskGitLinkDTO struct {
	ID         int64     `json:"id"`
	TaskID     int64     `json:"task_id"`
	Kind       string    `json:"kind"`
	ExternalID string    `json:"external_id"`
	Title      *string   `json:"title,omitempty"`
	State      *string   `json:"state,omitempty"`
	WebURL     *string   `json:"web_url,omitempty"`
	AuthorName *string   `json:"author_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type GitLabPipelineDTO struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id"`
	TaskID      *int64     `json:"task_id,omitempty"`
	PipelineID  int64      `json:"pipeline_id"`
	Ref         string     `json:"ref"`
	SHA         string     `json:"sha"`
	Status      string     `json:"status"`
	DurationSec *int       `json:"duration_sec,omitempty"`
	WebURL      *string    `json:"web_url,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TaskGitOverview struct {
	Connected       bool                `json:"connected"`
	TaskKey         string              `json:"task_key"`
	SuggestedBranch string              `json:"suggested_branch"`
	RepositoryURL   string              `json:"repository_url"`
	Links           []TaskGitLinkDTO    `json:"links"`
	Pipelines       []GitLabPipelineDTO `json:"pipelines"`
}

type ProjectGitSummaryItem struct {
	TaskID         int64   `json:"task_id"`
	Branches       int     `json:"branches"`
	MergeRequests  int     `json:"merge_requests"`
	PipelineStatus *string `json:"pipeline_status,omitempty"`
	PipelineURL    *string `json:"pipeline_url,omitempty"`
}

type GitLabWebhookEvent struct {
	ID         string    `json:"id"`
	ProjectID  int64     `json:"project_id"`
	EventType  string    `json:"event_type"`
	Payload    []byte    `json:"payload"`
	RetryCount int       `json:"retry_count"`
	CreatedAt  time.Time `json:"created_at"`
}
