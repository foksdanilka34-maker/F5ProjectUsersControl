package server

import (
	"context"
	"log"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"
	core "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/core"
	pb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/project_service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func taskToProto(task *models.Task) *pb.Task {
	pbTask := &pb.Task{
		Id:          task.ID,
		ProjectId:   task.ProjectID,
		Title:       task.TaskName,
		Description: task.Description,
		Status:      pb.TaskStatus(task.Status.ProtoValue()),
		Priority:    pb.TaskPriority(task.Priority.ProtoValue()),
		CreatorId:   task.CreatorID,
		OrderIndex:  task.OrderIndex,
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}

	if task.DueDate != nil {
		pbTask.DueDate = timestamppb.New(*task.DueDate)
	}

	if task.StartedAt != nil {
		pbTask.StartedAt = timestamppb.New(*task.StartedAt)
	}

	if task.CompletedAt != nil {
		pbTask.CompletedAt = timestamppb.New(*task.CompletedAt)
	}

	return pbTask
}

func (s *Server) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.Project, error) {
	log.Printf("RPC CreateProject called: name=%s, managerID=%s", req.Name, req.ManagerId)
	if req.GetName() == "" || req.GetManagerId() == "" {
		return nil, status.Error(codes.InvalidArgument, "name and manager_id are required")
	}
	coreReq := &models.CreateProjectRequest{
		Name:        req.Name,
		Description: req.Description,
		ManagerID:   req.ManagerId,
	}
	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		coreReq.DueDate = &dueDate
	}

	project, err := s.core.CreateProject(ctx, coreReq)
	if err != nil {
		log.Printf("failed to create project: %v", err)
		return nil, status.Error(codes.Internal, "failed to create project")
	}

	pbProject := &pb.Project{
		Id:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		ManagerId:   project.ManagerID,
		CreatedAt:   timestamppb.New(project.CreatedAt),
		UpdatedAt:   timestamppb.New(project.UpdatedAt),
	}
	if project.DueDate != nil {
		pbProject.DueDate = timestamppb.New(*project.DueDate)
	}

	return pbProject, nil
}

func (s *Server) GetProject(ctx context.Context, req *pb.GetProjectRequest) (*pb.Project, error) {
	log.Printf("RPC GetProject called: projectID=%s", req.ProjectId)
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "projectID required")
	}
	project, err := s.core.GetProject(ctx, req.ProjectId)
	if err != nil {
		log.Printf("failed to get project: %v", err)
		return nil, status.Error(codes.Internal, "failed to get project")
	}

	pbProject := &pb.Project{
		Id:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		ManagerId:   project.ManagerID,
		Status:      pb.ProjectStatus(project.Status.ProtoValue()),
		CreatedAt:   timestamppb.New(project.CreatedAt),
		UpdatedAt:   timestamppb.New(project.UpdatedAt),
	}
	return pbProject, nil
}

func (s *Server) ListProjects(ctx context.Context, req *pb.ListProjectsRequest) (*pb.ListProjectsResponse, error) {
	log.Printf("RPC ListProject called: managerId=%s, Status %s", req.GetManagerId(), req.GetStatus())

	coreReq := &models.ListProjectsFilter{
		PageSize:   req.PageSize,
		PageNumber: req.PageNumber,
		ManagerID:  req.ManagerId,
	}

	if req.Status != nil {
		projStatus := models.ProjectStatusFromProtoValue(int32(*req.Status))
		coreReq.Status = &projStatus
	}
	projects, err := s.core.ListProjects(ctx, coreReq)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list projects")
	}

	if projects == nil {
		projects = &models.ProjectsListResponse{
			Projects:   make([]*models.Project, 0),
			TotalCount: 0,
		}
	}

	pbProjects := make([]*pb.Project, 0, len(projects.Projects))
	for _, p := range projects.Projects {
		pbP := &pb.Project{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			ManagerId:   p.ManagerID,
			Status:      pb.ProjectStatus(p.Status.ProtoValue()),
			CreatedAt:   timestamppb.New(p.CreatedAt),
			UpdatedAt:   timestamppb.New(p.UpdatedAt),
		}
		if p.DueDate != nil {
			pbP.DueDate = timestamppb.New(*p.DueDate)
		}
		pbProjects = append(pbProjects, pbP)
	}

	listResponse := &pb.ListProjectsResponse{
		Projects:   pbProjects,
		TotalCount: projects.TotalCount,
	}

	return listResponse, nil
}

