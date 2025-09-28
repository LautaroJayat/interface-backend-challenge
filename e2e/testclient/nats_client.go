package testclient

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"messaging-app/internal/domain"

	"github.com/nats-io/nats.go"
)

// NATSClient represents a NATS WebSocket test client for consuming messages
type NATSClient struct {
	conn         *nats.Conn
	url          string
	userID       string
	subscriptions map[string]*nats.Subscription
	mu           sync.RWMutex
	messageHandlers map[string][]MessageHandler
}

// MessageHandler defines a function type for handling incoming messages
type MessageHandler func(envelope *domain.MessageEnvelope) error

// StatusHandler defines a function type for handling status updates
type StatusHandler func(envelope *domain.StatusUpdateEnvelope) error

// NATSConfig holds configuration for the NATS test client
type NATSConfig struct {
	URL    string
	UserID string
	// Optional connection options
	Timeout time.Duration
	MaxReconnects int
}

// NewNATSClient creates a new NATS WebSocket test client
func NewNATSClient(config NATSConfig) (*NATSClient, error) {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	maxReconnects := config.MaxReconnects
	if maxReconnects == 0 {
		maxReconnects = 5
	}

	// Configure NATS connection options
	opts := []nats.Option{
		nats.Timeout(timeout),
		nats.MaxReconnects(maxReconnects),
		nats.ReconnectWait(1 * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				fmt.Printf("NATS disconnected: %v\n", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("NATS reconnected to %s\n", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			fmt.Printf("NATS connection closed\n")
		}),
	}

	// Connect to NATS
	nc, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	client := &NATSClient{
		conn:            nc,
		url:             config.URL,
		userID:          config.UserID,
		subscriptions:   make(map[string]*nats.Subscription),
		messageHandlers: make(map[string][]MessageHandler),
	}

	return client, nil
}

// DefaultNATSConfig returns default NATS configuration for testing
func DefaultNATSConfig(userID string) NATSConfig {
	return NATSConfig{
		URL:           "ws://localhost:8080", // WebSocket connection
		UserID:        userID,
		Timeout:       30 * time.Second,
		MaxReconnects: 5,
	}
}

// SubscribeToMessages subscribes to message updates for the user
func (c *NATSClient) SubscribeToMessages(ctx context.Context, handler MessageHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	topic := domain.GetMessageTopic(c.userID)

	// Add handler to the list
	c.messageHandlers[topic] = append(c.messageHandlers[topic], handler)

	// If already subscribed, just add the handler
	if _, exists := c.subscriptions[topic]; exists {
		return nil
	}

	// Create subscription
	sub, err := c.conn.Subscribe(topic, func(msg *nats.Msg) {
		var envelope domain.MessageEnvelope
		if err := json.Unmarshal(msg.Data, &envelope); err != nil {
			fmt.Printf("Failed to unmarshal message envelope: %v\n", err)
			return
		}

		// Call all handlers for this topic
		c.mu.RLock()
		handlers := c.messageHandlers[topic]
		c.mu.RUnlock()

		for _, h := range handlers {
			if err := h(&envelope); err != nil {
				fmt.Printf("Message handler error: %v\n", err)
			}
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	c.subscriptions[topic] = sub
	return nil
}

// SubscribeToStatus subscribes to status updates for the user
func (c *NATSClient) SubscribeToStatus(ctx context.Context, handler StatusHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	topic := domain.GetStatusTopic(c.userID)

	// If already subscribed, return error (status subscriptions are singular)
	if _, exists := c.subscriptions[topic]; exists {
		return fmt.Errorf("already subscribed to status topic %s", topic)
	}

	// Create subscription
	sub, err := c.conn.Subscribe(topic, func(msg *nats.Msg) {
		var envelope domain.StatusUpdateEnvelope
		if err := json.Unmarshal(msg.Data, &envelope); err != nil {
			fmt.Printf("Failed to unmarshal status envelope: %v\n", err)
			return
		}

		if err := handler(&envelope); err != nil {
			fmt.Printf("Status handler error: %v\n", err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to status topic %s: %w", topic, err)
	}

	c.subscriptions[topic] = sub
	return nil
}

// SubscribeToTopic subscribes to a custom topic with raw message handling
func (c *NATSClient) SubscribeToTopic(ctx context.Context, topic string, handler func([]byte) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If already subscribed, return error
	if _, exists := c.subscriptions[topic]; exists {
		return fmt.Errorf("already subscribed to topic %s", topic)
	}

	// Create subscription
	sub, err := c.conn.Subscribe(topic, func(msg *nats.Msg) {
		if err := handler(msg.Data); err != nil {
			fmt.Printf("Topic handler error for %s: %v\n", topic, err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	c.subscriptions[topic] = sub
	return nil
}

// Unsubscribe removes subscription from a topic
func (c *NATSClient) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sub, exists := c.subscriptions[topic]
	if !exists {
		return fmt.Errorf("not subscribed to topic %s", topic)
	}

	if err := sub.Unsubscribe(); err != nil {
		return fmt.Errorf("failed to unsubscribe from topic %s: %w", topic, err)
	}

	delete(c.subscriptions, topic)
	delete(c.messageHandlers, topic)
	return nil
}

// WaitForMessage waits for a specific message to arrive within timeout
func (c *NATSClient) WaitForMessage(ctx context.Context, expectedContent string, timeout time.Duration) (*domain.MessageEnvelope, error) {
	resultChan := make(chan *domain.MessageEnvelope, 1)
	errorChan := make(chan error, 1)

	// Create a temporary handler that looks for the expected content
	handler := func(envelope *domain.MessageEnvelope) error {
		if envelope.Data.Content == expectedContent {
			select {
			case resultChan <- envelope:
			default:
				// Channel already has a result
			}
		}
		return nil
	}

	// Subscribe to messages
	if err := c.SubscribeToMessages(ctx, handler); err != nil {
		return nil, fmt.Errorf("failed to subscribe for message waiting: %w", err)
	}

	// Wait for result or timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case envelope := <-resultChan:
		return envelope, nil
	case err := <-errorChan:
		return nil, err
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("timeout waiting for message with content '%s'", expectedContent)
	}
}

// GetConnectionStatus returns the current connection status
func (c *NATSClient) GetConnectionStatus() nats.Status {
	return c.conn.Status()
}

// IsConnected returns true if the client is connected to NATS
func (c *NATSClient) IsConnected() bool {
	return c.conn.IsConnected()
}

// GetActiveSubscriptions returns a list of active subscription topics
func (c *NATSClient) GetActiveSubscriptions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	topics := make([]string, 0, len(c.subscriptions))
	for topic := range c.subscriptions {
		topics = append(topics, topic)
	}
	return topics
}

// Close closes all subscriptions and the NATS connection
func (c *NATSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Unsubscribe from all topics
	for topic, sub := range c.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			fmt.Printf("Error unsubscribing from %s: %v\n", topic, err)
		}
	}

	// Clear subscriptions and handlers
	c.subscriptions = make(map[string]*nats.Subscription)
	c.messageHandlers = make(map[string][]MessageHandler)

	// Close NATS connection
	c.conn.Close()
	return nil
}

// Flush ensures all published messages have been processed by the server
func (c *NATSClient) Flush() error {
	return c.conn.Flush()
}

// WaitForConnection waits for the NATS connection to be established
func (c *NATSClient) WaitForConnection(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for NATS connection")
		case <-ticker.C:
			if c.IsConnected() {
				return nil
			}
		}
	}
}