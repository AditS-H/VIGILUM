// Package api provides HTTP API routing and middleware setup.
package api

import (
	"database/sql"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/vigilum/backend/internal/api/handlers"
	"github.com/vigilum/backend/internal/db/repositories"
	"github.com/vigilum/backend/internal/proof"
	zkproof "github.com/vigilum/backend/internal/proof/zkproof"
)

// APIServer wraps the Gin router and handlers.
type APIServer struct {
	router       *gin.Engine
	proofHandler *handlers.ProofHandler
	logger       *slog.Logger
}

// NewAPIServer creates a new API server with routing.
func NewAPIServer(
	db *sql.DB,
	zkConfig zkproof.ProofServiceConfig,
	logger *slog.Logger,
) *APIServer {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(nil, nil))
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	proofRepo := repositories.NewHumanProofRepository(db)

	// Initialize proof verifier
	verifier := proof.NewHumanProofVerifier(db, zkConfig, logger)

	// Create handlers
	proofHandler := handlers.NewProofHandler(verifier, logger)

	// Create router
	router := gin.Default()

	// Add middleware
	router.Use(LoggingMiddleware(logger))
	router.Use(ErrorHandlingMiddleware(logger))
	router.Use(CORSMiddleware())

	server := &APIServer{
		router:       router,
		proofHandler: proofHandler,
		logger:       logger,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures all API routes.
func (as *APIServer) setupRoutes() {
	v1 := as.router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", as.proofHandler.Health)

		// Proof endpoints
		proofs := v1.Group("/proofs")
		{
			// Generate challenge
			proofs.POST("/challenges", as.proofHandler.GenerateChallenge)

			// Submit proof for verification
			proofs.POST("/verify", as.proofHandler.SubmitProof)

			// Get user proofs with pagination
			proofs.GET("", as.proofHandler.GetUserProofs)

			// Get challenge status
			proofs.GET("/challenges/:challenge_id", as.proofHandler.GetChallengeStatus)
		}

		// User endpoints
		users := v1.Group("/users")
		{
			// Get user verification score
			users.GET("/verification-score", as.proofHandler.GetVerificationScore)
		}

		// Firewall endpoints (documented in design)
		firewall := v1.Group("/firewall")
		{
			// Verify proof (alias for /proofs/verify)
			firewall.POST("/verify-proof", as.proofHandler.SubmitProof)

			// Get challenge
			firewall.GET("/challenge", as.proofHandler.GenerateChallenge)

			// Get risk score
			firewall.GET("/risk/:address", as.proofHandler.GetVerificationScore)

			// Stats
			firewall.GET("/stats", as.proofHandler.Health)
		}
	}

	as.logger.Info("API routes configured")
}

// Router returns the underlying Gin router.
func (as *APIServer) Router() *gin.Engine {
	return as.router
}

// Start starts the API server.
func (as *APIServer) Start(addr string) error {
	as.logger.Info("Starting API server", slog.String("address", addr))
	return as.router.Run(addr)
}

// LoggingMiddleware logs HTTP requests and responses.
func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request
		logger.Info("API request received",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("remote_addr", c.RemoteIP()),
		)

		// Process request
		c.Next()

		// Log response
		logger.Info("API response sent",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status_code", c.Writer.Status()),
		)
	}
}

// ErrorHandlingMiddleware handles panics and errors.
func ErrorHandlingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("API panic recovered",
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.Any("panic", r),
				)
				c.JSON(500, gin.H{
					"error":   "internal_server_error",
					"message": "An unexpected error occurred",
				})
			}
		}()

		c.Next()
	}
}

// RateLimitingMiddleware implements rate limiting (placeholder).
func RateLimitingMiddleware(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// In production: Use redis-based rate limiting
		// For now: Just pass through
		c.Next()
	}
}

// AuthenticationMiddleware validates API keys (placeholder).
func AuthenticationMiddleware(userRepo repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(401, gin.H{
				"error":   "unauthorized",
				"message": "API key required",
			})
			c.Abort()
			return
		}

		// In production: Validate API key against database
		// For now: Just pass through
		c.Next()
	}
}

// CORSMiddleware handles CORS headers.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
