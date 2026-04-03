package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func timePtr(t time.Time) *time.Time { return &t }

func TestUrgency(t *testing.T) {
	now := time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
	nowLabel := "now"

	// ── Priority factor ──────────────────────────────────────────

	t.Run("high priority adds 6.0 to urgency", func(t *testing.T) {
		task := Task{Priority: PriorityH, CreatedAt: now}
		assert.InDelta(t, 6.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("medium priority adds 3.9 to urgency", func(t *testing.T) {
		task := Task{Priority: PriorityM, CreatedAt: now}
		assert.InDelta(t, 3.9, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("low priority adds 1.8 to urgency", func(t *testing.T) {
		task := Task{Priority: PriorityL, CreatedAt: now}
		assert.InDelta(t, 1.8, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("no priority adds 0 to urgency", func(t *testing.T) {
		task := Task{Priority: PriorityNone, CreatedAt: now}
		assert.InDelta(t, 0.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	// ── Due date factor ──────────────────────────────────────────

	t.Run("no due date adds 0 to urgency", func(t *testing.T) {
		task := Task{CreatedAt: now}
		assert.InDelta(t, 0.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("due 7 or more days overdue adds 12.0", func(t *testing.T) {
		due := now.Add(-7 * 24 * time.Hour)
		task := Task{Due: timePtr(due), CreatedAt: now}
		assert.InDelta(t, 12.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("due 30 days overdue adds 12.0", func(t *testing.T) {
		due := now.Add(-30 * 24 * time.Hour)
		task := Task{Due: timePtr(due), CreatedAt: now}
		assert.InDelta(t, 12.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("due today applies linear interpolation", func(t *testing.T) {
		// days_overdue = 0, factor = ((0+14)*0.8/21) + 0.2 = (11.2/21) + 0.2 ≈ 0.7333
		due := now
		task := Task{Due: timePtr(due), CreatedAt: now}
		expected := 12.0 * ((14.0 * 0.8 / 21.0) + 0.2) // 12 * 0.7333 = 8.8
		assert.InDelta(t, expected, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("due in 14 days is at low end of interpolation", func(t *testing.T) {
		// days_overdue = -14, factor = ((-14+14)*0.8/21) + 0.2 = 0.2
		due := now.Add(14 * 24 * time.Hour)
		task := Task{Due: timePtr(due), CreatedAt: now}
		expected := 12.0 * 0.2 // 2.4
		assert.InDelta(t, expected, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("due in 7 days applies interpolation midpoint", func(t *testing.T) {
		// days_overdue = -7, factor = ((-7+14)*0.8/21) + 0.2 = (7*0.8/21) + 0.2 = 0.2667 + 0.2 = 0.4667
		due := now.Add(7 * 24 * time.Hour)
		task := Task{Due: timePtr(due), CreatedAt: now}
		expected := 12.0 * ((7.0 * 0.8 / 21.0) + 0.2) // 12 * 0.4667 = 5.6
		assert.InDelta(t, expected, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("due more than 14 days out adds 12*0.2", func(t *testing.T) {
		due := now.Add(30 * 24 * time.Hour)
		task := Task{Due: timePtr(due), CreatedAt: now}
		expected := 12.0 * 0.2 // 2.4
		assert.InDelta(t, expected, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	// ── Active (started) factor ──────────────────────────────────

	t.Run("active task with now label adds 15.0", func(t *testing.T) {
		task := Task{Labels: []string{"now"}, CreatedAt: now}
		// next=15.0 + tags(1 label)=0.8
		assert.InDelta(t, 15.8, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("task without now label adds 0 for next", func(t *testing.T) {
		task := Task{Labels: []string{"other"}, CreatedAt: now}
		// only tag factor: 1 tag = 0.8
		expected := 1.0 * 0.8
		assert.InDelta(t, expected, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("custom now label is respected", func(t *testing.T) {
		task := Task{Labels: []string{"started"}, CreatedAt: now}
		// "started" is the nowLabel, so next=15.0 + tags(1)=0.8
		assert.InDelta(t, 15.8, CalculateUrgency(task, "started", now), 0.001)
	})

	// ── Age factor ───────────────────────────────────────────────

	t.Run("age 0 days adds 0 to urgency", func(t *testing.T) {
		task := Task{CreatedAt: now}
		assert.InDelta(t, 0.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("age 30 days adds proportional urgency", func(t *testing.T) {
		task := Task{CreatedAt: now.AddDate(0, 0, -30)}
		expected := 2.0 * (30.0 / 365.0) // ≈ 0.1644
		assert.InDelta(t, expected, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("age 365 days caps at coefficient 2.0", func(t *testing.T) {
		task := Task{CreatedAt: now.AddDate(0, 0, -365)}
		assert.InDelta(t, 2.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("age 730 days still caps at 2.0", func(t *testing.T) {
		task := Task{CreatedAt: now.AddDate(0, 0, -730)}
		assert.InDelta(t, 2.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	// ── Annotations factor ───────────────────────────────────────

	t.Run("0 annotations adds 0", func(t *testing.T) {
		task := Task{CommentCount: 0, CreatedAt: now}
		assert.InDelta(t, 0.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("1 annotation adds 0.8", func(t *testing.T) {
		task := Task{CommentCount: 1, CreatedAt: now}
		assert.InDelta(t, 0.8, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("2 annotations adds 0.9", func(t *testing.T) {
		task := Task{CommentCount: 2, CreatedAt: now}
		assert.InDelta(t, 0.9, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("3 annotations adds 1.0", func(t *testing.T) {
		task := Task{CommentCount: 3, CreatedAt: now}
		assert.InDelta(t, 1.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("5 annotations still adds 1.0", func(t *testing.T) {
		task := Task{CommentCount: 5, CreatedAt: now}
		assert.InDelta(t, 1.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	// ── Tags factor ──────────────────────────────────────────────

	t.Run("0 labels adds 0 for tags", func(t *testing.T) {
		task := Task{CreatedAt: now}
		assert.InDelta(t, 0.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("1 label adds 0.8 for tags", func(t *testing.T) {
		// label is not nowLabel so no active bonus
		task := Task{Labels: []string{"bug"}, CreatedAt: now}
		assert.InDelta(t, 0.8, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("2 labels adds 0.9 for tags", func(t *testing.T) {
		task := Task{Labels: []string{"bug", "urgent"}, CreatedAt: now}
		assert.InDelta(t, 0.9, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("3 labels adds 1.0 for tags", func(t *testing.T) {
		task := Task{Labels: []string{"bug", "urgent", "backend"}, CreatedAt: now}
		assert.InDelta(t, 1.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("now label counts as a tag too", func(t *testing.T) {
		// "now" label: next=15.0, tags: 1 label = 0.8
		task := Task{Labels: []string{"now"}, CreatedAt: now}
		assert.InDelta(t, 15.0+0.8, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	// ── Project factor ───────────────────────────────────────────

	t.Run("task with project adds 1.0", func(t *testing.T) {
		task := Task{Project: "work", CreatedAt: now}
		assert.InDelta(t, 1.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("task without project adds 0", func(t *testing.T) {
		task := Task{CreatedAt: now}
		assert.InDelta(t, 0.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	// ── Scheduled vs Due ─────────────────────────────────────────

	t.Run("scheduled date does not affect urgency", func(t *testing.T) {
		// Only Due (deadline) drives urgency, not Scheduled.
		sched := now.Add(-10 * 24 * time.Hour) // 10 days overdue scheduled
		task := Task{Scheduled: timePtr(sched), CreatedAt: now}
		assert.InDelta(t, 0.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("urgency uses Due not Scheduled when both set", func(t *testing.T) {
		due := now.Add(-7 * 24 * time.Hour)   // 7 days overdue → factor 1.0
		sched := now.Add(30 * 24 * time.Hour) // 30 days in future (irrelevant)
		task := Task{Due: timePtr(due), Scheduled: timePtr(sched), CreatedAt: now}
		// Only due contributes: 12.0 * 1.0 = 12.0
		assert.InDelta(t, 12.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	// ── Edge cases ───────────────────────────────────────────────

	t.Run("zero task has zero urgency", func(t *testing.T) {
		task := Task{CreatedAt: now}
		assert.InDelta(t, 0.0, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("all factors at maximum", func(t *testing.T) {
		due := now.Add(-10 * 24 * time.Hour) // 10 days overdue -> due factor = 1.0
		task := Task{
			Priority:     PriorityH,
			Due:          timePtr(due),
			Labels:       []string{"now", "bug", "urgent"}, // active + 3 tags
			Project:      "work",
			CreatedAt:    now.AddDate(-2, 0, 0), // 730 days, age capped at 1.0
			CommentCount: 5,                     // annotations capped at 1.0
		}
		// priority H:     6.0 * 1.0 = 6.0
		// due (overdue):  12.0 * 1.0 = 12.0
		// next:           15.0 * 1.0 = 15.0
		// age:            2.0 * 1.0 = 2.0
		// annotations:    1.0 * 1.0 = 1.0
		// tags (3):       1.0 * 1.0 = 1.0
		// project:        1.0 * 1.0 = 1.0
		expected := 6.0 + 12.0 + 15.0 + 2.0 + 1.0 + 1.0 + 1.0
		assert.InDelta(t, expected, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	// ── Combined / worked example ────────────────────────────────

	t.Run("combined: H priority + due today + active + 2 tags + 1 annotation + project + 30 days old", func(t *testing.T) {
		due := now
		task := Task{
			Priority:     PriorityH,
			Due:          timePtr(due),
			Labels:       []string{"now", "backend"}, // active + 2 tags
			Project:      "tdo",
			CreatedAt:    now.AddDate(0, 0, -30),
			CommentCount: 1,
		}
		// priority H:     6.0 * 1.0 = 6.0
		// due (today):    12.0 * ((0+14)*0.8/21 + 0.2) = 12.0 * 0.73333 = 8.8
		// next:           15.0 * 1.0 = 15.0
		// age:            2.0 * (30/365) ≈ 0.16438
		// annotations:    1.0 * 0.8 = 0.8
		// tags (2):       1.0 * 0.9 = 0.9
		// project:        1.0 * 1.0 = 1.0
		dueFactor := (14.0*0.8/21.0 + 0.2)
		expected := 6.0 + 12.0*dueFactor + 15.0 + 2.0*(30.0/365.0) + 0.8 + 0.9 + 1.0
		assert.InDelta(t, expected, CalculateUrgency(task, nowLabel, now), 0.001)
	})

	t.Run("combined: L priority + due in 3 days + no active + 1 tag + no annotations + no project + 180 days old", func(t *testing.T) {
		due := now.Add(3 * 24 * time.Hour)
		task := Task{
			Priority:  PriorityL,
			Due:       timePtr(due),
			Labels:    []string{"personal"},
			CreatedAt: now.AddDate(0, 0, -180),
		}
		// priority L:     1.8
		// due (-3 days):  12.0 * ((-3+14)*0.8/21 + 0.2) = 12.0 * (11*0.8/21 + 0.2) = 12.0 * (0.41905 + 0.2) = 12.0 * 0.61905 ≈ 7.4286
		// active:         0
		// age:            2.0 * (180/365) ≈ 0.98630
		// annotations:    0
		// tags (1):       0.8
		// project:        0
		dueFactor := (11.0*0.8/21.0 + 0.2)
		expected := 1.8 + 12.0*dueFactor + 2.0*(180.0/365.0) + 0.8
		assert.InDelta(t, expected, CalculateUrgency(task, nowLabel, now), 0.001)
	})
}
