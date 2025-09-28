package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpAdapter "messaging-app/internal/adapters/http"
	"messaging-app/internal/domain"
	"messaging-app/internal/mocks"
	"messaging-app/internal/ports"
	"messaging-app/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MessageHandlerTestSuite struct {
	suite.Suite
	handler       *MessageHandler
	mockRepo      *mocks.MessageRepository
	mockPublisher *mocks.MessagePublisher
	mockLogger    *mocks.Logger
}

func (s *MessageHandlerTestSuite) SetupTest() {
	s.mockRepo = &mocks.MessageRepository{}
	s.mockPublisher = &mocks.MessagePublisher{}
	s.mockLogger = &mocks.Logger{}
	s.handler = NewMessageHandler(s.mockRepo, s.mockPublisher, s.mockLogger)
}

func (s *MessageHandlerTestSuite) TearDownTest() {
	s.mockRepo.AssertExpectations(s.T())
	s.mockPublisher.AssertExpectations(s.T())
	s.mockLogger.AssertExpectations(s.T())
}

// Helper function to create request with user context
func (s *MessageHandlerTestSuite) createRequestWithUser(method, url string, body interface{}, user domain.UserContext) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Add user context
	ctx := context.WithValue(req.Context(), httpAdapter.UserContextKey, user)
	return req.WithContext(ctx)
}

