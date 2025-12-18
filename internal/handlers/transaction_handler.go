package handlers

import (
	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// TransactionHandler handles transaction-related HTTP requests
type TransactionHandler struct {
	service *services.TransactionService
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(service *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

// Checkout handles POST /api/v1/transactions
// @Summary Process checkout transaction
// @Tags Transactions
// @Accept json
// @Produce json
// @Param transaction body models.CheckoutRequest true "Checkout data"
// @Success 201 {object} models.Response{data=models.Transaction}
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/transactions [post]
func (h *TransactionHandler) Checkout(c *gin.Context) {
	var req models.CheckoutRequest
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

	// Get user ID from context (set by auth middleware)
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Message: "Invalid user data",
		})
		return
	}

	// Process checkout
	transaction, err := h.service.Checkout(req, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Transaction processed successfully",
		Data:    transaction,
	})
}

// GetTransaction handles GET /api/v1/transactions/:id
// @Summary Get transaction by ID
// @Tags Transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.Response{data=models.Transaction}
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "Invalid transaction ID",
		})
		return
	}

	transaction, err := h.service.GetTransactionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    transaction,
	})
}

// GetTransactionByReceipt handles GET /api/v1/transactions/receipt/:number
// @Summary Get transaction by receipt number
// @Tags Transactions
// @Produce json
// @Param number path string true "Receipt Number"
// @Success 200 {object} models.Response{data=models.Transaction}
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/transactions/receipt/{number} [get]
func (h *TransactionHandler) GetTransactionByReceipt(c *gin.Context) {
	receiptNumber := c.Param("number")
	if receiptNumber == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "Receipt number is required",
		})
		return
	}

	transaction, err := h.service.GetTransactionByReceiptNumber(receiptNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    transaction,
	})
}

// ListTransactions handles GET /api/v1/transactions
// @Summary List transactions with pagination
// @Tags Transactions
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param store_id query int false "Filter by store ID"
// @Param user_id query int false "Filter by user ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} models.PaginatedResponse{data=[]models.Transaction}
// @Security BearerAuth
// @Router /api/v1/transactions [get]
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
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

	var storeID *uint
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if id, err := strconv.ParseUint(storeIDStr, 10, 32); err == nil {
			sid := uint(id)
			storeID = &sid
		}
	}

	var userID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			uid := uint(id)
			userID = &uid
		}
	}

	var startDate, endDate *time.Time
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &t
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			endOfDay := t.Add(24*time.Hour - time.Second)
			endDate = &endOfDay
		}
	}

	transactions, total, err := h.service.ListTransactions(page, limit, storeID, userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve transactions",
		})
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success: true,
		Data:    transactions,
		Pagination: models.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GetReceipt handles GET /api/v1/transactions/:id/receipt
// @Summary Get receipt for printing
// @Tags Transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.Response{data=models.ReceiptResponse}
// @Failure 404 {object} models.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/transactions/{id}/receipt [get]
func (h *TransactionHandler) GetReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Message: "Invalid transaction ID",
		})
		return
	}

	receipt, err := h.service.GenerateReceipt(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Data:    receipt,
	})
}
