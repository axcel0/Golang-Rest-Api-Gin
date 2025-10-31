# 🚀 GO Lang Project 01

Production-ready REST API built with Go 1.25.3, following best practices and clean architecture.

## ✨ Features

- ✅ **Clean Architecture** - Separation of concerns with layered design
- ✅ **Request Validation** - Using `validator/v10` with detailed error responses
- ✅ **Pagination & Filtering** - Efficient data retrieval with search, sort, filter
- ✅ **Configuration Management** - Viper-based config with environment override
- ✅ **Rate Limiting** - Per-IP protection using token bucket algorithm
- ✅ **Structured Logging** - Ready for `log/slog` integration
- ✅ **Type-Safe SQL** - SQLC for compile-time SQL verification
- ✅ **Database Migrations** - Version-controlled schema changes
- ✅ **Comprehensive Testing** - Unit tests with race detection
- ✅ **API Documentation** - Swagger/OpenAPI ready
- ✅ **CI/CD Pipeline** - Automated quality gates
- ✅ **Security** - JWT authentication, vulnerability scanning
- ✅ **Docker Support** - Multi-stage production build
- ✅ **NO DEPRECATED CODE** - SA1019 enforcement in CI

## �️ Tech Stack

- **Go**: 1.25.3 (Latest Stable)
- **Framework**: Gin v1.11.0
- **ORM**: GORM v1.31.0
- **Config**: Viper v1.21.0
- **Validation**: go-playground/validator/v10
- **Rate Limiting**: golang.org/x/time/rate
- **Type-Safe SQL**: sqlc
- **Migrations**: golang-migrate/migrate/v4
- **Auth**: golang-jwt/jwt/v5
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
make help           # Show all available commands
make build          # Build the application
make run            # Build and run
make test           # Run tests
make test-race      # Run tests with race detector
make test-cover     # Generate coverage report
make lint           # Run linters (includes SA1019)
make vet            # Run go vet
make vuln           # Check vulnerabilities
make staticcheck    # Run staticcheck
make fmt            # Format code
make ci             # Run all CI checks
make pre-commit     # Quick checks before commit
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

#### Users

```http
GET    /api/v1/users          # List users (paginated)
GET    /api/v1/users/:id      # Get user by ID
POST   /api/v1/users          # Create user
PUT    /api/v1/users/:id      # Update user
DELETE /api/v1/users/:id      # Delete user
```

#### Query Parameters (Pagination)

```
page=1               # Page number (default: 1)
limit=10             # Items per page (default: 10)
sort=created_at      # Sort field
order=desc           # Sort order (asc/desc)
search=john          # Search in name/email
active=true          # Filter by status
```

### Example Request

```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=10&sort=name&order=asc"
```

### Example Response

```json
{
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "age": 30,
      "active": true
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

- **Rate Limiting**: 100 requests per minute per IP
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

### Run all tests
```bash
make test
```

### Run with race detection
```bash
make test-race
```

### Generate coverage report
```bash
make test-cover
open coverage.html
```

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


