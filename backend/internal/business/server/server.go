package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/business"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/business/repo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type BusinessServer struct {
	business.UnimplementedBusinessServiceServer
	projectService   *core.ProjectService
	taskService      *core.TaskService
	analyticsService *core.AnalyticsService
}

func NewBusinessServer(projectService *core.ProjectService, taskService *core.TaskService, analyticsService *core.AnalyticsService) *BusinessServer {
	return &BusinessServer{
		projectService:   projectService,
		taskService:      taskService,
		analyticsService: analyticsService,
	}
}

func (s *BusinessServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	business.RegisterBusinessServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	log.Printf("Business gRPC server listening on %s", addr)
	return grpcServer.Serve(lis)
}

// === PROJECT ===

func (s *BusinessServer) CreateProject(ctx context.Context, req *business.CreateProjectRequest) (*business.Project, error) {
	createReq := &core.CreateProjectRequest{
		Name:    req.Name,
		OwnerID: req.ManagerId,
	}
	if req.Description != nil {
		createReq.Description = *req.Description
	}
	if req.DueDate != nil {
		t := req.DueDate.AsTime()
		createReq.EndDate = &t
	}

	project, err := s.projectService.CreateProject(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create project: %v", err)
	}

	return projectToProto(project), nil
}

func (s *BusinessServer) GetProject(ctx context.Context, req *business.GetProjectRequest) (*business.Project, error) {
	project, err := s.projectService.GetProject(ctx, req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "project not found: %v", err)
	}
	return projectToProto(project), nil
}

func (s *BusinessServer) ListProjects(ctx context.Context, req *business.ListProjectsRequest) (*business.ListProjectsResponse, error) {
	filter := &core.ListProjectsFilter{
		PageSize:   int(req.PageSize),
		PageNumber: int(req.PageNumber),
	}
	if req.ManagerId != nil {
		filter.OwnerID = *req.ManagerId
	}
	if req.Status != nil {
		filter.Status = protoStatusToString(*req.Status)
	}

	projects, total, err := s.projectService.ListProjects(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list projects: %v", err)
	}

	protoProjects := make([]*business.Project, len(projects))
	for i, p := range projects {
		protoProjects[i] = projectToProto(p)
	}

	return &business.ListProjectsResponse{
		Projects:   protoProjects,
		TotalCount: int32(total),
	}, nil
}

func (s *BusinessServer) UpdateProject(ctx context.Context, req *business.UpdateProjectRequest) (*business.Project, error) {
	updateReq := &core.UpdateProjectRequest{}
	if req.Name != nil {
		updateReq.Name = req.Name
	}
	if req.Description != nil {
		updateReq.Description = req.Description
	}
	if req.Status != nil {
		st := protoStatusToString(*req.Status)
		updateReq.Status = &st
	}
	if req.DueDate != nil {
		t := req.DueDate.AsTime()
		updateReq.EndDate = &t
	}

	project, err := s.projectService.UpdateProject(ctx, req.ProjectId, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update project: %v", err)
	}

	return projectToProto(project), nil
}

