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
	
	updProject := &models.Project{}
	err := s.pgx.QueryRow(ctx, query, updRequest.Name, updRequest.Description, updRequest.Status.DBValue(), updRequest.DueDate, updRequest.ID).Scan(
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
