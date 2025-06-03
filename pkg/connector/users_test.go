package connector

import (
	"context"
	"fmt"
	"testing"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	test "github.com/conductorone/baton-sdk/pkg/test"
	"github.com/conductorone/baton-sumo-logic/pkg/client"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper function to create a test builder with mocks.
func newTestUserBuilder(includeServiceAccounts bool) (*userBuilder, *client.MockClientService) {
	mockClient := &client.Client{}
	mockClientService := &client.MockClientService{}

	builder := newUserBuilder(mockClient, includeServiceAccounts)
	// Replace the service with our mock.
	builder.service = mockClientService

	return builder, mockClientService
}

func TestUsersList(t *testing.T) {
	ctx := context.Background()

	t.Run("should get ratelimit annotations from users (without service accounts)", func(t *testing.T) {
		// Create a new user builder with a mock client service.
		userBuilder, mockClientService := newTestUserBuilder(false)

		mockClientService.GetUsersFunc = func(
			ctx context.Context,
			pageToken *string,
		) (
			[]*client.UserResponse,
			*string,
			*v2.RateLimitDescription,
			error,
		) {
			rateLimitData := v2.RateLimitDescription{
				ResetAt: timestamppb.New(time.Now().Add(10 * time.Second)),
			}
			err := fmt.Errorf("ratelimit error")
			return nil, nil, &rateLimitData, err
		}

		resources, token, annotations, err := userBuilder.List(ctx, nil, &pagination.Token{})

		require.Nil(t, resources)
		require.Empty(t, token)
		require.NotNil(t, err)

		// There should be annotations.
		require.Len(t, annotations, 1)
		rateLimitData := v2.RateLimitDescription{}
		err = annotations[0].UnmarshalTo(&rateLimitData)
		if err != nil {
			t.Errorf("couldn't unmarshal the ratelimit annotation")
		}
		require.NotNil(t, rateLimitData.ResetAt)
	})

	t.Run("should get ratelimit annotations from service accounts", func(t *testing.T) {
		// Create a new user builder with a mock client service.
		userBuilder, mockClientService := newTestUserBuilder(true)

		mockClientService.GetServiceAccountsFunc = func(
			ctx context.Context,
		) (
			[]*client.ServiceAccountResponse,
			*v2.RateLimitDescription,
			error,
		) {
			rateLimitData := v2.RateLimitDescription{
				ResetAt: timestamppb.New(time.Now().Add(10 * time.Second)),
			}
			err := fmt.Errorf("ratelimit error")
			return nil, &rateLimitData, err
		}

		resources, token, annotations, err := userBuilder.List(ctx, nil, &pagination.Token{})

		require.Nil(t, resources)
		require.Empty(t, token)
		require.NotNil(t, err)

		// There should be annotations.
		require.Len(t, annotations, 1)
		rateLimitData := v2.RateLimitDescription{}
		err = annotations[0].UnmarshalTo(&rateLimitData)
		if err != nil {
			t.Errorf("couldn't unmarshal the ratelimit annotation")
		}
		require.NotNil(t, rateLimitData.ResetAt)
	})

	t.Run("should get passed a pagination token", func(t *testing.T) {
		// Create a new user builder with a mock client service.
		userBuilder, mockClientService := newTestUserBuilder(false)

		startToken := "start-token"
		mockClientService.GetUsersFunc = func(
			ctx context.Context,
			pageToken *string,
		) (
			[]*client.UserResponse,
			*string,
			*v2.RateLimitDescription,
			error,
		) {
			require.Equal(t, startToken, *pageToken)
			return nil, nil, nil, nil
		}

		_, _, _, _ = userBuilder.List(ctx, nil, &pagination.Token{Token: startToken})
	})

	t.Run("should get users without service accounts", func(t *testing.T) {
		// Create a new user builder with a mock client service.
		userBuilder, mockClientService := newTestUserBuilder(false)

		mockClientService.GetUsersFunc = func(
			ctx context.Context,
			pageToken *string,
		) (
			[]*client.UserResponse,
			*string,
			*v2.RateLimitDescription,
			error,
		) {
			email := "marcos@conductorone.com"
			isActive := true
			isMfaEnabled := false
			isLocked := false
			lastLoginTimestamp := time.Now()
			createdAt := time.Now()
			modifiedAt := time.Now()
			users := []*client.UserResponse{
				{
					BaseAccount: client.BaseAccount{
						ID:         "1",
						Email:      email,
						IsActive:   &isActive,
						CreatedAt:  createdAt,
						CreatedBy:  "test",
						ModifiedBy: "test",
						ModifiedAt: modifiedAt,
						RoleIDs:    []string{"1", "2"},
					},
					FirstName:          "Marcos",
					LastName:           "Garcia",
					IsMfaEnabled:       &isMfaEnabled,
					IsLocked:           &isLocked,
					LastLoginTimestamp: &lastLoginTimestamp,
				},
			}
			return users, nil, nil, nil
		}

		resources, token, annotations, err := userBuilder.List(ctx, nil, &pagination.Token{})

		// Assert the returned user has an ID.
		require.NotNil(t, resources)
		require.Len(t, resources, 1)
		require.NotEmpty(t, resources[0].Id)

		require.NotNil(t, token)
		test.AssertNoRatelimitAnnotations(t, annotations)
		require.Nil(t, err)
	})

	t.Run("should get users with service accounts", func(t *testing.T) {
		// Create a new user builder with a mock client service.
		userBuilder, mockClientService := newTestUserBuilder(true)

		// Mock the service accounts.
		mockClientService.GetServiceAccountsFunc = func(
			ctx context.Context,
		) (
			[]*client.ServiceAccountResponse,
			*v2.RateLimitDescription,
			error,
		) {
			email := "baton-service-account@conductorone.com"
			isActive := true
			createdAt := time.Now()
			modifiedAt := time.Now()
			serviceAccounts := []*client.ServiceAccountResponse{
				{
					BaseAccount: client.BaseAccount{
						ID:         "1",
						Email:      email,
						IsActive:   &isActive,
						CreatedAt:  createdAt,
						CreatedBy:  "test",
						ModifiedBy: "test",
						ModifiedAt: modifiedAt,
					},
					Name: "baton-service-account",
				},
			}
			return serviceAccounts, nil, nil
		}

		// Mock the users.
		mockClientService.GetUsersFunc = func(
			ctx context.Context,
			pageToken *string,
		) (
			[]*client.UserResponse,
			*string,
			*v2.RateLimitDescription,
			error,
		) {
			email := "baton-user@conductorone.com"
			users := []*client.UserResponse{
				{
					BaseAccount: client.BaseAccount{
						ID:    "2",
						Email: email,
					},
					FirstName: "Baton",
					LastName:  "User",
				},
			}
			return users, nil, nil, nil
		}

		resources, token, annotations, err := userBuilder.List(ctx, nil, &pagination.Token{})

		// Assert the returned user has an ID.
		require.NotNil(t, resources)
		require.Len(t, resources, 2)
		require.NotEmpty(t, resources[0].Id)
		require.NotEmpty(t, resources[1].Id)

		require.NotNil(t, token)
		test.AssertNoRatelimitAnnotations(t, annotations)
		require.Nil(t, err)
	})
}
