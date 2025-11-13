package project

import (
	"context"
	"fmt"
	"log"
	"time"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/client/nats"

	repo "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/repo"
	ev "github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/eventbus"
)

type projectCore struct {
	project   repo.Storage
	publisher *nats.Publisher
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

func NewCore(project repo.Storage, publisher *nats.Publisher) *projectCore {
	return &projectCore{
		project:   project,
		publisher: publisher,
	}
}

func (p *projectCore) CreateProject(ctx context.Context, regProfile *models.CreateProjectRequest) (*models.Project, error) {
	if regProfile == nil {
		log.Printf("empty data provided")
		return nil, fmt.Errorf("provided empty data")
	}
	if regProfile.Name == "" || regProfile.ManagerID == "" {
		log.Printf("all fields should be filled")
		return nil, fmt.Errorf("all fields should be filled")
	}

	if regProfile.DueDate != nil {
		dueDate := *regProfile.DueDate
		if dueDate.Before(time.Now()) {
			log.Printf("due_date must be in the future, got: %v", dueDate)
			return nil, fmt.Errorf("due_date must be in the future")
		}
	}

	project, err := p.project.CreateProject(ctx, regProfile)
	if err != nil {
		return nil, err
	}
	log.Printf("Project created, uuid: %s", project.ID)

	if p.publisher != nil {
		event := &ev.ProjectEvent{
			ProjectID: project.ID,
			ManagerID: project.ManagerID,
			Status:    project.Status.String(),
			DueDate:   project.DueDate,
			Timestamp: time.Now(),
		}
		if err := p.publisher.PublishProjectEvent(ctx, ev.EventTypeProjectCreated, event); err != nil {
			log.Printf("warning: failed to publish project created event: %v", err)
		}
	}

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

	if p.publisher != nil {
		event := &ev.ProjectEvent{
			ProjectID: updProject.ID,
			ManagerID: updProject.ManagerID,
			Status:    updProject.Status.String(),
			DueDate:   updProject.DueDate,
			Timestamp: time.Now(),
		}
		if err := p.publisher.PublishProjectEvent(ctx, ev.EventTypeProjectUpdated, event); err != nil {
			log.Printf("warning: failed to publish project update event: %v", err)
		}
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

	if p.publisher != nil {
		event := &ev.ProjectEvent{
			ProjectID: projectID,
			Timestamp: time.Now(),
		}
		if err := p.publisher.PublishProjectEvent(ctx, ev.EventTypeProjectDeleted, event); err != nil {
			log.Printf("warning: failed to publish project delete event: %v", err)
		}
	}
	return nil
}

func (p *projectCore) CreateTask(ctx context.Context, createTask *models.CreateTaskRequest) (*models.Task, error) {
	if createTask.AssigneeID != nil {
		if err := p.validateProjectMember(ctx, createTask.ProjectID, *createTask.AssigneeID); err != nil {
			log.Printf("assignee %s is not a member of project %s", *createTask.AssigneeID, createTask.ProjectID)
			return nil, fmt.Errorf("assignee must be a member of the project")
		}
	}

	newTask, err := p.project.CreateTask(ctx, createTask)
	if err != nil {
		return nil, err
	}

	if p.publisher != nil {
		if createTask.AssigneeID != nil {
			eventAssign := &ev.TaskEvent{
				TaskID:     newTask.ID,
				ProjectID:  createTask.ProjectID,
				Priority:   createTask.Priority.String(),
				AssigneeID: createTask.AssigneeID,
			}
			if err := p.publisher.PublishTaskEvent(ctx, ev.EventTypeTaskAssigned, eventAssign); err != nil {
				log.Printf("warning: failed to publish task assigned event: %v", err)
			}
			return newTask, nil
		}
		event := &ev.TaskEvent{
			TaskID:      newTask.ID,
			ProjectID:   newTask.ProjectID,
			Status:      newTask.Status.String(),
			AssigneeID:  newTask.AssigneeID,
			CreatorID:   newTask.CreatorID,
			Priority:    newTask.Priority.String(),
			DueDate:     newTask.DueDate,
			StartedAt:   newTask.StartedAt,
			CompletedAt: newTask.CompletedAt,
			Timestamp:   time.Now(),
		}
		if err := p.publisher.PublishTaskEvent(ctx, ev.EventTypeTaskCreated, event); err != nil {
			log.Printf("warning: failed to publish task created event: %v", err)
		}
	}
	return newTask, nil
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

	oldTask, err := p.project.GetTask(ctx, updRequest.ID)
	if err != nil {
		log.Printf("failed to get existing task %s: %v", updRequest.ID, err)
		return nil, fmt.Errorf("task not found")
	}

	if updRequest.AssigneeID != nil {
		if err := p.validateProjectMember(ctx, oldTask.ProjectID, *updRequest.AssigneeID); err != nil {
			log.Printf("assignee %s is not a member of project %s", *updRequest.AssigneeID, oldTask.ProjectID)
			return nil, fmt.Errorf("assignee must be a member of the project")
		}
	}

	updTask, err := p.project.UpdateTask(ctx, updRequest)
	if err != nil {
		return nil, err
	}

	if p.publisher != nil {
		event := &ev.TaskEvent{
			TaskID:      updTask.ID,
			ProjectID:   updTask.ProjectID,
			Status:      updTask.Status.String(),
			AssigneeID:  updTask.AssigneeID,
			Priority:    updTask.Priority.String(),
			DueDate:     updTask.DueDate,
			StartedAt:   updTask.StartedAt,
			CompletedAt: updTask.CompletedAt,
			Timestamp:   time.Now(),
		}
		if oldTask != nil && updRequest.Status != nil && oldTask.Status != *updRequest.Status {
			event.OldStatus = oldTask.Status.String()
			if err := p.publisher.PublishTaskEvent(ctx, ev.EventTypeTaskStatusChanged, event); err != nil {
			log.Printf("warning: failed to publish task updated event: %v", err)
		}
		}
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

	if p.publisher != nil {
		event := &ev.TaskEvent{
			TaskID:    taskID,
			Timestamp: time.Now(),
		}
		if err := p.publisher.PublishTaskEvent(ctx, ev.EventTypeTaskDeleted, event); err != nil {
			log.Printf("warning: failed to publish task delete event: %v", err)
		}
	}
	return nil
}

func (p *projectCore) MoveTask(ctx context.Context, moveRequest *models.MoveTaskRequest) (*models.Task, error) {
	if moveRequest.TaskID == "" {
		log.Printf("empty taskID provided for moving")
		return nil, fmt.Errorf("taskID cannot be empty")
	}

	oldTask, err := p.project.GetTask(ctx, moveRequest.TaskID)
	if err != nil {
		log.Printf("failed to get task %s: %v", moveRequest.TaskID, err)
		return nil, fmt.Errorf("task not found")
	}
	movedTask, err := p.project.MoveTask(ctx, moveRequest)
	if err != nil {
		return nil, err
	}

	if p.publisher != nil && oldTask != nil {
		event := &ev.TaskEvent{
			TaskID:      movedTask.ID,
			ProjectID:   movedTask.ProjectID,
			Status:      movedTask.Status.String(),
			OldStatus:   oldTask.Status.String(),
			AssigneeID:  movedTask.AssigneeID,
			Priority:    movedTask.Priority.String(),
			DueDate:     movedTask.DueDate,
			StartedAt:   movedTask.StartedAt,
			CompletedAt: movedTask.CompletedAt,
			Timestamp:   time.Now(),
		}
		if moveRequest.NewStatus == models.TaskStatusDone {
			if err := p.publisher.PublishTaskEvent(ctx, ev.EventTypeTaskCompleted, event); err != nil {
				log.Printf("warning: failed to publish task completed event: %v", err)
			}
			return movedTask, nil
		}
		if err := p.publisher.PublishTaskEvent(ctx, ev.EventTypeTaskStatusChanged, event); err != nil {
			log.Printf("warning: failed to publish task moved event: %v", err)
		}
	}
	return movedTask, nil
}

func (p *projectCore) AssignTask(ctx context.Context, assignRequest *models.AssignTaskRequest) (*models.Task, error) {
	if assignRequest.TaskID == "" {
		log.Printf("empty taskID provided for assignment")
		return nil, fmt.Errorf("taskID cannot be empty")
	}

	task, err := p.project.GetTask(ctx, assignRequest.TaskID)
	if err != nil {
		log.Printf("failed to get task %s: %v", assignRequest.TaskID, err)
		return nil, fmt.Errorf("task not found")
	}

	if assignRequest.AssigneeID != nil {
		if err := p.validateProjectMember(ctx, task.ProjectID, *assignRequest.AssigneeID); err != nil {
			log.Printf("assignee %s is not a member of project %s", *assignRequest.AssigneeID, task.ProjectID)
			return nil, fmt.Errorf("assignee must be a member of the project")
		}
	}

	assignedTask, err := p.project.AssignTask(ctx, assignRequest)
	if err != nil {
		return nil, err
	}

	if p.publisher != nil {
		event := &ev.TaskEvent{
			TaskID:     assignedTask.ID,
			ProjectID:  task.ProjectID,
			Timestamp:  time.Now(),
			AssigneeID: assignRequest.AssigneeID,
		}
		if err := p.publisher.PublishTaskEvent(ctx, ev.EventTypeTaskAssigned, event); err != nil {
			log.Printf("warning: failed to publish task delete event: %v", err)
		}
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

	if p.publisher != nil {
		event := &ev.ProjectEvent{
			ProjectID: projectID,
			MemberID:  userID,
			Timestamp: time.Now(),
		}
		if err := p.publisher.PublishProjectEvent(ctx, ev.EventTypeProjectMemberAdd, event); err != nil {
			log.Printf("warning: failed to publish project addMember event: %v", err)
		}
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

	if p.publisher != nil {
		event := &ev.ProjectEvent{
			ProjectID: projectID,
			MemberID:  userID,
			Timestamp: time.Now(),
		}
		if err := p.publisher.PublishProjectEvent(ctx, ev.EventTypeProjectMemberDel, event); err != nil {
			log.Printf("warning: failed to publish project delMember event: %v", err)
		}
	}
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

func (p *projectCore) validateProjectMember(ctx context.Context, projectID, userID string) error {
	members, err := p.project.ListProjectMembers(ctx, projectID)
	if err != nil {
		log.Printf("failed to get project members: %v", err)
		return fmt.Errorf("failed to validate project membership")
	}

	for _, member := range members.Members {
		if member.UserID == userID {
			return nil
		}
	}

	return fmt.Errorf("user is not a member of the project")
}
