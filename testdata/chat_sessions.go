package testdata

import (
	"time"

	"messaging-app/internal/domain"
)

// AliceChatSessions returns Alice's expected chat sessions based on the test conversations
func AliceChatSessions() []domain.ChatSession {
	return []domain.ChatSession{
		{
			ChatID:           domain.ComputeChatID(Alice.UserID, Bob.UserID),
			OtherParticipant: Bob.UserID,
			LastMessageAt:    BaseTime.Add(14 * time.Minute),
			UnreadCount:      0, // Alice has read all messages
			LastMessage:      "Sounds good! I'll keep an eye out for the PR.",
			LastMessageBy:    Alice.UserID,
		},
		{
			ChatID:           domain.ComputeChatID(Alice.UserID, Charlie.UserID),
			OtherParticipant: Charlie.UserID,
			LastMessageAt:    BaseTime.Add(42 * time.Minute),
			UnreadCount:      0, // Alice sent the last message
			LastMessage:      "It's intuitive, but maybe we could make the search more prominent?",
			LastMessageBy:    Alice.UserID,
		},
		{
			ChatID:           domain.ComputeChatID(Alice.UserID, Diana.UserID),
			OtherParticipant: Diana.UserID,
			LastMessageAt:    BaseTime.Add(2 * time.Hour),
			UnreadCount:      1, // Diana's message is unread
			LastMessage:      "Alice, quick question about the database schema changes.",
			LastMessageBy:    Diana.UserID,
		},
		{
			ChatID:           domain.ComputeChatID(Alice.UserID, "eve"),
			OtherParticipant: "eve",
			LastMessageAt:    BaseTime.Add(2*time.Hour + 30*time.Minute),
			UnreadCount:      0, // Alice sent the last message
			LastMessage:      "I can help with that if Bob is busy!",
			LastMessageBy:    Alice.UserID,
		},
	}
}

// BobChatSessions returns Bob's expected chat sessions
func BobChatSessions() []domain.ChatSession {
	return []domain.ChatSession{
		{
			ChatID:           domain.ComputeChatID(Alice.UserID, Bob.UserID),
			OtherParticipant: Alice.UserID,
			LastMessageAt:    BaseTime.Add(14 * time.Minute),
			UnreadCount:      2, // Last two messages from Alice are unread by Bob
			LastMessage:      "Sounds good! I'll keep an eye out for the PR.",
			LastMessageBy:    Alice.UserID,
		},
		{
			ChatID:           domain.ComputeChatID(Bob.UserID, Charlie.UserID),
			OtherParticipant: Charlie.UserID,
			LastMessageAt:    BaseTime.Add(1*time.Hour + 5*time.Minute),
			UnreadCount:      1, // Charlie's response is unread
			LastMessage:      "Excellent! I'll start connecting it to the new dashboard.",
			LastMessageBy:    Charlie.UserID,
		},
		{
			ChatID:           domain.ComputeChatID(Bob.UserID, "eve"),
			OtherParticipant: "eve",
			LastMessageAt:    BaseTime.Add(2*time.Hour + 15*time.Minute),
			UnreadCount:      1, // Eve's bug report is unread
			LastMessage:      "Found a small bug in the authentication flow. Can you take a look?",
			LastMessageBy:    "eve",
		},
	}
}

// CharlieChatSessions returns Charlie's expected chat sessions
func CharlieChatSessions() []domain.ChatSession {
	return []domain.ChatSession{
		{
			ChatID:           domain.ComputeChatID(Alice.UserID, Charlie.UserID),
			OtherParticipant: Alice.UserID,
			LastMessageAt:    BaseTime.Add(42 * time.Minute),
			UnreadCount:      1, // Alice's suggestion about search is unread
			LastMessage:      "It's intuitive, but maybe we could make the search more prominent?",
			LastMessageBy:    Alice.UserID,
		},
		{
			ChatID:           domain.ComputeChatID(Bob.UserID, Charlie.UserID),
			OtherParticipant: Bob.UserID,
			LastMessageAt:    BaseTime.Add(1*time.Hour + 5*time.Minute),
			UnreadCount:      0, // Charlie sent the last message
			LastMessage:      "Excellent! I'll start connecting it to the new dashboard.",
			LastMessageBy:    Charlie.UserID,
		},
	}
}

// EmptyChatSessions returns an empty chat session list for testing
func EmptyChatSessions() []domain.ChatSession {
	return []domain.ChatSession{}
}

// SingleChatSession returns a single chat session for testing
func SingleChatSession() []domain.ChatSession {
	return []domain.ChatSession{
		{
			ChatID:           domain.ComputeChatID(Alice.UserID, Bob.UserID),
			OtherParticipant: Bob.UserID,
			LastMessageAt:    BaseTime,
			UnreadCount:      0,
			LastMessage:      "Just a quick message to test single chat scenarios.",
			LastMessageBy:    Alice.UserID,
		},
	}
}

// GetChatSessionsForUser returns expected chat sessions for a specific user
func GetChatSessionsForUser(userID string) []domain.ChatSession {
	switch userID {
	case Alice.UserID:
		return AliceChatSessions()
	case Bob.UserID:
		return BobChatSessions()
	case Charlie.UserID:
		return CharlieChatSessions()
	default:
		return EmptyChatSessions()
	}
}
