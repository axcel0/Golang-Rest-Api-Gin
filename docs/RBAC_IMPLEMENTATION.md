# Role-Based Access Control (RBAC) Implementation

## Overview

This document describes the Role-Based Access Control (RBAC) implementation in the Go REST API project. The system implements a three-tier role hierarchy with granular permissions for different user operations.

## Role Hierarchy

```
┌─────────────────┐
│   SUPERADMIN    │  ← Full system access, can promote/demote users
└────────┬────────┘
         │
┌────────▼────────┐
│      ADMIN      │  ← Can manage users (CRUD operations)
└────────┬────────┘
         │
┌────────▼────────┐
│      USER       │  ← Can view users and own profile
└─────────────────┘
```

## Roles & Permissions

### 1. Superadmin (`superadmin`)
**Full Access** - Complete control over the system

**Permissions:**
- ✅ View all users
- ✅ Create users
- ✅ Update users
- ✅ Delete users
- ✅ **Change user roles** (exclusive permission)
- ✅ Access all protected endpoints

**Use Cases:**
- System administrator
- Initial setup and configuration
- Role management
- Security oversight

### 2. Admin (`admin`)
**Management Access** - Can manage users but not change roles

**Permissions:**
- ✅ View all users
- ✅ Create users
- ✅ Update users  
- ✅ Delete users
- ❌ Cannot change user roles
- ✅ Access all protected endpoints

**Use Cases:**
- User management
- Daily operations
- Content moderation
- Support team

### 3. User (`user`)
**Basic Access** - Read-only access with own profile management

**Permissions:**
- ✅ View all users (list)
- ✅ View user stats
- ✅ View individual user profiles
- ❌ Cannot create users
- ❌ Cannot update users
- ❌ Cannot delete users
- ❌ Cannot change roles
- ✅ Can view/update own profile

**Use Cases:**
- Regular application users
- Default role for new registrations
- Limited access to system functions

## Implementation Details

### Models

#### Role Type
```go
type Role string

const (
    RoleSuperAdmin Role = "superadmin"
    RoleAdmin      Role = "admin"
    RoleUser       Role = "user"
)
```

#### User Model
```go
type User struct {
    ID        uint
    Name      string
    Email     string
    Password  string `json:"-"`
    Age       int
    Role      string  // "superadmin", "admin", or "user"
    IsActive  bool
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt
}
```

#### Helper Methods
```go
func (u *User) HasRole(role Role) bool
func (u *User) IsSuperAdmin() bool
func (u *User) IsAdmin() bool
func (u *User) CanManageUsers() bool
func (u *User) CanDeleteUsers() bool
func (u *User) CanPromoteUsers() bool
```

### Middleware

#### JWTAuth Middleware
```go
func JWTAuth(jwtManager *auth.JWTManager, userRepo *repository.UserRepository) gin.HandlerFunc
```
- Validates JWT token
- Fetches user from database
- Checks if user is active
- Sets user object in context for RBAC checks

#### Role-Based Middleware

**RequireSuperAdmin()**
```go
func RequireSuperAdmin() gin.HandlerFunc
```
- Ensures user has `superadmin` role
- Returns 403 Forbidden if not superadmin

**RequireAdmin()**
```go
func RequireAdmin() gin.HandlerFunc
```
- Ensures user has `admin` or `superadmin` role
- Returns 403 Forbidden if neither role

**RequireUser()**
```go
func RequireUser() gin.HandlerFunc
```
- Ensures user is authenticated (any role)
- Returns 401 Unauthorized if not authenticated

**RequireRole(...roles)**
```go
func RequireRole(allowedRoles ...models.Role) gin.HandlerFunc
```
- Generic middleware for custom role combinations
- Accepts variable number of roles

### API Endpoints & Access Control

#### Health Endpoints (Public)
```
GET  /health                     [Public]
GET  /ready                      [Public]
```

