package http

import (
	httpAdapter "messaging-app/internal/adapters/http"
	"messaging-app/internal/ports"
)

type ChatRoutes struct {
	messageRepo ports.MessageRepository
	logger      ports.Logger
}

func NewChatRoutes(messageRepo ports.MessageRepository, logger ports.Logger) *ChatRoutes {
	return &ChatRoutes{
		messageRepo: messageRepo,
		logger:      logger,
	}
}

func (cr *ChatRoutes) GetRoutes() []httpAdapter.Route {
	handler := NewChatHandler(cr.messageRepo, cr.logger)

	return []httpAdapter.Route{
		{
			Method:      "GET",
			Pattern:     "/api/v1/chats",
			Handler:     handler.GetChats,
			RequireAuth: true,
		},
	}
}
