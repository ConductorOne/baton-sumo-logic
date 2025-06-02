package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sumo-logic/pkg/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type userBuilder struct {
	service                client.ClientService
	includeServiceAccounts bool
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

func (o *userBuilder) CreateAccountCapabilityDetails(_ context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
	}, nil, nil
}

func (o *userBuilder) CreateAccount(
	ctx context.Context,
	accountInfo *v2.AccountInfo,
	credentialOptions *v2.CredentialOptions,
) (
	connectorbuilder.CreateAccountResponse,
	[]*v2.PlaintextData,
	annotations.Annotations,
	error,
) {
	userRequest, err := accountInfoToUserRequest(accountInfo)
	if err != nil {
		return nil, nil, nil, err
	}

	outputAnnotations := annotations.New()
	user, rateLimit, err := o.service.CreateUser(ctx, *userRequest)
	outputAnnotations.WithRateLimiting(rateLimit)
	if err != nil {
		return nil, nil, outputAnnotations, fmt.Errorf("failed to create user: %w", err)
	}

	userResource, err := createUserResource(user)
	if err != nil {
		return nil, nil, nil, err
	}

	car := &v2.CreateAccountResponse_SuccessResult{
		Resource: userResource,
	}

	return car, nil, nil, nil
}

// Delete implements the ResourceDeleter interface.
func (o *userBuilder) Delete(ctx context.Context, resourceId *v2.ResourceId) (annotations.Annotations, error) {
	accountID := resourceId.GetResource()
	if len(accountID) == 0 {
		return nil, fmt.Errorf("missing resource ID")
	}
	l := ctxzap.Extract(ctx).With(zap.String("accountID", accountID))

	// check the account exists
	outputAnnotations := annotations.New()
	account, rateLimit, err := o.service.GetUserByID(ctx, accountID)
	outputAnnotations.WithRateLimiting(rateLimit)
	if err != nil {
		l.Error("baton-sumo-logic: delete-user: failed to get account by user ID", zap.Error(err))
		return outputAnnotations, err
	}

	// delete the account
	rateLimit, err = o.service.DeleteUser(ctx, account.ID)
	outputAnnotations.WithRateLimiting(rateLimit)
	if err != nil {
		l.Error("baton-sumo-logic: delete-user: failed to delete account with user ID", zap.Error(err))
		return outputAnnotations, err
	}

	// verify the account no longer exists
	_, rateLimit, err = o.service.GetUserByID(ctx, account.ID)
	outputAnnotations.WithRateLimiting(rateLimit)
	if err == nil || status.Code(err) != codes.NotFound {
		l.Error("baton-sumo-logic: delete-user: failed: Account with ID should have been deleted", zap.Error(err))
		return outputAnnotations, err
	}
	// log the deleted account success
	l.Info("baton-sumo-logic: delete-user: success")
	return nil, nil
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
			serviceAccountCopy := serviceAccount
			userResource, err := createUserResource(&serviceAccountCopy)
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
		humanAccountCopy := humanAccount
		userResource, err := createUserResource(&humanAccountCopy)
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
		"created_at":  base.CreatedAt,
		"created_by":  base.CreatedBy,
		"modified_by": base.ModifiedBy,
		"modified_at": base.ModifiedAt,
	}

	userTraitOptions := []rs.UserTraitOption{}

	// Handle active status
	if base.IsActive == nil || !*base.IsActive {
		userTraitOptions = append(userTraitOptions, rs.WithStatus(v2.UserTrait_Status_STATUS_DISABLED))
	} else {
		userTraitOptions = append(userTraitOptions, rs.WithStatus(v2.UserTrait_Status_STATUS_ENABLED))
	}

	// Handle specific account types
	switch a := account.(type) {
	case *client.UserResponse:
		fullName = a.FirstName + " " + a.LastName
		profile["full_name"] = fullName

		// Add optional human user fields
		if a.IsLocked != nil && *a.IsLocked {
			profile["is_locked"] = *a.IsLocked
		}
		if a.IsMfaEnabled != nil {
			profile["mfa_enabled"] = *a.IsMfaEnabled
		}
		if a.LastLoginTimestamp != nil {
			profile["last_login"] = *a.LastLoginTimestamp
		}
		userTraitOptions = append(userTraitOptions, rs.WithAccountType(v2.UserTrait_ACCOUNT_TYPE_HUMAN))

	case *client.ServiceAccountResponse:
		fullName = a.Name
		profile["full_name"] = fullName
		userTraitOptions = append(userTraitOptions, rs.WithAccountType(v2.UserTrait_ACCOUNT_TYPE_SERVICE))

	default:
		return nil, fmt.Errorf("unsupported account type: %T", account)
	}

	// Create the resource
	return rs.NewUserResource(
		fullName,
		userResourceType,
		base.ID,
		userTraitOptions,
	)
}

func accountInfoToUserRequest(accountInfo *v2.AccountInfo) (*client.UserRequest, error) {
	pMap := accountInfo.Profile.AsMap()

	firstName, ok := pMap["first_name"]
	if !ok {
		return nil, fmt.Errorf("missing first name in account info")
	}

	lastName, ok := pMap["last_name"]
	if !ok {
		return nil, fmt.Errorf("missing last name in account info")
	}

	email, ok := pMap["email"]
	if !ok {
		return nil, fmt.Errorf("missing email in account info")
	}

	roleID, ok := pMap["default_role_id"]
	if !ok {
		return nil, fmt.Errorf("missing default role ID in account info")
	}

	return &client.UserRequest{
		FirstName: firstName.(string),
		LastName:  lastName.(string),
		Email:     email.(string),
		RoleIDs:   []string{roleID.(string)},
	}, nil
}
