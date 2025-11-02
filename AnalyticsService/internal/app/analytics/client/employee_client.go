package client

import (
	"context"
	"log"

	employeesv1 "github.com/foksdanilka34-maker/F5ProjectUsersControl/gen/go/employee_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type EmployeeClient struct {
	client employeesv1.EmployeeServiceClient
	conn   *grpc.ClientConn
}

func NewEmployeeClient(addr string) (*EmployeeClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to dial employee service: %v", err)
		return nil, err
	}

	client := employeesv1.NewEmployeeServiceClient(conn)

	return &EmployeeClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *EmployeeClient) GetProfile(ctx context.Context, userID string) (*employeesv1.Profile, error) {
	req := &employeesv1.GetProfileRequest{
		UserId: userID,
	}

	profile, err := c.client.GetProfile(ctx, req)
	if err != nil {
		log.Printf("failed to get profile: %v", err)
		return nil, err
	}

	return profile, nil
}

func (c *EmployeeClient) ListProfiles(ctx context.Context, pageSize, pageNumber int32, departmentID, positionID string) ([]*employeesv1.Profile, int32, error) {
	req := &employeesv1.ListProfilesRequest{
		PageSize:     pageSize,
		PageNumber:   pageNumber,
		DepartmentId: departmentID,
		PositionId:   positionID,
	}

	resp, err := c.client.ListProfiles(ctx, req)
	if err != nil {
		log.Printf("failed to list profiles: %v", err)
		return nil, 0, err
	}

	return resp.Profiles, resp.TotalCount, nil
}

func (c *EmployeeClient) GetDepartment(ctx context.Context, departmentID string) (*employeesv1.Department, error) {
	req := &employeesv1.GetDepartmentRequest{
		Id: departmentID,
	}

	dept, err := c.client.GetDepartment(ctx, req)
	if err != nil {
		log.Printf("failed to get department: %v", err)
		return nil, err
	}

	return dept, nil
}

func (c *EmployeeClient) ListDepartments(ctx context.Context) ([]*employeesv1.Department, error) {
	resp, err := c.client.ListDepartments(ctx, nil)
	if err != nil {
		log.Printf("failed to list departments: %v", err)
		return nil, err
	}

	return resp.Departments, nil
}

func (c *EmployeeClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
