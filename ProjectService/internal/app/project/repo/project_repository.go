package project

import (
	"context"
	"errors"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app"

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
			app.Logger.Warn("project not created, 0 rows returned")
			return nil, err
		}
		app.Logger.Error("project not created, system error")
		return nil, err
	}
	return project, nil
}
