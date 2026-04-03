package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDateExpr parses a TaskWarrior-compatible date expression relative to now.
// Supports named dates, period boundaries, weekdays, months, ISO dates,
// relative durations, and arithmetic (base +/- duration).
func ParseDateExpr(input string, now time.Time) (time.Time, error) {
	if input == "" {
		return time.Time{}, fmt.Errorf("empty date expression")
	}

	s := strings.ToLower(input)

	// Split into base and optional arithmetic (+/- duration).
	base, op, durStr := splitArithmetic(s)
	origBase, _, _ := splitArithmetic(input)

	t, err := parseBase(base, origBase, now)
	if err != nil {
		return time.Time{}, err
	}

	if op == 0 {
		return t, nil
	}

	d, calendarMonths, calendarYears, err := parseDuration(durStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration %q: %w", durStr, err)
	}

	if op == '-' {
		d = -d
		calendarMonths = -calendarMonths
		calendarYears = -calendarYears
	}

	t = t.AddDate(calendarYears, calendarMonths, 0)
	t = t.Add(d)
	return t, nil
}

// splitArithmetic splits "base+dur" or "base-dur" into components.
// Returns (base, op, dur) where op is '+', '-', or 0.
func splitArithmetic(s string) (string, byte, string) {
	// Walk past a potential ISO date prefix (YYYY-MM-DD) before looking
	// for arithmetic operators, so the hyphens inside dates are not
	// treated as subtraction.
	searchStart := 0
	if len(s) >= 10 && s[4] == '-' && s[7] == '-' {
		searchStart = 10
	}

	for i := searchStart; i < len(s); i++ {
		if s[i] == '+' || s[i] == '-' {
			return s[:i], s[i], s[i+1:]
		}
	}
	return s, 0, ""
}

// parseBase parses the base component of a date expression.
// orig is the original (non-lowercased) input for ISO datetime parsing.
func parseBase(s string, orig string, now time.Time) (time.Time, error) {
	// Named dates
	switch s {
	case "now":
		return now, nil
	case "today", "sod":
		return startOfDay(now), nil
	case "eod":
		return endOfDay(now), nil
	case "tomorrow":
		return startOfDay(now).AddDate(0, 0, 1), nil
	case "yesterday":
		return startOfDay(now).AddDate(0, 0, -1), nil
	case "later", "someday":
		return time.Date(9999, 12, 30, 0, 0, 0, 0, now.Location()), nil
	}

	// Period boundaries
	if t, ok := parsePeriodBoundary(s, now); ok {
		return t, nil
	}

	// Weekdays
	if wd, ok := parseWeekday(s); ok {
		return nextWeekday(now, wd), nil
	}

	// Months
	if m, ok := parseMonth(s); ok {
		return nextMonth(now, m), nil
	}

	// Relative duration (e.g. "2d", "3w")
	if d, cm, cy, err := parseDuration(s); err == nil {
		return now.AddDate(cy, cm, 0).Add(d), nil
	}

	// ISO date (use orig to preserve 'T' separator)
	if t, err := time.Parse("2006-01-02T15:04:05", orig); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unrecognized date expression %q", s)
}

// parseDuration parses a TW duration like "2d", "3w", "1mo", "1q", "1y".
// Returns (time.Duration for h/min/s/d/w, calendar months, calendar years, error).
func parseDuration(s string) (time.Duration, int, int, error) {
	if s == "" {
		return 0, 0, 0, fmt.Errorf("empty duration")
	}

	// Find where digits end.
	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9') {
		i++
	}
	if i == 0 || i == len(s) {
		return 0, 0, 0, fmt.Errorf("invalid duration %q", s)
	}

	n, err := strconv.Atoi(s[:i])
	if err != nil {
		return 0, 0, 0, err
	}

	unit := s[i:]
	switch unit {
	case "s":
		return time.Duration(n) * time.Second, 0, 0, nil
	case "min":
		return time.Duration(n) * time.Minute, 0, 0, nil
	case "h":
		return time.Duration(n) * time.Hour, 0, 0, nil
	case "d":
		return time.Duration(n) * 24 * time.Hour, 0, 0, nil
	case "w":
		return time.Duration(n) * 7 * 24 * time.Hour, 0, 0, nil
	case "mo":
		return 0, n, 0, nil
	case "q":
		return 0, n * 3, 0, nil
	case "y":
		return 0, 0, n, nil
	default:
		return 0, 0, 0, fmt.Errorf("unknown duration unit %q", unit)
	}
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

func startOfWeek(t time.Time) time.Time {
	wd := t.Weekday()
	offset := (int(wd) - int(time.Monday) + 7) % 7
	return startOfDay(t).AddDate(0, 0, -offset)
}

