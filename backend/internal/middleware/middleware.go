// Package middleware provides HTTP middleware for the VIGILUM API Gateway.
package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// ContextKey is used for context values.
type ContextKey string

const (
	// ContextKeyRequestID is the context key for request ID.
	ContextKeyRequestID ContextKey = "request_id"
	// ContextKeyAPIKey is the context key for authenticated API key.
	ContextKeyAPIKey ContextKey = "api_key"
	// ContextKeyUserID is the context key for user ID.
	ContextKeyUserID ContextKey = "user_id"
)

// ResponseWriter wraps http.ResponseWriter to capture status code.
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Written    int64
}

// WriteHeader captures the status code.
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures bytes written.
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.Written += int64(n)
	return n, err
}

// Chain applies middleware in order.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// ═══════════════════════════════════════════════════════════════════════════
// RECOVERY MIDDLEWARE
// ═══════════════════════════════════════════════════════════════════════════

// Recovery middleware recovers from panics and returns 500.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					logger.Error("Panic recovered",
						"error", err,
						"path", r.URL.Path,
						"stack", string(stack),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"success": false,
						"error": map[string]interface{}{
							"code":    "INTERNAL_ERROR",
							"message": "An unexpected error occurred",
						},
						"timestamp": time.Now().UTC().Format(time.RFC3339),
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// REQUEST ID MIDDLEWARE
// ═══════════════════════════════════════════════════════════════════════════

var requestIDCounter uint64
var requestIDMu sync.Mutex

// RequestID middleware adds a unique request ID to each request.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing request ID from load balancer
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestIDMu.Lock()
			requestIDCounter++
			requestID = fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), requestIDCounter)
			requestIDMu.Unlock()
		}

		// Add to context and response header
		ctx := context.WithValue(r.Context(), ContextKeyRequestID, requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return id
	}
	return ""
}

// ═══════════════════════════════════════════════════════════════════════════
// LOGGING MIDDLEWARE
// ═══════════════════════════════════════════════════════════════════════════

// Logger middleware logs all HTTP requests.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer
			wrapped := &ResponseWriter{
				ResponseWriter: w,
				StatusCode:     http.StatusOK,
			}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Log request
			duration := time.Since(start)
			logger.Info("HTTP request",
				"request_id", GetRequestID(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", wrapped.StatusCode,
				"bytes", wrapped.Written,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// CORS MIDDLEWARE
// ═══════════════════════════════════════════════════════════════════════════

// CORSConfig configures CORS settings.
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int // seconds
}

// DefaultCORSConfig returns default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-API-Key", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORS middleware handles Cross-Origin Resource Sharing.
func CORS(config CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				origin = "*"
			}

			// Check if origin is allowed
			allowedOrigin := "*"
			for _, allowed := range config.AllowOrigins {
				if allowed == "*" || allowed == origin {
					allowedOrigin = origin
					break
				}
			}

			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// TIMEOUT MIDDLEWARE
// ═══════════════════════════════════════════════════════════════════════════

// Timeout middleware adds a timeout to request handling.
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusGatewayTimeout)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"code":    "TIMEOUT",
						"message": "Request timed out",
					},
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				})
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// SECURITY HEADERS MIDDLEWARE
// ═══════════════════════════════════════════════════════════════════════════

// SecurityHeaders adds security-related HTTP headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Clickjacking protection
		w.Header().Set("X-Frame-Options", "DENY")
		// Strict transport security
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// Content security policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		next.ServeHTTP(w, r)
	})
}
