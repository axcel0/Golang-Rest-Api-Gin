package models

import (
	"time"
)

// AuditAction represents the type of action performed
type AuditAction string

const (
	// Authentication actions
	AuditActionLogin         AuditAction = "login"
	AuditActionLoginFailed   AuditAction = "login_failed"
	AuditActionLogout        AuditAction = "logout"
	AuditActionRefreshToken  AuditAction = "refresh_token"
	AuditActionRegister      AuditAction = "register"

	// User CRUD actions
	AuditActionUserCreate    AuditAction = "user_create"
	AuditActionUserRead      AuditAction = "user_read"
	AuditActionUserUpdate    AuditAction = "user_update"
	AuditActionUserDelete    AuditAction = "user_delete"
	AuditActionUserBatchCreate AuditAction = "user_batch_create"

	// Profile actions
	AuditActionProfileUpdate   AuditAction = "profile_update"
	AuditActionPasswordChange  AuditAction = "password_change"

	// Role management
	AuditActionRoleChange    AuditAction = "role_change"

	// System actions
	AuditActionSystemAccess  AuditAction = "system_access"
)

// AuditResource represents the resource being accessed
type AuditResource string

const (
	AuditResourceAuth    AuditResource = "auth"
	AuditResourceUser    AuditResource = "user"
	AuditResourceProfile AuditResource = "profile"
	AuditResourceSystem  AuditResource = "system"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	UserID      *uint         `gorm:"index" json:"user_id,omitempty"` // Nullable for failed logins
	Action      AuditAction   `gorm:"type:varchar(50);index" json:"action"`
	Resource    AuditResource `gorm:"type:varchar(50);index" json:"resource"`
	ResourceID  *uint         `gorm:"index" json:"resource_id,omitempty"` // ID of affected resource
	Details     string        `gorm:"type:text" json:"details,omitempty"` // JSON details
	IPAddress   string        `gorm:"type:varchar(45)" json:"ip_address"` // IPv4 or IPv6
	UserAgent   string        `gorm:"type:text" json:"user_agent,omitempty"`
	Success     bool          `gorm:"default:true;index" json:"success"`
	ErrorMsg    string        `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt   time.Time     `gorm:"index" json:"created_at"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}
