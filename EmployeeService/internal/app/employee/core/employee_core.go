package employee

import (
	"context"
	"fmt"

	emp "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee"
	authClient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client"

	natsclient "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/client/nats"
	empl "github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app/employee/repo"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/EmployeeService/internal/app"
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

	CreateDepartment(ctx context.Context, name string) (*emp.Department, error)
	GetDepartment(ctx context.Context, id string) (*emp.Department, error)
	ListDepartments(ctx context.Context) ([]*emp.Department, error)
	UpdateDepartment(ctx context.Context, id, name string) (*emp.Department, error)
	DeleteDepartment(ctx context.Context, id string) error

	CreatePosition(ctx context.Context, name string) (*emp.Position, error)
	GetPosition(ctx context.Context, id string) (*emp.Position, error)
	ListPositions(ctx context.Context) ([]*emp.Position, error)
	UpdatePosition(ctx context.Context, id, name string) (*emp.Position, error)
	DeletePosition(ctx context.Context, id string) error

	CreateSkill(ctx context.Context, name string) (*emp.Skill, error)
	ListSkills(ctx context.Context) ([]*emp.Skill, error)
	AddSkillToEmployee(ctx context.Context, employeeID, skillID string) error
	RemoveSkillFromEmployee(ctx context.Context, employeeID, skillID string) error
}

func NewCore(employee empl.EmployeeStorage, client *authClient.Client, publisher *natsclient.Publisher) CoreLogic {
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
	app.Logger.Info("CreateProfile called", zap.String("login", regProfile.Login), zap.String("email", regProfile.Email))
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
		app.Logger.Error("failed to create credentials", zap.Error(authCred))
		return nil, authCred
	}

	if err = tx.Commit(ctx); err != nil {
		app.Logger.Error("transaction not completed", zap.Error(err))
		return nil, err
	}

	app.Logger.Info("Profile created successfully", zap.String("userID", newUser.UserID))
	return newUser, nil
}

func (l *loginCore) GetProfile(ctx context.Context, userID string) (*emp.Profile, error) {
	app.Logger.Info("GetProfile called", zap.String("userID", userID))
	if userID == "" {
		return nil, fmt.Errorf("empty user ID")
	}
	getProfile, err := l.employee.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	app.Logger.Info("Profile retrieved successfully", zap.String("userID", userID))
	return getProfile, nil
}

func (l *loginCore) ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error) {
	app.Logger.Info("ListProfile called", zap.Int("pageSize", pageSize), zap.Int("pageNum", pageNum), zap.String("departmentID", departmentID), zap.String("positionID", positionID))
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
	app.Logger.Info("Profiles listed successfully", zap.Int("count", len(profiles)))
	return profiles, nil
}

func (l *loginCore) UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error) {
	app.Logger.Info("UpdateProfile called", zap.String("userID", userID))
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	updProfile, err := l.employee.UpdateProfile(ctx, userID, updProf)
	if err != nil {
		return nil, err
	}
	app.Logger.Info("Profile updated successfully", zap.String("userID", userID))
	return updProfile, nil
}

func (l *loginCore) DeactivateProfile(ctx context.Context, userID string, status bool) error {
	app.Logger.Info("DeactivateProfile called", zap.String("userID", userID))
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if err := l.publisher.PublishDeactivateUserCommand(ctx, userID, status); err != nil {
		return fmt.Errorf("failed to publish deactivate user command: %w", err)
	}
	app.Logger.Info("NATS: deactivate command published", zap.String("userID", userID))
	return nil
}

func (l *loginCore) CreateDepartment(ctx context.Context, name string) (*emp.Department, error) {
	app.Logger.Info("CreateDepartment called", zap.String("name", name))
	if name == "" {
		return nil, fmt.Errorf("department name cannot be empty")
	}
	department, err := l.employee.CreateDepartment(ctx, name)
	if err != nil {
		app.Logger.Error("failed to create department", zap.Error(err))
		return nil, err
	}
	app.Logger.Info("Department created successfully", zap.String("id", department.ID))
	return department, nil
}

func (l *loginCore) GetDepartment(ctx context.Context, id string) (*emp.Department, error) {
	app.Logger.Info("GetDepartment called", zap.String("id", id))
	if id == "" {
		return nil, fmt.Errorf("department ID cannot be empty")
	}
	department, err := l.employee.GetDepartment(ctx, id)
	if err != nil {
		app.Logger.Error("failed to get department", zap.Error(err))
		return nil, err
	}
	app.Logger.Info("Department retrieved successfully", zap.String("id", id))
	return department, nil
}

func (l *loginCore) ListDepartments(ctx context.Context) ([]*emp.Department, error) {
	app.Logger.Info("ListDepartments called")
	departments, err := l.employee.ListDepartments(ctx)
	if err != nil {
		app.Logger.Error("failed to list departments", zap.Error(err))
		return nil, err
	}
	app.Logger.Info("Departments listed successfully", zap.Int("count", len(departments)))
	return departments, nil
}

