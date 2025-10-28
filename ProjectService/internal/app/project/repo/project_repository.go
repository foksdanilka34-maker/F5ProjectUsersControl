package employee

import (
	"context"
	"errors"

	proj "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/employee"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Storage struct {
	pgx *pgxpool.Pool
}

func NewStorage(p *pgxpool.Pool) *Storage {
	return &Storage{
		pgx: p,
	}
}

type ProjectStorage interface {
	CreateProject(ctx context.Context, req *proj.CreateProjectRequest) (*proj.Project, error)
	GetProject(ctx context.Context, projectID string) (*proj.Project, error)
	ListProjects(ctx context.Context, filter *proj.ListProjectsFilter) (*proj.ProjectsListResponse, error)
	UpdateProject(ctx context.Context, req *proj.UpdateProjectRequest) (*proj.Project, error)
	DeleteProject(ctx context.Context, projectID string) error

	CreateTask(ctx context.Context, req *proj.CreateTaskRequest) (*proj.Task, error)
	GetTask(ctx context.Context, taskID string) (*proj.Task, error)
	ListTasksByProject(ctx context.Context, filter *proj.ListTasksFilter) ([]*proj.Task, error)
	UpdateTask(ctx context.Context, req *proj.UpdateTaskRequest) (*proj.Task, error)
	DeleteTask(ctx context.Context, taskID string) error
	MoveTask(ctx context.Context, req *proj.MoveTaskRequest) (*proj.Task, error)
	AssignTask(ctx context.Context, req *proj.AssignTaskRequest) (*proj.Task, error)

	AddMemberToProject(ctx context.Context, req *proj.AddMemberRequest) error
	RemoveMemberFromProject(ctx context.Context, req *proj.RemoveMemberRequest) error
	ListProjectMembers(ctx context.Context, projectID string) ([]*proj.ProjectMember, error)

	BeginTransaction(ctx context.Context) (pgx.Tx, error)
}

func (s *Storage) BeginTransaction(ctx context.Context) (pgx.Tx, error) {
	return s.pgx.Begin(ctx)
}

func (s *Storage) CreateProject(ctx context.Context, req *proj.CreateProjectRequest) (*proj.Project, error) {
	query := `INSERT INTO projects.projects (name, description, manager_id, due_date)
			VALUES ($1, $2, $3, $4)
			RETURNING id, name, description, manager_id, status, due_date, created_at, updated_at`

	project := &proj.Project{}
	err := s.pgx.QueryRow(ctx, query, req.Name, req.Description, req.ManagerID, req.DueDate).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.ManagerID,
		&project.Status,
		&project.DueDate,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		storage.Logger.Error("error creating project", zap.Error(err))
		return nil, err
	}
	return project, nil
}

func (s *Storage) GetProject(ctx context.Context, projectID string) (*proj.Project, error) {
	query := `SELECT id, name, description, manager_id, status, due_date, created_at, updated_at
			FROM projects.projects WHERE id = $1`

	project := &proj.Project{}
	err := s.pgx.QueryRow(ctx, query, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.ManagerID,
		&project.Status,
		&project.DueDate,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			storage.Logger.Warn("project not found", zap.String("projectID", projectID))
		} else {
			storage.Logger.Error("error getting project", zap.Error(err))
		}
		return nil, err
	}
	return project, nil
}

func (s *Storage) ListProjects(ctx context.Context, filter *proj.ListProjectsFilter) (*proj.ProjectsListResponse, error) {
	countQuery := `SELECT COUNT(*) FROM projects.projects
				WHERE ($1::UUID IS NULL OR manager_id = $1)
				AND ($2::SMALLINT IS NULL OR status = $2)`

	var totalCount int32
	err := s.pgx.QueryRow(ctx, countQuery, filter.ManagerID, filter.Status).Scan(&totalCount)
	if err != nil {
		storage.Logger.Error("error counting projects", zap.Error(err))
		return nil, err
	}

	query := `SELECT id, name, description, manager_id, status, due_date, created_at, updated_at
			FROM projects.projects
			WHERE ($1::UUID IS NULL OR manager_id = $1)
			AND ($2::SMALLINT IS NULL OR status = $2)
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4`

	offset := (filter.PageNumber - 1) * filter.PageSize
	rows, err := s.pgx.Query(ctx, query, filter.ManagerID, filter.Status, filter.PageSize, offset)
	if err != nil {
		storage.Logger.Error("error listing projects", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	projects := make([]proj.Project, 0, filter.PageSize)
	for rows.Next() {
		project := proj.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.ManagerID,
			&project.Status,
			&project.DueDate,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			storage.Logger.Error("error scanning project row", zap.Error(err))
			return nil, err
		}
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		storage.Logger.Error("error iterating project rows", zap.Error(err))
		return nil, err
	}

	return &proj.ProjectsListResponse{
		Projects:   projects,
		TotalCount: totalCount,
	}, nil
}

func (s *Storage) UpdateProject(ctx context.Context, req *proj.UpdateProjectRequest) (*proj.Project, error) {
	query := `UPDATE projects.projects SET
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			status = COALESCE($3, status),
			due_date = COALESCE($4, due_date),
			updated_at = NOW()
			WHERE id = $5
			RETURNING id, name, description, manager_id, status, due_date, created_at, updated_at`

	project := &proj.Project{}
	err := s.pgx.QueryRow(ctx, query, req.Name, req.Description, req.Status, req.DueDate, req.ID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.ManagerID,
		&project.Status,
		&project.DueDate,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			storage.Logger.Warn("project not found for update", zap.String("projectID", req.ID))
		} else {
			storage.Logger.Error("error updating project", zap.Error(err))
		}
		return nil, err
	}
	return project, nil
}

