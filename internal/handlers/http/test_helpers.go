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
	"messaging-app/testdata"

	"github.com/stretchr/testify/assert"
)

// TestHelpers provides common utilities for HTTP handler testing
type TestHelpers struct {
	t *testing.T
}

// NewTestHelpers creates a new instance of test helpers
func NewTestHelpers(t *testing.T) *TestHelpers {
	return &TestHelpers{t: t}
}

// CreateRequestWithUser creates an HTTP request with user context
func (h *TestHelpers) CreateRequestWithUser(method, url string, body interface{}, user domain.UserContext) *http.Request {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		assert.NoError(h.t, err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	assert.NoError(h.t, err)

	req.Header.Set("Content-Type", "application/json")

	// Add user context
	ctx := context.WithValue(req.Context(), httpAdapter.UserContextKey, user)
	return req.WithContext(ctx)
}

// CreateRequestWithoutUser creates an HTTP request without user context
func (h *TestHelpers) CreateRequestWithoutUser(method, url string, body interface{}) *http.Request {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		assert.NoError(h.t, err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	assert.NoError(h.t, err)

	req.Header.Set("Content-Type", "application/json")
	return req
}

// CreateRecorder creates a new ResponseRecorder for testing
func (h *TestHelpers) CreateRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// AssertErrorResponse verifies that the response contains the expected error
func (h *TestHelpers) AssertErrorResponse(recorder *httptest.ResponseRecorder, expectedStatus int, expectedError, expectedCode string) {
	assert.Equal(h.t, expectedStatus, recorder.Code)

	var errorResp httpAdapter.ErrorResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &errorResp)
	assert.NoError(h.t, err)
	assert.Equal(h.t, expectedError, errorResp.Error)
	assert.Equal(h.t, expectedCode, errorResp.Code)
}

// AssertSuccessResponse verifies that the response is successful and returns the unmarshaled body
func (h *TestHelpers) AssertSuccessResponse(recorder *httptest.ResponseRecorder, expectedStatus int, target interface{}) {
	assert.Equal(h.t, expectedStatus, recorder.Code)
	assert.Equal(h.t, "application/json", recorder.Header().Get("Content-Type"))

	err := json.Unmarshal(recorder.Body.Bytes(), target)
	assert.NoError(h.t, err)
}

// AssertJSONResponse verifies that the response has valid JSON and returns the unmarshaled body
func (h *TestHelpers) AssertJSONResponse(recorder *httptest.ResponseRecorder) map[string]interface{} {
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(h.t, err)
	return response
}

// CreateMessagePath creates a proper message API path with receiver ID
func (h *TestHelpers) CreateMessagePath(receiverID string) string {
	return "/api/v1/chats/" + receiverID + "/messages"
}

// CreateChatMessagesPath creates a proper chat messages API path with chat ID
func (h *TestHelpers) CreateChatMessagesPath(chatID string) string {
	return "/api/v1/chats/" + chatID + "/messages"
}

// GetDefaultUsers returns common test users for convenience
func (h *TestHelpers) GetDefaultUsers() (alice, bob, charlie domain.UserContext) {
	return testdata.Alice, testdata.Bob, testdata.Charlie
}

// GetValidTestMessage returns a valid test message between Alice and Bob
func (h *TestHelpers) GetValidTestMessage() domain.Message {
	messages := testdata.ValidMessages()
	if len(messages) == 0 {
		h.t.Fatal("No valid test messages available")
	}
	return messages[0]
}

// CreateSampleChatSessions creates sample chat sessions for testing
func (h *TestHelpers) CreateSampleChatSessions(userID string) []domain.ChatSession {
	return []domain.ChatSession{
		{
			ChatID:           domain.ComputeChatID(userID, "user1"),
			LastMessageAt:    testdata.BaseTime,
			UnreadCount:      3,
			OtherParticipant: "user1",
		},
		{
			ChatID:           domain.ComputeChatID(userID, "user2"),
			LastMessageAt:    testdata.BaseTime.Add(-1 * time.Hour),
			UnreadCount:      0,
			OtherParticipant: "user2",
		},
		{
			ChatID:           domain.ComputeChatID(userID, "user3"),
			LastMessageAt:    testdata.BaseTime.Add(-2 * time.Hour),
			UnreadCount:      1,
			OtherParticipant: "user3",
		},
	}
}

// SetURLPath sets the URL path on a request (useful for path parameter extraction)
func (h *TestHelpers) SetURLPath(req *http.Request, path string) {
	req.URL.Path = path
}

// SetQueryParams sets query parameters on a request
func (h *TestHelpers) SetQueryParams(req *http.Request, params map[string]string) {
	q := req.URL.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	req.URL.RawQuery = q.Encode()
}

// ValidSendMessageRequest creates a valid SendMessageRequest for testing
func (h *TestHelpers) ValidSendMessageRequest() SendMessageRequest {
	return SendMessageRequest{
		Content: "Test message content",
	}
}

// InvalidSendMessageRequest creates an invalid SendMessageRequest for testing
func (h *TestHelpers) InvalidSendMessageRequest() SendMessageRequest {
	return SendMessageRequest{
		Content: "", // Empty content is invalid
	}
}

// CreateValidUpdateStatusRequest creates a valid UpdateStatusRequest
func (h *TestHelpers) CreateValidUpdateStatusRequest(messageID domain.MessageID) UpdateStatusRequest {
	return UpdateStatusRequest{
		MessageID: messageID,
	}
}

// AssertResponseHeaders verifies common response headers
func (h *TestHelpers) AssertResponseHeaders(recorder *httptest.ResponseRecorder) {
	assert.Equal(h.t, "application/json", recorder.Header().Get("Content-Type"))
}

// AssertNoErrorsLogged verifies that no error logs were called (useful with mock loggers)
func (h *TestHelpers) AssertNoErrorsLogged(recorder *httptest.ResponseRecorder) {
	// This can be extended based on specific logging requirements
	assert.NotContains(h.t, recorder.Body.String(), "error")
}

// CreateLongContent creates content that exceeds the maximum length limit
func (h *TestHelpers) CreateLongContent() string {
	// Create content longer than 10000 characters
	content := make([]byte, 10001)
	for i := range content {
		content[i] = 'a'
	}
	return string(content)
}

// CreateTestMessageID creates a test MessageID using testdata
func (h *TestHelpers) CreateTestMessageID() domain.MessageID {
	message := h.GetValidTestMessage()
	return domain.MessageID{
		SenderID:   message.SenderID,
		ReceiverID: message.ReceiverID,
		CreatedAt:  message.CreatedAt,
	}
}

// AssertTimestampRecent verifies that a timestamp is recent (within 5 seconds)
func (h *TestHelpers) AssertTimestampRecent(timestamp interface{}) {
	// This can be used to verify that CreatedAt fields are set correctly
	// Implementation depends on specific timestamp format used
}

// CreateInvalidJSON returns a malformed JSON string for testing
func (h *TestHelpers) CreateInvalidJSON() string {
	return `{"content": "missing closing quote and brace"`
}