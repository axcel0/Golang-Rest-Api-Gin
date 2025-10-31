package main

import (
	"log"

	"Go-Lang-project-01/internal/handlers"
	"Go-Lang-project-01/internal/middleware"
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"Go-Lang-project-01/internal/services"
	"Go-Lang-project-01/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connect to database (SQLite)
	if err := database.Connect(); err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Auto migrate
	db := database.GetDB()
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}
	log.Println("✅ Database migration completed!")

	// Initialize dependencies (Dependency Injection)
	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)
	healthHandler := handlers.NewHealthHandler()

	// Set Gin mode
	gin.SetMode(gin.ReleaseMode) // Use gin.DebugMode for development

	// Initialize Gin router
	r := gin.New()

	// Apply global middleware
	r.Use(middleware.Recovery())     // Panic recovery
	r.Use(middleware.Logger())       // Custom logger
	r.Use(middleware.CORS())         // CORS support
	r.Use(middleware.ErrorHandler()) // Centralized error handling

	// Health check routes
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/ready", healthHandler.ReadinessCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.GET("", userHandler.GetAllUsers)
			users.GET("/stats", userHandler.GetUserStats) // Must be before /:id
			users.GET("/:id", userHandler.GetUserByID)
			users.POST("", userHandler.CreateUser)
			users.POST("/batch", userHandler.BatchCreateUsers)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	// Start server
	port := ":8080"
	log.Println("🚀 Server starting...")
	log.Println("📊 API Endpoints:")
	log.Println("   Health:")
	log.Println("     - GET    /health         (liveness probe)")
	log.Println("     - GET    /ready          (readiness probe)")
	log.Println("   Users:")
	log.Println("     - GET    /api/v1/users")
	log.Println("     - GET    /api/v1/users/stats")
	log.Println("     - GET    /api/v1/users/:id")
	log.Println("     - POST   /api/v1/users")
	log.Println("     - POST   /api/v1/users/batch")
	log.Println("     - PUT    /api/v1/users/:id")
	log.Println("     - DELETE /api/v1/users/:id")
	log.Println("🎯 Framework: Gin v1.11.0")
	log.Printf("🌐 Server running on http://localhost%s\n", port)

	if err := r.Run(port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
