# Activity Logging & Audit Trail

## Overview

The audit logging system provides a comprehensive trail of all user actions and system events. It automatically tracks authentication attempts, user management operations, profile changes, and administrative actions with detailed context including IP addresses, user agents, and timestamps.

**Status**: ✅ Fully Implemented & Tested

**Version**: 1.0.0

**Last Updated**: October 2025

---

## Features

### ✅ Comprehensive Action Tracking
- Authentication events (login, logout, registration, token refresh)
- User CRUD operations (create, read, update, delete)
- Profile management (profile updates, password changes)
- Role modifications (role changes tracked with context)
- Failed login attempts (security monitoring)

###  ✅ Detailed Context Capture
- User ID (who performed the action)
- Action type (what was done)
- Resource type (what was affected)
- Resource ID (specific entity affected)
- IP Address (where from - IPv4/IPv6 support)
- User Agent (client information)
- Success status (true/false)
- Error messages (for failed actions)
- Timestamp (when)

### ✅ Flexible Query API
- Filter by user, action, resource, success status, date range
- Pagination support (configurable page size)
- User-specific audit logs (view own history)
- Admin statistics and analytics
- Failed login monitoring
- Automatic cleanup of old logs

### ✅ Async Logging
- Non-blocking audit log creation
- Doesn't impact request performance
- Goroutine-based asynchronous writes

---

## Database Schema

### AuditLog Model

```go
type AuditLog struct {
    ID          uint          `json:"id"`
    UserID      *uint         `json:"user_id,omitempty"`      // Nullable for failed logins
    Action      AuditAction   `json:"action"`                 // login, user_create, etc.
    Resource    AuditResource `json:"resource"`               // auth, user, profile, system
    ResourceID  *uint         `json:"resource_id,omitempty"`  // ID of affected resource
    Details     string        `json:"details,omitempty"`      // JSON details
    IPAddress   string        `json:"ip_address"`             // Client IP (IPv4/IPv6)
    UserAgent   string        `json:"user_agent,omitempty"`   // Browser/client info
    Success     bool          `json:"success"`                // true/false
    ErrorMsg    string        `json:"error_message,omitempty"`// Error if failed
    CreatedAt   time.Time     `json:"created_at"`             // Timestamp
}
```

### Action Types

**Authentication Actions:**
- `login` - Successful user login
- `login_failed` - Failed login attempt
- `logout` - User logout
- `refresh_token` - Token refresh operation
- `register` - New user registration

**User Management Actions:**
- `user_create` - New user created by admin
- `user_read` - User data accessed
- `user_update` - User information updated
- `user_delete` - User deleted
- `user_batch_create` - Multiple users created

**Profile Actions:**
- `profile_update` - User profile modified
- `password_change` - Password changed

**Administrative Actions:**
- `role_change` - User role modified
- `system_access` - System-level access

### Resource Types

- `auth` - Authentication/authorization
- `user` - User management
- `profile` - Profile management
- `system` - System operations

---

## API Endpoints

### 1. Get My Audit Logs (User)

**Endpoint**: `GET /api/v1/audit-logs/me`

**Authentication**: Required (JWT Bearer token)

**Authorization**: Any authenticated user

**Description**: Retrieve audit logs for the authenticated user

**Query Parameters**:
- `limit` (optional): Number of logs to return (default: 50, max: 100)

