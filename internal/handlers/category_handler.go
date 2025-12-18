package handlers

import (
	"net/http"
	"strconv"

	"Go-Lang-project-01/internal/models"
	"Go-Lang-project-01/internal/services"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	service      *services.CategoryService
	auditService *services.AuditService
}

type createCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

type updateCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

func NewCategoryHandler(service *services.CategoryService, audit *services.AuditService) *CategoryHandler {
	return &CategoryHandler{service: service, auditService: audit}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req createCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	cat := &models.Category{Name: req.Name, Description: req.Description}

	if err := h.service.ValidateCategory(cat); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	created, err := h.service.CreateCategory(cat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Message: err.Error()})
		h.logAudit(c, nil, models.AuditActionCategoryCreate, created, false, err.Error())
		return
	}

	c.JSON(http.StatusCreated, models.Response{Success: true, Message: "category created", Data: created})
	resourceID := created.ID
	h.logAudit(c, &resourceID, models.AuditActionCategoryCreate, created, true, "")
}

func (h *CategoryHandler) List(c *gin.Context) {
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

	categories, total, err := h.service.ListCategories(page, limit, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	totalPages := (int(total) + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (h *CategoryHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: "invalid ID"})
		return
	}

	cat, err := h.service.GetCategory(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.Response{Success: true, Data: cat})
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: "invalid ID"})
		return
	}

	var req updateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	payload := &models.Category{Name: req.Name, Description: req.Description}

	if err := h.service.ValidateCategory(payload); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: err.Error()})
		return
	}

	updated, err := h.service.UpdateCategory(uint(id), payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Message: err.Error()})
		h.logAudit(c, (*uint)(nil), models.AuditActionCategoryUpdate, payload, false, err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Response{Success: true, Message: "category updated", Data: updated})
	resourceID := updated.ID
	h.logAudit(c, &resourceID, models.AuditActionCategoryUpdate, updated, true, "")
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Success: false, Message: "invalid ID"})
		return
	}

	if err := h.service.DeleteCategory(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Success: false, Message: err.Error()})
		h.logAudit(c, (*uint)(nil), models.AuditActionCategoryDelete, nil, false, err.Error())
		return
	}

	c.JSON(http.StatusOK, models.Response{Success: true, Message: "category deleted"})
	resourceID := uint(id)
	h.logAudit(c, &resourceID, models.AuditActionCategoryDelete, nil, true, "")
}

func (h *CategoryHandler) logAudit(c *gin.Context, resourceID *uint, action models.AuditAction, details interface{}, success bool, errMsg string) {
	userInterface, exists := c.Get("user")
	if !exists {
		return
	}
	user, ok := userInterface.(*models.User)
	if !ok {
		return
	}
	h.auditService.LogAction(c, &user.ID, action, models.AuditResourceCategory, resourceID, details, success, errMsg)
}