#### Authentication Endpoints (Public/Protected)
```
POST /api/v1/auth/register       [Public]    - New users start as 'user' role
POST /api/v1/auth/login          [Public]
POST /api/v1/auth/refresh        [Public]
GET  /api/v1/auth/profile        [All]       - Any authenticated user
```

#### User Endpoints (Protected with RBAC)
```
GET    /api/v1/users                [All]           - List users
GET    /api/v1/users/stats          [All]           - User statistics
GET    /api/v1/users/:id            [All]           - Get user by ID
POST   /api/v1/users                [Admin+]        - Create user
POST   /api/v1/users/batch          [Admin+]        - Batch create users
PUT    /api/v1/users/:id            [Admin+]        - Update user
DELETE /api/v1/users/:id            [Admin+]        - Delete user
PUT    /api/v1/users/:id/role       [Superadmin]    - Change user role
```

**Legend:**
- `[Public]` - No authentication required
- `[All]` - Any authenticated user
- `[Admin+]` - Admin or Superadmin only
- `[Superadmin]` - Superadmin only

### Route Configuration

```go
// User routes with RBAC
users := v1.Group("/users")
users.Use(middleware.JWTAuth(jwtManager, userRepo))
{
    // Anyone authenticated can view
    users.GET("", userHandler.GetAllUsers)
    users.GET("/stats", userHandler.GetUserStats)
    users.GET("/:id", userHandler.GetUserByID)
    
    // Admin and superadmin only
    users.POST("", middleware.RequireAdmin(), userHandler.CreateUser)
    users.POST("/batch", middleware.RequireAdmin(), userHandler.BatchCreateUsers)
    users.PUT("/:id", middleware.RequireAdmin(), userHandler.UpdateUser)
    users.DELETE("/:id", middleware.RequireAdmin(), userHandler.DeleteUser)
    
    // Superadmin only
    users.PUT("/:id/role", middleware.RequireSuperAdmin(), userHandler.UpdateUserRole)
}
```

## Database Migration

### Migration: Add Role Column
```sql
-- 000003_add_role_to_users.up.sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'user' NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
UPDATE users SET role = 'user' WHERE role IS NULL OR role = '';
```

### Migration: Rollback
```sql
-- 000003_add_role_to_users.down.sql
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN IF EXISTS role;
```

## Usage Examples

### Register New User (Starts as 'user')
```bash
curl -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123",
    "age": 30
  }'

# Response: User created with role="user"
```

### Promote User to Admin (Superadmin Only)
```bash
curl -X PUT "http://localhost:8080/api/v1/users/2/role" \
  -H "Authorization: Bearer SUPERADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "admin"
  }'
```

### Create User as Admin
```bash
curl -X POST "http://localhost:8080/api/v1/users" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New User",
    "email": "newuser@example.com",
    "password": "password123",
    "age": 25
  }'
```

### Regular User Trying to Create User (Should Fail)
```bash
curl -X POST "http://localhost:8080/api/v1/users" \
  -H "Authorization: Bearer USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Another User",
    "email": "another@example.com",
    "password": "password123",
    "age": 28
  }'

# Response: 403 Forbidden
{
  "success": false,
  "message": "forbidden: insufficient permissions"
}
```

## Initial Setup

### Creating the First Superadmin

**Option 1: Database Seed (Recommended)**
```sql
-- Insert first superadmin
INSERT INTO users (name, email, password, age, role, is_active, created_at, updated_at)
VALUES (
  'System Admin',
  'admin@system.com',
  '$2a$10$hashed_password_here',  -- Use bcrypt to hash password
  30,
  'superadmin',
  true,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
);
```

**Option 2: Registration + Manual Promotion**
```bash
# 1. Register user normally
curl -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "System Admin",
    "email": "admin@system.com",
    "password": "securepassword",
    "age": 30
  }'

# 2. Manually update role in database
sqlite3 test.db "UPDATE users SET role = 'superadmin' WHERE email = 'admin@system.com';"

# 3. Login to get token with superadmin role
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@system.com",
    "password": "securepassword"
  }'
```

## Security Considerations

