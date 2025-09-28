package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"messaging-app/internal/ports"
)

type Server struct {
	config Config
	logger ports.Logger
	server *http.Server
	mux    *http.ServeMux
}

type Config struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Auth         AuthConfig
	CORS         CORSConfig
}

type AuthConfig struct {
	UserIDHeader  string
	EmailHeader   string
	HandlerHeader string
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type Route struct {
	Method      string
	Pattern     string
	Handler     http.HandlerFunc
	RequireAuth bool
}

func NewServer(config Config, logger ports.Logger) *Server {
	return &Server{
		config: config,
		logger: logger,
		mux:    http.NewServeMux(),
	}
}

func (s *Server) RegisterRoutes(routes []Route) {
	for _, route := range routes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Pattern)

		var handler http.HandlerFunc
		if route.RequireAuth {
			handler = s.withUserContext(route.Handler)
		} else {
			handler = route.Handler
		}

		s.mux.HandleFunc(pattern, handler)
		s.logger.Debug("Registered route", "method", route.Method, "pattern", route.Pattern, "auth_required", route.RequireAuth)
	}
}

func (s *Server) Initialize() error {
	s.logger.Info("Initializing HTTP server...")

	// Add default health check route
	s.mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Create HTTP server
	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      s.withMiddleware(s.mux),
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	s.logger.Info("HTTP server initialized successfully")
	return nil
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server...", "address", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server failed: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	s.logger.Info("HTTP server shutdown completed")
	return nil
}

// Address returns the server address
func (s *Server) Address() string {
	return s.server.Addr
}