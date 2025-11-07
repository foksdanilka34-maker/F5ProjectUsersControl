package project

import (
	"context"
	"errors"
	"log"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
	CreateProject(ctx context.Context, createProject *models.CreateProjectRequest) (*models.Project, error)
	GetProject(ctx context.Context, projectID string) (*models.Project, error)
	ListProjects(ctx context.Context, listProject *models.ListProjectsFilter) (*models.ProjectsListResponse, error)
	UpdateProject(ctx context.Context, updRequest *models.UpdateProjectRequest) (*models.Project, error)
	DeleteProject(ctx context.Context, projectID string) error

	CreateTask(ctx context.Context, createTask *models.CreateTaskRequest) (*models.Task, error)
	GetTask(ctx context.Context, taskID string) (*models.Task, error)
	UpdateTask(ctx context.Context, updRequest *models.UpdateTaskRequest) (*models.Task, error)
	DeleteTask(ctx context.Context, taskID string) error
	MoveTask(ctx context.Context, moveRequest *models.MoveTaskRequest) (*models.Task, error)
	AssignTask(ctx context.Context, assignRequest *models.AssignTaskRequest) (*models.Task, error)
	ListTasksByProject(ctx context.Context, filter *models.ListTasksFilter) (*models.TasksListResponse, error)

	AddMemberToProject(ctx context.Context, projectID, userID string) error
	RemoveMemberFromProject(ctx context.Context, projectID, userID string) error
	ListProjectMembers(ctx context.Context, projectID string) (*models.ProjectMembersResponse, error)
}

type UserMetaStorage interface {
	UpsertUserMeta(ctx context.Context, userID, userName, userPhoto string) error
	DeleteUserMeta(ctx context.Context, userID string) error
}

func (s *Storage) CreateProject(ctx context.Context, createProject *models.CreateProjectRequest) (*models.Project, error) {
	query := `INSERT INTO project.projects (project_name, project_description, 
			manager_id, due_date) VALUES ($1,$2,$3,$4)
			RETURNING project_id, project_name, project_description, manager_id, project_status,
			created_at, updated_at, due_date`

	project := &models.Project{}
	err := s.pgx.QueryRow(ctx, query, createProject.Name, createProject.Description, createProject.ManagerID, createProject.DueDate).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.ManagerID,
		&project.Status,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.DueDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("project not created, 0 rows returned")
			return nil, err
		}
		log.Printf("project not created, system error: %v", err)
		return nil, err
	}
	return project, nil
}

func (s *Storage) GetProject(ctx context.Context, projectID string) (*models.Project, error) {
	query := `SELECT project_id, project_name, project_description, manager_id, project_status,
				created_at, updated_at, due_date FROM project.projects
				WHERE project_id = $1`

	project := &models.Project{}
	err := s.pgx.QueryRow(ctx, query, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.ManagerID,
		&project.Status,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.DueDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("project not found")
			return nil, err
		}
		log.Printf("system error while getting project: %v", err)
		return nil, err
	}
	return project, nil
}

