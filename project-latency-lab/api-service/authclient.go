package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// AuthClient talks to the auth-service to verify caller tokens. The HTTP
// transport is otelhttp-instrumented so the call shows up as a child span of
// the current handler and propagates traceparent downstream.
type AuthClient struct {
	baseURL string
	http    *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 60 * time.Second, Transport: tracedHTTPClient().Transport},
	}
}

var ErrUnauthorized = errors.New("unauthorized")

// Verify exchanges a caller token for a username. Returns ErrUnauthorized
// on a 401 from auth-service. The latency-knob query string from the
// inbound request is forwarded so injected latency / faults propagate.
func (c *AuthClient) Verify(ctx context.Context, token, rawQuery string) (string, error) {
	ctx, span := tracer.Start(ctx, "authclient.Verify")
	defer span.End()

	if c == nil || c.baseURL == "" {
		return "", fmt.Errorf("auth client not configured")
	}
	body, _ := json.Marshal(map[string]string{"token": token})

	parsed, err := url.Parse(c.baseURL + "/api/auth/verify")
	if err != nil {
		return "", err
	}
	parsed.RawQuery = rawQuery
	u := parsed.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		logger(ctx).Error("auth call network error", "url", u, "error", err)
		return "", fmt.Errorf("auth call: %w", err)
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("auth.status", resp.StatusCode))
	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrUnauthorized
	}
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(resp.Body)
		logger(ctx).Error("auth call non-2xx", "url", u, "status", resp.StatusCode, "body", string(b))
		return "", fmt.Errorf("auth %d: %s", resp.StatusCode, string(b))
	}

	var out struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	span.SetAttributes(attribute.String("username", out.Username))
	return out.Username, nil
}
