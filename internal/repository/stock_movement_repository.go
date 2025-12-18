package repository

import (
	"Go-Lang-project-01/internal/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// StockMovementRepository handles stock movement database operations
type StockMovementRepository struct {
	db *gorm.DB
}

// NewStockMovementRepository creates a new stock movement repository
func NewStockMovementRepository(db *gorm.DB) *StockMovementRepository {
	return &StockMovementRepository{db: db}
}

// Create creates a new stock movement record
// Must be called within a database transaction for audit trail
func (r *StockMovementRepository) Create(tx *gorm.DB, movement *models.StockMovement) error {
	if err := tx.Create(movement).Error; err != nil {
		return fmt.Errorf("failed to create stock movement: %w", err)
	}
	return nil
}

// FindByID retrieves a stock movement by ID
func (r *StockMovementRepository) FindByID(id uint) (*models.StockMovement, error) {
	var movement models.StockMovement
	if err := r.db.
		Preload("Product").
		Preload("Product.Category").
		Preload("User").
		Preload("Transaction").
		First(&movement, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("stock movement with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to find stock movement: %w", err)
	}
	return &movement, nil
}

// List retrieves stock movements with pagination and filters
func (r *StockMovementRepository) List(page, limit int, productID *uint, movementType *string, startDate, endDate *time.Time) ([]models.StockMovement, int64, error) {
	var movements []models.StockMovement
	var total int64

	query := r.db.Model(&models.StockMovement{})

	// Apply filters
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	}

	if movementType != nil {
		query = query.Where("type = ?", *movementType)
	}

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count stock movements: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.
		Preload("Product").
		Preload("Product.Category").
		Preload("User").
		Preload("Transaction").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&movements).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list stock movements: %w", err)
	}

	return movements, total, nil
}

// GetByProduct retrieves all stock movements for a product
func (r *StockMovementRepository) GetByProduct(productID uint, limit int) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	query := r.db.Where("product_id = ?", productID).
		Preload("User").
		Preload("Transaction").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&movements).Error; err != nil {
		return nil, fmt.Errorf("failed to get product stock movements: %w", err)
	}

	return movements, nil
}

// GetByTransaction retrieves all stock movements for a transaction
func (r *StockMovementRepository) GetByTransaction(transactionID uint) ([]models.StockMovement, error) {
	var movements []models.StockMovement
	if err := r.db.
		Where("transaction_id = ?", transactionID).
		Preload("Product").
		Preload("Product.Category").
		Order("created_at ASC").
		Find(&movements).Error; err != nil {
		return nil, fmt.Errorf("failed to get transaction stock movements: %w", err)
	}

	return movements, nil
}

// GetTotalMovementByType calculates total stock movement by type for a product
func (r *StockMovementRepository) GetTotalMovementByType(productID uint, movementType string, startDate, endDate *time.Time) (int, error) {
	var total struct {
		Sum int
	}

	query := r.db.Model(&models.StockMovement{}).
		Select("COALESCE(SUM(ABS(quantity)), 0) as sum").
		Where("product_id = ? AND type = ?", productID, movementType)

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	if err := query.Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate total movement: %w", err)
	}

	return total.Sum, nil
}
