package services

import (
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TransactionService handles transaction business logic
type TransactionService struct {
	transactionRepo   *repository.TransactionRepository
	productRepo       *repository.ProductRepository
	stockMovementRepo *repository.StockMovementRepository
	db                *gorm.DB
}

// NewTransactionService creates a new transaction service
func NewTransactionService(
	transactionRepo *repository.TransactionRepository,
	productRepo *repository.ProductRepository,
	stockMovementRepo *repository.StockMovementRepository,
	db *gorm.DB,
) *TransactionService {
	return &TransactionService{
		transactionRepo:   transactionRepo,
		productRepo:       productRepo,
		stockMovementRepo: stockMovementRepo,
		db:                db,
	}
}

// Checkout processes a transaction with atomic stock reduction
func (s *TransactionService) Checkout(req models.CheckoutRequest, userID uint) (*models.Transaction, error) {
	// Start database transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Validate and get all products
	productIDs := make([]uint, len(req.Items))
	for i, item := range req.Items {
		productIDs[i] = item.ProductID
	}

	products, err := s.productRepo.GetProductsByIDs(productIDs)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create product map for quick lookup
	productMap := make(map[uint]*models.Product)
	for i := range products {
		productMap[products[i].ID] = &products[i]
	}

	// 2. Validate stock availability and calculate totals
	var subtotal decimal.Decimal = decimal.Zero
	transactionItems := make([]models.TransactionItem, 0, len(req.Items))

	for _, item := range req.Items {
		product, exists := productMap[item.ProductID]
		if !exists {
			tx.Rollback()
			return nil, fmt.Errorf("product with ID %d not found", item.ProductID)
		}

		// Check stock availability
		if product.Stock < item.Quantity {
			tx.Rollback()
			return nil, fmt.Errorf("insufficient stock for product %s (available: %d, requested: %d)",
				product.Name, product.Stock, item.Quantity)
		}

		// Check if product is active
		if !product.IsActive {
			tx.Rollback()
			return nil, fmt.Errorf("product %s is not active", product.Name)
		}

		// Calculate item total
		price := product.BasePrice
		quantity := decimal.NewFromInt(int64(item.Quantity))
		itemSubtotal := price.Mul(quantity)
		subtotal = subtotal.Add(itemSubtotal)

		// Create transaction item with snapshot data
		transactionItems = append(transactionItems, models.TransactionItem{
			ProductID:   product.ID,
			ProductSKU:  product.SKU,
			ProductName: product.Name,
			ProductUnit: product.Unit,
			Price:       price,             // Price snapshot at transaction time
			CostPrice:   product.CostPrice, // Cost snapshot for profit calculation
			Quantity:    item.Quantity,
			Subtotal:    itemSubtotal,
			Discount:    decimal.Zero, // Item-level discount (for future)
			Total:       itemSubtotal, // Total after item discount
		})
	}

	// 3. Calculate tax, discount and final total
	tax := decimal.NewFromFloat(req.Tax)
	if tax.IsNegative() {
		tx.Rollback()
		return nil, fmt.Errorf("tax cannot be negative")
	}

	discount := decimal.NewFromFloat(req.Discount)
	if discount.IsNegative() {
		tx.Rollback()
		return nil, fmt.Errorf("discount cannot be negative")
	}

	if discount.GreaterThan(subtotal) {
		tx.Rollback()
		return nil, fmt.Errorf("discount cannot exceed subtotal")
	}

	// Total = Subtotal + Tax - Discount
	total := subtotal.Add(tax).Sub(discount)

	// 4. Validate payment (only for cash)
	cashReceived := decimal.NewFromFloat(req.CashReceived)
	cashChange := decimal.Zero

	if req.PaymentMethod == "cash" {
		if cashReceived.LessThan(total) {
			tx.Rollback()
			return nil, fmt.Errorf("insufficient cash payment (required: %s, received: %s)",
				total.String(), cashReceived.String())
		}
		cashChange = cashReceived.Sub(total)
	}

	// 5. Generate receipt number (format: RCP-YYYYMMDD-XXXXX)
	receiptNumber := fmt.Sprintf("RCP-%s-%05d",
		time.Now().Format("20060102"),
		time.Now().Unix()%100000,
	)

	// 6. Create transaction
	paymentMethod := models.PaymentMethod(req.PaymentMethod)
	transaction := &models.Transaction{
		StoreID:       req.StoreID,
		UserID:        userID,
		ReceiptNumber: receiptNumber,
		Status:        models.TransactionStatusCompleted, // Auto-complete after checkout
		Subtotal:      subtotal,
		Tax:           tax,
		Discount:      discount,
		Total:         total,
		CashReceived:  cashReceived,
		CashChange:    cashChange,
		PaymentMethod: paymentMethod,
		Notes:         req.Notes,
		Items:         transactionItems,
	}

	now := time.Now()
	transaction.CompletedAt = &now

	if err := s.transactionRepo.Create(tx, transaction); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// 7. Reduce stock atomically and create audit trail
	for _, item := range transactionItems {
		// Atomic stock reduction
		if err := s.productRepo.UpdateStock(tx, item.ProductID, -item.Quantity); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update stock for product ID %d: %w", item.ProductID, err)
		}

		// Create stock movement audit record
		stockMovement := &models.StockMovement{
			ProductID:     item.ProductID,
			TransactionID: &transaction.ID,
			UserID:        userID,
			Type:          "sale",
			Quantity:      -item.Quantity, // Negative for sale
			Notes:         fmt.Sprintf("Sale via transaction %s", receiptNumber),
		}

		if err := s.stockMovementRepo.Create(tx, stockMovement); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create stock movement: %w", err)
		}
	}

	// 8. Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 9. Reload transaction with all relationships
	result, err := s.transactionRepo.FindByID(transaction.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload transaction: %w", err)
	}

	return result, nil
}

