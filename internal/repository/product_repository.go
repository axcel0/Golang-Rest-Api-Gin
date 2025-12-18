package repository

import (
	"Go-Lang-project-01/internal/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ProductRepository handles product data operations
type ProductRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create creates a new product
func (r *ProductRepository) Create(product *models.Product) error {
	// Check if SKU already exists
	var existing models.Product
	if err := r.db.Where("sku = ?", product.SKU).First(&existing).Error; err == nil {
		return fmt.Errorf("product with SKU %s already exists", product.SKU)
	}

	// Check if barcode already exists
	if err := r.db.Where("barcode = ?", product.Barcode).First(&existing).Error; err == nil {
		return fmt.Errorf("product with barcode %s already exists", product.Barcode)
	}

	return r.db.Create(product).Error
}

// FindByID retrieves a product by ID
func (r *ProductRepository) FindByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Store").Preload("Category").First(&product, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("product with ID %d not found", id)
		}
		return nil, err
	}
	return &product, nil
}

// FindByBarcode retrieves a product by barcode (for scanner)
func (r *ProductRepository) FindByBarcode(barcode string) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Store").Preload("Category").Where("barcode = ? AND is_active = ?", barcode, true).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("product with barcode %s not found or inactive", barcode)
		}
		return nil, err
	}
	return &product, nil
}

// FindBySKU retrieves a product by SKU
func (r *ProductRepository) FindBySKU(sku string) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Store").Preload("Category").Where("sku = ?", sku).First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("product with SKU %s not found", sku)
		}
		return nil, err
	}
	return &product, nil
}

// List retrieves products with pagination and filtering
func (r *ProductRepository) List(page, limit int, storeID *uint, categoryID *uint, search string, activeOnly bool) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64

	query := r.db.Model(&models.Product{})

	// Apply filters
	if storeID != nil {
		query = query.Where("store_id = ?", *storeID)
	}

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR sku LIKE ? OR barcode LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	err := query.Preload("Store").Preload("Category").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&products).Error

	return products, total, err
}

// Update updates a product
func (r *ProductRepository) Update(product *models.Product) error {
	// Check if product exists
	var existing models.Product
	if err := r.db.First(&existing, product.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("product with ID %d not found", product.ID)
		}
		return err
	}

	// Check SKU uniqueness if changed
	if product.SKU != existing.SKU {
		var duplicate models.Product
		if err := r.db.Where("sku = ? AND id != ?", product.SKU, product.ID).First(&duplicate).Error; err == nil {
			return fmt.Errorf("product with SKU %s already exists", product.SKU)
		}
	}

	return r.db.Save(product).Error
}

// Delete soft deletes a product
func (r *ProductRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Product{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("product with ID %d not found", id)
	}
	return nil
}

// UpdateStock updates product stock atomically
func (r *ProductRepository) UpdateStock(tx *gorm.DB, productID uint, quantity int) error {
	// Use transaction if provided, otherwise use default db
	db := r.db
	if tx != nil {
		db = tx
	}

	// Atomic update with validation
	result := db.Model(&models.Product{}).
		Where("id = ? AND stock + ? >= 0", productID, quantity).
		Update("stock", gorm.Expr("stock + ?", quantity))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("insufficient stock or product not found")
	}

	return nil
}

// GetLowStockProducts retrieves products with low stock
func (r *ProductRepository) GetLowStockProducts(storeID *uint) ([]models.ProductStockAlert, error) {
	var alerts []models.ProductStockAlert

	query := r.db.Table("products").
		Select("products.id as product_id, products.sku, products.name, products.stock, products.min_stock, categories.name as category_name").
		Joins("LEFT JOIN categories ON products.category_id = categories.id").
		Where("products.stock <= products.min_stock AND products.is_active = ?", true).
		Where("products.deleted_at IS NULL")

	if storeID != nil {
		query = query.Where("products.store_id = ?", *storeID)
	}

	err := query.Order("products.stock ASC").Find(&alerts).Error
	return alerts, err
}

// GetProductsByIDs retrieves multiple products by IDs (for checkout)
func (r *ProductRepository) GetProductsByIDs(ids []uint) ([]models.Product, error) {
	var products []models.Product
	err := r.db.Where("id IN ? AND is_active = ?", ids, true).Find(&products).Error
	return products, err
}

// CountByStore counts products by store
func (r *ProductRepository) CountByStore(storeID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Product{}).Where("store_id = ?", storeID).Count(&count).Error
	return count, err
}

// CountByCategory counts products by category
func (r *ProductRepository) CountByCategory(categoryID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Product{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}
