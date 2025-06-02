package client

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

// ClientService defines the interface for client operations.
type ClientService interface {
	GetUserByID(ctx context.Context, userId string) (*UserResponse, *v2.RateLimitDescription, error)
	GetUsers(ctx context.Context, pageToken *string) ([]UserResponse, *string, *v2.RateLimitDescription, error)
	CreateUser(ctx context.Context, userRequest UserRequest) (*UserResponse, *v2.RateLimitDescription, error)
	DeleteUser(ctx context.Context, userId string) (*v2.RateLimitDescription, error)
	GetServiceAccounts(ctx context.Context) ([]ServiceAccountResponse, *v2.RateLimitDescription, error)
	GetRoles(ctx context.Context, pageToken *string) ([]RoleResponse, *string, *v2.RateLimitDescription, error)
	GetRole(ctx context.Context, roleId string) (*RoleResponse, *v2.RateLimitDescription, error)
	AssignRoleToUser(ctx context.Context, roleId string, userId string) (*RoleResponse, *v2.RateLimitDescription, error)
	RemoveRoleFromUser(ctx context.Context, roleId string, userId string) (*v2.RateLimitDescription, error)
}

// ClientServiceImpl is the default implementation that calls the actual API.
type ClientServiceImpl struct {
	client Client
}

func NewClientService(client *Client) ClientService {
	return &ClientServiceImpl{client: *client}
}

func (s *ClientServiceImpl) GetUserByID(ctx context.Context, userId string) (*UserResponse, *v2.RateLimitDescription, error) {
	return s.client.getUserByID(ctx, userId)
}

func (s *ClientServiceImpl) CreateUser(ctx context.Context, userRequest UserRequest) (*UserResponse, *v2.RateLimitDescription, error) {
	return s.client.createUser(ctx, userRequest)
}

func (s *ClientServiceImpl) DeleteUser(ctx context.Context, userId string) (*v2.RateLimitDescription, error) {
	return s.client.deleteUser(ctx, userId)
}

func (s *ClientServiceImpl) GetUsers(ctx context.Context, pageToken *string) ([]UserResponse, *string, *v2.RateLimitDescription, error) {
	return s.client.getUsers(ctx, pageToken)
}

func (s *ClientServiceImpl) GetServiceAccounts(ctx context.Context) ([]ServiceAccountResponse, *v2.RateLimitDescription, error) {
	return s.client.getServiceAccounts(ctx)
}

func (s *ClientServiceImpl) GetRoles(ctx context.Context, pageToken *string) ([]RoleResponse, *string, *v2.RateLimitDescription, error) {
	return s.client.getRoles(ctx, pageToken)
}

func (s *ClientServiceImpl) GetRole(ctx context.Context, roleId string) (*RoleResponse, *v2.RateLimitDescription, error) {
	return s.client.getRole(ctx, roleId)
}

func (s *ClientServiceImpl) AssignRoleToUser(ctx context.Context, roleId string, userId string) (*RoleResponse, *v2.RateLimitDescription, error) {
	return s.client.assignRoleToUser(ctx, roleId, userId)
}

func (s *ClientServiceImpl) RemoveRoleFromUser(ctx context.Context, roleId string, userId string) (*v2.RateLimitDescription, error) {
	return s.client.removeRoleFromUser(ctx, roleId, userId)
}
