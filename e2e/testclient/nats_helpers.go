package testclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"messaging-app/internal/domain"
)

// NATSTestManager manages multiple NATS clients for different users in tests
type NATSTestManager struct {
	clients map[string]*NATSClient
	baseURL string
	mu      sync.RWMutex
}

// NewNATSTestManager creates a new NATS test manager
func NewNATSTestManager(natsURL string) *NATSTestManager {
	return &NATSTestManager{
		clients: make(map[string]*NATSClient),
		baseURL: natsURL,
	}
}

// GetClient returns a NATS client for a specific user, creating it if necessary
func (m *NATSTestManager) GetClient(userID string) (*NATSClient, error) {
	m.mu.RLock()
	if client, exists := m.clients[userID]; exists {
		m.mu.RUnlock()
		return client, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := m.clients[userID]; exists {
		return client, nil
	}

	config := NATSConfig{
		URL:           m.baseURL,
		UserID:        userID,
		Timeout:       30 * time.Second,
		MaxReconnects: 5,
	}

	client, err := NewNATSClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS client for user %s: %w", userID, err)
	}

	m.clients[userID] = client
	return client, nil
}

// GetAliceNATSClient returns a NATS client for Alice
func (m *NATSTestManager) GetAliceNATSClient() (*NATSClient, error) {
	return m.GetClient("alice_123")
}

// GetBobNATSClient returns a NATS client for Bob
func (m *NATSTestManager) GetBobNATSClient() (*NATSClient, error) {
	return m.GetClient("bob_456")
}

// GetCharlieNATSClient returns a NATS client for Charlie
func (m *NATSTestManager) GetCharlieNATSClient() (*NATSClient, error) {
	return m.GetClient("charlie_789")
}

// CloseAll closes all NATS clients
func (m *NATSTestManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for userID, client := range m.clients {
		if err := client.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close client for %s: %w", userID, err)
		}
	}

	m.clients = make(map[string]*NATSClient)
	return lastErr
}

// MessageCollector helps collect and verify messages in tests
type MessageCollector struct {
	messages []domain.MessageEnvelope
	mu       sync.Mutex
}

// NewMessageCollector creates a new message collector
func NewMessageCollector() *MessageCollector {
	return &MessageCollector{
		messages: make([]domain.MessageEnvelope, 0),
	}
}

// Handler returns a message handler that collects messages
func (c *MessageCollector) Handler() MessageHandler {
	return func(envelope *domain.MessageEnvelope) error {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.messages = append(c.messages, *envelope)
		return nil
	}
}

// GetMessages returns all collected messages
func (c *MessageCollector) GetMessages() []domain.MessageEnvelope {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Return a copy to avoid race conditions
	result := make([]domain.MessageEnvelope, len(c.messages))
	copy(result, c.messages)
	return result
}

// GetMessageCount returns the number of collected messages
func (c *MessageCollector) GetMessageCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.messages)
}

// FindMessageByContent finds a message with specific content
func (c *MessageCollector) FindMessageByContent(content string) *domain.MessageEnvelope {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, msg := range c.messages {
		if msg.Data.Content == content {
			return &msg
		}
	}
	return nil
}

// Clear removes all collected messages
func (c *MessageCollector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = c.messages[:0]
}

// WaitForMessageCount waits until the collector has the expected number of messages
func (c *MessageCollector) WaitForMessageCount(ctx context.Context, expectedCount int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if c.GetMessageCount() >= expectedCount {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			// Continue checking
		}
	}

	return fmt.Errorf("timeout waiting for %d messages, got %d", expectedCount, c.GetMessageCount())
}

// StatusCollector helps collect and verify status updates in tests
type StatusCollector struct {
	updates []domain.StatusUpdateEnvelope
	mu      sync.Mutex
}

// NewStatusCollector creates a new status collector
func NewStatusCollector() *StatusCollector {
	return &StatusCollector{
		updates: make([]domain.StatusUpdateEnvelope, 0),
	}
}

// Handler returns a status handler that collects status updates
func (c *StatusCollector) Handler() StatusHandler {
	return func(envelope *domain.StatusUpdateEnvelope) error {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.updates = append(c.updates, *envelope)
		return nil
	}
}

// GetUpdates returns all collected status updates
func (c *StatusCollector) GetUpdates() []domain.StatusUpdateEnvelope {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Return a copy to avoid race conditions
	result := make([]domain.StatusUpdateEnvelope, len(c.updates))
	copy(result, c.updates)
	return result
}

// GetUpdateCount returns the number of collected status updates
func (c *StatusCollector) GetUpdateCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.updates)
}

// Clear removes all collected status updates
func (c *StatusCollector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.updates = c.updates[:0]
}

// NATSTestScenario provides utilities for common NATS testing scenarios
type NATSTestScenario struct {
	manager *NATSTestManager
}

// NewNATSTestScenario creates a new NATS test scenario helper
func NewNATSTestScenario(natsURL string) *NATSTestScenario {
	return &NATSTestScenario{
		manager: NewNATSTestManager(natsURL),
	}
}

// SetupMessageSubscription sets up a message subscription for a user with a collector
func (s *NATSTestScenario) SetupMessageSubscription(ctx context.Context, userID string) (*NATSClient, *MessageCollector, error) {
	client, err := s.manager.GetClient(userID)
	if err != nil {
		return nil, nil, err
	}

	collector := NewMessageCollector()
	if err := client.SubscribeToMessages(ctx, collector.Handler()); err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to messages: %w", err)
	}

	return client, collector, nil
}

// SetupStatusSubscription sets up a status subscription for a user with a collector
func (s *NATSTestScenario) SetupStatusSubscription(ctx context.Context, userID string) (*NATSClient, *StatusCollector, error) {
	client, err := s.manager.GetClient(userID)
	if err != nil {
		return nil, nil, err
	}

	collector := NewStatusCollector()
	if err := client.SubscribeToStatus(ctx, collector.Handler()); err != nil {
		return nil, nil, fmt.Errorf("failed to subscribe to status: %w", err)
	}

	return client, collector, nil
}

// WaitForAllClientsConnected waits for all managed clients to be connected
func (s *NATSTestScenario) WaitForAllClientsConnected(timeout time.Duration) error {
	s.manager.mu.RLock()
	defer s.manager.mu.RUnlock()

	for userID, client := range s.manager.clients {
		if err := client.WaitForConnection(timeout); err != nil {
			return fmt.Errorf("failed to connect client for %s: %w", userID, err)
		}
	}

	return nil
}

// GetManager returns the underlying NATS test manager
func (s *NATSTestScenario) GetManager() *NATSTestManager {
	return s.manager
}

// Close closes all clients in the scenario
func (s *NATSTestScenario) Close() error {
	return s.manager.CloseAll()
}

// DefaultNATSURL returns the default NATS WebSocket URL for testing
func DefaultNATSURL() string {
	return "ws://localhost:8080"
}