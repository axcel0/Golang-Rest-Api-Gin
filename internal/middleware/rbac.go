package middleware

import (
	"net/http"

	"Go-Lang-project-01/internal/models"

	"github.com/gin-gonic/gin"
)

// RequireRole middleware ensures user has one of the specified roles
func RequireRole(allowedRoles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get user role from context (set by JWT middleware)
		roleInterface, exists := c.Get("user_role")
		if !exists {
			// Fallback: try to get from user object
			userInterface, userExists := c.Get("user")
			if !userExists {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Success: false,
					Message: "unauthorized: user not found in context",
				})
				c.Abort()
				return
			}

			user, ok := userInterface.(*models.User)
			if !ok {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Success: false,
					Message: "internal error: invalid user type",
				})
				c.Abort()
				return
			}
			roleInterface = user.Role
		}

		// Get role string
		roleStr, ok := roleInterface.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Message: "internal error: invalid role type",
			})
			c.Abort()
			return
		}

		// Check if user's role is in allowed roles
		userRole := models.Role(roleStr)
		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Success: false,
			Message: "forbidden: insufficient permissions",
		})
		c.Abort()
	}
}

// RequireSuperAdmin middleware ensures user is superadmin
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleSuperAdmin)
}

// RequireAdmin middleware ensures user is admin or superadmin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin, models.RoleSuperAdmin)
}

// RequireUser middleware ensures user is authenticated (any role)
func RequireUser() gin.HandlerFunc {
	return RequireRole(models.RoleUser, models.RoleAdmin, models.RoleSuperAdmin)
}

// CheckOwnershipOrAdmin checks if user is the owner of resource or admin
func CheckOwnershipOrAdmin(userID uint, requestingUser *models.User) bool {
	// Superadmin and admin can access any resource
	if requestingUser.IsAdmin() {
		return true
	}

	// User can only access their own resources
	return requestingUser.ID == userID
}
