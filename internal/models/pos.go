package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Store represents a store/outlet in the POS system
// Future-proof for multi-store support
type Store struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name" binding:"required,min=3,max=100"`
	Address     string         `gorm:"type:text" json:"address,omitempty"`
	PhoneNumber string         `gorm:"type:varchar(20)" json:"phone_number,omitempty"`
	Email       string         `gorm:"type:varchar(100)" json:"email,omitempty"`
	IsActive    bool           `gorm:"default:true;not null" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Products     []Product     `gorm:"foreignKey:StoreID" json:"products,omitempty"`
	Transactions []Transaction `gorm:"foreignKey:StoreID" json:"transactions,omitempty"`
}

// Category represents a product category
type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null;uniqueIndex" json:"name" binding:"required,min=2,max=100"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Products []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// Product represents a product in the POS system
type Product struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	StoreID     uint            `gorm:"not null;index" json:"store_id" binding:"required"`
	CategoryID  uint            `gorm:"not null;index" json:"category_id" binding:"required"`
	SKU         string          `gorm:"type:varchar(50);uniqueIndex;not null" json:"sku" binding:"required,min=3,max=50"`
	Barcode     string          `gorm:"type:varchar(100);uniqueIndex;not null" json:"barcode" binding:"required"`
	Name        string          `gorm:"type:varchar(200);not null;index" json:"name" binding:"required,min=3,max=200"`
	Description string          `gorm:"type:text" json:"description,omitempty"`
	Unit        string          `gorm:"type:varchar(20);not null" json:"unit" binding:"required"` // pcs, kg, liter, box, dll
	BasePrice   decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"base_price" binding:"required,gt=0"`
	CostPrice   decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"cost_price" binding:"required,gt=0"`
	Stock       int             `gorm:"not null;default:0" json:"stock" binding:"gte=0"`
	MinStock    int             `gorm:"not null;default:0" json:"min_stock" binding:"gte=0"` // Low stock alert threshold
	IsActive    bool            `gorm:"default:true;not null" json:"is_active"`
	ImageURL    string          `gorm:"type:varchar(255)" json:"image_url,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`

	// Relationships
	Store            Store             `gorm:"foreignKey:StoreID;references:ID" json:"store,omitempty"`
	Category         Category          `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"`
	TransactionItems []TransactionItem `gorm:"foreignKey:ProductID" json:"transaction_items,omitempty"`
	StockMovements   []StockMovement   `gorm:"foreignKey:ProductID" json:"stock_movements,omitempty"`
}

// BeforeCreate hook to validate product before creation
func (p *Product) BeforeCreate(_ *gorm.DB) error {
	// Ensure cost price is not greater than base price (normally)
	// Allow products with base price < cost price (loss leaders)

	// Ensure stock is not negative
	if p.Stock < 0 {
		p.Stock = 0
	}

	return nil
}

// IsLowStock checks if product stock is below minimum threshold
func (p *Product) IsLowStock() bool {
	return p.Stock <= p.MinStock
}

// CalculateProfit calculates profit per unit
func (p *Product) CalculateProfit() decimal.Decimal {
	return p.BasePrice.Sub(p.CostPrice)
}

// CalculateProfitMargin calculates profit margin percentage
func (p *Product) CalculateProfitMargin() decimal.Decimal {
	if p.BasePrice.IsZero() {
		return decimal.Zero
	}
	profit := p.CalculateProfit()
	return profit.Div(p.BasePrice).Mul(decimal.NewFromInt(100))
}

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

// PaymentMethod represents payment method
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodDebitCard    PaymentMethod = "debit_card"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodEWallet      PaymentMethod = "e_wallet"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

