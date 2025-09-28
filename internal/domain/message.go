package domain

import (
	"fmt"
	"strings"
	"time"
)

type Message struct {
	SenderID   string    `json:"sender_id" validate:"required,max=100"`
	ReceiverID string    `json:"receiver_id" validate:"required,max=100"`
	CreatedAt  time.Time `json:"created_at" validate:"required"`
	Content    string    `json:"content" validate:"required,max=10000"`
	Status     string    `json:"status" validate:"required,oneof=sent delivered read"`
}

// MessageID represents the composite primary key
type MessageID struct {
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// Validate performs domain-level validation
func (m *Message) Validate() error {
	if strings.TrimSpace(m.SenderID) == "" {
		return ErrInvalidSenderID
	}
	if strings.TrimSpace(m.ReceiverID) == "" {
		return ErrInvalidReceiverID
	}
	if m.SenderID == m.ReceiverID {
		return ErrSelfMessage
	}
	if strings.TrimSpace(m.Content) == "" {
		return ErrEmptyContent
	}
	if len(m.Content) > 10000 {
		return ErrContentTooLong
	}
	if !IsValidStatus(m.Status) {
		return ErrInvalidStatus
	}
	return nil
}

const (
	MessageStatusSent      = "sent"
	MessageStatusDelivered = "delivered"
	MessageStatusRead      = "read"
)

func IsValidStatus(status string) bool {
	validStatuses := []string{MessageStatusSent, MessageStatusDelivered, MessageStatusRead}
	for _, v := range validStatuses {
		if v == status {
			return true
		}
	}
	return false
}

// ComputeChatID creates a consistent chat identifier using a separator that won't conflict with user IDs
func ComputeChatID(user1, user2 string) string {
	if user1 < user2 {
		return fmt.Sprintf("%s---%s", user1, user2)
	}
	return fmt.Sprintf("%s---%s", user2, user1)
}

// GetOtherParticipant returns the other participant in a 1:1 chat
func (m *Message) GetOtherParticipant(currentUserID string) string {
	if m.SenderID == currentUserID {
		return m.ReceiverID
	}
	return m.SenderID
}
