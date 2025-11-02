package client

import (
	"context"
	"log"

	projectv1 "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/project_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProjectClient struct {
	client projectv1.ProjectServiceClient
	conn   *grpc.ClientConn
}

func NewProjectClient(addr string) (*ProjectClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to dial project service: %v", err)
		return nil, err
	}

	client := projectv1.NewProjectServiceClient(conn)

	return &ProjectClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *ProjectClient) GetProject(ctx context.Context, projectID string) (*projectv1.Project, error) {
	req := &projectv1.GetProjectRequest{
		ProjectId: projectID,
	}

	project, err := c.client.GetProject(ctx, req)
	if err != nil {
		log.Printf("failed to get project: %v", err)
		return nil, err
	}

	return project, nil
}

func (c *ProjectClient) ListProjects(ctx context.Context, pageSize, pageNumber int32, managerID string, status *projectv1.ProjectStatus) ([]*projectv1.Project, error) {
	req := &projectv1.ListProjectsRequest{
		PageSize:   pageSize,
		PageNumber: pageNumber,
		ManagerId:  &managerID,
		Status:     status,
	}

	resp, err := c.client.ListProjects(ctx, req)
	if err != nil {
		log.Printf("failed to list projects: %v", err)
		return nil, err
	}

	return resp.Projects, nil
}

func (c *ProjectClient) ListTasksByProject(ctx context.Context, projectID string) ([]*projectv1.Task, error) {
	req := &projectv1.ListTasksByProjectRequest{
		ProjectId: projectID,
	}

	resp, err := c.client.ListTasksByProject(ctx, req)
	if err != nil {
		log.Printf("failed to list tasks by project: %v", err)
		return nil, err
	}

	return resp.Tasks, nil
}

func (c *ProjectClient) GetTaskStatusHistory(ctx context.Context, taskID string) ([]*projectv1.TaskStatusHistoryEntry, error) {
	req := &projectv1.GetTaskStatusHistoryRequest{
		TaskId: taskID,
	}

	resp, err := c.client.GetTaskStatusHistory(ctx, req)
	if err != nil {
		log.Printf("failed to get task status history: %v", err)
		return nil, err
	}

	return resp.History, nil
}

func (c *ProjectClient) GetProjectMetrics(ctx context.Context, projectID string) (*projectv1.ProjectMetrics, error) {
	req := &projectv1.GetProjectMetricsRequest{
		ProjectId: projectID,
	}

	metrics, err := c.client.GetProjectMetrics(ctx, req)
	if err != nil {
		log.Printf("failed to get project metrics: %v", err)
		return nil, err
	}

	return metrics, nil
}

func (c *ProjectClient) ListProjectMembers(ctx context.Context, projectID string) ([]*projectv1.ProjectMember, error) {
	req := &projectv1.ListProjectMembersRequest{
		ProjectId: projectID,
	}

	resp, err := c.client.ListProjectMembers(ctx, req)
	if err != nil {
		log.Printf("failed to list project members: %v", err)
		return nil, err
	}

	return resp.Members, nil
}

func (c *ProjectClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