// Helper function to create request without user context
func (s *MessageHandlerTestSuite) createRequestWithoutUser(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// SendMessage Tests

func (s *MessageHandlerTestSuite) TestSendMessage_Success() {
	alice := testdata.Alice
	bob := testdata.Bob

	requestBody := SendMessageRequest{
		Content: "Hello Bob!",
	}

	expectedMessage := domain.Message{
		SenderID:   alice.UserID,
		ReceiverID: bob.UserID,
		Content:    requestBody.Content,
		Status:     "sent",
	}

	// Mock expectations
	s.mockRepo.On("SaveMessage", mock.Anything, mock.MatchedBy(func(msg domain.Message) bool {
		return msg.SenderID == expectedMessage.SenderID &&
			msg.ReceiverID == expectedMessage.ReceiverID &&
			msg.Content == expectedMessage.Content &&
			msg.Status == expectedMessage.Status
	})).Return(nil)

	s.mockPublisher.On("PublishMessage", mock.Anything, mock.MatchedBy(func(msg domain.Message) bool {
		return msg.SenderID == expectedMessage.SenderID &&
			msg.ReceiverID == expectedMessage.ReceiverID &&
			msg.Content == expectedMessage.Content
	})).Return(nil)

	s.mockLogger.On("Debug", "Message sent successfully", "sender", alice.UserID, "receiver", bob.UserID).Return()

	// Create request
	req := s.createRequestWithUser("POST", "/api/v1/chats/"+bob.UserID+"/messages", requestBody, alice)

	// Mock the URL path parsing by setting the correct path
	req.URL.Path = "/api/v1/chats/" + bob.UserID + "/messages"

	recorder := httptest.NewRecorder()

	// Execute
	s.handler.SendMessage(recorder, req)

	// Assertions
	s.Equal(http.StatusCreated, recorder.Code)

	var response SendMessageResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	s.NoError(err)

	s.Equal(alice.UserID, response.SenderID)
	s.Equal(bob.UserID, response.ReceiverID)
	s.Equal(requestBody.Content, response.Content)
	s.Equal("sent", response.Status)
	s.WithinDuration(time.Now(), response.CreatedAt, 5*time.Second)
}

func (s *MessageHandlerTestSuite) TestSendMessage_NoUserContext() {
	requestBody := SendMessageRequest{
		Content: "Hello!",
	}

	req := s.createRequestWithoutUser("POST", "/api/v1/chats/user123/messages", requestBody)
	req.URL.Path = "/api/v1/chats/user123/messages"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.SendMessage(recorder, req)

	// Assertions
	s.Equal(http.StatusUnauthorized, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("User context not found", errorResp.Error)
	s.Equal("NO_USER_CONTEXT", errorResp.Code)
}

func (s *MessageHandlerTestSuite) TestSendMessage_MissingReceiverID() {
	alice := testdata.Alice
	requestBody := SendMessageRequest{
		Content: "Hello!",
	}

	req := s.createRequestWithUser("POST", "/api/v1/chats//messages", requestBody, alice)
	req.URL.Path = "/api/v1/chats//messages"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.SendMessage(recorder, req)

	// Assertions
	s.Equal(http.StatusBadRequest, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Missing receiver ID", errorResp.Error)
	s.Equal("MISSING_RECEIVER_ID", errorResp.Code)
}

func (s *MessageHandlerTestSuite) TestSendMessage_InvalidJSON() {
	alice := testdata.Alice

	req := s.createRequestWithUser("POST", "/api/v1/chats/user123/messages", nil, alice)
	req.URL.Path = "/api/v1/chats/user123/messages"
	req.Body = http.NoBody // Invalid JSON

	recorder := httptest.NewRecorder()

	// Execute
	s.handler.SendMessage(recorder, req)

	// Assertions
	s.Equal(http.StatusBadRequest, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Invalid JSON", errorResp.Error)
	s.Equal("INVALID_JSON", errorResp.Code)
}

func (s *MessageHandlerTestSuite) TestSendMessage_EmptyContent() {
	alice := testdata.Alice
	bob := testdata.Bob

	requestBody := SendMessageRequest{
		Content: "", // Empty content should fail validation
	}

	req := s.createRequestWithUser("POST", "/api/v1/chats/"+bob.UserID+"/messages", requestBody, alice)
	req.URL.Path = "/api/v1/chats/" + bob.UserID + "/messages"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.SendMessage(recorder, req)

	// Assertions
	s.Equal(http.StatusBadRequest, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Validation failed", errorResp.Error)
	s.Equal("VALIDATION_ERROR", errorResp.Code)
}

func (s *MessageHandlerTestSuite) TestSendMessage_SelfMessage() {
	alice := testdata.Alice

	requestBody := SendMessageRequest{
		Content: "Talking to myself",
	}

	req := s.createRequestWithUser("POST", "/api/v1/chats/"+alice.UserID+"/messages", requestBody, alice)
	req.URL.Path = "/api/v1/chats/" + alice.UserID + "/messages"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.SendMessage(recorder, req)

	// Assertions
	s.Equal(http.StatusBadRequest, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Validation failed", errorResp.Error)
	s.Equal("VALIDATION_ERROR", errorResp.Code)
}

func (s *MessageHandlerTestSuite) TestSendMessage_DuplicateMessage() {
	alice := testdata.Alice
	bob := testdata.Bob

	requestBody := SendMessageRequest{
		Content: "Hello Bob!",
	}

	// Mock expectations - repository returns duplicate error
	s.mockRepo.On("SaveMessage", mock.Anything, mock.Anything).Return(domain.ErrDuplicateMessage)

	req := s.createRequestWithUser("POST", "/api/v1/chats/"+bob.UserID+"/messages", requestBody, alice)
	req.URL.Path = "/api/v1/chats/" + bob.UserID + "/messages"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.SendMessage(recorder, req)

	// Assertions
	s.Equal(http.StatusConflict, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Duplicate message", errorResp.Error)
	s.Equal("DUPLICATE_MESSAGE", errorResp.Code)
}

func (s *MessageHandlerTestSuite) TestSendMessage_RepositoryError() {
	alice := testdata.Alice
	bob := testdata.Bob

	requestBody := SendMessageRequest{
		Content: "Hello Bob!",
	}

	// Mock expectations - repository returns generic error
	repoError := assert.AnError
	s.mockRepo.On("SaveMessage", mock.Anything, mock.Anything).Return(repoError)
	s.mockLogger.On("Error", "Failed to save message", "error", repoError, "sender", alice.UserID, "receiver", bob.UserID).Return()

	req := s.createRequestWithUser("POST", "/api/v1/chats/"+bob.UserID+"/messages", requestBody, alice)
	req.URL.Path = "/api/v1/chats/" + bob.UserID + "/messages"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.SendMessage(recorder, req)

	// Assertions
	s.Equal(http.StatusInternalServerError, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Failed to save message", errorResp.Error)
	s.Equal("SAVE_ERROR", errorResp.Code)
}

func (s *MessageHandlerTestSuite) TestSendMessage_PublisherError() {
	// Should still succeed even if publisher fails
	alice := testdata.Alice
	bob := testdata.Bob

	requestBody := SendMessageRequest{
		Content: "Hello Bob!",
	}

	// Mock expectations
	s.mockRepo.On("SaveMessage", mock.Anything, mock.Anything).Return(nil)

	publishError := assert.AnError
	s.mockPublisher.On("PublishMessage", mock.Anything, mock.Anything).Return(publishError)
	s.mockLogger.On("Error", "Failed to publish message", "error", publishError, "sender", alice.UserID, "receiver", bob.UserID).Return()
	s.mockLogger.On("Debug", "Message sent successfully", "sender", alice.UserID, "receiver", bob.UserID).Return()

	req := s.createRequestWithUser("POST", "/api/v1/chats/"+bob.UserID+"/messages", requestBody, alice)
	req.URL.Path = "/api/v1/chats/" + bob.UserID + "/messages"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.SendMessage(recorder, req)

	// Assertions - should still return 201 even if publishing fails
	s.Equal(http.StatusCreated, recorder.Code)
}

// GetMessages Tests

func (s *MessageHandlerTestSuite) TestGetMessages_Success() {
	alice := testdata.Alice
	validMessages := testdata.ValidMessages()
	chatID := "alice_bob"

	// Mock expectations
	s.mockRepo.On("GetMessages", mock.Anything, chatID, mock.AnythingOfType("time.Time"), 50).Return(validMessages, nil)
	s.mockLogger.On("Debug", "Messages retrieved successfully", "chat_id", chatID, "user", alice.UserID, "count", len(validMessages)).Return()

	req := s.createRequestWithUser("GET", "/api/v1/chats/"+chatID+"/messages", nil, alice)
	req.URL.Path = "/api/v1/chats/" + chatID + "/messages"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetMessages(recorder, req)

	// Assertions
	s.Equal(http.StatusOK, recorder.Code)

	var response GetMessagesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	s.NoError(err)

	s.Equal(len(validMessages), len(response.Messages))
	s.Equal(len(validMessages) == 50, response.HasMore)
}

func (s *MessageHandlerTestSuite) TestGetMessages_WithPagination() {
	alice := testdata.Alice
	validMessages := testdata.ValidMessages()
	chatID := "alice_bob"

	// Request with cursor and limit
	req := s.createRequestWithUser("GET", "/api/v1/chats/"+chatID+"/messages?cursor=2024-01-15T10:05:00Z&limit=10", nil, alice)
	req.URL.Path = "/api/v1/chats/" + chatID + "/messages"
	req.URL.RawQuery = "cursor=2024-01-15T10:05:00Z&limit=10"

	// Mock expectations
	s.mockRepo.On("GetMessages", mock.Anything, chatID, mock.AnythingOfType("time.Time"), 10).Return(validMessages[:2], nil)
	s.mockLogger.On("Debug", "Messages retrieved successfully", "chat_id", chatID, "user", alice.UserID, "count", 2).Return()

	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetMessages(recorder, req)

	// Assertions
	s.Equal(http.StatusOK, recorder.Code)

	var response GetMessagesResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	s.NoError(err)

	s.Equal(2, len(response.Messages))
	s.False(response.HasMore) // Less than limit means no more
}

func (s *MessageHandlerTestSuite) TestGetMessages_InvalidCursor() {
	alice := testdata.Alice
	chatID := "alice_bob"

	req := s.createRequestWithUser("GET", "/api/v1/chats/"+chatID+"/messages?cursor=invalid-date", nil, alice)
	req.URL.Path = "/api/v1/chats/" + chatID + "/messages"
	req.URL.RawQuery = "cursor=invalid-date"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetMessages(recorder, req)

	// Assertions
	s.Equal(http.StatusBadRequest, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Invalid cursor format", errorResp.Error)
	s.Equal("INVALID_CURSOR", errorResp.Code)
}

func (s *MessageHandlerTestSuite) TestGetMessages_InvalidLimit() {
	alice := testdata.Alice
	chatID := "alice_bob"

	req := s.createRequestWithUser("GET", "/api/v1/chats/"+chatID+"/messages?limit=500", nil, alice)
	req.URL.Path = "/api/v1/chats/" + chatID + "/messages"
	req.URL.RawQuery = "limit=500"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetMessages(recorder, req)

	// Assertions
	s.Equal(http.StatusBadRequest, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Invalid limit", errorResp.Error)
	s.Equal("INVALID_LIMIT", errorResp.Code)
}

func (s *MessageHandlerTestSuite) TestGetMessages_RepositoryError() {
	alice := testdata.Alice
	chatID := "alice_bob"

	repoError := assert.AnError
	s.mockRepo.On("GetMessages", mock.Anything, chatID, mock.AnythingOfType("time.Time"), 50).Return(nil, repoError)
	s.mockLogger.On("Error", "Failed to get messages", "error", repoError, "chat_id", chatID, "user", alice.UserID).Return()

	req := s.createRequestWithUser("GET", "/api/v1/chats/"+chatID+"/messages", nil, alice)
	req.URL.Path = "/api/v1/chats/" + chatID + "/messages"
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.GetMessages(recorder, req)

	// Assertions
	s.Equal(http.StatusInternalServerError, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Failed to get messages", errorResp.Error)
	s.Equal("GET_MESSAGES_ERROR", errorResp.Code)
}

// UpdateMessageStatus Tests

func (s *MessageHandlerTestSuite) TestUpdateMessageStatus_Success() {
	bob := testdata.Bob
	validMessages := testdata.ValidMessages()
	testMessage := validMessages[0] // This is from Alice to Bob

	messageID := domain.MessageID{
		SenderID:   testMessage.SenderID,
		ReceiverID: testMessage.ReceiverID,
		CreatedAt:  testMessage.CreatedAt,
	}

	requestBody := UpdateStatusRequest{
		MessageID: messageID,
	}

	// Mock expectations - Bob is updating status of message he received
	s.mockRepo.On("MarkMessagesUpToRead", mock.Anything, messageID).Return(int64(3), nil)
	s.mockPublisher.On("PublishStatusUpdate", mock.Anything, bob.UserID, mock.MatchedBy(func(status ports.StatusUpdate) bool {
		return status.MessageID == messageID &&
			status.Status == domain.MessageStatusRead &&
			status.UpdatedBy == bob.UserID
	})).Return(nil)
	s.mockLogger.On("Debug", "Message status updated successfully", "user", bob.UserID, "count", int64(3), "status", domain.MessageStatusRead).Return()

	req := s.createRequestWithUser("PATCH", "/api/v1/messages/status", requestBody, bob)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.UpdateMessageStatus(recorder, req)

	// Assertions
	s.Equal(http.StatusOK, recorder.Code)

	var response UpdateStatusResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	s.NoError(err)

	s.Equal(int64(3), response.UpdatedCount)
}

func (s *MessageHandlerTestSuite) TestUpdateMessageStatus_AccessDenied() {
	alice := testdata.Alice
	validMessages := testdata.ValidMessages()
	testMessage := validMessages[0] // This is from Alice to Bob

	messageID := domain.MessageID{
		SenderID:   testMessage.SenderID,
		ReceiverID: testMessage.ReceiverID,
		CreatedAt:  testMessage.CreatedAt,
	}

	requestBody := UpdateStatusRequest{
		MessageID: messageID,
	}

	// Alice tries to update status of message she sent (not received)
	req := s.createRequestWithUser("PATCH", "/api/v1/messages/status", requestBody, alice)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.UpdateMessageStatus(recorder, req)

	// Assertions
	s.Equal(http.StatusForbidden, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Access denied", errorResp.Error)
	s.Equal("ACCESS_DENIED", errorResp.Code)
	s.Equal("Can only update status of messages you received", errorResp.Details)
}

func (s *MessageHandlerTestSuite) TestUpdateMessageStatus_RepositoryError() {
	bob := testdata.Bob
	validMessages := testdata.ValidMessages()
	testMessage := validMessages[0]

	messageID := domain.MessageID{
		SenderID:   testMessage.SenderID,
		ReceiverID: testMessage.ReceiverID,
		CreatedAt:  testMessage.CreatedAt,
	}

	requestBody := UpdateStatusRequest{
		MessageID: messageID,
	}

	repoError := assert.AnError
	s.mockRepo.On("MarkMessagesUpToRead", mock.Anything, messageID).Return(int64(0), repoError)
	s.mockLogger.On("Error", "Failed to update message status", "error", repoError, "user", bob.UserID, "message_id", messageID).Return()

	req := s.createRequestWithUser("PATCH", "/api/v1/messages/status", requestBody, bob)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.UpdateMessageStatus(recorder, req)

	// Assertions
	s.Equal(http.StatusInternalServerError, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	s.NoError(err)
	s.Equal("Failed to update status", errorResp.Error)
	s.Equal("UPDATE_STATUS_ERROR", errorResp.Code)
}

func (s *MessageHandlerTestSuite) TestUpdateMessageStatus_PublisherError() {
	// Should still succeed even if publisher fails
	bob := testdata.Bob
	validMessages := testdata.ValidMessages()
	testMessage := validMessages[0]

	messageID := domain.MessageID{
		SenderID:   testMessage.SenderID,
		ReceiverID: testMessage.ReceiverID,
		CreatedAt:  testMessage.CreatedAt,
	}

	requestBody := UpdateStatusRequest{
		MessageID: messageID,
	}

	publishError := assert.AnError
	s.mockRepo.On("MarkMessagesUpToRead", mock.Anything, messageID).Return(int64(2), nil)
	s.mockPublisher.On("PublishStatusUpdate", mock.Anything, bob.UserID, mock.Anything).Return(publishError)
	s.mockLogger.On("Error", "Failed to publish status update", "error", publishError, "user", bob.UserID).Return()
	s.mockLogger.On("Debug", "Message status updated successfully", "user", bob.UserID, "count", int64(2), "status", domain.MessageStatusRead).Return()

	req := s.createRequestWithUser("PATCH", "/api/v1/messages/status", requestBody, bob)
	recorder := httptest.NewRecorder()

	// Execute
	s.handler.UpdateMessageStatus(recorder, req)

	// Assertions - should still return 200 even if publishing fails
	s.Equal(http.StatusOK, recorder.Code)
}

func TestMessageHandlerSuite(t *testing.T) {
	suite.Run(t, new(MessageHandlerTestSuite))
}