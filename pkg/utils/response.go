package utils

import (
	"net/http"
	"strings"

	"Go-Lang-project-01/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

// ValidationErrorResponse sends a validation error response with detailed field errors
func ValidationErrorResponse(c *gin.Context, err error) {
	var validationErrors []models.ValidationError

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			validationErrors = append(validationErrors, models.ValidationError{
				Field:   strings.ToLower(fe.Field()),
				Message: getValidationErrorMessage(fe),
			})
		}
	} else {
		// Generic error fallback
		validationErrors = append(validationErrors, models.ValidationError{
			Field:   "request",
			Message: err.Error(),
		})
	}

	c.JSON(http.StatusBadRequest, models.ErrorResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  validationErrors,
	})
}

// getValidationErrorMessage returns human-readable error message for validation
func getValidationErrorMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		if fe.Type().String() == "string" {
			return field + " must be at least " + fe.Param() + " characters"
		}
		return field + " must be at least " + fe.Param()
	case "max":
		if fe.Type().String() == "string" {
			return field + " must be at most " + fe.Param() + " characters"
		}
		return field + " must be at most " + fe.Param()
	case "oneof":
		return field + " must be one of: " + fe.Param()
	case "dive":
		return "invalid item in " + field
	default:
		return field + " is invalid"
	}
}

// PaginatedResponse sends paginated response
func PaginatedResponse(c *gin.Context, data interface{}, meta models.PaginationMeta) {
	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: meta,
	})
}