func (s *Storage) ListProjects(ctx context.Context, listProject *models.ListProjectsFilter) (*models.ProjectsListResponse, error) {
	query := `SELECT project_id, project_name, project_description, manager_id, project_status,
			created_at, updated_at, due_date FROM project.projects
			WHERE ($1::uuid IS NULL OR manager_id = $1) 
			AND ($2::text IS NULL OR project_status = $2::project.project_status)
			LIMIT $3 OFFSET $4`

	offset := (listProject.PageNumber - 1) * listProject.PageSize

	log.Printf("ListProjects query params: ManagerID=%v, Status=%v, PageSize=%d, Offset=%d",
		listProject.ManagerID, listProject.Status, listProject.PageSize, offset)

	rows, err := s.pgx.Query(ctx, query, listProject.ManagerID, listProject.Status, listProject.PageSize, offset)
	if err != nil {
		log.Printf("query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	projects := make([]*models.Project, 0)
	for rows.Next() {
		project := &models.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.ManagerID,
			&project.Status,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.DueDate,
		)
		if err != nil {
			log.Printf("error scanning rows: %v", err)
			return nil, err
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows iteration error: %v", err)
		return nil, err
	}

	log.Printf("ListProjects found %d projects", len(projects))

	response := &models.ProjectsListResponse{
		Projects:   projects,
		TotalCount: int32(len(projects)),
	}
	return response, nil
}

func (s *Storage) UpdateProject(ctx context.Context, updRequest *models.UpdateProjectRequest) (*models.Project, error) {
	query := `UPDATE project.projects SET 
			project_name = COALESCE($1, project_name),
			project_description = COALESCE($2, project_description),
			project_status = COALESCE($3::project.project_status, project_status),
			due_date = COALESCE($4, due_date),
			updated_at = NOW()
			WHERE project_id = $5
			RETURNING project_id, project_name, project_description, 
			manager_id, project_status, created_at, updated_at, due_date`

	var statusDB *string
	if updRequest.Status != nil {
		s := updRequest.Status.DBValue()
		statusDB = &s
	}

	updProject := &models.Project{}
	err := s.pgx.QueryRow(ctx, query, updRequest.Name, updRequest.Description, statusDB, updRequest.DueDate, updRequest.ID).Scan(
		&updProject.ID,
		&updProject.Name,
		&updProject.Description,
		&updProject.ManagerID,
		&updProject.Status,
		&updProject.CreatedAt,
		&updProject.UpdatedAt,
		&updProject.DueDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("project not updated")
			return nil, err
		}
		log.Printf("system error while updating project: %v", err)
		return nil, err
	}
	return updProject, nil
}

func (s *Storage) DeleteProject(ctx context.Context, projectID string) error {
	query := `DELETE FROM project.projects WHERE project_id = $1`

	result, err := s.pgx.Exec(ctx, query, projectID)
	if err != nil {
		log.Printf("system error while deleting project: %v", err)
		return err
	}
	if result.RowsAffected() == 0 {
		log.Printf("project with id %s not found for deletion", projectID)
		return errors.New("project not found")
	}

	log.Printf("project with id %s deleted", projectID)
	return nil
}

func (s *Storage) CreateTask(ctx context.Context, createTask *models.CreateTaskRequest) (*models.Task, error) {
	query := `INSERT INTO project.tasks (project_id, task_name, task_description,
			task_priority, task_status, creator_id, assign_id, due_date)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING task_id, project_id, task_name, task_description,
			task_priority, task_status, creator_id, assign_id, order_index,
			created_at, updated_at, due_date, started_at, completed_at`

	newTask := &models.Task{}
	err := s.pgx.QueryRow(ctx, query, createTask.ProjectID, createTask.TaskName, createTask.Description,
		createTask.Priority.DBValue(), createTask.Status.DBValue(), createTask.CreatorID,
		createTask.AssigneeID, createTask.DueDate).Scan(
		&newTask.ID,
		&newTask.ProjectID,
		&newTask.TaskName,
		&newTask.Description,
		&newTask.Priority,
		&newTask.Status,
		&newTask.CreatorID,
		&newTask.AssigneeID,
		&newTask.OrderIndex,
		&newTask.CreatedAt,
		&newTask.UpdatedAt,
		&newTask.DueDate,
		&newTask.StartedAt,
		&newTask.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("error creating task, 0 rows returned %v", err)
			return nil, err
		}
		log.Printf("system error %v", err)
		return nil, err
	}

	return newTask, nil
}