func (s *Server) UpdateProject(ctx context.Context, req *pb.UpdateProjectRequest) (*pb.Project, error) {
	log.Printf("RPC UpdateProject called %s:", req.GetProjectId())
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project id required")
	}

	updReq := &models.UpdateProjectRequest{
		ID:          req.GetProjectId(),
		Name:        req.Name,
		Description: req.Description,
	}

	if req.Status != nil {
		pstatus := models.ProjectStatusFromProtoValue(int32(*req.Status))
		updReq.Status = &pstatus
	}

	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		updReq.DueDate = &dueDate
	}

	updProject, err := s.core.UpdateProject(ctx, updReq)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update project")
	}

	pbUpdProject := &pb.Project{
		Id:          updProject.ID,
		Name:        updProject.Name,
		Description: updProject.Description,
		ManagerId:   updProject.ManagerID,
		CreatedAt:   timestamppb.New(updProject.CreatedAt),
		UpdatedAt:   timestamppb.New(updProject.UpdatedAt),
		Status:      pb.ProjectStatus(updProject.Status.ProtoValue()),
	}

	if updProject.DueDate != nil {
		pbUpdProject.DueDate = timestamppb.New(*updProject.DueDate)
	}

	return pbUpdProject, nil
}

func (s *Server) DeleteProject(ctx context.Context, req *pb.DeleteProjectRequest) (*emptypb.Empty, error) {
	log.Printf("RPC DeleteProject called for project %s", req.GetProjectId())
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	err := s.core.DeleteProject(ctx, req.GetProjectId())
	if err != nil {
		log.Printf("failed to delete project: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete project")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
	log.Printf("RPC CreateTask called for project %s", req.GetProjectId())
	if req.GetProjectId() == "" || req.GetTitle() == "" || req.GetCreatorId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id, title and creator_id are required")
	}

	priority := models.TaskPriorityMedium
	if req.Priority != nil {
		priority = models.TaskPriorityFromProtoValue(int32(*req.Priority))
	}

	coreReq := &models.CreateTaskRequest{
		ProjectID:   req.ProjectId,
		TaskName:    req.Title,
		Description: req.Description,
		Status:      models.TaskStatusTodo,
		Priority:    priority,
		CreatorID:   req.CreatorId,
		DueDate:     req.DueDate.AsTime(),
	}

	task, err := s.core.CreateTask(ctx, coreReq)
	if err != nil {
		log.Printf("failed to create task: %v", err)
		return nil, status.Error(codes.Internal, "failed to create task")
	}

	return taskToProto(task), nil
}

func (s *Server) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.Task, error) {
	log.Printf("RPC GetTask called: taskID=%s", req.GetTaskId())
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	task, err := s.core.GetTask(ctx, req.TaskId)
	if err != nil {
		log.Printf("failed to get task: %v", err)
		return nil, status.Error(codes.Internal, "failed to get task")
	}

	return taskToProto(task), nil
}

func (s *Server) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.Task, error) {
	log.Printf("RPC UpdateTask called: taskID=%s", req.GetTaskId())
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	updReq := &models.UpdateTaskRequest{
		ID:          req.TaskId,
		TaskName:    req.Title,
		Description: req.Description,
		AssigneeID:  req.AssigneeId,
	}

	if req.Status != nil {
		status := models.TaskStatusFromProtoValue(int32(*req.Status))
		updReq.Status = &status
	}

	if req.Priority != nil {
		priority := models.TaskPriorityFromProtoValue(int32(*req.Priority))
		updReq.Priority = &priority
	}

	if req.DueDate != nil {
		dueDate := req.DueDate.AsTime()
		updReq.DueDate = &dueDate
	}

	task, err := s.core.UpdateTask(ctx, updReq)
	if err != nil {
		log.Printf("failed to update task: %v", err)
		return nil, status.Error(codes.Internal, "failed to update task")
	}

	return taskToProto(task), nil
}

