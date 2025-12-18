package handlers

import (
	"net/http"
	"strconv"

	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/services"

	"github.com/gin-gonic/gin"
)

type StoreHandler struct {
	service      *services.StoreService
	auditService *services.AuditService
}

type createStoreRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Address     string `json:"address" binding:"omitempty,max=255"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,max=20"`
	Email       string `json:"email" binding:"omitempty,email"`
	IsActive    *bool  `json:"is_active" binding:"omitempty"`
}

type updateStoreRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Address     string `json:"address" binding:"omitempty,max=255"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,max=20"`
	Email       string `json:"email" binding:"omitempty,email"`
	IsActive    *bool  `json:"is_active" binding:"omitempty"`
}

func NewStoreHandler(service *services.StoreService, audit *services.AuditService) *StoreHandler {
	return &StoreHandler{service: service, auditService: audit}
}

func (h *StoreHandler) Create(c *gin.Context) {
	var req createStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	store := &models.Store{
		Name:        req.Name,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		IsActive:    true,
	}
	if req.IsActive != nil {
		store.IsActive = *req.IsActive
	}

	if err := h.service.ValidateStore(store); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	created, err := h.service.CreateStore(store)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Message: err.Error()})
		h.logAudit(c, nil, models.AuditActionStoreCreate, store, false, err.Error())
		return
	}

	c.JSON(http.StatusCreated, models.Response{Success: true, Message: "store created", Data: created})
	resourceID := created.ID
	h.logAudit(c, &resourceID, models.AuditActionStoreCreate, created, true, "")
}

func (h *StoreHandler) List(c *gin.Context) {
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

	stores, total, err := h.service.ListStores(page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	totalPages := (int(total) + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stores,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (h *StoreHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: "invalid ID"})
		return
	}

	store, err := h.service.GetStore(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.Response{Success: true, Data: store})
}

func (h *StoreHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: "invalid ID"})
		return
	}

	var req updateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	payload := &models.Store{
		Name:        req.Name,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		IsActive:    true,
	}
	if req.IsActive != nil {
		payload.IsActive = *req.IsActive
	}

	if err := h.service.ValidateStore(payload); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	updated, err := h.service.UpdateStore(uint(id), payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Message: err.Error()})
		h.logAudit(c, (*uint)(nil), models.AuditActionStoreUpdate, payload, false, err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Response{Success: true, Message: "store updated", Data: updated})
	resourceID := updated.ID
	h.logAudit(c, &resourceID, models.AuditActionStoreUpdate, updated, true, "")
}

func (h *StoreHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: "invalid ID"})
		return
	}

	if err := h.service.DeleteStore(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Message: err.Error()})
		h.logAudit(c, (*uint)(nil), models.AuditActionStoreDelete, nil, false, err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Response{Success: true, Message: "store deleted"})
	resourceID := uint(id)
	h.logAudit(c, &resourceID, models.AuditActionStoreDelete, nil, true, "")
}

func (h *StoreHandler) logAudit(c *gin.Context, resourceID *uint, action models.AuditAction, details interface{}, success bool, errMsg string) {
	userInterface, exists := c.Get("user")
	if !exists {
		return
	}
	user, ok := userInterface.(*models.User)
	if !ok {
		return
	}
	h.auditService.LogAction(c, &user.ID, action, models.AuditResourceStore, resourceID, details, success, errMsg)
}
