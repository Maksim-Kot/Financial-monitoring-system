package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/repository"
	"fms-project/internal/domain/valueobject"
	"fms-project/internal/infrastructure/logger"
	"fms-project/internal/infrastructure/storage/postgres"
	storageModel "fms-project/internal/infrastructure/storage/postgres/model"

	"github.com/google/uuid"
)

const moneyDecimals uint8 = 2

type AnalyticsRepositoryConfig struct {
	Logger logger.Logger
	Client *postgres.Client
}

type AnalyticsRepository struct {
	logger logger.Logger
	client *postgres.Client
}

func NewAnalyticsRepository(cfg *AnalyticsRepositoryConfig) repository.AnalyticsRepository {
	return &AnalyticsRepository{
		logger: cfg.Logger.With("layer", "repository", "repository", "Analytics"),
		client: cfg.Client,
	}
}

// GetTotal returns total expenses and counts for the period
func (r *AnalyticsRepository) GetTotal(ctx context.Context, in repository.AnalyticsRepositoryGetSummaryIn) (repository.AnalyticsRepositoryGetSummaryOut, error) {
	db := postgres.CheckTx(ctx, r.client)

	// Query: total sum, purchase count, expense count
	type result struct {
		TotalMinor    int64 `bun:"total_minor"`
		PurchaseCount int   `bun:"purchase_count"`
		ExpenseCount  int   `bun:"expense_count"`
	}

	var res result
	err := db.NewSelect().
		Model((*storageModel.Expense)(nil)).
		ColumnExpr("COALESCE(SUM(e.total_price_minor), 0) as total_minor").
		ColumnExpr("COUNT(DISTINCT e.purchase_id) as purchase_count").
		ColumnExpr("COUNT(e.id) as expense_count").
		Join("JOIN purchases p ON p.id = e.purchase_id").
		Where("p.user_id = ?", in.UserID).
		Where("p.purchase_date >= ?", in.Period.Start).
		Where("p.purchase_date <= ?", in.Period.End).
		Scan(ctx, &res)

	if err != nil {
		return repository.AnalyticsRepositoryGetSummaryOut{}, fmt.Errorf("query total: %w", err)
	}

	total, err := valueobject.NewMoneyAmountFromInt64(
		res.TotalMinor,
		moneyDecimals,
		valueobject.MoneyAmountCurrencyBYN,
		nil,
	)
	if err != nil {
		return repository.AnalyticsRepositoryGetSummaryOut{}, fmt.Errorf("mapping total: %w", err)
	}

	return repository.AnalyticsRepositoryGetSummaryOut{
		Total:         total,
		PurchaseCount: res.PurchaseCount,
		ExpenseCount:  res.ExpenseCount,
	}, nil
}

// GetTopCategories returns top N categories by total amount
func (r *AnalyticsRepository) GetTopCategories(ctx context.Context, in repository.AnalyticsRepositoryGetCategoryTotalsIn) ([]entity.CategoryStat, error) {
	db := postgres.CheckTx(ctx, r.client)

	type result struct {
		CategoryID   uuid.UUID `bun:"category_id"`
		CategoryName string    `bun:"category_name"`
		CategoryIcon string    `bun:"category_icon"`
		TotalMinor   int64     `bun:"total_minor"`
		ExpenseCount int       `bun:"expense_count"`
	}

	var results []result
	err := db.NewSelect().
		Model((*storageModel.Expense)(nil)).
		ColumnExpr("e.category_id").
		ColumnExpr("c.name as category_name").
		ColumnExpr("c.icon as category_icon").
		ColumnExpr("SUM(e.total_price_minor) as total_minor").
		ColumnExpr("COUNT(e.id) as expense_count").
		Join("JOIN purchases p ON p.id = e.purchase_id").
		Join("JOIN categories c ON c.id = e.category_id").
		Where("p.user_id = ?", in.UserID).
		Where("p.purchase_date >= ?", in.Period.Start).
		Where("p.purchase_date <= ?", in.Period.End).
		GroupExpr("e.category_id, c.name, c.icon").
		OrderExpr("total_minor DESC").
		Limit(in.Limit).
		Scan(ctx, &results)

	if err != nil {
		return nil, fmt.Errorf("query top categories: %w", err)
	}

	stats := make([]entity.CategoryStat, 0, len(results))
	for _, r := range results {
		catID, err := valueobject.NewUUID(r.CategoryID.String())
		if err != nil {
			return nil, fmt.Errorf("mapping category id: %w", err)
		}

		total, err := valueobject.NewMoneyAmountFromInt64(
			r.TotalMinor,
			moneyDecimals,
			valueobject.MoneyAmountCurrencyBYN,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("mapping category total: %w", err)
		}

		stats = append(stats, entity.CategoryStat{
			CategoryID:    catID,
			CategoryName:  r.CategoryName,
			CategoryIcon:  r.CategoryIcon,
			Total:         total,
			PurchaseCount: r.ExpenseCount,
		})
	}

	return stats, nil
}

