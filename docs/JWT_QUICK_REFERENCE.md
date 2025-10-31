# 🔐 JWT Authentication Quick Reference

## 🚀 Quick Start

### 1. Start Server
```bash
make run
# or
./bin/server
```

### 2. Test Authentication
```bash
# Make test script executable
chmod +x test_auth.sh

# Run tests
./test_auth.sh
```

## 📋 API Endpoints

### Public Endpoints (No Authentication)

#### Register
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepass123",
  "age": 30
}
```

#### Login
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "securepass123"
}
```

#### Refresh Token
```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Protected Endpoints (Requires Bearer Token)

#### Get Profile
```bash
GET /api/v1/auth/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## 🔑 Configuration

### Environment Variables
```bash
# JWT Configuration
JWT_SECRETKEY=your-super-secret-jwt-key-change-this-in-production
JWT_ACCESSTOKENDURATION=24h      # Access token expiry
JWT_REFRESHTOKENDURATION=168h    # Refresh token expiry (7 days)
```

### Config File (`configs/config.yaml`)
```yaml
jwt:
  secretkey: "change-this-secret-key-in-production"
  accesstokenduration: "24h"
  refreshtokenduration: "168h"
```

## 📝 Usage Examples

### 1. Register New User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Smith",
    "email": "jane@example.com",
    "password": "mypassword123",
    "age": 28
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "user registered successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "name": "Jane Smith",
      "email": "jane@example.com",
      "age": 28,
      "is_active": true
    }
  }
}
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "password": "mypassword123"
  }'
```

**Response:** Same as registration

### 3. Access Protected Route
```bash
# Save token from login response
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Use token to access protected route
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Jane Smith",
    "email": "jane@example.com",
    "age": 28,
    "is_active": true,
    "created_at": "2025-10-31T10:00:00Z",
    "updated_at": "2025-10-31T10:00:00Z"
  }
}
```

### 4. Refresh Access Token
```bash
# When access token expires, use refresh token to get new one
REFRESH_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}"
```

**Response:**
```json
{
  "success": true,
  "message": "token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

## 🔒 Security Best Practices

### Production Checklist
- [ ] Change JWT secret key (JWT_SECRETKEY)
- [ ] Use strong secret (min 32 characters, random)
- [ ] Store secret in secure vault (not in code)
- [ ] Use HTTPS in production
- [ ] Enable rate limiting (already configured: 100 req/min)
- [ ] Set appropriate token expiry times
- [ ] Implement token blacklist for logout (optional)
- [ ] Add CORS configuration for your domain
- [ ] Monitor failed login attempts
- [ ] Implement account lockout after X failed attempts (optional)

### Password Requirements
- Minimum length: 6 characters (configurable in validation)
- Hashed with bcrypt (cost 10)
- Never stored in plaintext
- Never returned in API responses

### Token Management
- **Access Token**: Short-lived (24h default)
  - Used for API authentication
  - Include in Authorization header
  
- **Refresh Token**: Long-lived (7 days default)
  - Used to get new access tokens
  - Should be stored securely by client

## 🐛 Troubleshooting

### 401 Unauthorized
```json
{
  "success": false,
  "message": "invalid or expired token"
}
```
**Solution**: 
- Check if token is included in Authorization header
- Verify format: `Authorization: Bearer YOUR_TOKEN`
- Token may be expired, use refresh endpoint
- Ensure JWT secret matches between requests

### Email Already Registered
```json
{
  "success": false,
  "message": "email already registered"
}
```
**Solution**: Use a different email or login instead

### Invalid Email or Password
```json
{
  "success": false,
  "message": "invalid email or password"
}
```
**Solution**:
- Check email is correct
- Check password is correct
- Ensure account is active

## 📊 Token Structure

### JWT Claims
```json
{
  "user_id": 1,
  "email": "jane@example.com",
  "exp": 1761968846,  // Expiry timestamp
  "nbf": 1761882446,  // Not before timestamp
  "iat": 1761882446   // Issued at timestamp
}
```

### Accessing User Info in Handlers
After authentication middleware, user info is available in Gin context:
```go
// Get user ID
userID, exists := c.Get("user_id")
if !exists {
    // User not authenticated
}

// Get user email
userEmail, exists := c.Get("user_email")
```

## 🧪 Testing

### Manual Testing
```bash
# 1. Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@test.com","password":"test123","age":25}'

# 2. Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123"}'

# 3. Get Profile (replace TOKEN)
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer TOKEN"
```

### Automated Testing
```bash
./test_auth.sh
```

## 📚 Documentation

For detailed implementation details, see:
- [JWT_AUTH_IMPLEMENTATION.md](./docs/JWT_AUTH_IMPLEMENTATION.md)
- [COPILOT.md](./COPILOT.md)
- [README.md](./README.md)

## 🎯 Next Steps

1. **Add Swagger Documentation**
   ```bash
   go get github.com/swaggo/swag/cmd/swag
   swag init -g cmd/api/main.go
   ```

2. **Write Unit Tests**
   - Test auth handlers
   - Test JWT manager
   - Test password hashing
   - Test middleware

3. **Optional Enhancements**
   - Email verification
   - Forgot password
   - Token blacklist (logout)
   - Role-based access control (RBAC)
   - OAuth2/SSO integration

---

**Version**: 1.0.0  
**Last Updated**: October 31, 2025  
**Go Version**: 1.25.3
