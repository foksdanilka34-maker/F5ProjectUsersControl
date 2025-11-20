package service

import (
	"context"
	"time"

	employeepb "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/employee_service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EmployeeServiceClient struct {
	conn   *grpc.ClientConn
	client employeepb.EmployeeServiceClient
}

func NewEmployeeServiceClient(host, port string) *EmployeeServiceClient {
	conn := dialGRPC("employee service", host, port, 10*time.Second)
	return &EmployeeServiceClient{
		conn:   conn,
		client: employeepb.NewEmployeeServiceClient(conn),
	}
}

func (c *EmployeeServiceClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Profile methods
func (c *EmployeeServiceClient) CreateProfile(ctx context.Context, req *employeepb.CreateProfileRequest) (*employeepb.Profile, error) {
	return c.client.CreateProfile(ctx, req)
}

func (c *EmployeeServiceClient) GetProfile(ctx context.Context, userID string) (*employeepb.Profile, error) {
	return c.client.GetProfile(ctx, &employeepb.GetProfileRequest{
		UserId: userID,
	})
}

func (c *EmployeeServiceClient) ListProfiles(ctx context.Context, pageSize, pageNumber int32, departmentID, positionID *string) (*employeepb.ListProfilesResponse, error) {
	return c.client.ListProfiles(ctx, &employeepb.ListProfilesRequest{
		PageSize:     pageSize,
		PageNumber:   pageNumber,
		DepartmentId: departmentID,
		PositionId:   positionID,
	})
}

func (c *EmployeeServiceClient) UpdateProfile(ctx context.Context, req *employeepb.UpdateProfileRequest) (*employeepb.Profile, error) {
	return c.client.UpdateProfile(ctx, req)
}

func (c *EmployeeServiceClient) ChangeUserStatusProfile(ctx context.Context, userID string, status bool) error {
	_, err := c.client.ChangeUserStatusProfile(ctx, &employeepb.DeactivateProfileRequest{
		UserId: userID,
		Status: status,
	})
	return err
}

func (c *EmployeeServiceClient) CreateDepartment(ctx context.Context, name string) (*employeepb.Department, error) {
	return c.client.CreateDepartment(ctx, &employeepb.CreateDepartmentRequest{
		Name: name,
	})
}

func (c *EmployeeServiceClient) GetDepartment(ctx context.Context, id string) (*employeepb.Department, error) {
	return c.client.GetDepartment(ctx, &employeepb.GetDepartmentRequest{
		Id: id,
	})
}

func (c *EmployeeServiceClient) ListDepartments(ctx context.Context) (*employeepb.ListDepartmentsResponse, error) {
	return c.client.ListDepartments(ctx, &emptypb.Empty{})
}

func (c *EmployeeServiceClient) UpdateDepartment(ctx context.Context, id, name string) (*employeepb.Department, error) {
	return c.client.UpdateDepartment(ctx, &employeepb.UpdateDepartmentRequest{
		Id:   id,
		Name: name,
	})
}

func (c *EmployeeServiceClient) DeleteDepartment(ctx context.Context, id string) error {
	_, err := c.client.DeleteDepartment(ctx, &employeepb.DeleteDepartmentRequest{
		Id: id,
	})
	return err
}

func (c *EmployeeServiceClient) CreatePosition(ctx context.Context, name string) (*employeepb.Position, error) {
	return c.client.CreatePosition(ctx, &employeepb.CreatePositionRequest{
		Name: name,
	})
}

func (c *EmployeeServiceClient) GetPosition(ctx context.Context, id string) (*employeepb.Position, error) {
	return c.client.GetPosition(ctx, &employeepb.GetPositionRequest{
		Id: id,
	})
}

func (c *EmployeeServiceClient) ListPositions(ctx context.Context) (*employeepb.ListPositionsResponse, error) {
	return c.client.ListPositions(ctx, &emptypb.Empty{})
}

func (c *EmployeeServiceClient) UpdatePosition(ctx context.Context, id, name string) (*employeepb.Position, error) {
	return c.client.UpdatePosition(ctx, &employeepb.UpdatePositionRequest{
		Id:   id,
		Name: name,
	})
}

func (c *EmployeeServiceClient) DeletePosition(ctx context.Context, id string) error {
	_, err := c.client.DeletePosition(ctx, &employeepb.DeletePositionRequest{
		Id: id,
	})
	return err
}

func (c *EmployeeServiceClient) CreateSkill(ctx context.Context, name string) (*employeepb.Skill, error) {
	return c.client.CreateSkill(ctx, &employeepb.CreateSkillRequest{
		Name: name,
	})
}

func (c *EmployeeServiceClient) ListSkills(ctx context.Context) (*employeepb.ListSkillsResponse, error) {
	return c.client.ListSkills(ctx, &emptypb.Empty{})
}

func (c *EmployeeServiceClient) AddSkillToEmployee(ctx context.Context, employeeID, skillID string) error {
	_, err := c.client.AddSkillToEmployee(ctx, &employeepb.AddSkillToEmployeeRequest{
		EmployeeId: employeeID,
		SkillId:    skillID,
	})
	return err
}

func (c *EmployeeServiceClient) RemoveSkillFromEmployee(ctx context.Context, employeeID, skillID string) error {
	_, err := c.client.RemoveSkillFromEmployee(ctx, &employeepb.RemoveSkillFromEmployeeRequest{
		EmployeeId: employeeID,
		SkillId:    skillID,
	})
	return err
}

func TimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