**Example Request**:
```bash
curl -X GET "http://localhost:8080/api/v1/audit-logs/me?limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Example Response**:
```json
{
  "success": true,
  "count": 3,
  "data": [
    {
      "id": 3,
      "user_id": 25,
      "action": "login",
      "resource": "auth",
      "ip_address": "::1",
      "user_agent": "curl/8.14.1",
      "success": true,
      "created_at": "2025-10-31T17:30:31.933911271+07:00"
    },
    {
      "id": 2,
      "user_id": 25,
      "action": "login",
      "resource": "auth",
      "ip_address": "::1",
      "user_agent": "curl/8.14.1",
      "success": true,
      "created_at": "2025-10-31T17:29:09.465953436+07:00"
    },
    {
      "id": 1,
      "user_id": 25,
      "action": "register",
      "resource": "auth",
      "ip_address": "::1",
      "user_agent": "curl/8.14.1",
      "success": true,
      "created_at": "2025-10-31T17:17:51.687370848+07:00"
    }
  ]
}
```

---

### 2. Get All Audit Logs (Admin)

**Endpoint**: `GET /api/v1/audit-logs`

**Authentication**: Required (JWT Bearer token)

**Authorization**: Admin or Superadmin only

**Description**: Retrieve all audit logs with optional filters

**Query Parameters**:
- `user_id` (optional): Filter by specific user ID
- `action` (optional): Filter by action type (e.g., "login", "user_create")
- `resource` (optional): Filter by resource type (e.g., "auth", "user")
- `success` (optional): Filter by success status (true/false)
- `start_date` (optional): Start date (RFC3339 format: 2025-10-01T00:00:00Z)
- `end_date` (optional): End date (RFC3339 format)
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 20, max: 100)

**Example Requests**:

```bash
# Get all logs (first page)
curl -X GET "http://localhost:8080/api/v1/audit-logs" \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Filter by specific user
curl -X GET "http://localhost:8080/api/v1/audit-logs?user_id=25" \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Filter by action type
curl -X GET "http://localhost:8080/api/v1/audit-logs?action=login_failed" \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Filter by date range
curl -X GET "http://localhost:8080/api/v1/audit-logs?start_date=2025-10-01T00:00:00Z&end_date=2025-10-31T23:59:59Z" \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Pagination
curl -X GET "http://localhost:8080/api/v1/audit-logs?page=2&page_size=50" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Example Response**:
```json
{
  "success": true,
  "data": [
    {
      "id": 5,
      "user_id": 1,
      "action": "user_create",
      "resource": "user",
      "resource_id": 26,
      "details": "{\"name\":\"New User\",\"email\":\"new@example.com\"}",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "success": true,
      "created_at": "2025-10-31T15:30:00Z"
    },
    {
      "id": 4,
      "user_id": null,
      "action": "login_failed",
      "resource": "auth",
      "ip_address": "192.168.1.50",
      "user_agent": "curl/7.81.0",
      "success": false,
      "error_message": "Invalid password",
      "created_at": "2025-10-31T14:20:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 2,
    "total_pages": 1
  }
}
```

---

### 3. Get Single Audit Log (Admin)

**Endpoint**: `GET /api/v1/audit-logs/:id`

**Authentication**: Required (JWT Bearer token)

**Authorization**: Admin or Superadmin only

**Description**: Retrieve a specific audit log by ID

**Example Request**:
```bash
curl -X GET "http://localhost:8080/api/v1/audit-logs/5" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Example Response**:
```json
{
  "success": true,
  "data": {
    "id": 5,
    "user_id": 1,
    "action": "role_change",
    "resource": "user",
    "resource_id": 25,
    "details": "{\"old_role\":\"user\",\"new_role\":\"admin\",\"changed_by\":1}",
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "success": true,
    "created_at": "2025-10-31T16:00:00Z"
  }
}
```

---

### 4. Get Audit Statistics (Admin)

**Endpoint**: `GET /api/v1/audit-logs/stats`

**Authentication**: Required (JWT Bearer token)

**Authorization**: Admin or Superadmin only

**Description**: Retrieve audit log statistics and analytics

**Example Request**:
```bash
curl -X GET "http://localhost:8080/api/v1/audit-logs/stats" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Example Response**:
```json
{
  "success": true,
  "data": {
    "total_logs": 1523,
    "by_action": [
      {"action": "login", "count": 450},
      {"action": "user_read", "count": 320},
      {"action": "profile_update", "count": 215},
      {"action": "login_failed", "count": 45},
      {"action": "user_create", "count": 38}
    ],
    "failed_last_24h": 12,
    "most_active_users": [
      {"user_id": 1, "count": 156},
      {"user_id": 5, "count": 98},
      {"user_id": 12, "count": 87}
    ]
  }
}
```

---

### 5. Cleanup Old Logs (Admin)

**Endpoint**: `DELETE /api/v1/audit-logs/cleanup`

**Authentication**: Required (JWT Bearer token)

**Authorization**: Admin or Superadmin only

**Description**: Delete audit logs older than specified retention period

**Query Parameters**:
- `days` (required): Retention period in days (logs older than this will be deleted)

**Example Request**:
```bash
# Delete logs older than 90 days
curl -X DELETE "http://localhost:8080/api/v1/audit-logs/cleanup?days=90" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Example Response**:
```json
{
  "success": true,
  "message": "Old audit logs cleaned up successfully",
  "deleted": 234
}
```

---

## Usage Examples

### Automatically Logged Actions

The audit service is integrated into all authentication and user management handlers:

#### 1. Registration
```go
// Automatically logs successful registration
// Action: register
// Resource: auth
// Success: true
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com","password":"pass123","age":30}'
```

#### 2. Login (Success)
```go
// Automatically logs successful login
// Action: login
// Resource: auth
// Success: true
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"pass123"}'
```

#### 3. Login (Failed)
```go
// Automatically logs failed login attempt
// Action: login_failed
// Resource: auth
// Success: false
// Error: "Invalid password" or "User not found" or "Account inactive"
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","password":"wrong"}'
```

#### 4. Token Refresh
```go
// Automatically logs token refresh
// Action: refresh_token
// Resource: auth
// Success: true
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

