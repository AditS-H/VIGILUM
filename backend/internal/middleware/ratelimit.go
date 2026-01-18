// Package middleware provides rate limiting for the VIGILUM API Gateway.
package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimitTier defines rate limit tiers.
type RateLimitTier string

const (
	// TierAnonymous for unauthenticated requests (by IP).
	TierAnonymous RateLimitTier = "anonymous"
	// TierFree for free API key users.
	TierFree RateLimitTier = "free"
	// TierPaid for paid API key users.
	TierPaid RateLimitTier = "paid"
	// TierEnterprise for enterprise users.
	TierEnterprise RateLimitTier = "enterprise"
)

// RateLimitConfig configures the rate limiter.
type RateLimitConfig struct {
	// Requests per minute per tier.
	AnonymousLimit   int
	FreeLimit        int
	PaidLimit        int
	EnterpriseLimit  int
	CleanupInterval  time.Duration
	WindowSize       time.Duration
}

// DefaultRateLimitConfig returns default rate limit configuration.
// Based on SYSTEM_DESIGN.md specifications.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		AnonymousLimit:   10,    // 10 req/min per IP
		FreeLimit:        100,   // 100 req/min
		PaidLimit:        1000,  // 1000 req/min
		EnterpriseLimit:  10000, // 10000 req/min
		CleanupInterval:  5 * time.Minute,
		WindowSize:       time.Minute,
	}
}

// rateLimitEntry tracks request counts for a client.
type rateLimitEntry struct {
	Count     int
	WindowEnd time.Time
	Tier      RateLimitTier
}

// RateLimiter implements a sliding window rate limiter.
type RateLimiter struct {
	config  RateLimitConfig
	entries map[string]*rateLimitEntry
	mu      sync.RWMutex
	logger  *slog.Logger
	done    chan struct{}
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(config RateLimitConfig, logger *slog.Logger) *RateLimiter {
	rl := &RateLimiter{
		config:  config,
		entries: make(map[string]*rateLimitEntry),
		logger:  logger,
		done:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// Stop stops the rate limiter cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.done)
}

// cleanupLoop periodically removes expired entries.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.done:
			return
		}
	}
}

// cleanup removes expired entries.
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, entry := range rl.entries {
		if now.After(entry.WindowEnd) {
			delete(rl.entries, key)
		}
	}
}

// getLimitForTier returns the rate limit for a tier.
func (rl *RateLimiter) getLimitForTier(tier RateLimitTier) int {
	switch tier {
	case TierAnonymous:
		return rl.config.AnonymousLimit
	case TierFree:
		return rl.config.FreeLimit
	case TierPaid:
		return rl.config.PaidLimit
	case TierEnterprise:
		return rl.config.EnterpriseLimit
	default:
		return rl.config.AnonymousLimit
	}
}

// Allow checks if a request is allowed and updates the counter.
func (rl *RateLimiter) Allow(key string, tier RateLimitTier) (allowed bool, remaining int, resetAt time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit := rl.getLimitForTier(tier)

	entry, exists := rl.entries[key]
	if !exists || now.After(entry.WindowEnd) {
		// New window
		entry = &rateLimitEntry{
			Count:     1,
			WindowEnd: now.Add(rl.config.WindowSize),
			Tier:      tier,
		}
		rl.entries[key] = entry
		return true, limit - 1, entry.WindowEnd
	}

	if entry.Count >= limit {
		return false, 0, entry.WindowEnd
	}

	entry.Count++
	return true, limit - entry.Count, entry.WindowEnd
}

// RateLimit middleware enforces rate limits.
func RateLimit(limiter *RateLimiter, getTier func(r *http.Request) (string, RateLimitTier)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, tier := getTier(r)
			allowed, remaining, resetAt := limiter.Allow(key, tier)

			// Set rate limit headers
			limit := limiter.getLimitForTier(tier)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))

			if !allowed {
				retryAfter := int(time.Until(resetAt).Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))

				limiter.logger.Warn("Rate limit exceeded",
					"key", key,
					"tier", tier,
					"path", r.URL.Path,
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": fmt.Sprintf("Rate limit exceeded. Try again in %d seconds.", retryAfter),
					},
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// AUTHENTICATION MIDDLEWARE
// ═══════════════════════════════════════════════════════════════════════════

// APIKeyInfo contains information about an API key.
type APIKeyInfo struct {
	Key       string
	UserID    string
	Tier      RateLimitTier
	Active    bool
	ExpiresAt *time.Time
}

// APIKeyValidator is a function that validates an API key.
type APIKeyValidator func(ctx context.Context, key string) (*APIKeyInfo, error)

// Authentication middleware validates API keys.
func Authentication(validator APIKeyValidator, requireAuth bool, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract API key from header or query
			apiKey := r.Header.Get("Authorization")
			if apiKey != "" {
				// Remove "Bearer " prefix if present
				if len(apiKey) > 7 && apiKey[:7] == "Bearer " {
					apiKey = apiKey[7:]
				}
			}
			if apiKey == "" {
				apiKey = r.Header.Get("X-API-Key")
			}
			if apiKey == "" {
				apiKey = r.URL.Query().Get("api_key")
			}

			// If no API key and auth is required, reject
			if apiKey == "" {
				if requireAuth {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"success": false,
						"error": map[string]interface{}{
							"code":    "UNAUTHORIZED",
							"message": "API key required",
						},
						"timestamp": time.Now().UTC().Format(time.RFC3339),
					})
					return
				}
				// Continue without auth for public endpoints
				next.ServeHTTP(w, r)
				return
			}

			// Validate API key
			keyInfo, err := validator(r.Context(), apiKey)
			if err != nil {
				logger.Error("API key validation failed", "error", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"code":    "INVALID_API_KEY",
						"message": "Invalid or expired API key",
					},
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				})
				return
			}

			if !keyInfo.Active {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"code":    "API_KEY_INACTIVE",
						"message": "API key is inactive",
					},
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				})
				return
			}

			if keyInfo.ExpiresAt != nil && time.Now().After(*keyInfo.ExpiresAt) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"code":    "API_KEY_EXPIRED",
						"message": "API key has expired",
					},
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				})
				return
			}

			// Add key info to context
			ctx := context.WithValue(r.Context(), ContextKeyAPIKey, apiKey)
			ctx = context.WithValue(ctx, ContextKeyUserID, keyInfo.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAPIKey retrieves the API key from context.
func GetAPIKey(ctx context.Context) string {
	if key, ok := ctx.Value(ContextKeyAPIKey).(string); ok {
		return key
	}
	return ""
}

// GetUserID retrieves the user ID from context.
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return id
	}
	return ""
}

// ═══════════════════════════════════════════════════════════════════════════
// IP EXTRACTION UTILITIES
// ═══════════════════════════════════════════════════════════════════════════

// GetClientIP extracts the client IP from a request.
// Handles X-Forwarded-For and X-Real-IP headers from load balancers.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain multiple IPs)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP (original client)
		ips := splitIPs(xff)
		if len(ips) > 0 {
			return ips[0]
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// splitIPs splits a comma-separated list of IPs.
func splitIPs(xff string) []string {
	var ips []string
	for _, ip := range splitTrim(xff, ",") {
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

// splitTrim splits a string and trims whitespace.
func splitTrim(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, trimSpace(s[start:i]))
			start = i + len(sep)
		}
	}
	result = append(result, trimSpace(s[start:]))
	return result
}

// trimSpace removes leading/trailing whitespace.
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