// GetTopPurchases returns top N most expensive purchases
func (r *AnalyticsRepository) GetTopPurchases(ctx context.Context, in repository.AnalyticsRepositoryGetPurchasesIn) ([]entity.PurchaseStat, error) {
	db := postgres.CheckTx(ctx, r.client)

	type result struct {
		PurchaseID       uuid.UUID `bun:"purchase_id"`
		PurchaseDate     time.Time `bun:"purchase_date"`
		OrganisationName string    `bun:"organisation_name"`
		TotalMinor       int64     `bun:"total_minor"`
		ItemCount        int       `bun:"item_count"`
	}

	var results []result
	err := db.NewSelect().
		Model((*storageModel.Expense)(nil)).
		ColumnExpr("e.purchase_id").
		ColumnExpr("p.purchase_date").
		ColumnExpr("p.organisation_name").
		ColumnExpr("SUM(e.total_price_minor) as total_minor").
		ColumnExpr("COUNT(e.id) as item_count").
		Join("JOIN purchases p ON p.id = e.purchase_id").
		Where("p.user_id = ?", in.UserID).
		Where("p.purchase_date >= ?", in.Period.Start).
		Where("p.purchase_date <= ?", in.Period.End).
		GroupExpr("e.purchase_id, p.purchase_date, p.organisation_name").
		OrderExpr("total_minor DESC").
		Limit(in.Limit).
		Scan(ctx, &results)

	if err != nil {
		return nil, fmt.Errorf("query top purchases: %w", err)
	}

	stats := make([]entity.PurchaseStat, 0, len(results))
	for _, r := range results {
		purchaseID, err := valueobject.NewUUID(r.PurchaseID.String())
		if err != nil {
			return nil, fmt.Errorf("mapping purchase id: %w", err)
		}

		total, err := valueobject.NewMoneyAmountFromInt64(
			r.TotalMinor,
			moneyDecimals,
			valueobject.MoneyAmountCurrencyBYN,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("mapping purchase total: %w", err)
		}

		stats = append(stats, entity.PurchaseStat{
			PurchaseID:       purchaseID,
			PurchaseDate:     r.PurchaseDate,
			OrganisationName: r.OrganisationName,
			Total:            total,
			ItemCount:        r.ItemCount,
		})
	}

	return stats, nil
}

// GetCategoryTotals returns all category totals for the period
func (r *AnalyticsRepository) GetCategoryTotals(ctx context.Context, in repository.AnalyticsRepositoryGetCategoryTotalsIn) (repository.AnalyticsRepositoryGetCategoryTotalsOut, error) {
	db := postgres.CheckTx(ctx, r.client)

	type result struct {
		CategoryID   uuid.UUID `bun:"category_id"`
		CategoryName string    `bun:"category_name"`
		CategoryIcon string    `bun:"category_icon"`
		TotalMinor   int64     `bun:"total_minor"`
		ExpenseCount int       `bun:"expense_count"`
	}

	var results []result
	err := db.NewSelect().
		Model((*storageModel.Expense)(nil)).
		ColumnExpr("e.category_id").
		ColumnExpr("c.name as category_name").
		ColumnExpr("c.icon as category_icon").
		ColumnExpr("SUM(e.total_price_minor) as total_minor").
		ColumnExpr("COUNT(e.id) as expense_count").
		Join("JOIN purchases p ON p.id = e.purchase_id").
		Join("JOIN categories c ON c.id = e.category_id").
		Where("p.user_id = ?", in.UserID).
		Where("p.purchase_date >= ?", in.Period.Start).
		Where("p.purchase_date <= ?", in.Period.End).
		GroupExpr("e.category_id, c.name, c.icon").
		OrderExpr("total_minor DESC").
		Scan(ctx, &results)

	if err != nil {
		return repository.AnalyticsRepositoryGetCategoryTotalsOut{}, fmt.Errorf("query category totals: %w", err)
	}

	stats := make([]entity.CategoryStat, 0, len(results))
	for _, r := range results {
		catID, err := valueobject.NewUUID(r.CategoryID.String())
		if err != nil {
			return repository.AnalyticsRepositoryGetCategoryTotalsOut{}, fmt.Errorf("mapping category id: %w", err)
		}

		total, err := valueobject.NewMoneyAmountFromInt64(
			r.TotalMinor,
			moneyDecimals,
			valueobject.MoneyAmountCurrencyBYN,
			nil,
		)
		if err != nil {
			return repository.AnalyticsRepositoryGetCategoryTotalsOut{}, fmt.Errorf("mapping category total: %w", err)
		}

		stats = append(stats, entity.CategoryStat{
			CategoryID:    catID,
			CategoryName:  r.CategoryName,
			CategoryIcon:  r.CategoryIcon,
			Total:         total,
			PurchaseCount: r.ExpenseCount,
		})
	}

	return repository.AnalyticsRepositoryGetCategoryTotalsOut{Totals: stats}, nil
}

