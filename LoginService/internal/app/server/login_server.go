package auth

import (
	"context"
	"errors"
	"strings"

	core "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/core"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/storage"
	auth "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/login_service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	auth.UnimplementedLoginServiceServer
	core core.CoreLogic
}

func NewAuthServer(core core.CoreLogic) *Server {
	return &Server{
		core: core,
	}
}

func (s *Server) Register(gRPC *grpc.Server) {
	auth.RegisterLoginServiceServer(gRPC, s)
}

func (s *Server) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	storage.Logger.Info("Login server called", zap.String("login", req.GetLogin()), zap.String("userAgent", req.GetUserAgent()), zap.String("ipAddress", req.GetIpAddress()))
	if req.GetLogin() == "" {
		storage.Logger.Warn("Login: login is required")
		return nil, status.Error(codes.InvalidArgument, "login is required")
	}
	if req.GetPassword() == "" {
		storage.Logger.Warn("Login: password is required")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	accessToken, refreshToken, err := s.core.Login(ctx, req.GetLogin(), req.GetPassword(),
		req.GetUserAgent(), req.GetIpAddress())
	if err != nil {
		storage.Logger.Error("Login failed", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid login or password")
	}
	setRefreshTokenInMetadata(ctx, refreshToken)
	storage.Logger.Info("Login server successful", zap.String("login", req.GetLogin()))
	return &auth.LoginResponse{AccessToken: accessToken}, nil
}

func (s *Server) Refresh(ctx context.Context, req *emptypb.Empty) (*auth.RefreshResponse, error) {
	storage.Logger.Info("Refresh server called")
	refreshToken, err := getRefreshTokenFromMetadata(ctx)
	if err != nil {
		storage.Logger.Error("Refresh: error getting refresh token", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	userAgent, ipAddress := getUserAgentAndIPFromMetadata(ctx)

	newAccessToken, newRefreshToken, err := s.core.Refresh(ctx, refreshToken, userAgent, ipAddress)
	if err != nil {
		storage.Logger.Error("Refresh failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "refresh failed")
	}

	setRefreshTokenInMetadata(ctx, newRefreshToken)
	storage.Logger.Info("Refresh server successful")
	return &auth.RefreshResponse{
		AccessToken: newAccessToken,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	storage.Logger.Info("Logout server called")
	refreshToken, err := getRefreshTokenFromMetadata(ctx)
	if err != nil {
		storage.Logger.Error("Logout: error getting authorization header", zap.Error(err))
		return &emptypb.Empty{}, nil
	}

	if err := s.core.Logout(ctx, refreshToken); err != nil {
		storage.Logger.Error("Logout: error during logout function", zap.Error(err))
	}

	clearRefreshTokenInMetadata(ctx)
	storage.Logger.Info("Logout server successful")
	return &emptypb.Empty{}, nil
}

func (s *Server) CreateCredentials(ctx context.Context, req *auth.CreateCredentialsRequest) (*emptypb.Empty, error) {
	storage.Logger.Info("CreateCredentials server called", zap.String("userID", req.GetUserId()), zap.String("login", req.GetLogin()), zap.String("role", req.GetRole()))
	if req.GetUserId() == "" {
		storage.Logger.Warn("CreateCredentials: user_id is required")
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetLogin() == "" {
		storage.Logger.Warn("CreateCredentials: login is required")
		return nil, status.Error(codes.InvalidArgument, "login is required")
	}
	if req.GetPassword() == "" {
		storage.Logger.Warn("CreateCredentials: password is required")
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if req.GetRole() == "" {
		storage.Logger.Warn("CreateCredentials: role is required")
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}

	err := s.core.CreateCredentials(ctx, req.GetUserId(), req.GetLogin(), req.GetPassword(), req.GetRole())
	if err != nil {
		storage.Logger.Error("CreateCredentials failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create credentials")
	}
	storage.Logger.Info("CreateCredentials server successful", zap.String("userID", req.GetUserId()), zap.String("login", req.GetLogin()))
	return &emptypb.Empty{}, nil
}

func (s *Server) ChangeUserStatus(ctx context.Context, req *auth.ChangeUserStatusRequest) (*emptypb.Empty, error) {
	storage.Logger.Info("ChangeUserStatus server called", zap.String("userID", req.GetUserId()), zap.Bool("isActive", req.GetIsActive()))
	if req.GetUserId() == "" {
		storage.Logger.Warn("ChangeUserStatus: user_id is required")
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.core.ChangeUserStatus(ctx, req.GetUserId(), req.GetIsActive())
	if err != nil {
		storage.Logger.Error("ChangeUserStatus failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to change user status")
	}
	storage.Logger.Info("ChangeUserStatus server successful", zap.String("userID", req.GetUserId()), zap.Bool("isActive", req.GetIsActive()))
	return &emptypb.Empty{}, nil
}

func (s *Server) ChangePassword(ctx context.Context, req *auth.ChangePasswordRequest) (*emptypb.Empty, error) {
	storage.Logger.Info("ChangePassword server called", zap.String("userID", req.GetUserId()))
	if req.GetUserId() == "" {
		storage.Logger.Warn("ChangePassword: user_id is required")
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetNewPassword() == "" {
		storage.Logger.Warn("ChangePassword: new_password is required")
		return nil, status.Error(codes.InvalidArgument, "new_password is required")
	}

	err := s.core.ChangePassword(ctx, req.GetUserId(), req.GetNewPassword())
	if err != nil {
		storage.Logger.Error("ChangePassword failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to change password")
	}
	storage.Logger.Info("ChangePassword server successful", zap.String("userID", req.GetUserId()))
	return &emptypb.Empty{}, nil
}

func setRefreshTokenInMetadata(ctx context.Context, token string) {
	grpc.SetHeader(ctx, metadata.Pairs("authorization", "Bearer "+token))
}

func clearRefreshTokenInMetadata(ctx context.Context) {
	grpc.SetHeader(ctx, metadata.Pairs("authorization", ""))
}

func getRefreshTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata is not provided")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", errors.New("authorization header is not provided")
	}

	authHeader := strings.TrimSpace(values[0])
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("invalid authorization header format, expected 'Bearer <token>'")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("refresh token is empty")
	}

	return token, nil
}

func getUserAgentAndIPFromMetadata(ctx context.Context) (userAgent, ipAddress string) {
	userAgent = "unknown"
	ipAddress = "unknown"

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}

	uaValues := md.Get("x-user-agent")
	if len(uaValues) > 0 {
		userAgent = uaValues[0]
	}

	ipValues := md.Get("x-real-ip")
	if len(ipValues) > 0 {
		ipAddress = ipValues[0]
	}

	return userAgent, ipAddress
}
