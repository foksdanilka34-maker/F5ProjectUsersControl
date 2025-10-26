package employee

import (
	"context"
	"fmt"

	emp "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	authClient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client"

	natsclient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client/nats"
	empl "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/storage"
	"go.uber.org/zap"
)

const maxPageSize = 100

type loginCore struct {
	employee  empl.EmployeeStorage
	client    *authClient.Client
	publisher *natsclient.Publisher
}

type CoreLogic interface {
	CreateProfile(ctx context.Context, regProfile *emp.RegisterData) (*emp.Profile, error)
	GetProfile(ctx context.Context, userID string) (*emp.Profile, error)
	ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error)
	UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error)
	DeactivateProfile(ctx context.Context, userID string, status bool) error
}

func NewCore(employee empl.EmployeeStorage, client *authClient.Client, publisher *natsclient.Publisher) CoreLogic {
	return &loginCore{
		employee:  employee,
		client:    client,
		publisher: publisher,
	}
}

func (l *loginCore) CreateProfile(ctx context.Context, regProfile *emp.RegisterData) (*emp.Profile, error) {
	storage.Logger.Info("CreateProfile called", zap.String("login", regProfile.Login), zap.String("email", regProfile.Email))
	if regProfile == nil {
		return nil, fmt.Errorf("profile data cannot be nil")
	}
	if regProfile.FirstName == "" || regProfile.LastName == "" || regProfile.Email == "" {
		return nil, fmt.Errorf("first name, last name, and email are required fields")
	}

	tx, err := l.employee.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	newUser, err := l.employee.CreateProfile(ctx, tx, regProfile)
	if err != nil {
		return nil, err
	}
	authCred := l.client.CreateCredentials(ctx, newUser.UserID, regProfile.Login, regProfile.Password, regProfile.Role)
	if authCred != nil {
		storage.Logger.Error("failed to create credentials", zap.Error(authCred))
		return nil, authCred
	}

	if err = tx.Commit(ctx); err != nil {
		storage.Logger.Error("transaction not completed", zap.Error(err))
		return nil, err
	}

	storage.Logger.Info("Profile created successfully", zap.String("userID", newUser.UserID))
	return newUser, nil
}

func (l *loginCore) GetProfile(ctx context.Context, userID string) (*emp.Profile, error) {
	storage.Logger.Info("GetProfile called", zap.String("userID", userID))
	if userID == "" {
		return nil, fmt.Errorf("empty user ID")
	}
	getProfile, err := l.employee.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	storage.Logger.Info("Profile retrieved successfully", zap.String("userID", userID))
	return getProfile, nil
}

func (l *loginCore) ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error) {
	storage.Logger.Info("ListProfile called", zap.Int("pageSize", pageSize), zap.Int("pageNum", pageNum), zap.String("departmentID", departmentID), zap.String("positionID", positionID))
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if pageNum <= 0 {
		pageNum = 1
	}

	profiles, err := l.employee.ListProfile(ctx, pageSize, pageNum, departmentID, positionID)
	if err != nil {
		return nil, err
	}
	storage.Logger.Info("Profiles listed successfully", zap.Int("count", len(profiles)))
	return profiles, nil
}

func (l *loginCore) UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error) {
	storage.Logger.Info("UpdateProfile called", zap.String("userID", userID))
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	updProfile, err := l.employee.UpdateProfile(ctx, userID, updProf)
	if err != nil {
		return nil, err
	}
	storage.Logger.Info("Profile updated successfully", zap.String("userID", userID))
	return updProfile, nil
}

func (l *loginCore) DeactivateProfile(ctx context.Context, userID string, status bool) error {
	storage.Logger.Info("DeactivateProfile called", zap.String("userID", userID))
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if err := l.publisher.PublishDeactivateUserCommand(ctx, userID, status); err != nil {
		return fmt.Errorf("failed to publish deactivate user command: %w", err)
	}
	storage.Logger.Info("NATS: deactivate command published", zap.String("userID", userID))
	return nil
}
