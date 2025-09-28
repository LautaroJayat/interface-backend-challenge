package http

import (
	"time"

	"messaging-app/internal/domain"
)

// Request models
type SendMessageRequest struct {
	Content string `json:"content" validate:"required,max=10000"`
}

type UpdateStatusRequest struct {
	MessageID domain.MessageID `json:"message_id" validate:"required"`
}

type GetMessagesRequest struct {
	Cursor string `json:"cursor"` // RFC3339 timestamp
	Limit  int    `json:"limit"`  // Max 100, default 50
}

// Response models
type SendMessageResponse struct {
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	CreatedAt  time.Time `json:"created_at"`
	Content    string    `json:"content"`
	Status     string    `json:"status"`
}

type GetChatsResponse struct {
	Chats []domain.ChatSession `json:"chats"`
}

type GetMessagesResponse struct {
	Messages   []domain.Message `json:"messages"`
	NextCursor string           `json:"next_cursor,omitempty"`
	HasMore    bool             `json:"has_more"`
}

type UpdateStatusResponse struct {
	UpdatedCount int64 `json:"updated_count"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// Success response wrapper
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}
