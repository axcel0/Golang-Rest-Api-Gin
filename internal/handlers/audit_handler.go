package handlers

import (
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"Go-Lang-project-01/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditHandler handles audit log HTTP requests
type AuditHandler struct {
	service *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(service *services.AuditService) *AuditHandler {
	return &AuditHandler{service: service}
}

// GetAuditLogs godoc
// @Summary      Get audit logs
// @Description  Retrieve audit logs with optional filters (admin only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        user_id      query  int     false  "Filter by user ID"
// @Param        action       query  string  false  "Filter by action"
// @Param        resource     query  string  false  "Filter by resource"
// @Param        success      query  bool    false  "Filter by success status"
// @Param        start_date   query  string  false  "Start date (RFC3339)"
// @Param        end_date     query  string  false  "End date (RFC3339)"
// @Param        page         query  int     false  "Page number (default: 1)"
// @Param        page_size    query  int     false  "Page size (default: 20, max: 100)"
// @Security     Bearer
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /audit-logs [get]
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	filter := &repository.AuditLogFilter{
		Page:     1,
		PageSize: 20,
	}

	// Parse query parameters
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uid := uint(userID)
			filter.UserID = &uid
		}
	}

	if action := c.Query("action"); action != "" {
		act := models.AuditAction(action)
		filter.Action = &act
	}

	if resource := c.Query("resource"); resource != "" {
		res := models.AuditResource(resource)
		filter.Resource = &res
	}

	if successStr := c.Query("success"); successStr != "" {
		if success, err := strconv.ParseBool(successStr); err == nil {
			filter.Success = &success
		}
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			filter.PageSize = pageSize
		}
	}

	logs, total, err := h.service.GetLogs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to retrieve audit logs",
			"error":   err.Error(),
		})
		return
	}

	totalPages := (int(total) + filter.PageSize - 1) / filter.PageSize

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"pagination": gin.H{
			"page":        filter.Page,
			"page_size":   filter.PageSize,
			"total_items": total,
			"total_pages": totalPages,
		},
	})
}

// GetAuditLog godoc
// @Summary      Get single audit log
// @Description  Retrieve a specific audit log by ID (admin only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        id   path  int  true  "Audit Log ID"
// @Security     Bearer
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /audit-logs/{id} [get]
func (h *AuditHandler) GetAuditLog(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid audit log ID",
		})
		return
	}

	log, err := h.service.GetLogByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Audit log not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    log,
	})
}

// GetMyAuditLogs godoc
// @Summary      Get my audit logs
// @Description  Retrieve audit logs for the authenticated user
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        limit   query  int  false  "Limit (default: 50)"
// @Security     Bearer
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /audit-logs/me [get]
func (h *AuditHandler) GetMyAuditLogs(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	userID := userIDInterface.(uint)
	limit := 50

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	logs, err := h.service.GetRecentByUser(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to retrieve audit logs",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"count":   len(logs),
	})
}

// GetAuditStats godoc
// @Summary      Get audit statistics
// @Description  Retrieve audit log statistics (admin only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /audit-logs/stats [get]
func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to retrieve audit statistics",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// CleanupOldLogs godoc
// @Summary      Cleanup old audit logs
// @Description  Delete audit logs older than specified days (admin only)
// @Tags         audit
// @Accept       json
// @Produce      json
// @Param        days  query  int  true  "Retention days (logs older than this will be deleted)"
// @Security     Bearer
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /audit-logs/cleanup [delete]
func (h *AuditHandler) CleanupOldLogs(c *gin.Context) {
	daysStr := c.Query("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid days parameter (must be positive integer)",
		})
		return
	}

	deleted, err := h.service.CleanupOldLogs(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to cleanup old logs",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Old audit logs cleaned up successfully",
		"deleted": deleted,
	})
}
