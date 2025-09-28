package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"messaging-app/internal/domain"
)

// withMiddleware applies all global middleware to the handler
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return s.withRecovery(
		s.withLogging(
			s.withCORS(
				s.withContentType(next),
			),
		),
	)
}

// withUserContext extracts user context from configured headers
func (s *Server) withUserContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get(s.config.Auth.UserIDHeader)
		email := r.Header.Get(s.config.Auth.EmailHeader)
		handler := r.Header.Get(s.config.Auth.HandlerHeader)

		if userID == "" {
			s.writeErrorResponse(w, http.StatusUnauthorized, "Missing user context", "MISSING_USER_ID",
				s.config.Auth.UserIDHeader+" header is required")
			return
		}

		userContext := domain.UserContext{
			UserID:  userID,
			Email:   email,
			Handler: handler,
		}

		if err := userContext.Validate(); err != nil {
			s.writeErrorResponse(w, http.StatusUnauthorized, "Invalid user context", "INVALID_USER_CONTEXT", err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userContext)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// withLogging logs HTTP requests
func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		s.logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", duration,
			"user_agent", r.UserAgent(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

// withCORS handles CORS headers using configuration
func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set allowed origins
		if len(s.config.CORS.AllowedOrigins) > 0 {
			origin := r.Header.Get("Origin")
			if origin != "" {
				for _, allowedOrigin := range s.config.CORS.AllowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
						break
					}
				}
			}
		} else {
			// Default to allow all if not configured
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set allowed methods
		if len(s.config.CORS.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(s.config.CORS.AllowedMethods, ", "))
		} else {
			// Default methods
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		// Set allowed headers
		if len(s.config.CORS.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(s.config.CORS.AllowedHeaders, ", "))
		} else {
			// Default headers including auth headers
			defaultHeaders := []string{
				"Content-Type",
				"Authorization",
				s.config.Auth.UserIDHeader,
				s.config.Auth.EmailHeader,
				s.config.Auth.HandlerHeader,
			}
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(defaultHeaders, ", "))
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// withRecovery recovers from panics
func (s *Server) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.Error("Panic recovered", "error", err, "path", r.URL.Path)
				s.writeErrorResponse(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR", "An unexpected error occurred")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// withContentType ensures JSON content type for API endpoints
func (s *Server) withContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Helper functions

func (s *Server) writeErrorResponse(w http.ResponseWriter, statusCode int, message, code, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to write error response", "error", err)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
