// Package services provides audit logging business logic with asynchronous
// event tracking, IP address extraction, and comprehensive action logging
// for security and compliance monitoring.
package services

import (
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"Go-Lang-project-01/pkg/logger"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditService handles audit logging business logic
type AuditService struct {
	repo *repository.AuditLogRepository
}

// NewAuditService creates a new audit service
func NewAuditService(repo *repository.AuditLogRepository) *AuditService {
	return &AuditService{repo: repo}
}

// LogAction creates an audit log entry asynchronously
func (s *AuditService) LogAction(c *gin.Context, userID *uint, action models.AuditAction, resource models.AuditResource, resourceID *uint, details interface{}, success bool, errorMsg string) {
	// Create audit log in goroutine to not block the request
	go func() {
		detailsJSON := ""
		if details != nil {
			if jsonBytes, err := json.Marshal(details); err == nil {
				detailsJSON = string(jsonBytes)
			}
		}

		log := &models.AuditLog{
			UserID:     userID,
			Action:     action,
			Resource:   resource,
			ResourceID: resourceID,
			Details:    detailsJSON,
			IPAddress:  s.getClientIP(c),
			UserAgent:  c.GetHeader("User-Agent"),
			Success:    success,
			ErrorMsg:   errorMsg,
			CreatedAt:  time.Now(),
		}

		if err := s.repo.Create(log); err != nil {
			logger.Error("Failed to create audit log", "error", err, "action", action)
		}
	}()
}

// LogAuthAction logs authentication-related actions
func (s *AuditService) LogAuthAction(c *gin.Context, userID *uint, action models.AuditAction, success bool, errorMsg string) {
	s.LogAction(c, userID, action, models.AuditResourceAuth, nil, nil, success, errorMsg)
}

// LogUserAction logs user management actions
func (s *AuditService) LogUserAction(c *gin.Context, userID uint, action models.AuditAction, targetUserID uint, details interface{}, success bool, errorMsg string) {
	s.LogAction(c, &userID, action, models.AuditResourceUser, &targetUserID, details, success, errorMsg)
}

// LogProfileAction logs profile-related actions
func (s *AuditService) LogProfileAction(c *gin.Context, userID uint, action models.AuditAction, details interface{}, success bool, errorMsg string) {
	s.LogAction(c, &userID, action, models.AuditResourceProfile, &userID, details, success, errorMsg)
}

// GetLogs retrieves audit logs with filters
func (s *AuditService) GetLogs(filter *repository.AuditLogFilter) ([]models.AuditLog, int64, error) {
	return s.repo.List(filter)
}

// GetLogByID retrieves a single audit log
func (s *AuditService) GetLogByID(id uint) (*models.AuditLog, error) {
	return s.repo.GetByID(id)
}

// GetRecentByUser retrieves recent logs for a user
func (s *AuditService) GetRecentByUser(userID uint, limit int) ([]models.AuditLog, error) {
	return s.repo.GetRecentByUser(userID, limit)
}

// GetFailedLoginAttempts gets failed login count from an IP
func (s *AuditService) GetFailedLoginAttempts(ipAddress string, since time.Time) (int64, error) {
	return s.repo.GetFailedLoginAttempts(ipAddress, since)
}

// GetStats retrieves audit log statistics
func (s *AuditService) GetStats() (map[string]interface{}, error) {
	return s.repo.GetStats()
}

// CleanupOldLogs deletes logs older than the retention period
func (s *AuditService) CleanupOldLogs(retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	deleted, err := s.repo.DeleteOlderThan(cutoffDate)
	if err != nil {
		logger.Error("Failed to cleanup old audit logs", "error", err)
		return 0, err
	}
	logger.Info("Cleaned up old audit logs", "deleted", deleted, "cutoff_date", cutoffDate)
	return deleted, nil
}

// getClientIP extracts the real client IP from the request
func (s *AuditService) getClientIP(c *gin.Context) string {
	// Try X-Forwarded-For header first (for proxies/load balancers)
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		return xff
	}

	// Try X-Real-IP header
	xri := c.GetHeader("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}
