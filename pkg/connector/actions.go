package connector

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	config "github.com/conductorone/baton-sdk/pb/c1/config/v1"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/actions"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sumo-logic/pkg/client"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	ActionEnableUser  = "enable_user"
	ActionDisableUser = "disable_user"
)

var enableUserAction = &v2.BatonActionSchema{
	Name: ActionEnableUser,
	Arguments: []*config.Field{
		{
			Name:        "userId",
			DisplayName: "User ID",
			Field:       &config.Field_StringField{},
			IsRequired:  true,
		},
	},
	ReturnTypes: []*config.Field{
		{
			Name:        "success",
			DisplayName: "Success",
			Field:       &config.Field_BoolField{},
		},
	},
	ActionType: []v2.ActionType{
		v2.ActionType_ACTION_TYPE_ACCOUNT,
		v2.ActionType_ACTION_TYPE_ACCOUNT_ENABLE,
	},
}

var disableUserAction = &v2.BatonActionSchema{
	Name: ActionDisableUser,
	Arguments: []*config.Field{
		{
			Name:        "userId",
			DisplayName: "User ID",
			Field:       &config.Field_StringField{},
			IsRequired:  true,
		},
	},
	ReturnTypes: []*config.Field{
		{
			Name:        "success",
			DisplayName: "Success",
			Field:       &config.Field_BoolField{},
		},
	},
	ActionType: []v2.ActionType{
		v2.ActionType_ACTION_TYPE_ACCOUNT,
		v2.ActionType_ACTION_TYPE_ACCOUNT_DISABLE,
	},
}

func (c *Connector) RegisterActionManager(ctx context.Context) (connectorbuilder.CustomActionManager, error) {
	actionManager := actions.NewActionManager(ctx)

	err := actionManager.RegisterAction(ctx, enableUserAction.Name, enableUserAction, c.enableUser)
	if err != nil {
		return nil, err
	}

	err = actionManager.RegisterAction(ctx, disableUserAction.Name, disableUserAction, c.disableUser)
	if err != nil {
		return nil, err
	}

	return actionManager, nil
}

func (c *Connector) enableUser(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if args == nil {
		return nil, nil, fmt.Errorf("arguments cannot be nil")
	}

	if args.Fields == nil {
		return nil, nil, fmt.Errorf("arguments fields cannot be nil")
	}

	userId, ok := args.Fields["userId"]
	if !ok {
		return nil, nil, fmt.Errorf("missing required argument userId")
	}

	if userId == nil {
		return nil, nil, fmt.Errorf("userId value cannot be nil")
	}

	userIdStr := userId.GetStringValue()
	if userIdStr == "" {
		return nil, nil, fmt.Errorf("userId cannot be empty")
	}

	l.Debug("enabling user", zap.String("userId", userIdStr))

	clientService := client.NewClientService(c.client)

	// First, get the current user data to preserve all required fields
	annos := annotations.New()
	currentUser, rateLimit, err := clientService.GetUserByID(ctx, userIdStr)
	annos.WithRateLimiting(rateLimit)
	if err != nil {
		l.Error("failed to get current user data", zap.String("userId", userIdStr), zap.Error(err))
		return nil, annos, fmt.Errorf("failed to get current user data for %s: %w", userIdStr, err)
	}

	// Create update request with all required fields preserved, only changing isActive
	userUpdateRequest := client.UserUpdateRequest{
		FirstName: currentUser.FirstName,
		LastName:  currentUser.LastName,
		IsActive:  &[]bool{true}[0], // Set to true
	}
	updatedUser, rateLimit, err := clientService.UpdateUser(ctx, userIdStr, userUpdateRequest)
	annos.WithRateLimiting(rateLimit)
	if err != nil {
		l.Error("failed to enable user", zap.String("userId", userIdStr), zap.Error(err))
		return nil, annos, fmt.Errorf("failed to enable user %s: %w", userIdStr, err)
	}

	success := updatedUser.IsActive != nil && *updatedUser.IsActive
	if !success {
		l.Warn("user enable operation completed but user is still inactive", zap.String("userId", userIdStr), zap.Boolp("isActive", updatedUser.IsActive))
	}

	response := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"success": structpb.NewBoolValue(success),
		},
	}

	return response, annos, nil
}

func (c *Connector) disableUser(ctx context.Context, args *structpb.Struct) (*structpb.Struct, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if args == nil {
		return nil, nil, fmt.Errorf("arguments cannot be nil")
	}

	if args.Fields == nil {
		return nil, nil, fmt.Errorf("arguments fields cannot be nil")
	}

	userId, ok := args.Fields["userId"]
	if !ok {
		return nil, nil, fmt.Errorf("missing required argument userId")
	}

	if userId == nil {
		return nil, nil, fmt.Errorf("userId value cannot be nil")
	}

	userIdStr := userId.GetStringValue()
	if userIdStr == "" {
		return nil, nil, fmt.Errorf("userId cannot be empty")
	}

	l.Debug("disabling user", zap.String("userId", userIdStr))

	clientService := client.NewClientService(c.client)

	// First, get the current user data to preserve all required fields
	annos := annotations.New()
	currentUser, rateLimit, err := clientService.GetUserByID(ctx, userIdStr)
	annos.WithRateLimiting(rateLimit)
	if err != nil {
		l.Error("failed to get current user data", zap.String("userId", userIdStr), zap.Error(err))
		return nil, annos, fmt.Errorf("failed to get current user data for %s: %w", userIdStr, err)
	}

	l.Debug("current user data", zap.Any("currentUser", currentUser))

	// Create update request with all required fields preserved, only changing isActive
	userUpdateRequest := client.UserUpdateRequest{
		FirstName: currentUser.FirstName,
		LastName:  currentUser.LastName,
		IsActive:  &[]bool{false}[0], // Set to false
	}

	updatedUser, rateLimit, err := clientService.UpdateUser(ctx, userIdStr, userUpdateRequest)
	annos.WithRateLimiting(rateLimit)
	if err != nil {
		l.Error("failed to disable user", zap.String("userId", userIdStr), zap.Error(err))
		return nil, annos, fmt.Errorf("failed to disable user %s: %w", userIdStr, err)
	}

	success := updatedUser.IsActive != nil && !*updatedUser.IsActive
	if !success {
		l.Warn("user disable operation completed but user is still active", zap.String("userId", userIdStr), zap.Boolp("isActive", updatedUser.IsActive))
	}

	response := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"success": structpb.NewBoolValue(success),
		},
	}

	return response, annos, nil
}
