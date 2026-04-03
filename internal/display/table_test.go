package display

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/abushady/tdo/internal/domain"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func datePtr(y int, m time.Month, d int) *time.Time {
	t := date(y, m, d)
	return &t
}

// --- FormatRelativeDue ---

func TestFormatRelativeDue_Today(t *testing.T) {
	now := date(2025, 4, 3)
	assert.Equal(t, "today", FormatRelativeDue(date(2025, 4, 3), now))
}

func TestFormatRelativeDue_Tomorrow(t *testing.T) {
	now := date(2025, 4, 3)
	assert.Equal(t, "tomorrow", FormatRelativeDue(date(2025, 4, 4), now))
}

func TestFormatRelativeDue_Yesterday(t *testing.T) {
	now := date(2025, 4, 3)
	assert.Equal(t, "yesterday", FormatRelativeDue(date(2025, 4, 2), now))
}

func TestFormatRelativeDue_PastDays(t *testing.T) {
	now := date(2025, 4, 3)
	assert.Equal(t, "-2d", FormatRelativeDue(date(2025, 4, 1), now))
	assert.Equal(t, "-30d", FormatRelativeDue(date(2025, 3, 4), now))
}

func TestFormatRelativeDue_FutureDays(t *testing.T) {
	now := date(2025, 4, 3)
	assert.Equal(t, "+2d", FormatRelativeDue(date(2025, 4, 5), now))
	assert.Equal(t, "+14d", FormatRelativeDue(date(2025, 4, 17), now))
}

func TestFormatRelativeDue_FarFuture(t *testing.T) {
	now := date(2025, 4, 3)
	assert.Equal(t, "2025-06-15", FormatRelativeDue(date(2025, 6, 15), now))
}

// --- truncate ---

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "hello", truncate("hello", 5))
	assert.Equal(t, "he...", truncate("hello world", 5))
	assert.Equal(t, "this is a really long str...", truncate("this is a really long string that goes on", 28))
}

// --- FormatTaskTable ---

func TestFormatTaskTable_Empty(t *testing.T) {
	out, posMap := FormatTaskTable(nil, "", time.Now())
	assert.Contains(t, out, "No tasks found")
	assert.Empty(t, posMap)
}

func TestFormatTaskTable_PositionMap(t *testing.T) {
	tasks := []domain.Task{
		{ID: "abc123", Content: "First task"},
		{ID: "def456", Content: "Second task"},
		{ID: "ghi789", Content: "Third task"},
	}
	_, posMap := FormatTaskTable(tasks, "", time.Now())
	assert.Len(t, posMap, 3)
	assert.Equal(t, "abc123", posMap[1])
	assert.Equal(t, "def456", posMap[2])
	assert.Equal(t, "ghi789", posMap[3])
}

func TestFormatTaskTable_RowCount(t *testing.T) {
	tasks := []domain.Task{
		{ID: "a", Content: "Task 1"},
		{ID: "b", Content: "Task 2"},
	}
	out, _ := FormatTaskTable(tasks, "", time.Now())
	// 1 header line + 2 data lines
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	assert.Equal(t, 3, len(lines))
}

func TestFormatTaskTable_ContentTruncation(t *testing.T) {
	longContent := "This is a very long task content that should be truncated"
	tasks := []domain.Task{
		{ID: "x", Content: longContent},
	}
	out, _ := FormatTaskTable(tasks, "", time.Now())
	assert.Contains(t, out, "...")
	assert.NotContains(t, out, longContent)
}

func TestFormatTaskTable_DueFormatting(t *testing.T) {
	now := date(2025, 4, 3)
	tasks := []domain.Task{
		{ID: "a", Content: "Due today", Due: datePtr(2025, 4, 3)},
		{ID: "b", Content: "Due tomorrow", Due: datePtr(2025, 4, 4)},
		{ID: "c", Content: "Overdue", Due: datePtr(2025, 4, 1)},
	}
	out, _ := FormatTaskTable(tasks, "", now)
	assert.Contains(t, out, "today")
	assert.Contains(t, out, "tomorrow")
	assert.Contains(t, out, "-2d")
}

func TestClassifyTask_OverdueUseDueNotScheduled(t *testing.T) {
	now := date(2025, 4, 3)

	// Overdue scheduled but no due → should be normal, not overdue
	schedOnly := domain.Task{
		ID:        "a",
		Content:   "Scheduled overdue",
		Scheduled: datePtr(2025, 4, 1),
	}
	assert.Equal(t, statusNormal, classifyTask(schedOnly, "", now))

	// Due overdue → should be overdue regardless of Scheduled
	dueOverdue := domain.Task{
		ID:        "b",
		Content:   "Due overdue",
		Due:       datePtr(2025, 4, 1),
		Scheduled: datePtr(2025, 4, 10),
	}
	assert.Equal(t, statusOverdue, classifyTask(dueOverdue, "", now))

	// Due today → should be dueTodayStatus regardless of Scheduled
	dueToday := domain.Task{
		ID:        "c",
		Content:   "Due today",
		Due:       datePtr(2025, 4, 3),
		Scheduled: datePtr(2025, 3, 1),
	}
	assert.Equal(t, statusDueToday, classifyTask(dueToday, "", now))
}

