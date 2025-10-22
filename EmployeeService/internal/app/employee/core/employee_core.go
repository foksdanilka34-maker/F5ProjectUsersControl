package employee

import (
	"context"
	"fmt"
	"log"

	emp "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	authClient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client"
	empl "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/repo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type loginCore struct {
	employee empl.EmployeeStorage
	client authClient.Client
}

type CoreLogic interface {
	CreateProfile(ctx context.Context, regProfile *emp.RegisterData) (*emp.Profile, error)
	GetProfile(ctx context.Context, userID string) (*emp.Profile, error)
	ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error)
	UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error)
	DeactivateProfile(ctx context.Context) (error)
}

func NewCore(employee empl.EmployeeStorage, client authClient.Client) CoreLogic {
	return &loginCore{
		employee: employee,
		client: client,
	}
}

func (l *loginCore) CreateProfile(ctx context.Context, regProfile *emp.RegisterData) (*emp.Profile, error) {
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
	profiles, err := l.employee.ListProfile(ctx, pageSize, pageNum, departmentID, positionID)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (l *loginCore) UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error) {
	updProfile, err := l.employee.UpdateProfile(ctx, userID, updProf)
	if err != nil {
		return nil, err
	}
	return updProfile, nil
}

func (l *loginCore) DeactivateProfile(ctx context.Context) (error) {
	
}
	
	
