# 📚 Documentation Directory

## Main Documentation

**👉 [DOCUMENTATION.md](DOCUMENTATION.md)** - **Complete consolidated documentation (5,085 lines)**

This file contains **ALL project documentation** in one place:
- JWT Authentication & Authorization
- Role-Based Access Control (RBAC)
- Swagger/OpenAPI Documentation
- Prometheus Metrics & Monitoring
- Enhanced Health Checks
- WebSocket Real-Time Updates
- User Profile Management
- Activity Logging / Audit Trail
- Repository Tests & Coverage
- Quick Start Guide
- Architecture Overview
- Best Practices
- Development Roadmap

**All individual markdown files have been consolidated and removed.** This single file is now the **single source of truth** for all project documentation.

## Generated Documentation

### GoDoc (Go Documentation)
Located in `godoc/` directory:
- `handlers.txt` - HTTP handlers documentation
- `services.txt` - Business logic services
- `repository.txt` - Data access layer
- `middleware.txt` - HTTP middleware
- `auth.txt` - Authentication utilities
- `models.txt` - Data models
- `metrics.txt` - Prometheus metrics
- `health.txt` - Health check system
- `websocket.txt` - WebSocket support
- `database.txt` - Database connection
- `logger.txt` - Structured logging
- `utils.txt` - Utility functions
- `configs.txt` - Configuration management

### Swagger/OpenAPI
- `swagger.json` - OpenAPI 3.0 specification (JSON)
- `swagger.yaml` - OpenAPI 3.0 specification (YAML)
- Access interactive UI at: `http://localhost:8080/swagger/index.html`

## Quick Access

### For Developers
1. **Getting Started**: Read [DOCUMENTATION.md](DOCUMENTATION.md) Quick Start section
2. **API Reference**: Open Swagger UI at `http://localhost:8080/swagger/index.html`
3. **Code Documentation**: Browse `godoc/` directory
4. **Testing**: See DOCUMENTATION.md → Repository Tests section

### For API Users
1. **API Endpoints**: See [DOCUMENTATION.md](DOCUMENTATION.md) or Swagger UI
2. **Authentication**: See JWT Authentication section in DOCUMENTATION.md
3. **Examples**: Find cURL examples throughout DOCUMENTATION.md

### For DevOps
1. **Monitoring**: See Prometheus Metrics section in DOCUMENTATION.md
2. **Health Checks**: See Enhanced Health Checks section
3. **Configuration**: See config.yaml and .env.example

## Documentation Statistics

- **Main Documentation**: 5,085 lines (DOCUMENTATION.md)
- **GoDoc Files**: 5 package documentations
- **API Endpoints Documented**: 25+
- **Code Examples**: 100+
- **Test Cases Documented**: 50+

## Updating Documentation

### Regenerate Swagger Documentation
```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate Swagger docs
swag init -g cmd/api/main.go -o ./docs

# Access Swagger UI
# http://localhost:8080/swagger/index.html
```

### Regenerate GoDoc
```bash
# Generate documentation for all packages
mkdir -p docs/godoc
go doc ./internal/handlers > docs/godoc/handlers.txt
go doc ./internal/services > docs/godoc/services.txt
go doc ./internal/repository > docs/godoc/repository.txt
# ... etc for all packages
```

### Update Main Documentation
To update the documentation, edit `DOCUMENTATION.md` directly. This is now the single source of truth.

```bash
# Edit the main documentation
code docs/DOCUMENTATION.md

# Or use any text editor
nano docs/DOCUMENTATION.md
```

---

**Last Updated**: November 3, 2025

**Note**: All individual markdown files have been consolidated into **DOCUMENTATION.md** for easier maintenance and navigation.