func TestFormatTaskTable_SchedColumn(t *testing.T) {
	now := date(2025, 4, 3)
	tasks := []domain.Task{
		{ID: "a", Content: "With sched", Scheduled: datePtr(2025, 4, 5)},
		{ID: "b", Content: "No sched"},
	}
	out, _ := FormatTaskTable(tasks, "", now)
	assert.Contains(t, out, "Sched")
	assert.Contains(t, out, "+2d")
}

func TestFormatTaskTable_UrgencyColumn(t *testing.T) {
	now := date(2025, 4, 3)
	tasks := []domain.Task{
		{
			ID:        "abc123",
			Content:   "High priority overdue",
			Priority:  domain.PriorityH,
			Due:       datePtr(2025, 4, 1),
			Project:   "Work",
			Labels:    []string{"urgent"},
			CreatedAt: date(2025, 3, 1),
		},
		{
			ID:        "def456",
			Content:   "No attributes",
			CreatedAt: now,
		},
	}
	out, _ := FormatTaskTable(tasks, "", now)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// Task with priority+due+project+labels should have a non-zero urgency
	assert.Contains(t, lines[1], "abc123")
	assert.NotContains(t, lines[1], "  0.0")
	// Task with no attributes should show 0.0
	assert.Contains(t, lines[2], "def456")
	assert.Contains(t, lines[2], "0.0")
}

// --- FormatProjectList ---

func TestFormatProjectList_Empty(t *testing.T) {
	out := FormatProjectList(nil)
	assert.Contains(t, out, "No projects found")
}

func TestFormatProjectList(t *testing.T) {
	projects := []domain.Project{
		{ID: "p1", Name: "Backend"},
		{ID: "p2", Name: "Frontend"},
	}
	out := FormatProjectList(projects)
	assert.Contains(t, out, "Backend")
	assert.Contains(t, out, "Frontend")
}

// --- FormatLabelList ---

func TestFormatLabelList_Empty(t *testing.T) {
	out := FormatLabelList(nil)
	assert.Contains(t, out, "No labels found")
}

func TestFormatLabelList(t *testing.T) {
	labels := []domain.Label{
		{ID: "l1", Name: "urgent"},
		{ID: "l2", Name: "docs"},
	}
	out := FormatLabelList(labels)
	assert.Contains(t, out, "urgent")
	assert.Contains(t, out, "docs")
}

// --- FormatTaskDetail ---

func TestFormatTaskDetail_Basic(t *testing.T) {
	due := date(2025, 4, 3)
	task := domain.Task{
		ID:        "abc123",
		Content:   "Fix login bug",
		Project:   "Backend",
		Priority:  domain.PriorityH,
		Due:       &due,
		Labels:    []string{"urgent", "backend"},
		CreatedAt: date(2025, 3, 15),
	}
	out := FormatTaskDetail(task, nil, 21.7)
	assert.Contains(t, out, "Fix login bug")
	assert.Contains(t, out, "abc123")
	assert.Contains(t, out, "Backend")
	assert.Contains(t, out, "H")
	assert.Contains(t, out, "urgent, backend")
	assert.Contains(t, out, "21.7")
	assert.Contains(t, out, "2025-03-15")
}

func TestFormatTaskDetail_WithComments(t *testing.T) {
	task := domain.Task{
		ID:        "abc123",
		Content:   "Some task",
		CreatedAt: date(2025, 3, 15),
	}
	comments := []domain.Comment{
		{Content: "First note", PostedAt: date(2025, 3, 20)},
		{Content: "Second note", PostedAt: date(2025, 3, 21)},
	}
	out := FormatTaskDetail(task, comments, 5.0)
	assert.Contains(t, out, "Annotations:")
	assert.Contains(t, out, "First note")
	assert.Contains(t, out, "2025-03-20")
	assert.Contains(t, out, "Second note")
}

func TestFormatTaskDetail_ActiveStatus(t *testing.T) {
	task := domain.Task{
		ID:        "abc",
		Content:   "Active task",
		Labels:    []string{"now"},
		CreatedAt: date(2025, 3, 15),
	}
	out := FormatTaskDetail(task, nil, 10.0)
	assert.Contains(t, out, "Active")
}

func TestFormatTaskDetail_CompletedStatus(t *testing.T) {
	task := domain.Task{
		ID:          "abc",
		Content:     "Done task",
		IsCompleted: true,
		CreatedAt:   date(2025, 3, 15),
	}
	out := FormatTaskDetail(task, nil, 0.0)
	assert.Contains(t, out, "Completed")
}

func TestFormatTaskDetail_WithScheduled(t *testing.T) {
	sched := date(2025, 4, 10)
	due := date(2025, 4, 15)
	task := domain.Task{
		ID:        "abc",
		Content:   "Both dates",
		Scheduled: &sched,
		Due:       &due,
		CreatedAt: date(2025, 3, 15),
	}
	out := FormatTaskDetail(task, nil, 5.0)
	assert.Contains(t, out, "Scheduled")
	assert.Contains(t, out, "2025-04-10")
	assert.Contains(t, out, "Due")
	assert.Contains(t, out, "2025-04-15")
}
