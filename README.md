# 🚀 GO Lang Project 01

Production-ready REST API built with Go 1.25.3, following best practices and clean architecture.

## ✨ Features

- ✅ **Clean Architecture** - Separation of concerns with layered design
- ✅ **Request Validation** - Using `validator/v10` with detailed error responses
- ✅ **Pagination & Filtering** - Efficient data retrieval with search, sort, filter
- ✅ **Configuration Management** - Viper-based config with environment override
- ✅ **Rate Limiting** - Per-IP protection using token bucket algorithm (100 req/min)
- ✅ **Structured Logging** - log/slog with JSON format for production monitoring
- ✅ **JWT Authentication** - Secure token-based auth with bcrypt password hashing
- ✅ **RBAC (Role-Based Access Control)** - Three-tier role system (superadmin/admin/user)
- ✅ **Type-Safe SQL** - SQLC for compile-time SQL verification
- ✅ **Database Migrations** - Version-controlled schema changes
- ✅ **Comprehensive Testing** - Unit tests with race detection
- ✅ **API Documentation** - Swagger/OpenAPI ready
- ✅ **CI/CD Pipeline** - Automated quality gates with golangci-lint, govulncheck
- ✅ **Docker Support** - Multi-stage production build
- ✅ **NO DEPRECATED CODE** - SA1019 enforcement in CI

## 🛠️ Tech Stack

- **Go**: 1.25.3 (Latest Stable)
- **Framework**: Gin v1.11.0
- **ORM**: GORM v1.31.0
- **Config**: Viper v1.21.0
- **Validation**: go-playground/validator/v10
- **Rate Limiting**: golang.org/x/time/rate
- **Logging**: log/slog (stdlib)
- **Auth**: golang-jwt/jwt/v5, golang.org/x/crypto/bcrypt
- **Type-Safe SQL**: sqlc
- **Migrations**: golang-migrate/migrate/v4
- **Testing**: testify, httptest
- **Docs**: swaggo/swag

## 📁 Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── handlers/                # HTTP handlers
│   ├── services/                # Business logic
│   ├── repository/              # Data access layer
│   │   ├── queries/             # SQLC queries
│   │   └── sqlc/                # Generated type-safe code
│   ├── models/                  # Domain models
│   └── middleware/              # HTTP middleware
├── pkg/
│   └── utils/                   # Shared utilities
├── configs/                     # Configuration files
├── scripts/
│   └── migrations/              # Database migrations
├── .github/
│   └── workflows/               # CI/CD pipelines
├── Makefile                     # Build automation
├── Dockerfile                   # Production container
└── COPILOT.md                   # Development guidelines
```

## 🚀 Quick Start

### Prerequisites

- Go 1.25.3+
- Make
- Docker (optional)

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd "GO Lang Project 01"
```

2. **Install dependencies**
```bash
go mod download
```

3. **Install development tools**
```bash
make install-tools
```

4. **Setup configuration**
```bash
cp .env.example .env
# Edit .env with your settings
```

5. **Run the application**
```bash
make run
```

## 🧪 Development

### Available Make Commands

```bash
make help             # Show all available commands
make build            # Build the application
make run              # Build and run
make test             # Run all tests
make test-repo        # Run repository tests only
make test-race        # Run tests with race detector
make test-cover       # Generate coverage report
make test-cover-repo  # Repository coverage (83%)
make bench            # Run benchmarks
make lint             # Run linters (includes SA1019)
make vet              # Run go vet
make vuln             # Check vulnerabilities
make staticcheck      # Run staticcheck
make fmt              # Format code
make swagger          # Generate Swagger docs
make ci               # Run all CI checks
make pre-commit       # Quick checks before commit
```

### Quality Gates

Before every commit, ensure all checks pass:

```bash
make pre-commit
```

This runs:
- ✅ Code formatting (`gofmt`)
- ✅ Static analysis (`go vet`)
- ✅ Linting (`golangci-lint` with SA1019)
- ✅ Tests with race detection
- ✅ Security scanning (`govulncheck`)

## 📝 API Documentation

### Endpoints

#### Health
```http
GET    /health                # Liveness probe
GET    /ready                 # Readiness probe
```

#### Authentication
```http
POST   /api/v1/auth/register  # Register new user (default role: user)
POST   /api/v1/auth/login     # Login with email/password
POST   /api/v1/auth/refresh   # Refresh access token
GET    /api/v1/auth/profile   # Get authenticated user profile (requires Bearer token)
```

#### Users
```http
GET    /api/v1/users          # List users (paginated) [All authenticated users]
GET    /api/v1/users/stats    # Get user statistics [All authenticated users]
GET    /api/v1/users/:id      # Get user by ID [All authenticated users]
POST   /api/v1/users          # Create user [Admin+]
POST   /api/v1/users/batch    # Batch create users [Admin+]
PUT    /api/v1/users/:id      # Update user [Admin+]
DELETE /api/v1/users/:id      # Delete user [Admin+]
PUT    /api/v1/users/:id/role # Change user role [Superadmin only]
```

**Legend**: `[All]` = Any authenticated user, `[Admin+]` = Admin or Superadmin, `[Superadmin only]` = Superadmin only

## 🔐 Role-Based Access Control (RBAC)

This API implements a three-tier role-based permission system:

### Role Hierarchy
```
👑 superadmin (highest)  →  👮 admin (middle)  →  👤 user (default)
```

### Role Permissions

| Action | User | Admin | Superadmin |
|--------|------|-------|------------|
| View users | ✅ | ✅ | ✅ |
| Create/Update/Delete users | ❌ | ✅ | ✅ |
| Change user roles | ❌ | ❌ | ✅ |

