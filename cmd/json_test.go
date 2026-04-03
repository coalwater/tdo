package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/abushady/tdo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToTaskJSON(t *testing.T) {
	now := time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
	due := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)
	created := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		task     domain.Task
		wantID   string
		wantPri  string
		wantDue  string
		wantUrgN bool // urgency > 0
	}{
		{
			name: "full task with all fields",
			task: domain.Task{
				ID:           "abc123",
				Content:      "Fix the bug",
				Description:  "Check auth",
				Priority:     domain.PriorityH,
				Due:          &due,
				Labels:       []string{"urgent", "backend"},
				Project:      "Work",
				ProjectID:    "proj-1",
				CreatedAt:    created,
				CommentCount: 2,
				URL:          "https://todoist.com/task/abc123",
			},
			wantID:   "abc123",
			wantPri:  "H",
			wantDue:  "2026-04-05",
			wantUrgN: true,
		},
		{
			name:     "empty task has zero urgency and empty labels array",
			task:     domain.Task{ID: "empty"},
			wantID:   "empty",
			wantPri:  "",
			wantDue:  "",
			wantUrgN: false,
		},
		{
			name: "nil labels become empty array",
			task: domain.Task{
				ID:     "no-labels",
				Labels: nil,
			},
			wantID: "no-labels",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tj := toTaskJSON(tt.task, "now", now)

			assert.Equal(t, tt.wantID, tj.ID)
			assert.Equal(t, tt.wantPri, tj.Priority)
			assert.Equal(t, tt.wantDue, tj.Due)
			assert.NotNil(t, tj.Labels, "labels should never be nil")

			if tt.wantUrgN {
				assert.Greater(t, tj.Urgency, 0.0)
			}
		})
	}
}

func TestToCommentJSON(t *testing.T) {
	posted := time.Date(2026, 4, 1, 8, 0, 0, 0, time.UTC)

	cj := toCommentJSON(domain.Comment{
		ID:       "c1",
		TaskID:   "t1",
		Content:  "Found root cause",
		PostedAt: posted,
	})

	assert.Equal(t, "c1", cj.ID)
	assert.Equal(t, "t1", cj.TaskID)
	assert.Equal(t, "Found root cause", cj.Content)
	assert.Equal(t, "2026-04-01T08:00:00Z", cj.PostedAt)
}

func TestToCommentJSONZeroTime(t *testing.T) {
	cj := toCommentJSON(domain.Comment{ID: "c2", Content: "test"})
	assert.Empty(t, cj.PostedAt)
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer

	data := []projectJSON{
		{ID: "p1", Name: "Inbox"},
		{ID: "p2", Name: "Work"},
	}

	err := writeJSON(&buf, data)
	require.NoError(t, err)

	var result []projectJSON
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Inbox", result[0].Name)
}

func TestWriteJSONRoundTripsTaskFields(t *testing.T) {
	now := time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
	due := time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC)

	task := domain.Task{
		ID:           "t1",
		Content:      "Test task",
		Priority:     domain.PriorityM,
		Due:          &due,
		Labels:       []string{"claude", "now"},
		Project:      "Next",
		ProjectID:    "proj-1",
		CreatedAt:    now.Add(-24 * time.Hour),
		CommentCount: 1,
	}

	var buf bytes.Buffer
	tj := toTaskJSON(task, "now", now)
	err := writeJSON(&buf, tj)
	require.NoError(t, err)

	var result taskJSON
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, "t1", result.ID)
	assert.Equal(t, "Test task", result.Content)
	assert.Equal(t, "M", result.Priority)
	assert.Equal(t, "2026-04-03", result.Due)
	assert.Equal(t, []string{"claude", "now"}, result.Labels)
	assert.Equal(t, "Next", result.Project)
	assert.Equal(t, 1, result.CommentCount)
	assert.Greater(t, result.Urgency, 0.0)
}
