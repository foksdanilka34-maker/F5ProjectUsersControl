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
	log.Printf("CreateProfile called: login=%s, email=%s", regProfile.Login, regProfile.Email)
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
		log.Printf("failed to create credentials: %v", authCred)
		return nil, authCred
	}

	if err = tx.Commit(ctx); err != nil {
		log.Printf("transaction not completed: %v", err)
		return nil, err
	}

	fullName := regProfile.FirstName + " " + regProfile.LastName
	var photoURL *string
	if regProfile.AvatarUrl != "" {
		photoURL = &regProfile.AvatarUrl
	}
	if err := l.publisher.PublishEmployeeCreated(ctx, newUser.UserID, fullName, photoURL); err != nil {
		log.Printf("failed to publish employee created event: %v", err)
	}

	log.Printf("Profile created successfully: userID=%s", newUser.UserID)
	return newUser, nil
}

func (l *loginCore) GetProfile(ctx context.Context, userID string) (*emp.Profile, error) {
	log.Printf("GetProfile called: userID=%s", userID)
	if userID == "" {
		return nil, fmt.Errorf("empty user ID")
	}
	getProfile, err := l.employee.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	log.Printf("Profile retrieved successfully: userID=%s", userID)
	return getProfile, nil
}

func (l *loginCore) ListProfile(ctx context.Context, pageSize, pageNum int, departmentID, positionID string) ([]*emp.Profile, error) {
	log.Printf("ListProfile called: pageSize=%d, pageNum=%d, departmentID=%s, positionID=%s", pageSize, pageNum, departmentID, positionID)
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
	log.Printf("Profiles listed successfully: count=%d", len(profiles))
	return profiles, nil
}

func (l *loginCore) UpdateProfile(ctx context.Context, userID string, updProf *emp.UpdateProfile) (*emp.Profile, error) {
	log.Printf("UpdateProfile called: userID=%s", userID)
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	updProfile, err := l.employee.UpdateProfile(ctx, userID, updProf)
	if err != nil {
		return nil, err
	}

	// Publish employee updated event to NATS
	fullName := updProfile.FirstName + " " + updProfile.LastName
	var photoURL *string
	if updProfile.AvatarUrl != "" {
		photoURL = &updProfile.AvatarUrl
	}
	if err := l.publisher.PublishEmployeeUpdated(ctx, updProfile.UserID, fullName, photoURL); err != nil {
		log.Printf("failed to publish employee updated event: %v", err)
		// Don't fail the whole operation, just log the error
	}

	log.Printf("Profile updated successfully: userID=%s", userID)
	return updProfile, nil
}

func (l *loginCore) DeactivateProfile(ctx context.Context, userID string, status bool) error {
	log.Printf("DeactivateProfile called: userID=%s", userID)
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	if err := l.publisher.PublishDeactivateUserCommand(ctx, userID, status); err != nil {
		return fmt.Errorf("failed to publish deactivate user command: %w", err)
	}
	log.Printf("NATS: deactivate command published: userID=%s", userID)
	return nil
}

func (l *loginCore) CreateDepartment(ctx context.Context, name string) (*emp.Department, error) {
	log.Printf("CreateDepartment called: name=%s", name)
	if name == "" {
		return nil, fmt.Errorf("department name cannot be empty")
	}
	department, err := l.employee.CreateDepartment(ctx, name)
	if err != nil {
		log.Printf("failed to create department: %v", err)
		return nil, err
	}
	log.Printf("Department created successfully: id=%s", department.ID)
	return department, nil
}

func (l *loginCore) GetDepartment(ctx context.Context, id string) (*emp.Department, error) {
	log.Printf("GetDepartment called: id=%s", id)
	if id == "" {
		return nil, fmt.Errorf("department ID cannot be empty")
	}
	department, err := l.employee.GetDepartment(ctx, id)
	if err != nil {
		log.Printf("failed to get department: %v", err)
		return nil, err
	}
	log.Printf("Department retrieved successfully: id=%s", id)
	return department, nil
}

func (l *loginCore) ListDepartments(ctx context.Context) ([]*emp.Department, error) {
	log.Printf("ListDepartments called")
	departments, err := l.employee.ListDepartments(ctx)
	if err != nil {
		log.Printf("failed to list departments: %v", err)
		return nil, err
	}
	log.Printf("Departments listed successfully: count=%d", len(departments))
	return departments, nil
}

