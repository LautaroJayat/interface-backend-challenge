package domain

import "time"

type SeedData struct {
	Users    []UserContext
	Messages []Message
}

func GetTestSeedData() SeedData {
	now := time.Now()
	return SeedData{
		Users: []UserContext{
			{UserID: "alice", Email: "alice@test.com", Handler: "alice_h"},
			{UserID: "bob", Email: "bob@test.com", Handler: "bob_h"},
			{UserID: "charlie", Email: "charlie@test.com", Handler: "charlie_h"},
		},
		Messages: []Message{
			{
				SenderID:   "alice",
				ReceiverID: "bob",
				CreatedAt:  now.Add(-2 * time.Hour),
				Content:    "Hello Bob! How are you doing?",
				Status:     "read",
			},
			{
				SenderID:   "bob",
				ReceiverID: "alice",
				CreatedAt:  now.Add(-1 * time.Hour),
				Content:    "Hi Alice! I'm doing great, thanks for asking!",
				Status:     "delivered",
			},
			{
				SenderID:   "alice",
				ReceiverID: "bob",
				CreatedAt:  now.Add(-30 * time.Minute),
				Content:    "That's wonderful to hear!",
				Status:     "sent",
			},
			{
				SenderID:   "charlie",
				ReceiverID: "alice",
				CreatedAt:  now.Add(-15 * time.Minute),
				Content:    "Hey Alice, are you free for lunch today?",
				Status:     "sent",
			},
		},
	}
}

func GetMinimalSeedData() SeedData {
	now := time.Now()
	return SeedData{
		Users: []UserContext{
			{UserID: "test1", Email: "test1@example.com", Handler: "test1"},
			{UserID: "test2", Email: "test2@example.com", Handler: "test2"},
		},
		Messages: []Message{
			{
				SenderID:   "test1",
				ReceiverID: "test2",
				CreatedAt:  now.Add(-1 * time.Minute),
				Content:    "Test message",
				Status:     "sent",
			},
		},
	}
}