package undo_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/abushady/tdo/internal/domain"
	"github.com/abushady/tdo/internal/undo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPushPopLIFO(t *testing.T) {
	dir := t.TempDir()
	log := undo.NewLog(dir, 10)

	e1 := undo.Entry{Op: undo.OpDone, TaskID: "t1", Timestamp: time.Now()}
	e2 := undo.Entry{Op: undo.OpDelete, TaskID: "t2", Timestamp: time.Now()}

	require.NoError(t, log.Push(e1))
	require.NoError(t, log.Push(e2))

	got, err := log.Pop()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "t2", got.TaskID)

	got, err = log.Pop()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "t1", got.TaskID)
}

func TestPopOnEmpty(t *testing.T) {
	dir := t.TempDir()
	log := undo.NewLog(dir, 10)

	got, err := log.Pop()
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPeekDoesNotRemove(t *testing.T) {
	dir := t.TempDir()
	log := undo.NewLog(dir, 10)

	e := undo.Entry{Op: undo.OpDone, TaskID: "t1", Timestamp: time.Now()}
	require.NoError(t, log.Push(e))

	got, err := log.Peek()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "t1", got.TaskID)

	// Second peek still returns it.
	got2, err := log.Peek()
	require.NoError(t, err)
	require.NotNil(t, got2)
	assert.Equal(t, "t1", got2.TaskID)
}

func TestPeekOnEmpty(t *testing.T) {
	dir := t.TempDir()
	log := undo.NewLog(dir, 10)

	got, err := log.Peek()
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestMaxSizeEviction(t *testing.T) {
	dir := t.TempDir()
	log := undo.NewLog(dir, 10)

	for i := 0; i < 12; i++ {
		require.NoError(t, log.Push(undo.Entry{
			Op:        undo.OpDone,
			TaskID:    fmt.Sprintf("t%d", i),
			Timestamp: time.Now(),
		}))
	}

	var ids []string
	for {
		e, err := log.Pop()
		require.NoError(t, err)
		if e == nil {
			break
		}
		ids = append(ids, e.TaskID)
	}
	assert.Len(t, ids, 10)
	assert.NotContains(t, ids, "t0")
	assert.NotContains(t, ids, "t1")
}

func TestPersistenceAcrossInstances(t *testing.T) {
	dir := t.TempDir()

	log1 := undo.NewLog(dir, 10)
	require.NoError(t, log1.Push(undo.Entry{Op: undo.OpAdd, CreatedID: "new-id", Timestamp: time.Now()}))

	log2 := undo.NewLog(dir, 10)
	got, err := log2.Pop()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, undo.OpAdd, got.Op)
	assert.Equal(t, "new-id", got.CreatedID)
}

func TestCorruptFileReturnedAsEmpty(t *testing.T) {
	dir := t.TempDir()

	// Write corrupt JSON directly to the undo log file.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "undo_log.json"), []byte("not json!!"), 0o600))

	log := undo.NewLog(dir, 10)

	// Peek on corrupt file should return nil, nil (treated as empty).
	got, err := log.Peek()
	require.NoError(t, err)
	assert.Nil(t, got)

	// Pop on corrupt file should return nil, nil.
	got, err = log.Pop()
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestSnapshotToCreateParams(t *testing.T) {
	scheduled := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	task := &domain.Task{
		Content:     "Buy groceries",
		Description: "milk and eggs",
		Priority:    domain.PriorityH,
		Scheduled:   &scheduled,
		Due:         &due,
		Labels:      []string{"shopping", "urgent"},
		ProjectID:   "proj-123",
		ParentID:    "parent-456",
		Recurrence:  "every day",
	}

	p := undo.SnapshotToCreateParams(task)

	assert.Equal(t, "Buy groceries", p.Content)
	assert.Equal(t, "milk and eggs", p.Description)
	assert.Equal(t, domain.PriorityH, p.Priority)
	assert.Equal(t, "2026-04-15", p.ScheduledString)
	assert.Equal(t, "2026-04-30", p.DueDate)
	assert.Equal(t, []string{"shopping", "urgent"}, p.Labels)
	assert.Equal(t, "proj-123", p.ProjectID)
	assert.Equal(t, "parent-456", p.ParentID)
	assert.Equal(t, "every day", p.Recurrence)
}

func TestSnapshotToCreateParams_NilTimes(t *testing.T) {
	task := &domain.Task{
		Content: "Simple task",
	}

	p := undo.SnapshotToCreateParams(task)

	assert.Equal(t, "Simple task", p.Content)
	assert.Empty(t, p.ScheduledString)
	assert.Empty(t, p.DueDate)
}

func TestSnapshotToUpdateParams(t *testing.T) {
	scheduled := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	task := &domain.Task{
		Content:     "Fix bug",
		Description: "in auth module",
		Priority:    domain.PriorityM,
		Scheduled:   &scheduled,
		Due:         &due,
		Labels:      []string{"backend"},
		ProjectID:   "proj-999",
	}

	p := undo.SnapshotToUpdateParams(task)

	require.NotNil(t, p.Content)
	assert.Equal(t, "Fix bug", *p.Content)
	require.NotNil(t, p.Description)
	assert.Equal(t, "in auth module", *p.Description)
	require.NotNil(t, p.Priority)
	assert.Equal(t, domain.PriorityM, *p.Priority)
	require.NotNil(t, p.ScheduledString)
	assert.Equal(t, "2026-04-15", *p.ScheduledString)
	require.NotNil(t, p.DueDate)
	assert.Equal(t, "2026-04-30", *p.DueDate)
	assert.Equal(t, []string{"backend"}, p.Labels)
	require.NotNil(t, p.ProjectID)
	assert.Equal(t, "proj-999", *p.ProjectID)
}

func TestSnapshotToUpdateParams_NilTimes(t *testing.T) {
	task := &domain.Task{
		Content:  "No dates task",
		Priority: domain.PriorityNone,
	}

	p := undo.SnapshotToUpdateParams(task)

	require.NotNil(t, p.ScheduledString)
	assert.Equal(t, "", *p.ScheduledString)
	require.NotNil(t, p.DueDate)
	assert.Equal(t, "", *p.DueDate)
}
