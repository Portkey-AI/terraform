package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client manages communication with the Portkey Admin API
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Portkey API client
func NewClient(baseURL, apiKey string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// doRequest performs an HTTP request
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-portkey-api-key", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Workspace represents a Portkey workspace
type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateWorkspaceRequest represents the request to create a workspace
type CreateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateWorkspaceRequest represents the request to update a workspace
type UpdateWorkspaceRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateWorkspace creates a new workspace
func (c *Client) CreateWorkspace(ctx context.Context, req CreateWorkspaceRequest) (*Workspace, error) {
	respBody, err := c.doRequest(ctx, http.MethodPost, "/admin/workspaces", req)
	if err != nil {
		return nil, err
	}

	var workspace Workspace
	if err := json.Unmarshal(respBody, &workspace); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &workspace, nil
}

// GetWorkspace retrieves a workspace by ID
func (c *Client) GetWorkspace(ctx context.Context, id string) (*Workspace, error) {
	respBody, err := c.doRequest(ctx, http.MethodGet, "/admin/workspaces/"+id, nil)
	if err != nil {
		return nil, err
	}

	var workspace Workspace
	if err := json.Unmarshal(respBody, &workspace); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &workspace, nil
}

// ListWorkspaces retrieves all workspaces
func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	respBody, err := c.doRequest(ctx, http.MethodGet, "/admin/workspaces", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []Workspace `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return response.Data, nil
}

// UpdateWorkspace updates a workspace
func (c *Client) UpdateWorkspace(ctx context.Context, id string, req UpdateWorkspaceRequest) (*Workspace, error) {
	respBody, err := c.doRequest(ctx, http.MethodPut, "/admin/workspaces/"+id, req)
	if err != nil {
		return nil, err
	}

	var workspace Workspace
	if err := json.Unmarshal(respBody, &workspace); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &workspace, nil
}

// DeleteWorkspaceRequest represents the request to delete a workspace
type DeleteWorkspaceRequest struct {
	Name        string `json:"name"`
	ForceDelete bool   `json:"force_delete,omitempty"`
}

// DeleteWorkspace deletes a workspace
func (c *Client) DeleteWorkspace(ctx context.Context, id string, name string) error {
	req := DeleteWorkspaceRequest{
		Name: name,
	}
	_, err := c.doRequest(ctx, http.MethodDelete, "/admin/workspaces/"+id, req)
	return err
}

// User represents a Portkey user
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetUser retrieves a user by ID
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	respBody, err := c.doRequest(ctx, http.MethodGet, "/admin/users/"+id, nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &user, nil
}

// ListUsers retrieves all users
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	respBody, err := c.doRequest(ctx, http.MethodGet, "/admin/users", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []User `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return response.Data, nil
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Role string `json:"role,omitempty"`
}

// UpdateUser updates a user
func (c *Client) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	respBody, err := c.doRequest(ctx, http.MethodPut, "/admin/users/"+id, req)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &user, nil
}

// DeleteUser removes a user
func (c *Client) DeleteUser(ctx context.Context, id string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, "/admin/users/"+id, nil)
	return err
}

// WorkspaceMember represents a workspace member
type WorkspaceMember struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	WorkspaceID string    `json:"workspace_id"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

// AddWorkspaceMemberRequest represents the request to add a workspace member
type AddWorkspaceMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

// AddWorkspaceMember adds a member to a workspace
func (c *Client) AddWorkspaceMember(ctx context.Context, workspaceID string, req AddWorkspaceMemberRequest) (*WorkspaceMember, error) {
	path := fmt.Sprintf("/admin/workspaces/%s/members", workspaceID)
	respBody, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}

	var member WorkspaceMember
	if err := json.Unmarshal(respBody, &member); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &member, nil
}

// GetWorkspaceMember retrieves a workspace member
func (c *Client) GetWorkspaceMember(ctx context.Context, workspaceID, memberID string) (*WorkspaceMember, error) {
	path := fmt.Sprintf("/admin/workspaces/%s/members/%s", workspaceID, memberID)
	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var member WorkspaceMember
	if err := json.Unmarshal(respBody, &member); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &member, nil
}

// ListWorkspaceMembers retrieves all members of a workspace
func (c *Client) ListWorkspaceMembers(ctx context.Context, workspaceID string) ([]WorkspaceMember, error) {
	path := fmt.Sprintf("/admin/workspaces/%s/members", workspaceID)
	respBody, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []WorkspaceMember `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return response.Data, nil
}

