package server

import (
	"context"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project"
	core "github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app/project/core"
	pb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/project_service"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ProjectService/internal/app"

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

func (s *Server) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.Project, error) {
	app.Logger.Info("RPC CreateProject called", zap.String("name", req.Name), zap.String("managerID", req.ManagerId))
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
		app.Logger.Error("failed to create project", zap.Error(err))
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


