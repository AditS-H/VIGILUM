// Package e2e provides end-to-end integration tests for VIGILUM.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vigilum/backend/internal/middleware"
	"log/slog"
)

// TestAPIGatewayIntegration tests the full API Gateway flow.
func TestAPIGatewayIntegration(t *testing.T) {
	// Create a mock backend service
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back the request details for verification
		response := map[string]interface{}{
			"success":   true,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"data": map[string]interface{}{
				"path":      r.URL.Path,
				"method":    r.Method,
				"requestId": r.Header.Get("X-Request-ID"),
				"apiKey":    r.Header.Get("X-API-Key"),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer backendServer.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create middleware chain similar to API Gateway
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    map[string]interface{}{"status": "healthy"},
		})
	})
	mux.HandleFunc("/api/v1/test/", func(w http.ResponseWriter, r *http.Request) {
		// Forward to mock backend
		resp, err := http.Get(backendServer.URL + r.URL.Path)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	})

	rateLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
		AnonymousLimit:  100,
		FreeLimit:       1000,
		CleanupInterval: time.Minute,
		WindowSize:      time.Minute,
	}, logger)
	defer rateLimiter.Stop()

	handler := middleware.Chain(
		mux,
		middleware.Recovery(logger),
		middleware.RequestID,
		middleware.Logger(logger),
		middleware.CORS(middleware.DefaultCORSConfig()),
		middleware.SecurityHeaders,
	)

	// Create test server
	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("health check", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("health check failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["success"] != true {
			t.Error("expected success=true")
		}
	})

	t.Run("security headers present", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		headers := []string{
			"X-Content-Type-Options",
			"X-XSS-Protection",
			"X-Frame-Options",
		}

		for _, h := range headers {
			if resp.Header.Get(h) == "" {
				t.Errorf("missing security header: %s", h)
			}
		}
	})

	t.Run("request ID propagation", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		requestID := resp.Header.Get("X-Request-ID")
		if requestID == "" {
			t.Error("expected X-Request-ID header")
		}
	})

	t.Run("CORS preflight", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", server.URL+"/health", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("preflight request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected 204 for preflight, got %d", resp.StatusCode)
		}

		if resp.Header.Get("Access-Control-Allow-Origin") == "" {
			t.Error("missing CORS header")
		}
	})
}

// TestRateLimitingIntegration tests rate limiting behavior.
func TestRateLimitingIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create rate limiter with very low limits for testing
	rateLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
		AnonymousLimit:  3,
		FreeLimit:       5,
		CleanupInterval: time.Minute,
		WindowSize:      time.Minute,
	}, logger)
	defer rateLimiter.Stop()

	getTier := func(r *http.Request) (string, middleware.RateLimitTier) {
		// Use a fixed key for all requests in this test
		return "test-key", middleware.TierAnonymous
	}

	handler := middleware.RateLimit(rateLimiter, getTier)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("allows requests within limit", func(t *testing.T) {
		// First 3 requests should pass
		for i := 0; i < 3; i++ {
			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatalf("request %d failed: %v", i, err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("request %d: expected 200, got %d", i, resp.StatusCode)
			}
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		// 4th request should be blocked
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d", resp.StatusCode)
		}

		// Check rate limit headers
		if resp.Header.Get("X-RateLimit-Limit") == "" {
			t.Error("missing X-RateLimit-Limit header")
		}
		if resp.Header.Get("Retry-After") == "" {
			t.Error("missing Retry-After header")
		}
	})
}

