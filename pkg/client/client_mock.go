package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type MockClientService struct {
	GetUsersFunc           func(ctx context.Context, pageToken *string) ([]UserResponse, *string, *v2.RateLimitDescription, error)
	GetServiceAccountsFunc func(ctx context.Context) ([]ServiceAccountResponse, *v2.RateLimitDescription, error)
	GetRolesFunc           func(ctx context.Context, pageToken *string) ([]RoleResponse, *string, *v2.RateLimitDescription, error)
	GetRoleFunc            func(ctx context.Context, roleId string) (*RoleResponse, *v2.RateLimitDescription, error)
	AssignRoleToUserFunc   func(ctx context.Context, roleId string, userId string) (*RoleResponse, *v2.RateLimitDescription, error)
	RemoveRoleFromUserFunc func(ctx context.Context, roleId string, userId string) (*RoleResponse, *v2.RateLimitDescription, error)
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