func (s *Storage) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	query := `SELECT task_id, project_id, task_name, task_description,
			task_priority, task_status, creator_id, assign_id, order_index,
			created_at, updated_at, due_date, started_at, completed_at FROM project.tasks
			WHERE task_id = $1`

	newTask := &models.Task{}
	err := s.pgx.QueryRow(ctx, query, taskID).Scan(
		&newTask.ID,
		&newTask.ProjectID,
		&newTask.TaskName,
		&newTask.Description,
		&newTask.Priority,
		&newTask.Status,
		&newTask.CreatorID,
		&newTask.AssigneeID,
		&newTask.OrderIndex,
		&newTask.CreatedAt,
		&newTask.UpdatedAt,
		&newTask.DueDate,
		&newTask.StartedAt,
		&newTask.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("task is not existed %v", err)
			return nil, err
		}
		log.Printf("system error %v", err)
		return nil, err
	}
	return newTask, nil
}

func (s *Storage) UpdateTask(ctx context.Context, updRequest *models.UpdateTaskRequest) (*models.Task, error) {
	query := `UPDATE project.tasks SET
			task_name = COALESCE($1, task_name),
			task_description = COALESCE($2, task_description),
			task_status = COALESCE($3::project.task_status, task_status),
			task_priority = COALESCE($4::project.task_priority, task_priority),
			assign_id = COALESCE($5, assign_id),
			due_date = COALESCE($6, due_date),
			order_index = COALESCE($7, order_index),
			started_at = CASE 
				WHEN $3::project.task_status IN ('IN_PROGRESS', 'REVIEW') AND started_at IS NULL 
				THEN NOW() 
				ELSE started_at 
			END,
			completed_at = CASE 
				WHEN $3::project.task_status = 'DONE' AND completed_at IS NULL 
				THEN NOW() 
				ELSE completed_at 
			END,
			updated_at = NOW()
			WHERE task_id = $8
			RETURNING task_id, project_id, task_name, task_description,
			task_priority, task_status, creator_id, assign_id, order_index,
			created_at, updated_at, due_date, started_at, completed_at`

	var statusDB *string
	if updRequest.Status != nil {
		s := updRequest.Status.DBValue()
		statusDB = &s
	}

	var priorityDB *string
	if updRequest.Priority != nil {
		p := updRequest.Priority.DBValue()
		priorityDB = &p
	}

	updTask := &models.Task{}
	err := s.pgx.QueryRow(ctx, query, updRequest.TaskName, updRequest.Description,
		statusDB, priorityDB, updRequest.AssigneeID, updRequest.DueDate,
		updRequest.OrderIndex, updRequest.ID).Scan(
		&updTask.ID,
		&updTask.ProjectID,
		&updTask.TaskName,
		&updTask.Description,
		&updTask.Priority,
		&updTask.Status,
		&updTask.CreatorID,
		&updTask.AssigneeID,
		&updTask.OrderIndex,
		&updTask.CreatedAt,
		&updTask.UpdatedAt,
		&updTask.DueDate,
		&updTask.StartedAt,
		&updTask.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("task not updated, no rows affected")
			return nil, err
		}
		log.Printf("system error while updating task: %v", err)
		return nil, err
	}

	return updTask, nil
}

func (s *Storage) DeleteTask(ctx context.Context, taskID string) error {
	query := `DELETE FROM project.tasks WHERE task_id = $1`

	result, err := s.pgx.Exec(ctx, query, taskID)
	if err != nil {
		log.Printf("system error while deleting task: %v", err)
		return err
	}
	if result.RowsAffected() == 0 {
		log.Printf("task with id %s not found for deletion", taskID)
		return errors.New("task not found")
	}

	log.Printf("task with id %s deleted", taskID)
	return nil
}

