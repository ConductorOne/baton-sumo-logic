package connector

import (
	"context"
	"fmt"
	"time"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sumo-logic/pkg/client"
)

type userBuilder struct {
	service                client.ClientService
	includeServiceAccounts bool
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all accounts (human and service accounts) from Sumo Logic as resource objects.
func (o *userBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	outputAnnotations := annotations.New()
	resources := make([]*v2.Resource, 0)

	// Service accounts endpoint does not support pagination, so we only fetch them on the first page.
	if pToken.Token == "" && o.includeServiceAccounts {
		// Fetch both human and service accounts
		serviceAccounts, rateLimit, err := o.service.GetServiceAccounts(ctx)
		outputAnnotations.WithRateLimiting(rateLimit)
		if err != nil {
			return nil, "", outputAnnotations, fmt.Errorf("failed to get service accounts: %w", err)
		}

		// Process service accounts
		for _, serviceAccount := range serviceAccounts {
			userResource, err := createUserResource(serviceAccount)
			if err != nil {
				return nil, "", outputAnnotations, fmt.Errorf("failed to create user resource from service account: %w", err)
			}
			resources = append(resources, userResource)
		}
	}

	// Fetch and process human accounts
	humanAccounts, nextPageToken, rateLimit, err := o.service.GetUsers(ctx, parsePageToken(pToken))
	outputAnnotations.WithRateLimiting(rateLimit)
	if err != nil {
		return nil, "", outputAnnotations, fmt.Errorf("failed to get human accounts: %w", err)
	}

	// Process human accounts
	for _, humanAccount := range humanAccounts {
		userResource, err := createUserResource(humanAccount)
		if err != nil {
			return nil, "", outputAnnotations, fmt.Errorf("failed to create user resource from human account: %w", err)
		}
		resources = append(resources, userResource)
	}

	return resources, createPageToken(nextPageToken), outputAnnotations, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(cclient *client.Client, includeServiceAccounts bool) *userBuilder {
	return &userBuilder{
		service:                client.NewClientService(cclient),
		includeServiceAccounts: includeServiceAccounts,
	}
}

// createUserResource creates a resource object for either a UserResponse or ServiceAccountResponse.
func createUserResource(account interface{}) (*v2.Resource, error) {
	var fullName string
	var base client.BaseAccount
	switch a := account.(type) {
	case *client.UserResponse:
		base = a.BaseAccount
	case *client.ServiceAccountResponse:
		base = a.BaseAccount
	default:
		return nil, fmt.Errorf("unsupported account type: %T", account)
	}

	profile := map[string]interface{}{
		"id":          base.ID,
		"email":       base.Email,
		"created_at":  base.CreatedAt.Format(time.RFC3339),
		"created_by":  base.CreatedBy,
		"modified_at": base.ModifiedAt.Format(time.RFC3339),
		"modified_by": base.ModifiedBy,
	}

	// Initialize base user trait options with common fields (email, login, and creation time).
	userTraitOptions := []rs.UserTraitOption{
		rs.WithUserLogin(base.Email),
		rs.WithEmail(base.Email, true),
		rs.WithCreatedAt(base.CreatedAt),
	}

	// default baton-sdk is enabled, so we only need to set disabled if the account is disabled.
	if base.IsActive == nil || !*base.IsActive {
		userTraitOptions = append(userTraitOptions, rs.WithStatus(v2.UserTrait_Status_STATUS_DISABLED))
	}

	// Handle specific account types
	switch a := account.(type) {
	case *client.UserResponse:
		fullName = a.FirstName + " " + a.LastName
		profile["full_name"] = fullName

		// This has the value true if the user's account has been locked.
		// If a user tries to log into their account several times and fails, his or her account will be locked for security reasons.
		if a.IsLocked != nil {
			profile["is_locked"] = *a.IsLocked
		}

		// True if multi factor authentication is enabled for the user.
		if a.IsMfaEnabled != nil {
			userTraitOptions = append(userTraitOptions, rs.WithMFAStatus(&v2.UserTrait_MFAStatus{
				MfaEnabled: *a.IsMfaEnabled,
			}))
		}

		// Last login timestamp in UTC in RFC3339 format <date-time> (YYYY-MM-DDTHH:MM:SSZ).
		if a.LastLoginTimestamp != nil {
			userTraitOptions = append(userTraitOptions, rs.WithLastLogin(*a.LastLoginTimestamp))
		}

		userTraitOptions = append(userTraitOptions, rs.WithAccountType(v2.UserTrait_ACCOUNT_TYPE_HUMAN))

	case *client.ServiceAccountResponse:
		fullName = a.Name
		profile["full_name"] = fullName

		userTraitOptions = append(userTraitOptions, rs.WithAccountType(v2.UserTrait_ACCOUNT_TYPE_SERVICE))

	default:
		return nil, fmt.Errorf("unsupported account type: %T", account)
	}

	// The profile is assigned last because it needs to be built up with account-specific fields
	// that are only known after determining whether this is a human or service account.
	// This includes fields like full_name, is_locked, account_type, and other type-specific attributes.
	userTraitOptions = append(userTraitOptions, rs.WithUserProfile(profile))

	// Create the resource
	return rs.NewUserResource(
		fullName,
		userResourceType,
		base.ID,
		userTraitOptions,
	)
}
