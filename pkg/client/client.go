package client

import (
	"context"
	"fmt"
	"net/url"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	apiVersion       = "v1"
	resourcePageSize = 1000 // API: Default value is 100 and the range is 1-1000.
)

type Client struct {
	httpClient *uhttp.BaseHttpClient
	apiBaseURL *url.URL
}

func NewClient(ctx context.Context, apiBaseURL, apiAccessID, apiAccessKey string) (*Client, error) {
	// Create API base URL
	url, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing API base URL: %w", err)
	}

	// Create a basic auth client with proper options
	httpClient, err := uhttp.NewBasicAuth(apiAccessID, apiAccessKey).GetClient(ctx,
		uhttp.WithUserAgent("baton-sumo-logic"),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating http client: %w", err)
	}

	// Create the base HTTP client with the authenticated client
	baseClient, err := uhttp.NewBaseHttpClientWithContext(ctx, httpClient)
	if err != nil {
		return nil, fmt.Errorf("error creating base http client: %w", err)
	}

	return &Client{
		httpClient: baseClient,
		apiBaseURL: url,
	}, nil
}

// GetUsers retrieves user accounts (and service accounts when specified) from the API.
func (c *Client) getUsers(ctx context.Context, pageToken *string, includeServiceAccounts bool) (
	[]*Account,
	*string,
	*v2.RateLimitDescription,
	error,
) {
	l := ctxzap.Extract(ctx)
	// API Doc: https://api.sumologic.com/docs/#operation/listUsers
	path := "/api/{{.apiVersion}}/users"
	pathParameters := map[string]string{"apiVersion": apiVersion}
	queryParams := map[string]string{"includeServiceAccounts": fmt.Sprintf("%t", includeServiceAccounts)}

	var response ApiResponse[Account]

	pageSize := uint(resourcePageSize)
	url, err := c.constructURL(path, pathParameters, queryParams, pageToken, &pageSize)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error generating user list URL: %w", err)
	}

	rateLimit, err := c.get(ctx, url, &response)
	if err != nil {
		return nil, nil, rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	l.Debug("get-users: results",
		zap.String("request_url", url.String()),
		zap.Bool("include_service_accounts", includeServiceAccounts),
		zap.Int("users_count", len(response.Data)),
		zap.Stringp("next_page_token", response.Next),
	)

	return response.Data, response.Next, rateLimit, nil
}

// GetRoles retrieves roles from the API.
func (c *Client) getRoles(ctx context.Context, pageToken *string) (
	[]*RoleResponse,
	*string,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/listRoles
	path := "/api/{{.apiVersion}}/roles"
	pathParameters := map[string]string{"apiVersion": apiVersion}

	var response ApiResponse[RoleResponse]

	pageSize := uint(resourcePageSize)
	url, err := c.constructURL(path, pathParameters, nil, pageToken, &pageSize)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error generating role list URL: %w", err)
	}

	rateLimit, err := c.get(ctx, url, &response)
	if err != nil {
		return nil, nil, rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	return response.Data, response.Next, rateLimit, nil
}

// GetRole retrieves a role by ID.
func (c *Client) getRole(ctx context.Context, roleId string) (
	*RoleResponse,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/listRoles
	path := "/api/{{.apiVersion}}/roles/{{.roleID}}"
	pathParameters := map[string]string{"apiVersion": apiVersion, "roleID": roleId}

	var response RoleResponse

	url, err := c.constructURL(path, pathParameters, nil, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating role list URL: %w", err)
	}

	rateLimit, err := c.get(ctx, url, &response)
	if err != nil {
		return nil, rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	return &response, rateLimit, nil
}

func (c *Client) assignRoleToUser(ctx context.Context, roleId string, userId string) (
	*RoleResponse,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/assignRoleToUser
	path := "/api/{{.apiVersion}}/roles/{{.roleID}}/users/{{.userID}}"
	pathParameters := map[string]string{"apiVersion": apiVersion, "roleID": roleId, "userID": userId}

	var response RoleResponse

	url, err := c.constructURL(path, pathParameters, nil, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating assign role to user URL: %w", err)
	}

	rateLimit, err := c.put(ctx, url, &response)
	if err != nil {
		return nil, rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	return &response, rateLimit, nil
}

func (c *Client) removeRoleFromUser(ctx context.Context, roleId string, userId string) (
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/removeRoleFromUser
	path := "/api/{{.apiVersion}}/roles/{{.roleID}}/users/{{.userID}}"
	pathParameters := map[string]string{"apiVersion": apiVersion, "roleID": roleId, "userID": userId}

	url, err := c.constructURL(path, pathParameters, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error generating remove role from user URL: %w", err)
	}

	rateLimit, err := c.delete(ctx, url, nil)
	if err != nil {
		return rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	return rateLimit, nil
}

func (c *Client) getUserByID(ctx context.Context, userId string) (
	*Account,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/getUser
	path := "/api/{{.apiVersion}}/users/{{.userID}}"
	pathParameters := map[string]string{"apiVersion": apiVersion, "userID": userId}

	url, err := c.constructURL(path, pathParameters, nil, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating get user by ID URL: %w", err)
	}

	var response Account
	rateLimit, err := c.get(ctx, url, &response)
	if err != nil {
		return nil, rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	return &response, rateLimit, nil
}

func (c *Client) createUser(ctx context.Context, userRequest UserRequest) (
	*Account,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/createUser
	path := "/api/{{.apiVersion}}/users"
	pathParameters := map[string]string{"apiVersion": apiVersion}

	url, err := c.constructURL(path, pathParameters, nil, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating create user URL: %w", err)
	}

	payload := map[string]interface{}{
		"firstName": userRequest.FirstName,
		"lastName":  userRequest.LastName,
		"email":     userRequest.Email,
		"roleIds":   userRequest.RoleIDs,
	}

	var response Account
	rateLimit, err := c.post(ctx, url, &response, payload)
	if err != nil {
		return nil, rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	return &response, rateLimit, nil
}

func (c *Client) deleteUser(ctx context.Context, userId string) (
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/deleteUser
	path := "/api/{{.apiVersion}}/users/{{.userID}}"
	pathParameters := map[string]string{"apiVersion": apiVersion, "userID": userId}

	url, err := c.constructURL(path, pathParameters, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error generating delete user URL: %w", err)
	}

	rateLimit, err := c.delete(ctx, url, nil)
	if err != nil {
		return rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	return rateLimit, nil
}

// SearchRoleByName retrieves a role by its display name using the `name` query parameter. Returns the first match.
func (c *Client) SearchRoleByName(ctx context.Context, roleName string) (*RoleResponse, *v2.RateLimitDescription, error) {
	if roleName == "" {
		return nil, nil, fmt.Errorf("role name cannot be empty")
	}

	// API Doc: https://api.sumologic.com/docs/#operation/listRoles
	path := "/api/{{.apiVersion}}/roles"
	pathParameters := map[string]string{"apiVersion": apiVersion}
	queryParams := map[string]string{"name": roleName}

	var response ApiResponse[RoleResponse]

	pageSize := uint(resourcePageSize)
	url, err := c.constructURL(path, pathParameters, queryParams, nil, &pageSize)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating role search URL: %w", err)
	}

	rateLimit, err := c.get(ctx, url, &response)
	if err != nil {
		return nil, rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, rateLimit, fmt.Errorf("no role found with name '%s'", roleName)
	}

	return response.Data[0], rateLimit, nil
}
