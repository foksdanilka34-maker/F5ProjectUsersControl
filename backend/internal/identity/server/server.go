package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/identity"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/identity/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/identity/repo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// IdentityServer - gRPC сервер Identity сервиса
type IdentityServer struct {
	identity.UnimplementedIdentityServiceServer
	authService    *core.AuthService
	profileService *core.ProfileService
}

func NewIdentityServer(authService *core.AuthService, profileService *core.ProfileService) *IdentityServer {
	return &IdentityServer{
		authService:    authService,
		profileService: profileService,
	}
}

func (s *IdentityServer) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	identity.RegisterIdentityServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	log.Printf("Identity gRPC server listening on %s", addr)
	return grpcServer.Serve(lis)
}

// Login - авторизация пользователя
func (s *IdentityServer) Login(ctx context.Context, req *identity.LoginRequest) (*identity.LoginResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	userAgent := ""
	ipAddress := ""
	if ua := md.Get("user-agent"); len(ua) > 0 {
		userAgent = ua[0]
	}
	if ip := md.Get("x-forwarded-for"); len(ip) > 0 {
		ipAddress = ip[0]
	}
	if req.UserAgent != "" {
		userAgent = req.UserAgent
	}
	if req.IpAddress != "" {
		ipAddress = req.IpAddress
	}

	userID, accessToken, _, err := s.authService.Login(ctx, req.Login, req.Password, userAgent, ipAddress)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", err)
	}

	// Получаем профиль для UserInfo
	profile, err := s.profileService.GetProfile(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get profile: %v", err)
	}

	return &identity.LoginResponse{
		AccessToken: accessToken,
		User: &identity.UserInfo{
			Id:        profile.ID,
			Login:     profile.Login,
			FullName:  profile.FirstName + " " + profile.LastName,
			Role:      profile.Role,
			AvatarUrl: profile.AvatarURL,
		},
	}, nil
}

// GetMe - получение информации о текущем пользователе
func (s *IdentityServer) GetMe(ctx context.Context, req *emptypb.Empty) (*identity.UserInfo, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	userIDStr := md.Get("x-user-id")
	if len(userIDStr) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing user id")
	}

	userID, err := core.ParseUserID(userIDStr[0])
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	profile, err := s.profileService.GetProfile(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "profile not found: %v", err)
	}

	return &identity.UserInfo{
		Id:        profile.ID,
		Login:     profile.Login,
		FullName:  profile.FirstName + " " + profile.LastName,
		Role:      profile.Role,
		AvatarUrl: profile.AvatarURL,
	}, nil
}

// Logout - выход из системы
func (s *IdentityServer) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// Refresh - обновление токена
func (s *IdentityServer) Refresh(ctx context.Context, req *emptypb.Empty) (*identity.RefreshResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	userAgent := ""
	ipAddress := ""
	if ua := md.Get("user-agent"); len(ua) > 0 {
		userAgent = ua[0]
	}
	if ip := md.Get("x-forwarded-for"); len(ip) > 0 {
		ipAddress = ip[0]
	}

	auth := md.Get("authorization")
	if len(auth) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization")
	}

	newAccessToken, _, err := s.authService.Refresh(ctx, auth[0], userAgent, ipAddress)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "refresh failed: %v", err)
	}

	return &identity.RefreshResponse{AccessToken: newAccessToken}, nil
}

// ChangePassword - смена пароля
func (s *IdentityServer) ChangePassword(ctx context.Context, req *identity.ChangePasswordRequest) (*emptypb.Empty, error) {
	if err := s.authService.ChangePassword(ctx, req.UserId, req.NewPassword); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to change password: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// CreateProfile - создание профиля
func (s *IdentityServer) CreateProfile(ctx context.Context, req *identity.CreateProfileRequest) (*identity.Profile, error) {
	createReq := &core.CreateProfileRequest{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		PositionID: req.PositionId,
		Email:      req.Email,
		Login:      req.Login,
		Password:   req.Password,
		Role:       req.Role,
	}
	if req.DepartmentId != nil {
		createReq.DepartmentID = req.DepartmentId
	}

	profile, err := s.profileService.CreateProfile(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create profile: %v", err)
	}

	return profileToProto(profile), nil
}

// GetProfile - получение профиля
func (s *IdentityServer) GetProfile(ctx context.Context, req *identity.GetProfileRequest) (*identity.Profile, error) {
	profile, err := s.profileService.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "profile not found: %v", err)
	}
	return profileToProto(profile), nil
}