### Creating First Superadmin

**Option 1: Using Go script** (Recommended)
```bash
# 1. Register a user
curl -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"Admin","email":"admin@system.com","password":"securepass","age":30}'

# 2. Promote to superadmin
go run scripts/promote_user.go admin@system.com superadmin
```

**Option 2: Direct database update**
```bash
sqlite3 goproject.db "UPDATE users SET role = 'superadmin' WHERE email = 'admin@system.com';"
```

### Promoting Other Users

As a superadmin, use the API:
```bash
# Promote user to admin
curl -X PUT "http://localhost:8080/api/v1/users/2/role" \
  -H "Authorization: Bearer YOUR_SUPERADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}'
```

### Testing RBAC
```bash
./test_rbac_simple.sh  # Quick RBAC verification
```

📖 **Full RBAC Documentation**: See [docs/RBAC_IMPLEMENTATION.md](docs/RBAC_IMPLEMENTATION.md)

---

#### Query Parameters (Pagination)
```
page=1               # Page number (default: 1)
limit=10             # Items per page (default: 10)
sort=created_at      # Sort field
order=desc           # Sort order (asc/desc)
search=john          # Search in name/email
active=true          # Filter by status
```

### Authentication Examples

#### Register
```bash
curl -X POST "http://localhost:8080/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepass123",
    "age": 30
  }'
```

#### Login
```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

#### Get Profile (Protected)
```bash
curl -X GET "http://localhost:8080/api/v1/auth/profile" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Example Response (Login/Register)
```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "age": 30,
      "is_active": true,
      "created_at": "2025-10-31T10:00:00Z",
      "updated_at": "2025-10-31T10:00:00Z"
    }
  }
}
```

### Example Request (Users)
```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=10&sort=name&order=asc"
```

### Example Response (Users)
```json
{
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "age": 30,
      "is_active": true
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

## 🔒 Security

- **JWT Authentication**: Secure token-based authentication with HS256
- **Password Security**: Bcrypt hashing with cost 10, passwords never exposed
- **Token Management**: Short-lived access tokens (24h), long-lived refresh tokens (7d)
- **Protected Routes**: Middleware-based authorization
- **Rate Limiting**: 100 requests per minute per IP with burst of 10
- **Input Validation**: All requests validated with detailed error responses
- **SQL Injection**: Protected via GORM/SQLC parameterized queries
- **Vulnerability Scanning**: Automated with `govulncheck`
- **Dependency Updates**: Weekly automated PRs via Dependabot
- **No Deprecated Code**: SA1019 check in CI prevents deprecated imports

## 🐳 Docker

### Build and run with Docker

```bash
docker build -t go-api .
docker run -p 8080:8080 go-api
```

### Production deployment

The Dockerfile uses multi-stage builds with:
- Alpine Linux (minimal size)
- Non-root user
- Health checks
- Security hardening

## 📊 Testing

We maintain comprehensive test coverage with production-ready testing practices:

### Test Coverage by Layer

- **Repository Layer**: 83% coverage ([See docs/REPOSITORY_TESTS.md](docs/REPOSITORY_TESTS.md))
  - ✅ All 9 CRUD methods tested
  - ✅ Table-driven test patterns
  - ✅ In-memory SQLite for fast execution
  - ✅ Edge cases and boundary conditions
  - ✅ Performance benchmarks

- **Service Layer**: Coming soon
- **Handler Layer**: Coming soon

### Running Tests

```bash
# Run all tests
make test

# Run repository tests only
make test-repo

# Run with race detection
make test-race

# Generate coverage report
make test-cover
open coverage.html

# Run repository tests with coverage
make test-cover-repo
open coverage-repo.html

# Run benchmarks
make bench
```

### Test Results

```
Repository Layer Tests:
✅ 14 test suites
✅ 19 subtests
✅ 83.0% code coverage
✅ ~38ms execution time
✅ 3 performance benchmarks

Benchmarks:
- Create: ~86μs per operation
- GetByID: ~35μs per operation  
- GetAllPaginated: ~150μs per operation
```

### Test Features

- ✅ **Table-Driven Tests**: Multiple scenarios per test function
- ✅ **Test Isolation**: Each test uses independent in-memory database
- ✅ **Context Handling**: All tests respect context cancellation
- ✅ **Edge Case Coverage**: Empty databases, special characters, large datasets
- ✅ **Descriptive Names**: Clear test case identification
- ✅ **Fast Execution**: In-memory SQLite for speed

See [docs/REPOSITORY_TESTS.md](docs/REPOSITORY_TESTS.md) for detailed testing documentation.

## 🔄 CI/CD

GitHub Actions workflow runs on every push/PR:

1. ✅ Format check (`gofmt`)
2. ✅ Vet analysis (`go vet`)
3. ✅ Static analysis (`staticcheck` with SA1019)
4. ✅ Linting (`golangci-lint`)
5. ✅ Vulnerability scan (`govulncheck`)
6. ✅ Tests with race detection
7. ✅ Build verification
8. ✅ Coverage report

## � Documentation

- **COPILOT.md** - Comprehensive development guidelines
- **API Docs** - Swagger UI (coming soon)
- **Code Comments** - Godoc compatible

## 🤝 Contributing

1. Follow guidelines in `COPILOT.md`
2. Run `make pre-commit` before committing
3. Ensure all CI checks pass
4. Update documentation as needed

## 📄 License

MIT License

## � Author

**Axel**

---

**Built with ❤️ using Go 1.25.3**


