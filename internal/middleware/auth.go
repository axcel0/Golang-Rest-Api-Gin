package middleware

import (
	"net/http"
	"strings"

	"Go-Lang-project-01/internal/auth"
	"Go-Lang-project-01/internal/repository"
	"Go-Lang-project-01/pkg/logger"

	"github.com/gin-gonic/gin"
)

// JWTAuth validates JWT token and loads user into context (for RBAC)
func JWTAuth(jwtManager *auth.JWTManager, userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header", "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "authorization header required",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid authorization format", "header", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid authorization format (use: Bearer <token>)",
			})
			c.Abort()
			return
		}

		// Validate token
		token := parts[1]
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			logger.Warn("Invalid token", "error", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Fetch user from database to get role
		user, err := userRepo.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			logger.Warn("User not found", "user_id", claims.UserID)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "user not found",
			})
			c.Abort()
			return
		}

		// Check if user is active
		if !user.IsActive {
			logger.Warn("Inactive user attempted access", "user_id", user.ID)
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "account is inactive",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user", user) // For RBAC checks

		logger.Debug("User authenticated", "user_id", claims.UserID, "email", claims.Email, "role", user.Role)

		c.Next()
	}
}

// AuthMiddleware validates JWT token from Authorization header (backward compatibility)
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header", "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "authorization header required",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid authorization format", "header", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid authorization format (use: Bearer <token>)",
			})
			c.Abort()
			return
		}

		// Validate token
		token := parts[1]
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			logger.Warn("Invalid token", "error", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role) // Add role for RBAC

		logger.Debug("User authenticated", "user_id", claims.UserID, "email", claims.Email, "role", claims.Role)

		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT token but doesn't abort if missing
func OptionalAuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			if claims, err := jwtManager.ValidateToken(token); err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
			}
		}

		c.Next()
	}
}
