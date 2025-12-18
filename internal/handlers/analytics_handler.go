package handlers

import (
	"net/http"
	"strconv"
	"time"

	"Go-Lang-project-01/internal/services"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	service *services.AnalyticsService
}

func NewAnalyticsHandler(service *services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: service}
}

func (h *AnalyticsHandler) DailySummary(c *gin.Context) {
	dateStr := c.Query("date")
	loc := time.Local
	if dateStr == "" {
		dateStr = time.Now().In(loc).Format("2006-01-02")
	}
	date, err := time.ParseInLocation("2006-01-02", dateStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid date format, use YYYY-MM-DD"})
		return
	}

	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, loc)
	end := start.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	summary, err := h.service.GetSummary(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": summary})
}

func (h *AnalyticsHandler) RangeSummary(c *gin.Context) {
	start, end, ok := parseDateRange(c)
	if !ok {
		return
	}

	summary, err := h.service.GetSummary(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": summary})
}

func (h *AnalyticsHandler) PaymentBreakdown(c *gin.Context) {
	start, end, ok := parseDateRange(c)
	if !ok {
		return
	}

	data, err := h.service.GetPaymentBreakdown(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func (h *AnalyticsHandler) TopProducts(c *gin.Context) {
	start, end, ok := parseDateRange(c)
	if !ok {
		return
	}

	limit := parseLimit(c.DefaultQuery("limit", "10"))

	data, err := h.service.GetTopProducts(start, end, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// Helpers
func parseDateRange(c *gin.Context) (time.Time, time.Time, bool) {
	loc := time.Local
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	// default: last 30 days
	if startStr == "" || endStr == "" {
		end := time.Now().In(loc)
		start := end.AddDate(0, 0, -30)
		return start, end, true
	}

	start, err := time.ParseInLocation("2006-01-02", startStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid start_date format, use YYYY-MM-DD"})
		return time.Time{}, time.Time{}, false
	}

	end, err := time.ParseInLocation("2006-01-02", endStr, loc)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid end_date format, use YYYY-MM-DD"})
		return time.Time{}, time.Time{}, false
	}

	// make end inclusive end of day
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	return start, end, true
}

func parseLimit(limitStr string) int {
	limit := 10
	if parsed, err := strconv.Atoi(limitStr); err == nil {
		limit = parsed
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return limit
}
