// Package testdata provides comprehensive test data using our domain models.
// This serves as the single source of truth for all testing scenarios.
package testdata

import "messaging-app/internal/domain"

// TestDataset represents a complete dataset for testing
type TestDataset struct {
	Users        []domain.UserContext
	Messages     []domain.Message
	ChatSessions []domain.ChatSession
}

// FullTestDataset returns a complete dataset with all test scenarios
func FullTestDataset() TestDataset {
	return TestDataset{
		Users:        ValidUsers(),
		Messages:     MultiUserScenario(),
		ChatSessions: getAllChatSessions(),
	}
}

// MinimalTestDataset returns a minimal dataset for simple testing
func MinimalTestDataset() TestDataset {
	return TestDataset{
		Users:        ValidUsers()[:2], // Just Alice and Bob
		Messages:     SingleMessageConversation(),
		ChatSessions: SingleChatSession(),
	}
}

// ValidationTestDataset returns data specifically for validation testing
func ValidationTestDataset() TestDataset {
	validUsers := ValidUsers()[:2] // Get Alice and Bob
	invalidUsers := InvalidUsers()
	allUsers := append(validUsers, invalidUsers...)

	validMessages := ValidMessages()[:2] // Get first two valid messages
	invalidMessages := InvalidMessages()
	allMessages := append(validMessages, invalidMessages...)

	return TestDataset{
		Users:        allUsers,
		Messages:     allMessages,
		ChatSessions: EmptyChatSessions(), // No chat sessions for validation testing
	}
}

// PerformanceTestDataset returns a large dataset for performance testing
func PerformanceTestDataset() TestDataset {
	return TestDataset{
		Users:        ValidUsers(), // All users
		Messages:     PerformanceTestData(),
		ChatSessions: EmptyChatSessions(), // Chat sessions computed from messages
	}
}

// getAllChatSessions returns all predefined chat sessions
func getAllChatSessions() []domain.ChatSession {
	var allSessions []domain.ChatSession
	allSessions = append(allSessions, AliceChatSessions()...)
	allSessions = append(allSessions, BobChatSessions()...)
	allSessions = append(allSessions, CharlieChatSessions()...)
	return allSessions
}

// GetTestDataByScenario returns test data for specific scenarios
func GetTestDataByScenario(scenario string) TestDataset {
	switch scenario {
	case "full":
		return FullTestDataset()
	case "minimal":
		return MinimalTestDataset()
	case "validation":
		return ValidationTestDataset()
	case "performance":
		return PerformanceTestDataset()
	case "pagination":
		return TestDataset{
			Users:        ValidUsers(),
			Messages:     PaginationTestData(),
			ChatSessions: EmptyChatSessions(),
		}
	case "concurrent":
		return TestDataset{
			Users:        ValidUsers(),
			Messages:     ConcurrentMessagingData(),
			ChatSessions: EmptyChatSessions(),
		}
	case "duplicates":
		return TestDataset{
			Users:        ValidUsers()[:2],
			Messages:     DuplicateMessageScenario(),
			ChatSessions: EmptyChatSessions(),
		}
	case "status_updates":
		return TestDataset{
			Users:        ValidUsers()[:2],
			Messages:     StatusUpdateScenario(),
			ChatSessions: EmptyChatSessions(),
		}
	case "timezone":
		return TestDataset{
			Users:        ValidUsers(),
			Messages:     CrossTimeZoneScenario(),
			ChatSessions: EmptyChatSessions(),
		}
	default:
		return MinimalTestDataset()
	}
}