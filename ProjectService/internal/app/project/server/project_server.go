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