func endOfWeek(t time.Time) time.Time {
	return endOfDay(startOfWeek(t).AddDate(0, 0, 6))
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

func endOfMonth(t time.Time) time.Time {
	return endOfDay(startOfMonth(t).AddDate(0, 1, -1))
}

func quarterStart(t time.Time) time.Month {
	return time.Month(((int(t.Month()) - 1) / 3 * 3) + 1)
}

func startOfQuarter(t time.Time) time.Time {
	return time.Date(t.Year(), quarterStart(t), 1, 0, 0, 0, 0, t.Location())
}

func endOfQuarter(t time.Time) time.Time {
	return endOfDay(startOfQuarter(t).AddDate(0, 3, -1))
}

func startOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

func endOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 0, t.Location())
}

// parsePeriodBoundary handles sow/eow/soww/eoww/som/eom/soq/eoq/soy/eoy
// and their next/prev variants (sonw, eopw, etc.).
func parsePeriodBoundary(s string, now time.Time) (time.Time, bool) {
	type periodFn struct {
		start func(time.Time) time.Time
		end   func(time.Time) time.Time
	}

	periods := map[string]periodFn{
		"w":  {startOfWeek, endOfWeek},
		"ww": {startOfWeek, func(t time.Time) time.Time { return endOfDay(startOfWeek(t).AddDate(0, 0, 4)) }},
		"m":  {startOfMonth, endOfMonth},
		"q":  {startOfQuarter, endOfQuarter},
		"y":  {startOfYear, endOfYear},
	}

	offsets := map[string]func(time.Time, string) time.Time{
		"": func(t time.Time, _ string) time.Time { return t },
		"n": func(t time.Time, period string) time.Time {
			switch period {
			case "w", "ww":
				return t.AddDate(0, 0, 7)
			case "m":
				return t.AddDate(0, 1, 0)
			case "q":
				return t.AddDate(0, 3, 0)
			case "y":
				return t.AddDate(1, 0, 0)
			}
			return t
		},
		"p": func(t time.Time, period string) time.Time {
			switch period {
			case "w", "ww":
				return t.AddDate(0, 0, -7)
			case "m":
				return t.AddDate(0, -1, 0)
			case "q":
				return t.AddDate(0, -3, 0)
			case "y":
				return t.AddDate(-1, 0, 0)
			}
			return t
		},
	}

	// Try patterns: so[n|p]<period>, eo[n|p]<period>
	for _, prefix := range []string{"so", "eo"} {
		if !strings.HasPrefix(s, prefix) {
			continue
		}
		rest := s[2:]

		for _, modifier := range []string{"n", "p", ""} {
			if modifier != "" && !strings.HasPrefix(rest, modifier) {
				continue
			}
			periodKey := rest[len(modifier):]

			pf, ok := periods[periodKey]
			if !ok {
				continue
			}

			shifted := offsets[modifier](now, periodKey)

			if prefix == "so" {
				return pf.start(shifted), true
			}
			return pf.end(shifted), true
		}
	}

	return time.Time{}, false
}

var weekdayMap = map[string]time.Weekday{
	"monday": time.Monday, "mon": time.Monday,
	"tuesday": time.Tuesday, "tue": time.Tuesday,
	"wednesday": time.Wednesday, "wed": time.Wednesday,
	"thursday": time.Thursday, "thu": time.Thursday,
	"friday": time.Friday, "fri": time.Friday,
	"saturday": time.Saturday, "sat": time.Saturday,
	"sunday": time.Sunday, "sun": time.Sunday,
}

func parseWeekday(s string) (time.Weekday, bool) {
	wd, ok := weekdayMap[s]
	return wd, ok
}

func nextWeekday(now time.Time, wd time.Weekday) time.Time {
	today := startOfDay(now)
	current := now.Weekday()
	days := (int(wd) - int(current) + 7) % 7
	if days == 0 {
		days = 7
	}
	return today.AddDate(0, 0, days)
}

var monthMap = map[string]time.Month{
	"january": time.January, "jan": time.January,
	"february": time.February, "feb": time.February,
	"march": time.March, "mar": time.March,
	"april": time.April, "apr": time.April,
	"may":  time.May,
	"june": time.June, "jun": time.June,
	"july": time.July, "jul": time.July,
	"august": time.August, "aug": time.August,
	"september": time.September, "sep": time.September,
	"october": time.October, "oct": time.October,
	"november": time.November, "nov": time.November,
	"december": time.December, "dec": time.December,
}

func parseMonth(s string) (time.Month, bool) {
	m, ok := monthMap[s]
	return m, ok
}

func nextMonth(now time.Time, m time.Month) time.Time {
	y := now.Year()
	if m <= now.Month() {
		y++
	}
	return time.Date(y, m, 1, 0, 0, 0, 0, now.Location())
}
