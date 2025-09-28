package ports

import (
	"context"
	"time"

	"messaging-app/internal/domain"
)

//go:generate mockery --name=MessagePublisher --output=../mocks --outpkg=mocks

type MessagePublisher interface {
	// PublishMessage sends a message to the real-time delivery system
	// Subject pattern: messages.{receiver_id}
	PublishMessage(ctx context.Context, message domain.Message) error

	// PublishStatusUpdate notifies about message status changes
	// Subject pattern: status.{user_id}
	PublishStatusUpdate(ctx context.Context, userID string, statusUpdate StatusUpdate) error

	// Close gracefully shuts down the publisher
	Close() error
}

type StatusUpdate struct {
	MessageID domain.MessageID `json:"message_id"`
	Status    string           `json:"status"`
	UpdatedBy string           `json:"updated_by"`
	UpdatedAt time.Time        `json:"updated_at"`
}
