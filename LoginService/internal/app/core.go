package app

import "context"


type LoginCore interface {
	Login(ctx context.Context, login, password string) (accesToken, refreshToken string, userID string, err error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, oldRefreshToken string) (accesToken, newRefreshToken string, err error)
	CreateCredentials(ctx context.Context, userID, login, password, role string) error
}