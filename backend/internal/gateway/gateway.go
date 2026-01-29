// Package gateway implements the API Gateway for request routing and rate limiting
package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RouteConfig defines a backend service route
type RouteConfig struct {
	Path           string
	TargetURL      string
	Timeout        time.Duration
	RateLimit      int // requests per second
	CircuitBreaker bool
}

// APIGateway routes requests to backend services
type APIGateway struct {
	logger        *slog.Logger
	routes        map[string]*routeHandler
	rateLimiters  map[string]*rate.Limiter
	circuitBreaker map[string]*CircuitBreaker
	mu            sync.RWMutex
}

// routeHandler handles a single route
type routeHandler struct {
	proxy      *httputil.ReverseProxy
	config     RouteConfig
	limiter    *rate.Limiter
	breaker    *CircuitBreaker
	statsLock  sync.RWMutex
	stats      RouteStats
}

// RouteStats tracks metrics for a route
type RouteStats struct {
	Requests      int64
	Success       int64
	Failures      int64
	Timeouts      int64
	TotalLatency   time.Duration
	LastError     string
	LastErrorTime time.Time
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	maxFailures      int
	resetTimeout     time.Duration
	failureCount     int
	lastFailureTime  time.Time
	state            string // "closed", "open", "half-open"
	stateChangeTime  time.Time
	mu               sync.RWMutex
}

// NewAPIGateway creates a new API Gateway
func NewAPIGateway(logger *slog.Logger) *APIGateway {
	return &APIGateway{
		logger:         logger.With("service", "api-gateway"),
		routes:         make(map[string]*routeHandler),
		rateLimiters:   make(map[string]*rate.Limiter),
		circuitBreaker: make(map[string]*CircuitBreaker),
	}
}

// RegisterRoute adds a new backend service route
func (gw *APIGateway) RegisterRoute(config RouteConfig) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	// Parse target URL
	target, err := url.Parse(config.TargetURL)
	if err != nil {
		return fmt.Errorf("invalid target URL: %w", err)
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Create rate limiter (requests per second)
	limiter := rate.NewLimiter(rate.Limit(config.RateLimit), config.RateLimit*10)

	// Create circuit breaker
	var breaker *CircuitBreaker
	if config.CircuitBreaker {
		breaker = &CircuitBreaker{
			maxFailures:  5,
			resetTimeout: 30 * time.Second,
			state:        "closed",
		}
	}

	handler := &routeHandler{
		proxy:   proxy,
		config:  config,
		limiter: limiter,
		breaker: breaker,
	}

	gw.routes[config.Path] = handler

	gw.logger.Info("route registered",
		"path", config.Path,
		"target", config.TargetURL,
		"rate_limit", config.RateLimit,
		"circuit_breaker", config.CircuitBreaker,
	)

	return nil
}

// ServeHTTP handles incoming requests
func (gw *APIGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	gw.mu.RLock()
	handler, exists := gw.routes[r.URL.Path]
	gw.mu.RUnlock()

	if !exists {
		http.Error(w, "route not found", http.StatusNotFound)
		return
	}

	// Check circuit breaker
	if handler.breaker != nil {
		if !handler.breaker.Allow() {
			gw.logger.Warn("circuit breaker open", "path", r.URL.Path)
			http.Error(w, "service unavailable", http.StatusServiceUnavailable)
			return
		}
	}

	// Check rate limit
	if !handler.limiter.Allow() {
		gw.logger.Warn("rate limit exceeded",
			"path", r.URL.Path,
			"client", r.RemoteAddr,
		)
		w.Header().Set("Retry-After", "1")
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	// Create response writer wrapper to track status
	wrapped := &responseWriterWrapper{ResponseWriter: w}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), handler.config.Timeout)
	defer cancel()

	// Add trace headers
	r.Header.Set("X-Request-ID", generateRequestID())
	r.Header.Set("X-Forwarded-For", r.RemoteAddr)

	// Proxy request
	done := make(chan struct{})
	go func() {
		handler.proxy.ServeHTTP(wrapped, r.WithContext(ctx))
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		// Request completed
	case <-ctx.Done():
		http.Error(w, "request timeout", http.StatusGatewayTimeout)
		handler.recordFailure(time.Since(start), "timeout")
		return
	}

	// Update statistics
	latency := time.Since(start)
	handler.statsLock.Lock()
	handler.stats.Requests++
	handler.stats.TotalLatency += latency

	if wrapped.statusCode >= 200 && wrapped.statusCode < 300 {
		handler.stats.Success++
		if handler.breaker != nil {
			handler.breaker.RecordSuccess()
		}
	} else if wrapped.statusCode >= 500 {
		handler.stats.Failures++
		handler.stats.LastError = fmt.Sprintf("HTTP %d", wrapped.statusCode)
		handler.stats.LastErrorTime = time.Now()
		if handler.breaker != nil {
			handler.breaker.RecordFailure()
		}
	}
	handler.statsLock.Unlock()

	gw.logger.Info("request processed",
		"path", r.URL.Path,
		"method", r.Method,
		"status", wrapped.statusCode,
		"latency_ms", latency.Milliseconds(),
	)
}

// GetStats returns statistics for all routes
func (gw *APIGateway) GetStats() map[string]RouteStats {
	gw.mu.RLock()
	defer gw.mu.RUnlock()

	stats := make(map[string]RouteStats)
	for path, handler := range gw.routes {
		handler.statsLock.RLock()
		stats[path] = handler.stats
		handler.statsLock.RUnlock()
	}

	return stats
}

// Helper types and methods

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	if !w.written {
		w.statusCode = statusCode
		w.written = true
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}
	return w.ResponseWriter.Write(b)
}

func (h *routeHandler) recordFailure(latency time.Duration, reason string) {
	h.statsLock.Lock()
	defer h.statsLock.Unlock()

	h.stats.Requests++
	h.stats.Failures++
	h.stats.LastError = reason
	h.stats.LastErrorTime = time.Now()
	h.stats.TotalLatency += latency

	if h.breaker != nil {
		h.breaker.RecordFailure()
	}
}

// CircuitBreaker methods

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == "open" {
		if time.Since(cb.stateChangeTime) > cb.resetTimeout {
			cb.state = "half-open"
			return true
		}
		return false
	}

	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	if cb.state != "closed" {
		cb.state = "closed"
		cb.stateChangeTime = time.Now()
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.maxFailures {
		cb.state = "open"
		cb.stateChangeTime = time.Now()
	}
}

func generateRequestID() string {
	// In production, use proper UUID generation
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
