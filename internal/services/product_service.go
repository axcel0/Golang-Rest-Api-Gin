package services

import (
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/repository"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ProductService handles product business logic
type ProductService struct {
	repo *repository.ProductRepository
	db   *gorm.DB
}

// NewProductService creates a new product service
func NewProductService(repo *repository.ProductRepository, db *gorm.DB) *ProductService {
	return &ProductService{
		repo: repo,
		db:   db,
	}
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(req models.CreateProductRequest) (*models.Product, error) {
	// Convert float64 to decimal for precise calculations
	basePrice := decimal.NewFromFloat(req.BasePrice)
	costPrice := decimal.NewFromFloat(req.CostPrice)

	// Business validation
	// Allow products with base price < cost price (loss leaders)

	if costPrice.IsNegative() {
		return nil, fmt.Errorf("cost price cannot be negative")
	}

	if basePrice.IsNegative() {
		return nil, fmt.Errorf("base price cannot be negative")
	}

	product := &models.Product{
		StoreID:     req.StoreID,
		CategoryID:  req.CategoryID,
		SKU:         req.SKU,
		Barcode:     req.Barcode,
		Name:        req.Name,
		Description: req.Description,
		Unit:        req.Unit,
		BasePrice:   basePrice,
		CostPrice:   costPrice,
		Stock:       req.Stock,
		MinStock:    req.MinStock,
		ImageURL:    req.ImageURL,
		IsActive:    true,
	}

	if err := s.repo.Create(product); err != nil {
		return nil, err
	}

	return product, nil
}

// GetProductByID retrieves a product by ID
func (s *ProductService) GetProductByID(id uint) (*models.Product, error) {
	return s.repo.FindByID(id)
}

// GetProductByBarcode retrieves a product by barcode (for scanner)
func (s *ProductService) GetProductByBarcode(barcode string) (*models.Product, error) {
	return s.repo.FindByBarcode(barcode)
}

// GetProductBySKU retrieves a product by SKU
func (s *ProductService) GetProductBySKU(sku string) (*models.Product, error) {
	return s.repo.FindBySKU(sku)
}

// ListProducts retrieves products with pagination
func (s *ProductService) ListProducts(page, limit int, storeID *uint, categoryID *uint, search string, activeOnly bool) ([]models.Product, int64, error) {
	// Default pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	return s.repo.List(page, limit, storeID, categoryID, search, activeOnly)
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(id uint, req models.UpdateProductRequest) (*models.Product, error) {
	// Get existing product
	product, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.CategoryID != nil {
		product.CategoryID = *req.CategoryID
	}

	if req.Name != nil {
		product.Name = *req.Name
	}

	if req.Description != nil {
		product.Description = *req.Description
	}

	if req.Unit != nil {
		product.Unit = *req.Unit
	}

	if req.BasePrice != nil {
		basePrice := decimal.NewFromFloat(*req.BasePrice)
		if basePrice.IsNegative() {
			return nil, fmt.Errorf("base price cannot be negative")
		}
		product.BasePrice = basePrice
	}

	if req.CostPrice != nil {
		costPrice := decimal.NewFromFloat(*req.CostPrice)
		if costPrice.IsNegative() {
			return nil, fmt.Errorf("cost price cannot be negative")
		}
		product.CostPrice = costPrice
	}

	if req.MinStock != nil {
		product.MinStock = *req.MinStock
	}

	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if req.ImageURL != nil {
		product.ImageURL = *req.ImageURL
	}

	// Save updates
	if err := s.repo.Update(product); err != nil {
		return nil, err
	}

	return product, nil
}

// DeleteProduct soft deletes a product
func (s *ProductService) DeleteProduct(id uint) error {
	return s.repo.Delete(id)
}

// GetLowStockAlerts retrieves products with low stock
func (s *ProductService) GetLowStockAlerts(storeID *uint) ([]models.ProductStockAlert, error) {
	return s.repo.GetLowStockProducts(storeID)
}

// ValidateProductsForCheckout validates products exist and have sufficient stock
func (s *ProductService) ValidateProductsForCheckout(items []models.CheckoutItemRequest) ([]models.Product, error) {
	// Extract product IDs
	productIDs := make([]uint, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	// Fetch products
	products, err := s.repo.GetProductsByIDs(productIDs)
	if err != nil {
		return nil, err
	}

	// Create map for quick lookup
	productMap := make(map[uint]*models.Product)
	for i := range products {
		productMap[products[i].ID] = &products[i]
	}

	// Validate each item
	for _, item := range items {
		product, exists := productMap[item.ProductID]
		if !exists {
			return nil, fmt.Errorf("product with ID %d not found or inactive", item.ProductID)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s (available: %d, requested: %d)",
				product.Name, product.Stock, item.Quantity)
		}
	}

	return products, nil
}
