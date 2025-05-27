package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

// get performs a GET request to the API.
func (c *Client) get(
	ctx context.Context,
	url *url.URL,
	target interface{},
) (
	*v2.RateLimitDescription,
	error,
) {
	return c.doRequest(
		ctx,
		http.MethodGet,
		url,
		target,
		withBasicAuth(c.base64Credentials),
	)
}

func (c *Client) doRequest(
	ctx context.Context,
	method string,
	url *url.URL,
	target interface{},
	options ...uhttp.RequestOption,
) (
	*v2.RateLimitDescription,
	error,
) {
	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"making request",
		zap.String("method", method),
		zap.String("url", url.String()),
	)

	options = append(
		options,
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
	)

	request, err := c.httpClient.NewRequest(
		ctx,
		method,
		url,
		options...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var ratelimitData v2.RateLimitDescription
	response, err := c.httpClient.Do(
		request,
		uhttp.WithRatelimitData(&ratelimitData),
		uhttp.WithJSONResponse(&target),
		uhttp.WithErrorResponse(&ErrorResponse{}),
	)
	if err != nil {
		return &ratelimitData, fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	return &ratelimitData, nil
}
