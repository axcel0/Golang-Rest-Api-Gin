package services

import (
	"Go-Lang-project-01/internal/models"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AnalyticsService struct {
	db *gorm.DB
}

func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

type SummaryMetrics struct {
	Revenue      decimal.Decimal `json:"revenue"`
	Profit       decimal.Decimal `json:"profit"`
	Transactions int64           `json:"transactions"`
	ItemsSold    int64           `json:"items_sold"`
}

type PaymentBreakdown struct {
	PaymentMethod models.PaymentMethod `json:"payment_method"`
	Total         decimal.Decimal      `json:"total"`
	Count         int64                `json:"count"`
}

type TopProduct struct {
	ProductID   uint            `json:"product_id"`
	ProductName string          `json:"product_name"`
	SKU         string          `json:"sku"`
	Quantity    int64           `json:"quantity"`
	Revenue     decimal.Decimal `json:"revenue"`
}

// GetSummary returns revenue, profit, transaction count, items sold in range (inclusive)
func (s *AnalyticsService) GetSummary(startDate, endDate time.Time) (*SummaryMetrics, error) {
	var revenueStr, profitStr string
	var transactionsCount, itemsSold int64

	// Revenue & transactions
	err := s.db.Model(&models.Transaction{}).
		Select("COALESCE(SUM(total),0) as revenue, COUNT(*) as transactions").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Row().Scan(&revenueStr, &transactionsCount)
	if err != nil {
		return nil, err
	}

	// Items sold and profit
	err = s.db.Model(&models.TransactionItem{}).
		Select("COALESCE(SUM(quantity),0) as items_sold, COALESCE(SUM((price - cost_price) * quantity),0) as profit").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Row().Scan(&itemsSold, &profitStr)
	if err != nil {
		return nil, err
	}

	revenue := decimal.Zero
	if r, err := decimal.NewFromString(revenueStr); err == nil {
		revenue = r
	}
	profit := decimal.Zero
	if p, err := decimal.NewFromString(profitStr); err == nil {
		profit = p
	}

	return &SummaryMetrics{
		Revenue:      revenue,
		Profit:       profit,
		Transactions: transactionsCount,
		ItemsSold:    itemsSold,
	}, nil
}

// GetPaymentBreakdown groups totals by payment method
func (s *AnalyticsService) GetPaymentBreakdown(startDate, endDate time.Time) ([]PaymentBreakdown, error) {
	rows, err := s.db.Model(&models.Transaction{}).
		Select("payment_method, COALESCE(SUM(total),0) as total, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Group("payment_method").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PaymentBreakdown
	for rows.Next() {
		var method string
		var totalStr string
		var count int64
		if err := rows.Scan(&method, &totalStr, &count); err != nil {
			return nil, err
		}
		total := decimal.Zero
		if t, err := decimal.NewFromString(totalStr); err == nil {
			total = t
		}
		result = append(result, PaymentBreakdown{
			PaymentMethod: models.PaymentMethod(method),
			Total:         total,
			Count:         count,
		})
	}

	return result, nil
}

// GetTopProducts returns top products by quantity sold
func (s *AnalyticsService) GetTopProducts(startDate, endDate time.Time, limit int) ([]TopProduct, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}

	rows, err := s.db.Table("transaction_items ti").
		Select("ti.product_id, ti.product_name, ti.product_sku, COALESCE(SUM(ti.quantity),0) as qty, COALESCE(SUM(ti.price * ti.quantity),0) as revenue").
		Where("ti.created_at >= ? AND ti.created_at <= ?", startDate, endDate).
		Group("ti.product_id, ti.product_name, ti.product_sku").
		Order("qty DESC").
		Limit(limit).
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TopProduct
	for rows.Next() {
		var tp TopProduct
		var revenueStr string
		if err := rows.Scan(&tp.ProductID, &tp.ProductName, &tp.SKU, &tp.Quantity, &revenueStr); err != nil {
			return nil, err
		}
		revenue := decimal.Zero
		if r, err := decimal.NewFromString(revenueStr); err == nil {
			revenue = r
		}
		tP := tp
		tP.Revenue = revenue
		result = append(result, tP)
	}

	return result, nil
}
