package employee

import (
	"context"

	loginClient "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/login_service"
	"google.golang.org/grpc"
)

type AuthService interface {
	CreateCredentials(ctx context.Context, userID, login, password, role string) error
}

type Client struct {
	client loginClient.LoginServiceClient
}

func New(c *grpc.ClientConn) (*Client, error) {
	client := loginClient.NewLoginServiceClient(c)

	return &Client{
		client: client,
	}, nil
}

func (c *Client) CreateCredentials(ctx context.Context, userID, login, password, role string) error {
	request := &loginClient.CreateCredentialsRequest{
		UserId: userID,
		Login: login,
		Password: password,
		Role: role,
	}

	_, err := c.client.CreateCredentials(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

