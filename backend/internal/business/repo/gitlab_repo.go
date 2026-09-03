package repo

import (
	"context"
	"errors"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/jackc/pgx/v5"
)

type GitLabRepo struct {
	db DBExecutor
}

func NewGitLabRepo(db DBExecutor) *GitLabRepo {
	return &GitLabRepo{db: db}
}

func (r *GitLabRepo) GetIntegration(ctx context.Context, projectID int64) (*dto.GitLabIntegration, error) {
	query := `
		SELECT project_id, base_url, gitlab_project_id, access_token_enc, webhook_secret, default_branch,
		       task_key_prefix, auto_move_in_progress, auto_move_review, auto_close_on_merge, is_active,
		       created_at, updated_at
		FROM business.gitlab_integrations
		WHERE project_id = $1
	`
	var g dto.GitLabIntegration
	err := r.db.QueryRow(ctx, query, projectID).Scan(
		&g.ProjectID, &g.BaseURL, &g.GitLabProjectID, &g.AccessTokenEnc, &g.WebhookSecret, &g.DefaultBranch,
		&g.TaskKeyPrefix, &g.AutoMoveInProgress, &g.AutoMoveReview, &g.AutoCloseOnMerge, &g.IsActive,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GitLabRepo) UpsertIntegration(ctx context.Context, g *dto.GitLabIntegration) error {
	query := `
		INSERT INTO business.gitlab_integrations (
			project_id, base_url, gitlab_project_id, access_token_enc, webhook_secret, default_branch,
			task_key_prefix, auto_move_in_progress, auto_move_review, auto_close_on_merge, is_active,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (project_id) DO UPDATE SET
			base_url = EXCLUDED.base_url,
			gitlab_project_id = EXCLUDED.gitlab_project_id,
			access_token_enc = COALESCE(EXCLUDED.access_token_enc, business.gitlab_integrations.access_token_enc),
			default_branch = EXCLUDED.default_branch,
			task_key_prefix = EXCLUDED.task_key_prefix,
			auto_move_in_progress = EXCLUDED.auto_move_in_progress,
			auto_move_review = EXCLUDED.auto_move_review,
			auto_close_on_merge = EXCLUDED.auto_close_on_merge,
			is_active = EXCLUDED.is_active,
			updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query,
		g.ProjectID, g.BaseURL, g.GitLabProjectID, g.AccessTokenEnc, g.WebhookSecret, g.DefaultBranch,
		g.TaskKeyPrefix, g.AutoMoveInProgress, g.AutoMoveReview, g.AutoCloseOnMerge, g.IsActive,
	)
	return err
}

func (r *GitLabRepo) DeleteIntegration(ctx context.Context, projectID int64) error {
	query := `DELETE FROM business.gitlab_integrations WHERE project_id = $1`
	_, err := r.db.Exec(ctx, query, projectID)
	return err
}

func (r *GitLabRepo) UpsertLink(ctx context.Context, l *dto.TaskGitLinkDTO) error {
	query := `
		INSERT INTO business.task_git_links (task_id, kind, external_id, title, state, web_url, author_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT (task_id, kind, external_id) DO UPDATE SET
			title = COALESCE(EXCLUDED.title, business.task_git_links.title),
			state = COALESCE(EXCLUDED.state, business.task_git_links.state),
			web_url = COALESCE(EXCLUDED.web_url, business.task_git_links.web_url),
			author_name = COALESCE(EXCLUDED.author_name, business.task_git_links.author_name),
			updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, l.TaskID, l.Kind, l.ExternalID, l.Title, l.State, l.WebURL, l.AuthorName)
	return err
}

func (r *GitLabRepo) ListLinks(ctx context.Context, taskID int64) ([]dto.TaskGitLinkDTO, error) {
	query := `
		SELECT id, task_id, kind, external_id, title, state, web_url, author_name, created_at, updated_at
		FROM business.task_git_links
		WHERE task_id = $1
		ORDER BY updated_at DESC
		LIMIT 50
	`
	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []dto.TaskGitLinkDTO
	for rows.Next() {
		var l dto.TaskGitLinkDTO
		if err := rows.Scan(
			&l.ID, &l.TaskID, &l.Kind, &l.ExternalID, &l.Title, &l.State, &l.WebURL, &l.AuthorName, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, nil
}

func (r *GitLabRepo) UpsertPipeline(ctx context.Context, p *dto.GitLabPipelineDTO) error {
	query := `
		INSERT INTO business.gitlab_pipelines (
			project_id, task_id, pipeline_id, ref, sha, status, duration_sec, web_url, started_at, finished_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (project_id, pipeline_id) DO UPDATE SET
			task_id = COALESCE(EXCLUDED.task_id, business.gitlab_pipelines.task_id),
			status = EXCLUDED.status,
			duration_sec = EXCLUDED.duration_sec,
			web_url = COALESCE(EXCLUDED.web_url, business.gitlab_pipelines.web_url),
			started_at = COALESCE(EXCLUDED.started_at, business.gitlab_pipelines.started_at),
			finished_at = EXCLUDED.finished_at,
			updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query,
		p.ProjectID, p.TaskID, p.PipelineID, p.Ref, p.SHA, p.Status, p.DurationSec, p.WebURL, p.StartedAt, p.FinishedAt,
	)
	return err
}

func (r *GitLabRepo) ListPipelines(ctx context.Context, taskID int64) ([]dto.GitLabPipelineDTO, error) {
	query := `
		SELECT id, project_id, task_id, pipeline_id, ref, sha, status, duration_sec, web_url, started_at, finished_at, updated_at
		FROM business.gitlab_pipelines
		WHERE task_id = $1
		ORDER BY updated_at DESC
		LIMIT 10
	`
	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pipelines []dto.GitLabPipelineDTO
	for rows.Next() {
		var p dto.GitLabPipelineDTO
		if err := rows.Scan(
			&p.ID, &p.ProjectID, &p.TaskID, &p.PipelineID, &p.Ref, &p.SHA, &p.Status,
			&p.DurationSec, &p.WebURL, &p.StartedAt, &p.FinishedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		pipelines = append(pipelines, p)
	}
	return pipelines, nil
}

func (r *GitLabRepo) ProjectSummary(ctx context.Context, projectID int64) ([]dto.ProjectGitSummaryItem, error) {
	query := `
		SELECT t.id,
		       COUNT(*) FILTER (WHERE l.kind = 'BRANCH') AS branches,
		       COUNT(*) FILTER (WHERE l.kind = 'MERGE_REQUEST') AS merge_requests,
		       p.status,
		       p.web_url
		FROM business.tasks t
		LEFT JOIN business.task_git_links l ON l.task_id = t.id
		LEFT JOIN LATERAL (
			SELECT status, web_url
			FROM business.gitlab_pipelines gp
			WHERE gp.task_id = t.id
			ORDER BY gp.updated_at DESC
			LIMIT 1
		) p ON TRUE
		WHERE t.project_id = $1
		GROUP BY t.id, p.status, p.web_url
		HAVING COUNT(l.id) > 0 OR p.status IS NOT NULL
	`
	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []dto.ProjectGitSummaryItem
	for rows.Next() {
		var item dto.ProjectGitSummaryItem
		if err := rows.Scan(&item.TaskID, &item.Branches, &item.MergeRequests, &item.PipelineStatus, &item.PipelineURL); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *GitLabRepo) InsertWebhookEvent(ctx context.Context, projectID int64, eventType, deliveryID string, payload []byte) (bool, error) {
	query := `
		INSERT INTO business.gitlab_webhook_events (project_id, event_type, delivery_id, payload, status, created_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, 'PENDING', NOW())
		ON CONFLICT (delivery_id) DO NOTHING
	`
	tag, err := r.db.Exec(ctx, query, projectID, eventType, deliveryID, payload)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *GitLabRepo) FetchPendingWebhookEvents(ctx context.Context, limit int) ([]dto.GitLabWebhookEvent, error) {
	query := `
		WITH picked AS (
			SELECT id
			FROM business.gitlab_webhook_events
			WHERE status = 'PENDING'
			ORDER BY created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE business.gitlab_webhook_events e
		SET status = 'PROCESSING'
		FROM picked
		WHERE e.id = picked.id
		RETURNING e.id::text, e.project_id, e.event_type, e.payload, e.retry_count, e.created_at
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []dto.GitLabWebhookEvent
	for rows.Next() {
		var e dto.GitLabWebhookEvent
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.EventType, &e.Payload, &e.RetryCount, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *GitLabRepo) RequeueStaleWebhookEvents(ctx context.Context, olderThan time.Duration) error {
	query := `
		UPDATE business.gitlab_webhook_events
		SET status = 'PENDING'
		WHERE status = 'PROCESSING' AND created_at < NOW() - make_interval(secs => $1)
	`
	_, err := r.db.Exec(ctx, query, olderThan.Seconds())
	return err
}

func (r *GitLabRepo) MarkWebhookProcessed(ctx context.Context, id string) error {
	query := `
		UPDATE business.gitlab_webhook_events
		SET status = 'PROCESSED', processed_at = NOW()
		WHERE id = $1::uuid
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *GitLabRepo) MarkWebhookFailed(ctx context.Context, id, errMsg string) error {
	query := `
		UPDATE business.gitlab_webhook_events
		SET status = CASE WHEN retry_count >= 3 THEN 'FAILED' ELSE 'PENDING' END,
		    retry_count = retry_count + 1,
		    error_message = $1
		WHERE id = $2::uuid
	`
	_, err := r.db.Exec(ctx, query, errMsg, id)
	return err
}
