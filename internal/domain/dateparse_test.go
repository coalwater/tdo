package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ref is a fixed reference time: Wednesday 2026-04-15 10:30:00 UTC.
var ref = time.Date(2026, 4, 15, 10, 30, 0, 0, time.UTC)

func TestParseDateExpr_Named(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		{"now", ref},
		{"today", time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)},
		{"sod", time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)},
		{"eod", time.Date(2026, 4, 15, 23, 59, 59, 0, time.UTC)},
		{"tomorrow", time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)},
		{"yesterday", time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC)},
		{"later", time.Date(9999, 12, 30, 0, 0, 0, 0, time.UTC)},
		{"someday", time.Date(9999, 12, 30, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDateExpr(tt.input, ref)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDateExpr_PeriodBoundaries(t *testing.T) {
	// ref = Wednesday 2026-04-15
	tests := []struct {
		input string
		want  time.Time
	}{
		// Week boundaries (Mon-Sun)
		{"sow", time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)},     // Monday
		{"eow", time.Date(2026, 4, 19, 23, 59, 59, 0, time.UTC)},  // Sunday
		{"soww", time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)},    // Monday
		{"eoww", time.Date(2026, 4, 17, 23, 59, 59, 0, time.UTC)}, // Friday
		// Month boundaries
		{"som", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)},
		{"eom", time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)},
		// Quarter boundaries (Q2: Apr-Jun)
		{"soq", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)},
		{"eoq", time.Date(2026, 6, 30, 23, 59, 59, 0, time.UTC)},
		// Year boundaries
		{"soy", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"eoy", time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDateExpr(tt.input, ref)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDateExpr_NextPrevPeriods(t *testing.T) {
	// ref = Wednesday 2026-04-15
	tests := []struct {
		input string
		want  time.Time
	}{
		{"sonw", time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)},
		{"eonw", time.Date(2026, 4, 26, 23, 59, 59, 0, time.UTC)},
		{"sopw", time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC)},
		{"eopw", time.Date(2026, 4, 12, 23, 59, 59, 0, time.UTC)},
		{"sonm", time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
		{"eonm", time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC)},
		{"sopm", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
		{"eopm", time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)},
		{"sonq", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
		{"eonq", time.Date(2026, 9, 30, 23, 59, 59, 0, time.UTC)},
		{"sopq", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"eopq", time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)},
		{"sony", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"eony", time.Date(2027, 12, 31, 23, 59, 59, 0, time.UTC)},
		{"sopy", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"eopy", time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDateExpr(tt.input, ref)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDateExpr_Weekdays(t *testing.T) {
	// ref = Wednesday 2026-04-15
	tests := []struct {
		input string
		want  time.Time
	}{
		{"monday", time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)},
		{"mon", time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)},
		{"tuesday", time.Date(2026, 4, 21, 0, 0, 0, 0, time.UTC)},
		{"tue", time.Date(2026, 4, 21, 0, 0, 0, 0, time.UTC)},
		{"wednesday", time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC)}, // next wed, not today
		{"wed", time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC)},
		{"thursday", time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)},
		{"thu", time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)},
		{"friday", time.Date(2026, 4, 17, 0, 0, 0, 0, time.UTC)},
		{"fri", time.Date(2026, 4, 17, 0, 0, 0, 0, time.UTC)},
		{"saturday", time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)},
		{"sat", time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)},
		{"sunday", time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC)},
		{"sun", time.Date(2026, 4, 19, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDateExpr(tt.input, ref)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDateExpr_Months(t *testing.T) {
	// ref = April 2026
	tests := []struct {
		input string
		want  time.Time
	}{
		{"january", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"jan", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"february", time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)},
		{"feb", time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)},
		{"march", time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)},
		{"mar", time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)},
		{"april", time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC)}, // next april, not this month
		{"apr", time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC)},
		{"may", time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
		{"june", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
		{"jun", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)},
		{"december", time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)},
		{"dec", time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDateExpr(tt.input, ref)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDateExpr_RelativeDurations(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		{"2d", ref.AddDate(0, 0, 2)},
		{"3w", ref.AddDate(0, 0, 21)},
		{"1mo", ref.AddDate(0, 1, 0)},
		{"8h", ref.Add(8 * time.Hour)},
		{"30min", ref.Add(30 * time.Minute)},
		{"1q", ref.AddDate(0, 3, 0)},
		{"1y", ref.AddDate(1, 0, 0)},
		{"90s", ref.Add(90 * time.Second)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDateExpr(tt.input, ref)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDateExpr_ISO(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		{"2026-04-20", time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)},
		{"2026-04-20T14:30:00", time.Date(2026, 4, 20, 14, 30, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDateExpr(tt.input, ref)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDateExpr_Arithmetic(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
	}{
		{"friday+1w", time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC)},   // Apr 17 + 7d
		{"eom-1d", time.Date(2026, 4, 29, 23, 59, 59, 0, time.UTC)},   // Apr 30 23:59:59 - 1d
		{"today+8h", time.Date(2026, 4, 15, 8, 0, 0, 0, time.UTC)},    // today 00:00 + 8h
		{"tomorrow-1d", time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)}, // tomorrow - 1d = today
		{"2026-04-20+2w", time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)},
		{"now+30min", ref.Add(30 * time.Minute)},
		{"sow+3d", time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)}, // Mon + 3d = Thu
		{"sonm-1d", time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDateExpr(tt.input, ref)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDateExpr_CaseInsensitive(t *testing.T) {
	tests := []string{"Today", "TODAY", "Friday", "FRIDAY", "EOD", "SOM"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDateExpr(input, ref)
			require.NoError(t, err)
		})
	}
}

func TestParseDateExpr_Errors(t *testing.T) {
	tests := []string{
		"",
		"notadate",
		"2026-13-01",
		"today+",
		"today+abc",
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseDateExpr(input, ref)
			assert.Error(t, err)
		})
	}
}