// GetTransactionByID retrieves a transaction by ID
func (s *TransactionService) GetTransactionByID(id uint) (*models.Transaction, error) {
	return s.transactionRepo.FindByID(id)
}

// GetTransactionByReceiptNumber retrieves a transaction by receipt number
func (s *TransactionService) GetTransactionByReceiptNumber(receiptNumber string) (*models.Transaction, error) {
	return s.transactionRepo.FindByReceiptNumber(receiptNumber)
}

// ListTransactions retrieves transactions with filters
func (s *TransactionService) ListTransactions(page, limit int, storeID *uint, userID *uint, startDate, endDate *time.Time) ([]models.Transaction, int64, error) {
	// Default pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.transactionRepo.List(page, limit, storeID, userID, startDate, endDate)
}

// GenerateReceipt generates receipt data for printing
func (s *TransactionService) GenerateReceipt(transactionID uint) (*models.ReceiptResponse, error) {
	transaction, err := s.transactionRepo.FindByID(transactionID)
	if err != nil {
		return nil, err
	}

	// Map transaction items to receipt items
	items := make([]models.ReceiptItemResponse, len(transaction.Items))
	for i, item := range transaction.Items {
		items[i] = models.ReceiptItemResponse{
			Name:     item.ProductName,
			Quantity: item.Quantity,
			Unit:     item.ProductUnit,
			Price:    item.Price.String(),
			Discount: item.Discount.String(),
			Total:    item.Total.String(),
		}
	}

	receipt := &models.ReceiptResponse{
		ReceiptNumber: transaction.ReceiptNumber,
		StoreName:     transaction.Store.Name,
		StoreAddress:  transaction.Store.Address,
		StorePhone:    transaction.Store.PhoneNumber,
		KasirName:     transaction.User.Name,
		TransactionID: transaction.ID,
		Date:          transaction.CreatedAt,
		Items:         items,
		Subtotal:      transaction.Subtotal.String(),
		Tax:           transaction.Tax.String(),
		Discount:      transaction.Discount.String(),
		Total:         transaction.Total.String(),
		PaymentMethod: string(transaction.PaymentMethod),
		CashReceived:  transaction.CashReceived.String(),
		CashChange:    transaction.CashChange.String(),
		Notes:         transaction.Notes,
	}

	return receipt, nil
}
