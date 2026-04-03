package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	tasks := []Task{
		{ID: "abc123", Content: "Fix login bug"},
		{ID: "def456", Content: "Report bug"},
		{ID: "ghi789", Content: "Write documentation"},
		{ID: "42", Content: "Numeric ID task"},
	}

	tests := []struct {
		name      string
		tasks     []Task
		positions map[int]string
		input     string
		wantID    string
		wantErr   string
	}{
		{
			name:      "exact ID match",
			tasks:     tasks,
			positions: map[int]string{},
			input:     "abc123",
			wantID:    "abc123",
		},
		{
			name:      "positional match",
			tasks:     tasks,
			positions: map[int]string{1: "abc123", 2: "def456"},
			input:     "1",
			wantID:    "abc123",
		},
		{
			name:      "position not found",
			tasks:     tasks,
			positions: map[int]string{1: "abc123", 2: "def456"},
			input:     "3",
			wantErr:   "position 3 not found",
		},
		{
			name:      "fuzzy prefix match",
			tasks:     tasks,
			positions: map[int]string{},
			input:     "Write",
			wantID:    "ghi789",
		},
		{
			name:      "fuzzy substring match",
			tasks:     tasks,
			positions: map[int]string{},
			input:     "documentation",
			wantID:    "ghi789",
		},
		{
			name:      "fuzzy case insensitive",
			tasks:     tasks,
			positions: map[int]string{},
			input:     "write",
			wantID:    "ghi789",
		},
		{
			name:      "fuzzy ambiguous match",
			tasks:     tasks,
			positions: map[int]string{},
			input:     "bug",
			wantErr:   "ambiguous",
		},
		{
			name:      "fuzzy no match",
			tasks:     tasks,
			positions: map[int]string{},
			input:     "nonexistent",
			wantErr:   "no task found",
		},
		{
			name:      "exact ID wins over fuzzy",
			tasks:     tasks,
			positions: map[int]string{},
			input:     "def456",
			wantID:    "def456",
		},
		{
			name:      "position wins for numeric input over fuzzy",
			tasks:     tasks,
			positions: map[int]string{1: "abc123"},
			input:     "1",
			wantID:    "abc123",
		},
		{
			name:      "numeric input matches exact ID before position",
			tasks:     tasks,
			positions: map[int]string{42: "def456"},
			input:     "42",
			wantID:    "42",
		},
		{
			name:      "ambiguous error lists matching tasks",
			tasks:     tasks,
			positions: map[int]string{},
			input:     "bug",
			wantErr:   "Fix login bug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Resolver{
				Tasks:     tt.tasks,
				Positions: tt.positions,
			}
			result, err := r.Resolve(tt.input)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantID, result.TaskID)
				assert.NotNil(t, result.Task)
				assert.Equal(t, tt.wantID, result.Task.ID)
			}
		})
	}
}
