# Go Lang Project 01 - Learning Project

Project belajar Go dengan arsitektur clean dan best practices! 🚀

## ✨ Features

- ✅ **Clean Architecture** - Separation of concerns dengan layers
- ✅ **Gin Framework** - Fast HTTP router dan middleware
- ✅ **GORM ORM** - Database operations dengan SQLite
- ✅ **Goroutines** - Concurrent operations untuk batch & stats
- ✅ **Middleware** - Custom logging, CORS, error handling, recovery
- ✅ **Health Checks** - Liveness & readiness probes
- ✅ **API Versioning** - Versioned endpoints (`/api/v1/`)
- ✅ **Context Management** - Request timeouts & cancellation
- ✅ **Response Helpers** - Consistent API responses

## 📁 Struktur Project

```
GO Lang Project 01/
├── cmd/api/                    # Entry point aplikasi
│   └── main.go                 # Server initialization & routing
├── internal/                   # Private application code
│   ├── handlers/               # HTTP handlers (controller layer)
│   │   ├── user_handler.go    # User CRUD endpoints
│   │   └── health_handler.go  # Health check endpoints
│   ├── middleware/             # Custom middleware
│   │   ├── logger.go          # Structured logging
│   │   ├── cors.go            # Cross-origin resource sharing
│   │   └── error_handler.go   # Error handling & recovery
│   ├── models/                 # Data structures
│   │   └── models.go          # User model & DTOs
│   ├── services/               # Business logic layer
│   │   └── user_service.go    # User business logic
│   └── repository/             # Data access layer
│       └── user_repository.go # Database operations
├── pkg/                        # Public reusable code
│   ├── database/               # Database connection
│   │   └── database.go        # SQLite connection manager
│   └── utils/                  # Utility functions
│       ├── response.go        # Response helper functions
│       └── id_generator.go    # ID generation utility
├── bin/                        # Compiled binaries
├── go.mod                      # Go modules
├── go.sum                      # Dependency checksums
├── goproject.db                # SQLite database file
└── README.md
```

## 🎯 Konsep Arsitektur

**Layered Architecture (Clean Architecture):**

1. **Handlers** (Presentation Layer)
   - Menerima HTTP requests dari client
   - Validasi input & binding JSON
   - Memanggil service layer
   - Mengembalikan HTTP responses (menggunakan response helpers)
   - Error handling untuk user-facing errors

2. **Services** (Business Logic Layer)
   - Business rules & validations
   - Orchestration logic
   - Koordinasi antara multiple repositories
   - Concurrent operations dengan goroutines
   - Context management untuk timeouts

3. **Repository** (Data Access Layer)
   - CRUD operations dengan GORM
   - Data persistence ke SQLite
   - Database queries & transactions
   - Soft delete support
   - Context-aware database operations

4. **Middleware** (Cross-Cutting Concerns)
   - Logging: Request/response logging dengan latency tracking
   - CORS: Cross-origin resource sharing
   - Error Handler: Centralized error handling
   - Recovery: Panic recovery untuk stability

5. **Utils** (Helper Functions)
   - Response helpers: Consistent API responses
   - ID generators: Unique identifier generation
   - Reusable utilities

## 🚀 Cara Menjalankan

```bash
# Install dependencies
go mod download

# Run the application (development)
go run cmd/api/main.go

# Build the application
go build -o bin/server cmd/api/main.go

# Run the binary (production)
./bin/server
```

Server akan berjalan di `http://localhost:8080`

**Framework & Dependencies:**
- Gin v1.11.0 - HTTP web framework
- GORM v1.31.0 - ORM library
- SQLite - Embedded database (zero configuration)

## 📝 API Endpoints

### Health Check Endpoints

**Liveness Probe**
```bash
GET /health
```
Response:
```json
{
  "status": "OK",
  "service": "Go-Lang-project-01"
}
```

**Readiness Probe** (dengan database connectivity check)
```bash
GET /ready
```
Response:
```json
{
  "status": "ready",
  "database": "connected"
}
```

### Users API (Versioned: `/api/v1/`)

