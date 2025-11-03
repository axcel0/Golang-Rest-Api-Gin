package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"Go-Lang-project-01/configs"
	_ "Go-Lang-project-01/docs" // Import generated docs
	"Go-Lang-project-01/graph"
	"Go-Lang-project-01/internal/auth"
	"Go-Lang-project-01/internal/handlers"
	"Go-Lang-project-01/internal/health"
	"Go-Lang-project-01/internal/metrics"
	"Go-Lang-project-01/internal/middleware"
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"Go-Lang-project-01/internal/services"
	"Go-Lang-project-01/internal/websocket"
	"Go-Lang-project-01/pkg/database"
	"Go-Lang-project-01/pkg/logger"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

// @title           Go REST API with JWT Authentication
// @version         1.0
// @description     Production-ready REST API built with Go, Gin, GORM, and JWT authentication
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := configs.LoadConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger from config
	logger.Init(cfg.Logger.Level, cfg.Logger.Format)
	logger.Info("✅ Configuration loaded successfully",
		"environment", cfg.App.Environment,
		"port", cfg.Server.Port,
	)

	// Connect to database (SQLite)
	if err := database.Connect(); err != nil {
		logger.Error("❌ Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Auto migrate
	db := database.GetDB()
	if err := db.AutoMigrate(&models.User{}, &models.AuditLog{}); err != nil {
		logger.Error("❌ Failed to migrate database", "error", err)
		os.Exit(1)
	}
	logger.Info("✅ Database migration completed")

	// Parse JWT token durations
	accessDuration, err := time.ParseDuration(cfg.JWT.AccessTokenDuration)
	if err != nil {
		logger.Error("❌ Invalid access token duration", "error", err)
		os.Exit(1)
	}
	refreshDuration, err := time.ParseDuration(cfg.JWT.RefreshTokenDuration)
	if err != nil {
		logger.Error("❌ Invalid refresh token duration", "error", err)
		os.Exit(1)
	}

	// Initialize JWT Manager
	jwtManager := auth.NewJWTManager(cfg.JWT.SecretKey, accessDuration, refreshDuration)
	logger.Info("✅ JWT authentication initialized")

	// Initialize health service with checkers
	healthService := health.NewHealthService()

	// Register database health checker (5 second timeout)
	healthService.RegisterChecker("database", &health.DatabaseChecker{
		DB:      db,
		Timeout: 5 * time.Second,
	})

	// Register disk space checker (80% warning, 90% critical)
	healthService.RegisterChecker("disk", &health.DiskSpaceChecker{
		Path:              "/",
		WarningThreshold:  80.0,
		CriticalThreshold: 90.0,
	})

	// Register memory checker (500MB warning, 1GB critical)
	healthService.RegisterChecker("memory", &health.MemoryChecker{
		WarningThresholdMB:  500,
		CriticalThresholdMB: 1024,
	})

	logger.Info("✅ Health checks configured")

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run() // Start hub in background
	logger.Info("✅ WebSocket hub initialized")

	// Initialize dependencies (Dependency Injection)
	userRepo := repository.NewUserRepository(db)
	auditRepo := repository.NewAuditLogRepository(db)
	auditService := services.NewAuditService(auditRepo)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager, auditService)
	healthHandler := handlers.NewHealthHandler(healthService)
	wsHandler := handlers.NewWebSocketHandler(wsHub, jwtManager)
	auditHandler := handlers.NewAuditHandler(auditService)

	// Set Gin mode from config
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize Gin router
	r := gin.New()

	// Initialize Prometheus metrics
	prometheusMetrics := metrics.NewMetrics()

	// Apply global middleware
	r.Use(middleware.Recovery())          // Panic recovery
	r.Use(middleware.Logger())            // Custom logger
	r.Use(middleware.CORS())              // CORS support
	r.Use(prometheusMetrics.Middleware()) // Prometheus metrics
	r.Use(middleware.ErrorHandler())      // Centralized error handling

	// Rate limiting middleware (from config)
	// Convert per-minute to per-second: 100 req/min = 100/60 req/sec
	ratePerSecond := rate.Limit(float64(cfg.App.RateLimitPerMinute) / 60.0)
	rateLimiter := middleware.NewRateLimiter(
		ratePerSecond,
		cfg.App.RateLimitBurst,
	)
	r.Use(rateLimiter.RateLimit())

	// Health check routes
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadinessCheck)

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// GraphQL endpoints
	graphqlResolver := &graph.Resolver{
		UserService: userService,
		UserRepo:    userRepo,
		JWTManager:  jwtManager,
	}
	graphqlServer := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: graphqlResolver}))

	// GraphQL playground (only in development)
	if cfg.App.Environment != "production" {
		r.GET("/graphql", gin.WrapH(playground.Handler("GraphQL Playground", "/query")))
		logger.Info("📊 GraphQL Playground enabled", "url", "http://localhost:8080/graphql")
	}

	// GraphQL query endpoint (with JWT authentication)
	r.POST("/query", func(c *gin.Context) {
		// Extract JWT token and add userID to context
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString := authHeader[7:]
			claims, err := jwtManager.ValidateToken(tokenString)
			if err == nil {
				// Add userID to context for resolvers
				c.Request = c.Request.WithContext(
					context.WithValue(c.Request.Context(), "userID", claims.UserID),
				)
			}
		}
		graphqlServer.ServeHTTP(c.Writer, c.Request)
	})
	logger.Info("✅ GraphQL API configured")

	// WebSocket routes
	r.GET("/ws", wsHandler.HandleWebSocket)

	// WebSocket management endpoints (protected)
	wsRoutes := r.Group("/ws")
	wsRoutes.Use(middleware.JWTAuth(jwtManager, userRepo))
	{
		wsRoutes.GET("/stats", wsHandler.GetStats)
		wsRoutes.POST("/broadcast", wsHandler.BroadcastMessage)
	}
	logger.Info("✅ WebSocket endpoints configured")

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public auth routes (no authentication required)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected auth routes (requires authentication)
		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(jwtManager))
		{
			authProtected.GET("/profile", authHandler.GetProfile)
		}

		// User routes (protected with RBAC)
		users := v1.Group("/users")
		users.Use(middleware.JWTAuth(jwtManager, userRepo)) // All user endpoints require authentication
		{
			// Profile management - any authenticated user can access their own profile
			users.GET("/me", userHandler.GetMe)
			users.PUT("/me", userHandler.UpdateMe)
			users.PUT("/me/password", userHandler.ChangePassword)

			// Anyone authenticated can view users
			users.GET("", userHandler.GetAllUsers)
			users.GET("/stats", userHandler.GetUserStats) // Must be before /:id
			users.GET("/:id", userHandler.GetUserByID)

			// Only admin and superadmin can create/update/delete users
			users.POST("", middleware.RequireAdmin(), userHandler.CreateUser)
			users.POST("/batch", middleware.RequireAdmin(), userHandler.BatchCreateUsers)
			users.PUT("/:id", middleware.RequireAdmin(), userHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequireAdmin(), userHandler.DeleteUser)

			// Only superadmin can change roles
			users.PUT("/:id/role", middleware.RequireSuperAdmin(), userHandler.UpdateUserRole)
		}

		// Audit log routes (protected)
		auditLogs := v1.Group("/audit-logs")
		auditLogs.Use(middleware.JWTAuth(jwtManager, userRepo))
		{
			// Any authenticated user can view their own audit logs
			auditLogs.GET("/me", auditHandler.GetMyAuditLogs)

			// Admin endpoints
			auditLogs.GET("", middleware.RequireAdmin(), auditHandler.GetAuditLogs)
			auditLogs.GET("/stats", middleware.RequireAdmin(), auditHandler.GetAuditStats)
			auditLogs.GET("/:id", middleware.RequireAdmin(), auditHandler.GetAuditLog)
			auditLogs.DELETE("/cleanup", middleware.RequireAdmin(), auditHandler.CleanupOldLogs)
		}
	}

	// Start server
	port := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("🚀 Server starting...")
	logger.Info("⚙️  Environment", "mode", cfg.App.Environment)
	logger.Info("🛡️  Rate Limit", "per_minute", cfg.App.RateLimitPerMinute, "burst", cfg.App.RateLimitBurst)
	logger.Info("� JWT Authentication", "access_expiry", cfg.JWT.AccessTokenDuration, "refresh_expiry", cfg.JWT.RefreshTokenDuration)
	logger.Info(" API Endpoints registered")
	logger.Info("   Health endpoints", "liveness", "/health", "readiness", "/ready")
	logger.Info("   Metrics endpoint", "prometheus", "GET /metrics")
	logger.Info("   Auth endpoints",
		"register", "POST /api/v1/auth/register",
		"login", "POST /api/v1/auth/login",
		"refresh", "POST /api/v1/auth/refresh",
		"profile", "GET /api/v1/auth/profile [protected]",
	)
	logger.Info("   User endpoints",
		"list", "GET /api/v1/users",
		"stats", "GET /api/v1/users/stats",
		"get", "GET /api/v1/users/:id",
		"create", "POST /api/v1/users",
		"batch", "POST /api/v1/users/batch",
		"update", "PUT /api/v1/users/:id",
		"delete", "DELETE /api/v1/users/:id",
	)
	logger.Info("🎯 Framework", "name", "Gin", "version", "v1.11.0")
	logger.Info("🌐 Server listening", "address", fmt.Sprintf("http://localhost%s", port))

	if err := r.Run(port); err != nil {
		logger.Error("❌ Failed to start server", "error", err)
		os.Exit(1)
	}
}
