package platformapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

type Application struct {
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	Type                string               `json:"type"`
	Runtime             string               `json:"runtime"`
	OwningTeam          string               `json:"owning_team"`
	CreatedBy           PlatformUser         `json:"created_by"`
	CreatedAt           string               `json:"created_at"`
	ProvisioningRequest *ProvisioningRequest `json:"provisioning_request"`
}

type ProvisioningRequest struct {
	ID            string `json:"id"`
	Application   string `json:"application"`
	Status        string `json:"status"`
	RequestedByID string `json:"requested_by_id"`
	CreatedAt     string `json:"created_at"`
	Error         string `json:"error"`
}

type PlatformUser struct {
	ID            string `json:"id"`
	EntraObjectID string `json:"entra_object_id"`
	TenantID      string `json:"tenant_id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
}

type CreateApplicationRequest struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Runtime    string `json:"runtime"`
	OwningTeam string `json:"owning_team"`
}

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("API request failed (%d): %s", e.StatusCode, e.Message)
}

type Client struct {
	baseURL     string
	tokenSource oauth2.TokenSource
	httpClient  *http.Client
}

func NewClient(baseURL string, tokenSource oauth2.TokenSource) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), tokenSource: tokenSource, httpClient: http.DefaultClient}
}

func NewClientWithHTTPClient(baseURL string, tokenSource oauth2.TokenSource, httpClient *http.Client) *Client {
	client := NewClient(baseURL, tokenSource)
	client.httpClient = httpClient
	return client
}

func (c *Client) Me(ctx context.Context) (PlatformUser, error) {
	var user PlatformUser
	err := c.get(ctx, "/api/v1/me", &user)
	return user, err
}
func (c *Client) ListApplications(ctx context.Context) ([]Application, error) {
	var apps []Application
	err := c.get(ctx, "/api/v1/applications", &apps)
	return apps, err
}
func (c *Client) GetApplication(ctx context.Context, id string) (Application, error) {
	var app Application
	err := c.get(ctx, "/api/v1/applications/"+id, &app)
	return app, err
}
func (c *Client) GetProvisioning(ctx context.Context, id string) (ProvisioningRequest, error) {
	var request ProvisioningRequest
	err := c.get(ctx, "/api/v1/provisioning/"+id, &request)
	return request, err
}

func (c *Client) CreateApplication(ctx context.Context, input CreateApplicationRequest) (Application, error) {
	var app Application
	err := c.send(ctx, http.MethodPost, "/api/v1/applications", input, &app)
	return app, err
}

func (c *Client) get(ctx context.Context, path string, output any) error {
	return c.send(ctx, http.MethodGet, path, nil, output)
}

func (c *Client) send(ctx context.Context, method, path string, input, output any) error {
	token, err := c.tokenSource.Token()
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = strings.NewReader(string(encoded))
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call platform API: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode >= 300 {
		raw, _ := io.ReadAll(response.Body)
		return &Error{StatusCode: response.StatusCode, Message: strings.TrimSpace(string(raw))}
	}
	if output == nil {
		return nil
	}
	return json.NewDecoder(response.Body).Decode(output)
}
