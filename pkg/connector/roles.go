package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sumo-logic/pkg/client"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const roleAssignmentEntitlement = "assigned"

type roleBuilder struct {
	service client.ClientService
}

func (o *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return roleResourceType
}

// List returns all the roles from the database as resource objects.
func (o *roleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	outputAnnotations := annotations.New()

	roles, nextPageToken, rateLimit, err := o.service.GetRoles(ctx, parsePageToken(pToken))
	outputAnnotations.WithRateLimiting(rateLimit)
	if err != nil {
		return nil, "", outputAnnotations, fmt.Errorf("failed to list roles: %w", err)
	}

	resources := make([]*v2.Resource, 0, len(roles))
	for _, role := range roles {
		roleResource, err := createRoleResource(role)
		if err != nil {
			return nil, "", outputAnnotations, fmt.Errorf("failed to create role resource: %w", err)
		}
		resources = append(resources, roleResource)
	}

	return resources, createPageToken(nextPageToken), outputAnnotations, nil
}

func (o *roleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDisplayName(fmt.Sprintf("%s Role Member", resource.DisplayName)),
		ent.WithDescription(fmt.Sprintf("Has the %s role in Sumo Logic", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(resource, roleAssignmentEntitlement, assignmentOptions...))

	return rv, "", nil, nil
}

func (o *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	outputAnnotations := annotations.New()
	role, rateLimit, err := o.service.GetRole(ctx, resource.Id.Resource)
	outputAnnotations.WithRateLimiting(rateLimit)
	if err != nil {
		return nil, "", outputAnnotations, fmt.Errorf("failed to get role: %w", err)
	}

	if role.Users == nil {
		return nil, "", outputAnnotations, nil
	}

	rv := make([]*v2.Grant, 0, len(*role.Users))
	for _, userId := range *role.Users {
		userResource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     userId,
			},
		}

		rv = append(rv, grant.NewGrant(resource, roleAssignmentEntitlement, userResource))
	}

	return rv, "", outputAnnotations, nil
}

func (o *roleBuilder) Grant(
	ctx context.Context,
	principal *v2.Resource,
	entitlement *v2.Entitlement,
) (annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)

	if principal.Id.ResourceType != userResourceType.Id {
		logger.Error(
			"baton-sumo-logic: only users can be assigned to a role",
			zap.String("principal_type", principal.Id.ResourceType),
			zap.String("principal_id", principal.Id.Resource),
		)
		return nil, fmt.Errorf("baton-sumo-logic: only users can be assigned to a role")
	}

	outputAnnotations := annotations.New()
	_, rateLimitData, err := o.service.AssignRoleToUser(
		ctx,
		entitlement.Resource.Id.Resource,
		principal.Id.Resource,
	)
	outputAnnotations.WithRateLimiting(rateLimitData)

	if err != nil {
		// We are not checking if the grant is already exists because the API DOC does not provide specific information.
		// API Doc: https://api.sumologic.com/docs/#operation/assignRoleToUser
		return outputAnnotations, fmt.Errorf("baton-sumo-logic: failed to assign role to user: %w", err)
	}

	return outputAnnotations, nil
}

func (o *roleBuilder) Revoke(
	ctx context.Context,
	grant *v2.Grant,
) (
	annotations.Annotations,
	error,
) {
	logger := ctxzap.Extract(ctx)

	if grant.Principal.Id.ResourceType != userResourceType.Id {
		logger.Error(
			"baton-sumo-logic: only users can be assigned to a role",
			zap.String("principal_type", grant.Principal.Id.ResourceType),
			zap.String("principal_id", grant.Principal.Id.Resource),
		)
		return nil, fmt.Errorf("baton-sumo-logic: only users can be revoked from a role")
	}

	outputAnnotations := annotations.New()

	rateLimitData, err := o.service.RemoveRoleFromUser(
		ctx,
		grant.Entitlement.Resource.Id.Resource,
		grant.Principal.Id.Resource,
	)
	outputAnnotations.WithRateLimiting(rateLimitData)

	if err != nil {
		// We are not checking if the grant was already revoked because the API DOC does not provide specific information.
		// API Doc: https://api.sumologic.com/docs/#operation/assignRoleToUser
		return outputAnnotations, fmt.Errorf("baton-sumo-logic: failed to revoke role from user: %w", err)
	}

	return outputAnnotations, nil
}

func newRoleBuilder(cclient *client.Client) *roleBuilder {
	return &roleBuilder{
		service: client.NewClientService(cclient),
	}
}

func createRoleResource(role *client.RoleResponse) (*v2.Resource, error) {
	var description string
	if role.Description != nil {
		description = *role.Description
	}

	profile := map[string]interface{}{
		"role_id":     role.ID,
		"role_name":   role.Name,
		"description": description,
		"modified_by": role.ModifiedBy,
		"modified_at": role.ModifiedAt,
		"created_by":  role.CreatedBy,
		"created_at":  role.CreatedAt,
	}

	roleTraitOptions := []rs.RoleTraitOption{
		rs.WithRoleProfile(profile),
	}

	resource, err := rs.NewRoleResource(
		role.Name,
		roleResourceType,
		role.ID,
		roleTraitOptions,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}