// TestAuthenticationIntegration tests authentication flow.
func TestAuthenticationIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	validator := func(ctx context.Context, key string) (*middleware.APIKeyInfo, error) {
		validKeys := map[string]*middleware.APIKeyInfo{
			"valid-key": {
				Key:    "valid-key",
				UserID: "user-123",
				Tier:   middleware.TierFree,
				Active: true,
			},
			"inactive-key": {
				Key:    "inactive-key",
				UserID: "user-456",
				Tier:   middleware.TierFree,
				Active: false,
			},
		}

		if info, ok := validKeys[key]; ok {
			return info, nil
		}
		return nil, fmt.Errorf("invalid key")
	}

	// Protected endpoint
	protectedHandler := middleware.Authentication(validator, true, logger)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := middleware.GetUserID(r.Context())
			w.Write([]byte("Hello, " + userID))
		}),
	)

	server := httptest.NewServer(protectedHandler)
	defer server.Close()

	t.Run("rejects unauthenticated request", func(t *testing.T) {
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("accepts valid API key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header.Set("X-API-Key", "valid-key")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "Hello, user-123" {
			t.Errorf("unexpected body: %s", body)
		}
	})

	t.Run("accepts Bearer token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header.Set("Authorization", "Bearer valid-key")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("rejects inactive key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header.Set("X-API-Key", "inactive-key")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("expected 403, got %d", resp.StatusCode)
		}
	})

	t.Run("rejects invalid key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header.Set("X-API-Key", "invalid-key")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})
}

// TestRecoveryIntegration tests panic recovery.
func TestRecoveryIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("intentional panic for testing")
	})

	handler := middleware.Recovery(logger)(panicHandler)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500 after panic, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["success"] != false {
		t.Error("expected success=false after panic")
	}
}

// TestTimeoutIntegration tests request timeout handling.
func TestTimeoutIntegration(t *testing.T) {
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			return
		}
	})

	handler := middleware.Timeout(100 * time.Millisecond)(slowHandler)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", resp.StatusCode)
	}
}

// TestFullStackIntegration tests a complete request flow through all middleware.
func TestFullStackIntegration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(middleware.DefaultRateLimitConfig(), logger)
	defer rateLimiter.Stop()

	// Create validator
	validator := func(ctx context.Context, key string) (*middleware.APIKeyInfo, error) {
		if key == "test-key" {
			return &middleware.APIKeyInfo{
				Key:    key,
				UserID: "test-user",
				Tier:   middleware.TierFree,
				Active: true,
			}, nil
		}
		return nil, fmt.Errorf("invalid key")
	}

	getTier := func(r *http.Request) (string, middleware.RateLimitTier) {
		if key := r.Header.Get("X-API-Key"); key != "" {
			return "key:" + key, middleware.TierFree
		}
		return "ip:" + r.RemoteAddr, middleware.TierAnonymous
	}

	// Create handler
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/scan", func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		requestID := middleware.GetRequestID(r.Context())

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"userId":    userID,
				"requestId": requestID,
				"message":   "Scan initiated",
			},
		})
	})

	// Apply full middleware stack
	handler := middleware.Chain(
		mux,
		middleware.Recovery(logger),
		middleware.RequestID,
		middleware.Logger(logger),
		middleware.CORS(middleware.DefaultCORSConfig()),
		middleware.SecurityHeaders,
		middleware.Timeout(5*time.Second),
		middleware.RateLimit(rateLimiter, getTier),
		middleware.Authentication(validator, false, logger),
	)

	server := httptest.NewServer(handler)
	defer server.Close()

	// Test authenticated request
	t.Run("authenticated request flow", func(t *testing.T) {
		req, _ := http.NewRequest("POST", server.URL+"/api/v1/scan", bytes.NewReader([]byte(`{"address":"0x123"}`)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		// Check status
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected 200, got %d: %s", resp.StatusCode, body)
		}

		// Check headers
		if resp.Header.Get("X-Request-ID") == "" {
			t.Error("missing request ID header")
		}
		if resp.Header.Get("X-RateLimit-Limit") == "" {
			t.Error("missing rate limit header")
		}

		// Check body
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["success"] != true {
			t.Error("expected success=true")
		}

		data := result["data"].(map[string]interface{})
		if data["userId"] != "test-user" {
			t.Errorf("expected userId=test-user, got %v", data["userId"])
		}
	})
}