func (s *Storage) DeleteProject(ctx context.Context, projectID string) error {
	query := `UPDATE projects.projects SET status = $1, updated_at = NOW() WHERE id = $2`
	cmdTag, err := s.pgx.Exec(ctx, query, proj.ProjectStatusArchived, projectID)
	if err != nil {
		storage.Logger.Error("error archiving project", zap.Error(err))
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		storage.Logger.Warn("project not found for archiving", zap.String("projectID", projectID))
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Storage) CreateTask(ctx context.Context, req *proj.CreateTaskRequest) (*proj.Task, error) {
	query := `INSERT INTO projects.tasks (project_id, title, description, priority, assignee_id, creator_id, due_date, order_index)
			VALUES ($1, $2, $3, $4, $5, $6, $7, COALESCE((SELECT MAX(order_index) + 1 FROM projects.tasks WHERE project_id = $1 AND status = 1), 0))
			RETURNING id, project_id, title, description, status, priority, assignee_id, creator_id, order_index, due_date, created_at, updated_at`

	task := &proj.Task{}
	err := s.pgx.QueryRow(ctx, query, req.ProjectID, req.Title, req.Description, req.Priority, req.AssigneeID, req.CreatorID, req.DueDate).Scan(
		&task.ID,
		&task.ProjectID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.AssigneeID,
		&task.CreatorID,
		&task.OrderIndex,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		storage.Logger.Error("error creating task", zap.Error(err))
		return nil, err
	}
	return task, nil
}

func (s *Storage) GetTask(ctx context.Context, taskID string) (*proj.Task, error) {
	query := `SELECT id, project_id, title, description, status, priority, assignee_id, creator_id, order_index, due_date, created_at, updated_at
			FROM projects.tasks WHERE id = $1`

	task := &proj.Task{}
	err := s.pgx.QueryRow(ctx, query, taskID).Scan(
		&task.ID,
		&task.ProjectID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.AssigneeID,
		&task.CreatorID,
		&task.OrderIndex,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			storage.Logger.Warn("task not found", zap.String("taskID", taskID))
		} else {
			storage.Logger.Error("error getting task", zap.Error(err))
		}
		return nil, err
	}
	return task, nil
}

func (s *Storage) ListTasksByProject(ctx context.Context, filter *proj.ListTasksFilter) ([]*proj.Task, error) {
	query := `SELECT id, project_id, title, description, status, priority, assignee_id, creator_id, order_index, due_date, created_at, updated_at
			FROM projects.tasks
			WHERE project_id = $1
			AND ($2::SMALLINT IS NULL OR status = $2)
			AND ($3::UUID IS NULL OR assignee_id = $3)
			AND ($4::SMALLINT IS NULL OR priority = $4)
			ORDER BY order_index ASC`

	rows, err := s.pgx.Query(ctx, query, filter.ProjectID, filter.Status, filter.AssigneeID, filter.Priority)
	if err != nil {
		storage.Logger.Error("error listing tasks", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*proj.Task, 0)
	for rows.Next() {
		task := &proj.Task{}
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.AssigneeID,
			&task.CreatorID,
			&task.OrderIndex,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			storage.Logger.Error("error scanning task row", zap.Error(err))
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		storage.Logger.Error("error iterating task rows", zap.Error(err))
		return nil, err
	}

	return tasks, nil
}

func (s *Storage) UpdateTask(ctx context.Context, req *proj.UpdateTaskRequest) (*proj.Task, error) {
	query := `UPDATE projects.tasks SET
			title = COALESCE($1, title),
			description = COALESCE($2, description),
			status = COALESCE($3, status),
			priority = COALESCE($4, priority),
			assignee_id = COALESCE($5, assignee_id),
			due_date = COALESCE($6, due_date),
			updated_at = NOW()
			WHERE id = $7
			RETURNING id, project_id, title, description, status, priority, assignee_id, creator_id, order_index, due_date, created_at, updated_at`

	task := &proj.Task{}
	err := s.pgx.QueryRow(ctx, query, req.Title, req.Description, req.Status, req.Priority, req.AssigneeID, req.DueDate, req.ID).Scan(
		&task.ID,
		&task.ProjectID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.AssigneeID,
		&task.CreatorID,
		&task.OrderIndex,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			storage.Logger.Warn("task not found for update", zap.String("taskID", req.ID))
		} else {
			storage.Logger.Error("error updating task", zap.Error(err))
		}
		return nil, err
	}
	return task, nil
}

