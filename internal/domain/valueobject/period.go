package valueobject

import (
	"fmt"
	"time"
)

type PeriodType string

const (
	PeriodTypeDay      PeriodType = "day"
	PeriodTypeWeek     PeriodType = "week"
	PeriodTypeMonth    PeriodType = "month"
	PeriodTypeHalfYear PeriodType = "half_year"
)

// Period represents a specific time range for analytics
type Period struct {
	Type  PeriodType
	Start time.Time
	End   time.Time
}

// NewPeriod creates a new Period based on the type and anchor date
// The period includes the entire duration:
// - Day: from 00:00:00 to 23:59:59 of the anchor day
// - Week: from Monday 00:00:00 to Sunday 23:59:59 of the anchor week
// - Month: from 1st 00:00:00 to last day 23:59:59 of the anchor month
// - HalfYear: 6 months before anchor (including anchor month)
func NewPeriod(t PeriodType, anchor time.Time) Period {
	// Normalize anchor to start of day in local timezone
	anchor = time.Date(anchor.Year(), anchor.Month(), anchor.Day(), 0, 0, 0, 0, anchor.Location())

	switch t {
	case PeriodTypeDay:
		return Period{
			Type:  t,
			Start: anchor,
			End:   anchor.Add(24*time.Hour - time.Second),
		}

	case PeriodTypeWeek:
		// Find Monday (weekday 1)
		weekday := int(anchor.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := anchor.AddDate(0, 0, -weekday+1)
		return Period{
			Type:  t,
			Start: monday,
			End:   monday.AddDate(0, 0, 7).Add(-time.Second),
		}

	case PeriodTypeMonth:
		start := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location())
		// First day of next month minus 1 second
		end := start.AddDate(0, 1, 0).Add(-time.Second)
		return Period{
			Type:  t,
			Start: start,
			End:   end,
		}

	case PeriodTypeHalfYear:
		// 6 months back from anchor month
		start := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location()).AddDate(0, -5, 0)
		end := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, anchor.Location()).AddDate(0, 1, 0).Add(-time.Second)
		return Period{
			Type:  t,
			Start: start,
			End:   end,
		}

	default:
		return Period{
			Type:  PeriodTypeDay,
			Start: anchor,
			End:   anchor.Add(24*time.Hour - time.Second),
		}
	}
}

// Previous returns the previous period of the same type
func (p Period) Previous() Period {
	switch p.Type {
	case PeriodTypeDay:
		prevStart := p.Start.AddDate(0, 0, -1)
		return NewPeriod(p.Type, prevStart)

	case PeriodTypeWeek:
		prevStart := p.Start.AddDate(0, 0, -7)
		return NewPeriod(p.Type, prevStart)

	case PeriodTypeMonth:
		prevStart := p.Start.AddDate(0, -1, 0)
		return NewPeriod(p.Type, prevStart)

	case PeriodTypeHalfYear:
		prevAnchor := p.Start.AddDate(0, -6, 0)
		return NewPeriod(p.Type, prevAnchor)

	default:
		return p
	}
}

// Contains checks if a given time falls within this period
func (p Period) Contains(t time.Time) bool {
	return !t.Before(p.Start) && !t.After(p.End)
}

// Duration returns the length of the period
func (p Period) Duration() time.Duration {
	return p.End.Sub(p.Start)
}

// Format returns a human-readable representation of the period
func (p Period) Format() string {
	// Russian month names
	months := []string{
		"январь", "февраль", "март", "апрель", "май", "июнь",
		"июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь",
	}

	switch p.Type {
	case PeriodTypeDay:
		return fmt.Sprintf("%d %s %d",
			p.Start.Day(),
			months[p.Start.Month()-1],
			p.Start.Year())

	case PeriodTypeWeek:
		if p.Start.Month() == p.End.Month() {
			return fmt.Sprintf("%d-%d %s %d",
				p.Start.Day(),
				p.End.Day(),
				months[p.Start.Month()-1],
				p.Start.Year())
		}
		return fmt.Sprintf("%d %s - %d %s %d",
			p.Start.Day(),
			months[p.Start.Month()-1],
			p.End.Day(),
			months[p.End.Month()-1],
			p.End.Year())

	case PeriodTypeMonth:
		return fmt.Sprintf("%s %d",
			months[p.Start.Month()-1],
			p.Start.Year())

	case PeriodTypeHalfYear:
		return fmt.Sprintf("%s %d - %s %d",
			months[p.Start.Month()-1],
			p.Start.Year(),
			months[p.End.Month()-1],
			p.End.Year())

	default:
		return fmt.Sprintf("%s - %s", p.Start.Format("02.01.2006"), p.End.Format("02.01.2006"))
	}
}

// TypeDisplayName returns a human-readable name for the period type in Russian
func (p Period) TypeDisplayName() string {
	switch p.Type {
	case PeriodTypeDay:
		return "день"
	case PeriodTypeWeek:
		return "неделя"
	case PeriodTypeMonth:
		return "месяц"
	case PeriodTypeHalfYear:
		return "полгода"
	default:
		return string(p.Type)
	}
}
