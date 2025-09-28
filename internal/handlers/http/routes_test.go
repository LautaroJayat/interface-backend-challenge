package http

import (
	"testing"

	httpAdapter "messaging-app/internal/adapters/http"
	"messaging-app/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RoutesTestSuite struct {
	suite.Suite
	mockRepo      *mocks.MessageRepository
	mockPublisher *mocks.MessagePublisher
	mockLogger    *mocks.Logger
}

func (s *RoutesTestSuite) SetupTest() {
	s.mockRepo = &mocks.MessageRepository{}
	s.mockPublisher = &mocks.MessagePublisher{}
	s.mockLogger = &mocks.Logger{}
}

func (s *RoutesTestSuite) TestMessageRoutes_GetRoutes() {
	messageRoutes := NewMessageRoutes(s.mockRepo, s.mockPublisher, s.mockLogger)
	routes := messageRoutes.GetRoutes()

	// Verify we have the expected number of routes
	s.Len(routes, 3)

	// Create a map for easier lookup
	routeMap := make(map[string]httpAdapter.Route)
	for _, route := range routes {
		key := route.Method + " " + route.Pattern
		routeMap[key] = route
	}

	// Verify SendMessage route
	sendRoute, exists := routeMap["POST /api/v1/chats/{receiverId}/messages"]
	s.True(exists, "SendMessage route should exist")
	s.Equal("POST", sendRoute.Method)
	s.Equal("/api/v1/chats/{receiverId}/messages", sendRoute.Pattern)
	s.True(sendRoute.RequireAuth)
	s.NotNil(sendRoute.Handler)

	// Verify GetMessages route
	getRoute, exists := routeMap["GET /api/v1/chats/{chatId}/messages"]
	s.True(exists, "GetMessages route should exist")
	s.Equal("GET", getRoute.Method)
	s.Equal("/api/v1/chats/{chatId}/messages", getRoute.Pattern)
	s.True(getRoute.RequireAuth)
	s.NotNil(getRoute.Handler)

	// Verify UpdateMessageStatus route
	updateRoute, exists := routeMap["PATCH /api/v1/messages/status"]
	s.True(exists, "UpdateMessageStatus route should exist")
	s.Equal("PATCH", updateRoute.Method)
	s.Equal("/api/v1/messages/status", updateRoute.Pattern)
	s.True(updateRoute.RequireAuth)
	s.NotNil(updateRoute.Handler)
}

func (s *RoutesTestSuite) TestChatRoutes_GetRoutes() {
	chatRoutes := NewChatRoutes(s.mockRepo, s.mockLogger)
	routes := chatRoutes.GetRoutes()

	// Verify we have the expected number of routes
	s.Len(routes, 1)

	route := routes[0]

	// Verify GetChats route
	s.Equal("GET", route.Method)
	s.Equal("/api/v1/chats", route.Pattern)
	s.True(route.RequireAuth)
	s.NotNil(route.Handler)
}

func (s *RoutesTestSuite) TestMessageRoutes_AllRoutesRequireAuth() {
	messageRoutes := NewMessageRoutes(s.mockRepo, s.mockPublisher, s.mockLogger)
	routes := messageRoutes.GetRoutes()

	for _, route := range routes {
		s.True(route.RequireAuth, "Route %s %s should require authentication", route.Method, route.Pattern)
	}
}

func (s *RoutesTestSuite) TestChatRoutes_AllRoutesRequireAuth() {
	chatRoutes := NewChatRoutes(s.mockRepo, s.mockLogger)
	routes := chatRoutes.GetRoutes()

	for _, route := range routes {
		s.True(route.RequireAuth, "Route %s %s should require authentication", route.Method, route.Pattern)
	}
}

func (s *RoutesTestSuite) TestMessageRoutes_HandlerNotNil() {
	messageRoutes := NewMessageRoutes(s.mockRepo, s.mockPublisher, s.mockLogger)
	routes := messageRoutes.GetRoutes()

	for _, route := range routes {
		s.NotNil(route.Handler, "Handler for route %s %s should not be nil", route.Method, route.Pattern)
	}
}

func (s *RoutesTestSuite) TestChatRoutes_HandlerNotNil() {
	chatRoutes := NewChatRoutes(s.mockRepo, s.mockLogger)
	routes := chatRoutes.GetRoutes()

	for _, route := range routes {
		s.NotNil(route.Handler, "Handler for route %s %s should not be nil", route.Method, route.Pattern)
	}
}