// Transaction represents a sales transaction (struk penjualan)
type Transaction struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	StoreID       uint              `gorm:"not null;index" json:"store_id"`
	UserID        uint              `gorm:"not null;index" json:"user_id"` // Kasir who created this transaction
	ReceiptNumber string            `gorm:"type:varchar(50);uniqueIndex;not null" json:"receipt_number"`
	Status        TransactionStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	PaymentMethod PaymentMethod     `gorm:"type:varchar(30);not null" json:"payment_method" binding:"required"`
	Subtotal      decimal.Decimal   `gorm:"type:decimal(15,2);not null" json:"subtotal"`
	Tax           decimal.Decimal   `gorm:"type:decimal(15,2);not null;default:0" json:"tax"`
	Discount      decimal.Decimal   `gorm:"type:decimal(15,2);not null;default:0" json:"discount"`
	Total         decimal.Decimal   `gorm:"type:decimal(15,2);not null" json:"total"`
	CashReceived  decimal.Decimal   `gorm:"type:decimal(15,2)" json:"cash_received,omitempty"` // For cash payment
	CashChange    decimal.Decimal   `gorm:"type:decimal(15,2)" json:"cash_change,omitempty"`   // For cash payment
	Notes         string            `gorm:"type:text" json:"notes,omitempty"`
	CompletedAt   *time.Time        `json:"completed_at,omitempty"`
	CancelledAt   *time.Time        `json:"cancelled_at,omitempty"`
	CancelReason  string            `gorm:"type:text" json:"cancel_reason,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `gorm:"index" json:"-"`

	// Relationships
	Store Store             `gorm:"foreignKey:StoreID;references:ID" json:"store,omitempty"`
	User  User              `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"` // Kasir
	Items []TransactionItem `gorm:"foreignKey:TransactionID" json:"items,omitempty"`
}

// BeforeCreate hook to generate receipt number
func (t *Transaction) BeforeCreate(_ *gorm.DB) error {
	if t.ReceiptNumber == "" {
		// Generate receipt number: TRX-YYYYMMDD-XXXXXX (timestamp-based)
		t.ReceiptNumber = "TRX-" + time.Now().Format("20060102-150405")
	}

	// Set default status
	if t.Status == "" {
		t.Status = TransactionStatusPending
	}

	return nil
}

// MarkAsCompleted marks transaction as completed
func (t *Transaction) MarkAsCompleted(tx *gorm.DB) error {
	now := time.Now()
	t.Status = TransactionStatusCompleted
	t.CompletedAt = &now
	return tx.Save(t).Error
}

// MarkAsCancelled marks transaction as cancelled
func (t *Transaction) MarkAsCancelled(tx *gorm.DB, reason string) error {
	now := time.Now()
	t.Status = TransactionStatusCancelled
	t.CancelledAt = &now
	t.CancelReason = reason
	return tx.Save(t).Error
}

// CalculateTotal calculates transaction total
func (t *Transaction) CalculateTotal() decimal.Decimal {
	// Total = Subtotal + Tax - Discount
	return t.Subtotal.Add(t.Tax).Sub(t.Discount)
}

// CalculateChange calculates change for cash payment
func (t *Transaction) CalculateChange() decimal.Decimal {
	if t.PaymentMethod != PaymentMethodCash {
		return decimal.Zero
	}
	return t.CashReceived.Sub(t.Total)
}

// TransactionItem represents an item in a transaction
// Stores snapshot of product data at time of purchase
type TransactionItem struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	TransactionID uint            `gorm:"not null;index" json:"transaction_id"`
	ProductID     uint            `gorm:"not null;index" json:"product_id"`
	ProductSKU    string          `gorm:"type:varchar(50);not null" json:"product_sku"`          // Snapshot
	ProductName   string          `gorm:"type:varchar(200);not null" json:"product_name"`        // Snapshot
	ProductUnit   string          `gorm:"type:varchar(20);not null" json:"product_unit"`         // Snapshot
	Price         decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"price"`              // Price at time of sale (snapshot)
	CostPrice     decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"cost_price"`         // Cost at time of sale (for profit calc)
	Quantity      int             `gorm:"not null" json:"quantity" binding:"required,gt=0"`      // Quantity sold
	Subtotal      decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"subtotal"`           // Price * Quantity
	Discount      decimal.Decimal `gorm:"type:decimal(15,2);not null;default:0" json:"discount"` // Item-level discount
	Total         decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"total"`              // Subtotal - Discount
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`

	// Relationships
	Transaction Transaction `gorm:"foreignKey:TransactionID;references:ID" json:"transaction,omitempty"`
	Product     Product     `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
}

