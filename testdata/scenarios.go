package testdata

import (
	"time"

	"messaging-app/internal/domain"
)

// PaginationTestData returns messages designed for testing pagination
func PaginationTestData() []domain.Message {
	messages := make([]domain.Message, 50)

	for i := 0; i < 50; i++ {
		// Alternate between alice->bob and bob->alice
		var senderID, receiverID string
		if i%2 == 0 {
			senderID, receiverID = Alice.UserID, Bob.UserID
		} else {
			senderID, receiverID = Bob.UserID, Alice.UserID
		}

		messages[i] = domain.Message{
			SenderID:   senderID,
			ReceiverID: receiverID,
			CreatedAt:  BaseTime.Add(time.Duration(i) * time.Minute),
			Content:    generatePaginationContent(i),
			Status:     domain.MessageStatusSent,
		}
	}

	return messages
}

// ConcurrentMessagingData returns messages sent at nearly the same time for concurrency testing
func ConcurrentMessagingData() []domain.Message {
	baseTime := BaseTime

	return []domain.Message{
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  baseTime,
			Content:    "Concurrent message 1",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Bob.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  baseTime.Add(1 * time.Millisecond),
			Content:    "Concurrent message 2",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  baseTime.Add(2 * time.Millisecond),
			Content:    "Concurrent message 3",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Bob.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  baseTime.Add(3 * time.Millisecond),
			Content:    "Concurrent message 4",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Charlie.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  baseTime.Add(4 * time.Millisecond),
			Content:    "Concurrent message from third party",
			Status:     domain.MessageStatusSent,
		},
	}
}

// PerformanceTestData returns a large dataset for performance testing
func PerformanceTestData() []domain.Message {
	messages := make([]domain.Message, 1000)
	users := []string{Alice.UserID, Bob.UserID, Charlie.UserID, "diana", "eve"}

	for i := 0; i < 1000; i++ {
		// Rotate through users
		senderIdx := i % len(users)
		receiverIdx := (i + 1) % len(users)

		messages[i] = domain.Message{
			SenderID:   users[senderIdx],
			ReceiverID: users[receiverIdx],
			CreatedAt:  BaseTime.Add(time.Duration(i) * time.Second),
			Content:    generatePerformanceContent(i),
			Status:     getRandomStatus(i),
		}
	}

	return messages
}

// DuplicateMessageScenario returns identical messages for testing duplicate handling
func DuplicateMessageScenario() []domain.Message {
	duplicateMessage := domain.Message{
		SenderID:   Alice.UserID,
		ReceiverID: Bob.UserID,
		CreatedAt:  BaseTime,
		Content:    "This is a duplicate message for testing",
		Status:     domain.MessageStatusSent,
	}

	return []domain.Message{
		duplicateMessage,
		duplicateMessage, // Exact duplicate
		duplicateMessage, // Another duplicate
	}
}

// StatusUpdateScenario returns messages designed for testing status updates
func StatusUpdateScenario() []domain.Message {
	return []domain.Message{
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime,
			Content:    "Message 1 - should become delivered",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime.Add(1 * time.Minute),
			Content:    "Message 2 - should become read",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime.Add(2 * time.Minute),
			Content:    "Message 3 - stays sent",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Bob.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  BaseTime.Add(3 * time.Minute),
			Content:    "Message 4 - response from Bob",
			Status:     domain.MessageStatusSent,
		},
	}
}

// CrossTimeZoneScenario returns messages with various timezone considerations
func CrossTimeZoneScenario() []domain.Message {
	utc := time.UTC
	est, _ := time.LoadLocation("America/New_York")
	pst, _ := time.LoadLocation("America/Los_Angeles")

	return []domain.Message{
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  time.Date(2024, 1, 15, 10, 0, 0, 0, utc),
			Content:    "UTC message",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Bob.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  time.Date(2024, 1, 15, 5, 0, 0, 0, est).UTC(),
			Content:    "EST message (converted to UTC)",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Charlie.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  time.Date(2024, 1, 15, 2, 0, 0, 0, pst).UTC(),
			Content:    "PST message (converted to UTC)",
			Status:     domain.MessageStatusSent,
		},
	}
}

// Helper functions

func generatePaginationContent(index int) string {
	return "Pagination test message number " + string(rune(index+48)) // Simple number conversion
}

func generatePerformanceContent(index int) string {
	return "Performance test message with index " + string(rune(index%10+48)) + " and some additional content to make it realistic"
}

func getRandomStatus(index int) string {
	statuses := []string{domain.MessageStatusSent, domain.MessageStatusDelivered, domain.MessageStatusRead}
	return statuses[index%3]
}
