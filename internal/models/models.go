package models

import (
	"time"

	"gorm.io/gorm"
)

// Role represents user role in the system
type Role string

const (
	RoleSuperAdmin Role = "superadmin"
	RoleAdmin      Role = "admin"
	RoleUser       Role = "user"
)

// IsValid checks if role is valid
func (r Role) IsValid() bool {
	return r == RoleSuperAdmin || r == RoleAdmin || r == RoleUser
}

// String returns string representation of role
func (r Role) String() string {
	return string(r)
}

// User represents a user in the system
type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Email       string         `gorm:"uniqueIndex;not null" json:"email"`
	Password    string         `gorm:"default:''" json:"-"` // Password is optional for migration, never exposed in JSON
	Age         int            `gorm:"not null" json:"age"`
	Role        string         `gorm:"type:varchar(20);default:'user';not null" json:"role"` // Role: superadmin, admin, user
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	AvatarURL   string         `gorm:"type:varchar(255)" json:"avatar_url,omitempty"`  // Profile avatar URL
	Bio         string         `gorm:"type:text" json:"bio,omitempty"`                 // User biography
	PhoneNumber string         `gorm:"type:varchar(20)" json:"phone_number,omitempty"` // Contact phone number
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// HasRole checks if user has specific role
func (u *User) HasRole(role Role) bool {
	return Role(u.Role) == role
}

// IsSuperAdmin checks if user is superadmin
func (u *User) IsSuperAdmin() bool {
	return u.HasRole(RoleSuperAdmin)
}

// IsAdmin checks if user is admin or superadmin
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin) || u.HasRole(RoleSuperAdmin)
}

// CanManageUsers checks if user can manage other users
func (u *User) CanManageUsers() bool {
	return u.IsAdmin()
}

// CanDeleteUsers checks if user can delete users
func (u *User) CanDeleteUsers() bool {
	return u.IsAdmin()
}

// CanPromoteUsers checks if user can change roles
func (u *User) CanPromoteUsers() bool {
	return u.IsSuperAdmin()
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6,max=100" example:"password123"`
	Age      int    `json:"age" binding:"required,min=1,max=150" example:"25"`
	Role     string `json:"role" binding:"omitempty,oneof=user admin superadmin" example:"user"` // Optional, defaults to 'user'
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"Jane Doe"`
	Email *string `json:"email,omitempty" binding:"omitempty,email" example:"jane@example.com"`
	Age   *int    `json:"age,omitempty" binding:"omitempty,min=1,max=150" example:"26"`
	Role  *string `json:"role,omitempty" binding:"omitempty,oneof=user admin superadmin" example:"admin"` // Only superadmin can change roles
}

// UpdateRoleRequest represents the request body for updating user role
type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user admin superadmin" example:"admin"`
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

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6,max=100" example:"password123"`
	Age      int    `json:"age" binding:"required,min=1,max=150" example:"25"`
	// Role is not included in registration - all new users start as 'user'
}

// LoginRequest represents the request body for login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse represents the response body for login
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
	User         User   `json:"user"`
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents the response body for token refresh
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"` // seconds
}

// UpdateProfileRequest represents the request body for updating own profile
type UpdateProfileRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=2,max=100" example:"John Doe"`
	Age         *int    `json:"age,omitempty" binding:"omitempty,min=1,max=150" example:"26"`
	AvatarURL   *string `json:"avatar_url,omitempty" binding:"omitempty,url" example:"https://example.com/avatar.jpg"`
	Bio         *string `json:"bio,omitempty" binding:"omitempty,max=500" example:"Software developer"`
	PhoneNumber *string `json:"phone_number,omitempty" binding:"omitempty,min=10,max=20" example:"+628123456789"`
}

// ChangePasswordRequest represents the request body for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6" example:"oldpassword123"`
	NewPassword     string `json:"new_password" binding:"required,min=6,max=100" example:"newpassword123"`
}
