package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpAdapter "messaging-app/internal/adapters/http"
	"messaging-app/internal/domain"
	"messaging-app/internal/mocks"
	"messaging-app/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ChatHandlerTestSuite struct {
	suite.Suite
	handler    *ChatHandler
	mockRepo   *mocks.MessageRepository
	mockLogger *mocks.Logger
}

func (s *ChatHandlerTestSuite) SetupTest() {
	s.mockRepo = &mocks.MessageRepository{}
	s.mockLogger = &mocks.Logger{}
	s.handler = NewChatHandler(s.mockRepo, s.mockLogger)
}

func (s *ChatHandlerTestSuite) TearDownTest() {
	s.mockRepo.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
}

// Helper function to create request with user context
func (s *ChatHandlerTestSuite) createRequestWithUser(method, url string, user domain.UserContext) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Set("Content-Type", "application/json")

	// Add user context
	ctx := context.WithValue(req.Context(), httpAdapter.UserContextKey, user)
	return req.WithContext(ctx)
}

// Helper function to create request without user context
func (s *ChatHandlerTestSuite) createRequestWithoutUser(method, url string) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// GetChats Tests

func (s *ChatHandlerTestSuite) TestGetChats_Success() {
	alice := testdata.Alice

	// Create test chat sessions
	expectedSessions := []domain.ChatSession{
		{
			ChatID:           "alice_bob",
			LastMessageAt:    testdata.BaseTime,
			UnreadCount:      3,
			OtherParticipant: "bob",
		},
		{
			ChatID:           "alice_charlie",
			LastMessageAt:    testdata.BaseTime.Add(-1 * time.Hour),
			UnreadCount:      0,
			OtherParticipant: "charlie",
		},
		{
			ChatID:           "alice_diana",
			LastMessageAt:    testdata.BaseTime.Add(-2 * time.Hour),
			UnreadCount:      1,
			OtherParticipant: "diana",
		},
	}

	// Mock expectations
	s.mockRepo.On("GetChatSessions", mock.Anything, alice.UserID).Return(expectedSessions, nil)
	s.mockLogger.On("Debug", "Chat sessions retrieved successfully", "user", alice.UserID, "count", len(expectedSessions)).Return()

	// Create request
	req := s.createRequestWithUser("GET", "/api/v1/chats", alice)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetChats(recorder, req)

	// Assertions
	s.Equal(http.StatusOK, recorder.Code)

	var response GetChatsResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	s.NoError(err)

	s.Equal(len(expectedSessions), len(response.Chats))
	s.Equal(expectedSessions[0].ChatID, response.Chats[0].ChatID)
	s.Equal(expectedSessions[0].OtherParticipant, response.Chats[0].OtherParticipant)
	s.Equal(expectedSessions[0].UnreadCount, response.Chats[0].UnreadCount)
}

func (s *ChatHandlerTestSuite) TestGetChats_EmptyResult() {
	alice := testdata.Alice

	// Empty chat sessions
	expectedSessions := []domain.ChatSession{}

	// Mock expectations
	s.mockRepo.On("GetChatSessions", mock.Anything, alice.UserID).Return(expectedSessions, nil)
	s.mockLogger.On("Debug", "Chat sessions retrieved successfully", "user", alice.UserID, "count", 0).Return()

	// Create request
	req := s.createRequestWithUser("GET", "/api/v1/chats", alice)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetChats(recorder, req)

	// Assertions
	s.Equal(http.StatusOK, recorder.Code)

	var response GetChatsResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	s.NoError(err)

	s.Equal(0, len(response.Chats))
	s.NotNil(response.Chats) // Should be empty array, not null
}