// ListProfiles - список профилей
func (s *IdentityServer) ListProfiles(ctx context.Context, req *identity.ListProfilesRequest) (*identity.ListProfilesResponse, error) {
	filter := &core.ListProfilesFilter{
		PageSize:     int(req.PageSize),
		PageNumber:   int(req.PageNumber),
		DepartmentID: req.DepartmentId,
		PositionID:   req.PositionId,
	}
	profiles, total, err := s.profileService.ListProfiles(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list profiles: %v", err)
	}

	protoProfiles := make([]*identity.Profile, len(profiles))
	for i, p := range profiles {
		protoProfiles[i] = profileToProto(p)
	}

	return &identity.ListProfilesResponse{
		Profiles:   protoProfiles,
		TotalCount: int32(total),
	}, nil
}

// UpdateProfile - обновление профиля
func (s *IdentityServer) UpdateProfile(ctx context.Context, req *identity.UpdateProfileRequest) (*identity.Profile, error) {
	updateReq := &core.UpdateProfileRequest{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PositionID:   req.PositionId,
		DepartmentID: req.DepartmentId,
		Email:        req.Email,
		AvatarURL:    req.AvatarUrl,
	}

	profile, err := s.profileService.UpdateProfile(ctx, req.UserId, updateReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update profile: %v", err)
	}

	return profileToProto(profile), nil
}

// ChangeUserStatus - изменение статуса пользователя
func (s *IdentityServer) ChangeUserStatus(ctx context.Context, req *identity.ChangeUserStatusRequest) (*emptypb.Empty, error) {
	if err := s.profileService.ChangeUserStatus(ctx, req.UserId, req.IsActive); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to change user status: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// CreateDepartment - создание отдела
func (s *IdentityServer) CreateDepartment(ctx context.Context, req *identity.CreateDepartmentRequest) (*identity.Department, error) {
	dept, err := s.profileService.CreateDepartment(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create department: %v", err)
	}
	return &identity.Department{Id: dept.ID, Name: dept.Name}, nil
}

// GetDepartment - получение отдела
func (s *IdentityServer) GetDepartment(ctx context.Context, req *identity.GetDepartmentRequest) (*identity.Department, error) {
	dept, err := s.profileService.GetDepartment(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "department not found: %v", err)
	}
	return &identity.Department{Id: dept.ID, Name: dept.Name}, nil
}

// ListDepartments - список отделов
func (s *IdentityServer) ListDepartments(ctx context.Context, req *emptypb.Empty) (*identity.ListDepartmentsResponse, error) {
	depts, err := s.profileService.ListDepartments(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list departments: %v", err)
	}
	protoDepts := make([]*identity.Department, len(depts))
	for i, d := range depts {
		protoDepts[i] = &identity.Department{Id: d.ID, Name: d.Name}
	}
	return &identity.ListDepartmentsResponse{Departments: protoDepts}, nil
}

// UpdateDepartment - обновление отдела
func (s *IdentityServer) UpdateDepartment(ctx context.Context, req *identity.UpdateDepartmentRequest) (*identity.Department, error) {
	dept, err := s.profileService.UpdateDepartment(ctx, req.Id, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update department: %v", err)
	}
	return &identity.Department{Id: dept.ID, Name: dept.Name}, nil
}