// GetPurchases returns all purchases in the period with their totals
// Used for anomaly detection
func (r *AnalyticsRepository) GetPurchases(ctx context.Context, in repository.AnalyticsRepositoryGetPurchasesIn) (repository.AnalyticsRepositoryGetPurchasesOut, error) {
	db := postgres.CheckTx(ctx, r.client)

	type result struct {
		PurchaseID       uuid.UUID `bun:"purchase_id"`
		PurchaseDate     time.Time `bun:"purchase_date"`
		OrganisationName string    `bun:"organisation_name"`
		TotalMinor       int64     `bun:"total_minor"`
		ItemCount        int       `bun:"item_count"`
	}

	var results []result
	err := db.NewSelect().
		Model((*storageModel.Expense)(nil)).
		ColumnExpr("e.purchase_id").
		ColumnExpr("p.purchase_date").
		ColumnExpr("p.organisation_name").
		ColumnExpr("SUM(e.total_price_minor) as total_minor").
		ColumnExpr("COUNT(e.id) as item_count").
		Join("JOIN purchases p ON p.id = e.purchase_id").
		Where("p.user_id = ?", in.UserID).
		Where("p.purchase_date >= ?", in.Period.Start).
		Where("p.purchase_date <= ?", in.Period.End).
		GroupExpr("e.purchase_id, p.purchase_date, p.organisation_name").
		OrderExpr("p.purchase_date DESC").
		Scan(ctx, &results)

	if err != nil {
		return repository.AnalyticsRepositoryGetPurchasesOut{}, fmt.Errorf("query purchases: %w", err)
	}

	stats := make([]entity.PurchaseStat, 0, len(results))
	for _, r := range results {
		purchaseID, err := valueobject.NewUUID(r.PurchaseID.String())
		if err != nil {
			return repository.AnalyticsRepositoryGetPurchasesOut{}, fmt.Errorf("mapping purchase id: %w", err)
		}

		total, err := valueobject.NewMoneyAmountFromInt64(
			r.TotalMinor,
			moneyDecimals,
			valueobject.MoneyAmountCurrencyBYN,
			nil,
		)
		if err != nil {
			return repository.AnalyticsRepositoryGetPurchasesOut{}, fmt.Errorf("mapping purchase total: %w", err)
		}

		stats = append(stats, entity.PurchaseStat{
			PurchaseID:       purchaseID,
			PurchaseDate:     r.PurchaseDate,
			OrganisationName: r.OrganisationName,
			Total:            total,
			ItemCount:        r.ItemCount,
		})
	}

	return repository.AnalyticsRepositoryGetPurchasesOut{Purchases: stats}, nil
}

// GetAverageExpense returns average expense amount for the period
// If CategoryID is nil, calculates overall average across all purchases
// If CategoryID is set, calculates average for purchases containing that category
func (r *AnalyticsRepository) GetAverageExpense(ctx context.Context, in repository.AnalyticsRepositoryGetAverageExpenseIn) (repository.AnalyticsRepositoryGetAverageExpenseOut, error) {
	db := postgres.CheckTx(ctx, r.client)

	type result struct {
		AvgMinor sql.NullFloat64 `bun:"avg_minor"`
	}

	var res result
	var err error

	if in.CategoryID == nil {
		// Overall average - all purchases using raw query
		query := `
			SELECT AVG(total_minor) as avg_minor
			FROM (
				SELECT SUM(e.total_price_minor) as total_minor
				FROM expenses e
				JOIN purchases p ON p.id = e.purchase_id
				WHERE p.user_id = ?
				  AND p.purchase_date >= ?
				  AND p.purchase_date <= ?
				GROUP BY e.purchase_id
			) as purchase_totals`
		err = db.QueryRowContext(ctx, query, in.UserID, in.Period.Start, in.Period.End).Scan(&res.AvgMinor)
	} else {
		// Category-specific average using raw query
		query := `
			SELECT AVG(total_minor) as avg_minor
			FROM (
				SELECT SUM(e.total_price_minor) as total_minor
				FROM expenses e
				JOIN purchases p ON p.id = e.purchase_id
				WHERE p.user_id = ?
				  AND p.purchase_date >= ?
				  AND p.purchase_date <= ?
				  AND EXISTS (
					  SELECT 1 FROM expenses e2
					  WHERE e2.purchase_id = e.purchase_id
					    AND e2.category_id = ?
				  )
				GROUP BY e.purchase_id
			) as purchase_totals`
		err = db.QueryRowContext(ctx, query, in.UserID, in.Period.Start, in.Period.End, in.CategoryID.Value()).Scan(&res.AvgMinor)
	}

	if err != nil {
		return repository.AnalyticsRepositoryGetAverageExpenseOut{}, fmt.Errorf("query average: %w", err)
	}

	// Handle NaN/NULL when no data
	var avgMinor int64
	if res.AvgMinor.Valid {
		avgMinor = int64(res.AvgMinor.Float64)
	} else {
		avgMinor = 0
	}

	avg, err := valueobject.NewMoneyAmountFromInt64(
		avgMinor,
		moneyDecimals,
		valueobject.MoneyAmountCurrencyBYN,
		nil,
	)
	if err != nil {
		return repository.AnalyticsRepositoryGetAverageExpenseOut{}, fmt.Errorf("mapping average: %w", err)
	}

	return repository.AnalyticsRepositoryGetAverageExpenseOut{AvgTotal: avg}, nil
}
