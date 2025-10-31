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

func (p *projectCore) DeleteProject(ctx context.Context, projectID string) error {
	if projectID == "" {
		log.Printf("empty projectID provided for deletion")
		return fmt.Errorf("projectID cannot be empty")
	}
	err := p.project.DeleteProject(ctx, projectID)
	if err != nil {
		log.Printf("failed to delete project %s: %v", projectID, err)
		return err
	}
	log.Printf("project %s deleted successfully", projectID)
	return nil
}

func (p *projectCore) CreateTask(ctx context.Context, createTask *models.CreateTaskRequest) (*models.Task, error) {
	newTask, err := p.project.CreateTask(ctx, createTask)
	if err != nil {
		return nil, err
	}
	return newTask, err
}

func (p *projectCore) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	task, err := p.project.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (p *projectCore) UpdateTask(ctx context.Context, updRequest *models.UpdateTaskRequest) (*models.Task, error) {
	if updRequest.ID == "" {
		log.Printf("task id is empty")
		return nil, fmt.Errorf("taskID is empty")
	}
	updTask, err := p.project.UpdateTask(ctx, updRequest)
	if err != nil {
		return nil, err
	}
	return updTask, nil
}

func (p *projectCore) DeleteTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		log.Printf("empty taskID provided for deletion")
		return fmt.Errorf("taskID cannot be empty")
	}
	err := p.project.DeleteTask(ctx, taskID)
	if err != nil {
		log.Printf("failed to delete task %s: %v", taskID, err)
		return err
	}
	log.Printf("task %s deleted successfully", taskID)
	return nil
}

func (p *projectCore) MoveTask(ctx context.Context, moveRequest *models.MoveTaskRequest) (*models.Task, error) {
	if moveRequest.TaskID == "" {
		log.Printf("empty taskID provided for moving")
		return nil, fmt.Errorf("taskID cannot be empty")
	}
	movedTask, err := p.project.MoveTask(ctx, moveRequest)
	if err != nil {
		return nil, err
	}
	return movedTask, nil
}

func (p *projectCore) AssignTask(ctx context.Context, assignRequest *models.AssignTaskRequest) (*models.Task, error) {
	if assignRequest.TaskID == "" {
		log.Printf("empty taskID provided for assignment")
		return nil, fmt.Errorf("taskID cannot be empty")
	}
	assignedTask, err := p.project.AssignTask(ctx, assignRequest)
	if err != nil {
		return nil, err
	}
	return assignedTask, nil
}

func (p *projectCore) ListTasksByProject(ctx context.Context, filter *models.ListTasksFilter) (*models.TasksListResponse, error) {
	if filter.ProjectID == "" {
		log.Printf("empty projectID provided for listing tasks")
		return nil, fmt.Errorf("projectID cannot be empty")
	}
	tasks, err := p.project.ListTasksByProject(ctx, filter)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (p *projectCore) AddMemberToProject(ctx context.Context, projectID, userID string) error {
	if projectID == "" || userID == "" {
		log.Printf("empty projectID or userID provided")
		return fmt.Errorf("projectID and userID cannot be empty")
	}
	err := p.project.AddMemberToProject(ctx, projectID, userID)
	if err != nil {
		log.Printf("failed to add member %s to project %s: %v", userID, projectID, err)
		return err
	}
	log.Printf("member %s added to project %s successfully", userID, projectID)
	return nil
}

func (p *projectCore) RemoveMemberFromProject(ctx context.Context, projectID, userID string) error {
	if projectID == "" || userID == "" {
		log.Printf("empty projectID or userID provided")
		return fmt.Errorf("projectID and userID cannot be empty")
	}
	err := p.project.RemoveMemberFromProject(ctx, projectID, userID)
	if err != nil {
		log.Printf("failed to remove member %s from project %s: %v", userID, projectID, err)
		return err
	}
	log.Printf("member %s removed from project %s successfully", userID, projectID)
	return nil
}

func (p *projectCore) ListProjectMembers(ctx context.Context, projectID string) (*models.ProjectMembersResponse, error) {
	if projectID == "" {
		log.Printf("empty projectID provided for listing members")
		return nil, fmt.Errorf("projectID cannot be empty")
	}
	members, err := p.project.ListProjectMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return members, nil
}