---

## Integration with Services

### Manual Audit Logging

You can manually log actions in your services:

```go
// In your service/handler
func (s *UserService) UpdateUser(c *gin.Context, id uint, updates *UpdateRequest) error {
    // Perform update
    user, err := s.repo.Update(id, updates)
    if err != nil {
        // Log failed action
        s.auditService.LogUserAction(c, currentUserID, models.AuditActionUserUpdate, id, updates, false, err.Error())
        return err
    }
    
    // Log successful action
    s.auditService.LogUserAction(c, currentUserID, models.AuditActionUserUpdate, id, updates, true, "")
    return nil
}
```

### Audit Service Methods

```go
type AuditService interface {
    // General logging
    LogAction(c *gin.Context, userID *uint, action AuditAction, resource AuditResource, resourceID *uint, details interface{}, success bool, errorMsg string)
    
    // Convenience methods
    LogAuthAction(c *gin.Context, userID *uint, action AuditAction, success bool, errorMsg string)
    LogUserAction(c *gin.Context, userID uint, action AuditAction, targetUserID uint, details interface{}, success bool, errorMsg string)
    LogProfileAction(c *gin.Context, userID uint, action AuditAction, details interface{}, success bool, errorMsg string)
    
    // Query methods
    GetLogs(filter *AuditLogFilter) ([]AuditLog, int64, error)
    GetLogByID(id uint) (*AuditLog, error)
    GetRecentByUser(userID uint, limit int) ([]AuditLog, error)
    GetFailedLoginAttempts(ipAddress string, since time.Time) (int64, error)
    GetStats() (map[string]interface{}, error)
    
    // Maintenance
    CleanupOldLogs(retentionDays int) (int64, error)
}
```

---

## Security Features

### 1. Failed Login Monitoring

Track and monitor failed login attempts for security analysis:

```go
// Get failed login attempts from specific IP in last hour
failedCount, err := auditService.GetFailedLoginAttempts("192.168.1.50", time.Now().Add(-1*time.Hour))
if failedCount > 5 {
    // Trigger security alert or rate limiting
}
```

### 2. IP Address Tracking

All actions are logged with the client's real IP address, considering proxies:

- Checks `X-Forwarded-For` header (for load balancers)
- Falls back to `X-Real-IP` header
- Uses `RemoteAddr` as last resort

### 3. User Agent Tracking

Identifies the client application/browser used for each action, useful for:
- Detecting automated attacks
- Identifying compromised accounts (unusual clients)
- Security analytics

### 4. Nullable User ID

For failed login attempts where the user doesn't exist or can't be identified, `user_id` is nullable, preventing information disclosure.

---

## Performance Considerations

### Asynchronous Logging

All audit logging is performed asynchronously to prevent blocking the main request:

```go
func (s *AuditService) LogAction(...) {
    // Create audit log in goroutine to not block the request
    go func() {
        log := &models.AuditLog{...}
        if err := s.repo.Create(log); err != nil {
            logger.Error("Failed to create audit log", "error", err)
        }
    }()
}
```

**Benefits**:
- Zero impact on request latency
- Fire-and-forget pattern
- Failures logged but don't affect user experience

### Database Indexes

The audit_logs table has indexes on:
- `user_id` - Fast user-specific queries
- `action` - Fast action-type filtering
- `resource` - Fast resource-type filtering
- `success` - Fast success/failure filtering
- `created_at` - Fast date range queries

### Pagination

All list endpoints support pagination to prevent memory issues with large datasets.

---

## Maintenance

### Log Retention

Implement a retention policy to manage database growth:

```bash
# Manual cleanup - delete logs older than 90 days
curl -X DELETE "http://localhost:8080/api/v1/audit-logs/cleanup?days=90" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

**Recommended Policies**:
- **Development**: 30 days
- **Staging**: 60 days
- **Production**: 90-365 days (based on compliance requirements)

### Automated Cleanup

Set up a cron job or scheduled task:

```bash
# Daily cleanup at 2 AM
0 2 * * * curl -X DELETE "http://localhost:8080/api/v1/audit-logs/cleanup?days=90" -H "Authorization: Bearer $ADMIN_TOKEN"
```

Or implement in-application scheduler:

```go
// In main.go
go func() {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        deleted, err := auditService.CleanupOldLogs(90) // 90 days retention
        if err != nil {
            logger.Error("Failed to cleanup old audit logs", "error", err)
        } else {
            logger.Info("Cleaned up old audit logs", "deleted", deleted)
        }
    }
}()
```

---

## Compliance & Auditing

### GDPR Compliance

For GDPR compliance, implement user data export and deletion:

```go
// Export user's audit logs (for GDPR data portability)
func ExportUserAuditLogs(userID uint) ([]byte, error) {
    logs, _, err := auditService.GetLogs(&AuditLogFilter{
        UserID: &userID,
        PageSize: 10000, // All logs
    })
    return json.Marshal(logs)
}