func (s *BusinessServer) DeleteProject(ctx context.Context, req *business.DeleteProjectRequest) (*emptypb.Empty, error) {
	if err := s.projectService.DeleteProject(ctx, req.ProjectId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete project: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// === TASK ===

func (s *BusinessServer) CreateTask(ctx context.Context, req *business.CreateTaskRequest) (*business.Task, error) {
	createReq := &core.CreateTaskRequest{
		ProjectID:   req.ProjectId,
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   req.CreatorId,
	}
	if req.AssigneeId != nil {
		createReq.AssigneeID = req.AssigneeId
	}
	if req.Priority != nil {
		p := protoPriorityToString(*req.Priority)
		createReq.Priority = p
	}
	if req.DueDate != nil {
		t := req.DueDate.AsTime()
		createReq.DueDate = &t
	}

	task, err := s.taskService.CreateTask(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create task: %v", err)
	}

	return taskToProto(task), nil
}

func (s *BusinessServer) GetTask(ctx context.Context, req *business.GetTaskRequest) (*business.Task, error) {
	task, err := s.taskService.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found: %v", err)
	}
	return taskToProto(task), nil
}

func (s *BusinessServer) UpdateTask(ctx context.Context, req *business.UpdateTaskRequest) (*business.Task, error) {
	updateReq := &core.UpdateTaskRequest{}
	if req.Title != nil {
		updateReq.Title = req.Title
	}
	if req.Description != nil {
		updateReq.Description = req.Description
	}
	if req.Status != nil {
		st := protoTaskStatusToString(*req.Status)
		updateReq.Status = &st
	}
	if req.Priority != nil {
		p := protoPriorityToString(*req.Priority)
		updateReq.Priority = &p
	}
	if req.AssigneeId != nil {
		updateReq.AssigneeID = req.AssigneeId
	}
	if req.DueDate != nil {
		t := req.DueDate.AsTime()
		updateReq.DueDate = &t
	}

	task, err := s.taskService.UpdateTask(ctx, req.TaskId, 0, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}

	return taskToProto(task), nil
}

func (s *BusinessServer) DeleteTask(ctx context.Context, req *business.DeleteTaskRequest) (*emptypb.Empty, error) {
	if err := s.taskService.DeleteTask(ctx, req.TaskId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete task: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *BusinessServer) MoveTask(ctx context.Context, req *business.MoveTaskRequest) (*business.Task, error) {
	st := protoTaskStatusToString(req.NewStatus)
	updateReq := &core.UpdateTaskRequest{
		Status: &st,
	}

	task, err := s.taskService.UpdateTask(ctx, req.TaskId, 0, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to move task: %v", err)
	}
	return taskToProto(task), nil
}

func (s *BusinessServer) AssignTask(ctx context.Context, req *business.AssignTaskRequest) (*business.Task, error) {
	updateReq := &core.UpdateTaskRequest{
		AssigneeID: &req.AssigneeId,
	}

	task, err := s.taskService.UpdateTask(ctx, req.TaskId, 0, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign task: %v", err)
	}
	return taskToProto(task), nil
}

func (s *BusinessServer) ListTasksByProject(ctx context.Context, req *business.ListTasksByProjectRequest) (*business.ListTasksByProjectResponse, error) {
	filter := &core.ListTasksFilter{
		ProjectID: req.ProjectId,
	}
	if req.Status != nil {
		filter.Status = protoTaskStatusToString(*req.Status)
	}
	if req.AssigneeId != nil {
		filter.AssigneeID = *req.AssigneeId
	}
	if req.Priority != nil {
		filter.Priority = protoPriorityToString(*req.Priority)
	}

	tasks, _, err := s.taskService.ListTasks(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
	}

	protoTasks := make([]*business.Task, len(tasks))
	for i, t := range tasks {
		protoTasks[i] = taskToProto(t)
	}

	return &business.ListTasksByProjectResponse{
		Tasks: protoTasks,
	}, nil
}

// === PROJECT MEMBERS ===

func (s *BusinessServer) AddMemberToProject(ctx context.Context, req *business.AddMemberToProjectRequest) (*emptypb.Empty, error) {
	if err := s.projectService.AddMember(ctx, req.ProjectId, req.UserId, "member"); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add member: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *BusinessServer) RemoveMemberFromProject(ctx context.Context, req *business.RemoveMemberFromProjectRequest) (*emptypb.Empty, error) {
	if err := s.projectService.RemoveMember(ctx, req.ProjectId, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove member: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *BusinessServer) ListProjectMembers(ctx context.Context, req *business.ListProjectMembersRequest) (*business.ListProjectMembersResponse, error) {
	members, err := s.projectService.GetMembers(ctx, req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list members: %v", err)
	}

	protoMembers := make([]*business.ProjectMember, len(members))
	for i, m := range members {
		protoMembers[i] = &business.ProjectMember{
			UserId:   m.UserID,
			FullName: m.UserName,
			Role:     m.Role,
		}
	}

	return &business.ListProjectMembersResponse{
		Members: protoMembers,
	}, nil
}

// === ANALYTICS ===

func (s *BusinessServer) GetEmployeeMetrics(ctx context.Context, req *business.GetEmployeeMetricsRequest) (*business.EmployeeMetricsResponse, error) {
	analytics, err := s.analyticsService.GetEmployeeAnalytics(ctx, req.EmployeeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get employee metrics: %v", err)
	}

	var completionRate float64
	if analytics.AssignedTasks > 0 {
		completionRate = float64(analytics.CompletedTasks) / float64(analytics.AssignedTasks) * 100
	}

	return &business.EmployeeMetricsResponse{
		Metrics: &business.EmployeeMetrics{
			EmployeeId:      analytics.UserID,
			AssignedTasks:   analytics.AssignedTasks,
			CompletedTasks:  analytics.CompletedTasks,
			InProgressTasks: analytics.InProgressTasks,
			OverdueTasks:    analytics.OverdueTasks,
			CompletionRate:  completionRate,
			OnTimeRate:      100 - float64(analytics.OverdueTasks)/float64(max(analytics.CompletedTasks, 1))*100,
		},
	}, nil
}

func (s *BusinessServer) ListEmployeeMetrics(ctx context.Context, req *business.ListEmployeeMetricsRequest) (*business.ListEmployeeMetricsResponse, error) {
	return &business.ListEmployeeMetricsResponse{
		Metrics:    []*business.EmployeeMetrics{},
		TotalCount: 0,
	}, nil
}

func (s *BusinessServer) GetTopPerformers(ctx context.Context, req *business.GetTopPerformersRequest) (*business.ListEmployeeMetricsResponse, error) {
	return &business.ListEmployeeMetricsResponse{
		Metrics:    []*business.EmployeeMetrics{},
		TotalCount: 0,
	}, nil
}

func (s *BusinessServer) GetProjectMetrics(ctx context.Context, req *business.GetProjectMetricsRequest) (*business.ProjectMetricsResponse, error) {
	analytics, err := s.analyticsService.GetProjectAnalytics(ctx, req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get project metrics: %v", err)
	}

	var progressPercent float64
	if analytics.TotalTasks > 0 {
		progressPercent = float64(analytics.CompletedTasks) / float64(analytics.TotalTasks) * 100
	}

	healthStatus := business.HealthStatus_HEALTH_STATUS_HEALTHY
	if analytics.OverdueTasks > int32(analytics.TotalTasks/2) {
		healthStatus = business.HealthStatus_HEALTH_STATUS_CRITICAL
	} else if analytics.OverdueTasks > 0 {
		healthStatus = business.HealthStatus_HEALTH_STATUS_AT_RISK
	}

	return &business.ProjectMetricsResponse{
		Metrics: &business.ProjectMetrics{
			ProjectId:       analytics.ProjectID,
			ManagerId:       0,
			TotalTasks:      analytics.TotalTasks,
			CompletedTasks:  analytics.CompletedTasks,
			InProgressTasks: analytics.InProgressTasks,
			OverdueTasks:    analytics.OverdueTasks,
			TeamSize:        analytics.MemberCount,
			ProgressPercent: progressPercent,
			OnTimeRate:      100 - float64(analytics.OverdueTasks)/float64(max(int32(1), analytics.CompletedTasks))*100,
			HealthStatus:    healthStatus,
		},
		CalculatedAt: timestamppb.Now(),
	}, nil
}

func (s *BusinessServer) ListProjectMetrics(ctx context.Context, req *business.ListProjectMetricsRequest) (*business.ListProjectMetricsResponse, error) {
	return &business.ListProjectMetricsResponse{
		Metrics:    []*business.ProjectMetrics{},
		TotalCount: 0,
	}, nil
}

func (s *BusinessServer) GetProductivityTrends(ctx context.Context, req *business.GetProductivityTrendsRequest) (*business.ProductivityTrendsResponse, error) {
	return &business.ProductivityTrendsResponse{
		Entries: []*business.ProductivityTrendEntry{},
		Period:  req.Period,
	}, nil
}

func (s *BusinessServer) GetCompletionRateTrends(ctx context.Context, req *business.GetCompletionRateTrendsRequest) (*business.CompletionRateTrendsResponse, error) {
	return &business.CompletionRateTrendsResponse{
		Entries: []*business.CompletionRateTrendEntry{},
		Period:  req.Period,
	}, nil
}

func (s *BusinessServer) GetDashboardStats(ctx context.Context, req *business.GetDashboardStatsRequest) (*business.DashboardStatsResponse, error) {
	summary, err := s.analyticsService.GetSummary(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get dashboard stats: %v", err)
	}

	var avgCompletionRate float64
	if summary.TotalTasks > 0 {
		avgCompletionRate = float64(summary.CompletedTasks) / float64(summary.TotalTasks) * 100
	}

	return &business.DashboardStatsResponse{
		TotalEmployees:      summary.TotalEmployees,
		ActiveEmployees:     summary.ActiveEmployees,
		TotalProjects:       summary.TotalProjects,
		ActiveProjects:      summary.ActiveProjects,
		TotalTasks:          summary.TotalTasks,
		CompletedTasks:      summary.CompletedTasks,
		OverdueTasks:        summary.OverdueTasks,
		AvgCompletionRate:   avgCompletionRate,
		AvgOnTimeRate:       100 - float64(summary.OverdueTasks)/float64(max(int32(1), summary.CompletedTasks))*100,
		TopEmployees:        []*business.TopEmployee{},
		ProblematicProjects: []*business.ProblematicProject{},
		CalculatedAt:        timestamppb.Now(),
	}, nil
}

// === CONVERTERS ===

func projectToProto(p *repo.Project) *business.Project {
	project := &business.Project{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		ManagerId:   p.OwnerID,
		Status:      stringToProtoStatus(p.Status),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
	if p.EndDate != nil {
		project.DueDate = timestamppb.New(*p.EndDate)
	}
	return project
}

func taskToProto(t *repo.Task) *business.Task {
	task := &business.Task{
		Id:          t.ID,
		ProjectId:   t.ProjectID,
		Title:       t.Title,
		Description: t.Description,
		Status:      stringToProtoTaskStatus(t.Status),
		Priority:    stringToProtoPriority(t.Priority),
		CreatorId:   t.CreatorID,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
	if t.AssigneeID != nil {
		task.AssigneeId = *t.AssigneeID
	}
	if t.DueDate != nil {
		task.DueDate = timestamppb.New(*t.DueDate)
	}
	return task
}

// Status converters
func protoStatusToString(s business.ProjectStatus) string {
	switch s {
	case business.ProjectStatus_ACTIVE:
		return "active"
	case business.ProjectStatus_ON_HOLD:
		return "on_hold"
	case business.ProjectStatus_ARCHIVED:
		return "archived"
	default:
		return "active"
	}
}

func stringToProtoStatus(s string) business.ProjectStatus {
	switch s {
	case "active":
		return business.ProjectStatus_ACTIVE
	case "on_hold":
		return business.ProjectStatus_ON_HOLD
	case "archived":
		return business.ProjectStatus_ARCHIVED
	default:
		return business.ProjectStatus_PROJECT_STATUS_UNSPECIFIED
	}
}

func protoTaskStatusToString(s business.TaskStatus) string {
	switch s {
	case business.TaskStatus_TODO:
		return "todo"
	case business.TaskStatus_IN_PROGRESS:
		return "in_progress"
	case business.TaskStatus_REVIEW:
		return "review"
	case business.TaskStatus_DONE:
		return "done"
	default:
		return "todo"
	}
}

func stringToProtoTaskStatus(s string) business.TaskStatus {
	switch s {
	case "todo":
		return business.TaskStatus_TODO
	case "in_progress":
		return business.TaskStatus_IN_PROGRESS
	case "review":
		return business.TaskStatus_REVIEW
	case "done":
		return business.TaskStatus_DONE
	default:
		return business.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

func protoPriorityToString(p business.TaskPriority) string {
	switch p {
	case business.TaskPriority_LOW:
		return "low"
	case business.TaskPriority_MEDIUM:
		return "medium"
	case business.TaskPriority_HIGH:
		return "high"
	case business.TaskPriority_CRITICAL:
		return "critical"
	default:
		return "medium"
	}
}

func stringToProtoPriority(s string) business.TaskPriority {
	switch s {
	case "low":
		return business.TaskPriority_LOW
	case "medium":
		return business.TaskPriority_MEDIUM
	case "high":
		return business.TaskPriority_HIGH
	case "critical":
		return business.TaskPriority_CRITICAL
	default:
		return business.TaskPriority_PRIORITY_UNSPECIFIED
	}
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
