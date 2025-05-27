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
		roleCopy := role
		roleResource, err := createRoleResource(&roleCopy)
		if err != nil {
			return nil, "", outputAnnotations, fmt.Errorf("failed to create role resource: %w", err)
		}
		resources = append(resources, roleResource)
	}

	return resources, createPageToken(nextPageToken), outputAnnotations, nil
}

// Entitlements always returns an empty slice for roles.
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
