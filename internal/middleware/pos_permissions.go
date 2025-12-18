package middleware

import (
	"Go-Lang-project-01/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireSuperadmin middleware ensures user is superadmin
func RequireSuperadmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by auth middleware)
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Message: "Invalid user data",
			})
			c.Abort()
			return
		}

		if !user.IsSuperAdmin() {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Success: false,
				Message: "Superadmin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireUserOrSuperadmin middleware ensures user is regular user or superadmin
func RequireUserOrSuperadmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Message: "Invalid user data",
			})
			c.Abort()
			return
		}

		// User can create transactions, superadmin has all permissions
		if !user.CanCreateTransaction() {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Success: false,
				Message: "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CanManageProducts middleware checks if user can manage products
func CanManageProducts() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Message: "Invalid user data",
			})
			c.Abort()
			return
		}

		if !user.CanManageProducts() {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Success: false,
				Message: "Product management requires superadmin access",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CanManageStock middleware checks if user can manage stock
func CanManageStock() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Message: "Invalid user data",
			})
			c.Abort()
			return
		}

		if !user.CanManageStock() {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Success: false,
				Message: "Stock management requires superadmin access",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CanViewAnalytics middleware checks if user can view analytics
func CanViewAnalytics() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Message: "Invalid user data",
			})
			c.Abort()
			return
		}

		if !user.CanViewAnalytics() {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Success: false,
				Message: "Analytics access requires admin or superadmin access",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CanManageCatalog middleware checks if user can manage categories/stores
func CanManageCatalog() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Message: "Authentication required",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Message: "Invalid user data",
			})
			c.Abort()
			return
		}

		if !user.CanManageCategories() {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Success: false,
				Message: "Catalog management requires admin or superadmin access",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
