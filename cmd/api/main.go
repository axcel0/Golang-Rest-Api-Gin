package main

import (
	"fmt"
	"os"
	"time"

	"Go-Lang-project-01/configs"
	_ "Go-Lang-project-01/docs" // Import generated docs
	"Go-Lang-project-01/internal/auth"
	"Go-Lang-project-01/internal/handlers"
	"Go-Lang-project-01/internal/middleware"
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"Go-Lang-project-01/internal/services"
	"Go-Lang-project-01/pkg/database"
	"Go-Lang-project-01/pkg/logger"

	"github.com/gin-gonic/gin"
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
	if err := db.AutoMigrate(&models.User{}); err != nil {
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

	// Initialize dependencies (Dependency Injection)
	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	healthHandler := handlers.NewHealthHandler()

	// Set Gin mode from config
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize Gin router
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.Recovery())     // Panic recovery
	r.Use(middleware.Logger())       // Custom logger
	r.Use(middleware.CORS())         // CORS support
	r.Use(middleware.ErrorHandler()) // Centralized error handling

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

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
	}

	// Start server
	port := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("🚀 Server starting...")
	logger.Info("⚙️  Environment", "mode", cfg.App.Environment)
	logger.Info("🛡️  Rate Limit", "per_minute", cfg.App.RateLimitPerMinute, "burst", cfg.App.RateLimitBurst)
	logger.Info("� JWT Authentication", "access_expiry", cfg.JWT.AccessTokenDuration, "refresh_expiry", cfg.JWT.RefreshTokenDuration)
	logger.Info("�📊 API Endpoints registered")
	logger.Info("   Health endpoints", "liveness", "/health", "readiness", "/ready")
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