// BeforeCreate hook to calculate totals
func (ti *TransactionItem) BeforeCreate(_ *gorm.DB) error {
	// Calculate subtotal
	ti.Subtotal = ti.Price.Mul(decimal.NewFromInt(int64(ti.Quantity)))

	// Calculate total (subtotal - discount)
	ti.Total = ti.Subtotal.Sub(ti.Discount)

	return nil
}

// CalculateProfit calculates profit for this transaction item
func (ti *TransactionItem) CalculateProfit() decimal.Decimal {
	profitPerUnit := ti.Price.Sub(ti.CostPrice)
	return profitPerUnit.Mul(decimal.NewFromInt(int64(ti.Quantity)))
}

// StockMovementType represents type of stock movement
type StockMovementType string

const (
	StockMovementTypeIn         StockMovementType = "in"         // Restok masuk
	StockMovementTypeOut        StockMovementType = "out"        // Keluar (manual)
	StockMovementTypeSale       StockMovementType = "sale"       // Terjual via transaction
	StockMovementTypeAdjustment StockMovementType = "adjustment" // Penyesuaian stock
	StockMovementTypeDamage     StockMovementType = "damage"     // Rusak/hilang
	StockMovementTypeReturn     StockMovementType = "return"     // Return dari customer
	StockMovementTypeTransfer   StockMovementType = "transfer"   // Transfer antar gudang/toko
)

// StockMovement represents stock movement audit trail (immutable record)
type StockMovement struct {
	ID            uint              `gorm:"primaryKey" json:"id"`
	ProductID     uint              `gorm:"not null;index" json:"product_id"`
	UserID        uint              `gorm:"not null;index" json:"user_id"`         // User who performed the action
	TransactionID *uint             `gorm:"index" json:"transaction_id,omitempty"` // Optional: linked transaction
	Type          StockMovementType `gorm:"type:varchar(20);not null;index" json:"type"`
	Quantity      int               `gorm:"not null" json:"quantity"`                         // Positive for IN, negative for OUT
	StockBefore   int               `gorm:"not null" json:"stock_before"`                     // Stock before movement
	StockAfter    int               `gorm:"not null" json:"stock_after"`                      // Stock after movement
	Notes         string            `gorm:"type:text" json:"notes,omitempty"`                 // Reason/notes
	ReferenceDoc  string            `gorm:"type:varchar(100)" json:"reference_doc,omitempty"` // PO number, DO number, etc
	CreatedAt     time.Time         `json:"created_at"`

	// Relationships (no updates/deletes for audit trail)
	Product     Product      `gorm:"foreignKey:ProductID;references:ID" json:"product,omitempty"`
	User        User         `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Transaction *Transaction `gorm:"foreignKey:TransactionID;references:ID" json:"transaction,omitempty"`
}

// Note: StockMovement is immutable (audit trail)
// No UpdatedAt or DeletedAt - once created, never modified or deleted

// CreateStockMovement creates a stock movement record
func CreateStockMovement(tx *gorm.DB, product *Product, userID uint, movementType StockMovementType, quantity int, notes string, transactionID *uint) error {
	stockBefore := product.Stock
	stockAfter := stockBefore + quantity // quantity can be negative

	movement := StockMovement{
		ProductID:     product.ID,
		UserID:        userID,
		TransactionID: transactionID,
		Type:          movementType,
		Quantity:      quantity,
		StockBefore:   stockBefore,
		StockAfter:    stockAfter,
		Notes:         notes,
	}

	return tx.Create(&movement).Error
}

// --- Request/Response DTOs ---

// CreateStoreRequest represents request to create a store
type CreateStoreRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Address     string `json:"address,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Email       string `json:"email,omitempty" binding:"omitempty,email"`
}

// CreateCategoryRequest represents request to create a category
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description,omitempty"`
}

