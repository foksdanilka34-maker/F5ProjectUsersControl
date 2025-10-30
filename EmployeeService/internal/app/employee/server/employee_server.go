package server

import (
	"context"
	"log"

	models "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	core "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/core"
	employee "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/employee_service"

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
	log.Printf("RPC CreateProfile called: firstName=%s, lastName=%s, email=%s", req.FirstName, req.LastName, req.Email)
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
	log.Printf("created profile: %+v", newProfile)
	if err != nil {
		log.Printf("failed to create profile: %v", err)
		return nil, status.Error(codes.Internal, "failed to create profile")
	}

	return convertCoreProfileToProto(newProfile), nil
}

func (s *Server) GetProfile(ctx context.Context, req *employee.GetProfileRequest) (*employee.Profile, error) {
	log.Printf("RPC GetProfile called: userID=%s", req.UserId)
	if req.GetUserId() == "" {
		log.Printf("RPC GetProfile: user ID required")
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	profile, err := s.core.GetProfile(ctx, req.UserId)
	if err != nil {
		log.Printf("failed to get profile: %v", err)
		return nil, status.Error(codes.Internal, "failed to get profile")
	}

	return convertCoreProfileToProto(profile), nil
}

func (s *Server) ListProfiles(ctx context.Context, req *employee.ListProfilesRequest) (*employee.ListProfilesResponse, error) {
	log.Printf("RPC ListProfiles called: pageSize=%d, pageNumber=%d", req.PageSize, req.PageNumber)
	profiles, err := s.core.ListProfile(ctx, int(req.GetPageSize()), int(req.GetPageNumber()), req.GetDepartmentId(), req.GetPositionId())
	if err != nil {
		log.Printf("failed to list profiles: %v", err)
		return nil, status.Error(codes.Internal, "failed to list profiles")
	}

	protoProfiles := make([]*employee.Profile, len(profiles))
	for i, p := range profiles {
		protoProfiles[i] = convertCoreProfileToProto(p)
	}

	return &employee.ListProfilesResponse{Profiles: protoProfiles}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *employee.UpdateProfileRequest) (*employee.Profile, error) {
	log.Printf("RPC UpdateProfile called: userID=%s", req.UserId)
	if req.GetUserId() == "" {
		log.Printf("RPC UpdateProfile: user ID required")
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
		log.Printf("failed to update profile: %v", err)
		return nil, status.Error(codes.Internal, "failed to update profile")
	}

	return convertCoreProfileToProto(updatedProfile), nil
}

func (s *Server) ChangeUserStatusProfile(ctx context.Context, req *employee.DeactivateProfileRequest) (*emptypb.Empty, error) {
	log.Printf("RPC ChangeUserStatusProfile called: userID=%s", req.UserId)
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	if err := s.core.DeactivateProfile(ctx, req.GetUserId(), req.GetStatus()); err != nil {
		log.Printf("failed to change profile status: %v", err)
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
	log.Printf("RPC CreateDepartment called: name=%s", req.Name)
	if req.GetName() == "" {
		log.Printf("RPC CreateDepartment: name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	department, err := s.core.CreateDepartment(ctx, req.GetName())
	if err != nil {
		log.Printf("failed to create department: %v", err)
		return nil, status.Error(codes.Internal, "failed to create department")
	}

	return &employee.Department{
		Id:   department.ID,
		Name: department.Name,
	}, nil
}

func (s *Server) GetDepartment(ctx context.Context, req *employee.GetDepartmentRequest) (*employee.Department, error) {
	log.Printf("RPC GetDepartment called: id=%s", req.Id)
	if req.GetId() == "" {
		log.Printf("RPC GetDepartment: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	department, err := s.core.GetDepartment(ctx, req.GetId())
	if err != nil {
		log.Printf("failed to get department: %v", err)
		return nil, status.Error(codes.NotFound, "department not found")
	}

	return &employee.Department{
		Id:   department.ID,
		Name: department.Name,
	}, nil
}

func (s *Server) ListDepartments(ctx context.Context, req *emptypb.Empty) (*employee.ListDepartmentsResponse, error) {
	log.Printf("RPC ListDepartments called")
	departments, err := s.core.ListDepartments(ctx)
	if err != nil {
		log.Printf("failed to list departments: %v", err)
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
	log.Printf("RPC UpdateDepartment called: id=%s, name=%s", req.Id, req.Name)
	if req.GetId() == "" {
		log.Printf("RPC UpdateDepartment: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetName() == "" {
		log.Printf("RPC UpdateDepartment: name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	department, err := s.core.UpdateDepartment(ctx, req.GetId(), req.GetName())
	if err != nil {
		log.Printf("failed to update department: %v", err)
		return nil, status.Error(codes.Internal, "failed to update department")
	}

	return &employee.Department{
		Id:   department.ID,
		Name: department.Name,
	}, nil
}

func (s *Server) DeleteDepartment(ctx context.Context, req *employee.DeleteDepartmentRequest) (*emptypb.Empty, error) {
	log.Printf("RPC DeleteDepartment called: id=%s", req.Id)
	if req.GetId() == "" {
		log.Printf("RPC DeleteDepartment: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := s.core.DeleteDepartment(ctx, req.GetId())
	if err != nil {
		log.Printf("failed to delete department: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete department")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) CreatePosition(ctx context.Context, req *employee.CreatePositionRequest) (*employee.Position, error) {
	log.Printf("RPC CreatePosition called: name=%s", req.Name)
	if req.GetName() == "" {
		log.Printf("RPC CreatePosition: name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	position, err := s.core.CreatePosition(ctx, req.GetName())
	if err != nil {
		log.Printf("failed to create position: %v", err)
		return nil, status.Error(codes.Internal, "failed to create position")
	}

	return &employee.Position{
		Id:   position.ID,
		Name: position.Name,
	}, nil
}

func (s *Server) GetPosition(ctx context.Context, req *employee.GetPositionRequest) (*employee.Position, error) {
	log.Printf("RPC GetPosition called: id=%s", req.Id)
	if req.GetId() == "" {
		log.Printf("RPC GetPosition: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	position, err := s.core.GetPosition(ctx, req.GetId())
	if err != nil {
		log.Printf("failed to get position: %v", err)
		return nil, status.Error(codes.NotFound, "position not found")
	}

	return &employee.Position{
		Id:   position.ID,
		Name: position.Name,
	}, nil
}

func (s *Server) ListPositions(ctx context.Context, req *emptypb.Empty) (*employee.ListPositionsResponse, error) {
	log.Printf("RPC ListPositions called")
	positions, err := s.core.ListPositions(ctx)
	if err != nil {
		log.Printf("failed to list positions: %v", err)
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
	log.Printf("RPC UpdatePosition called: id=%s, name=%s", req.Id, req.Name)
	if req.GetId() == "" {
		log.Printf("RPC UpdatePosition: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetName() == "" {
		log.Printf("RPC UpdatePosition: name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	position, err := s.core.UpdatePosition(ctx, req.GetId(), req.GetName())
	if err != nil {
		log.Printf("failed to update position: %v", err)
		return nil, status.Error(codes.Internal, "failed to update position")
	}

	return &employee.Position{
		Id:   position.ID,
		Name: position.Name,
	}, nil
}

func (s *Server) DeletePosition(ctx context.Context, req *employee.DeletePositionRequest) (*emptypb.Empty, error) {
	log.Printf("RPC DeletePosition called: id=%s", req.Id)
	if req.GetId() == "" {
		log.Printf("RPC DeletePosition: id is required")
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	err := s.core.DeletePosition(ctx, req.GetId())
	if err != nil {
		log.Printf("failed to delete position: %v", err)
		return nil, status.Error(codes.Internal, "failed to delete position")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) CreateSkill(ctx context.Context, req *employee.CreateSkillRequest) (*employee.Skill, error) {
	log.Printf("RPC CreateSkill called: name=%s", req.Name)
	if req.GetName() == "" {
		log.Printf("RPC CreateSkill: name is required")
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	skill, err := s.core.CreateSkill(ctx, req.GetName())
	if err != nil {
		log.Printf("failed to create skill: %v", err)
		return nil, status.Error(codes.Internal, "failed to create skill")
	}

	return &employee.Skill{
		Id:   skill.ID,
		Name: skill.Name,
	}, nil
}

func (s *Server) ListSkills(ctx context.Context, req *emptypb.Empty) (*employee.ListSkillsResponse, error) {
	log.Printf("RPC ListSkills called")
	skills, err := s.core.ListSkills(ctx)
	if err != nil {
		log.Printf("failed to list skills: %v", err)
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
	log.Printf("RPC AddSkillToEmployee called: employeeID=%s, skillID=%s", req.EmployeeId, req.SkillId)
	if req.GetEmployeeId() == "" {
		log.Printf("RPC AddSkillToEmployee: employee_id is required")
		return nil, status.Error(codes.InvalidArgument, "employee_id is required")
	}
	if req.GetSkillId() == "" {
		log.Printf("RPC AddSkillToEmployee: skill_id is required")
		return nil, status.Error(codes.InvalidArgument, "skill_id is required")
	}

	err := s.core.AddSkillToEmployee(ctx, req.GetEmployeeId(), req.GetSkillId())
	if err != nil {
		log.Printf("failed to add skill to employee: %v", err)
		return nil, status.Error(codes.Internal, "failed to add skill to employee")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RemoveSkillFromEmployee(ctx context.Context, req *employee.RemoveSkillFromEmployeeRequest) (*emptypb.Empty, error) {
	log.Printf("RPC RemoveSkillFromEmployee called: employeeID=%s, skillID=%s", req.EmployeeId, req.SkillId)
	if req.GetEmployeeId() == "" {
		log.Printf("RPC RemoveSkillFromEmployee: employee_id is required")
		return nil, status.Error(codes.InvalidArgument, "employee_id is required")
	}
	if req.GetSkillId() == "" {
		log.Printf("RPC RemoveSkillFromEmployee: skill_id is required")
		return nil, status.Error(codes.InvalidArgument, "skill_id is required")
	}

	err := s.core.RemoveSkillFromEmployee(ctx, req.GetEmployeeId(), req.GetSkillId())
	if err != nil {
		log.Printf("failed to remove skill from employee: %v", err)
		return nil, status.Error(codes.Internal, "failed to remove skill from employee")
	}

	return &emptypb.Empty{}, nil
}
