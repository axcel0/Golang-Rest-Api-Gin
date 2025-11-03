package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"Go-Lang-project-01/internal/auth"
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/services"
	"Go-Lang-project-01/pkg/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests
type UserHandler struct {
	service *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetAllUsers godoc
// @Summary      List all users
// @Description  Get all users with pagination, search, filter, and sort
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        page     query     int     false  "Page number (default: 1)"
// @Param        limit    query     int     false  "Items per page (default: 10)"
// @Param        sort     query     string  false  "Sort field (default: created_at)"
// @Param        order    query     string  false  "Sort order: asc or desc (default: desc)"
// @Param        search   query     string  false  "Search in name and email"
// @Param        active   query     bool    false  "Filter by active status"
// @Success      200      {object}  map[string]interface{}  "List of users with pagination metadata"
// @Failure      400      {object}  map[string]interface{}  "Invalid query parameters"
// @Failure      500      {object}  map[string]interface{}  "Internal server error"
// @Router       /users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse and validate query parameters
	var query models.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get paginated users
	users, meta, err := h.service.GetAllUsersPaginated(ctx, query)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.PaginatedResponse(c, users, meta)
}

// GetUserByID godoc
// @Summary      Get user by ID
// @Description  Get a single user by their ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int                     true  "User ID"
// @Success      200  {object}  map[string]interface{}  "User found"
// @Failure      400  {object}  map[string]interface{}  "Invalid user ID"
// @Failure      404  {object}  map[string]interface{}  "User not found"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.GetUserByID(ctx, uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.SuccessResponse(c, user)
}

// CreateUser godoc
// @Summary      Create new user
// @Description  Create a new user with the provided information
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      models.CreateUserRequest  true  "User creation request"
// @Success      201      {object}  map[string]interface{}    "User created successfully"
// @Failure      400      {object}  map[string]interface{}    "Invalid request body"
// @Failure      500      {object}  map[string]interface{}    "Internal server error"
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	user, err := h.service.CreateUser(ctx, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.CreatedResponse(c, "user created successfully", user)
}

// UpdateUser godoc
// @Summary      Update user
// @Description  Update an existing user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id       path      int                       true  "User ID"
// @Param        request  body      models.UpdateUserRequest  true  "User update request"
// @Success      200      {object}  map[string]interface{}    "User updated successfully"
// @Failure      400      {object}  map[string]interface{}    "Invalid request"
// @Failure      404      {object}  map[string]interface{}    "User not found"
// @Failure      500      {object}  map[string]interface{}    "Internal server error"
// @Router       /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	user, err := h.service.UpdateUser(ctx, uint(id), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "user updated successfully",
		Data:    user,
	})
}

// DeleteUser godoc
// @Summary      Delete user
// @Description  Soft delete a user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int                     true  "User ID"
// @Success      200  {object}  map[string]interface{}  "User deleted successfully"
// @Failure      400  {object}  map[string]interface{}  "Invalid user ID"
// @Failure      404  {object}  map[string]interface{}  "User not found"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.service.DeleteUser(ctx, uint(id)); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "user deleted successfully",
	})
}

// BatchCreateUsers godoc
// @Summary      Batch create users
// @Description  Create multiple users in a single request
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      models.BatchCreateUsersRequest  true  "Batch user creation request"
// @Success      201      {object}  map[string]interface{}          "Users created successfully"
// @Failure      400      {object}  map[string]interface{}          "Invalid request body"
// @Failure      500      {object}  map[string]interface{}          "Internal server error"
// @Router       /users/batch [post]
func (h *UserHandler) BatchCreateUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var requests []*models.CreateUserRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	users, err := h.service.BatchCreateUsers(ctx, requests)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: err.Error(),
			Data:    users,
		})
		return
	}

	utils.CreatedResponse(c, "users created successfully", users)
}

// GetUserStats godoc
// @Summary      Get user statistics
// @Description  Get statistics about users (total count, active count, etc.)
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "User statistics"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /users/stats [get]
func (h *UserHandler) GetUserStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats, err := h.service.GetUserStats(ctx)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, stats)
}