### 1. Role Assignment
- ✅ New registrations automatically assigned `user` role
- ✅ Only superadmin can change roles
- ✅ Users cannot escalate their own privileges
- ✅ Superadmin cannot demote themselves

### 2. Token Security
- ✅ JWT tokens include user ID and email
- ✅ Role is fetched from database on each request (not stored in JWT)
- ✅ Real-time role changes without token refresh
- ✅ Inactive users automatically blocked

### 3. Access Control
- ✅ Middleware-based enforcement
- ✅ Database-level role validation
- ✅ Context-based user information
- ✅ Comprehensive error messages

### 4. Audit Trail
- ✅ All role changes logged
- ✅ Structured logging with slog
- ✅ User ID and email in logs
- ✅ Action timestamps recorded

## Testing

### Automated RBAC Testing
```bash
# Run comprehensive RBAC test suite
./test_rbac.sh
```

The test script validates:
- ✅ User registration with default role
- ✅ Role promotions (superadmin only)
- ✅ READ access for all authenticated users
- ✅ CREATE access (admin and superadmin only)
- ✅ UPDATE access (admin and superadmin only)
- ✅ DELETE access (admin and superadmin only)
- ✅ ROLE CHANGE access (superadmin only)
- ✅ Forbidden responses for unauthorized access

### Manual Testing
```bash
# 1. Register three users
# 2. Promote second user to admin (via SQL)
# 3. Promote third user to superadmin (via SQL)
# 4. Test each endpoint with each role
# 5. Verify 403 responses for unauthorized actions
```

## Error Responses

### 401 Unauthorized
```json
{
  "success": false,
  "message": "authorization header required"
}
```

### 403 Forbidden
```json
{
  "success": false,
  "message": "forbidden: insufficient permissions"
}
```

### 403 Inactive Account
```json
{
  "success": false,
  "message": "account is inactive"
}
```

## Best Practices

### 1. Principle of Least Privilege
- Start users with minimum permissions (`user` role)
- Promote only when necessary
- Regular audit of role assignments

### 2. Role Naming Convention
- Use lowercase for database storage
- Clear, descriptive role names
- Consistent across the system

### 3. Middleware Ordering
```go
// Correct order
users.POST("", 
    middleware.JWTAuth(),      // 1. Authenticate
    middleware.RequireAdmin(), // 2. Check role
    handler)                    // 3. Execute

// Not
users.POST("", 
    middleware.RequireAdmin(), // ❌ User not set yet
    middleware.JWTAuth(),      // ❌ Wrong order
    handler)
```

### 4. Database Indexing
```sql
-- Index on role for fast queries
CREATE INDEX idx_users_role ON users(role);

-- Composite index for common queries
CREATE INDEX idx_users_role_active ON users(role, is_active);
```

## Future Enhancements

### Planned Features
- [ ] Permission-based access (more granular than roles)
- [ ] Role hierarchies with inheritance
- [ ] API key authentication for service accounts
- [ ] Role-based rate limiting
- [ ] Audit log for role changes
- [ ] UI for role management
- [ ] Bulk role assignment
- [ ] Time-based role assignments (temporary promotions)

### Advanced RBAC
- [ ] Resource-level permissions (e.g., own resources only)
- [ ] Department/team-based access control
- [ ] Dynamic role loading from database
- [ ] External identity provider integration (OAuth, SAML)

## References

- [OWASP RBAC Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html)
- [NIST RBAC Model](https://csrc.nist.gov/projects/role-based-access-control)
- JWT Best Practices
- Go Security Guidelines

## Summary

✅ **Three-tier role system**: Superadmin → Admin → User  
✅ **Middleware-based enforcement**: Clean separation of concerns  
✅ **Database-backed roles**: Real-time role changes  
✅ **Comprehensive testing**: Automated RBAC test suite  
✅ **Security-first design**: Cannot self-promote, inactive users blocked  
✅ **Production-ready**: Complete error handling and logging  

---

**Next Steps**: Run `./test_rbac.sh` to verify RBAC implementation! 🚀
