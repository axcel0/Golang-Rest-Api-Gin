package handlers

import (
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	service *services.ProductService
}

// NewProductHandler creates a new product handler
func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// CreateProduct handles POST /api/v1/products
// @Summary Create a new product
// @Tags Products
// @Accept json
// @Produce json
// @Param product body models.CreateProductRequest true "Product data"
// @Success 201 {object} models.Response{data=models.Product}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "Invalid request data",
			Errors: []models.ValidationError{
				{
					Field:   "request",
					Message: err.Error(),
				},
			},
		})
		return
	}

	product, err := h.service.CreateProduct(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Product created successfully",
		Data:    product,
	})
}

// GetProduct handles GET /api/v1/products/:id
// @Summary Get product by ID
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} models.Response{data=models.Product}
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "Invalid product ID",
		})
		return
	}

	product, err := h.service.GetProductByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    product,
	})
}

// GetProductByBarcode handles GET /api/v1/products/by-barcode/:barcode
// @Summary Get product by barcode (for scanner)
// @Tags Products
// @Produce json
// @Param barcode path string true "Product Barcode"
// @Success 200 {object} models.Response{data=models.Product}
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/products/by-barcode/{barcode} [get]
func (h *ProductHandler) GetProductByBarcode(c *gin.Context) {
	barcode := c.Param("barcode")
	if barcode == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "Barcode is required",
		})
		return
	}

	product, err := h.service.GetProductByBarcode(barcode)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    product,
	})
}

// ListProducts handles GET /api/v1/products
// @Summary List products with pagination
// @Tags Products
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param store_id query int false "Filter by store ID"
// @Param category_id query int false "Filter by category ID"
// @Param search query string false "Search by name, SKU, or barcode"
// @Param active_only query bool false "Show only active products" default(true)
// @Success 200 {object} models.PaginatedResponse{data=[]models.Product}
// @Security BearerAuth
// @Router /api/v1/products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	search := c.Query("search")
	activeOnly := c.DefaultQuery("active_only", "true") == "true"

	var storeID *uint
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if id, err := strconv.ParseUint(storeIDStr, 10, 32); err == nil {
			sid := uint(id)
			storeID = &sid
		}
	}

	var categoryID *uint
	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if id, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
			cid := uint(id)
			categoryID = &cid
		}
	}

	products, total, err := h.service.ListProducts(page, limit, storeID, categoryID, search, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve products",
		})
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success: true,
		Data:    products,
		Pagination: models.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// UpdateProduct handles PUT /api/v1/products/:id
// @Summary Update product
// @Tags Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body models.UpdateProductRequest true "Product update data"
// @Success 200 {object} models.Response{data=models.Product}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "Invalid product ID",
		})
		return
	}

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "Invalid request data",
			Errors: []models.ValidationError{
				{
					Field:   "request",
					Message: err.Error(),
				},
			},
		})
		return
	}

	product, err := h.service.UpdateProduct(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Product updated successfully",
		Data:    product,
	})
}

// DeleteProduct handles DELETE /api/v1/products/:id
// @Summary Delete product (soft delete)
// @Tags Products
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} models.Response
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "Invalid product ID",
		})
		return
	}

	if err := h.service.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Product deleted successfully",
	})
}

// GetLowStockAlerts handles GET /api/v1/products/low-stock
// @Summary Get products with low stock
// @Tags Products
// @Produce json
// @Param store_id query int false "Filter by store ID"
// @Success 200 {object} models.Response{data=[]models.ProductStockAlert}
// @Security BearerAuth
// @Router /api/v1/products/low-stock [get]
func (h *ProductHandler) GetLowStockAlerts(c *gin.Context) {
	var storeID *uint
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if id, err := strconv.ParseUint(storeIDStr, 10, 32); err == nil {
			sid := uint(id)
			storeID = &sid
		}
	}

	alerts, err := h.service.GetLowStockAlerts(storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve low stock alerts",
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    alerts,
	})
}
