package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

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

// GetAllUsers handles GET /api/v1/users with pagination
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

// GetUserByID handles GET /api/users/:id
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

// CreateUser handles POST /api/users
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

// UpdateUser handles PUT /api/users/:id
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

// DeleteUser handles DELETE /api/users/:id
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

// BatchCreateUsers handles POST /api/users/batch
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

// GetUserStats handles GET /api/users/stats
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
