package repo

import (
	"context"
	"fmt"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/dto"
)

type ProjectRepo struct {
	db DBExecutor
}

func NewProjectRepo(db DBExecutor) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(ctx context.Context, p *dto.ProjectDTO) (int64, error) {
	query := `
		INSERT INTO business.projects (name, description, owner_id, status, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, p.Name, p.Description, p.ManagerID, p.Status, p.DueDate).Scan(&id)
	return id, err
}

func (r *ProjectRepo) GetByID(ctx context.Context, id int64) (*dto.ProjectDTO, error) {
	query := `
		SELECT id, name, description, owner_id, status, start_date, end_date, created_at, updated_at
		FROM business.projects
		WHERE id = $1
	`
	var p dto.ProjectDTO
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.ManagerID, &p.Status, &p.StartDate, &p.DueDate, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepo) List(ctx context.Context, filter dto.ListProjectsFilter) ([]dto.ProjectDTO, int, error) {
	offset := (filter.PageNumber - 1) * filter.PageSize

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.ManagerID > 0 {
		whereClause += fmt.Sprintf(" AND p.owner_id = $%d", argIdx)
		args = append(args, filter.ManagerID)
		argIdx++
	}
	if filter.MemberID > 0 {
		whereClause += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM business.project_members pm WHERE pm.project_id = p.id AND pm.user_id = $%d)", argIdx)
		args = append(args, filter.MemberID)
		argIdx++
	}
	if filter.Status != "" {
		whereClause += fmt.Sprintf(" AND p.status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM business.projects p %s", whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.owner_id, p.status, p.start_date, p.end_date, p.created_at, p.updated_at
		FROM business.projects p
		%s
		ORDER BY p.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var projects []dto.ProjectDTO
	for rows.Next() {
		var p dto.ProjectDTO
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.ManagerID, &p.Status, &p.StartDate, &p.DueDate, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		projects = append(projects, p)
	}

	return projects, total, nil
}

func (r *ProjectRepo) Update(ctx context.Context, id int64, req dto.UpdateProjectRequest) error {
	query := `
		UPDATE business.projects
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description),
		    status = COALESCE($3, status),
		    owner_id = COALESCE($4, owner_id),
		    end_date = COALESCE($5, end_date),
		    updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query, req.Name, req.Description, req.Status, req.ManagerID, req.DueDate, id)
	return err
}

func (r *ProjectRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM business.projects WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *ProjectRepo) AddMember(ctx context.Context, projectID, userID int64, role string) error {
	query := `
		INSERT INTO business.project_members (project_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (project_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`
	_, err := r.db.Exec(ctx, query, projectID, userID, role)
	return err
}

func (r *ProjectRepo) RemoveMember(ctx context.Context, projectID, userID int64) error {
	query := `DELETE FROM business.project_members WHERE project_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, projectID, userID)
	return err
}

func (r *ProjectRepo) GetMembers(ctx context.Context, projectID int64) ([]dto.ProjectMemberDTO, error) {
	query := `
		SELECT pm.user_id, COALESCE(um.full_name, 'Сотрудник #' || pm.user_id) as full_name, pm.role, um.photo_url
		FROM business.project_members pm
		LEFT JOIN business.users_meta um ON pm.user_id = um.user_id
		WHERE pm.project_id = $1
		ORDER BY pm.joined_at ASC
	`
	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []dto.ProjectMemberDTO
	for rows.Next() {
		var m dto.ProjectMemberDTO
		if err := rows.Scan(&m.UserID, &m.FullName, &m.Role, &m.PhotoURL); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}
