package connector

import (
	"github.com/conductorone/baton-sdk/pkg/pagination"
)

func parsePageToken(pToken *pagination.Token) *string {
	if pToken == nil {
		return nil
	}
	return &pToken.Token
}

func createPageToken(pageToken *string) string {
	if pageToken == nil {
		return ""
	}
	return *pageToken
}
