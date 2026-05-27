package repository

import (
	"context"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/valueobject"
)

type AnalyticsRepositoryGetSummaryIn struct {
	UserID             int64
	Period             valueobject.Period
	TopCategoriesLimit int
	TopPurchasesLimit  int
}

type AnalyticsRepositoryGetSummaryOut struct {
	Total         valueobject.MoneyAmount
	PurchaseCount int
	ExpenseCount  int
	TopCategories []entity.CategoryStat
	TopPurchases  []entity.PurchaseStat
}

type AnalyticsRepositoryGetCategoryTotalsIn struct {
	UserID int64
	Period valueobject.Period
	Limit  int
}

type AnalyticsRepositoryGetCategoryTotalsOut struct {
	Totals []entity.CategoryStat
}

type AnalyticsRepositoryGetPurchasesIn struct {
	UserID int64
	Period valueobject.Period
	Limit  int
}

type AnalyticsRepositoryGetPurchasesOut struct {
	Purchases []entity.PurchaseStat
}

type AnalyticsRepositoryGetAverageExpenseIn struct {
	UserID     int64
	Period     valueobject.Period
	CategoryID *valueobject.UUID // Optional - if nil, calculates overall average
}

type AnalyticsRepositoryGetAverageExpenseOut struct {
	AvgTotal valueobject.MoneyAmount
}

// AnalyticsRepository is a read-only repository for analytics data
type AnalyticsRepository interface {
	// GetTotal returns total expenses for the period
	GetTotal(ctx context.Context, in AnalyticsRepositoryGetSummaryIn) (AnalyticsRepositoryGetSummaryOut, error)

	// GetTopCategories returns top N categories by total amount
	GetTopCategories(ctx context.Context, in AnalyticsRepositoryGetCategoryTotalsIn) ([]entity.CategoryStat, error)

	// GetTopPurchases returns top N most expensive purchases
	GetTopPurchases(ctx context.Context, in AnalyticsRepositoryGetPurchasesIn) ([]entity.PurchaseStat, error)

	// GetCategoryTotals returns all category totals for the period
	GetCategoryTotals(ctx context.Context, in AnalyticsRepositoryGetCategoryTotalsIn) (AnalyticsRepositoryGetCategoryTotalsOut, error)

	// GetPurchases returns all purchases in the period with their totals
	// Used for anomaly detection
	GetPurchases(ctx context.Context, in AnalyticsRepositoryGetPurchasesIn) (AnalyticsRepositoryGetPurchasesOut, error)

	// GetAverageExpense returns average expense amount
	// If CategoryID is nil, returns overall average across all categories
	// If CategoryID is set, returns average for that specific category
	GetAverageExpense(ctx context.Context, in AnalyticsRepositoryGetAverageExpenseIn) (AnalyticsRepositoryGetAverageExpenseOut, error)
}
