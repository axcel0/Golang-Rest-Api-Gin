package utils

import (
	"net/http"

	"Go-Lang-project-01/internal/models"
	"github.com/gin-gonic/gin"
)

// Response helpers untuk mengurangi boilerplate code

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    data,
	})
}

// CreatedResponse sends a created response
func CreatedResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.Response{
		Success: false,
		Message: message,
	})
}

// ValidationErrorResponse sends a validation error response
func ValidationErrorResponse(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, models.Response{
		Success: false,
		Message: "Validation error: " + err.Error(),
	})
}
