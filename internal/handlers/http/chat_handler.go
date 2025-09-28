package http

import (
	"encoding/json"
	"net/http"

	httpAdapter "messaging-app/internal/adapters/http"
	"messaging-app/internal/ports"
)

// ChatHandler handles chat-related requests
type ChatHandler struct {
	MessageRepo ports.MessageRepository
	Logger      ports.Logger
}

func NewChatHandler(messageRepo ports.MessageRepository, logger ports.Logger) *ChatHandler {
	return &ChatHandler{
		MessageRepo: messageRepo,
		Logger:      logger,
	}
}

// GetChats handles GET /chats
func (h *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	user, ok := httpAdapter.GetUserFromContext(r.Context())
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "User context not found", "NO_USER_CONTEXT", "")
		return
	}

	// Get chat sessions for the user
	sessions, err := h.MessageRepo.GetChatSessions(r.Context(), user.UserID)
	if err != nil {
		h.Logger.Error("Failed to get chat sessions", "error", err, "user", user.UserID)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get chats", "GET_CHATS_ERROR", "")
		return
	}

	response := GetChatsResponse{
		Chats: sessions,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.Logger.Debug("Chat sessions retrieved successfully", "user", user.UserID, "count", len(sessions))
}

func (h *ChatHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message, code, details string) {
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