func (l *loginCore) UpdateDepartment(ctx context.Context, id, name string) (*emp.Department, error) {
	log.Printf("UpdateDepartment called: id=%s, name=%s", id, name)
	if id == "" {
		return nil, fmt.Errorf("department ID cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("department name cannot be empty")
	}
	department, err := l.employee.UpdateDepartment(ctx, id, name)
	if err != nil {
		log.Printf("failed to update department: %v", err)
		return nil, err
	}
	log.Printf("Department updated successfully: id=%s", id)
	return department, nil
}

func (l *loginCore) DeleteDepartment(ctx context.Context, id string) error {
	log.Printf("DeleteDepartment called: id=%s", id)
	if id == "" {
		return fmt.Errorf("department ID cannot be empty")
	}
	err := l.employee.DeleteDepartment(ctx, id)
	if err != nil {
		log.Printf("failed to delete department: %v", err)
		return err
	}
	log.Printf("Department deleted successfully: id=%s", id)
	return nil
}

func (l *loginCore) CreatePosition(ctx context.Context, name string) (*emp.Position, error) {
	log.Printf("CreatePosition called: name=%s", name)
	if name == "" {
		return nil, fmt.Errorf("position name cannot be empty")
	}
	position, err := l.employee.CreatePosition(ctx, name)
	if err != nil {
		log.Printf("failed to create position: %v", err)
		return nil, err
	}
	log.Printf("Position created successfully: id=%s", position.ID)
	return position, nil
}

func (l *loginCore) GetPosition(ctx context.Context, id string) (*emp.Position, error) {
	log.Printf("GetPosition called: id=%s", id)
	if id == "" {
		return nil, fmt.Errorf("position ID cannot be empty")
	}
	position, err := l.employee.GetPosition(ctx, id)
	if err != nil {
		log.Printf("failed to get position: %v", err)
		return nil, err
	}
	log.Printf("Position retrieved successfully: id=%s", id)
	return position, nil
}

func (l *loginCore) ListPositions(ctx context.Context) ([]*emp.Position, error) {
	log.Printf("ListPositions called")
	positions, err := l.employee.ListPositions(ctx)
	if err != nil {
		log.Printf("failed to list positions: %v", err)
		return nil, err
	}
	log.Printf("Positions listed successfully: count=%d", len(positions))
	return positions, nil
}

func (l *loginCore) UpdatePosition(ctx context.Context, id, name string) (*emp.Position, error) {
	log.Printf("UpdatePosition called: id=%s, name=%s", id, name)
	if id == "" {
		return nil, fmt.Errorf("position ID cannot be empty")
	}
	if name == "" {
		return nil, fmt.Errorf("position name cannot be empty")
	}
	position, err := l.employee.UpdatePosition(ctx, id, name)
	if err != nil {
		log.Printf("failed to update position: %v", err)
		return nil, err
	}
	log.Printf("Position updated successfully: id=%s", id)
	return position, nil
}

func (l *loginCore) DeletePosition(ctx context.Context, id string) error {
	log.Printf("DeletePosition called: id=%s", id)
	if id == "" {
		return fmt.Errorf("position ID cannot be empty")
	}
	err := l.employee.DeletePosition(ctx, id)
	if err != nil {
		log.Printf("failed to delete position: %v", err)
		return err
	}
	log.Printf("Position deleted successfully: id=%s", id)
	return nil
}

func (l *loginCore) CreateSkill(ctx context.Context, name string) (*emp.Skill, error) {
	log.Printf("CreateSkill called: name=%s", name)
	if name == "" {
		return nil, fmt.Errorf("skill name cannot be empty")
	}
	skill, err := l.employee.CreateSkill(ctx, name)
	if err != nil {
		log.Printf("failed to create skill: %v", err)
		return nil, err
	}
	log.Printf("Skill created successfully: id=%s", skill.ID)
	return skill, nil
}

func (l *loginCore) ListSkills(ctx context.Context) ([]*emp.Skill, error) {
	log.Printf("ListSkills called")
	skills, err := l.employee.ListSkills(ctx)
	if err != nil {
		log.Printf("failed to list skills: %v", err)
		return nil, err
	}
	log.Printf("Skills listed successfully: count=%d", len(skills))
	return skills, nil
}

func (l *loginCore) AddSkillToEmployee(ctx context.Context, employeeID, skillID string) error {
	log.Printf("AddSkillToEmployee called: employeeID=%s, skillID=%s", employeeID, skillID)
	if employeeID == "" {
		return fmt.Errorf("employee ID cannot be empty")
	}
	if skillID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}
	err := l.employee.AddSkillToEmployee(ctx, employeeID, skillID)
	if err != nil {
		log.Printf("failed to add skill to employee: %v", err)
		return err
	}
	log.Printf("Skill added to employee successfully: employeeID=%s, skillID=%s", employeeID, skillID)
	return nil
}

func (l *loginCore) RemoveSkillFromEmployee(ctx context.Context, employeeID, skillID string) error {
	log.Printf("RemoveSkillFromEmployee called: employeeID=%s, skillID=%s", employeeID, skillID)
	if employeeID == "" {
		return fmt.Errorf("employee ID cannot be empty")
	}
	if skillID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}
	err := l.employee.RemoveSkillFromEmployee(ctx, employeeID, skillID)
	if err != nil {
		log.Printf("failed to remove skill from employee: %v", err)
		return err
	}
	log.Printf("Skill removed from employee successfully: employeeID=%s, skillID=%s", employeeID, skillID)
	return nil
}