func (l *loginCore) UpdateDepartment(ctx context.Context, id, name string) (*emp.Department, error) {
	app.Logger.Info("UpdateDepartment called", zap.String("id", id), zap.String("name", name))
	if id == "" {
		return nil, fmt.Errorf("department ID cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("department name cannot be empty")
	}
	department, err := l.employee.UpdateDepartment(ctx, id, name)
	if err != nil {
		app.Logger.Error("failed to update department", zap.Error(err))
		return nil, err
	}
	app.Logger.Info("Department updated successfully", zap.String("id", id))
	return department, nil
}

func (l *loginCore) DeleteDepartment(ctx context.Context, id string) error {
	app.Logger.Info("DeleteDepartment called", zap.String("id", id))
	if id == "" {
		return fmt.Errorf("department ID cannot be empty")
	}
	err := l.employee.DeleteDepartment(ctx, id)
	if err != nil {
		app.Logger.Error("failed to delete department", zap.Error(err))
		return err
	}
	app.Logger.Info("Department deleted successfully", zap.String("id", id))
	return nil
}

func (l *loginCore) CreatePosition(ctx context.Context, name string) (*emp.Position, error) {
	app.Logger.Info("CreatePosition called", zap.String("name", name))
	if name == "" {
		return nil, fmt.Errorf("position name cannot be empty")
	}
	position, err := l.employee.CreatePosition(ctx, name)
	if err != nil {
		app.Logger.Error("failed to create position", zap.Error(err))
		return nil, err
	}
	app.Logger.Info("Position created successfully", zap.String("id", position.ID))
	return position, nil
}

func (l *loginCore) GetPosition(ctx context.Context, id string) (*emp.Position, error) {
	app.Logger.Info("GetPosition called", zap.String("id", id))
	if id == "" {
		return nil, fmt.Errorf("position ID cannot be empty")
	}
	position, err := l.employee.GetPosition(ctx, id)
	if err != nil {
		app.Logger.Error("failed to get position", zap.Error(err))
		return nil, err
	}
	app.Logger.Info("Position retrieved successfully", zap.String("id", id))
	return position, nil
}

func (l *loginCore) ListPositions(ctx context.Context) ([]*emp.Position, error) {
	app.Logger.Info("ListPositions called")
	positions, err := l.employee.ListPositions(ctx)
	if err != nil {
		app.Logger.Error("failed to list positions", zap.Error(err))
		return nil, err
	}
	app.Logger.Info("Positions listed successfully", zap.Int("count", len(positions)))
	return positions, nil
}

func (l *loginCore) UpdatePosition(ctx context.Context, id, name string) (*emp.Position, error) {
	app.Logger.Info("UpdatePosition called", zap.String("id", id), zap.String("name", name))
	if id == "" {
		return nil, fmt.Errorf("position ID cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("position name cannot be empty")
	}
	position, err := l.employee.UpdatePosition(ctx, id, name)
	if err != nil {
		app.Logger.Error("failed to update position", zap.Error(err))
		return nil, err
	}
	app.Logger.Info("Position updated successfully", zap.String("id", id))
	return position, nil
}

func (l *loginCore) DeletePosition(ctx context.Context, id string) error {
	app.Logger.Info("DeletePosition called", zap.String("id", id))
	if id == "" {
		return fmt.Errorf("position ID cannot be empty")
	}
	err := l.employee.DeletePosition(ctx, id)
	if err != nil {
		app.Logger.Error("failed to delete position", zap.Error(err))
		return err
	}
	app.Logger.Info("Position deleted successfully", zap.String("id", id))
	return nil
}

func (l *loginCore) CreateSkill(ctx context.Context, name string) (*emp.Skill, error) {
	app.Logger.Info("CreateSkill called", zap.String("name", name))
	if name == "" {
		return nil, fmt.Errorf("skill name cannot be empty")
	}
	skill, err := l.employee.CreateSkill(ctx, name)
	if err != nil {
		app.Logger.Error("failed to create skill", zap.Error(err))
		return nil, err
	}
	app.Logger.Info("Skill created successfully", zap.String("id", skill.ID))
	return skill, nil
}

func (l *loginCore) ListSkills(ctx context.Context) ([]*emp.Skill, error) {
	app.Logger.Info("ListSkills called")
	skills, err := l.employee.ListSkills(ctx)
	if err != nil {
		app.Logger.Error("failed to list skills", zap.Error(err))
		return nil, err
	}
	app.Logger.Info("Skills listed successfully", zap.Int("count", len(skills)))
	return skills, nil
}

func (l *loginCore) AddSkillToEmployee(ctx context.Context, employeeID, skillID string) error {
	app.Logger.Info("AddSkillToEmployee called", zap.String("employeeID", employeeID), zap.String("skillID", skillID))
	if employeeID == "" {
		return fmt.Errorf("employee ID cannot be empty")
	}
	if skillID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}
	err := l.employee.AddSkillToEmployee(ctx, employeeID, skillID)
	if err != nil {
		app.Logger.Error("failed to add skill to employee", zap.Error(err))
		return err
	}
	app.Logger.Info("Skill added to employee successfully", zap.String("employeeID", employeeID), zap.String("skillID", skillID))
	return nil
}

func (l *loginCore) RemoveSkillFromEmployee(ctx context.Context, employeeID, skillID string) error {
	app.Logger.Info("RemoveSkillFromEmployee called", zap.String("employeeID", employeeID), zap.String("skillID", skillID))
	if employeeID == "" {
		return fmt.Errorf("employee ID cannot be empty")
	}
	if skillID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}
	err := l.employee.RemoveSkillFromEmployee(ctx, employeeID, skillID)
	if err != nil {
		app.Logger.Error("failed to remove skill from employee", zap.Error(err))
		return err
	}
	app.Logger.Info("Skill removed from employee successfully", zap.String("employeeID", employeeID), zap.String("skillID", skillID))
	return nil
}
