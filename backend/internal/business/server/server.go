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
	if req.MemberId != nil {
		filter.MemberID = *req.MemberId
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
	// If moving TO review, use SubmitForReview flow
	if req.NewStatus == business.TaskStatus_REVIEW {
		task, err := s.taskService.SubmitForReview(ctx, req.TaskId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to submit for review: %v", err)
		}
		return taskToProto(task), nil
	}

	// Check if task is currently in review - block moving
	existingTask, err := s.taskService.GetTask(ctx, req.TaskId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found: %v", err)
	}
	if existingTask.Status == "REVIEW" {
		return nil, status.Errorf(codes.FailedPrecondition, "task is in review and cannot be moved manually")
	}

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

func (s *BusinessServer) GetEmployeeMetrics(ctx context.Context, req *business.GetEmployeeMetricsRequest) (*business.EmployeeMetricsResponse, error) {
	log.Printf("GetEmployeeMetrics called for employee_id: %d", req.EmployeeId)

	analytics, err := s.analyticsService.GetEmployeeAnalytics(ctx, req.EmployeeId)
	if err != nil {
		log.Printf("GetEmployeeAnalytics error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get employee metrics: %v", err)
	}

	log.Printf("Analytics for user %d: assigned=%d, completed=%d, on_time=%d, late=%d, weighted_on_time=%.1f, weighted_total=%.1f",
		analytics.UserID, analytics.AssignedTasks, analytics.CompletedTasks, analytics.CompletedOnTime, analytics.CompletedLate,
		analytics.WeightedOnTime, analytics.WeightedTotal)

	var completionRate float64
	if analytics.AssignedTasks > 0 {
		completionRate = float64(analytics.CompletedTasks) / float64(analytics.AssignedTasks) * 100
	}

	var onTimeRate float64
	if analytics.CompletedTasks > 0 {
		onTimeRate = float64(analytics.CompletedOnTime) / float64(analytics.CompletedTasks) * 100
	} else {
		onTimeRate = 0 // Нет завершённых задач - нечего оценивать
	}

	var weightedOnTimeRate float64
	if analytics.WeightedTotal > 0 {
		weightedOnTimeRate = analytics.WeightedOnTime / analytics.WeightedTotal * 100
	} else {
		weightedOnTimeRate = 0 // Нет завершённых задач - эффективность 0
	}

	var efficiency float64
	if analytics.CompletedTasks > 0 && analytics.AssignedTasks > 0 {
		efficiency = completionRate * weightedOnTimeRate / 100
	} else {
		efficiency = 0 // Без выполненных задач эффективность 0
	}

	log.Printf("Calculated rates: completion=%.2f%%, on_time=%.2f%%, weighted_on_time=%.2f%%, efficiency=%.2f%%",
		completionRate, onTimeRate, weightedOnTimeRate, efficiency)

	return &business.EmployeeMetricsResponse{
		Metrics: &business.EmployeeMetrics{
			EmployeeId:      analytics.UserID,
			AssignedTasks:   analytics.AssignedTasks,
			CompletedTasks:  analytics.CompletedTasks,
			CompletedOnTime: analytics.CompletedOnTime,
			CompletedLate:   analytics.CompletedLate,
			InProgressTasks: analytics.InProgressTasks,
			OverdueTasks:    analytics.OverdueTasks,
			CompletionRate:  efficiency,         // Взвешенная эффективность
			OnTimeRate:      weightedOnTimeRate, // Взвешенный % вовремя
		},
	}, nil
}

func (s *BusinessServer) ListEmployeeMetrics(ctx context.Context, req *business.ListEmployeeMetricsRequest) (*business.ListEmployeeMetricsResponse, error) {
	all, err := s.analyticsService.GetAllEmployeeAnalytics(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list employee metrics: %v", err)
	}

	var metrics []*business.EmployeeMetrics
	for _, a := range all {
		var completionRate, onTimeRate float64
		if a.AssignedTasks > 0 {
			completionRate = float64(a.CompletedTasks) / float64(a.AssignedTasks) * 100
		}
		if a.CompletedTasks > 0 {
			onTimeRate = float64(a.CompletedOnTime) / float64(a.CompletedTasks) * 100
		} else {
			onTimeRate = 100
		}
		metrics = append(metrics, &business.EmployeeMetrics{
			EmployeeId:      a.UserID,
			AssignedTasks:   a.AssignedTasks,
			CompletedTasks:  a.CompletedTasks,
			CompletedOnTime: a.CompletedOnTime,
			CompletedLate:   a.CompletedLate,
			InProgressTasks: a.InProgressTasks,
			OverdueTasks:    a.OverdueTasks,
			CompletionRate:  completionRate,
			OnTimeRate:      onTimeRate,
		})
	}

	return &business.ListEmployeeMetricsResponse{
		Metrics:    metrics,
		TotalCount: int32(len(metrics)),
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

	var onTimeRate float64
	if analytics.CompletedTasks > 0 {
		onTimeRate = float64(analytics.CompletedOnTime) / float64(analytics.CompletedTasks) * 100
	} else {
		onTimeRate = 100 // Если завершённых задач нет, считаем что все в срок
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
			CompletedOnTime: analytics.CompletedOnTime,
			CompletedLate:   analytics.CompletedLate,
			InProgressTasks: analytics.InProgressTasks,
			OverdueTasks:    analytics.OverdueTasks,
			TeamSize:        analytics.MemberCount,
			ProgressPercent: progressPercent,
			OnTimeRate:      onTimeRate,
			HealthStatus:    healthStatus,
		},
		CalculatedAt: timestamppb.Now(),
	}, nil
}

func (s *BusinessServer) ListProjectMetrics(ctx context.Context, req *business.ListProjectMetricsRequest) (*business.ListProjectMetricsResponse, error) {
	all, err := s.analyticsService.GetAllProjectAnalytics(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list project metrics: %v", err)
	}

	var metrics []*business.ProjectMetrics
	for _, a := range all {
		var progressPercent, onTimeRate float64
		if a.TotalTasks > 0 {
			progressPercent = float64(a.CompletedTasks) / float64(a.TotalTasks) * 100
		}
		if a.CompletedTasks > 0 {
			onTimeRate = float64(a.CompletedOnTime) / float64(a.CompletedTasks) * 100
		} else {
			onTimeRate = 100
		}

		healthStatus := business.HealthStatus_HEALTH_STATUS_HEALTHY
		if a.OverdueTasks > a.TotalTasks/2 {
			healthStatus = business.HealthStatus_HEALTH_STATUS_CRITICAL
		} else if a.OverdueTasks > 0 {
			healthStatus = business.HealthStatus_HEALTH_STATUS_AT_RISK
		}

		metrics = append(metrics, &business.ProjectMetrics{
			ProjectId:       a.ProjectID,
			TotalTasks:      a.TotalTasks,
			CompletedTasks:  a.CompletedTasks,
			CompletedOnTime: a.CompletedOnTime,
			CompletedLate:   a.CompletedLate,
			InProgressTasks: a.InProgressTasks,
			OverdueTasks:    a.OverdueTasks,
			TeamSize:        a.MemberCount,
			ProgressPercent: progressPercent,
			OnTimeRate:      onTimeRate,
			HealthStatus:    healthStatus,
		})
	}

	return &business.ListProjectMetricsResponse{
		Metrics:    metrics,
		TotalCount: int32(len(metrics)),
	}, nil
}

func (s *BusinessServer) GetProductivityTrends(ctx context.Context, req *business.GetProductivityTrendsRequest) (*business.ProductivityTrendsResponse, error) {
	days := req.Limit
	if days <= 0 {
		days = 30
	}

	points, err := s.analyticsService.GetProductivityTrends(ctx, days)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get productivity trends: %v", err)
	}

	var entries []*business.ProductivityTrendEntry
	for _, p := range points {
		entries = append(entries, &business.ProductivityTrendEntry{
			Date:              timestamppb.New(p.Date),
			TasksCompleted:    p.TasksCompleted,
			AvgCompletionRate: p.AvgCompletionRate,
		})
	}

	return &business.ProductivityTrendsResponse{
		Entries: entries,
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

	var avgOnTimeRate float64
	if summary.CompletedTasks > 0 {
		avgOnTimeRate = float64(summary.CompletedOnTime) / float64(summary.CompletedTasks) * 100
	} else {
		avgOnTimeRate = 100 // Если завершённых задач нет, считаем что все в срок
	}

	var avgEfficiency float64
	if summary.TotalTasks > 0 {
		avgEfficiency = avgCompletionRate * avgOnTimeRate / 100
	} else {
		avgEfficiency = 100 // Без задач считаем 100%
	}

	return &business.DashboardStatsResponse{
		TotalEmployees:      summary.TotalEmployees,
		ActiveEmployees:     summary.ActiveEmployees,
		TotalProjects:       summary.TotalProjects,
		ActiveProjects:      summary.ActiveProjects,
		TotalTasks:          summary.TotalTasks,
		CompletedTasks:      summary.CompletedTasks,
		OverdueTasks:        summary.OverdueTasks,
		CompletedOnTime:     summary.CompletedOnTime,
		CompletedLate:       summary.CompletedLate,
		AvgCompletionRate:   avgEfficiency, // Теперь это комбинированная эффективность
		AvgOnTimeRate:       avgOnTimeRate,
		TopEmployees:        []*business.TopEmployee{},
		ProblematicProjects: []*business.ProblematicProject{},
		CalculatedAt:        timestamppb.Now(),
	}, nil
}

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

func protoStatusToString(s business.ProjectStatus) string {
	switch s {
	case business.ProjectStatus_ACTIVE:
		return "ACTIVE"
	case business.ProjectStatus_ON_HOLD:
		return "ON_HOLD"
	case business.ProjectStatus_ARCHIVED:
		return "ARCHIVED"
	default:
		return "ACTIVE"
	}
}

func stringToProtoStatus(s string) business.ProjectStatus {
	switch s {
	case "ACTIVE":
		return business.ProjectStatus_ACTIVE
	case "ON_HOLD":
		return business.ProjectStatus_ON_HOLD
	case "ARCHIVED":
		return business.ProjectStatus_ARCHIVED
	case "COMPLETED":
		return business.ProjectStatus_ARCHIVED
	default:
		return business.ProjectStatus_PROJECT_STATUS_UNSPECIFIED
	}
}

func protoTaskStatusToString(s business.TaskStatus) string {
	switch s {
	case business.TaskStatus_TODO:
		return "TODO"
	case business.TaskStatus_IN_PROGRESS:
		return "IN_PROGRESS"
	case business.TaskStatus_REVIEW:
		return "REVIEW"
	case business.TaskStatus_DONE:
		return "DONE"
	default:
		return "TODO"
	}
}

func stringToProtoTaskStatus(s string) business.TaskStatus {
	switch s {
	case "TODO":
		return business.TaskStatus_TODO
	case "IN_PROGRESS":
		return business.TaskStatus_IN_PROGRESS
	case "REVIEW", "IN_REVIEW":
		return business.TaskStatus_REVIEW
	case "DONE":
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

// Review RPCs

func (s *BusinessServer) SubmitForReview(ctx context.Context, req *business.SubmitForReviewRequest) (*business.Task, error) {
	task, err := s.taskService.SubmitForReview(ctx, req.TaskId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to submit for review: %v", err)
	}
	return taskToProto(task), nil
}

func (s *BusinessServer) ApproveTask(ctx context.Context, req *business.ApproveTaskRequest) (*business.Task, error) {
	task, err := s.taskService.ApproveTask(ctx, req.TaskId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to approve task: %v", err)
	}
	return taskToProto(task), nil
}

func (s *BusinessServer) GetReviewStatus(ctx context.Context, req *business.GetReviewStatusRequest) (*business.ReviewStatusResponse, error) {
	reviewStatus, err := s.taskService.GetReviewStatus(ctx, req.TaskId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get review status: %v", err)
	}

	reviewers := make([]*business.ReviewerInfo, len(reviewStatus.Reviewers))
	for i, r := range reviewStatus.Reviewers {
		reviewers[i] = &business.ReviewerInfo{
			UserId:   r.UserID,
			Approved: r.Approved,
		}
	}

	return &business.ReviewStatusResponse{
		Reviewers: reviewers,
		IsActive:  reviewStatus.IsActive,
	}, nil
}
