package mocks_test

import (
	"context"
	"testing"
	"time"

	"messaging-app/internal/domain"
	"messaging-app/internal/mocks"
	"messaging-app/internal/ports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Example test demonstrating how to use the generated mocks
func TestMockUsageExample(t *testing.T) {
	// Create mock instances
	mockRepo := &mocks.MessageRepository{}
	mockPublisher := &mocks.MessagePublisher{}
	mockLogger := &mocks.Logger{}

	// Example: Mock logger expectations
	mockLogger.On("Debug", "Message sent successfully", "sender", "user1", "receiver", "user2").Return()

	// Example: Mock repository expectations
	testMessage := domain.Message{
		SenderID:   "user1",
		ReceiverID: "user2",
		Content:    "Hello, World!",
		CreatedAt:  time.Now().UTC(),
		Status:     "sent",
	}

	mockRepo.On("SaveMessage", mock.Anything, testMessage).Return(nil)

	// Example: Mock publisher expectations
	mockPublisher.On("PublishMessage", mock.Anything, testMessage).Return(nil)

	// Example usage in a hypothetical handler
	ctx := context.Background()

	// Call the mocked methods
	err := mockRepo.SaveMessage(ctx, testMessage)
	assert.NoError(t, err)

	err = mockPublisher.PublishMessage(ctx, testMessage)
	assert.NoError(t, err)

	mockLogger.Debug("Message sent successfully", "sender", "user1", "receiver", "user2")

	// Verify all expectations were met
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

// Example showing how to mock error scenarios
func TestMockErrorScenarios(t *testing.T) {
	mockRepo := &mocks.MessageRepository{}

	// Mock a repository error
	mockRepo.On("SaveMessage", mock.Anything, mock.Anything).Return(domain.ErrDuplicateMessage)

	ctx := context.Background()
	testMessage := domain.Message{
		SenderID:   "user1",
		ReceiverID: "user2",
		Content:    "Test message",
		CreatedAt:  time.Now().UTC(),
		Status:     "sent",
	}

	err := mockRepo.SaveMessage(ctx, testMessage)
	assert.Equal(t, domain.ErrDuplicateMessage, err)

	mockRepo.AssertExpectations(t)
}

// Example showing how to mock return values
func TestMockReturnValues(t *testing.T) {
	mockRepo := &mocks.MessageRepository{}

	// Mock GetChatSessions to return test data
	expectedSessions := []domain.ChatSession{
		{
			ChatID:          "user1_user2",
			LastMessageAt:   time.Now().UTC(),
			UnreadCount:     5,
			OtherParticipant: "user2",
		},
	}

	mockRepo.On("GetChatSessions", mock.Anything, "user1").Return(expectedSessions, nil)

	ctx := context.Background()
	sessions, err := mockRepo.GetChatSessions(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, expectedSessions, sessions)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "user1_user2", sessions[0].ChatID)

	mockRepo.AssertExpectations(t)
}

// Example showing how to test status updates
func TestMockStatusUpdate(t *testing.T) {
	mockPublisher := &mocks.MessagePublisher{}

	statusUpdate := ports.StatusUpdate{
		MessageID: domain.MessageID{
			SenderID:   "user1",
			ReceiverID: "user2",
			CreatedAt:  time.Now().UTC(),
		},
		Status:    "read",
		UpdatedBy: "user2",
		UpdatedAt: time.Now().UTC(),
	}

	mockPublisher.On("PublishStatusUpdate", mock.Anything, "user2", statusUpdate).Return(nil)

	ctx := context.Background()
	err := mockPublisher.PublishStatusUpdate(ctx, "user2", statusUpdate)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}