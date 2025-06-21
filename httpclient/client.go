// Package httpclient provides a configurable HTTP client with sensible defaults
// and helper methods for making HTTP requests.
package httpclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

var (
	// ErrWrongStatusCode is returned when the server responds with an unexpected HTTP status code.
	ErrWrongStatusCode = errors.New("wrong status code")

	// ErrEmptyResponse is returned when the server response is nil and error is nil.
	ErrEmptyResponse = errors.New("empty response")
)

// Client wraps http.Client to provide additional functionality and configuration.
// It embeds the standard http.Client to expose all its methods while adding custom behavior.
type Client struct {
	http.Client
}

// New creates and returns a new Client instance configured with the given settings.
// The configuration includes timeouts, transport settings, and dialer parameters.
func New(cfg *Config) *Client {
	return &Client{
		http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:       cfg.Dialer.Timeout,
					Deadline:      cfg.Dialer.Deadline,
					FallbackDelay: cfg.Dialer.FallbackDelay,
					KeepAlive:     cfg.Dialer.KeepAlive,
				}).DialContext,
				TLSHandshakeTimeout: cfg.TLSHandshakeTimeout,
			},
		},
	}
}

// SendRequest executes an HTTP request with the given method, URI, and optional body.
// It handles the full request lifecycle including context cancellation, error handling,
// and response processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - method: HTTP method (GET, POST, etc.)
//   - uri: Target URL for the request
//   - body: Request body content (can be nil)
//
// Returns:
//   - Response body as byte slice if successful
//   - Error if request fails, response status is not 200 OK, or body read fails.
func (c *Client) SendRequest(ctx context.Context, method, uri string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, uri, body)
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, ErrEmptyResponse
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d %s", ErrWrongStatusCode, res.StatusCode, res.Status)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resBody, nil
}
