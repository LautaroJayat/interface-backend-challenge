package http

import (
	httpAdapter "messaging-app/internal/adapters/http"
	"messaging-app/internal/ports"
)

type MessageRoutes struct {
	messageRepo ports.MessageRepository
	publisher   ports.MessagePublisher
	logger      ports.Logger
}

func NewMessageRoutes(messageRepo ports.MessageRepository, publisher ports.MessagePublisher, logger ports.Logger) *MessageRoutes {
	return &MessageRoutes{
		messageRepo: messageRepo,
		publisher:   publisher,
		logger:      logger,
	}
}

func (mr *MessageRoutes) GetRoutes() []httpAdapter.Route {
	handler := NewMessageHandler(mr.messageRepo, mr.publisher, mr.logger)

	return []httpAdapter.Route{
		{
			Method:      "POST",
			Pattern:     "/api/v1/chats/{receiverId}/messages",
			Handler:     handler.SendMessage,
			RequireAuth: true,
		},
		{
			Method:      "GET",
			Pattern:     "/api/v1/chats/{chatId}/messages",
			Handler:     handler.GetMessages,
			RequireAuth: true,
		},
		{
			Method:      "PATCH",
			Pattern:     "/api/v1/messages/status",
			Handler:     handler.UpdateMessageStatus,
			RequireAuth: true,
		},
	}
}
