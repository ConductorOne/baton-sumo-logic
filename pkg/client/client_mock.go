package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type MockClientService struct {
	GetUserByIDFunc        func(ctx context.Context, userId string) (*UserResponse, *v2.RateLimitDescription, error)
	CreateUserFunc         func(ctx context.Context, userRequest UserRequest) (*UserResponse, *v2.RateLimitDescription, error)
	DeleteUserFunc         func(ctx context.Context, userId string) (*v2.RateLimitDescription, error)
	GetUsersFunc           func(ctx context.Context, pageToken *string) ([]UserResponse, *string, *v2.RateLimitDescription, error)
	GetServiceAccountsFunc func(ctx context.Context) ([]ServiceAccountResponse, *v2.RateLimitDescription, error)
	GetRolesFunc           func(ctx context.Context, pageToken *string) ([]RoleResponse, *string, *v2.RateLimitDescription, error)
	GetRoleFunc            func(ctx context.Context, roleId string) (*RoleResponse, *v2.RateLimitDescription, error)
	AssignRoleToUserFunc   func(ctx context.Context, roleId string, userId string) (*RoleResponse, *v2.RateLimitDescription, error)
	RemoveRoleFromUserFunc func(ctx context.Context, roleId string, userId string) (*v2.RateLimitDescription, error)
}

func (m *MockClientService) GetUserByID(ctx context.Context, userId string) (*UserResponse, *v2.RateLimitDescription, error) {
	return m.GetUserByIDFunc(ctx, userId)
}

func (m *MockClientService) CreateUser(ctx context.Context, userRequest UserRequest) (*UserResponse, *v2.RateLimitDescription, error) {
	return m.CreateUserFunc(ctx, userRequest)
}

func (m *MockClientService) DeleteUser(ctx context.Context, userId string) (*v2.RateLimitDescription, error) {
	return m.DeleteUserFunc(ctx, userId)
}

func (m *MockClientService) GetUsers(ctx context.Context, pageToken *string) ([]UserResponse, *string, *v2.RateLimitDescription, error) {
	return m.GetUsersFunc(ctx, pageToken)
}

func (m *MockClientService) GetServiceAccounts(ctx context.Context) ([]ServiceAccountResponse, *v2.RateLimitDescription, error) {
	return m.GetServiceAccountsFunc(ctx)
}

func (m *MockClientService) GetRoles(ctx context.Context, pageToken *string) ([]RoleResponse, *string, *v2.RateLimitDescription, error) {
	return m.GetRolesFunc(ctx, pageToken)
}

func (m *MockClientService) GetRole(ctx context.Context, roleId string) (*RoleResponse, *v2.RateLimitDescription, error) {
	return m.GetRoleFunc(ctx, roleId)
}

func (m *MockClientService) AssignRoleToUser(ctx context.Context, roleId string, userId string) (*RoleResponse, *v2.RateLimitDescription, error) {
	return m.AssignRoleToUserFunc(ctx, roleId, userId)
}

func (m *MockClientService) RemoveRoleFromUser(ctx context.Context, roleId string, userId string) (*v2.RateLimitDescription, error) {
	return m.RemoveRoleFromUserFunc(ctx, roleId, userId)
}
