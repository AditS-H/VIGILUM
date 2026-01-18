// Package main is the entry point for the VIGILUM API Gateway.
// The API Gateway is the single entry point for all client requests,
// providing rate limiting, authentication, request routing, and logging.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/vigilum/backend/internal/config"
	"github.com/vigilum/backend/internal/middleware"
)

const version = "0.1.0"

// ServiceConfig defines backend service endpoints.
type ServiceConfig struct {
	FirewallURL string // Identity Firewall Service
	OracleURL   string // Threat Oracle Service
	GenomeURL   string // Genome Analyzer Service
	RedteamURL  string // Red-Team DAO Service
}

// Gateway is the main API Gateway server.
type Gateway struct {
	config      *config.Config
	services    ServiceConfig
	logger      *slog.Logger
	rateLimiter *middleware.RateLimiter
	apiKeys     map[string]*middleware.APIKeyInfo // In-memory cache; use DB in production
}

// NewGateway creates a new API Gateway.
func NewGateway(cfg *config.Config, logger *slog.Logger) *Gateway {
	// Service URLs from environment or defaults
	services := ServiceConfig{
		FirewallURL: getEnvOrDefault("FIREWALL_SERVICE_URL", "http://localhost:8081"),
		OracleURL:   getEnvOrDefault("ORACLE_SERVICE_URL", "http://localhost:8082"),
		GenomeURL:   getEnvOrDefault("GENOME_SERVICE_URL", "http://localhost:8083"),
		RedteamURL:  getEnvOrDefault("REDTEAM_SERVICE_URL", "http://localhost:8084"),
	}

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(
		middleware.DefaultRateLimitConfig(),
		logger,
	)

	return &Gateway{
		config:      cfg,
		services:    services,
		logger:      logger,
		rateLimiter: rateLimiter,
		apiKeys:     make(map[string]*middleware.APIKeyInfo),
	}
}

// Stop stops the gateway.
func (g *Gateway) Stop() {
	g.rateLimiter.Stop()
}

// Handler returns the main HTTP handler with all middleware.
func (g *Gateway) Handler() http.Handler {
	mux := http.NewServeMux()

	// Health check - no auth required
	mux.HandleFunc("GET /health", g.handleHealth)

	// Gateway info
	mux.HandleFunc("GET /", g.handleInfo)

	// API v1 routes - proxy to backend services
	mux.Handle("/api/v1/firewall/", g.proxyHandler(g.services.FirewallURL, "/api/v1/firewall"))
	mux.Handle("/api/v1/oracle/", g.proxyHandler(g.services.OracleURL, "/api/v1/oracle"))
	mux.Handle("/api/v1/genome/", g.proxyHandler(g.services.GenomeURL, "/api/v1/genome"))
	mux.Handle("/api/v1/redteam/", g.proxyHandler(g.services.RedteamURL, "/api/v1/redteam"))

	// Contract scan endpoint - commonly used, route to scanner
	mux.Handle("/api/v1/scan/", g.proxyHandler(g.services.GenomeURL, "/api/v1/scan"))
	mux.Handle("/api/v1/contracts/", g.proxyHandler(g.services.GenomeURL, "/api/v1/contracts"))

	// Alert endpoints - route to oracle
	mux.Handle("/api/v1/alerts/", g.proxyHandler(g.services.OracleURL, "/api/v1/alerts"))

	// Apply middleware chain
	handler := middleware.Chain(
		mux,
		middleware.Recovery(g.logger),
		middleware.RequestID,
		middleware.Logger(g.logger),
		middleware.CORS(middleware.DefaultCORSConfig()),
		middleware.SecurityHeaders,
		middleware.Timeout(30*time.Second),
		middleware.RateLimit(g.rateLimiter, g.getTierForRequest),
		middleware.Authentication(g.validateAPIKey, false, g.logger), // Auth not required for all
	)

	return handler
}