// CreateProductRequest represents request to create a product
type CreateProductRequest struct {
	StoreID     uint    `json:"store_id" binding:"required"`
	CategoryID  uint    `json:"category_id" binding:"required"`
	SKU         string  `json:"sku" binding:"required,min=3,max=50"`
	Barcode     string  `json:"barcode" binding:"required"`
	Name        string  `json:"name" binding:"required,min=3,max=200"`
	Description string  `json:"description,omitempty"`
	Unit        string  `json:"unit" binding:"required"`
	BasePrice   float64 `json:"base_price" binding:"required,gt=0"`
	CostPrice   float64 `json:"cost_price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"gte=0"`
	MinStock    int     `json:"min_stock" binding:"gte=0"`
	ImageURL    string  `json:"image_url,omitempty"`
}

// UpdateProductRequest represents request to update a product
type UpdateProductRequest struct {
	CategoryID  *uint    `json:"category_id,omitempty"`
	Name        *string  `json:"name,omitempty" binding:"omitempty,min=3,max=200"`
	Description *string  `json:"description,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	BasePrice   *float64 `json:"base_price,omitempty" binding:"omitempty,gt=0"`
	CostPrice   *float64 `json:"cost_price,omitempty" binding:"omitempty,gt=0"`
	MinStock    *int     `json:"min_stock,omitempty" binding:"omitempty,gte=0"`
	IsActive    *bool    `json:"is_active,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty"`
}

// StockAdjustmentRequest represents request to adjust stock
type StockAdjustmentRequest struct {
	Quantity     int    `json:"quantity" binding:"required"` // Can be positive (in) or negative (out)
	Type         string `json:"type" binding:"required,oneof=in out adjustment damage return transfer"`
	Notes        string `json:"notes,omitempty"`
	ReferenceDoc string `json:"reference_doc,omitempty"`
}

// CheckoutItemRequest represents an item in checkout request
type CheckoutItemRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	Discount  float64 `json:"discount,omitempty" binding:"gte=0"` // Item-level discount
}

// CheckoutRequest represents checkout request
type CheckoutRequest struct {
	StoreID       uint                  `json:"store_id" binding:"required"`
	PaymentMethod string                `json:"payment_method" binding:"required,oneof=cash debit_card credit_card e_wallet bank_transfer"`
	Items         []CheckoutItemRequest `json:"items" binding:"required,min=1,dive"`
	Tax           float64               `json:"tax,omitempty" binding:"gte=0"`
	Discount      float64               `json:"discount,omitempty" binding:"gte=0"`      // Transaction-level discount
	CashReceived  float64               `json:"cash_received,omitempty" binding:"gte=0"` // Required if payment_method = cash
	Notes         string                `json:"notes,omitempty"`
}

// ReceiptResponse represents receipt data for printing
type ReceiptResponse struct {
	ReceiptNumber string                `json:"receipt_number"`
	StoreName     string                `json:"store_name"`
	StoreAddress  string                `json:"store_address,omitempty"`
	StorePhone    string                `json:"store_phone,omitempty"`
	KasirName     string                `json:"kasir_name"`
	TransactionID uint                  `json:"transaction_id"`
	Date          time.Time             `json:"date"`
	Items         []ReceiptItemResponse `json:"items"`
	Subtotal      string                `json:"subtotal"` // Formatted decimal string
	Tax           string                `json:"tax"`
	Discount      string                `json:"discount"`
	Total         string                `json:"total"`
	PaymentMethod string                `json:"payment_method"`
	CashReceived  string                `json:"cash_received,omitempty"`
	CashChange    string                `json:"cash_change,omitempty"`
	Notes         string                `json:"notes,omitempty"`
}

// ReceiptItemResponse represents an item in receipt
type ReceiptItemResponse struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit"`
	Price    string `json:"price"`    // Formatted
	Discount string `json:"discount"` // Formatted
	Total    string `json:"total"`    // Formatted
}

// ProductStockAlert represents product with low stock
type ProductStockAlert struct {
	ProductID    uint   `json:"product_id"`
	SKU          string `json:"sku"`
	Name         string `json:"name"`
	Stock        int    `json:"stock"`
	MinStock     int    `json:"min_stock"`
	CategoryName string `json:"category_name"`
}
