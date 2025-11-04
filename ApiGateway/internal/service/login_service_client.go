package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	auth "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/login_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type LoginServiceClient struct {
	client    auth.LoginServiceClient
	jwtSecret string
	conn      *grpc.ClientConn
}

func NewLoginServiceClient(host, port, jwtSecret string) *LoginServiceClient {
	addr := fmt.Sprintf("%s:%s", host, port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024),
		),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("failed to connect to login service: %v", err)
	}

	return &LoginServiceClient{
		client:    auth.NewLoginServiceClient(conn),
		jwtSecret: jwtSecret,
		conn:      conn,
	}
}

func (c *LoginServiceClient) Login(ctx context.Context, login, password, userAgent, ipAddress string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &auth.LoginRequest{
		Login:     login,
		Password:  password,
		UserAgent: userAgent,
		IpAddress: ipAddress,
	}

	header := metadata.MD{}
	resp, err := c.client.Login(ctx, req, grpc.Header(&header))
	if err != nil {
		log.Printf("Login error: %v", err)
		return "", "", err
	}

	refreshToken := extractRefreshToken(header)

	return resp.GetAccessToken(), refreshToken, nil
}

func (c *LoginServiceClient) Logout(ctx context.Context, refreshToken string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if refreshToken == "" {
		return fmt.Errorf("refresh token is empty")
	}

	md := map[string][]string{
		"authorization": {"Bearer " + refreshToken},
	}
	ctx = addMetadata(ctx, md)

	_, err := c.client.Logout(ctx, &emptypb.Empty{})
	if err != nil {
		log.Printf("Logout error: %v", err)
		return err
	}

	return nil
}

func (c *LoginServiceClient) Refresh(ctx context.Context, refreshToken, userAgent, ipAddress string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if refreshToken == "" {
		return "", "", fmt.Errorf("refresh token is empty")
	}

	md := map[string][]string{
		"authorization": {"Bearer " + refreshToken},
		"x-user-agent":  {userAgent},
		"x-real-ip":     {ipAddress},
	}
	ctx = addMetadata(ctx, md)

	header := metadata.MD{}
	resp, err := c.client.Refresh(ctx, &emptypb.Empty{}, grpc.Header(&header))
	if err != nil {
		log.Printf("Refresh error: %v", err)
		return "", "", err
	}

	newRefreshToken := extractRefreshToken(header)

	return resp.GetAccessToken(), newRefreshToken, nil
}

func (c *LoginServiceClient) CreateCredentials(ctx context.Context, userID, login, password, role string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &auth.CreateCredentialsRequest{
		UserId:   userID,
		Login:    login,
		Password: password,
		Role:     role,
	}

	_, err := c.client.CreateCredentials(ctx, req)
	if err != nil {
		log.Printf("CreateCredentials error: %v", err)
		return err
	}

	return nil
}

func (c *LoginServiceClient) ChangeUserStatus(ctx context.Context, userID string, isActive bool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &auth.ChangeUserStatusRequest{
		UserId:   userID,
		IsActive: isActive,
	}

	_, err := c.client.ChangeUserStatus(ctx, req)
	if err != nil {
		log.Printf("ChangeUserStatus error: %v", err)
		return err
	}

	return nil
}

func (c *LoginServiceClient) ChangePassword(ctx context.Context, userID, newPassword string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &auth.ChangePasswordRequest{
		UserId:      userID,
		NewPassword: newPassword,
	}

	_, err := c.client.ChangePassword(ctx, req)
	if err != nil {
		log.Printf("ChangePassword error: %v", err)
		return err
	}

	return nil
}

func (c *LoginServiceClient) Close() error {
	return c.conn.Close()
}

func addMetadata(ctx context.Context, values map[string][]string) context.Context {
	if len(values) == 0 {
		return ctx
	}

	md := metadata.MD{}
	for key, vals := range values {
		lowerKey := strings.ToLower(key)
		md[lowerKey] = append(md[lowerKey], vals...)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

func extractRefreshToken(md metadata.MD) string {
	if len(md) == 0 {
		return ""
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return ""
	}

	token := strings.TrimSpace(values[0])
	if token == "" {
		return ""
	}

	parts := strings.SplitN(token, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}

	return token
}
