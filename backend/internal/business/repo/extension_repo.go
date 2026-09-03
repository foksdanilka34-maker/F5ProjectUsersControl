package repo

import (
	"context"
	"errors"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
	"github.com/jackc/pgx/v5"
)

type ExtensionRepo struct {
	db DBExecutor
}

func NewExtensionRepo(db DBExecutor) *ExtensionRepo {
	return &ExtensionRepo{db: db}
}

func (r *ExtensionRepo) Create(ctx context.Context, e *dto.ExtensionDTO) (int64, error) {
	query := `
		INSERT INTO business.extensions (
			key, name, description, base_url, shared_secret_enc, task_panel_url,
			project_tab_url, project_tab_label, events, is_active, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, TRUE, NOW(), NOW())
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query,
		e.Key, e.Name, e.Description, e.BaseURL, e.SharedSecretEnc,
		e.TaskPanelURL, e.ProjectTabURL, e.ProjectTabLabel, e.Events,
	).Scan(&id)
	return id, err
}

func (r *ExtensionRepo) GetByKey(ctx context.Context, key string) (*dto.ExtensionDTO, error) {
	query := `
		SELECT id, key, name, description, base_url, shared_secret_enc, task_panel_url,
		       project_tab_url, project_tab_label, events, is_active, created_at, updated_at
		FROM business.extensions
		WHERE key = $1
	`
	return r.scanOne(r.db.QueryRow(ctx, query, key))
}

func (r *ExtensionRepo) GetByID(ctx context.Context, id int64) (*dto.ExtensionDTO, error) {
	query := `
		SELECT id, key, name, description, base_url, shared_secret_enc, task_panel_url,
		       project_tab_url, project_tab_label, events, is_active, created_at, updated_at
		FROM business.extensions
		WHERE id = $1
	`
	return r.scanOne(r.db.QueryRow(ctx, query, id))
}

func (r *ExtensionRepo) scanOne(row pgx.Row) (*dto.ExtensionDTO, error) {
	var e dto.ExtensionDTO
	err := row.Scan(
		&e.ID, &e.Key, &e.Name, &e.Description, &e.BaseURL, &e.SharedSecretEnc, &e.TaskPanelURL,
		&e.ProjectTabURL, &e.ProjectTabLabel, &e.Events, &e.IsActive, &e.CreatedAt, &e.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *ExtensionRepo) List(ctx context.Context) ([]dto.ExtensionDTO, error) {
	query := `
		SELECT id, key, name, description, base_url, shared_secret_enc, task_panel_url,
		       project_tab_url, project_tab_label, events, is_active, created_at, updated_at
		FROM business.extensions
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var extensions []dto.ExtensionDTO
	for rows.Next() {
		var e dto.ExtensionDTO
		if err := rows.Scan(
			&e.ID, &e.Key, &e.Name, &e.Description, &e.BaseURL, &e.SharedSecretEnc, &e.TaskPanelURL,
			&e.ProjectTabURL, &e.ProjectTabLabel, &e.Events, &e.IsActive, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		extensions = append(extensions, e)
	}
	return extensions, nil
}

func (r *ExtensionRepo) Delete(ctx context.Context, key string) error {
	query := `DELETE FROM business.extensions WHERE key = $1`
	_, err := r.db.Exec(ctx, query, key)
	return err
}

func (r *ExtensionRepo) SetProjectInstall(ctx context.Context, projectID, extensionID int64, enabled bool) error {
	query := `
		INSERT INTO business.project_extensions (project_id, extension_id, is_enabled, installed_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (project_id, extension_id) DO UPDATE SET is_enabled = EXCLUDED.is_enabled
	`
	_, err := r.db.Exec(ctx, query, projectID, extensionID, enabled)
	return err
}

func (r *ExtensionRepo) Uninstall(ctx context.Context, projectID, extensionID int64) error {
	query := `DELETE FROM business.project_extensions WHERE project_id = $1 AND extension_id = $2`
	_, err := r.db.Exec(ctx, query, projectID, extensionID)
	return err
}

func (r *ExtensionRepo) ListForProject(ctx context.Context, projectID int64) ([]dto.ProjectExtensionDTO, error) {
	query := `
		SELECT e.id, e.key, e.name, e.description, e.base_url, e.task_panel_url,
		       e.project_tab_url, e.project_tab_label, e.events, e.is_active, e.created_at, e.updated_at,
		       pe.project_id IS NOT NULL AS installed,
		       COALESCE(pe.is_enabled, FALSE) AS enabled
		FROM business.extensions e
		LEFT JOIN business.project_extensions pe ON pe.extension_id = e.id AND pe.project_id = $1
		WHERE e.is_active = TRUE
		ORDER BY e.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var extensions []dto.ProjectExtensionDTO
	for rows.Next() {
		var e dto.ProjectExtensionDTO
		if err := rows.Scan(
			&e.ID, &e.Key, &e.Name, &e.Description, &e.BaseURL, &e.TaskPanelURL,
			&e.ProjectTabURL, &e.ProjectTabLabel, &e.Events, &e.IsActive, &e.CreatedAt, &e.UpdatedAt,
			&e.Installed, &e.Enabled,
		); err != nil {
			return nil, err
		}
		extensions = append(extensions, e)
	}
	return extensions, nil
}

func (r *ExtensionRepo) ListEnabledForProjectByEvent(ctx context.Context, projectID int64, event string) ([]dto.ExtensionDTO, error) {
	query := `
		SELECT e.id, e.key, e.name, e.description, e.base_url, e.shared_secret_enc, e.task_panel_url,
		       e.project_tab_url, e.project_tab_label, e.events, e.is_active, e.created_at, e.updated_at
		FROM business.extensions e
		JOIN business.project_extensions pe ON pe.extension_id = e.id
		WHERE pe.project_id = $1 AND pe.is_enabled = TRUE AND e.is_active = TRUE AND $2 = ANY(e.events)
	`
	rows, err := r.db.Query(ctx, query, projectID, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var extensions []dto.ExtensionDTO
	for rows.Next() {
		var e dto.ExtensionDTO
		if err := rows.Scan(
			&e.ID, &e.Key, &e.Name, &e.Description, &e.BaseURL, &e.SharedSecretEnc, &e.TaskPanelURL,
			&e.ProjectTabURL, &e.ProjectTabLabel, &e.Events, &e.IsActive, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		extensions = append(extensions, e)
	}
	return extensions, nil
}

func (r *ExtensionRepo) IsInstalledAndEnabled(ctx context.Context, projectID, extensionID int64) (bool, error) {
	query := `
		SELECT is_enabled FROM business.project_extensions
		WHERE project_id = $1 AND extension_id = $2
	`
	var enabled bool
	err := r.db.QueryRow(ctx, query, projectID, extensionID).Scan(&enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return enabled, nil
}

func (r *ExtensionRepo) SetProperty(ctx context.Context, taskID, extensionID int64, key string, value []byte) error {
	query := `
		INSERT INTO business.task_entity_properties (task_id, extension_id, property_key, value, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (task_id, extension_id, property_key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, taskID, extensionID, key, value)
	return err
}

func (r *ExtensionRepo) GetProperty(ctx context.Context, taskID, extensionID int64, key string) ([]byte, error) {
	query := `
		SELECT value FROM business.task_entity_properties
		WHERE task_id = $1 AND extension_id = $2 AND property_key = $3
	`
	var value []byte
	err := r.db.QueryRow(ctx, query, taskID, extensionID, key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (r *ExtensionRepo) ListProperties(ctx context.Context, taskID, extensionID int64) ([]dto.TaskEntityPropertyDTO, error) {
	query := `
		SELECT task_id, extension_id, property_key, value, updated_at
		FROM business.task_entity_properties
		WHERE task_id = $1 AND extension_id = $2
		ORDER BY property_key ASC
	`
	rows, err := r.db.Query(ctx, query, taskID, extensionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var properties []dto.TaskEntityPropertyDTO
	for rows.Next() {
		var p dto.TaskEntityPropertyDTO
		if err := rows.Scan(&p.TaskID, &p.ExtensionID, &p.Key, &p.Value, &p.UpdatedAt); err != nil {
			return nil, err
		}
		properties = append(properties, p)
	}
	return properties, nil
}
