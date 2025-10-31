package middleware

import (
	"net/http"

	"Go-Lang-project-01/internal/models"

	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware for centralized error handling
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			c.JSON(http.StatusInternalServerError, models.Response{
				Success: false,
				Message: err.Error(),
			})
		}
	}
}

// Recovery middleware for panic recovery
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Internal server error",
		})
	})
}
