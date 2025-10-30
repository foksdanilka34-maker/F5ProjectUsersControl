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
