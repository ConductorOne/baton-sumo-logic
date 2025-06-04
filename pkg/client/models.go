package client

import (
	"fmt"
	"time"
)

type ErrorResponse struct {
	Code   string  `json:"code"`
	Msg    string  `json:"message"`
	Target *string `json:"target,omitempty"`
}

// Implement the required method for the interface.
func (e *ErrorResponse) Message() string {
	target := "none"
	if e.Target != nil {
		target = *e.Target
	}
	return fmt.Sprintf("code: %s, message: %s, target: %s", e.Code, e.Msg, target)
}

type ApiResponse[T any] struct {
	// Data is the list of items returned by the API.
	Data []*T `json:"data"`
	// Next is the token to get the next page of results.
	Next *string `json:"next,omitempty"`
}

type BaseAccount struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	// Creation timestamp in UTC in RFC3339 format <date-time> (YYYY-MM-DDTHH:MM:SSZ).
	// https://datatracker.ietf.org/doc/html/rfc3339 .
	CreatedAt time.Time `json:"createdAt"`
	// Identifier of the user who created the resource.
	CreatedBy string `json:"createdBy"`
	// Last modification timestamp in UTC in RFC3339 format <date-time> (YYYY-MM-DDTHH:MM:SSZ).
	ModifiedAt time.Time `json:"modifiedAt"`
	// Identifier of the user who last modified the resource.
	ModifiedBy string   `json:"modifiedBy"`
	RoleIDs    []string `json:"roleIds"`
	IsActive   *bool    `json:"isActive,omitempty"`
}

type UserResponse struct {
	BaseAccount
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	// This has the value true if the user's account has been locked.
	// If a user tries to log into their account several times and fails, his or her account will be locked for security reasons.
	IsLocked *bool `json:"isLocked,omitempty"`
	// True if multi factor authentication is enabled for the user.
	IsMfaEnabled *bool `json:"isMfaEnabled,omitempty"`
	// Last login timestamp in UTC in RFC3339 format <date-time> (YYYY-MM-DDTHH:MM:SSZ).
	LastLoginTimestamp *time.Time `json:"lastLoginTimestamp,omitempty"`
}

type ServiceAccountResponse struct {
	BaseAccount
	Name string `json:"name"`
}

type RoleResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	FilterPredicate *string `json:"filterPredicate,omitempty"`
	// List of user identifiers to assign the role to.
	Users *[]string `json:"users,omitempty"`
	// List of capabilities to assign the role to.
	Capabilities *[]string `json:"capabilities,omitempty"`
	// Set this to true if you want to automatically append all missing capability requirements.
	// If set to false an error will be thrown if any capabilities are missing their dependencies.
	AutofillDependencies *bool `json:"autofillDependencies,omitempty"`
	// Creation timestamp in UTC in RFC3339 format <date-time> (YYYY-MM-DDTHH:MM:SSZ).
	CreatedAt string `json:"createdAt"`
	// Identifier of the user who created the resource.
	CreatedBy string `json:"createdBy"`
	// Last modification timestamp in UTC in RFC3339 format <date-time> (YYYY-MM-DDTHH:MM:SSZ).
	ModifiedAt string `json:"modifiedAt"`
	// Identifier of the user who last modified the resource.
	ModifiedBy string `json:"modifiedBy"`
	// This has the value true if the role is defined by the system.
	// If a role is defined by the system, it cannot be deleted or modified.
	SystemDefined *bool `json:"systemDefined,omitempty"`
}
