package auth

import (
	"context"
	"errors"
	"log"
	"strings"

	auth "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/login_service"
	core "github.com/foksdanilka34-maker/F5ProjectUsersControl/LoginService/internal/app/core"
	
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
	if req.GetLogin() == "" {
		return nil, status.Error(codes.InvalidArgument, "login is required")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	accessToken, refreshToken, err := s.core.Login(ctx, req.GetLogin(), req.GetPassword(),
		req.GetUserAgent(), req.GetIpAddress())	
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid login or password")
	}
	setRefreshTokenInMetadata(ctx, refreshToken)

	return &auth.LoginResponse{AccessToken: accessToken}, nil
}

func (s *Server) Refresh(ctx context.Context, req *emptypb.Empty) (*auth.RefreshResponse, error) {
	refreshToken, err := getRefreshTokenFromMetadata(ctx)
	if err != nil {
		log.Printf("error getting refresh token")
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	
	userAgent, ipAddress := getUserAgentAndIPFromMetadata(ctx)

	newAccessToken, newRefreshToken, err := s.core.Refresh(ctx, refreshToken, userAgent, ipAddress)
	if err != nil {
		return nil, status.Error(codes.Internal, "refresh failed")
	}

	setRefreshTokenInMetadata(ctx, newRefreshToken)

	return &auth.RefreshResponse{
		AccessToken: newAccessToken,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	refreshToken, err := getRefreshTokenFromMetadata(ctx)
	if err != nil {
		log.Printf("error getting authorization header: %v", err)
		return &emptypb.Empty{}, nil
	}

	if err := s.core.Logout(ctx, refreshToken); err != nil {
		log.Printf("error during logout function: %v", err)
	}

	clearRefreshTokenInMetadata(ctx)

	return &emptypb.Empty{}, nil
}

func (s *Server) CreateCredentials(ctx context.Context, req *auth.CreateCredentialsRequest) (*emptypb.Empty, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetLogin() == "" {
		return nil, status.Error(codes.InvalidArgument, "login is required")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if req.GetRole() == "" {
		return nil, status.Error(codes.InvalidArgument, "role is required")
	}

	err := s.core.CreateCredentials(ctx, req.GetUserId(), req.GetLogin(), req.GetPassword(), req.GetRole())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create credentials")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) ChangeUserStatus(ctx context.Context, req *auth.ChangeUserStatusRequest) (*emptypb.Empty, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.core.ChangeUserStatus(ctx, req.GetUserId(), req.GetIsActive())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change user status")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) ChangePassword(ctx context.Context, req *auth.ChangePasswordRequest) (*emptypb.Empty, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "new_password is required")
	}

	err := s.core.ChangePassword(ctx, req.GetUserId(), req.GetNewPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to change password")
	}

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