package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

const (
	apiVersion       = "v1"
	resourcePageSize = 100 // API: Default value is 100 and the range is 1-100.
)

type Client struct {
	httpClient        *uhttp.BaseHttpClient
	apiBaseURL        *url.URL
	base64Credentials string
}

func NewClient(ctx context.Context, apiBaseURL, apiAccessID, apiAccessKey string) (*Client, error) {
	httpClient, err := uhttp.NewBaseHttpClientWithContext(ctx, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("error creating http client: %w", err)
	}

	// Create API base URL
	url, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing API base URL: %w", err)
	}

	return &Client{
		httpClient:        httpClient,
		apiBaseURL:        url,
		base64Credentials: encodeBase64(apiAccessID, apiAccessKey),
	}, nil
}

// GetUsers retrieves users from the API.
func (c *Client) getUsers(ctx context.Context, pageToken *string) (
	[]UserResponse,
	*string,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/listUsers
	path := "/api/{api-version}/users"
	pathParameters := map[string]string{"api-version": apiVersion}

	var response ApiResponse[UserResponse]

	pageSize := uint(resourcePageSize)
	url, err := c.constructURL(path, pathParameters, nil, pageToken, &pageSize)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error generating user list URL: %w", err)
	}

	rateLimit, err := c.get(ctx, url, &response)
	if err != nil {
		return nil, nil, rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	return response.Data, response.Next, rateLimit, nil
}

// GetServiceAccounts retrieves service accounts from the API.
func (c *Client) getServiceAccounts(ctx context.Context) (
	[]ServiceAccountResponse,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/listServiceAccounts
	path := "/api/{api-version}/serviceAccounts"
	pathParameters := map[string]string{"api-version": apiVersion}

	var response ApiResponse[ServiceAccountResponse]

	url, err := c.constructURL(path, pathParameters, nil, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating service account list URL: %w", err)
	}

	rateLimit, err := c.get(ctx, url, &response)
	if err != nil {
		return nil, rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	return response.Data, rateLimit, nil
}

// GetRoles retrieves roles from the API.
func (c *Client) getRoles(ctx context.Context, pageToken *string) (
	[]RoleResponse,
	*string,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/listRoles
	path := "/api/{api-version}/roles"
	pathParameters := map[string]string{"api-version": apiVersion}

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
	path := "/api/{api-version}/roles/{role-id}"
	pathParameters := map[string]string{"api-version": apiVersion, "role-id": roleId}

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

// AssignRoleToUser assigns a role to a user.
func (c *Client) assignRoleToUser(ctx context.Context, roleId string, userId string) (
	*RoleResponse,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/assignRoleToUser
	path := "/api/{api-version}/roles/{role-id}/users/{user-id}"
	pathParameters := map[string]string{"api-version": apiVersion, "role-id": roleId, "user-id": userId}

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

// RemoveRoleFromUser removes a role from a user.
func (c *Client) removeRoleFromUser(ctx context.Context, roleId string, userId string) (
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/removeRoleFromUser
	path := "/api/{api-version}/roles/{role-id}/users/{user-id}"
	pathParameters := map[string]string{"api-version": apiVersion, "role-id": roleId, "user-id": userId}

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
	*UserResponse,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/getUser
	path := "/api/{api-version}/users/{user-id}"
	pathParameters := map[string]string{"api-version": apiVersion, "user-id": userId}

	url, err := c.constructURL(path, pathParameters, nil, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating get user by ID URL: %w", err)
	}

	var response UserResponse
	rateLimit, err := c.get(ctx, url, &response)
	if err != nil {
		return nil, rateLimit, fmt.Errorf("error executing request: %w", err)
	}

	return &response, rateLimit, nil
}

func (c *Client) createUser(ctx context.Context, userRequest UserRequest) (
	*UserResponse,
	*v2.RateLimitDescription,
	error,
) {
	// API Doc: https://api.sumologic.com/docs/#operation/createUser
	path := "/api/{api-version}/users"
	pathParameters := map[string]string{"api-version": apiVersion}

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

	var response UserResponse
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
	path := "/api/{api-version}/users/{user-id}"
	pathParameters := map[string]string{"api-version": apiVersion, "user-id": userId}

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
