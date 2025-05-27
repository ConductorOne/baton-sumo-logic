package connector

import (
	"context"
	"fmt"
	"testing"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sumo-logic/pkg/client"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper function to create a test builder with mocks.
func newTestRoleBuilder() (*roleBuilder, *client.MockClientService) {
	mockClient := &client.Client{}
	mockClientService := &client.MockClientService{}

	builder := newRoleBuilder(mockClient)
	// Replace the service with our mock.
	builder.service = mockClientService

	return builder, mockClientService
}

func TestRolesList(t *testing.T) {
	ctx := context.Background()

	t.Run("should get ratelimit annotations", func(t *testing.T) {
		// Create a new role builder with a mock client service.
		roleBuilder, mockClientService := newTestRoleBuilder()

		mockClientService.GetRolesFunc = func(
			ctx context.Context,
			pageToken *string,
		) (
			[]client.RoleResponse,
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

		resources, token, annotations, err := roleBuilder.List(ctx, nil, &pagination.Token{})

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
		// Create a new role builder with a mock client service.
		roleBuilder, mockClientService := newTestRoleBuilder()

		startToken := "start-token"
		mockClientService.GetRolesFunc = func(
			ctx context.Context,
			pageToken *string,
		) (
			[]client.RoleResponse,
			*string,
			*v2.RateLimitDescription,
			error,
		) {
			require.Equal(t, startToken, *pageToken)
			return nil, nil, nil, nil
		}

		_, _, _, _ = roleBuilder.List(ctx, nil, &pagination.Token{Token: startToken})
	})

	t.Run("should get roles", func(t *testing.T) {
		// Create a new role builder with a mock client service.
		roleBuilder, mockClientService := newTestRoleBuilder()

		mockClientService.GetRolesFunc = func(
			ctx context.Context,
			pageToken *string,
		) (
			[]client.RoleResponse,
			*string,
			*v2.RateLimitDescription,
			error,
		) {
			description := "Test Role"
			roles := []client.RoleResponse{
				{
					ID:          "1",
					Name:        "baton-role",
					Description: &description,
				},
			}
			return roles, nil, nil, nil
		}

		resources, token, annotations, err := roleBuilder.List(ctx, nil, &pagination.Token{})

		// Assert the returned role has an ID.
		require.NotNil(t, resources)
		require.Len(t, resources, 1)
		require.NotEmpty(t, resources[0].Id)

		require.NotNil(t, token)
		AssertNoRatelimitAnnotations(t, annotations)
		require.Nil(t, err)
	})
}