func (s *ChatHandlerTestSuite) TestGetChats_NoUserContext() {
	// Create request without user context
	req := s.createRequestWithoutUser("GET", "/api/v1/chats")
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetChats(recorder, req)

	// Assertions
	s.Equal(http.StatusUnauthorized, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("User context not found", errorResp.Error)
	s.Equal("NO_USER_CONTEXT", errorResp.Code)
}

func (s *ChatHandlerTestSuite) TestGetChats_RepositoryError() {
	alice := testdata.Alice

	// Mock repository error
	repoError := assert.AnError
	s.mockRepo.On("GetChatSessions", mock.Anything, alice.UserID).Return(nil, repoError)
	s.mockLogger.On("Error", "Failed to get chat sessions", "error", repoError, "user", alice.UserID).Return()

	// Create request
	req := s.createRequestWithUser("GET", "/api/v1/chats", alice)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetChats(recorder, req)

	// Assertions
	s.Equal(http.StatusInternalServerError, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Failed to get chats", errorResp.Error)
	s.Equal("GET_CHATS_ERROR", errorResp.Code)
}

func (s *ChatHandlerTestSuite) TestGetChats_LargeNumberOfChats() {
	alice := testdata.Alice

	// Create a large number of chat sessions to test performance
	var expectedSessions []domain.ChatSession
	for i := 0; i < 100; i++ {
		expectedSessions = append(expectedSessions, domain.ChatSession{
			ChatID:           domain.ComputeChatID(alice.UserID, testdata.Bob.UserID+string(rune(i))),
			LastMessageAt:    testdata.BaseTime.Add(-time.Duration(i) * time.Hour),
			UnreadCount:      i % 5, // Vary unread counts
			OtherParticipant: testdata.Bob.UserID + string(rune(i)),
		})
	}

	// Mock expectations
	s.mockRepo.On("GetChatSessions", mock.Anything, alice.UserID).Return(expectedSessions, nil)
	s.mockLogger.On("Debug", "Chat sessions retrieved successfully", "user", alice.UserID, "count", 100).Return()

	// Create request
	req := s.createRequestWithUser("GET", "/api/v1/chats", alice)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetChats(recorder, req)

	// Assertions
	s.Equal(http.StatusOK, recorder.Code)

	var response GetChatsResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	s.NoError(err)

	s.Equal(100, len(response.Chats))
	s.Equal(expectedSessions[0].ChatID, response.Chats[0].ChatID)
}

func (s *ChatHandlerTestSuite) TestGetChats_UserWithSpecialCharacters() {
	// Test with user that has special characters in ID
	specialUser := testdata.UnicodeUser

	expectedSessions := []domain.ChatSession{
		{
			ChatID:           domain.ComputeChatID(specialUser.UserID, testdata.Alice.UserID),
			LastMessageAt:    testdata.BaseTime,
			UnreadCount:      1,
			OtherParticipant: testdata.Alice.UserID,
		},
	}

	// Mock expectations
	s.mockRepo.On("GetChatSessions", mock.Anything, specialUser.UserID).Return(expectedSessions, nil)
	s.mockLogger.On("Debug", "Chat sessions retrieved successfully", "user", specialUser.UserID, "count", 1).Return()

	// Create request
	req := s.createRequestWithUser("GET", "/api/v1/chats", specialUser)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetChats(recorder, req)

	// Assertions
	s.Equal(http.StatusOK, recorder.Code)

	var response GetChatsResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	s.NoError(err)

	s.Equal(1, len(response.Chats))
	s.Equal(expectedSessions[0].ChatID, response.Chats[0].ChatID)
}

func (s *ChatHandlerTestSuite) TestGetChats_SortedByLastMessage() {
	alice := testdata.Alice

	// Create chat sessions with different last message times (should be ordered by most recent first)
	expectedSessions := []domain.ChatSession{
		{
			ChatID:           "alice_charlie",
			LastMessageAt:    testdata.BaseTime, // Most recent
			UnreadCount:      2,
			OtherParticipant: "charlie",
		},
		{
			ChatID:           "alice_bob",
			LastMessageAt:    testdata.BaseTime.Add(-1 * time.Hour), // 1 hour ago
			UnreadCount:      0,
			OtherParticipant: "bob",
		},
		{
			ChatID:           "alice_diana",
			LastMessageAt:    testdata.BaseTime.Add(-24 * time.Hour), // 1 day ago
			UnreadCount:      5,
			OtherParticipant: "diana",
		},
	}

	// Mock expectations
	s.mockRepo.On("GetChatSessions", mock.Anything, alice.UserID).Return(expectedSessions, nil)
	s.mockLogger.On("Debug", "Chat sessions retrieved successfully", "user", alice.UserID, "count", 3).Return()

	// Create request
	req := s.createRequestWithUser("GET", "/api/v1/chats", alice)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetChats(recorder, req)

	// Assertions
	s.Equal(http.StatusOK, recorder.Code)

	var response GetChatsResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	s.NoError(err)

	s.Equal(3, len(response.Chats))

	// Verify order (most recent first)
	s.Equal("alice_charlie", response.Chats[0].ChatID)
	s.Equal("alice_bob", response.Chats[1].ChatID)
	s.Equal("alice_diana", response.Chats[2].ChatID)

	// Verify that each chat has the expected structure
	for i, chat := range response.Chats {
		s.NotEmpty(chat.ChatID)
		s.NotEmpty(chat.OtherParticipant)
		s.NotZero(chat.LastMessageAt)
		s.GreaterOrEqual(chat.UnreadCount, 0)

		// Verify time ordering
		if i > 0 {
			s.True(chat.LastMessageAt.Before(response.Chats[i-1].LastMessageAt) || chat.LastMessageAt.Equal(response.Chats[i-1].LastMessageAt))
		}
	}
}

func (s *ChatHandlerTestSuite) TestGetChats_ResponseFormat() {
	alice := testdata.Alice

	expectedSessions := []domain.ChatSession{
		{
			ChatID:           "alice_bob",
			LastMessageAt:    testdata.BaseTime,
			UnreadCount:      3,
			OtherParticipant: "bob",
		},
	}

	// Mock expectations
	s.mockRepo.On("GetChatSessions", mock.Anything, alice.UserID).Return(expectedSessions, nil)
	s.mockLogger.On("Debug", "Chat sessions retrieved successfully", "user", alice.UserID, "count", 1).Return()

	// Create request
	req := s.createRequestWithUser("GET", "/api/v1/chats", alice)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetChats(recorder, req)

	// Assertions
	s.Equal(http.StatusOK, recorder.Code)

	// Verify JSON structure
	var responseMap map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &responseMap)
	s.NoError(err)

	// Should have "chats" field
	chats, exists := responseMap["chats"]
	s.True(exists)
	s.NotNil(chats)

	// Chats should be an array
	chatsArray, ok := chats.([]interface{})
	s.True(ok)
	s.Equal(1, len(chatsArray))

	// Each chat should have required fields
	chat := chatsArray[0].(map[string]interface{})
	s.Contains(chat, "chat_id")
	s.Contains(chat, "last_message_at")
	s.Contains(chat, "unread_count")
	s.Contains(chat, "other_participant")
}

func TestChatHandlerSuite(t *testing.T) {
	suite.Run(t, new(ChatHandlerTestSuite))
}