// UpdateWorkspaceMemberRequest represents the request to update a workspace member
type UpdateWorkspaceMemberRequest struct {
	Role string `json:"role"`
}

// UpdateWorkspaceMember updates a workspace member's role
func (c *Client) UpdateWorkspaceMember(ctx context.Context, workspaceID, memberID string, req UpdateWorkspaceMemberRequest) (*WorkspaceMember, error) {
	path := fmt.Sprintf("/admin/workspaces/%s/members/%s", workspaceID, memberID)
	respBody, err := c.doRequest(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, err
	}

	var member WorkspaceMember
	if err := json.Unmarshal(respBody, &member); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &member, nil
}

// RemoveWorkspaceMember removes a member from a workspace
func (c *Client) RemoveWorkspaceMember(ctx context.Context, workspaceID, memberID string) error {
	path := fmt.Sprintf("/admin/workspaces/%s/members/%s", workspaceID, memberID)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

// UserInvite represents a user invitation
type UserInvite struct {
	ID         string                   `json:"id"`
	Email      string                   `json:"email"`
	Role       string                   `json:"role"`
	Status     string                   `json:"status"`
	Workspaces []WorkspaceInviteDetails `json:"workspaces,omitempty"`
	CreatedAt  time.Time                `json:"created_at"`
	ExpiresAt  time.Time                `json:"expires_at"`
}

// WorkspaceInviteDetails represents workspace details in an invitation
type WorkspaceInviteDetails struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

// CreateUserInviteRequest represents the request to invite a user
type CreateUserInviteRequest struct {
	Email                  string                   `json:"email"`
	Role                   string                   `json:"role"`
	Workspaces             []WorkspaceInviteDetails `json:"workspaces,omitempty"`
	WorkspaceAPIKeyDetails *APIKeyDetails           `json:"workspace_api_key_details,omitempty"`
}

// APIKeyDetails represents API key configuration for user invites
type APIKeyDetails struct {
	Scopes []string `json:"scopes"`
}

// InviteUser sends an invitation to a user
func (c *Client) InviteUser(ctx context.Context, req CreateUserInviteRequest) (*UserInvite, error) {
	respBody, err := c.doRequest(ctx, http.MethodPost, "/admin/users/invites", req)
	if err != nil {
		return nil, err
	}

	var invite UserInvite
	if err := json.Unmarshal(respBody, &invite); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &invite, nil
}

// GetUserInvite retrieves a user invitation by ID
func (c *Client) GetUserInvite(ctx context.Context, id string) (*UserInvite, error) {
	respBody, err := c.doRequest(ctx, http.MethodGet, "/admin/users/invites/"+id, nil)
	if err != nil {
		return nil, err
	}

	var invite UserInvite
	if err := json.Unmarshal(respBody, &invite); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &invite, nil
}

// ListUserInvites retrieves all user invitations
func (c *Client) ListUserInvites(ctx context.Context) ([]UserInvite, error) {
	respBody, err := c.doRequest(ctx, http.MethodGet, "/admin/users/invites", nil)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []UserInvite `json:"data"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return response.Data, nil
}

// DeleteUserInvite deletes a user invitation
func (c *Client) DeleteUserInvite(ctx context.Context, id string) error {
	_, err := c.doRequest(ctx, http.MethodDelete, "/admin/users/invites/"+id, nil)
	return err
}
