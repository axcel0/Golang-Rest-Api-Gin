package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Age       int            `json:"age"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	Age   int    `json:"age" binding:"required,min=1,max=150" example:"25"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"Jane Doe"`
	Email *string `json:"email,omitempty" binding:"omitempty,email" example:"jane@example.com"`
	Age   *int    `json:"age,omitempty" binding:"omitempty,min=1,max=150" example:"26"`
}

// BatchCreateUsersRequest represents batch user creation request
type BatchCreateUsersRequest struct {
	Users []*CreateUserRequest `json:"users" binding:"required,min=1,max=100,dive"`
}

// PaginationQuery represents pagination query parameters
type PaginationQuery struct {
	Page   int    `form:"page" binding:"omitempty,min=1" example:"1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
	Sort   string `form:"sort" binding:"omitempty,oneof=name email age created_at" example:"created_at"`
	Order  string `form:"order" binding:"omitempty,oneof=asc desc" example:"desc"`
	Search string `form:"search" binding:"omitempty,max=100" example:"john"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedResponse represents paginated API response
type PaginatedResponse struct {
	Success    bool           `json:"success"`
	Message    string         `json:"message,omitempty"`
	Data       interface{}    `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ValidationError represents field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors,omitempty"`
}
