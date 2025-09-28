package domain

import (
	"time"
)

type ChatSession struct {
	ChatID           string    `json:"chat_id"`
	OtherParticipant string    `json:"other_participant"`
	LastMessageAt    time.Time `json:"last_message_at"`
	UnreadCount      int       `json:"unread_count"`
	LastMessage      string    `json:"last_message"`
	LastMessageBy    string    `json:"last_message_by"`
}

// IsUnread checks if the session has unread messages
func (cs *ChatSession) IsUnread() bool {
	return cs.UnreadCount > 0
}

// ComputeChatIDFromParticipants creates chat ID from two participants
func ComputeChatIDFromParticipants(participant1, participant2 string) string {
	return ComputeChatID(participant1, participant2)
}