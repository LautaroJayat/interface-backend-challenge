package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	httpAdapter "messaging-app/internal/adapters/http"
	"messaging-app/internal/domain"
	"messaging-app/internal/ports"
)

// MessageHandler handles message-related requests
type MessageHandler struct {
	MessageRepo ports.MessageRepository
	Publisher   ports.MessagePublisher
	Logger      ports.Logger
}

func NewMessageHandler(messageRepo ports.MessageRepository, publisher ports.MessagePublisher, logger ports.Logger) *MessageHandler {
	return &MessageHandler{
		MessageRepo: messageRepo,
		Publisher:   publisher,
		Logger:      logger,
	}
}

// SendMessage handles POST /api/v1/chats/{receiverId}/messages
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// Extract receiverId from path: /api/v1/chats/{receiverId}/messages
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[3] == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing receiver ID", "MISSING_RECEIVER_ID", "receiverId path parameter is required")
		return
	}
	receiverID := pathParts[3]

	user, ok := httpAdapter.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User context not found", "NO_USER_CONTEXT", "")
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", "INVALID_JSON", err.Error())
		return
	}

	// Create message with current timestamp
	message := domain.Message{
		SenderID:   user.UserID,
		ReceiverID: receiverID,
		CreatedAt:  time.Now().UTC(),
		Content:    req.Content,
		Status:     "sent",
	}

	// Validate message
	if err := message.Validate(); err != nil {
		if domain.IsValidationError(err) {
			h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR", err.Error())
		} else {
			h.writeErrorResponse(w, http.StatusInternalServerError, "Internal error", "INTERNAL_ERROR", err.Error())
		}
		return
	}

	// Save to database
	if err := h.MessageRepo.SaveMessage(r.Context(), message); err != nil {
		if err == domain.ErrDuplicateMessage {
			h.writeErrorResponse(w, http.StatusConflict, "Duplicate message", "DUPLICATE_MESSAGE", "Message already exists")
		} else {
			h.Logger.Error("Failed to save message", "error", err, "sender", user.UserID, "receiver", receiverID)
			h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to save message", "SAVE_ERROR", "")
		}
		return
	}

	// Publish to real-time system
	if err := h.Publisher.PublishMessage(r.Context(), message); err != nil {
		h.Logger.Error("Failed to publish message", "error", err, "sender", user.UserID, "receiver", receiverID)
		// Don't fail the request if publishing fails - message is already saved
	}

	// Return response
	response := SendMessageResponse{
		SenderID:   message.SenderID,
		ReceiverID: message.ReceiverID,
		CreatedAt:  message.CreatedAt,
		Content:    message.Content,
		Status:     message.Status,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	h.Logger.Debug("Message sent successfully", "sender", user.UserID, "receiver", receiverID)
}

// GetMessages handles GET /api/v1/chats/{chatId}/messages
func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	// Extract chatId from path: /api/v1/chats/{chatId}/messages
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[3] == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing chat ID", "MISSING_CHAT_ID", "chatId path parameter is required")
		return
	}
	chatID := pathParts[3]

	user, ok := httpAdapter.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User context not found", "NO_USER_CONTEXT", "")
		return
	}

	// Validate user is participant in this chat
	if !h.isUserParticipant(user.UserID, chatID) {
		h.writeErrorResponse(w, http.StatusForbidden, "Access denied", "ACCESS_DENIED", "User is not a participant in this chat")
		return
	}

	// Parse query parameters
	cursorStr := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	var cursor time.Time
	if cursorStr != "" {
		var err error
		cursor, err = time.Parse(time.RFC3339, cursorStr)
		if err != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid cursor format", "INVALID_CURSOR", "Cursor must be RFC3339 formatted timestamp")
			return
		}
	}

	limit := 50 // Default limit
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			h.writeErrorResponse(w, http.StatusBadRequest, "Invalid limit", "INVALID_LIMIT", "Limit must be between 1 and 100")
			return
		}
	}

	// Get messages
	messages, err := h.MessageRepo.GetMessages(r.Context(), chatID, cursor, limit)
	if err != nil {
		h.Logger.Error("Failed to get messages", "error", err, "chat_id", chatID, "user", user.UserID)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get messages", "GET_MESSAGES_ERROR", "")
		return
	}

	// Build response
	response := GetMessagesResponse{
		Messages: messages,
		HasMore:  len(messages) == limit,
	}

	// Set next cursor if there are more messages
	if response.HasMore && len(messages) > 0 {
		lastMessage := messages[len(messages)-1]
		response.NextCursor = lastMessage.CreatedAt.Format(time.RFC3339)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.Logger.Debug("Messages retrieved successfully", "chat_id", chatID, "user", user.UserID, "count", len(messages))
}

// UpdateMessageStatus handles PATCH /api/v1/messages/status
func (h *MessageHandler) UpdateMessageStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := httpAdapter.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User context not found", "NO_USER_CONTEXT", "")
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", "INVALID_JSON", err.Error())
		return
	}

	if req.MessageID.ReceiverID != user.UserID {
		h.writeErrorResponse(w, http.StatusForbidden, "Access denied", "ACCESS_DENIED", "Can only update status of messages you received")
		return
	}

	// Update status
	affected, err := h.MessageRepo.MarkMessagesUpToRead(r.Context(), req.MessageID)

	if err != nil {
		h.Logger.Error("Failed to update message status", "error", err, "user", user.UserID, "message_id", req.MessageID)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update status", "UPDATE_STATUS_ERROR", "")
		return
	}

	// Publish status update
	statusUpdate := ports.StatusUpdate{
		MessageID: req.MessageID,
		Status:    domain.MessageStatusRead,
		UpdatedBy: user.UserID,
		UpdatedAt: time.Now().UTC(),
	}

	if err := h.Publisher.PublishStatusUpdate(r.Context(), user.UserID, statusUpdate); err != nil {
		h.Logger.Error("Failed to publish status update", "error", err, "user", user.UserID)
		// Don't fail the request if publishing fails - status is already updated
	}

	response := UpdateStatusResponse{
		UpdatedCount: affected,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.Logger.Debug("Message status updated successfully", "user", user.UserID, "count", affected, "status", domain.MessageStatusRead)
}

// Helper methods

func (h *MessageHandler) isUserParticipant(userID, chatID string) bool {
	return strings.Contains(chatID, userID)
}

func (h *MessageHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message, code, details string) {
	w.WriteHeader(statusCode)

	response := httpAdapter.ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Logger.Error("Failed to write error response", "error", err)
	}
}