func (s *Server) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*emptypb.Empty, error) {
	log.Printf("RPC DeleteTask called for task %s", req.GetTaskId())
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	err := s.core.DeleteTask(ctx, req.TaskId)
	if err != nil {
		log.Printf("failed to delete task: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete task")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) MoveTask(ctx context.Context, req *pb.MoveTaskRequest) (*pb.Task, error) {
	log.Printf("RPC MoveTask called for task %s to status %s", req.GetTaskId(), req.GetNewStatus())
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	moveReq := &models.MoveTaskRequest{
		TaskID:        req.TaskId,
		NewStatus:     models.TaskStatusFromProtoValue(int32(req.NewStatus)),
		NewOrderIndex: req.NewOrderIndex,
	}

	task, err := s.core.MoveTask(ctx, moveReq)
	if err != nil {
		log.Printf("failed to move task: %v", err)
		return nil, status.Error(codes.Internal, "failed to move task")
	}

	return taskToProto(task), nil
}

func (s *Server) AssignTask(ctx context.Context, req *pb.AssignTaskRequest) (*pb.Task, error) {
	log.Printf("RPC AssignTask called for task %s to assignee %s", req.GetTaskId(), req.GetAssigneeId())
	if req.GetTaskId() == "" || req.GetAssigneeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id and assignee_id are required")
	}

	assignReq := &models.AssignTaskRequest{
		TaskID:     req.TaskId,
		AssigneeID: req.AssigneeId,
	}

	task, err := s.core.AssignTask(ctx, assignReq)
	if err != nil {
		log.Printf("failed to assign task: %v", err)
		return nil, status.Error(codes.Internal, "failed to assign task")
	}

	return taskToProto(task), nil
}

func (s *Server) ListTasksByProject(ctx context.Context, req *pb.ListTasksByProjectRequest) (*pb.ListTasksByProjectResponse, error) {
	log.Printf("RPC ListTasksByProject called for project %s", req.GetProjectId())
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	filter := &models.ListTasksFilter{
		ProjectID:  req.ProjectId,
		AssigneeID: req.AssigneeId,
	}

	if req.Status != nil {
		taskStatus := models.TaskStatusFromProtoValue(int32(*req.Status))
		filter.Status = &taskStatus
	}

	if req.Priority != nil {
		taskPriority := models.TaskPriorityFromProtoValue(int32(*req.Priority))
		filter.Priority = &taskPriority
	}

	tasks, err := s.core.ListTasksByProject(ctx, filter)
	if err != nil {
		log.Printf("failed to list tasks: %v", err)
		return nil, status.Error(codes.Internal, "failed to list tasks")
	}

	pbTasks := make([]*pb.Task, 0, len(tasks.Tasks))
	for _, task := range tasks.Tasks {
		pbTasks = append(pbTasks, taskToProto(task))
	}

	return &pb.ListTasksByProjectResponse{
		Tasks: pbTasks,
	}, nil
}

func (s *Server) AddMemberToProject(ctx context.Context, req *pb.AddMemberToProjectRequest) (*emptypb.Empty, error) {
	log.Printf("RPC AddMemberToProject called: projectID=%s, userID=%s", req.GetProjectId(), req.GetUserId())
	if req.GetProjectId() == "" || req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id and user_id are required")
	}

	err := s.core.AddMemberToProject(ctx, req.ProjectId, req.UserId)
	if err != nil {
		log.Printf("failed to add member to project: %v", err)
		return nil, status.Error(codes.Internal, "failed to add member to project")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RemoveMemberFromProject(ctx context.Context, req *pb.RemoveMemberFromProjectRequest) (*emptypb.Empty, error) {
	log.Printf("RPC RemoveMemberFromProject called: projectID=%s, userID=%s", req.GetProjectId(), req.GetUserId())
	if req.GetProjectId() == "" || req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id and user_id are required")
	}

	err := s.core.RemoveMemberFromProject(ctx, req.ProjectId, req.UserId)
	if err != nil {
		log.Printf("failed to remove member from project: %v", err)
		return nil, status.Error(codes.Internal, "failed to remove member from project")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) ListProjectMembers(ctx context.Context, req *pb.ListProjectMembersRequest) (*pb.ListProjectMembersResponse, error) {
	log.Printf("RPC ListProjectMembers called for project %s", req.GetProjectId())
	if req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	members, err := s.core.ListProjectMembers(ctx, req.ProjectId)
	if err != nil {
		log.Printf("failed to list project members: %v", err)
		return nil, status.Error(codes.Internal, "failed to list project members")
	}

	pbMembers := make([]*pb.ProjectMember, 0, len(members.Members))
	for _, member := range members.Members {
		pbMembers = append(pbMembers, &pb.ProjectMember{
			UserId:   member.UserID,
			FullName: member.FullName,
			Role:     member.Role.String(),
		})
	}

	return &pb.ListProjectMembersResponse{
		Members: pbMembers,
	}, nil
}
