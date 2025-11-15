package service

import (
	"context"
	"time"

	projectpb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/project_service"
	"google.golang.org/grpc"
)

type ProjectServiceClient struct {
	conn   *grpc.ClientConn
	client projectpb.ProjectServiceClient
}

func NewProjectServiceClient(host, port string) *ProjectServiceClient {
	conn := dialGRPC("project service", host, port, 10*time.Second)
	return &ProjectServiceClient{
		conn:   conn,
		client: projectpb.NewProjectServiceClient(conn),
	}
}

func (c *ProjectServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *ProjectServiceClient) CreateProject(ctx context.Context, req *projectpb.CreateProjectRequest) (*projectpb.Project, error) {
	return c.client.CreateProject(ctx, req)
}

func (c *ProjectServiceClient) GetProject(ctx context.Context, projectID string) (*projectpb.Project, error) {
	return c.client.GetProject(ctx, &projectpb.GetProjectRequest{
		ProjectId: projectID,
	})
}

func (c *ProjectServiceClient) ListProjects(ctx context.Context, req *projectpb.ListProjectsRequest) (*projectpb.ListProjectsResponse, error) {
	return c.client.ListProjects(ctx, req)
}

func (c *ProjectServiceClient) UpdateProject(ctx context.Context, req *projectpb.UpdateProjectRequest) (*projectpb.Project, error) {
	return c.client.UpdateProject(ctx, req)
}

func (c *ProjectServiceClient) DeleteProject(ctx context.Context, projectID string) error {
	_, err := c.client.DeleteProject(ctx, &projectpb.DeleteProjectRequest{
		ProjectId: projectID,
	})
	return err
}

// Task methods
func (c *ProjectServiceClient) CreateTask(ctx context.Context, req *projectpb.CreateTaskRequest) (*projectpb.Task, error) {
	return c.client.CreateTask(ctx, req)
}

func (c *ProjectServiceClient) GetTask(ctx context.Context, taskID string) (*projectpb.Task, error) {
	return c.client.GetTask(ctx, &projectpb.GetTaskRequest{
		TaskId: taskID,
	})
}

func (c *ProjectServiceClient) UpdateTask(ctx context.Context, req *projectpb.UpdateTaskRequest) (*projectpb.Task, error) {
	return c.client.UpdateTask(ctx, req)
}

func (c *ProjectServiceClient) DeleteTask(ctx context.Context, taskID string) error {
	_, err := c.client.DeleteTask(ctx, &projectpb.DeleteTaskRequest{
		TaskId: taskID,
	})
	return err
}

func (c *ProjectServiceClient) MoveTask(ctx context.Context, req *projectpb.MoveTaskRequest) (*projectpb.Task, error) {
	return c.client.MoveTask(ctx, req)
}

func (c *ProjectServiceClient) AssignTask(ctx context.Context, taskID, assigneeID string) (*projectpb.Task, error) {
	return c.client.AssignTask(ctx, &projectpb.AssignTaskRequest{
		TaskId:     taskID,
		AssigneeId: assigneeID,
	})
}

func (c *ProjectServiceClient) ListTasksByProject(ctx context.Context, req *projectpb.ListTasksByProjectRequest) (*projectpb.ListTasksByProjectResponse, error) {
	return c.client.ListTasksByProject(ctx, req)
}

// Project Members methods
func (c *ProjectServiceClient) AddMemberToProject(ctx context.Context, projectID, userID string) error {
	_, err := c.client.AddMemberToProject(ctx, &projectpb.AddMemberToProjectRequest{
		ProjectId: projectID,
		UserId:    userID,
	})
	return err
}

func (c *ProjectServiceClient) RemoveMemberFromProject(ctx context.Context, projectID, userID string) error {
	_, err := c.client.RemoveMemberFromProject(ctx, &projectpb.RemoveMemberFromProjectRequest{
		ProjectId: projectID,
		UserId:    userID,
	})
	return err
}

func (c *ProjectServiceClient) ListProjectMembers(ctx context.Context, projectID string) (*projectpb.ListProjectMembersResponse, error) {
	return c.client.ListProjectMembers(ctx, &projectpb.ListProjectMembersRequest{
		ProjectId: projectID,
	})
}
