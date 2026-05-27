package usecase

import (
	"cmp"
	"context"
	"math"
	"slices"
	"time"

	"fms-project/internal/application/service"
	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/gateway"
	"fms-project/internal/domain/repository"
	"fms-project/internal/domain/valueobject"
	"fms-project/internal/infrastructure/logger"
)

type GetAnalyticsUseCaseRequest struct {
	UserID     int64
	PeriodType valueobject.PeriodType
	Detailed   bool // If true, include extended analysis
}

type GetAnalyticsUseCaseResponse struct {
	Summary  entity.Summary
	Detailed *entity.DetailedReport // nil if Detailed=false
}

type GetAnalyticsUseCaseConfig struct {
	Logger           logger.Logger
	AnalyticsRepo    repository.AnalyticsRepository
	Calculator       service.AnalyticsCalculator
	AnomalyDetector  service.AnomalyDetector
	InsightGenerator gateway.InsightGeneratorGateway
}

// GetAnalyticsUseCase implements the two-level analytics flow
// Level 1: Quick Summary (always returned)
// Level 2: Detailed Analysis (if Detailed=true)
type GetAnalyticsUseCase struct {
	logger           logger.Logger
	analyticsRepo    repository.AnalyticsRepository
	calculator       service.AnalyticsCalculator
	anomalyDetector  service.AnomalyDetector
	insightGenerator gateway.InsightGeneratorGateway
}

func NewGetAnalyticsUseCase(cfg *GetAnalyticsUseCaseConfig) *GetAnalyticsUseCase {
	return &GetAnalyticsUseCase{
		logger:           cfg.Logger.With("layer", "usecase", "usecase", "GetAnalytics"),
		analyticsRepo:    cfg.AnalyticsRepo,
		calculator:       cfg.Calculator,
		anomalyDetector:  cfg.AnomalyDetector,
		insightGenerator: cfg.InsightGenerator,
	}
}

func (uc *GetAnalyticsUseCase) Execute(ctx context.Context, in GetAnalyticsUseCaseRequest) (GetAnalyticsUseCaseResponse, error) {
	logger := uc.logger.With("userID", in.UserID, "periodType", in.PeriodType, "detailed", in.Detailed)

	// Build Period
	period := valueobject.NewPeriod(in.PeriodType, time.Now())
	logger.DebugContext(ctx, "built period", "start", period.Start, "end", period.End)

	summary, err := uc.getSummary(ctx, in.UserID, period)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get summary", "error", err)
		return GetAnalyticsUseCaseResponse{}, err
	}

	// If not detailed, return just the summary
	if !in.Detailed {
		return GetAnalyticsUseCaseResponse{
			Summary: summary,
		}, nil
	}

	// Build Detailed Report
	detailed, err := uc.buildDetailedReport(ctx, in.UserID, summary, period)
	if err != nil {
		logger.ErrorContext(ctx, "failed to build detailed report", "error", err)
		// Return summary even if detailed fails
		return GetAnalyticsUseCaseResponse{
			Summary: summary,
		}, nil
	}

	return GetAnalyticsUseCaseResponse{
		Summary:  summary,
		Detailed: detailed,
	}, nil
}

func (uc *GetAnalyticsUseCase) getSummary(ctx context.Context, userID int64, period valueobject.Period) (entity.Summary, error) {
	totalOut, err := uc.analyticsRepo.GetTotal(ctx, repository.AnalyticsRepositoryGetSummaryIn{
		UserID: userID,
		Period: period,
	})
	if err != nil {
		return entity.Summary{}, err
	}

	topCategories, err := uc.analyticsRepo.GetTopCategories(ctx, repository.AnalyticsRepositoryGetCategoryTotalsIn{
		UserID: userID,
		Period: period,
		Limit:  5,
	})
	if err != nil {
		return entity.Summary{}, err
	}

	// Calculate percentages for categories
	topCategories = uc.calculator.CalculatePercentages(topCategories, totalOut.Total)

	topPurchases, err := uc.analyticsRepo.GetTopPurchases(ctx, repository.AnalyticsRepositoryGetPurchasesIn{
		UserID: userID,
		Period: period,
		Limit:  3,
	})
	if err != nil {
		return entity.Summary{}, err
	}

	return entity.Summary{
		Period:        period,
		Total:         totalOut.Total,
		PurchaseCount: totalOut.PurchaseCount,
		ExpenseCount:  totalOut.ExpenseCount,
		TopCategories: topCategories,
		TopPurchases:  topPurchases,
	}, nil
}

