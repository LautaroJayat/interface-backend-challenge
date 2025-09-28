package testclient

import (
	"context"
	"fmt"
	"time"

	"messaging-app/internal/domain"
	httpHandlers "messaging-app/internal/handlers/http"
	"messaging-app/testdata"
)

// TestUserManager helps manage multiple test users for e2e testing
type TestUserManager struct {
	clients map[string]*Client
	baseURL string
}

// NewTestUserManager creates a new test user manager
func NewTestUserManager(baseURL string) *TestUserManager {
	return &TestUserManager{
		clients: make(map[string]*Client),
		baseURL: baseURL,
	}
}

// GetClient returns a client for a specific user, creating it if necessary
func (m *TestUserManager) GetClient(user domain.UserContext) *Client {
	if client, exists := m.clients[user.UserID]; exists {
		return client
	}

	config := Config{
		BaseURL:     m.baseURL,
		Timeout:     30 * time.Second,
		UserContext: user,
	}

	client := NewClient(config)
	m.clients[user.UserID] = client
	return client
}

// GetAliceClient returns a client for the Alice test user
func (m *TestUserManager) GetAliceClient() *Client {
	return m.GetClient(testdata.Alice)
}

// GetBobClient returns a client for the Bob test user
func (m *TestUserManager) GetBobClient() *Client {
	return m.GetClient(testdata.Bob)
}

// GetCharlieClient returns a client for the Charlie test user
func (m *TestUserManager) GetCharlieClient() *Client {
	return m.GetClient(testdata.Charlie)
}

// GetDianaClient returns a client for the Diana test user
func (m *TestUserManager) GetDianaClient() *Client {
	return m.GetClient(testdata.Diana)
}

// GetEveClient returns a client for the Eve test user
func (m *TestUserManager) GetEveClient() *Client {
	return m.GetClient(testdata.Eve)
}

// Conversation represents a sequence of messages between users
type Conversation struct {
	manager *TestUserManager
}

// NewConversation creates a new conversation helper
func NewConversation(manager *TestUserManager) *Conversation {
	return &Conversation{manager: manager}
}

// SendMessage sends a message from one user to another
func (c *Conversation) SendMessage(ctx context.Context, from, to domain.UserContext, content string) (*httpHandlers.SendMessageResponse, error) {
	client := c.manager.GetClient(from)
	return client.SendMessage(ctx, to.UserID, content)
}

// SendMessageAndWait sends a message and waits for it to be processed
func (c *Conversation) SendMessageAndWait(ctx context.Context, from, to domain.UserContext, content string) (*httpHandlers.SendMessageResponse, error) {
	client := c.manager.GetClient(from)
	return client.SendAndWaitForMessage(ctx, to.UserID, content)
}

// WaitForMessageAsReceiver waits for a specific message to appear for the receiver
func (c *Conversation) WaitForMessageAsReceiver(ctx context.Context, receiver domain.UserContext, sender domain.UserContext, expectedContent string, timeout time.Duration) (*domain.Message, error) {
	client := c.manager.GetClient(receiver)
	chatID := domain.ComputeChatID(sender.UserID, receiver.UserID)
	return client.WaitForMessageInChat(ctx, chatID, expectedContent, timeout)
}

// MarkAsRead marks messages as read by the receiver
func (c *Conversation) MarkAsRead(ctx context.Context, receiver domain.UserContext, messageID domain.MessageID) (*httpHandlers.UpdateStatusResponse, error) {
	client := c.manager.GetClient(receiver)
	return client.UpdateMessageStatus(ctx, messageID)
}

// ChatHelper provides utilities for testing chat functionality
type ChatHelper struct {
	manager *TestUserManager
}

// NewChatHelper creates a new chat helper
func NewChatHelper(manager *TestUserManager) *ChatHelper {
	return &ChatHelper{manager: manager}
}

// CreateChatBetween creates a chat by having user1 send a message to user2
func (h *ChatHelper) CreateChatBetween(ctx context.Context, user1, user2 domain.UserContext, initialMessage string) (*httpHandlers.SendMessageResponse, error) {
	client := h.manager.GetClient(user1)
	return client.SendAndWaitForMessage(ctx, user2.UserID, initialMessage)
}

