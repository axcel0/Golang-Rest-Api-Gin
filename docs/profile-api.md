# User Profile Management API

Self-service user profile management endpoints. Users can view and update their own profile without admin privileges.

## Endpoints

### 1. Get My Profile

Retrieve authenticated user's profile information.

**Endpoint:** `GET /api/v1/users/me`

**Authentication:** Required (JWT Bearer token)

**Request:**
```bash
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 20,
    "name": "John Doe",
    "email": "john@example.com",
    "age": 28,
    "role": "user",
    "is_active": true,
    "avatar_url": "https://example.com/avatar.jpg",
    "bio": "Software Engineer from Indonesia",
    "phone_number": "+628123456789",
    "created_at": "2025-10-31T15:31:41.143819195+07:00",
    "updated_at": "2025-10-31T15:31:41.231974549+07:00"
  }
}
```

**Error Response (401):**
```json
{
  "success": false,
  "message": "invalid or expired token"
}
```

---

### 2. Update My Profile

Update authenticated user's profile. All fields are optional - only send fields you want to update.

**Endpoint:** `PUT /api/v1/users/me`

**Authentication:** Required (JWT Bearer token)

**Request Body:**
```json
{
  "name": "Updated Name",           // Optional: 2-100 chars
  "age": 30,                         // Optional: 1-150
  "avatar_url": "https://example.com/new-avatar.jpg",  // Optional: valid URL
  "bio": "Full stack developer",    // Optional: max 500 chars
  "phone_number": "+628987654321"   // Optional: 10-20 chars
}
```

**Request Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Name",
    "avatar_url": "https://example.com/avatar.jpg",
    "bio": "Software Engineer",
    "phone_number": "+628123456789"
  }'
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "profile updated successfully",
    "user": {
      "id": 20,
      "name": "Updated Name",
      "email": "john@example.com",
      "age": 28,
      "role": "user",
      "is_active": true,
      "avatar_url": "https://example.com/avatar.jpg",
      "bio": "Software Engineer",
      "phone_number": "+628123456789",
      "created_at": "2025-10-31T15:31:41.143819195+07:00",
      "updated_at": "2025-10-31T15:31:41.231974549+07:00"
    }
  }
}
```

**Validation Errors (400):**
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": [
    {
      "field": "avatarurl",
      "message": "avatarurl is invalid"
    }
  ]
}
```

**Field Validation Rules:**
- `name`: min 2, max 100 characters
- `age`: min 1, max 150
- `avatar_url`: must be valid URL format (https://...)
- `bio`: max 500 characters
- `phone_number`: min 10, max 20 characters (supports international formats)

**Partial Updates:**
Only fields provided in request body will be updated. Omitted fields keep their current values.

Example - update only name:
```bash
curl -X PUT http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "New Name Only"}'
```

---

### 3. Change Password

Change authenticated user's password. Requires current password verification.

**Endpoint:** `PUT /api/v1/users/me/password`

**Authentication:** Required (JWT Bearer token)

**Request Body:**
```json
{
  "current_password": "oldpassword123",  // Required: min 6 chars
  "new_password": "newpassword456"       // Required: min 6, max 100 chars
}
```

**Request Example:**
```bash
curl -X PUT http://localhost:8080/api/v1/users/me/password \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "oldpass123",
    "new_password": "newpass456"
  }'
```

**Success Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "password changed successfully"
  }
}
```

**Wrong Current Password (400):**
```json
{
  "success": false,
  "message": "current password is incorrect"
}
```

**Validation Errors (400):**
```json
{
  "success": false,
  "message": "Validation failed",
  "errors": [
    {
      "field": "newpassword",
      "message": "newpassword must be at least 6 characters"
    }
  ]
}
```

**Security Notes:**
- Current password must match user's existing password
- Password is hashed using bcrypt before storage
- After changing password, existing JWT tokens remain valid until expiration
- Users should logout and login again to get fresh tokens

---

## Testing Results

### Comprehensive Test Suite (12 Scenarios)

✅ **Core Functionality (7 tests):**
1. Register new user - SUCCESS
2. JWT contains role field - VERIFIED
3. GET /users/me retrieves profile - SUCCESS
4. PUT /users/me updates profile - SUCCESS
5. GET /users/me verifies persistence - SUCCESS
6. PUT /users/me/password changes password - SUCCESS
7. Login with new password - SUCCESS

✅ **Validation Tests (5 tests):**
8. Invalid URL rejected - CORRECT (400)
9. Bio >500 chars rejected - CORRECT (400)
10. Phone <10 chars rejected - CORRECT (400)
11. Partial update (name only) - SUCCESS
12. Wrong current password rejected - CORRECT (400)

### Test Execution Summary
- **Total Tests:** 12
- **Passed:** 12
- **Failed:** 0
- **Success Rate:** 100%
- **Execution Time:** ~3 seconds

---

## Database Schema

New columns added to `users` table:

```sql
ALTER TABLE `users` ADD `avatar_url` varchar(255);
ALTER TABLE `users` ADD `bio` text;
ALTER TABLE `users` ADD `phone_number` varchar(20);
```

**Migration Time:** 13.1ms (8.3ms + 2.2ms + 2.5ms)

---

## Error Codes

| Code | Message | Description |
|------|---------|-------------|
| 200 | Success | Operation completed successfully |
| 400 | Validation failed | Request validation error (see errors array) |
| 400 | current password is incorrect | Wrong password provided |
| 401 | invalid or expired token | Authentication failed |
| 404 | user not found | User ID in token doesn't exist |
| 500 | internal server error | Server error (check logs) |

---

## Implementation Details

**Route Registration Order:**
```go
users.GET("/me", userHandler.GetMe)              // Before /:id
users.PUT("/me", userHandler.UpdateMe)           // Before /:id  
users.PUT("/me/password", userHandler.ChangePassword)
users.GET("/:id", userHandler.GetUserByID)       // After /me routes
```

**Important:** `/me` routes must be registered BEFORE `/:id` to avoid path collision (Gin matches first pattern).

**Middleware Stack:**
1. JWTAuth - validates token, fetches user from DB
2. Handler - extracts user_id from context

**Password Security:**
- Current password verified with `auth.CheckPassword(plaintext, hashedPassword)`
- New password hashed with `auth.HashPassword(plaintext)` using bcrypt
- Only hashed passwords stored in database

**Partial Updates:**
- Uses pointer fields (*string, *int) in UpdateProfileRequest
- `nil` = field not provided (keep existing)
- Non-nil = update to new value (even if empty for some fields)
- Empty string blocked for URL fields (validation requires valid URL)

---

## Next Steps

Profile Management is now complete. Ready to proceed with:

1. **Prometheus Metrics** - Add /metrics endpoint for monitoring
2. **Enhanced Health Checks** - Add component health (DB, disk, memory)
3. **GraphQL API** - Alternative query interface
4. **WebSocket** - Real-time profile update notifications
5. **Audit Logging** - Track profile changes for compliance

---

**Status:** ✅ COMPLETE - All endpoints tested and working
**Date:** 2025-10-31
**Version:** 1.0.0
