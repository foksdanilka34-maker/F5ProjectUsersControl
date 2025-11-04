package service

import (
	"context"
	"fmt"
	"log"
	"time"

	projectpb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/project_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProjectServiceClient struct {
	conn   *grpc.ClientConn
	client projectpb.ProjectServiceClient
}

func NewProjectServiceClient(host, port string) *ProjectServiceClient {
	address := fmt.Sprintf("%s:%s", host, port)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("Failed to connect to project service at %s: %v", address, err)
	}

	log.Printf("Successfully connected to project service at %s", address)

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

// Project methods
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

// History and Metrics
func (c *ProjectServiceClient) GetTaskStatusHistory(ctx context.Context, req *projectpb.GetTaskStatusHistoryRequest) (*projectpb.GetTaskStatusHistoryResponse, error) {
	return c.client.GetTaskStatusHistory(ctx, req)
}

func (c *ProjectServiceClient) GetProjectMetrics(ctx context.Context, req *projectpb.GetProjectMetricsRequest) (*projectpb.ProjectMetrics, error) {
	return c.client.GetProjectMetrics(ctx, req)
}
