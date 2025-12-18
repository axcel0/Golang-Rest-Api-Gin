package services

import (
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type StockService struct {
	productRepo       *repository.ProductRepository
	stockMovementRepo *repository.StockMovementRepository
	db                *gorm.DB
}

func NewStockService(
	productRepo *repository.ProductRepository,
	stockMovementRepo *repository.StockMovementRepository,
	db *gorm.DB,
) *StockService {
	return &StockService{
		productRepo:       productRepo,
		stockMovementRepo: stockMovementRepo,
		db:                db,
	}
}

// StockInRequest represents request for adding stock (restok)
type StockInRequest struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,gt=0"`
	CostPrice float64 `json:"cost_price" binding:"required,gt=0"`
	Notes     string  `json:"notes" binding:"required,min=3"`
}

// StockAdjustRequest represents request for stock adjustment
type StockAdjustRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"` // Can be positive or negative
	Type      string `json:"type" binding:"required,oneof=adjust damage expired return"`
	Notes     string `json:"notes" binding:"required,min=3"`
}

// StockMovementResponse represents stock movement for API response
type StockMovementResponse struct {
	ID            uint                     `json:"id"`
	ProductID     uint                     `json:"product_id"`
	Product       *models.Product          `json:"product,omitempty"`
	UserID        uint                     `json:"user_id"`
	User          *models.User             `json:"user,omitempty"`
	TransactionID *uint                    `json:"transaction_id,omitempty"`
	Type          models.StockMovementType `json:"type"`
	Quantity      int                      `json:"quantity"`
	StockBefore   int                      `json:"stock_before"`
	StockAfter    int                      `json:"stock_after"`
	Notes         string                   `json:"notes"`
	CreatedAt     time.Time                `json:"created_at"`
}

// StockIn adds stock to inventory (restok)
func (s *StockService) StockIn(req StockInRequest, userID uint) (*StockMovementResponse, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get product
	product, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Store stock before
	stockBefore := product.Stock

	// Update product stock
	if err := s.productRepo.UpdateStock(tx, req.ProductID, req.Quantity); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	// Update cost price only (avoid overwriting fresh stock value)
	costPriceDecimal := decimal.NewFromFloat(req.CostPrice)
	if err := tx.Model(&models.Product{}).
		Where("id = ?", req.ProductID).
		Update("cost_price", costPriceDecimal).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update cost price: %w", err)
	}

	// Create stock movement record
	stockAfter := stockBefore + req.Quantity
	movement := &models.StockMovement{
		ProductID:   req.ProductID,
		UserID:      userID,
		Type:        "in",
		Quantity:    req.Quantity, // Positive for stock in
		StockBefore: stockBefore,
		StockAfter:  stockAfter,
		Notes:       req.Notes,
	}

	if err := s.stockMovementRepo.Create(tx, movement); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create stock movement: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload movement with relationships
	movement, err = s.stockMovementRepo.FindByID(movement.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload movement: %w", err)
	}

	return s.toStockMovementResponse(movement), nil
}

// StockAdjust adjusts stock (for damage, expiration, returns, or corrections)
func (s *StockService) StockAdjust(req StockAdjustRequest, userID uint) (*StockMovementResponse, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get product
	product, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Store stock before
	stockBefore := product.Stock
	stockAfter := stockBefore + req.Quantity

	// Validate stock won't go negative
	if stockAfter < 0 {
		tx.Rollback()
		return nil, fmt.Errorf("adjustment would result in negative stock (current: %d, adjustment: %d)", stockBefore, req.Quantity)
	}

	// Update product stock
	if err := s.productRepo.UpdateStock(tx, req.ProductID, req.Quantity); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	// Determine movement type based on request type
	var movementType models.StockMovementType
	switch req.Type {
	case "adjust":
		movementType = "adjust"
	case "damage":
		movementType = "damage"
	case "expired":
		movementType = "expired"
	case "return":
		movementType = "return"
	default:
		movementType = "adjust"
	}

	// Create stock movement record
	movement := &models.StockMovement{
		ProductID:   req.ProductID,
		UserID:      userID,
		Type:        movementType,
		Quantity:    req.Quantity, // Can be positive or negative
		StockBefore: stockBefore,
		StockAfter:  stockAfter,
		Notes:       req.Notes,
	}

	if err := s.stockMovementRepo.Create(tx, movement); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create stock movement: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload movement with relationships
	movement, err = s.stockMovementRepo.FindByID(movement.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload movement: %w", err)
	}

	return s.toStockMovementResponse(movement), nil
}

// GetStockMovements lists stock movements with filters
func (s *StockService) GetStockMovements(page, limit int, productID uint, movementType string, startDate, endDate *time.Time) ([]StockMovementResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Convert parameters to pointers for repository
	var productIDPtr *uint
	if productID > 0 {
		productIDPtr = &productID
	}

	var typeFilterPtr *string
	if movementType != "" {
		typeFilterPtr = &movementType
	}

	movements, total, err := s.stockMovementRepo.List(page, limit, productIDPtr, typeFilterPtr, startDate, endDate)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list movements: %w", err)
	}

	responses := make([]StockMovementResponse, len(movements))
	for i, movement := range movements {
		responses[i] = *s.toStockMovementResponse(&movement)
	}

	return responses, total, nil
}

// GetProductStockHistory gets stock movement history for specific product
func (s *StockService) GetProductStockHistory(productID uint, limit int) ([]StockMovementResponse, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	movements, err := s.stockMovementRepo.GetByProduct(productID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get product history: %w", err)
	}

	responses := make([]StockMovementResponse, len(movements))
	for i, movement := range movements {
		responses[i] = *s.toStockMovementResponse(&movement)
	}

	return responses, nil
}

// Helper function to convert model to response
func (s *StockService) toStockMovementResponse(m *models.StockMovement) *StockMovementResponse {
	response := &StockMovementResponse{
		ID:            m.ID,
		ProductID:     m.ProductID,
		UserID:        m.UserID,
		Type:          m.Type,
		Quantity:      m.Quantity,
		StockBefore:   m.StockBefore,
		StockAfter:    m.StockAfter,
		Notes:         m.Notes,
		CreatedAt:     m.CreatedAt,
		TransactionID: m.TransactionID,
	}

	// Include product if loaded
	if m.Product.ID != 0 {
		response.Product = &m.Product
	}

	// Include user if loaded
	if m.User.ID != 0 {
		// Don't expose password
		user := m.User
		user.Password = ""
		response.User = &user
	}

	return response
}