// UpdateUserRole godoc
// @Summary      Update user role
// @Description  Update user role (superadmin only)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id       path      int                        true   "User ID"
// @Param        request  body      models.UpdateRoleRequest   true   "Role update request"
// @Success      200      {object}  map[string]interface{}     "User role updated successfully"
// @Failure      400      {object}  map[string]interface{}     "Invalid request"
// @Failure      403      {object}  map[string]interface{}     "Forbidden: superadmin only"
// @Failure      404      {object}  map[string]interface{}     "User not found"
// @Failure      500      {object}  map[string]interface{}     "Internal server error"
// @Router       /users/{id}/role [put]
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	var requestingUserID uint
	
	// Get requesting user role from context
	userRoleInterface, exists := c.Get("user_role")
	if !exists {
		// Fallback: try to get from user object
		requestingUserInterface, userExists := c.Get("user")
		if !userExists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
			return
		}

		requestingUser, ok := requestingUserInterface.(*models.User)
		if !ok {
			utils.ErrorResponse(c, http.StatusInternalServerError, "invalid user type")
			return
		}

		// Only superadmin can change roles
		if !requestingUser.IsSuperAdmin() {
			utils.ErrorResponse(c, http.StatusForbidden, "only superadmin can change user roles")
			return
		}
		requestingUserID = requestingUser.ID
	} else {
		// Check role from string
		userRole, ok := userRoleInterface.(string)
		if !ok {
			utils.ErrorResponse(c, http.StatusInternalServerError, "invalid role type")
			return
		}

		// Only superadmin can change roles
		if models.Role(userRole) != models.RoleSuperAdmin {
			utils.ErrorResponse(c, http.StatusForbidden, "only superadmin can change user roles")
			return
		}
		
		// Get user ID from context
		userIDInterface, _ := c.Get("user_id")
		requestingUserID = userIDInterface.(uint)
	}

	// Get user ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	// Bind request
	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get user to update
	user, err := h.service.GetUserByID(ctx, uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}

	// Prevent superadmin from demoting themselves
	if user.ID == requestingUserID && req.Role != string(models.RoleSuperAdmin) {
		utils.ErrorResponse(c, http.StatusBadRequest, "cannot demote yourself")
		return
	}

	// Update role using service
	updatedUser, err := h.service.UpdateUserRole(ctx, user.ID, req.Role)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to update role")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "user role updated successfully",
		"user":    updatedUser,
	})
}

// GetMe godoc
// @Summary      Get own profile
// @Description  Get authenticated user's profile
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "User profile"
// @Failure      401  {object}  map[string]interface{}  "Unauthorized"
// @Failure      404  {object}  map[string]interface{}  "User not found"
// @Router       /users/me [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get user ID from context (set by AuthMiddleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID := userIDInterface.(uint)

	// Get user
	user, err := h.service.GetUserByID(ctx, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}

	utils.SuccessResponse(c, user)
}

// UpdateMe godoc
// @Summary      Update own profile
// @Description  Update authenticated user's profile (name, age, avatar, bio, phone)
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      models.UpdateProfileRequest  true  "Profile update request"
// @Success      200      {object}  map[string]interface{}       "Profile updated successfully"
// @Failure      400      {object}  map[string]interface{}       "Invalid request"
// @Failure      401      {object}  map[string]interface{}       "Unauthorized"
// @Failure      500      {object}  map[string]interface{}       "Internal server error"
// @Router       /users/me [put]
func (h *UserHandler) UpdateMe(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get user ID from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID := userIDInterface.(uint)

	// Bind and validate request
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Update profile
	user, err := h.service.UpdateProfile(ctx, userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to update profile")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "profile updated successfully",
		"user":    user,
	})
}

// ChangePassword godoc
// @Summary      Change password
// @Description  Change authenticated user's password
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      models.ChangePasswordRequest  true  "Password change request"
// @Success      200      {object}  map[string]interface{}        "Password changed successfully"
// @Failure      400      {object}  map[string]interface{}        "Invalid request or wrong password"
// @Failure      401      {object}  map[string]interface{}        "Unauthorized"
// @Failure      500      {object}  map[string]interface{}        "Internal server error"
// @Router       /users/me/password [put]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Get user ID from context
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID := userIDInterface.(uint)

	// Bind and validate request
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get user to verify current password
	user, err := h.service.GetUserByID(ctx, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "user not found")
		return
	}

	// Verify current password
	if err := auth.CheckPassword(req.CurrentPassword, user.Password); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "current password is incorrect")
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Change password
	if err := h.service.ChangePassword(ctx, userID, user.Password, hashedPassword); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to change password")
		return
	}

	utils.SuccessResponse(c, gin.H{
		"message": "password changed successfully",
	})
}