// DeleteDepartment - удаление отдела
func (s *IdentityServer) DeleteDepartment(ctx context.Context, req *identity.DeleteDepartmentRequest) (*emptypb.Empty, error) {
	if err := s.profileService.DeleteDepartment(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete department: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// CreatePosition - создание должности
func (s *IdentityServer) CreatePosition(ctx context.Context, req *identity.CreatePositionRequest) (*identity.Position, error) {
	pos, err := s.profileService.CreatePosition(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create position: %v", err)
	}
	return &identity.Position{Id: pos.ID, Name: pos.Name}, nil
}

// GetPosition - получение должности
func (s *IdentityServer) GetPosition(ctx context.Context, req *identity.GetPositionRequest) (*identity.Position, error) {
	pos, err := s.profileService.GetPosition(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "position not found: %v", err)
	}
	return &identity.Position{Id: pos.ID, Name: pos.Name}, nil
}

// ListPositions - список должностей
func (s *IdentityServer) ListPositions(ctx context.Context, req *emptypb.Empty) (*identity.ListPositionsResponse, error) {
	positions, err := s.profileService.ListPositions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list positions: %v", err)
	}
	protoPositions := make([]*identity.Position, len(positions))
	for i, p := range positions {
		protoPositions[i] = &identity.Position{Id: p.ID, Name: p.Name}
	}
	return &identity.ListPositionsResponse{Positions: protoPositions}, nil
}

// UpdatePosition - обновление должности
func (s *IdentityServer) UpdatePosition(ctx context.Context, req *identity.UpdatePositionRequest) (*identity.Position, error) {
	pos, err := s.profileService.UpdatePosition(ctx, req.Id, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update position: %v", err)
	}
	return &identity.Position{Id: pos.ID, Name: pos.Name}, nil
}

// DeletePosition - удаление должности
func (s *IdentityServer) DeletePosition(ctx context.Context, req *identity.DeletePositionRequest) (*emptypb.Empty, error) {
	if err := s.profileService.DeletePosition(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete position: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// CreateSkill - создание навыка
func (s *IdentityServer) CreateSkill(ctx context.Context, req *identity.CreateSkillRequest) (*identity.Skill, error) {
	skill, err := s.profileService.CreateSkill(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create skill: %v", err)
	}
	return &identity.Skill{Id: skill.ID, Name: skill.Name}, nil
}

// ListSkills - список навыков
func (s *IdentityServer) ListSkills(ctx context.Context, req *emptypb.Empty) (*identity.ListSkillsResponse, error) {
	skills, err := s.profileService.ListSkills(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list skills: %v", err)
	}
	protoSkills := make([]*identity.Skill, len(skills))
	for i, sk := range skills {
		protoSkills[i] = &identity.Skill{Id: sk.ID, Name: sk.Name}
	}
	return &identity.ListSkillsResponse{Skills: protoSkills}, nil
}

// AddSkillToEmployee - добавление навыка сотруднику
func (s *IdentityServer) AddSkillToEmployee(ctx context.Context, req *identity.AddSkillToEmployeeRequest) (*emptypb.Empty, error) {
	if err := s.profileService.AddSkillToProfile(ctx, req.EmployeeId, req.SkillId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add skill: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// RemoveSkillFromEmployee - удаление навыка у сотрудника
func (s *IdentityServer) RemoveSkillFromEmployee(ctx context.Context, req *identity.RemoveSkillFromEmployeeRequest) (*emptypb.Empty, error) {
	if err := s.profileService.RemoveSkillFromProfile(ctx, req.EmployeeId, req.SkillId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove skill: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// profileToProto - конвертация профиля в proto
func profileToProto(p *repo.Profile) *identity.Profile {
	profile := &identity.Profile{
		Id:         p.ID,
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		PositionId: p.PositionID,
		Email:      p.Email,
		Login:      p.Login,
		Role:       p.Role,
		IsActive:   p.IsActive,
		AvatarUrl:  p.AvatarURL,
		HireDate:   timestamppb.New(p.HireDate),
		CreatedAt:  timestamppb.New(p.CreatedAt),
		UpdatedAt:  timestamppb.New(p.UpdatedAt),
	}
	if p.Department != nil {
		profile.Department = &identity.Department{
			Id:   p.Department.ID,
			Name: p.Department.Name,
		}
	}
	if len(p.Skills) > 0 {
		skills := make([]*identity.Skill, len(p.Skills))
		for i, sk := range p.Skills {
			skills[i] = &identity.Skill{Id: sk.ID, Name: sk.Name}
		}
		profile.Skills = skills
	}
	return profile
}
