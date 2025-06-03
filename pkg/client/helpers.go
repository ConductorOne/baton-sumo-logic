package client

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

// constructURL builds the full URL for an API request.
func (c *Client) constructURL(path string, pathParams map[string]string, queryParams map[string]string, pageToken *string, pageSize *uint) (*url.URL, error) {
	// Start with the base URL
	u := *c.apiBaseURL

	// Add the path parameters
	if path != "" {
		// Create a template for path parameter replacement
		tmpl, err := template.New("path").Parse(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse path template: %w", err)
		}

		// Create a buffer to hold the result
		var buf strings.Builder

		// Execute the template with path parameters
		if err := tmpl.Execute(&buf, pathParams); err != nil {
			return nil, fmt.Errorf("failed to execute path template: %w", err)
		}

		// Use the processed path
		u.Path += buf.String()
	}

	// Add pagination query parameters
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