// handleHealth returns the gateway health status.
func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"status":  "healthy",
			"version": version,
			"services": map[string]string{
				"firewall": g.services.FirewallURL,
				"oracle":   g.services.OracleURL,
				"genome":   g.services.GenomeURL,
				"redteam":  g.services.RedteamURL,
			},
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// handleInfo returns gateway information.
func (g *Gateway) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"name":        "VIGILUM API Gateway",
			"version":     version,
			"description": "Decentralized Blockchain Security Layer - API Gateway",
			"endpoints": map[string]string{
				"health":   "/health",
				"firewall": "/api/v1/firewall/*",
				"oracle":   "/api/v1/oracle/*",
				"genome":   "/api/v1/genome/*",
				"redteam":  "/api/v1/redteam/*",
				"scan":     "/api/v1/scan/*",
				"alerts":   "/api/v1/alerts/*",
			},
			"documentation": "https://docs.vigilum.network/api",
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// proxyHandler creates a reverse proxy handler for a backend service.
func (g *Gateway) proxyHandler(targetURL, prefix string) http.Handler {
	target, err := url.Parse(targetURL)
	if err != nil {
		g.logger.Error("Invalid service URL", "url", targetURL, "error", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error": map[string]interface{}{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Backend service configuration error",
				},
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		})
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize the director to modify the request
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Forward request ID
		if reqID := middleware.GetRequestID(req.Context()); reqID != "" {
			req.Header.Set("X-Request-ID", reqID)
		}

		// Forward API key info
		if apiKey := middleware.GetAPIKey(req.Context()); apiKey != "" {
			req.Header.Set("X-API-Key", apiKey)
		}
		if userID := middleware.GetUserID(req.Context()); userID != "" {
			req.Header.Set("X-User-ID", userID)
		}

		// Forward client IP
		req.Header.Set("X-Forwarded-For", middleware.GetClientIP(req))

		// Log the proxy request
		g.logger.Debug("Proxying request",
			"path", req.URL.Path,
			"target", targetURL,
		)
	}

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		g.logger.Error("Proxy error",
			"path", r.URL.Path,
			"target", targetURL,
			"error", err,
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": map[string]interface{}{
				"code":    "SERVICE_UNAVAILABLE",
				"message": "Backend service temporarily unavailable",
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	return http.StripPrefix(prefix, proxy)
}

// getTierForRequest determines the rate limit tier and key for a request.
func (g *Gateway) getTierForRequest(r *http.Request) (string, middleware.RateLimitTier) {
	// Check for API key
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		apiKey = r.Header.Get("Authorization")
		if strings.HasPrefix(apiKey, "Bearer ") {
			apiKey = apiKey[7:]
		}
	}

	if apiKey != "" {
		// Lookup API key info
		if info, ok := g.apiKeys[apiKey]; ok && info.Active {
			return "api:" + info.UserID, info.Tier
		}
		// Unknown key - treat as free tier
		return "key:" + apiKey, middleware.TierFree
	}

	// Anonymous - rate limit by IP
	ip := middleware.GetClientIP(r)
	return "ip:" + ip, middleware.TierAnonymous
}

// validateAPIKey validates an API key.
// In production, this would query a database.
func (g *Gateway) validateAPIKey(ctx context.Context, key string) (*middleware.APIKeyInfo, error) {
	// Check in-memory cache
	if info, ok := g.apiKeys[key]; ok {
		return info, nil
	}

	// For development, accept any key starting with "vgl_"
	if strings.HasPrefix(key, "vgl_") {
		info := &middleware.APIKeyInfo{
			Key:    key,
			UserID: "dev-user",
			Tier:   middleware.TierFree,
			Active: true,
		}
		g.apiKeys[key] = info
		return info, nil
	}

	// For development, accept test keys
	if key == "test-api-key" {
		info := &middleware.APIKeyInfo{
			Key:    key,
			UserID: "test-user",
			Tier:   middleware.TierFree,
			Active: true,
		}
		g.apiKeys[key] = info
		return info, nil
	}

	return nil, fmt.Errorf("invalid API key")
}

// RegisterAPIKey registers an API key (for development/testing).
func (g *Gateway) RegisterAPIKey(info *middleware.APIKeyInfo) {
	g.apiKeys[info.Key] = info
}

func main() {
	// Initialize structured logger
	logLevel := slog.LevelInfo
	if os.Getenv("VIGILUM_ENV") == "development" {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting VIGILUM API Gateway", "version", version)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	slog.Info("Configuration loaded", "env", cfg.Env)

	// Context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create gateway
	gateway := NewGateway(cfg, logger)
	defer gateway.Stop()

	// Create HTTP server
	port := getEnvOrDefaultInt("GATEWAY_PORT", 8080)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      gateway.Handler(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("Received shutdown signal", "signal", sig)

		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Gracefully shutdown HTTP server
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
		}

		cancel()
	}()

	// Start HTTP server
	slog.Info("API Gateway starting",
		"port", port,
		"firewall_service", gateway.services.FirewallURL,
		"oracle_service", gateway.services.OracleURL,
		"genome_service", gateway.services.GenomeURL,
		"redteam_service", gateway.services.RedteamURL,
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error", "error", err)
		os.Exit(1)
	}

	<-ctx.Done()
	slog.Info("VIGILUM API Gateway shutdown complete")
}

// getEnvOrDefault returns environment variable or default value.
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvOrDefaultInt returns environment variable as int or default value.
func getEnvOrDefaultInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var i int
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i
		}
	}
	return defaultVal
}
