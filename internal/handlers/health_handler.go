package handlers

import (
	"context"
	"net/http"
	"time"

	"Go-Lang-project-01/internal/health"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	healthService *health.HealthService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(healthService *health.HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// HealthCheck godoc
// @Summary      Enhanced health check
// @Description  Check service health including all components (database, disk, memory)
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  health.HealthResponse
// @Failure      503  {object}  health.HealthResponse
// @Router       /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	healthResp := h.healthService.CheckHealth(ctx)

	statusCode := http.StatusOK
	if healthResp.Status == health.StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if healthResp.Status == health.StatusDegraded {
		statusCode = http.StatusOK // Still return 200 for degraded
	}

	c.JSON(statusCode, healthResp)
}

// ReadinessCheck godoc
// @Summary      Readiness check
// @Description  Simple readiness probe for Kubernetes/Docker (lightweight check)
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Simple check - service is running
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"service": "Go-Lang-project-01",
	})
}
