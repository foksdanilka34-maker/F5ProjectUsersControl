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

	pstatus := models.ProjectStatusFromProtoValue(int32(*req.Status))
	fs := req.GetDueDate()
	dueDate := fs.AsTime()

	updReq := &models.UpdateProjectRequest{
		ID:          req.GetProjectId(),
		Name:        req.Name,
		Description: req.Description,
		Status:      &pstatus,
		DueDate:     &dueDate,
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
		DueDate:     timestamppb.New(*updProject.DueDate),
		Status:      pb.ProjectStatus(updProject.Status.ProtoValue()),
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