func (s *Storage) MoveTask(ctx context.Context, moveRequest *models.MoveTaskRequest) (*models.Task, error) {
	query := `UPDATE project.tasks SET
			task_status = $1::project.task_status,
			order_index = $2,
			started_at = CASE 
				WHEN $1::project.task_status IN ('IN_PROGRESS', 'REVIEW') AND started_at IS NULL 
				THEN NOW() 
				ELSE started_at 
			END,
			completed_at = CASE 
				WHEN $1::project.task_status = 'DONE' AND completed_at IS NULL 
				THEN NOW() 
				ELSE completed_at 
			END,
			updated_at = NOW()
			WHERE task_id = $3
			RETURNING task_id, project_id, task_name, task_description,
			task_priority, task_status, creator_id, assign_id, order_index,
			created_at, updated_at, due_date, started_at, completed_at`

	movedTask := &models.Task{}
	err := s.pgx.QueryRow(ctx, query, moveRequest.NewStatus.DBValue(),
		moveRequest.NewOrderIndex, moveRequest.TaskID).Scan(
		&movedTask.ID,
		&movedTask.ProjectID,
		&movedTask.TaskName,
		&movedTask.Description,
		&movedTask.Priority,
		&movedTask.Status,
		&movedTask.CreatorID,
		&movedTask.AssigneeID,
		&movedTask.OrderIndex,
		&movedTask.CreatedAt,
		&movedTask.UpdatedAt,
		&movedTask.DueDate,
		&movedTask.StartedAt,
		&movedTask.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("task not found for moving")
			return nil, err
		}
		log.Printf("system error while moving task: %v", err)
		return nil, err
	}

	return movedTask, nil
}

func (s *Storage) AssignTask(ctx context.Context, assignRequest *models.AssignTaskRequest) (*models.Task, error) {
	query := `UPDATE project.tasks SET
			assign_id = $1,
			updated_at = NOW()
			WHERE task_id = $2
			RETURNING task_id, project_id, task_name, task_description,
			task_priority, task_status, creator_id, assign_id, order_index,
			created_at, updated_at, due_date, started_at, completed_at`

	assignedTask := &models.Task{}
	err := s.pgx.QueryRow(ctx, query, assignRequest.AssigneeID, assignRequest.TaskID).Scan(
		&assignedTask.ID,
		&assignedTask.ProjectID,
		&assignedTask.TaskName,
		&assignedTask.Description,
		&assignedTask.Priority,
		&assignedTask.Status,
		&assignedTask.CreatorID,
		&assignedTask.AssigneeID,
		&assignedTask.OrderIndex,
		&assignedTask.CreatedAt,
		&assignedTask.UpdatedAt,
		&assignedTask.DueDate,
		&assignedTask.StartedAt,
		&assignedTask.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("task not found for assignment")
			return nil, err
		}
		log.Printf("system error while assigning task: %v", err)
		return nil, err
	}
	return assignedTask, nil
}

