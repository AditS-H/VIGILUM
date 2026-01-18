// Package middleware_test tests the API Gateway middleware.
package middleware_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vigilum/backend/internal/middleware"
)

// Discard logger for tests.
var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestRecovery(t *testing.T) {
	// Handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap with recovery middleware
	handler := middleware.Recovery(testLogger)(panicHandler)

	// Make request
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Should return 500, not panic
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	// Check JSON response
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["success"] != false {
		t.Error("expected success to be false")
	}
}

func TestRequestID(t *testing.T) {
	// Handler that checks for request ID
	var capturedID string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = middleware.GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequestID(testHandler)

	// Test without existing request ID
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if capturedID == "" {
		t.Error("expected request ID to be set")
	}

	// Check response header
	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}

	// Test with existing request ID
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Request-ID", "existing-id-123")
	rr2 := httptest.NewRecorder()

	handler.ServeHTTP(rr2, req2)

	if capturedID != "existing-id-123" {
		t.Errorf("expected request ID 'existing-id-123', got '%s'", capturedID)
	}
}

func TestLogger(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := middleware.Logger(testLogger)(testHandler)

	req := httptest.NewRequest("GET", "/test?foo=bar", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestCORS(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	config := middleware.DefaultCORSConfig()
	handler := middleware.CORS(config)(testHandler)

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d for preflight, got %d", http.StatusNoContent, rr.Code)
	}

	// Check CORS headers
	if rr.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("expected Access-Control-Allow-Origin header")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}

	// Test regular request
	req2 := httptest.NewRequest("GET", "/test", nil)
	rr2 := httptest.NewRecorder()

	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr2.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.SecurityHeaders(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Check security headers
	headers := []string{
		"X-Content-Type-Options",
		"X-XSS-Protection",
		"X-Frame-Options",
		"Strict-Transport-Security",
		"Content-Security-Policy",
	}

	for _, h := range headers {
		if rr.Header().Get(h) == "" {
			t.Errorf("expected %s header to be set", h)
		}
	}
}

func TestTimeout(t *testing.T) {
	// Handler that takes too long
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			// Context cancelled
			return
		}
	})

	handler := middleware.Timeout(100 * time.Millisecond)(slowHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status %d, got %d", http.StatusGatewayTimeout, rr.Code)
	}
}

func TestChain(t *testing.T) {
	var order []string

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw1-after")
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw2-after")
		})
	}

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Chain(finalHandler, mw1, mw2)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), len(order))
	}

	for i, v := range expected {
		if order[i] != v {
			t.Errorf("expected order[%d] = %s, got %s", i, v, order[i])
		}
	}
}

func TestRateLimiter(t *testing.T) {
	config := middleware.RateLimitConfig{
		AnonymousLimit:  2, // Low limit for testing
		FreeLimit:       5,
		PaidLimit:       10,
		EnterpriseLimit: 100,
		CleanupInterval: time.Minute,
		WindowSize:      time.Minute,
	}

	limiter := middleware.NewRateLimiter(config, testLogger)
	defer limiter.Stop()

	// Test anonymous requests
	for i := 0; i < 3; i++ {
		allowed, remaining, _ := limiter.Allow("ip:192.168.1.1", middleware.TierAnonymous)
		if i < 2 {
			if !allowed {
				t.Errorf("request %d should be allowed", i)
			}
			if remaining != 2-i-1 {
				t.Errorf("expected remaining %d, got %d", 2-i-1, remaining)
			}
		} else {
			if allowed {
				t.Error("request 3 should be rate limited")
			}
		}
	}

	// Test free tier with different key
	for i := 0; i < 6; i++ {
		allowed, _, _ := limiter.Allow("api:user1", middleware.TierFree)
		if i < 5 {
			if !allowed {
				t.Errorf("free tier request %d should be allowed", i)
			}
		} else {
			if allowed {
				t.Error("free tier request 6 should be rate limited")
			}
		}
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	config := middleware.RateLimitConfig{
		AnonymousLimit:  1,
		FreeLimit:       10,
		PaidLimit:       100,
		EnterpriseLimit: 1000,
		CleanupInterval: time.Minute,
		WindowSize:      time.Minute,
	}

	limiter := middleware.NewRateLimiter(config, testLogger)
	defer limiter.Stop()

	getTier := func(r *http.Request) (string, middleware.RateLimitTier) {
		return "ip:test", middleware.TierAnonymous
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RateLimit(limiter, getTier)(testHandler)

	// First request should pass
	req1 := httptest.NewRequest("GET", "/test", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("first request should pass, got status %d", rr1.Code)
	}

	// Check rate limit headers
	if rr1.Header().Get("X-RateLimit-Limit") != "1" {
		t.Errorf("expected X-RateLimit-Limit=1, got %s", rr1.Header().Get("X-RateLimit-Limit"))
	}

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("second request should be rate limited, got status %d", rr2.Code)
	}

	if rr2.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header when rate limited")
	}
}

func TestAuthentication(t *testing.T) {
	validator := func(ctx context.Context, key string) (*middleware.APIKeyInfo, error) {
		if key == "valid-key" {
			return &middleware.APIKeyInfo{
				Key:    key,
				UserID: "user123",
				Tier:   middleware.TierFree,
				Active: true,
			}, nil
		}
		return nil, http.ErrNotSupported
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserID(r.Context())
		w.Write([]byte(userID))
	})

	// Test with required auth
	handler := middleware.Authentication(validator, true, testLogger)(testHandler)

	// No key - should fail
	req1 := httptest.NewRequest("GET", "/test", nil)
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if rr1.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without key, got %d", rr1.Code)
	}

	// Valid key in header
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-API-Key", "valid-key")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("expected 200 with valid key, got %d", rr2.Code)
	}

	if rr2.Body.String() != "user123" {
		t.Errorf("expected user123, got %s", rr2.Body.String())
	}

	// Valid key as Bearer token
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.Header.Set("Authorization", "Bearer valid-key")
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("expected 200 with Bearer token, got %d", rr3.Code)
	}

	// Invalid key
	req4 := httptest.NewRequest("GET", "/test", nil)
	req4.Header.Set("X-API-Key", "invalid-key")
	rr4 := httptest.NewRecorder()
	handler.ServeHTTP(rr4, req4)

	if rr4.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid key, got %d", rr4.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "From RemoteAddr",
			remoteAddr: "192.168.1.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:       "From X-Forwarded-For single",
			headers:    map[string]string{"X-Forwarded-For": "10.0.0.1"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "10.0.0.1",
		},
		{
			name:       "From X-Forwarded-For multiple",
			headers:    map[string]string{"X-Forwarded-For": "10.0.0.1, 172.16.0.1"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "10.0.0.1",
		},
		{
			name:       "From X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "10.0.0.2"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "10.0.0.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			ip := middleware.GetClientIP(req)
			if ip != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, ip)
			}
		})
	}
}
