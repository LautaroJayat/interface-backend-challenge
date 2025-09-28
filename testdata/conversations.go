package testdata

import (
	"time"

	"messaging-app/internal/domain"
)

// AliceBob.UserIDConversation returns a realistic conversation between Alice and Bob.UserID
func AliceBobUserIDConversation() []domain.Message {
	return []domain.Message{
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime,
			Content:    "Hey Bob.UserID! How's the new API endpoint coming along?",
			Status:     domain.MessageStatusRead,
		},
		{
			SenderID:   Bob.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  BaseTime.Add(3 * time.Minute),
			Content:    "Going really well! Just finished the authentication middleware.",
			Status:     domain.MessageStatusRead,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime.Add(5 * time.Minute),
			Content:    "That's awesome! Any issues with the JWT validation?",
			Status:     domain.MessageStatusRead,
		},
		{
			SenderID:   Bob.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  BaseTime.Add(8 * time.Minute),
			Content:    "Nope, using the standard library made it pretty straightforward. Thanks for the suggestion!",
			Status:     domain.MessageStatusDelivered,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime.Add(10 * time.Minute),
			Content:    "Perfect! When do you think it'll be ready for code review?",
			Status:     domain.MessageStatusDelivered,
		},
		{
			SenderID:   Bob.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  BaseTime.Add(12 * time.Minute),
			Content:    "Should be ready by end of day. Just need to add a few more test cases.",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime.Add(14 * time.Minute),
			Content:    "Sounds good! I'll keep an eye out for the PR.",
			Status:     domain.MessageStatusSent,
		},
	}
}

// AliceCharlieConversation returns a conversation between Alice and Charlie about design
func AliceCharlieConversation() []domain.Message {
	return []domain.Message{
		{
			SenderID:   Charlie.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  BaseTime.Add(30 * time.Minute),
			Content:    "Alice, can you review the new dashboard mockups?",
			Status:     domain.MessageStatusRead,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Charlie.UserID,
			CreatedAt:  BaseTime.Add(35 * time.Minute),
			Content:    "Sure! They look fantastic. Love the color scheme.",
			Status:     domain.MessageStatusRead,
		},
		{
			SenderID:   Charlie.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  BaseTime.Add(38 * time.Minute),
			Content:    "Thanks! What do you think about the navigation layout?",
			Status:     domain.MessageStatusDelivered,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Charlie.UserID,
			CreatedAt:  BaseTime.Add(42 * time.Minute),
			Content:    "It's intuitive, but maybe we could make the search more prominent?",
			Status:     domain.MessageStatusSent,
		},
	}
}

// Bob.UserIDCharlieConversation returns a conversation between Bob.UserID and Charlie
func BobUserIDCharlieConversation() []domain.Message {
	return []domain.Message{
		{
			SenderID:   Bob.UserID,
			ReceiverID: Charlie.UserID,
			CreatedAt:  BaseTime.Add(1 * time.Hour),
			Content:    "Charlie, the API is ready for the frontend integration!",
			Status:     domain.MessageStatusDelivered,
		},
		{
			SenderID:   Charlie.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime.Add(1*time.Hour + 5*time.Minute),
			Content:    "Excellent! I'll start connecting it to the new dashboard.",
			Status:     domain.MessageStatusSent,
		},
	}
}

// MultiUserScenario returns messages involving multiple users for complex testing
func MultiUserScenario() []domain.Message {
	messages := []domain.Message{}

	// Add all individual conversations
	messages = append(messages, AliceBobUserIDConversation()...)
	messages = append(messages, AliceCharlieConversation()...)
	messages = append(messages, BobUserIDCharlieConversation()...)

	// Add additional cross-conversations
	messages = append(messages, []domain.Message{
		{
			SenderID:   Diana.UserID,
			ReceiverID: Alice.UserID,
			CreatedAt:  BaseTime.Add(2 * time.Hour),
			Content:    "Alice, quick question about the database schema changes.",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Eve.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime.Add(2*time.Hour + 15*time.Minute),
			Content:    "Found a small bug in the authentication flow. Can you take a look?",
			Status:     domain.MessageStatusSent,
		},
		{
			SenderID:   Alice.UserID,
			ReceiverID: Eve.UserID,
			CreatedAt:  BaseTime.Add(2*time.Hour + 30*time.Minute),
			Content:    "I can help with that if Bob.UserID is busy!",
			Status:     domain.MessageStatusSent,
		},
	}...)

	return messages
}

// EmptyConversation returns an empty slice for testing edge cases
func EmptyConversation() []domain.Message {
	return []domain.Message{}
}

// SingleMessageConversation returns a conversation with just one message
func SingleMessageConversation() []domain.Message {
	return []domain.Message{
		{
			SenderID:   Alice.UserID,
			ReceiverID: Bob.UserID,
			CreatedAt:  BaseTime,
			Content:    "Just a quick message to test single message scenarios.",
			Status:     domain.MessageStatusSent,
		},
	}
}

// GetConversationBetween returns all messages between two specific users
func GetConversationBetween(userID1, userID2 string) []domain.Message {
	allMessages := MultiUserScenario()
	var conversation []domain.Message

	for _, msg := range allMessages {
		if (msg.SenderID == userID1 && msg.ReceiverID == userID2) ||
			(msg.SenderID == userID2 && msg.ReceiverID == userID1) {
			conversation = append(conversation, msg)
		}
	}

	return conversation
}