func (s *Storage) ListTasksByProject(ctx context.Context, filter *models.ListTasksFilter) (*models.TasksListResponse, error) {
	query := `SELECT task_id, project_id, task_name, task_description,
			task_priority, task_status, creator_id, assign_id, order_index,
			created_at, updated_at, due_date, started_at, completed_at
			FROM project.tasks
			WHERE project_id = $1
			AND ($2::project.task_status IS NULL OR task_status = $2)
			AND ($3::uuid IS NULL OR assign_id = $3)
			AND ($4::project.task_priority IS NULL OR task_priority = $4)
			ORDER BY order_index ASC`

	var statusDB *string
	if filter.Status != nil {
		s := filter.Status.DBValue()
		statusDB = &s
	}

	var priorityDB *string
	if filter.Priority != nil {
		p := filter.Priority.DBValue()
		priorityDB = &p
	}

	rows, err := s.pgx.Query(ctx, query, filter.ProjectID, statusDB, filter.AssigneeID, priorityDB)
	if err != nil {
		log.Printf("query error while listing tasks: %v", err)
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*models.Task, 0)
	for rows.Next() {
		task := &models.Task{}
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.TaskName,
			&task.Description,
			&task.Priority,
			&task.Status,
			&task.CreatorID,
			&task.AssigneeID,
			&task.OrderIndex,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DueDate,
			&task.StartedAt,
			&task.CompletedAt,
		)
		if err != nil {
			log.Printf("error scanning task rows: %v", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows iteration error: %v", err)
		return nil, err
	}

	log.Printf("ListTasksByProject found %d tasks", len(tasks))

	response := &models.TasksListResponse{
		Tasks: tasks,
	}
	return response, nil
}

func (s *Storage) AddMemberToProject(ctx context.Context, projectID, userID string) error {
	query := `INSERT INTO project.project_members (project_id, user_id, role)
			VALUES ($1, $2, $3)
			ON CONFLICT (project_id, user_id) DO NOTHING`

	result, err := s.pgx.Exec(ctx, query, projectID, userID, "MEMBER")
	if err != nil {
		log.Printf("system error while adding member to project: %v", err)
		return err
	}

	if result.RowsAffected() == 0 {
		log.Printf("member already exists in project or no action taken")
	} else {
		log.Printf("member %s added to project %s", userID, projectID)
	}

	return nil
}

func (s *Storage) RemoveMemberFromProject(ctx context.Context, projectID, userID string) error {
	query := `DELETE FROM project.project_members
			WHERE project_id = $1 AND user_id = $2`

	result, err := s.pgx.Exec(ctx, query, projectID, userID)
	if err != nil {
		log.Printf("system error while removing member from project: %v", err)
		return err
	}

	if result.RowsAffected() == 0 {
		log.Printf("member %s not found in project %s", userID, projectID)
		return errors.New("member not found in project")
	}

	log.Printf("member %s removed from project %s", userID, projectID)
	return nil
}

func (s *Storage) ListProjectMembers(ctx context.Context, projectID string) (*models.ProjectMembersResponse, error) {
	query := `SELECT pm.user_id, COALESCE(um.user_name, 'Unknown'), pm.role
			FROM project.project_members pm
			LEFT JOIN project.users_meta um ON pm.user_id = um.user_id
			WHERE pm.project_id = $1
			ORDER BY pm.added_at ASC`

	rows, err := s.pgx.Query(ctx, query, projectID)
	if err != nil {
		log.Printf("query error while listing project members: %v", err)
		return nil, err
	}
	defer rows.Close()

	members := make([]*models.ProjectMember, 0)
	for rows.Next() {
		member := &models.ProjectMember{
			ProjectID: projectID,
		}
		var roleStr string
		err := rows.Scan(&member.UserID, &member.FullName, &roleStr)
		if err != nil {
			log.Printf("error scanning member rows: %v", err)
			return nil, err
		}
		member.Role = models.ParseProjectRole(roleStr)
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		log.Printf("rows iteration error: %v", err)
		return nil, err
	}

	log.Printf("ListProjectMembers found %d members", len(members))

	response := &models.ProjectMembersResponse{
		Members: members,
	}
	return response, nil
}

func (s *Storage) UpsertUserMeta(ctx context.Context, userID, userName, userPhoto string) error {
	query := `INSERT INTO project.users_meta (user_id, user_name, user_photo, updated_at)
			VALUES ($1, $2, $3, NOW())
			ON CONFLICT (user_id) 
			DO UPDATE SET 
			user_name = EXCLUDED.user_name,
			user_photo = EXCLUDED.user_photo,
			updated_at = NOW()`

	_, err := s.pgx.Exec(ctx, query, userID, userName, userPhoto)
	if err != nil {
		log.Printf("system error while upserting user meta: %v", err)
		return err
	}

	log.Printf("user meta upserted successfully: userID=%s", userID)
	return nil
}

func (s *Storage) DeleteUserMeta(ctx context.Context, userID string) error {
	query := `DELETE FROM project.users_meta WHERE user_id = $1`

	result, err := s.pgx.Exec(ctx, query, userID)
	if err != nil {
		log.Printf("system error while deleting user meta: %v", err)
		return err
	}

	if result.RowsAffected() == 0 {
		log.Printf("user meta with id %s not found for deletion", userID)
	} else {
		log.Printf("user meta with id %s deleted", userID)
	}

	return nil
}