// Package main is the entry point for the VIGILUM API server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vigilum/backend/internal/config"
	"github.com/vigilum/backend/internal/db"
	"github.com/vigilum/backend/internal/firewall"
	"github.com/vigilum/backend/internal/integration"
)

const version = "0.1.0"

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

	slog.Info("Starting VIGILUM API Server", "version", version)

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

	// Initialize database connection
	database, err := db.New(cfg.Database, logger)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// Initialize Ethereum client (optional - if ETH_RPC_URL is configured)
	var ethClient *integration.EthereumClient
	if os.Getenv("ETH_RPC_URL") != "" {
		ethConfig := integration.DefaultEthereumConfig()
		ethConfig.IdentityFirewallAddr = os.Getenv("IDENTITY_FIREWALL_ADDRESS")
		
		ethClient, err = integration.NewEthereumClient(ethConfig)
		if err != nil {
			slog.Warn("Failed to initialize Ethereum client - proceeding without on-chain integration", 
				"error", err)
		} else {
			slog.Info("Ethereum client initialized",
				"chain_id", ethClient.ChainID().String(),
				"contract", ethConfig.IdentityFirewallAddr,
			)
			defer ethClient.Close()
		}
	} else {
		slog.Info("ETH_RPC_URL not set - on-chain integration disabled")
	}

	// Initialize services
	var firewallService *firewall.Service
	if ethClient != nil {
		firewallService = firewall.NewServiceWithEthereum(database, logger, ethClient)
	} else {
		firewallService = firewall.NewService(database, logger)
	}

	// Initialize HTTP handlers
	firewallHandler := firewall.NewHandler(firewallService, logger)

	// Setup HTTP router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := database.HealthCheck(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","database":"disconnected"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","version":"` + version + `"}`))
	})

	// API info endpoint
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"name": "VIGILUM API",
			"version": "` + version + `",
			"description": "Decentralized Blockchain Security Layer",
			"endpoints": {
				"health": "/health",
				"firewall": "/api/v1/firewall/*"
			}
		}`))
	})

	// Register service routes
	firewallHandler.RegisterRoutes(mux)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      withMiddleware(mux, logger),
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
	slog.Info("HTTP server starting", "port", cfg.Server.HTTPPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error", "error", err)
		os.Exit(1)
	}

	<-ctx.Done()
	slog.Info("VIGILUM API Server shutdown complete")
}

// withMiddleware wraps the handler with common middleware.
func withMiddleware(h http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Serve request
		h.ServeHTTP(wrapped, r)

		// Log request
		logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
