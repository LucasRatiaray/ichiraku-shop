package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func (c *Client) do(ctx context.Context, method, path, token string, body, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("api: %s %s: %w", method, path, err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("api: %s %s: %w", method, path, err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("api: %s %s: %w", method, path, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		return &APIError{
			Status:  resp.StatusCode,
			Message: apiErr.Error,
		}
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("api: %s %s: %w", method, path, err)
		}
	}

	return nil
}

func (c *Client) Register(ctx context.Context, email, password string) error {
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	return c.do(ctx, http.MethodPost, "/auth/register", "", body, nil)
}

func (c *Client) Confirm(ctx context.Context, email, code string) error {
	body := map[string]string{
		"email": email,
		"code":  code,
	}
	return c.do(ctx, http.MethodPost, "/auth/confirm", "", body, nil)
}

func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	body := map[string]string{
		"email":    email,
		"password": password,
	}
	var res struct {
		Token string `json:"token"`
	}

	if err := c.do(ctx, http.MethodPost, "/auth/login", "", body, &res); err != nil {
		return "", err
	}
	return res.Token, nil
}

func (c *Client) Logout(ctx context.Context, token string) error {
	return c.do(ctx, http.MethodPost, "/auth/logout", token, nil, nil)
}

func (c *Client) ForgotPassword(ctx context.Context, email string) error {
	body := map[string]string{"email": email}
	return c.do(ctx, http.MethodPost, "/auth/password/forgot", "", body, nil)
}

func (c *Client) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	body := map[string]string{
		"email":        email,
		"code":         code,
		"new_password": newPassword,
	}
	return c.do(ctx, http.MethodPost, "/auth/password/reset", "", body, nil)
}
