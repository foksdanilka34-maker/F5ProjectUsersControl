package repo

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgreSQL placeholder format
var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type ProjectRepo struct {
	db *pgxpool.Pool
}

func NewProjectRepo(db *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(ctx context.Context, p *Project) (int64, error) {
	var id int64
	query := `
		INSERT INTO projects (name, description, status, start_date, end_date, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err := r.db.QueryRow(ctx, query, p.Name, p.Description, p.Status, p.StartDate, p.EndDate, p.OwnerID, p.CreatedAt, p.UpdatedAt).Scan(&id)
	return id, err
}

func (r *ProjectRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM projects WHERE LOWER(name) = LOWER($1))`
	err := r.db.QueryRow(ctx, query, name).Scan(&exists)
	return exists, err
}

func (r *ProjectRepo) GetByID(ctx context.Context, id int64) (*Project, error) {
	query := `
		SELECT p.id, p.name, p.description, p.status, p.start_date, p.end_date, p.owner_id, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM tasks WHERE project_id = p.id) as task_count,
			(SELECT COUNT(*) FROM project_members WHERE project_id = p.id) as member_count
		FROM projects p
		WHERE p.id = $1
	`
	var p Project
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Status, &p.StartDate, &p.EndDate, &p.OwnerID, &p.CreatedAt, &p.UpdatedAt,
		&p.TaskCount, &p.MemberCount,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepo) List(ctx context.Context, pageSize, offset int, status string, ownerID int64, memberID int64) ([]*Project, int, error) {
	// Count query
	countBuilder := psql.Select("COUNT(DISTINCT p.id)").From("projects p")
	if memberID > 0 {
		countBuilder = countBuilder.LeftJoin("project_members pm ON pm.project_id = p.id").
			Where(sq.Or{sq.Eq{"p.owner_id": memberID}, sq.Eq{"pm.user_id": memberID}})
	}
	if status != "" {
		countBuilder = countBuilder.Where(sq.Eq{"p.status": status})
	}
	if ownerID > 0 {
		countBuilder = countBuilder.Where(sq.Eq{"p.owner_id": ownerID})
	}

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Select query
	selectBuilder := psql.Select(
		"DISTINCT p.id", "p.name", "p.description", "p.status", "p.start_date", "p.end_date",
		"p.owner_id", "p.created_at", "p.updated_at",
		"(SELECT COUNT(*) FROM tasks WHERE project_id = p.id) as task_count",
		"(SELECT COUNT(*) FROM project_members WHERE project_id = p.id) as member_count",
	).From("projects p").
		OrderBy("p.created_at DESC").
		Limit(uint64(pageSize)).
		Offset(uint64(offset))

	if memberID > 0 {
		selectBuilder = selectBuilder.LeftJoin("project_members pm ON pm.project_id = p.id").
			Where(sq.Or{sq.Eq{"p.owner_id": memberID}, sq.Eq{"pm.user_id": memberID}})
	}
	if status != "" {
		selectBuilder = selectBuilder.Where(sq.Eq{"p.status": status})
	}
	if ownerID > 0 {
		selectBuilder = selectBuilder.Where(sq.Eq{"p.owner_id": ownerID})
	}

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Status, &p.StartDate, &p.EndDate,
			&p.OwnerID, &p.CreatedAt, &p.UpdatedAt,
			&p.TaskCount, &p.MemberCount,
		); err != nil {
			return nil, 0, err
		}
		projects = append(projects, &p)
	}

	return projects, total, nil
}

func (r *ProjectRepo) Update(ctx context.Context, id int64, name, description, status *string, startDate, endDate *time.Time) error {
	builder := psql.Update("projects").Where(sq.Eq{"id": id})

	if name != nil {
		builder = builder.Set("name", *name)
	}
	if description != nil {
		builder = builder.Set("description", *description)
	}
	if status != nil {
		builder = builder.Set("status", *status)
	}
	if startDate != nil {
		builder = builder.Set("start_date", *startDate)
	}
	if endDate != nil {
		builder = builder.Set("end_date", *endDate)
	}

	builder = builder.Set("updated_at", time.Now())

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	return err
}

func (r *ProjectRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM projects WHERE id = $1", id)
	return err
}

func (r *ProjectRepo) GetUserProjects(ctx context.Context, userID int64) ([]*Project, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.status, p.start_date, p.end_date, p.owner_id, p.created_at, p.updated_at,
			(SELECT COUNT(*) FROM tasks WHERE project_id = p.id) as task_count,
			(SELECT COUNT(*) FROM project_members WHERE project_id = p.id) as member_count
		FROM projects p
		LEFT JOIN project_members pm ON p.id = pm.project_id
		WHERE p.owner_id = $1 OR pm.user_id = $1
		ORDER BY p.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Status, &p.StartDate, &p.EndDate, &p.OwnerID, &p.CreatedAt, &p.UpdatedAt,
			&p.TaskCount, &p.MemberCount,
		); err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}
	return projects, nil
}

// Project Members
func (r *ProjectRepo) AddMember(ctx context.Context, projectID, userID int64, role string) error {
	query := `
		INSERT INTO project_members (project_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (project_id, user_id) DO UPDATE SET role = $3
	`
	_, err := r.db.Exec(ctx, query, projectID, userID, role, time.Now())
	return err
}

func (r *ProjectRepo) RemoveMember(ctx context.Context, projectID, userID int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM project_members WHERE project_id = $1 AND user_id = $2", projectID, userID)
	return err
}

func (r *ProjectRepo) GetMembers(ctx context.Context, projectID int64) ([]*ProjectMember, error) {
	query := `SELECT project_id, user_id, role, joined_at FROM project_members WHERE project_id = $1`
	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*ProjectMember
	for rows.Next() {
		var m ProjectMember
		if err := rows.Scan(&m.ProjectID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}
	return members, nil
}

func (r *ProjectRepo) IsMember(ctx context.Context, projectID, userID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM project_members WHERE project_id = $1 AND user_id = $2)`
	err := r.db.QueryRow(ctx, query, projectID, userID).Scan(&exists)
	return exists, err
}

func (r *ProjectRepo) GetTaskStats(ctx context.Context, projectID int64) (*TaskStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'todo') as todo,
			COUNT(*) FILTER (WHERE status = 'in_progress') as in_progress,
			COUNT(*) FILTER (WHERE status = 'done') as done
		FROM tasks WHERE project_id = $1
	`
	var stats TaskStats
	err := r.db.QueryRow(ctx, query, projectID).Scan(&stats.Total, &stats.Todo, &stats.InProgress, &stats.Done)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