func (s *RoutesTestSuite) TestRoutePatterns_FollowAPIConvention() {
	messageRoutes := NewMessageRoutes(s.mockRepo, s.mockPublisher, s.mockLogger)
	chatRoutes := NewChatRoutes(s.mockRepo, s.mockLogger)

	allRoutes := append(messageRoutes.GetRoutes(), chatRoutes.GetRoutes()...)

	for _, route := range allRoutes {
		// All routes should start with /api/v1
		s.True(len(route.Pattern) > 7, "Route pattern should be longer than '/api/v1'")
		s.Equal("/api/v1", route.Pattern[:7], "Route %s should start with /api/v1", route.Pattern)
	}
}

func (s *RoutesTestSuite) TestHTTPMethods_Valid() {
	messageRoutes := NewMessageRoutes(s.mockRepo, s.mockPublisher, s.mockLogger)
	chatRoutes := NewChatRoutes(s.mockRepo, s.mockLogger)

	allRoutes := append(messageRoutes.GetRoutes(), chatRoutes.GetRoutes()...)
	validMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"PATCH":  true,
		"DELETE": true,
	}

	for _, route := range allRoutes {
		s.True(validMethods[route.Method], "HTTP method %s should be valid", route.Method)
	}
}

// Test that we can create route structures without panics
func (s *RoutesTestSuite) TestRouteCreation_NoPanics() {
	s.NotPanics(func() {
		NewMessageRoutes(s.mockRepo, s.mockPublisher, s.mockLogger)
	}, "Creating MessageRoutes should not panic")

	s.NotPanics(func() {
		NewChatRoutes(s.mockRepo, s.mockLogger)
	}, "Creating ChatRoutes should not panic")
}

// Test route patterns for consistency
func (s *RoutesTestSuite) TestRoutePatterns_Consistency() {
	messageRoutes := NewMessageRoutes(s.mockRepo, s.mockPublisher, s.mockLogger)
	routes := messageRoutes.GetRoutes()

	// Check that chat-related routes use consistent path structure
	chatMessageRoutes := 0
	for _, route := range routes {
		if route.Pattern == "/api/v1/chats/{receiverId}/messages" || route.Pattern == "/api/v1/chats/{chatId}/messages" {
			chatMessageRoutes++
		}
	}

	s.Equal(2, chatMessageRoutes, "Should have exactly 2 chat message routes")
}

func TestRoutesSuite(t *testing.T) {
	suite.Run(t, new(RoutesTestSuite))
}

// Additional unit tests for specific route components
func TestNewMessageRoutes(t *testing.T) {
	mockRepo := &mocks.MessageRepository{}
	mockPublisher := &mocks.MessagePublisher{}
	mockLogger := &mocks.Logger{}

	routes := NewMessageRoutes(mockRepo, mockPublisher, mockLogger)

	assert.NotNil(t, routes)
	assert.Equal(t, mockRepo, routes.messageRepo)
	assert.Equal(t, mockPublisher, routes.publisher)
	assert.Equal(t, mockLogger, routes.logger)
}

func TestNewChatRoutes(t *testing.T) {
	mockRepo := &mocks.MessageRepository{}
	mockLogger := &mocks.Logger{}

	routes := NewChatRoutes(mockRepo, mockLogger)

	assert.NotNil(t, routes)
	assert.Equal(t, mockRepo, routes.messageRepo)
	assert.Equal(t, mockLogger, routes.logger)
}

func TestMessageHandler_Creation(t *testing.T) {
	mockRepo := &mocks.MessageRepository{}
	mockPublisher := &mocks.MessagePublisher{}
	mockLogger := &mocks.Logger{}

	handler := NewMessageHandler(mockRepo, mockPublisher, mockLogger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockRepo, handler.MessageRepo)
	assert.Equal(t, mockPublisher, handler.Publisher)
	assert.Equal(t, mockLogger, handler.Logger)
}

func TestChatHandler_Creation(t *testing.T) {
	mockRepo := &mocks.MessageRepository{}
	mockLogger := &mocks.Logger{}

	handler := NewChatHandler(mockRepo, mockLogger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockRepo, handler.MessageRepo)
	assert.Equal(t, mockLogger, handler.Logger)
}