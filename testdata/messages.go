package testdata

import (
	"time"

	"messaging-app/internal/domain"
)

// BaseTime provides a consistent reference point for all test timestamps
var BaseTime = time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

// ValidMessages returns a collection of valid messages for testing
func ValidMessages() []domain.Message {
	return []domain.Message{
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime,
			Content:    "Hey Bob! How's the new feature coming along?",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Bob.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  BaseTime.Add(5 * time.Minute),
			Content:    "Going well! Just finishing up the last few tests.",
			Status:     domain.MessageStatusDelivered,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime.Add(10 * time.Minute),
			Content:    "Awesome! Let me know when it's ready for review.",
			Status:     domain.MessageStatusRead,
		},
		{
			SenderID:   Charlie.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  BaseTime.Add(15 * time.Minute),
			Content:    "Alice, can you take a look at the new design mockups?",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Charlie.UserID,
			CreatedAt:  BaseTime.Add(20 * time.Minute),
			Content:    "Sure! They look great. Just one small suggestion on the color scheme.",
			Status:     domain.MessageStatusDelivered,
		},
	}
}

// InvalidMessages returns messages that should fail validation
func InvalidMessages() []domain.Message {
	return []domain.Message{
		{
			SenderID:   "",
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime,
			Content:    "Empty sender",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: "",
			CreatedAt:  BaseTime,
			Content:    "Empty receiver",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  BaseTime,
			Content:    "Self message",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime,
			Content:    "",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime,
			Content:    "Valid content",
			Status:     "invalid_status",
		},
	}
}

// LargeContentMessages returns messages with content at size boundaries
func LargeContentMessages() []domain.Message {
	return []domain.Message{
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime,
			Content:    generateContent(9999), // Just under limit
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime.Add(1 * time.Minute),
			Content:    generateContent(10000), // At limit
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime.Add(2 * time.Minute),
			Content:    generateContent(10001), // Over limit - should fail
			Status:     domain.MessageStatusSent,
		},
	}
}

// MessageStatusProgression returns messages showing status transitions
func MessageStatusProgression() []domain.Message {
	baseMessage := domain.Message{
		SenderID:   Alice.UserID,
		ReceiverID: Bob.UserID,
		CreatedAt:  BaseTime,
		Content:    "Status progression test message",
	}

	return []domain.Message{
		{
			SenderID:   baseMessage.SenderID,
			ReceiverID: baseMessage.ReceiverID,
			CreatedAt:  baseMessage.CreatedAt,
			Content:    baseMessage.Content,
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   baseMessage.SenderID,
			ReceiverID: baseMessage.ReceiverID,
			CreatedAt:  baseMessage.CreatedAt,
			Content:    baseMessage.Content,
			Status:     domain.MessageStatusDelivered,
		},
		{
			SenderID:   baseMessage.SenderID,
			ReceiverID: baseMessage.ReceiverID,
			CreatedAt:  baseMessage.CreatedAt,
			Content:    baseMessage.Content,
			Status:     domain.MessageStatusRead,
		},
	}
}

// generateContent creates content of specified length for testing
func generateContent(length int) string {
	if length <= 0 {
		return ""
	}

	content := make([]byte, length)
	for i := range content {
		content[i] = 'a'
	}
	return string(content)
}
