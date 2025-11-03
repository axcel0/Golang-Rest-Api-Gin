package repository

import (
	"Go-Lang-project-01/internal/models"
	"time"

	"gorm.io/gorm"
)

// AuditLogRepository handles database operations for audit logs
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditLogRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

// AuditLogFilter represents filter options for querying audit logs
type AuditLogFilter struct {
	UserID     *uint
	Action     *models.AuditAction
	Resource   *models.AuditResource
	Success    *bool
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
}

// List retrieves audit logs with optional filters
func (r *AuditLogRepository) List(filter *AuditLogFilter) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := r.db.Model(&models.AuditLog{})

	// Apply filters
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}
	if filter.Resource != nil {
		query = query.Where("resource = ?", *filter.Resource)
	}
	if filter.Success != nil {
		query = query.Where("success = ?", *filter.Success)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Default page size
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByID retrieves a single audit log by ID
func (r *AuditLogRepository) GetByID(id uint) (*models.AuditLog, error) {
	var log models.AuditLog
	if err := r.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// GetRecentByUser retrieves recent audit logs for a specific user
func (r *AuditLogRepository) GetRecentByUser(userID uint, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetFailedLoginAttempts retrieves failed login attempts within a time window
func (r *AuditLogRepository) GetFailedLoginAttempts(ipAddress string, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.AuditLog{}).
		Where("action = ? AND ip_address = ? AND success = ? AND created_at >= ?",
			models.AuditActionLoginFailed, ipAddress, false, since).
		Count(&count).Error
	return count, err
}

// DeleteOlderThan deletes audit logs older than the specified date
func (r *AuditLogRepository) DeleteOlderThan(date time.Time) (int64, error) {
	result := r.db.Where("created_at < ?", date).Delete(&models.AuditLog{})
	return result.RowsAffected, result.Error
}

// GetStats retrieves audit log statistics
func (r *AuditLogRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total logs
	var total int64
	if err := r.db.Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total_logs"] = total

	// Logs by action
	var actionStats []struct {
		Action string
		Count  int64
	}
	if err := r.db.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Order("count DESC").
		Limit(10).
		Scan(&actionStats).Error; err != nil {
		return nil, err
	}
	stats["by_action"] = actionStats

	// Failed actions in last 24 hours
	var failedCount int64
	if err := r.db.Model(&models.AuditLog{}).
		Where("success = ? AND created_at >= ?", false, time.Now().Add(-24*time.Hour)).
		Count(&failedCount).Error; err != nil {
		return nil, err
	}
	stats["failed_last_24h"] = failedCount

	// Most active users (last 7 days)
	var activeUsers []struct {
		UserID uint
		Count  int64
	}
	if err := r.db.Model(&models.AuditLog{}).
		Select("user_id, COUNT(*) as count").
		Where("user_id IS NOT NULL AND created_at >= ?", time.Now().Add(-7*24*time.Hour)).
		Group("user_id").
		Order("count DESC").
		Limit(5).
		Scan(&activeUsers).Error; err != nil {
		return nil, err
	}
	stats["most_active_users"] = activeUsers

	return stats, nil
}