**Get All Users**
```bash
GET /api/v1/users
```
Response:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "age": 25,
      "is_active": true,
      "created_at": "2025-10-30T10:00:00Z",
      "updated_at": "2025-10-30T10:00:00Z"
    }
  ]
}
```

**Get User by ID**
```bash
GET /api/v1/users/{id}
```

**Create User**
```bash
POST /api/v1/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 25
}
```
Response (201 Created):
```json
{
  "success": true,
  "message": "user created successfully",
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "age": 25,
    "is_active": true,
    "created_at": "2025-10-30T10:00:00Z",
    "updated_at": "2025-10-30T10:00:00Z"
  }
}
```

**Update User**
```bash
PUT /api/v1/users/{id}
Content-Type: application/json

{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "age": 26
}
```

**Delete User** (Soft Delete)
```bash
DELETE /api/v1/users/{id}
```

**Batch Create Users** 🚀 (Concurrent dengan Goroutines)
```bash
POST /api/v1/users/batch
Content-Type: application/json

[
  {
    "name": "User 1",
    "email": "user1@example.com",
    "age": 22
  },
  {
    "name": "User 2",
    "email": "user2@example.com",
    "age": 23
  }
]
```
Response:
```json
{
  "success": true,
  "message": "users created successfully",
  "data": [...]
}
```

**Get User Statistics** 🚀 (Concurrent dengan Goroutines)
```bash
GET /api/v1/users/stats
```
Response:
```json
{
  "success": true,
  "data": {
    "total_users": 10,
    "active_users": 8,
    "inactive_users": 2
  }
}
```

## 🧪 Testing dengan cURL

```bash
# Health checks
curl http://localhost:8080/health
curl http://localhost:8080/ready

# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","age":25}'

# Get all users
curl http://localhost:8080/api/v1/users | jq

# Get user by ID
curl http://localhost:8080/api/v1/users/1 | jq

# Update user
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com","age":26}' | jq

# Delete user (soft delete)
curl -X DELETE http://localhost:8080/api/v1/users/1 | jq

# Batch create users (concurrent with goroutines)
curl -X POST http://localhost:8080/api/v1/users/batch \
  -H "Content-Type: application/json" \
  -d '[
    {"name":"User 1","email":"user1@example.com","age":22},
    {"name":"User 2","email":"user2@example.com","age":23},
    {"name":"User 3","email":"user3@example.com","age":24}
  ]' | jq

