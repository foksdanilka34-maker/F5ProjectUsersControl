package server

import (
	"context"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	core "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/core"
	employee "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/employee_service"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	employee.UnimplementedEmployeeServiceServer
	core core.CoreLogic
}

func NewEmployeeServer(core core.CoreLogic) *Server {
	return &Server{
		core: core,
	}
}

func (s *Server) Register(gRPC *grpc.Server) {
	employee.RegisterEmployeeServiceServer(gRPC, s)
}

func (s *Server) CreateProfile(ctx context.Context, req *employee.CreateProfileRequest) (*employee.Profile, error) {
	storage.Logger.Info("RPC CreateProfile called", zap.String("firstName", req.FirstName), zap.String("lastName", req.LastName), zap.String("email", req.Email))
	if req.GetFirstName() == "" || req.GetLastName() == "" || req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "FirstName, LastName, and Email are required")
	}

	coreReq := &models.RegisterData{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Position:  req.PositionId,
		Email:     req.Email,
		HireDate:  req.HireDate.AsTime(),
		Departm:   req.DepartmentId,
		Login:     req.Login,
		Password:  req.Password,
		Role:      req.Role,
	}

	newProfile, err := s.core.CreateProfile(ctx, coreReq)
	storage.Logger.Debug("created profile", zap.Any("profile", newProfile))
	if err != nil {
		storage.Logger.Error("failed to create profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create profile")
	}

	return convertCoreProfileToProto(newProfile), nil
}

func (s *Server) GetProfile(ctx context.Context, req *employee.GetProfileRequest) (*employee.Profile, error) {
	storage.Logger.Info("RPC GetProfile called", zap.String("userID", req.UserId))
	if req.GetUserId() == "" {
		storage.Logger.Warn("RPC GetProfile: user ID required")
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	profile, err := s.core.GetProfile(ctx, req.UserId)
	if err != nil {
		storage.Logger.Error("failed to get profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get profile")
	}

	return convertCoreProfileToProto(profile), nil
}

func (s *Server) ListProfiles(ctx context.Context, req *employee.ListProfilesRequest) (*employee.ListProfilesResponse, error) {
	storage.Logger.Info("RPC ListProfiles called", zap.Int32("pageSize", req.PageSize), zap.Int32("pageNumber", req.PageNumber))
	profiles, err := s.core.ListProfile(ctx, int(req.GetPageSize()), int(req.GetPageNumber()), req.GetDepartmentId(), req.GetPositionId())
	if err != nil {
		storage.Logger.Error("failed to list profiles", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list profiles")
	}

	protoProfiles := make([]*employee.Profile, len(profiles))
	for i, p := range profiles {
		protoProfiles[i] = convertCoreProfileToProto(p)
	}

	return &employee.ListProfilesResponse{Profiles: protoProfiles}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *employee.UpdateProfileRequest) (*employee.Profile, error) {
	storage.Logger.Info("RPC UpdateProfile called", zap.String("userID", req.UserId))
	if req.GetUserId() == "" {
		storage.Logger.Warn("RPC UpdateProfile: user ID required")
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	coreReq := &models.UpdateProfile{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		PositionId: req.PositionId,
		Email:      req.Email,
		DepartmID:  req.DepartmentId,
		AvatarUrl:  req.AvatarUrl,
	}

	updatedProfile, err := s.core.UpdateProfile(ctx, req.GetUserId(), coreReq)
	if err != nil {
		storage.Logger.Error("failed to update profile", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update profile")
	}

	return convertCoreProfileToProto(updatedProfile), nil
}

func (s *Server) ChangeUserStatusProfile(ctx context.Context, req *employee.DeactivateProfileRequest) (*emptypb.Empty, error) {
	storage.Logger.Info("RPC ChangeUserStatusProfile called", zap.String("userID", req.UserId))
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	if err := s.core.DeactivateProfile(ctx, req.GetUserId(), req.GetStatus()); err != nil {
		storage.Logger.Error("failed to change profile status", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to change profile status")
	}

	return &emptypb.Empty{}, nil
}

func convertCoreProfileToProto(p *models.Profile) *employee.Profile {
	if p == nil {
		return nil
	}

	return &employee.Profile{
		Id:         p.UserID,
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		PositionId: p.PositionId,
		Email:      p.Email,
		Department: &employee.Department{
			Id:   p.Departm.ID,
			Name: p.Departm.Name,
		},
		AvatarUrl: p.AvatarUrl,
		HireDate:  timestamppb.New(p.HireDate),
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}
}
