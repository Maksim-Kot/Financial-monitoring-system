package service

import (
	"cmp"
	"math"
	"slices"

	"fms-project/internal/domain/entity"
	"fms-project/internal/domain/valueobject"
)

type AnalyticsCalculator interface {
	CalculateDelta(current, previous valueobject.MoneyAmount) float64
	CalculateCategoryDeltas(current, previous []entity.CategoryStat) []entity.CategoryDelta
	CalculatePercentages(stats []entity.CategoryStat, total valueobject.MoneyAmount) []entity.CategoryStat
	SortCategoryDeltasByAbsoluteDelta(deltas []entity.CategoryDelta) ([]entity.CategoryDelta, []entity.CategoryDelta)
}

type analyticsCalculatorService struct{}

func NewAnalyticsCalculatorService() AnalyticsCalculator {
	return &analyticsCalculatorService{}
}

// CalculateDelta calculates percentage change between current and previous amounts
// Formula: ((current - previous) / previous) * 100
// Returns positive for growth, negative for decline
// Edge cases:
//   - previous = 0, current > 0: returns +100.0
//   - previous = 0, current = 0: returns 0.0
//   - previous > 0, current = 0: returns -100.0
func (s *analyticsCalculatorService) CalculateDelta(current, previous valueobject.MoneyAmount) float64 {
	currentMinor := current.Int64()
	previousMinor := previous.Int64()

	// Edge case: no previous data
	if previousMinor == 0 {
		if currentMinor > 0 {
			return 100.0 // New expenses appeared
		}
		return 0.0 // No change (both zero)
	}

	// Edge case: no current data
	if currentMinor == 0 {
		return -100.0 // Expenses dropped to zero
	}

	// Standard calculation
	delta := float64(currentMinor-previousMinor) / float64(previousMinor) * 100.0

	// Cap extreme values for display
	if delta > 999.0 {
		return 999.0
	}
	if delta < -999.0 {
		return -999.0
	}

	return delta
}

// CalculateCategoryDeltas finds category intersections between periods
// and calculates delta for each. Returns only categories present in current period.
// Categories in previous but not in current are ignored.
// Categories in current but not in previous have DeltaPercent = +100.0
func (s *analyticsCalculatorService) CalculateCategoryDeltas(current, previous []entity.CategoryStat) []entity.CategoryDelta {
	// Build map of previous categories for fast lookup
	prevMap := make(map[valueobject.UUID]entity.CategoryStat)
	for _, p := range previous {
		prevMap[p.CategoryID] = p
	}

	deltas := make([]entity.CategoryDelta, 0, len(current))

	// Calculate delta for each current category
	for _, c := range current {
		delta := entity.CategoryDelta{
			CategoryID:   c.CategoryID,
			CategoryName: c.CategoryName,
			CurrentTotal: c.Total,
		}

		if prev, exists := prevMap[c.CategoryID]; exists {
			// Category existed in previous period
			delta.PreviousTotal = prev.Total
			delta.DeltaPercent = s.CalculateDelta(c.Total, prev.Total)
		} else {
			// New category (not in previous period)
			delta.PreviousTotal, _ = valueobject.NewMoneyAmountFromInt64(0, 2, valueobject.MoneyAmountCurrencyBYN, nil)
			delta.DeltaPercent = 100.0 // New appearance = 100% increase
		}

		deltas = append(deltas, delta)
	}

	return deltas
}

// CalculatePercentages adds Percentage field to each CategoryStat
// Formula: (category_total / total) * 100
// Percentages sum to ~100% (may have rounding differences)
func (s *analyticsCalculatorService) CalculatePercentages(stats []entity.CategoryStat, total valueobject.MoneyAmount) []entity.CategoryStat {
	totalMinor := total.Int64()
	if totalMinor == 0 {
		// Avoid division by zero - return stats with 0% each
		result := make([]entity.CategoryStat, len(stats))
		for i, stat := range stats {
			result[i] = stat
			result[i].Percentage = 0.0
		}
		return result
	}

	result := make([]entity.CategoryStat, len(stats))
	for i, stat := range stats {
		result[i] = stat
		statMinor := stat.Total.Int64()
		percentage := float64(statMinor) / float64(totalMinor) * 100.0
		result[i].Percentage = roundToOneDecimal(percentage)
	}

	return result
}

// SortCategoryDeltasByAbsoluteDelta sorts deltas by absolute change magnitude
// Returns: top increases (positive deltas) first, then top decreases
func (s *analyticsCalculatorService) SortCategoryDeltasByAbsoluteDelta(deltas []entity.CategoryDelta) ([]entity.CategoryDelta, []entity.CategoryDelta) {
	increases := make([]entity.CategoryDelta, 0)
	decreases := make([]entity.CategoryDelta, 0)

	for _, d := range deltas {
		if d.DeltaPercent > 0 {
			increases = append(increases, d)
		} else if d.DeltaPercent < 0 {
			decreases = append(decreases, d)
		}
		// Skip zero deltas (no change)
	}

	// Sort increases by delta descending (highest growth first)
	sortByDeltaDesc(increases)

	// Sort decreases by absolute delta descending (largest drop first)
	sortByAbsoluteDeltaDesc(decreases)

	return increases, decreases
}

func roundToOneDecimal(value float64) float64 {
	return float64(int(value*10+0.5)) / 10
}

func sortByDeltaDesc(items []entity.CategoryDelta) {
	slices.SortFunc(items, func(a, b entity.CategoryDelta) int {
		return cmp.Compare(b.DeltaPercent, a.DeltaPercent)
	})
}

func sortByAbsoluteDeltaDesc(items []entity.CategoryDelta) {
	slices.SortFunc(items, func(a, b entity.CategoryDelta) int {
		return cmp.Compare(
			math.Abs(b.DeltaPercent),
			math.Abs(a.DeltaPercent),
		)
	})
}