func (uc *GetAnalyticsUseCase) buildDetailedReport(ctx context.Context, userID int64, summary entity.Summary, period valueobject.Period,
) (*entity.DetailedReport, error) {
	logger := uc.logger.With("userID", userID)

	prevPeriod := period.Previous()

	prevTotalOut, err := uc.analyticsRepo.GetTotal(ctx, repository.AnalyticsRepositoryGetSummaryIn{
		UserID: userID,
		Period: prevPeriod,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get previous period total", "error", err)
		// Continue with zero previous data
		prevTotalOut.Total, _ = valueobject.NewMoneyAmountFromInt64(0, 2, valueobject.MoneyAmountCurrencyBYN, nil)
	}

	// Build comparison
	comparison := entity.Comparison{
		CurrentPeriod:  period,
		PreviousPeriod: prevPeriod,
		CurrentTotal:   summary.Total,
		PreviousTotal:  prevTotalOut.Total,
		DeltaPercent:   uc.calculator.CalculateDelta(summary.Total, prevTotalOut.Total),
	}

	// Get category data for both periods
	currentCatsOut, err := uc.analyticsRepo.GetCategoryTotals(ctx, repository.AnalyticsRepositoryGetCategoryTotalsIn{
		UserID: userID,
		Period: period,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get current categories", "error", err)
		currentCatsOut.Totals = nil
	}

	prevCatsOut, err := uc.analyticsRepo.GetCategoryTotals(ctx, repository.AnalyticsRepositoryGetCategoryTotalsIn{
		UserID: userID,
		Period: prevPeriod,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get previous categories", "error", err)
		prevCatsOut.Totals = nil
	}

	categoryDeltas := uc.calculator.CalculateCategoryDeltas(currentCatsOut.Totals, prevCatsOut.Totals)

	// Anomaly detection
	purchasesOut, err := uc.analyticsRepo.GetPurchases(ctx, repository.AnalyticsRepositoryGetPurchasesIn{
		UserID: userID,
		Period: period,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get purchases for anomaly detection", "error", err)
		purchasesOut.Purchases = nil
	}

	avgOut, err := uc.analyticsRepo.GetAverageExpense(ctx, repository.AnalyticsRepositoryGetAverageExpenseIn{
		UserID: userID,
		Period: period,
	})
	if err != nil {
		logger.ErrorContext(ctx, "failed to get average expense", "error", err)
		avgOut.AvgTotal, _ = valueobject.NewMoneyAmountFromInt64(0, 2, valueobject.MoneyAmountCurrencyBYN, nil)
	}

	// Detect anomalies
	anomalies := uc.anomalyDetector.Detect(service.DetectAnomaliesInput{
		Purchases:      purchasesOut.Purchases,
		AverageExpense: avgOut.AvgTotal,
		CategoryAvgs:   nil,
	})

	// Generate AI insights only if we have previous data to compare
	var insights []entity.Insight
	if prevTotalOut.Total.Int64() > 0 {
		aiInput := buildAIInput(summary, comparison, categoryDeltas, anomalies)
		aiOut, err := uc.insightGenerator.GenerateInsights(ctx, aiInput)
		if err != nil {
			logger.ErrorContext(ctx, "failed to generate insights", "error", err)
			// Continue without insights
		} else {
			insights = aiOut.Insights
		}
	}

	return &entity.DetailedReport{
		Summary:        summary,
		Comparison:     comparison,
		CategoryDeltas: categoryDeltas,
		Anomalies:      anomalies,
		Insights:       insights,
	}, nil
}

func buildAIInput(summary entity.Summary, comparison entity.Comparison, categoryDeltas []entity.CategoryDelta, anomalies []entity.Anomaly,
) gateway.InsightGeneratorGatewayIn {
	// Build top categories info (max 3)
	topCats := make([]gateway.InsightCategoryInfo, 0, 3)
	for i, cat := range summary.TopCategories {
		if i >= 3 {
			break
		}
		topCats = append(topCats, gateway.InsightCategoryInfo{
			Name:       cat.CategoryName,
			Percentage: cat.Percentage,
		})
	}

	// Build category changes (significant deltas only, max 3 each direction)
	increases, decreases := sortDeltasByMagnitude(categoryDeltas)

	changes := make([]gateway.InsightCategoryChange, 0)
	for i, d := range increases {
		if i >= 3 {
			break
		}
		changes = append(changes, gateway.InsightCategoryChange{
			Name:         d.CategoryName,
			DeltaPercent: d.DeltaPercent,
		})
	}
	for i, d := range decreases {
		if i >= 3 {
			break
		}
		changes = append(changes, gateway.InsightCategoryChange{
			Name:         d.CategoryName,
			DeltaPercent: d.DeltaPercent,
		})
	}

	return gateway.InsightGeneratorGatewayIn{
		PeriodType:      string(summary.Period.Type),
		Total:           summary.Total.Int64(),
		HasPrevious:     true,
		DeltaPercent:    comparison.DeltaPercent,
		TopCategories:   topCats,
		CategoryChanges: changes,
		AnomaliesCount:  len(anomalies),
	}
}

func sortDeltasByMagnitude(deltas []entity.CategoryDelta) ([]entity.CategoryDelta, []entity.CategoryDelta) {
	increases := make([]entity.CategoryDelta, 0)
	decreases := make([]entity.CategoryDelta, 0)

	for _, d := range deltas {
		if d.DeltaPercent > 0 {
			increases = append(increases, d)
		} else if d.DeltaPercent < 0 {
			decreases = append(decreases, d)
		}
	}

	// Sort increases by delta descending
	slices.SortFunc(increases, func(a, b entity.CategoryDelta) int {
		return cmp.Compare(b.DeltaPercent, a.DeltaPercent)
	})

	// Sort decreases by absolute delta descending
	slices.SortFunc(decreases, func(a, b entity.CategoryDelta) int {
		return cmp.Compare(
			math.Abs(b.DeltaPercent),
			math.Abs(a.DeltaPercent),
		)
	})

	return increases, decreases
}
