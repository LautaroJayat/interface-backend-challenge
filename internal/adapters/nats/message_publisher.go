package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"

	"messaging-app/internal/domain"
	"messaging-app/internal/ports"
)

type NATSMessagePublisher struct {
	conn   *nats.Conn
	logger ports.Logger
}

func NewNATSMessagePublisher(conn *nats.Conn, logger ports.Logger) *NATSMessagePublisher {
	return &NATSMessagePublisher{
		conn:   conn,
		logger: logger,
	}
}

// PublishMessage implements ports.MessagePublisher
func (p *NATSMessagePublisher) PublishMessage(ctx context.Context, message domain.Message) error {
	subject := domain.GetMessageTopic(message.ReceiverID)

	// Create message envelope with metadata
	envelope := domain.MessageEnvelope{
		Type:      domain.MessageTypeNewMessage,
		Timestamp: time.Now().UTC(),
		Data:      message,
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := p.conn.Publish(subject, payload); err != nil {
		return fmt.Errorf("failed to publish message to subject %s: %w", subject, err)
	}

	p.logger.Debug("Message published to NATS",
		"subject", subject,
		"sender", message.SenderID,
		"receiver", message.ReceiverID,
	)

	return nil
}

// PublishStatusUpdate implements ports.MessagePublisher
func (p *NATSMessagePublisher) PublishStatusUpdate(ctx context.Context, userID string, statusUpdate ports.StatusUpdate) error {
	subject := domain.GetStatusTopic(userID)

	// Create status update envelope
	envelope := domain.StatusUpdateEnvelope{
		Type:      domain.MessageTypeStatusUpdate,
		Timestamp: time.Now().UTC(),
		Data:      statusUpdate,
	}

	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal status update: %w", err)
	}

	if err := p.conn.Publish(subject, payload); err != nil {
		return fmt.Errorf("failed to publish status update to subject %s: %w", subject, err)
	}

	p.logger.Debug("Status update published to NATS",
		"subject", subject,
		"user", userID,
		"status", statusUpdate.Status,
	)

	return nil
}

// Close implements ports.MessagePublisher
func (p *NATSMessagePublisher) Close() error {
	if p.conn != nil {
		p.conn.Close()
		p.logger.Info("NATS connection closed")
	}
	return nil
}

