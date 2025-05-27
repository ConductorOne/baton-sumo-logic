package client

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

func encodeBase64(apiAccessID, apiAccessKey string) string {
	return base64.StdEncoding.EncodeToString([]byte(apiAccessID + ":" + apiAccessKey))
}

func withBasicAuth(basicAuthCredentials string) uhttp.RequestOption {
	return uhttp.WithHeader("Authorization", "Basic "+basicAuthCredentials)
}

// constructURL builds the full URL for an API request.
func (c *Client) constructURL(path string, pathParams map[string]string, queryParams map[string]string, pageToken *string, pageSize *uint) (*url.URL, error) {
	// Start with the base URL
	u := *c.apiBaseURL

	// Add the path
	if path != "" {
		// Replace path parameters
		for k, v := range pathParams {
			path = strings.ReplaceAll(path, "{"+k+"}", url.PathEscape(v))
		}
		u.Path += path
	}

	// Add query parameters
	q := u.Query()
	// Add token and limit parameters if provided
	if pageToken != nil {
		// API Doc: Continuation token to get the next page of results.
		// A page object with the next continuation token is returned in the response body.
		q.Set("token", *pageToken)
	}
	if pageSize != nil {
		// API Doc: Default value is 100 and the range is 1-100.
		q.Set("limit", fmt.Sprintf("%d", *pageSize))
	}
	// Add any additional query parameters
	for k, v := range queryParams {
		q.Set(k, v)
	}

	u.RawQuery = q.Encode()

	return &u, nil
}