// GetChatID computes the chat ID between two users
func (h *ChatHelper) GetChatID(user1, user2 domain.UserContext) string {
	return domain.ComputeChatID(user1.UserID, user2.UserID)
}

// WaitForChatToAppear waits for a chat to appear in a user's chat list
func (h *ChatHelper) WaitForChatToAppear(ctx context.Context, user domain.UserContext, otherUser domain.UserContext, timeout time.Duration) (*domain.ChatSession, error) {
	client := h.manager.GetClient(user)
	chatID := h.GetChatID(user, otherUser)
	return client.WaitForChatToExist(ctx, chatID, timeout)
}

// VerifyChatOrder verifies that chats are ordered correctly by last message time
func (h *ChatHelper) VerifyChatOrder(chats []domain.ChatSession) error {
	for i := 1; i < len(chats); i++ {
		if chats[i].LastMessageAt.After(chats[i-1].LastMessageAt) {
			return fmt.Errorf("chats are not ordered correctly: chat %d (%s) has later timestamp than chat %d (%s)",
				i, chats[i].ChatID, i-1, chats[i-1].ChatID)
		}
	}
	return nil
}

// MessageHelper provides utilities for testing message functionality
type MessageHelper struct {
	manager *TestUserManager
}

// NewMessageHelper creates a new message helper
func NewMessageHelper(manager *TestUserManager) *MessageHelper {
	return &MessageHelper{manager: manager}
}

// SendSequentialMessages sends multiple messages in sequence with delays
func (h *MessageHelper) SendSequentialMessages(ctx context.Context, from, to domain.UserContext, messages []string, delay time.Duration) ([]*httpHandlers.SendMessageResponse, error) {
	client := h.manager.GetClient(from)
	responses := make([]*httpHandlers.SendMessageResponse, 0, len(messages))

	for i, content := range messages {
		if i > 0 {
			time.Sleep(delay)
		}

		resp, err := client.SendAndWaitForMessage(ctx, to.UserID, content)
		if err != nil {
			return responses, fmt.Errorf("failed to send message %d: %w", i+1, err)
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// VerifyMessageOrder verifies that messages are ordered correctly by timestamp
func (h *MessageHelper) VerifyMessageOrder(messages []domain.Message) error {
	for i := 1; i < len(messages); i++ {
		if messages[i].CreatedAt.Before(messages[i-1].CreatedAt) {
			return fmt.Errorf("messages are not ordered correctly: message %d has earlier timestamp than message %d",
				i, i-1)
		}
	}
	return nil
}

// FindMessageByContent finds a message with specific content in a list
func (h *MessageHelper) FindMessageByContent(messages []domain.Message, content string) *domain.Message {
	for _, msg := range messages {
		if msg.Content == content {
			return &msg
		}
	}
	return nil
}

// CountMessagesBySender counts messages from a specific sender
func (h *MessageHelper) CountMessagesBySender(messages []domain.Message, senderID string) int {
	count := 0
	for _, msg := range messages {
		if msg.SenderID == senderID {
			count++
		}
	}
	return count
}

// WaitHelper provides utilities for waiting and polling
type WaitHelper struct{}

// NewWaitHelper creates a new wait helper
func NewWaitHelper() *WaitHelper {
	return &WaitHelper{}
}

// WaitWithTimeout waits for a condition to be true within a timeout
func (h *WaitHelper) WaitWithTimeout(ctx context.Context, condition func() bool, timeout time.Duration, pollInterval time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
			// Continue polling
		}
	}

	return fmt.Errorf("condition not met within timeout of %v", timeout)
}

// Retry retries an operation with exponential backoff
func (h *WaitHelper) Retry(ctx context.Context, operation func() error, maxRetries int, baseDelay time.Duration) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := operation(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if i < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<uint(i)) // Exponential backoff
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue to next retry
			}
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, lastErr)
}