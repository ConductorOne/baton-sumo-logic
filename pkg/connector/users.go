package connector

import (
	"context"
	"fmt"
	"time"

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

	// Fetch and process human (and service) accounts
	accounts, nextPageToken, rateLimit, err := o.service.GetUsers(ctx, parsePageToken(pToken), o.includeServiceAccounts)
	outputAnnotations.WithRateLimiting(rateLimit)
	if err != nil {
		return nil, "", outputAnnotations, fmt.Errorf("failed to get human accounts: %w", err)
	}
	fmt.Println("ALL USER AND SERVICE ACCOUNTS", accounts)

	// Process accounts
	for _, account := range accounts {
		resource, err := createUserResource(account)
		if err != nil {
			return nil, "", outputAnnotations, fmt.Errorf("failed to create user resource from human account: %w", err)
		}
		resources = append(resources, resource)
	}

	fmt.Println("ALL USER AND SERVICE RESOURCES", resources)
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
func createUserResource(account *client.AccountResponse) (*v2.Resource, error) {
	if account == nil {
		return nil, fmt.Errorf("account cannot be nil")
	}

	var fullName string
	profile := map[string]interface{}{
		"id":          account.ID,
		"email":       account.Email,
		"created_at":  account.CreatedAt.Format(time.RFC3339),
		"created_by":  account.CreatedBy,
		"modified_at": account.ModifiedAt.Format(time.RFC3339),
		"modified_by": account.ModifiedBy,
	}

	// Initialize base user trait options with common fields (email, login, and creation time).
	userTraitOptions := []rs.UserTraitOption{
		rs.WithUserLogin(account.Email),
		rs.WithEmail(account.Email, true),
		rs.WithCreatedAt(account.CreatedAt),
	}

	// default baton-sdk is enabled, so we only need to set disabled if the account is disabled.
	if account.IsActive == nil || !*account.IsActive {
		userTraitOptions = append(userTraitOptions, rs.WithStatus(v2.UserTrait_Status_STATUS_DISABLED))
	}

	// Handle specific account types
	// User accounts have a firstName and lastName
	if account.LastName != "" {
		fullName = account.FirstName + " " + account.LastName
		profile["full_name"] = fullName

		// This has the value true if the user's account has been locked.
		// If a user tries to log into their account several times and fails, his or her account will be locked for security reasons.
		if account.IsLocked != nil {
			profile["is_locked"] = *account.IsLocked
		}

		// True if multi factor authentication is enabled for the user.
		if account.IsMfaEnabled != nil {
			userTraitOptions = append(userTraitOptions, rs.WithMFAStatus(&v2.UserTrait_MFAStatus{
				MfaEnabled: *account.IsMfaEnabled,
			}))
		}

		// Last login timestamp in UTC in RFC3339 format <date-time> (YYYY-MM-DDTHH:MM:SSZ).
		if account.LastLoginTimestamp != nil {
			userTraitOptions = append(userTraitOptions, rs.WithLastLogin(*account.LastLoginTimestamp))
		}

		userTraitOptions = append(userTraitOptions, rs.WithAccountType(v2.UserTrait_ACCOUNT_TYPE_HUMAN))

	} else {
		// Service accounts only firstName populated and no lastName
		// (technically they only have a "name" field but it's shoved into "firstName")
		fullName = account.FirstName
		profile["full_name"] = fullName

		userTraitOptions = append(userTraitOptions, rs.WithAccountType(v2.UserTrait_ACCOUNT_TYPE_SERVICE))

	}

	// The profile is assigned last because it needs to be built up with account-specific fields
	// that are only known after determining whether this is a human or service account.
	// This includes fields like full_name, is_locked, account_type, and other type-specific attributes.
	userTraitOptions = append(userTraitOptions, rs.WithUserProfile(profile))

	// Create the resource
	return rs.NewUserResource(
		fullName,
		userResourceType,
		account.ID,
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
