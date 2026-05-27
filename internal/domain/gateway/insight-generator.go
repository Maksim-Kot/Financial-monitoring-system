package gateway

import (
	"context"

	"fms-project/internal/domain/entity"
)

// Contains pre-calculated analytical data
type InsightGeneratorGatewayIn struct {
	// Period information
	PeriodType  string
	Total       int64
	HasPrevious bool // Whether previous period data exists

	// Comparison data (empty if HasPrevious is false)
	DeltaPercent float64

	// Top categories
	TopCategories []InsightCategoryInfo

	// Category changes for insights (significant deltas only)
	CategoryChanges []InsightCategoryChange

	// Anomalies count
	AnomaliesCount int
}

// Category data for AI input
type InsightCategoryInfo struct {
	Name       string
	Percentage float64
}

// Category delta for AI input
type InsightCategoryChange struct {
	Name         string
	DeltaPercent float64
}

type InsightGeneratorGatewayOut struct {
	Insights []entity.Insight
}

type InsightGeneratorGateway interface {
	// GenerateInsights creates AI-powered insights based on analytical data
	GenerateInsights(ctx context.Context, in InsightGeneratorGatewayIn) (InsightGeneratorGatewayOut, error)
}