func (s *Storage) DeleteTask(ctx context.Context, taskID string) error {
	query := `DELETE FROM projects.tasks WHERE id = $1`
	cmdTag, err := s.pgx.Exec(ctx, query, taskID)
	if err != nil {
		storage.Logger.Error("error deleting task", zap.Error(err))
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		storage.Logger.Warn("task not found for deletion", zap.String("taskID", taskID))
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Storage) MoveTask(ctx context.Context, req *proj.MoveTaskRequest) (*proj.Task, error) {
	query := `UPDATE projects.tasks SET
			status = $1,
			order_index = $2,
			updated_at = NOW()
			WHERE id = $3
			RETURNING id, project_id, title, description, status, priority, assignee_id, creator_id, order_index, due_date, created_at, updated_at`

	task := &proj.Task{}
	err := s.pgx.QueryRow(ctx, query, req.NewStatus, req.NewOrderIndex, req.TaskID).Scan(
		&task.ID,
		&task.ProjectID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.AssigneeID,
		&task.CreatorID,
		&task.OrderIndex,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			storage.Logger.Warn("task not found for move", zap.String("taskID", req.TaskID))
		} else {
			storage.Logger.Error("error moving task", zap.Error(err))
		}
		return nil, err
	}
	return task, nil
}

func (s *Storage) AssignTask(ctx context.Context, req *proj.AssignTaskRequest) (*proj.Task, error) {
	query := `UPDATE projects.tasks SET
			assignee_id = $1,
			updated_at = NOW()
			WHERE id = $2
			RETURNING id, project_id, title, description, status, priority, assignee_id, creator_id, order_index, due_date, created_at, updated_at`

	task := &proj.Task{}
	err := s.pgx.QueryRow(ctx, query, req.AssigneeID, req.TaskID).Scan(
		&task.ID,
		&task.ProjectID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.AssigneeID,
		&task.CreatorID,
		&task.OrderIndex,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			storage.Logger.Warn("task not found for assignment", zap.String("taskID", req.TaskID))
		} else {
			storage.Logger.Error("error assigning task", zap.Error(err))
		}
		return nil, err
	}
	return task, nil
}

func (s *Storage) AddMemberToProject(ctx context.Context, req *proj.AddMemberRequest) error {
	query := `INSERT INTO projects.project_members (project_id, user_id, role)
			VALUES ($1, $2, $3)
			ON CONFLICT (project_id, user_id) DO UPDATE SET role = $3`

	cmdTag, err := s.pgx.Exec(ctx, query, req.ProjectID, req.UserID, req.Role)
	if err != nil {
		storage.Logger.Error("error adding member to project", zap.Error(err))
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		storage.Logger.Warn("unable to add member to project")
		return errors.New("unable to add member")
	}
	return nil
}

func (s *Storage) RemoveMemberFromProject(ctx context.Context, req *proj.RemoveMemberRequest) error {
	query := `DELETE FROM projects.project_members WHERE project_id = $1 AND user_id = $2`
	cmdTag, err := s.pgx.Exec(ctx, query, req.ProjectID, req.UserID)
	if err != nil {
		storage.Logger.Error("error removing member from project", zap.Error(err))
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		storage.Logger.Warn("member not found for removal", zap.String("projectID", req.ProjectID), zap.String("userID", req.UserID))
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Storage) ListProjectMembers(ctx context.Context, projectID string) ([]*proj.ProjectMember, error) {
	query := `SELECT user_id, role FROM projects.project_members WHERE project_id = $1 ORDER BY role`

	rows, err := s.pgx.Query(ctx, query, projectID)
	if err != nil {
		storage.Logger.Error("error listing project members", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	members := make([]*proj.ProjectMember, 0)
	for rows.Next() {
		member := &proj.ProjectMember{}
		err := rows.Scan(&member.UserID, &member.Role)
		if err != nil {
			storage.Logger.Error("error scanning project member row", zap.Error(err))
			return nil, err
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		storage.Logger.Error("error iterating project member rows", zap.Error(err))
		return nil, err
	}

	return members, nil
}