# Get user statistics (concurrent with goroutines)
curl http://localhost:8080/api/v1/users/stats | jq
```

## 📚 Best Practices yang Dipakai

### Architecture & Design Patterns
1. ✅ **Clean Architecture**: Separation of concerns dengan layers (handlers → services → repository)
2. ✅ **Standard Project Layout**: Mengikuti standar Go community (`cmd/`, `internal/`, `pkg/`)
3. ✅ **Dependency Injection**: Service dan repository di-inject melalui constructors
4. ✅ **Repository Pattern**: Abstraksi data access layer
5. ✅ **Middleware Pattern**: Cross-cutting concerns (logging, CORS, error handling)

### Code Quality
6. ✅ **Error Handling**: Proper error handling & propagation di setiap layer
7. ✅ **Centralized Error Handling**: Middleware untuk consistent error responses
8. ✅ **Response Helpers**: Reusable functions untuk consistent API responses
9. ✅ **Meaningful Names**: Variable, function, dan package names yang descriptive
10. ✅ **DRY Principle**: Don't Repeat Yourself - reusable utilities

### Concurrency & Performance
11. ✅ **Goroutines**: Concurrent operations untuk batch create & stats calculation
12. ✅ **Semaphore Pattern**: Limiting concurrent goroutines (max 5)
13. ✅ **sync.WaitGroup**: Proper goroutine synchronization
14. ✅ **Context Management**: Request timeouts (5-30 seconds) & cancellation
15. ✅ **Thread-Safe Operations**: Context-aware database operations

### API Design
16. ✅ **RESTful API**: REST principles dengan proper HTTP methods
17. ✅ **API Versioning**: Versioned endpoints (`/api/v1/`) untuk backward compatibility
18. ✅ **JSON API**: Consistent JSON request/response format
19. ✅ **HTTP Status Codes**: Proper usage (200, 201, 400, 404, 500)
20. ✅ **Health Checks**: Liveness & readiness probes untuk production

### Database & Persistence
21. ✅ **ORM Usage**: GORM untuk type-safe database operations
22. ✅ **Auto Migration**: Automatic schema migration
23. ✅ **Soft Delete**: Preserve data dengan deleted_at timestamp
24. ✅ **Database Indexing**: Email index untuk faster queries
25. ✅ **Connection Management**: Singleton database connection

## 🎓 Konsep Go yang Dipelajari

### Fundamentals
- ✅ **Package Organization**: Proper package structure (`cmd/`, `internal/`, `pkg/`)
- ✅ **Struct dan Methods**: Custom types dengan behavior
- ✅ **Interfaces**: Implicit interfaces untuk abstraction
- ✅ **Pointer vs Values**: Kapan pakai pointer, kapan pakai value
- ✅ **Error Handling**: Explicit error returns & handling
- ✅ **JSON Encoding/Decoding**: Marshaling & unmarshaling dengan struct tags

### Web Development
- ✅ **Gin Framework**: Fast HTTP router & middleware
- ✅ **HTTP Methods**: GET, POST, PUT, DELETE
- ✅ **Route Groups**: Organizing routes dengan versioning
- ✅ **Middleware**: Custom middleware untuk cross-cutting concerns
- ✅ **Request Binding**: JSON binding & validation

### Concurrency
- ✅ **Goroutines**: Lightweight concurrent execution
- ✅ **Channels**: Communication between goroutines (buffered channels)
- ✅ **sync.WaitGroup**: Waiting for multiple goroutines
- ✅ **Semaphore Pattern**: Limiting concurrent operations
- ✅ **Context**: Timeout & cancellation untuk goroutines

### Database
- ✅ **GORM ORM**: Object-relational mapping
- ✅ **SQLite**: Embedded database
- ✅ **Auto Migration**: Schema management
- ✅ **CRUD Operations**: Create, Read, Update, Delete
- ✅ **Soft Delete**: Logical deletion dengan GORM
- ✅ **Database Indexing**: Performance optimization

### Advanced Patterns
- ✅ **Dependency Injection**: Constructor injection
- ✅ **Repository Pattern**: Data access abstraction
- ✅ **Service Layer**: Business logic separation
- ✅ **Error Wrapping**: Context-aware errors
- ✅ **Panic Recovery**: Graceful error handling

## � Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.25.3 | Programming language |
| **Gin** | v1.11.0 | HTTP web framework |
| **GORM** | v1.31.0 | ORM library |
| **SQLite** | v1.6.0 | Embedded database |

## �🔜 Next Steps untuk Belajar

### Already Implemented ✅
- ✅ Clean Architecture
- ✅ Gin Framework
- ✅ GORM dengan SQLite
- ✅ Middleware (logging, CORS, error handling)
- ✅ Goroutines (batch & stats)
- ✅ Context management
- ✅ Health checks
- ✅ API versioning

### Future Enhancements 🎯
1. **Testing**
   - Unit tests untuk services & repositories
   - Integration tests
   - Test coverage reports

2. **Database**
   - Migration dari SQLite ke PostgreSQL/MySQL
   - Database migrations tool (golang-migrate)
   - Connection pooling optimization

3. **Authentication & Authorization**
   - JWT authentication
   - Middleware untuk protected routes
   - Role-based access control (RBAC)

4. **Validation & Security**
   - Request validation library (validator)
   - Input sanitization
   - Rate limiting
   - SQL injection prevention

5. **Configuration**
   - Environment variables (.env)
   - Configuration management (viper)
   - Multiple environments (dev, staging, prod)

6. **Observability**
   - Structured logging (logrus/zap)
   - Metrics collection (Prometheus)
   - Distributed tracing (OpenTelemetry)
   - Application monitoring

7. **DevOps**
   - Docker containerization
   - Docker Compose untuk local development
   - CI/CD pipeline (GitHub Actions)
   - Kubernetes deployment

8. **API Documentation**
   - Swagger/OpenAPI specification
   - API documentation generator
   - Postman collection

## 📖 Resources

### Official Documentation
- [Go Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Gin Framework](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)

### Learning Resources
- [Go by Example](https://gobyexample.com/)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)

### Community
- [Go Forum](https://forum.golangbridge.org/)
- [Go Reddit](https://www.reddit.com/r/golang/)
- [Gophers Slack](https://invite.slack.golangbridge.org/)

---

**Project Status**: 🟢 Production Ready with Best Practices

Happy coding! 🎉 Selamat belajar Go! 🚀
