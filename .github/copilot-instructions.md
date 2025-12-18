# GitHub Copilot Instructions for POS01 Project

## 🎯 Core Development Philosophy

**ALWAYS follow these principles when generating code:**

1. **Read Official Documentation First** - Before implementing any feature, reference official docs for Go, Gin, GORM, and any libraries used
2. **Use Established Libraries** - Prefer well-maintained, non-deprecated libraries over custom implementations
3. **Production-Ready Code** - This software will be sold commercially, ensure enterprise-grade quality
4. **Security First** - Especially for payment processing, calculations, and financial transactions

---

## 📚 Technology Stack & Best Practices

### **Go (1.25.5)**
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `golangci-lint` for code quality (must pass with 0 errors)
- Error handling: Always handle errors explicitly, never ignore
- Naming: Use idiomatic Go naming (camelCase for private, PascalCase for public)
- Package structure: Follow standard Go project layout
- Use `context.Context` for cancellation and timeouts
- Prefer `errors.Is()` and `errors.As()` for error checking

**Best Practices:**
```go
// ✅ DO: Explicit error handling
if err != nil {
    return fmt.Errorf("failed to create product: %w", err)
}

// ❌ DON'T: Ignore errors
db.Create(&product) // Missing error check
```

### **Gin Framework**
- Reference: https://gin-gonic.com/docs/
- Use middleware for cross-cutting concerns (auth, logging, CORS)
- Bind request data with proper validation (`ShouldBindJSON`, `ShouldBindQuery`)
- Return consistent response formats across all endpoints
- Use Gin's built-in validation with `binding:"required"` tags
- Handle panics with `gin.Recovery()` middleware

**Best Practices:**
```go
// ✅ DO: Proper request binding with validation
var req CreateProductRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}

// ❌ DON'T: Manual JSON parsing
// Avoid using json.Unmarshal manually when Gin provides binding
```

### **GORM (ORM)**
- Reference: https://gorm.io/docs/
- Always use transactions for operations that modify multiple tables
- Use `Preload` or `Joins` for eager loading (avoid N+1 queries)
- Implement soft deletes with `gorm.DeletedAt`
- Use database migrations for schema changes
- Add indexes for frequently queried fields
- Use `db.Model()` for updates to avoid loading full object

**Best Practices:**
```go
// ✅ DO: Use transactions for multi-table operations
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&transaction).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Model(&product).Update("stock", gorm.Expr("stock - ?", quantity)).Error; err != nil {
    tx.Rollback()
    return err
}

tx.Commit()

// ❌ DON'T: Multiple operations without transactions
db.Create(&transaction)
db.Model(&product).Update("stock", newStock) // Risk of inconsistency
```

### **JWT Authentication**
- Use existing `internal/auth` package (already implemented)
- Always validate tokens on protected endpoints
- Use short-lived access tokens (15-30 minutes)
- Implement refresh token mechanism
- Never expose JWT secret in code (use environment variables)

