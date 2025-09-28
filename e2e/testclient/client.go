package testclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	httpAdapter "messaging-app/internal/adapters/http"
	"messaging-app/internal/domain"
	httpHandlers "messaging-app/internal/handlers/http"
)

// Client represents an HTTP test client for the messaging API
type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	UserContext domain.UserContext
	AuthConfig  httpAdapter.AuthConfig
}

// Config holds configuration for the test client
type Config struct {
	BaseURL     string
	Timeout     time.Duration
	UserContext domain.UserContext
	AuthConfig  httpAdapter.AuthConfig
}

// NewClient creates a new API test client
func NewClient(config Config) *Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Use default auth config if not provided
	authConfig := config.AuthConfig
	if authConfig.UserIDHeader == "" {
		authConfig = DefaultAuthConfig()
	}

	return &Client{
		BaseURL: config.BaseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		UserContext: config.UserContext,
		AuthConfig:  authConfig,
	}
}

// DefaultAuthConfig returns the default authentication configuration
// This should match the server's default configuration
func DefaultAuthConfig() httpAdapter.AuthConfig {
	return httpAdapter.AuthConfig{
		UserIDHeader:  "x-interface-user-id",
		EmailHeader:   "x-interface-user-email",
		HandlerHeader: "x-interface-user-handler",
	}
}

// SetUser updates the user context for subsequent requests
func (c *Client) SetUser(user domain.UserContext) {
	c.UserContext = user
}

// makeRequest is a helper method to make HTTP requests with proper headers
func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := c.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set authentication headers using configured header names
	req.Header.Set(c.AuthConfig.UserIDHeader, c.UserContext.UserID)
	req.Header.Set(c.AuthConfig.EmailHeader, c.UserContext.Email)
	req.Header.Set(c.AuthConfig.HandlerHeader, c.UserContext.Handler)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// parseResponse is a helper to parse JSON responses
func (c *Client) parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResp httpAdapter.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    errorResp.Error,
				Code:       errorResp.Code,
				Details:    errorResp.Details,
			}
		}
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
			Code:       "HTTP_ERROR",
			Details:    string(body),
		}
	}

	if target != nil {
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Health checks the health endpoint
func (c *Client) Health(ctx context.Context) error {
	resp, err := c.makeRequest(ctx, "GET", "/health", nil)
	if err != nil {
		return err
	}

	var healthResp map[string]string
	return c.parseResponse(resp, &healthResp)
}

// SendMessage sends a message to a specific receiver
func (c *Client) SendMessage(ctx context.Context, receiverID, content string) (*httpHandlers.SendMessageResponse, error) {
	req := httpHandlers.SendMessageRequest{
		Content: content,
	}

	path := fmt.Sprintf("/api/v1/chats/%s/messages", receiverID)
	resp, err := c.makeRequest(ctx, "POST", path, req)
	if err != nil {
		return nil, err
	}

	var response httpHandlers.SendMessageResponse
	err = c.parseResponse(resp, &response)
	return &response, err
}

// GetMessages retrieves messages for a chat
func (c *Client) GetMessages(ctx context.Context, chatID string, options *GetMessagesOptions) (*httpHandlers.GetMessagesResponse, error) {
	path := fmt.Sprintf("/api/v1/chats/%s/messages", chatID)

	if options != nil {
		params := url.Values{}
		if options.Cursor != "" {
			params.Add("cursor", options.Cursor)
		}
		if options.Limit > 0 {
			params.Add("limit", fmt.Sprintf("%d", options.Limit))
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response httpHandlers.GetMessagesResponse
	err = c.parseResponse(resp, &response)
	return &response, err
}

// UpdateMessageStatus updates the status of messages
func (c *Client) UpdateMessageStatus(ctx context.Context, messageID domain.MessageID) (*httpHandlers.UpdateStatusResponse, error) {
	req := httpHandlers.UpdateStatusRequest{
		MessageID: messageID,
	}

	resp, err := c.makeRequest(ctx, "PATCH", "/api/v1/messages/status", req)
	if err != nil {
		return nil, err
	}

	var response httpHandlers.UpdateStatusResponse
	err = c.parseResponse(resp, &response)
	return &response, err
}

// GetChats retrieves all chat sessions for the current user
func (c *Client) GetChats(ctx context.Context) (*httpHandlers.GetChatsResponse, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/v1/chats", nil)
	if err != nil {
		return nil, err
	}

	var response httpHandlers.GetChatsResponse
	err = c.parseResponse(resp, &response)
	return &response, err
}

// Convenience methods for common operations

// SendAndWaitForMessage sends a message and waits for it to be sent
func (c *Client) SendAndWaitForMessage(ctx context.Context, receiverID, content string) (*httpHandlers.SendMessageResponse, error) {
	response, err := c.SendMessage(ctx, receiverID, content)
	if err != nil {
		return nil, err
	}

	// Add small delay to allow for processing
	time.Sleep(100 * time.Millisecond)
	return response, nil
}

// GetLatestMessages retrieves the most recent messages for a chat
func (c *Client) GetLatestMessages(ctx context.Context, chatID string, limit int) (*httpHandlers.GetMessagesResponse, error) {
	options := &GetMessagesOptions{
		Limit: limit,
	}
	return c.GetMessages(ctx, chatID, options)
}

// MarkAllMessagesAsRead marks all messages in a chat as read up to a specific message
func (c *Client) MarkAllMessagesAsRead(ctx context.Context, messageID domain.MessageID) (*httpHandlers.UpdateStatusResponse, error) {
	return c.UpdateMessageStatus(ctx, messageID)
}

// WaitForChatToExist polls until a chat appears in the user's chat list
func (c *Client) WaitForChatToExist(ctx context.Context, expectedChatID string, timeout time.Duration) (*domain.ChatSession, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		chats, err := c.GetChats(ctx)
		if err != nil {
			return nil, err
		}

		for _, chat := range chats.Chats {
			if chat.ChatID == expectedChatID {
				return &chat, nil
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("chat %s did not appear within timeout", expectedChatID)
}

// WaitForMessageInChat polls until a specific message appears in a chat
func (c *Client) WaitForMessageInChat(ctx context.Context, chatID, expectedContent string, timeout time.Duration) (*domain.Message, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		messages, err := c.GetLatestMessages(ctx, chatID, 10)
		if err != nil {
			return nil, err
		}

		for _, message := range messages.Messages {
			if message.Content == expectedContent {
				return &message, nil
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("message with content '%s' did not appear in chat %s within timeout", expectedContent, chatID)
}