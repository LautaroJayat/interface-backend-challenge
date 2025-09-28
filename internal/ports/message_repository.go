package ports

import (
	"context"
	"time"

	"messaging-app/internal/domain"
)

//go:generate mockery --name=MessageRepository --output=../mocks --outpkg=mocks

type MessageRepository interface {
	// SaveMessage stores a new message with idempotency protection
	// Returns ErrDuplicateMessage if message with same composite key exists
	SaveMessage(ctx context.Context, message domain.Message) error

	// GetMessages retrieves messages for a chat with cursor-based pagination
	// cursor: timestamp to start from (exclusive), use time.Time{} for first page
	// limit: maximum number of messages to return (1-100)
	// Returns messages in descending order by created_at (newest first)
	GetMessages(ctx context.Context, chatID string, cursor time.Time, limit int) ([]domain.Message, error)

	// GetChatSessions retrieves all chat sessions for a user
	// Returns sessions ordered by last_message_at descending
	GetChatSessions(ctx context.Context, userID string) ([]domain.ChatSession, error)

	// MarkMessagesUpToRead updates status for multiple messages to read
	// It expects a messageId and will mark all previous messages of the same conversation as read
	MarkMessagesUpToRead(ctx context.Context, msg domain.MessageID) (int64, error)

	// GetMessageByID retrieves a specific message by its composite key
	GetMessageByID(ctx context.Context, messageID domain.MessageID) (*domain.Message, error)

	// GetUnreadCount returns count of unread messages for a user in a specific chat
	GetUnreadCount(ctx context.Context, userID, chatID string) (int, error)

	// MarkChatAsRead marks all messages in a chat as read for the receiver
	MarkChatAsRead(ctx context.Context, userID, chatID string) error
}

// PaginationResult wraps paginated results
type PaginationResult struct {
	Messages   []domain.Message `json:"messages"`
	NextCursor *time.Time       `json:"next_cursor,omitempty"`
	HasMore    bool             `json:"has_more"`
	Total      int              `json:"total,omitempty"`
}