// Delete user's audit logs (for GDPR right to erasure)
func DeleteUserAuditLogs(userID uint) error {
    return db.Where("user_id = ?", userID).Delete(&AuditLog{}).Error
}
```

### SOC 2 Compliance

Audit logs support SOC 2 compliance requirements:
- **Access Logging**: All authentication attempts tracked
- **Change Tracking**: All modifications logged with who, what, when
- **Failed Access**: Failed login attempts monitored
- **Retention**: Configurable retention policies
- **Integrity**: Immutable logs (no update/delete for users)

### PCI DSS Compliance

For PCI DSS compliance:
- Track all access to cardholder data (implement custom logging)
- Monitor failed access attempts
- Retain audit trail for at least 1 year
- Implement log review processes

---

## Troubleshooting

### No Logs Appearing

**Problem**: Actions are performed but no audit logs created

**Solutions**:
1. Check audit service initialization in main.go
2. Verify database migration included AuditLog model
3. Check server logs for audit log errors
4. Ensure audit service is passed to handlers

### Performance Degradation

**Problem**: High volume of audit logs affecting performance

**Solutions**:
1. Verify async logging is enabled (goroutine-based)
2. Add database indexes if missing
3. Implement regular cleanup (reduce table size)
4. Consider archiving old logs to separate table/database

### Disk Space Issues

**Problem**: Audit logs consuming too much disk space

**Solutions**:
1. Implement retention policy (cleanup old logs)
2. Archive old logs to cold storage
3. Compress details field (use gzip)
4. Reduce details verbosity

---

## Testing

### Test Coverage

```bash
# Run audit-related tests
go test ./tests/integration/... -v -run TestAudit
```

### Manual Testing Checklist

- [ ] ✅ User registration creates audit log
- [ ] ✅ Successful login creates audit log
- [ ] ✅ Failed login creates audit log with error
- [ ] ✅ Token refresh creates audit log
- [ ] ✅ User can view own audit logs
- [ ] ✅ Non-admin cannot access admin endpoints
- [ ] ✅ Admin can view all audit logs
- [ ] ✅ Admin can filter by user/action/date
- [ ] ✅ Admin can view statistics
- [ ] ✅ Admin can cleanup old logs
- [ ] ✅ IP address captured correctly
- [ ] ✅ User agent captured correctly
- [ ] ✅ Pagination works correctly

---

## Future Enhancements

### Phase 2 (Planned)

- [ ] **Real-time Alerts**: Notify admins of suspicious activity
- [ ] **Anomaly Detection**: ML-based unusual pattern detection
- [ ] **Export Functionality**: CSV/PDF export for compliance
- [ ] **Advanced Search**: Full-text search in details
- [ ] **Audit Log Visualization**: Dashboards and charts
- [ ] **Webhook Integration**: Send audit events to external systems

### Phase 3 (Future)

- [ ] **Blockchain Integration**: Immutable audit trail
- [ ] **Multi-tenancy Support**: Separate logs per tenant
- [ ] **Log Signing**: Cryptographic proof of authenticity
- [ ] **Log Forwarding**: Send to SIEM systems (Splunk, ELK)
- [ ] **Compliance Reports**: Automated SOC2/PCI-DSS reports

---

## Related Documentation

- [Authentication & JWT](./authentication.md)
- [User Management](./user-management.md)
- [RBAC & Permissions](./rbac.md)
- [Security Best Practices](./security.md)

---

## API Summary

| Endpoint | Method | Auth | Role | Description |
|----------|--------|------|------|-------------|
| `/api/v1/audit-logs/me` | GET | Yes | User | Get my audit logs |
| `/api/v1/audit-logs` | GET | Yes | Admin | Get all logs (with filters) |
| `/api/v1/audit-logs/:id` | GET | Yes | Admin | Get specific log |
| `/api/v1/audit-logs/stats` | GET | Yes | Admin | Get statistics |
| `/api/v1/audit-logs/cleanup` | DELETE | Yes | Admin | Cleanup old logs |

---

**Documentation Version**: 1.0.0  
**Last Updated**: October 31, 2025  
**Status**: Production Ready ✅