### **Input Validation**
- Use `github.com/go-playground/validator/v10` (already in project)
- Validate ALL user inputs (never trust client)
- Sanitize inputs to prevent SQL injection (GORM handles this, but be aware)
- Validate business logic (e.g., stock can't be negative)

**Best Practices:**
```go
type CreateProductRequest struct {
    Name        string  `json:"name" binding:"required,min=3,max=100"`
    Barcode     string  `json:"barcode" binding:"required,len=13"` // EAN-13
    BasePrice   float64 `json:"base_price" binding:"required,gt=0"`
    CostPrice   float64 `json:"cost_price" binding:"required,gt=0"`
    Stock       int     `json:"stock" binding:"required,gte=0"`
    CategoryID  uint    `json:"category_id" binding:"required"`
}
```

---

## 🔒 Security Requirements (CRITICAL)

### **Payment & Financial Calculations**
⚠️ **CRITICAL: Financial calculations must be 100% accurate**

1. **Use Decimal Type for Money**
   - ❌ DON'T use `float64` for currency (has precision issues)
   - ✅ DO use `decimal.Decimal` from `github.com/shopspring/decimal`
   - Example: `price := decimal.NewFromFloat(10.50)`

2. **Atomic Transactions**
   - ALL financial operations MUST use database transactions
   - Rollback on ANY error in the transaction chain
   - Log all financial transactions for audit trail

3. **Stock Management**
   - Prevent negative stock with database constraints
   - Use atomic updates: `UPDATE products SET stock = stock - ? WHERE id = ? AND stock >= ?`
   - Lock rows during transactions to prevent race conditions

4. **Audit Logging**
   - Log ALL create/update/delete operations
   - Include: user ID, timestamp, action, old/new values
   - Never delete financial records (use soft deletes only)

**Example: Safe Stock Reduction**
```go
// ✅ DO: Atomic stock update with validation
result := tx.Model(&models.Product{}).
    Where("id = ? AND stock >= ?", productID, quantity).
    Update("stock", gorm.Expr("stock - ?", quantity))

if result.RowsAffected == 0 {
    tx.Rollback()
    return errors.New("insufficient stock")
}
```

### **Role-Based Access Control (RBAC)**
- Enforce permissions at middleware level (not just in handlers)
- Never trust `role` from request body (get from JWT token)
- Default to deny (explicit permission required)
- Log all permission denials

**Middleware Pattern:**
```go
func RequireSuperadmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("userRole")
        if !exists || role != "superadmin" {
            c.JSON(http.StatusForbidden, gin.H{"error": "superadmin access required"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## 📦 Library Usage Guidelines

### **When to Use Libraries vs Custom Code**

✅ **USE LIBRARIES for:**
- Authentication/Authorization (JWT, OAuth)
- Date/Time manipulation (`time` package built-in)
- UUID generation (`github.com/google/uuid`)
- Decimal calculations (`github.com/shopspring/decimal`)
- Validation (`github.com/go-playground/validator`)
- Database ORM (GORM)
- HTTP routing (Gin)
- Logging (`github.com/sirupsen/logrus` or existing logger)
- Configuration management (`github.com/spf13/viper`)

✅ **ONLY if library is:**
- Actively maintained (commits within last 6 months)
- Well documented
- Has >1000 GitHub stars (or is officially recommended)
- Not deprecated
- Has good test coverage

❌ **AVOID custom implementations for:**
- Cryptography (use Go's `crypto` package)
- JWT handling (use existing `internal/auth`)
- SQL query builders (GORM handles this)
- JSON parsing (use encoding/json or Gin binding)

### **Recommended Libraries for POS01**

```go
// Payment & Financial
"github.com/shopspring/decimal" // Precise decimal calculations

// Validation
"github.com/go-playground/validator/v10" // Already in project

// UUID
"github.com/google/uuid" // For transaction IDs

// Time
"time" // Go standard library (sufficient)

// Logging
// Use existing logger in project (internal/logger)

// Configuration
"github.com/spf13/viper" // If not already present, or use environment variables
```

---

## 🏗️ Code Architecture Patterns

### **Layered Architecture (Follow Existing Project Structure)**

```
cmd/api/main.go           → Application entry point
internal/
  ├── handlers/           → HTTP handlers (Gin controllers)
  ├── services/           → Business logic
  ├── repositories/       → Database operations
  ├── models/             → GORM models
  ├── middleware/         → Gin middleware
  ├── auth/               → JWT authentication (existing)
  └── logger/             → Logging (existing)
pkg/
  └── database/           → Database connection & migrations
```

### **Dependency Injection Pattern**

```go
// ✅ DO: Inject dependencies
type ProductService struct {
    repo   *repositories.ProductRepository
    logger *logger.Logger
}

func NewProductService(repo *repositories.ProductRepository, logger *logger.Logger) *ProductService {
    return &ProductService{repo: repo, logger: logger}
}

// ❌ DON'T: Use global variables
var db *gorm.DB // Avoid this pattern
```

### **Repository Pattern (Database Layer)**

```go
type ProductRepository interface {
    Create(product *models.Product) error
    FindByID(id uint) (*models.Product, error)
    FindByBarcode(barcode string) (*models.Product, error)
    Update(product *models.Product) error
    Delete(id uint) error
    List(page, limit int) ([]models.Product, int64, error)
}
```

### **Service Pattern (Business Logic)**

```go
type TransactionService struct {
    transactionRepo *repositories.TransactionRepository
    productRepo     *repositories.ProductRepository
    stockRepo       *repositories.StockMovementRepository
    db              *gorm.DB
}

func (s *TransactionService) Checkout(req CheckoutRequest) (*models.Transaction, error) {
    // Start transaction
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // Business logic here
    // - Validate stock availability
    // - Create transaction
    // - Reduce stock
    // - Create stock movement audit
    // - Commit or rollback

    return transaction, tx.Commit().Error
}
```

---

## 🧪 Testing Requirements

### **Write Tests for:**
- All business logic (services)
- API endpoints (handlers)
- Repository operations
- Middleware (especially auth & permissions)
- Financial calculations (CRITICAL)

### **Testing Libraries:**
```go
"testing"                          // Go standard
"github.com/stretchr/testify/assert" // Assertions
"github.com/stretchr/testify/mock"   // Mocking
```

### **Test Coverage Target:**
- Minimum: 80% code coverage
- Critical paths (payments, stock): 100% coverage

---

## 📐 Database Design Best Practices

### **Foreign Keys & Constraints**
```sql
-- ✅ DO: Add constraints at database level
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
CHECK (stock >= 0)
CHECK (base_price > 0)
```

### **Indexes**
```go
// Add indexes in GORM model
type Product struct {
    Barcode string `gorm:"uniqueIndex;not null"`
    Name    string `gorm:"index"`
}
```

### **Soft Deletes**
```go
// Use GORM's soft delete
type Product struct {
    gorm.Model  // Includes DeletedAt
    // ... other fields
}
```

---

## 🔄 Migration Strategy

### **Database Migrations**
```go
// Use GORM AutoMigrate for development
db.AutoMigrate(&models.Product{}, &models.Transaction{})

// For production: Use migration files (sql-migrate or golang-migrate)
```

---

## 📋 Code Review Checklist (Auto-Check Before Committing)

- [ ] All errors handled explicitly
- [ ] Input validation present on all endpoints
- [ ] Financial operations use `decimal.Decimal`
- [ ] Database transactions used for multi-table operations
- [ ] Permissions checked via middleware
- [ ] Audit logs created for sensitive operations
- [ ] Soft deletes used (no hard deletes of user data)
- [ ] Code passes `golangci-lint run ./...`
- [ ] Tests written for business logic
- [ ] No hardcoded secrets or credentials
- [ ] Logging added for important operations
- [ ] Documentation comments for exported functions

---

## 🚨 Common Mistakes to Avoid

❌ **DON'T:**
```go
// Using float64 for money
price := 10.50  // ❌ Precision issues

// Ignoring errors
db.Create(&product)  // ❌ No error check

// SQL injection vulnerable (GORM prevents this, but be aware)
db.Where("name = '" + userInput + "'")  // ❌ Never do this

// No transaction for multi-step operations
db.Create(&transaction)
db.Model(&product).Update("stock", newStock)  // ❌ Not atomic

// Trusting client-provided role
if req.Role == "superadmin" {  // ❌ Security vulnerability
```

✅ **DO:**
```go
// Use decimal for money
price := decimal.NewFromFloat(10.50)

// Handle errors
if err := db.Create(&product).Error; err != nil {
    return fmt.Errorf("failed to create: %w", err)
}

// Use parameterized queries (GORM does this automatically)
db.Where("name = ?", userInput)

// Use transactions
tx := db.Begin()
// ... multiple operations
tx.Commit()

// Get role from JWT token
role, _ := c.Get("userRole")
```

---

## 🎯 POS01-Specific Requirements

### **Transaction Checkout Flow**
1. Validate kasir has permission (from JWT)
2. Check all products exist and have sufficient stock
3. Start database transaction
4. Create Transaction record
5. Create TransactionItem records (with price snapshot)
6. Reduce product stock (atomic update)
7. Create StockMovement audit records
8. Commit transaction
9. Return transaction with receipt data

### **Receipt Generation**
- Format: 80mm thermal printer compatible
- Include: Store name, transaction ID, timestamp, items, total, kasir name
- Return as structured JSON (frontend handles printing)

### **Barcode Scanner Integration**
- Endpoint: `GET /api/v1/products/by-barcode/:barcode`
- Return full product details with current stock
- Cache product list for performance

### **Analytics Calculations**
- Aggregate data (don't return raw transactions)
- Calculate: revenue, profit (base_price - cost_price), item count
- Group by: date, product, kasir, payment method
- Optimize queries with proper indexes

---

## 📖 Required Reading Before Coding

**Before implementing each phase, review:**

1. **Phase 1 (Models):**
   - GORM documentation: https://gorm.io/docs/models.html
   - GORM associations: https://gorm.io/docs/belongs_to.html
   - docs/POS01_DOCUMENTATION.md (Section 3: Database Schema)

2. **Phase 2 (API Endpoints):**
   - Gin documentation: https://gin-gonic.com/docs/
   - IMPLEMENTATION_PLAN.md (Phase 2 section)
   - docs/POS01_DOCUMENTATION.md (Section 4: API Specification)

3. **Phase 3 (Analytics):**
   - GORM aggregation: https://gorm.io/docs/advanced_query.html
   - docs/POS01_DOCUMENTATION.md (Section 8: Analytics)

4. **Phase 4 (Testing):**
   - IMPLEMENTATION_PLAN.md (18 curl test cases)
   - Go testing: https://pkg.go.dev/testing

---

## 🎓 Final Reminders

1. **Security > Speed** - Take time to implement security correctly
2. **Test Financial Logic** - Double/triple check all calculations
3. **Read Docs First** - Don't guess, read official documentation
4. **Use Libraries Wisely** - Maintained, well-tested libraries only
5. **Production Quality** - This code will be sold, make it excellent

---

**When in doubt:**
1. Check official documentation
2. Review existing code patterns in the project
3. Prioritize security and correctness over convenience
4. Ask for clarification rather than making assumptions

**Good luck! Build something amazing! 🚀**
