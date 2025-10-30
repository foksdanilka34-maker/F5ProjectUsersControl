package employee

import (
	"context"
	"fmt"
	"log"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"
	repo "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/repo"
)

type projectCore struct {
	project repo.ProjectStorage
}

type CoreLogic interface {
	CreateProject(ctx context.Context, regProfile *models.CreateProjectRequest) (*models.Project, error)
	GetProject(ctx context.Context, projectID string) (*models.Project, error)
	ListProjects(ctx context.Context, listProject *models.ListProjectsFilter) (*models.ProjectsListResponse, error)
	UpdateProject(ctx context.Context, updRequest *models.UpdateProjectRequest) (*models.Project, error)
}

func NewCore(project repo.ProjectStorage) CoreLogic {
	return &projectCore{
		project: project,
	}
}

func (p *projectCore) CreateProject(ctx context.Context, regProfile *models.CreateProjectRequest) (*models.Project, error) {
	if regProfile == nil {
		log.Printf("empty data provided")
		return nil, fmt.Errorf("provided empty data")
	}
	if regProfile.Name == "" || regProfile.ManagerID == "" || regProfile.Description == nil {
		log.Printf("all fields should be filled")
		return nil, fmt.Errorf("all fields should be filled")
	}
	project, err := p.project.CreateProject(ctx, regProfile)
	if err != nil {
		return nil, err
	}
	log.Printf("Project created, uuid: %s", project.ID)
	return project, nil
}

func (p *projectCore) GetProject(ctx context.Context, projectID string) (*models.Project, error) {
	if projectID == "" {
		log.Printf("empty data provided, projectID")
		return nil, fmt.Errorf("empty data provided")
	}
	project, err := p.project.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (p *projectCore) ListProjects(ctx context.Context, listProject *models.ListProjectsFilter) (*models.ProjectsListResponse, error) {
	if listProject.PageNumber <= 0 {
		listProject.PageNumber = 1
	}
	if listProject.PageSize <= 0 {
		listProject.PageSize = 5
	}
	projects, err := p.project.ListProjects(ctx, listProject)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (p *projectCore) UpdateProject(ctx context.Context, updRequest *models.UpdateProjectRequest) (*models.Project, error) {
	if updRequest.ID == "" {
		log.Printf("project id is empty")
		return nil, fmt.Errorf("projectId is empty")
	}
	updProject, err := p.project.UpdateProject(ctx, updRequest)
	if err != nil {
		return nil, err
	}
	return updProject, nil
}