package server

import (
	"context"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"
	core "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/core"
	pb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/project_service"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/storage"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedProjectServiceServer
	core core.CoreLogic
}

func NewProjectServer(core core.CoreLogic) *Server {
	return &Server{
		core: core,
	}
}

func (s *Server) Register(gRPC *grpc.Server) {
	pb.RegisterProjectServiceServer(gRPC, s)
}

// ============ Project Methods ============

func (s *Server) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.Project, error) {
	storage.Logger.Info("RPC CreateProject called", zap.String("name", req.Name), zap.String("managerID", req.ManagerId))
	if req.GetName() == "" || req.GetManagerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "name and manager_id are required")
	}

	coreReq := &models.CreateProjectRequest{
		Name:        req.Name,
		Description: req.Description,
		ManagerID:   req.ManagerId,
		Status:      models.ProjectStatus(req.Status),
	}
	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		coreReq.DueDate = &dueDate
	}

	project, err := s.core.CreateProject(ctx, coreReq)
	if err != nil {
		storage.Logger.Error("failed to create project", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create project")
	}

	return convertProjectToProto(project), nil
}

func (s *Server) GetProject(ctx context.Context, req *pb.GetProjectRequest) (*pb.Project, error) {
	storage.Logger.Info("RPC GetProject called", zap.String("projectID", req.ProjectId))
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	project, err := s.core.GetProject(ctx, req.ProjectId)
	if err != nil {
		storage.Logger.Error("failed to get project", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get project")
	}

	return convertProjectToProto(project), nil
}

func (s *Server) ListProjects(ctx context.Context, req *pb.ListProjectsRequest) (*pb.ListProjectsResponse, error) {
	storage.Logger.Info("RPC ListProjects called", zap.Int32("pageSize", req.PageSize), zap.Int32("pageNumber", req.PageNumber))

	filter := &models.ListProjectsFilter{
		PageSize:   req.PageSize,
		PageNumber: req.PageNumber,
	}
	if req.ManagerId != nil {
		filter.ManagerID = req.ManagerId
	}
	if req.Status != nil {
		status := models.ProjectStatus(*req.Status)
		filter.Status = &status
	}

	response, err := s.core.ListProjects(ctx, filter)
	if err != nil {
		storage.Logger.Error("failed to list projects", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list projects")
	}

	protoProjects := make([]*pb.Project, len(response.Projects))
	for i, p := range response.Projects {
		protoProjects[i] = convertProjectToProto(p)
	}

	return &pb.ListProjectsResponse{
		Projects:   protoProjects,
		TotalCount: response.TotalCount,
	}, nil
}

func (s *Server) UpdateProject(ctx context.Context, req *pb.UpdateProjectRequest) (*pb.Project, error) {
	storage.Logger.Info("RPC UpdateProject called", zap.String("projectID", req.ProjectId))
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	coreReq := &models.UpdateProjectRequest{
		ID: req.ProjectId,
	}
	if req.Name != nil {
		coreReq.Name = req.Name
	}
	if req.Description != nil {
		coreReq.Description = req.Description
	}
	if req.Status != nil {
		status := models.ProjectStatus(*req.Status)
		coreReq.Status = &status
	}
	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		coreReq.DueDate = &dueDate
	}

	project, err := s.core.UpdateProject(ctx, coreReq)
	if err != nil {
		storage.Logger.Error("failed to update project", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update project")
	}

	return convertProjectToProto(project), nil
}

func (s *Server) DeleteProject(ctx context.Context, req *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	storage.Logger.Info("RPC DeleteProject called", zap.String("projectID", req.ProjectId))
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	err := s.core.DeleteProject(ctx, req.ProjectId)
	if err != nil {
		storage.Logger.Error("failed to delete project", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete project")
	}

	return &pb.DeleteProjectResponse{Success: true}, nil
}

// ============ Task Methods ============

func (s *Server) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	storage.Logger.Info("RPC CreateTask called", zap.String("projectID", req.ProjectId), zap.String("title", req.Title))
	if req.GetProjectId() == "" || req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id and title are required")
	}

	coreReq := &models.CreateTaskRequest{
		ProjectID:   req.ProjectId,
		Title:       req.Title,
		Description: req.Description,
		Status:      models.TaskStatus(req.Status),
		Priority:    models.TaskPriority(req.Priority),
		CreatorID:   req.CreatorId,
	}
	if req.AssigneeId != nil {
		coreReq.AssigneeID = req.AssigneeId
	}
	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		coreReq.DueDate = &dueDate
	}

	task, err := s.core.CreateTask(ctx, coreReq)
	if err != nil {
		storage.Logger.Error("failed to create task", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create task")
	}

	return convertTaskToProto(task), nil
}

func (s *Server) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	storage.Logger.Info("RPC GetTask called", zap.String("taskID", req.TaskId))
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	task, err := s.core.GetTask(ctx, req.TaskId)
	if err != nil {
		storage.Logger.Error("failed to get task", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get task")
	}

	return convertTaskToProto(task), nil
}

func (s *Server) ListTasksByProject(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	storage.Logger.Info("RPC ListTasksByProject called", zap.String("projectID", req.ProjectId))
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	filter := &models.ListTasksFilter{
		ProjectID: req.ProjectId,
	}
	if req.Status != nil {
		status := models.TaskStatus(*req.Status)
		filter.Status = &status
	}
	if req.AssigneeId != nil {
		filter.AssigneeID = req.AssigneeId
	}

	tasks, err := s.core.ListTasksByProject(ctx, filter)
	if err != nil {
		storage.Logger.Error("failed to list tasks", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list tasks")
	}

	protoTasks := make([]*pb.Task, len(tasks))
	for i, t := range tasks {
		protoTasks[i] = convertTaskToProto(t)
	}

	return &pb.ListTasksResponse{Tasks: protoTasks}, nil
}

func (s *Server) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
	storage.Logger.Info("RPC UpdateTask called", zap.String("taskID", req.TaskId))
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	coreReq := &models.UpdateTaskRequest{
		ID: req.TaskId,
	}
	if req.Title != nil {
		coreReq.Title = req.Title
	}
	if req.Description != nil {
		coreReq.Description = req.Description
	}
	if req.Priority != nil {
		priority := models.TaskPriority(*req.Priority)
		coreReq.Priority = &priority
	}
	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		coreReq.DueDate = &dueDate
	}

	task, err := s.core.UpdateTask(ctx, coreReq)
	if err != nil {
		storage.Logger.Error("failed to update task", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update task")
	}

	return convertTaskToProto(task), nil
}

func (s *Server) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	storage.Logger.Info("RPC DeleteTask called", zap.String("taskID", req.TaskId))
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	err := s.core.DeleteTask(ctx, req.TaskId)
	if err != nil {
		storage.Logger.Error("failed to delete task", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete task")
	}

	return &pb.DeleteTaskResponse{Success: true}, nil
}

func (s *Server) MoveTask(ctx context.Context, req *pb.MoveTaskRequest) (*pb.Task, error) {
	storage.Logger.Info("RPC MoveTask called", zap.String("taskID", req.TaskId), zap.Int32("newStatus", int32(req.NewStatus)))
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	coreReq := &models.MoveTaskRequest{
		TaskID:    req.TaskId,
		NewStatus: models.TaskStatus(req.NewStatus),
	}
	if req.NewOrderIndex != nil {
		coreReq.NewOrderIndex = req.NewOrderIndex
	}

	task, err := s.core.MoveTask(ctx, coreReq)
	if err != nil {
		storage.Logger.Error("failed to move task", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to move task")
	}

	return convertTaskToProto(task), nil
}

func (s *Server) AssignTask(ctx context.Context, req *pb.AssignTaskRequest) (*pb.Task, error) {
	storage.Logger.Info("RPC AssignTask called", zap.String("taskID", req.TaskId))
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	coreReq := &models.AssignTaskRequest{
		TaskID: req.TaskId,
	}
	if req.AssigneeId != nil {
		coreReq.AssigneeID = req.AssigneeId
	}

	task, err := s.core.AssignTask(ctx, coreReq)
	if err != nil {
		storage.Logger.Error("failed to assign task", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to assign task")
	}

	return convertTaskToProto(task), nil
}

// ============ Member Methods ============

func (s *Server) AddMemberToProject(ctx context.Context, req *pb.AddMemberRequest) (*pb.AddMemberResponse, error) {
	storage.Logger.Info("RPC AddMemberToProject called", zap.String("projectID", req.ProjectId), zap.String("userID", req.UserId))
	if req.GetProjectId() == "" || req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id and user_id are required")
	}

	coreReq := &models.AddMemberRequest{
		ProjectID: req.ProjectId,
		UserID:    req.UserId,
		Role:      models.ProjectRole(req.Role),
	}

	err := s.core.AddMemberToProject(ctx, coreReq)
	if err != nil {
		storage.Logger.Error("failed to add member to project", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to add member to project")
	}

	return &pb.AddMemberResponse{Success: true}, nil
}

func (s *Server) RemoveMemberFromProject(ctx context.Context, req *pb.RemoveMemberRequest) (*pb.RemoveMemberResponse, error) {
	storage.Logger.Info("RPC RemoveMemberFromProject called", zap.String("projectID", req.ProjectId), zap.String("userID", req.UserId))
	if req.GetProjectId() == "" || req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id and user_id are required")
	}

	coreReq := &models.RemoveMemberRequest{
		ProjectID: req.ProjectId,
		UserID:    req.UserId,
	}

	err := s.core.RemoveMemberFromProject(ctx, coreReq)
	if err != nil {
		storage.Logger.Error("failed to remove member from project", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to remove member from project")
	}

	return &pb.RemoveMemberResponse{Success: true}, nil
}

func (s *Server) ListProjectMembers(ctx context.Context, req *pb.ListProjectMembersRequest) (*pb.ListProjectMembersResponse, error) {
	storage.Logger.Info("RPC ListProjectMembers called", zap.String("projectID", req.ProjectId))
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	members, err := s.core.ListProjectMembers(ctx, req.ProjectId)
	if err != nil {
		storage.Logger.Error("failed to list project members", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list project members")
	}

	protoMembers := make([]*pb.ProjectMember, len(members))
	for i, m := range members {
		protoMembers[i] = convertProjectMemberToProto(m)
	}

	return &pb.ListProjectMembersResponse{Members: protoMembers}, nil
}

// ============ Converter Functions ============

func convertProjectToProto(p *models.Project) *pb.Project {
	protoProject := &pb.Project{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		ManagerId:   p.ManagerID,
		Status:      pb.ProjectStatus(p.Status),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
	if p.DueDate != nil {
		protoProject.DueDate = timestamppb.New(*p.DueDate)
	}
	return protoProject
}

func convertTaskToProto(t *models.Task) *pb.Task {
	protoTask := &pb.Task{
		Id:          t.ID,
		ProjectId:   t.ProjectID,
		Title:       t.Title,
		Description: t.Description,
		Status:      pb.TaskStatus(t.Status),
		Priority:    pb.TaskPriority(t.Priority),
		CreatorId:   t.CreatorID,
		OrderIndex:  t.OrderIndex,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
	if t.AssigneeID != nil {
		protoTask.AssigneeId = *t.AssigneeID
	}
	if t.DueDate != nil {
		protoTask.DueDate = timestamppb.New(*t.DueDate)
	}
	return protoTask
}

func convertProjectMemberToProto(m *models.ProjectMember) *pb.ProjectMember {
	return &pb.ProjectMember{
		UserId:    m.UserID,
		ProjectId: m.ProjectID,
		Role:      pb.ProjectRole(m.Role),
		JoinedAt:  timestamppb.New(m.JoinedAt),
	}
}
