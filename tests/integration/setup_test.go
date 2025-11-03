package integration

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"Go-Lang-project-01/internal/auth"
	"Go-Lang-project-01/internal/handlers"
	"Go-Lang-project-01/internal/middleware"
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"Go-Lang-project-01/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDB     *gorm.DB
	testRouter *gin.Engine
	jwtManager *auth.JWTManager
	cleanup    func()
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Setup
	setupTestEnvironment()

	// Run tests
	code := m.Run()

	// Cleanup
	if cleanup != nil {
		cleanup()
	}

	os.Exit(code)
}

// setupTestEnvironment initializes the test database and router
func setupTestEnvironment() {
	var err error

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create in-memory SQLite database for testing
	testDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
	})
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Configure SQLite for better concurrent access
	sqlDB, err := testDB.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	sqlDB.SetMaxOpenConns(1) // SQLite only supports 1 connection properly

	// Run migrations
	err = testDB.AutoMigrate(&models.User{}, &models.AuditLog{})
	if err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	// Initialize JWT manager with test config
	jwtManager = auth.NewJWTManager("test-secret-key-for-integration-tests-only", 1*time.Hour, 24*time.Hour)

	// Setup router
	testRouter = setupRouter()

	// Cleanup function
	cleanup = func() {
		sqlDB, _ := testDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// setupRouter creates the router with all routes and middleware
func setupRouter() *gin.Engine {
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Add rate limiting (very high limits for tests)
	rateLimiter := middleware.NewRateLimiter(10000, 1000) // 10000 requests per second, burst of 1000
	router.Use(rateLimiter.RateLimit())

	// Initialize repositories
	userRepo := repository.NewUserRepository(testDB)
	auditRepo := repository.NewAuditLogRepository(testDB)

	// Initialize services
	auditService := services.NewAuditService(auditRepo)
	userService := services.NewUserService(userRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager, auditService)

	// Setup routes
	api := router.Group("/api/v1")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// Protected routes
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtManager))
		{
			// All authenticated users can view
			users.GET("", userHandler.GetAllUsers)
			users.GET("/stats", userHandler.GetUserStats)
			users.GET("/:id", userHandler.GetUserByID)

			// Admin and above can create/update/delete
			users.POST("", middleware.RequireAdmin(), userHandler.CreateUser)
			users.POST("/batch", middleware.RequireAdmin(), userHandler.BatchCreateUsers)
			users.PUT("/:id", middleware.RequireAdmin(), userHandler.UpdateUser)
			users.DELETE("/:id", middleware.RequireAdmin(), userHandler.DeleteUser)

			// Only superadmin can change roles
			users.PUT("/:id/role", middleware.RequireSuperAdmin(), userHandler.UpdateUserRole)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return router
}

// seedTestUser creates a test user and returns the user object
func seedTestUser(role string) (*models.User, error) {
	hashedPassword, err := auth.HashPassword("password123")
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:     fmt.Sprintf("Test %s", role),
		Email:    fmt.Sprintf("%s@test.com", role),
		Password: hashedPassword,
		Age:      30,
		Role:     role,
		IsActive: true,
	}

	result := testDB.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

// getAuthToken generates a JWT token for a test user
func getAuthToken(user *models.User) (string, error) {
	return jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
}

// cleanDatabase truncates all tables
func cleanDatabase() {
	testDB.Exec("DELETE FROM users")
}

// countUsers returns the number of users in the database
func countUsers() int64 {
	var count int64
	testDB.Model(&models.User{}).Count(&count)
	return count
}
