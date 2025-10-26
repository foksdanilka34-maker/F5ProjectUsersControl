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

func (s *Server) CreateDepartment(ctx context.Context, req *employee.CreateDepartmentRequest) (*employee.Department, error) {
	storage.Logger.Info("RPC CreateDepartment called", zap.String("name", req.Name))
	if req.GetName() == "" {
		storage.Logger.Warn("RPC CreateDepartment: name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	department, err := s.core.CreateDepartment(ctx, req.GetName())
	if err != nil {
		storage.Logger.Error("failed to create department", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create department")
	}

	return &employee.Department{
		Id:   department.ID,
		Name: department.Name,
	}, nil
}

func (s *Server) GetDepartment(ctx context.Context, req *employee.GetDepartmentRequest) (*employee.Department, error) {
	storage.Logger.Info("RPC GetDepartment called", zap.String("id", req.Id))
	if req.GetId() == "" {
		storage.Logger.Warn("RPC GetDepartment: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	department, err := s.core.GetDepartment(ctx, req.GetId())
	if err != nil {
		storage.Logger.Error("failed to get department", zap.Error(err))
		return nil, status.Error(codes.NotFound, "department not found")
	}

	return &employee.Department{
		Id:   department.ID,
		Name: department.Name,
	}, nil
}

func (s *Server) ListDepartments(ctx context.Context, req *emptypb.Empty) (*employee.ListDepartmentsResponse, error) {
	storage.Logger.Info("RPC ListDepartments called")
	departments, err := s.core.ListDepartments(ctx)
	if err != nil {
		storage.Logger.Error("failed to list departments", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list departments")
	}

	protoDepartments := make([]*employee.Department, len(departments))
	for i, d := range departments {
		protoDepartments[i] = &employee.Department{
			Id:   d.ID,
			Name: d.Name,
		}
	}

	return &employee.ListDepartmentsResponse{Departments: protoDepartments}, nil
}

func (s *Server) UpdateDepartment(ctx context.Context, req *employee.UpdateDepartmentRequest) (*employee.Department, error) {
	storage.Logger.Info("RPC UpdateDepartment called", zap.String("id", req.Id), zap.String("name", req.Name))
	if req.GetId() == "" {
		storage.Logger.Warn("RPC UpdateDepartment: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetName() == "" {
		storage.Logger.Warn("RPC UpdateDepartment: name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	department, err := s.core.UpdateDepartment(ctx, req.GetId(), req.GetName())
	if err != nil {
		storage.Logger.Error("failed to update department", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update department")
	}

	return &employee.Department{
		Id:   department.ID,
		Name: department.Name,
	}, nil
}

func (s *Server) DeleteDepartment(ctx context.Context, req *employee.DeleteDepartmentRequest) (*emptypb.Empty, error) {
	storage.Logger.Info("RPC DeleteDepartment called", zap.String("id", req.Id))
	if req.GetId() == "" {
		storage.Logger.Warn("RPC DeleteDepartment: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := s.core.DeleteDepartment(ctx, req.GetId())
	if err != nil {
		storage.Logger.Error("failed to delete department", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete department")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) CreatePosition(ctx context.Context, req *employee.CreatePositionRequest) (*employee.Position, error) {
	storage.Logger.Info("RPC CreatePosition called", zap.String("name", req.Name))
	if req.GetName() == "" {
		storage.Logger.Warn("RPC CreatePosition: name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	position, err := s.core.CreatePosition(ctx, req.GetName())
	if err != nil {
		storage.Logger.Error("failed to create position", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create position")
	}

	return &employee.Position{
		Id:   position.ID,
		Name: position.Name,
	}, nil
}

func (s *Server) GetPosition(ctx context.Context, req *employee.GetPositionRequest) (*employee.Position, error) {
	storage.Logger.Info("RPC GetPosition called", zap.String("id", req.Id))
	if req.GetId() == "" {
		storage.Logger.Warn("RPC GetPosition: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	position, err := s.core.GetPosition(ctx, req.GetId())
	if err != nil {
		storage.Logger.Error("failed to get position", zap.Error(err))
		return nil, status.Error(codes.NotFound, "position not found")
	}

	return &employee.Position{
		Id:   position.ID,
		Name: position.Name,
	}, nil
}

func (s *Server) ListPositions(ctx context.Context, req *emptypb.Empty) (*employee.ListPositionsResponse, error) {
	storage.Logger.Info("RPC ListPositions called")
	positions, err := s.core.ListPositions(ctx)
	if err != nil {
		storage.Logger.Error("failed to list positions", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list positions")
	}

	protoPositions := make([]*employee.Position, len(positions))
	for i, p := range positions {
		protoPositions[i] = &employee.Position{
			Id:   p.ID,
			Name: p.Name,
		}
	}

	return &employee.ListPositionsResponse{Positions: protoPositions}, nil
}

func (s *Server) UpdatePosition(ctx context.Context, req *employee.UpdatePositionRequest) (*employee.Position, error) {
	storage.Logger.Info("RPC UpdatePosition called", zap.String("id", req.Id), zap.String("name", req.Name))
	if req.GetId() == "" {
		storage.Logger.Warn("RPC UpdatePosition: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetName() == "" {
		storage.Logger.Warn("RPC UpdatePosition: name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	position, err := s.core.UpdatePosition(ctx, req.GetId(), req.GetName())
	if err != nil {
		storage.Logger.Error("failed to update position", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update position")
	}

	return &employee.Position{
		Id:   position.ID,
		Name: position.Name,
	}, nil
}

func (s *Server) DeletePosition(ctx context.Context, req *employee.DeletePositionRequest) (*emptypb.Empty, error) {
	storage.Logger.Info("RPC DeletePosition called", zap.String("id", req.Id))
	if req.GetId() == "" {
		storage.Logger.Warn("RPC DeletePosition: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := s.core.DeletePosition(ctx, req.GetId())
	if err != nil {
		storage.Logger.Error("failed to delete position", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete position")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) CreateSkill(ctx context.Context, req *employee.CreateSkillRequest) (*employee.Skill, error) {
	storage.Logger.Info("RPC CreateSkill called", zap.String("name", req.Name))
	if req.GetName() == "" {
		storage.Logger.Warn("RPC CreateSkill: name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	skill, err := s.core.CreateSkill(ctx, req.GetName())
	if err != nil {
		storage.Logger.Error("failed to create skill", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create skill")
	}

	return &employee.Skill{
		Id:   skill.ID,
		Name: skill.Name,
	}, nil
}

func (s *Server) ListSkills(ctx context.Context, req *emptypb.Empty) (*employee.ListSkillsResponse, error) {
	storage.Logger.Info("RPC ListSkills called")
	skills, err := s.core.ListSkills(ctx)
	if err != nil {
		storage.Logger.Error("failed to list skills", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list skills")
	}

	protoSkills := make([]*employee.Skill, len(skills))
	for i, sk := range skills {
		protoSkills[i] = &employee.Skill{
			Id:   sk.ID,
			Name: sk.Name,
		}
	}

	return &employee.ListSkillsResponse{Skills: protoSkills}, nil
}

func (s *Server) AddSkillToEmployee(ctx context.Context, req *employee.AddSkillToEmployeeRequest) (*emptypb.Empty, error) {
	storage.Logger.Info("RPC AddSkillToEmployee called", zap.String("employeeID", req.EmployeeId), zap.String("skillID", req.SkillId))
	if req.GetEmployeeId() == "" {
		storage.Logger.Warn("RPC AddSkillToEmployee: employee_id is required")
		return nil, status.Error(codes.InvalidArgument, "employee_id is required")
	}
	if req.GetSkillId() == "" {
		storage.Logger.Warn("RPC AddSkillToEmployee: skill_id is required")
		return nil, status.Error(codes.InvalidArgument, "skill_id is required")
	}

	err := s.core.AddSkillToEmployee(ctx, req.GetEmployeeId(), req.GetSkillId())
	if err != nil {
		storage.Logger.Error("failed to add skill to employee", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to add skill to employee")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RemoveSkillFromEmployee(ctx context.Context, req *employee.RemoveSkillFromEmployeeRequest) (*emptypb.Empty, error) {
	storage.Logger.Info("RPC RemoveSkillFromEmployee called", zap.String("employeeID", req.EmployeeId), zap.String("skillID", req.SkillId))
	if req.GetEmployeeId() == "" {
		storage.Logger.Warn("RPC RemoveSkillFromEmployee: employee_id is required")
		return nil, status.Error(codes.InvalidArgument, "employee_id is required")
	}
	if req.GetSkillId() == "" {
		storage.Logger.Warn("RPC RemoveSkillFromEmployee: skill_id is required")
		return nil, status.Error(codes.InvalidArgument, "skill_id is required")
	}

	err := s.core.RemoveSkillFromEmployee(ctx, req.GetEmployeeId(), req.GetSkillId())
	if err != nil {
		storage.Logger.Error("failed to remove skill from employee", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to remove skill from employee")
	}

	return &emptypb.Empty{}, nil
}
