package service

import (
	"cmp"
	"slices"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/valueobject"
)

type DetectAnomaliesInput struct {
	Purchases      []entity.PurchaseStat
	AverageExpense valueobject.MoneyAmount
	CategoryAvgs   map[valueobject.UUID]valueobject.MoneyAmount
}

// AnomalyDetector defines operations for detecting unusual purchases
// Uses statistical approach without ML: purchase_total > avg * threshold
type AnomalyDetector interface {
	Detect(input DetectAnomaliesInput) []entity.Anomaly
	SetThreshold(threshold float64)
}

type anomalyDetectorService struct {
	threshold float64
}

func NewAnomalyDetectorService(threshold float64) AnomalyDetector {
	if threshold <= 0 {
		threshold = 2.5
	}
	return &anomalyDetectorService{
		threshold: threshold,
	}
}

// SetThreshold allows adjusting the detection sensitivity
// Default is 2.5 (purchase must be 2.5x above average)
// Higher values = fewer anomalies (stricter)
// Lower values = more anomalies (looser)
func (s *anomalyDetectorService) SetThreshold(threshold float64) {
	if threshold > 0 {
		s.threshold = threshold
	}
}

// Detect finds unusually expensive purchases
// Algorithm:
//  1. For each purchase, compare its total to average
//  2. Prefer category-specific average when available
//  3. Mark as anomaly if total > average * threshold
//  4. Calculate factor = total / average
//
// Returns sorted by factor (highest first)
func (s *anomalyDetectorService) Detect(input DetectAnomaliesInput) []entity.Anomaly {
	if len(input.Purchases) == 0 {
		return nil
	}

	avgMinor := input.AverageExpense.Int64()
	if avgMinor == 0 {
		// Cannot detect without baseline
		return nil
	}

	anomalies := make([]entity.Anomaly, 0)

	for _, purchase := range input.Purchases {
		purchaseMinor := purchase.Total.Int64()
		if purchaseMinor == 0 {
			continue
		}

		// Determine which average to use
		// For simplicity, we use global average for all purchases
		// In future: could determine dominant category of purchase
		comparisonAvg := avgMinor
		avgToStore := input.AverageExpense

		// If we have category averages, find the best match
		// (for now, use global average - category matching requires more data)
		_ = input.CategoryAvgs // Reserved for future category-specific detection

		// Check if purchase exceeds threshold
		factor := float64(purchaseMinor) / float64(comparisonAvg)

		if factor > s.threshold {
			anomaly := entity.Anomaly{
				PurchaseID:   purchase.PurchaseID,
				PurchaseDate: purchase.PurchaseDate,
				Name:         purchase.OrganisationName,
				Total:        purchase.Total,
				AvgCategory:  avgToStore,
				Factor:       roundToTwoDecimals(factor),
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	sortAnomaliesByFactorDesc(anomalies)

	return anomalies
}

func roundToTwoDecimals(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func sortAnomaliesByFactorDesc(items []entity.Anomaly) {
	slices.SortFunc(items, func(a, b entity.Anomaly) int {
		return cmp.Compare(b.Factor, a.Factor)
	})
}
