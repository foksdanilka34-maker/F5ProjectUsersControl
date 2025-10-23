package server

import (
	"context"

	employee "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/employee_service"
	core "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/core"
	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	
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
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create profile")
	}

	return convertCoreProfileToProto(newProfile), nil
}

func (s *Server) GetProfile(ctx context.Context, req *employee.GetProfileRequest) (*employee.Profile, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	profile, err := s.core.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get profile")
	}

	return convertCoreProfileToProto(profile), nil
}

func (s *Server) ListProfiles(ctx context.Context, req *employee.ListProfilesRequest) (*employee.ListProfilesResponse, error) {
	profiles, err := s.core.ListProfile(ctx, int(req.GetPageSize()), int(req.GetPageNumber()), req.GetDepartmentId(), req.GetSkillId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list profiles")
	}

	protoProfiles := make([]*employee.Profile, len(profiles))
	for i, p := range profiles {
		protoProfiles[i] = convertCoreProfileToProto(p)
	}

	return &employee.ListProfilesResponse{Profiles: protoProfiles}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *employee.UpdateProfileRequest) (*employee.Profile, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	coreReq := &models.UpdateProfile{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		PositionId: req.PositionId,
		Email:      req.Email,
		Departm:    req.DepartmentId,
		AvatarUrl:  req.AvatarUrl,
		HireDate:   req.HireDate.AsTime(),
	}

	updatedProfile, err := s.core.UpdateProfile(ctx, req.GetId(), coreReq)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update profile")
	}

	return convertCoreProfileToProto(updatedProfile), nil
}

func (s *Server) DeactivateProfile(ctx context.Context, req *employee.DeactivateProfileRequest) (*emptypb.Empty, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	if err := s.core.DeactivateProfile(ctx, req.GetId()); err != nil {
		return nil, status.Error(codes.Internal, "failed to deactivate profile")
	}

	return &emptypb.Empty{}, nil
}

// convertCoreProfileToProto converts the internal Profile model to the protobuf model.
func convertCoreProfileToProto(p *models.Profile) *employee.Profile {
	if p == nil {
		return nil
	}
	return &employee.Profile{
		Id:           p.UserID,
		FirstName:    p.FirstName,
		LastName:     p.LastName,
		PositionId:   p.PositionId,
		Email:        p.Email,
		DepartmentId: p.Departm.ID,
		AvatarUrl:    p.AvatarUrl,
		HireDate:     timestamppb.New(p.HireDate),
		CreatedAt:    timestamppb.New(p.CreatedAt),
		UpdatedAt:    timestamppb.New(p.UpdatedAt),
	}
}