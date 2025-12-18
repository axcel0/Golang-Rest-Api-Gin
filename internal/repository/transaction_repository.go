package repository

import (
	"Go-Lang-project-01/internal/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TransactionRepository handles transaction database operations
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create creates a new transaction with items
// Must be called within a database transaction
func (r *TransactionRepository) Create(tx *gorm.DB, transaction *models.Transaction) error {
	if err := tx.Create(transaction).Error; err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

// FindByID retrieves a transaction by ID with all relationships
func (r *TransactionRepository) FindByID(id uint) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := r.db.
		Preload("User").
		Preload("Store").
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Category").
		First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}
	return &transaction, nil
}

// FindByReceiptNumber retrieves a transaction by receipt number
func (r *TransactionRepository) FindByReceiptNumber(receiptNumber string) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := r.db.
		Preload("User").
		Preload("Store").
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Category").
		Where("receipt_number = ?", receiptNumber).
		First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction with receipt number %s not found", receiptNumber)
		}
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}
	return &transaction, nil
}

// List retrieves transactions with pagination and filters
func (r *TransactionRepository) List(page, limit int, storeID *uint, userID *uint, startDate, endDate *time.Time) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.db.Model(&models.Transaction{})

	// Apply filters
	if storeID != nil {
		query = query.Where("store_id = ?", *storeID)
	}

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.
		Preload("User").
		Preload("Store").
		Preload("Items").
		Preload("Items.Product").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	return transactions, total, nil
}

// GetDailySales retrieves total sales for a specific date and store
func (r *TransactionRepository) GetDailySales(storeID uint, date time.Time) (int64, error) {
	var count int64
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.db.Model(&models.Transaction{}).
		Where("store_id = ? AND created_at >= ? AND created_at < ?", storeID, startOfDay, endOfDay).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get daily sales count: %w", err)
	}

	return count, nil
}

// GetTotalRevenue calculates total revenue for a store within date range
func (r *TransactionRepository) GetTotalRevenue(storeID uint, startDate, endDate time.Time) (string, error) {
	var total struct {
		Sum string
	}

	if err := r.db.Model(&models.Transaction{}).
		Select("COALESCE(SUM(total_amount), '0') as sum").
		Where("store_id = ? AND created_at >= ? AND created_at <= ?", storeID, startDate, endDate).
		Scan(&total).Error; err != nil {
		return "0", fmt.Errorf("failed to calculate total revenue: %w", err)
	}

	return total.Sum, nil
}

// CountByStore counts transactions by store
func (r *TransactionRepository) CountByStore(storeID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Transaction{}).
		Where("store_id = ?", storeID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}
	return count, nil
}

// CountByUser counts transactions by user (kasir)
func (r *TransactionRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Transaction{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count user transactions: %w", err)
	}
	return count, nil
}
