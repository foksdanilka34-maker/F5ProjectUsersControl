package employee

import (
	"context"
	"fmt"
	"log"

	emp "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	authClient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client"
	
	natsclient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client/nats"
	empl "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/repo"
)

const maxPageSize = 100

type loginCore struct {
	employee  empl.EmployeeStorage
	client    authClient.Client
	publisher *natsclient.Publisher
}

type CoreLogic interface {
	CreateProfile(ctx context.Context, regProfile *emp.RegisterData) (*emp.Profile, error)
	GetProfile(ctx context.Context, userID string) (*emp.Profile, error)
	ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error)
	UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error)
	DeactivateProfile(ctx context.Context, userID string) error
}

func NewCore(employee empl.EmployeeStorage, client authClient.Client, publisher *natsclient.Publisher) CoreLogic {
	return &loginCore{
		employee:  employee,
		client:    client,
		publisher: publisher,
	}
}

func (l *loginCore) CreateProfile(ctx context.Context, regProfile *emp.RegisterData) (*emp.Profile, error) {
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
	newUser, err := l.employee.CreateProfile(ctx, regProfile)
	if err != nil {
		return nil, err
	}
	authCred := l.client.CreateCredentials(ctx, newUser.UserID, newUser.Login, newUser.Password, newUser.Role)

	if authCred != nil {
		return nil, err
	}
	
	if err = tx.Commit(ctx); err != nil {
		log.Printf("Transaction is not completed, error %v", err)
		return nil, err
	}

	return newUser, nil
}


func (l *loginCore) GetProfile(ctx context.Context, userID string) (*emp.Profile, error) {
	if userID == "" {
		return nil, fmt.Errorf("empty user ID")
	}
	getProfile, err := l.employee.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return getProfile, nil
}

func (l *loginCore) ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error) {
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
	return profiles, nil
}

func (l *loginCore) UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	updProfile, err := l.employee.UpdateProfile(ctx, userID, updProf)
	if err != nil {
		return nil, err
	}
	return updProfile, nil
}

func (l *loginCore) DeactivateProfile(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if err := l.publisher.PublishDeactivateUserCommand(ctx, userID); err != nil {
		return fmt.Errorf("failed to publish deactivate user command: %w", err)
	}
	log.Printf("NATS: Deactivate command for user %s published successfully", userID)
	return nil
}
