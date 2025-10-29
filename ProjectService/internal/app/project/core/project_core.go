package employee

import (
	"context"
	"fmt"

	repo "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/repo"
	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app"
	"go.uber.org/zap"
)


type projectCore struct {
	project repo.ProjectStorage
}

type CoreLogic interface {	
	CreateProject(ctx context.Context, regProfile *models.CreateProjectRequest) (*models.Project, error)
}

func NewCore(project repo.ProjectStorage) CoreLogic {
	return &projectCore{
		project:  project,
	}
}

func (p *projectCore) CreateProject(ctx context.Context, regProfile *models.CreateProjectRequest) (*models.Project, error) {
	if regProfile == nil {
		app.Logger.Info("empty data provided")
		return nil, fmt.Errorf("provided empty data")
	}
	if regProfile.Name == "" || regProfile.ManagerID == "" || regProfile.Description == nil {
		app.Logger.Info("all fields should be field")
		return nil, fmt.Errorf("all fields should be field")
	}
	project, err := p.project.CreateProject(ctx, regProfile)
	if err != nil {
		return nil, err
	}
	app.Logger.Info("Project created, uuid:", zap.String("pId", project.ID))
	return project, nil
}

