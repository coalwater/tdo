package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/abushady/tdo/internal/cache"
	"github.com/abushady/tdo/internal/domain"
	"github.com/abushady/tdo/internal/undo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// undoMockBackend extends mockBackend with additional tracking for undo ops.
type undoMockBackend struct {
	mockBackend
	reopenedID    string
	deletedID     string
	updatedID     string
	updatedParams domain.UpdateParams
	createdParams domain.CreateParams
	createdResult *domain.Task
	reopenErr     error
	deleteErr     error
	updateErr     error
	createErr     error
}

func (m *undoMockBackend) ReopenTask(_ context.Context, id string) error {
	m.reopenedID = id
	return m.reopenErr
}

func (m *undoMockBackend) DeleteTask(_ context.Context, id string) error {
	m.deletedID = id
	return m.deleteErr
}

func (m *undoMockBackend) UpdateTask(_ context.Context, id string, params domain.UpdateParams) error {
	m.updatedID = id
	m.updatedParams = params
	return m.updateErr
}

func (m *undoMockBackend) CreateTask(_ context.Context, params domain.CreateParams) (*domain.Task, error) {
	m.createdParams = params
	if m.createdResult != nil {
		return m.createdResult, m.createErr
	}
	return &domain.Task{ID: "re-created-id", Content: params.Content}, m.createErr
}

func setupUndoApp(t *testing.T, mb *undoMockBackend) *undo.Log {
	t.Helper()
	cacheDir := t.TempDir()
	undoLog := undo.NewLog(cacheDir, 10)
	app = &App{
		Backend:  mb,
		Cache:    cache.New(cacheDir, 0),
		NowLabel: "now",
		UndoLog:  undoLog,
	}
	t.Cleanup(func() {
		app = nil
		jsonOutput = false
	})
	return undoLog
}

func runUndoCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	cmdArgs := append([]string{"undo"}, args...)
	rootCmd.SetArgs(cmdArgs)
	err := rootCmd.ExecuteContext(context.Background())
	return buf.String(), err
}

func TestUndoCmd_EmptyLog(t *testing.T) {
	mb := &undoMockBackend{}
	setupUndoApp(t, mb)

	_, err := runUndoCmd(t)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nothing to undo")
}

func TestUndoCmd_Done(t *testing.T) {
	mb := &undoMockBackend{}
	undoLog := setupUndoApp(t, mb)

	task := &domain.Task{ID: "task-abc", Content: "Buy groceries"}
	require.NoError(t, undoLog.Push(undo.Entry{
		Op:        undo.OpDone,
		TaskID:    "task-abc",
		Snapshot:  task,
		Timestamp: time.Now(),
	}))

	out, err := runUndoCmd(t)
	require.NoError(t, err)
	assert.Equal(t, "task-abc", mb.reopenedID)
	assert.Contains(t, out, "reopened")
	assert.Contains(t, out, "Buy groceries")

	// Entry should be consumed.
	e, err := undoLog.Peek()
	require.NoError(t, err)
	assert.Nil(t, e)
}

func TestUndoCmd_Delete(t *testing.T) {
	mb := &undoMockBackend{}
	undoLog := setupUndoApp(t, mb)

	task := &domain.Task{
		ID:       "task-del",
		Content:  "Fix login bug",
		Priority: domain.PriorityH,
		Labels:   []string{"backend"},
	}
	require.NoError(t, undoLog.Push(undo.Entry{
		Op:        undo.OpDelete,
		TaskID:    "task-del",
		Snapshot:  task,
		Timestamp: time.Now(),
	}))

	out, err := runUndoCmd(t)
	require.NoError(t, err)
	assert.Equal(t, "Fix login bug", mb.createdParams.Content)
	assert.Equal(t, domain.PriorityH, mb.createdParams.Priority)
	assert.Equal(t, []string{"backend"}, mb.createdParams.Labels)
	assert.Contains(t, out, "re-created")
	assert.Contains(t, out, "Fix login bug")
}

func TestUndoCmd_Modify(t *testing.T) {
	mb := &undoMockBackend{}
	undoLog := setupUndoApp(t, mb)

	task := &domain.Task{
		ID:       "task-mod",
		Content:  "Deploy service",
		Priority: domain.PriorityL,
	}
	require.NoError(t, undoLog.Push(undo.Entry{
		Op:        undo.OpModify,
		TaskID:    "task-mod",
		Snapshot:  task,
		Timestamp: time.Now(),
	}))

	out, err := runUndoCmd(t)
	require.NoError(t, err)
	assert.Equal(t, "task-mod", mb.updatedID)
	require.NotNil(t, mb.updatedParams.Content)
	assert.Equal(t, "Deploy service", *mb.updatedParams.Content)
	assert.Contains(t, out, "reverted")
	assert.Contains(t, out, "Deploy service")
}

func TestUndoCmd_Add(t *testing.T) {
	mb := &undoMockBackend{}
	undoLog := setupUndoApp(t, mb)

	require.NoError(t, undoLog.Push(undo.Entry{
		Op:        undo.OpAdd,
		CreatedID: "accidental-task-id",
		Timestamp: time.Now(),
	}))

	out, err := runUndoCmd(t)
	require.NoError(t, err)
	assert.Equal(t, "accidental-task-id", mb.deletedID)
	assert.Contains(t, out, "removed")
}

func TestUndoCmd_Start(t *testing.T) {
	mb := &undoMockBackend{}
	undoLog := setupUndoApp(t, mb)

	task := &domain.Task{
		ID:      "task-st",
		Content: "Write report",
		Labels:  []string{},
	}
	require.NoError(t, undoLog.Push(undo.Entry{
		Op:        undo.OpStart,
		TaskID:    "task-st",
		Snapshot:  task,
		Timestamp: time.Now(),
	}))

	out, err := runUndoCmd(t)
	require.NoError(t, err)
	assert.Equal(t, "task-st", mb.updatedID)
	assert.Contains(t, out, "reverted")
}

func TestUndoCmd_Stop(t *testing.T) {
	mb := &undoMockBackend{}
	undoLog := setupUndoApp(t, mb)

	task := &domain.Task{
		ID:      "task-sp",
		Content: "Write report",
		Labels:  []string{"now"},
	}
	require.NoError(t, undoLog.Push(undo.Entry{
		Op:        undo.OpStop,
		TaskID:    "task-sp",
		Snapshot:  task,
		Timestamp: time.Now(),
	}))

	out, err := runUndoCmd(t)
	require.NoError(t, err)
	assert.Equal(t, "task-sp", mb.updatedID)
	assert.Contains(t, out, "reverted")
}

func TestUndoCmd_BackendFailure_EntryPreserved(t *testing.T) {
	mb := &undoMockBackend{reopenErr: errors.New("network error")}
	undoLog := setupUndoApp(t, mb)

	require.NoError(t, undoLog.Push(undo.Entry{
		Op:        undo.OpDone,
		TaskID:    "task-fail",
		Snapshot:  &domain.Task{ID: "task-fail", Content: "Failing task"},
		Timestamp: time.Now(),
	}))

	_, err := runUndoCmd(t)
	require.Error(t, err)

	// Entry must still be in log.
	e, err2 := undoLog.Peek()
	require.NoError(t, err2)
	require.NotNil(t, e)
	assert.Equal(t, "task-fail", e.TaskID)
}

func TestUndoCmd_MissingSnapshot_Delete(t *testing.T) {
	mb := &undoMockBackend{}
	undoLog := setupUndoApp(t, mb)

	require.NoError(t, undoLog.Push(undo.Entry{
		Op:        undo.OpDelete,
		TaskID:    "task-no-snap",
		Timestamp: time.Now(),
		// Snapshot intentionally nil
	}))

	_, err := runUndoCmd(t)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing task snapshot")
}

func TestUndoCmd_JSONOutput_Done(t *testing.T) {
	mb := &undoMockBackend{}
	undoLog := setupUndoApp(t, mb)

	require.NoError(t, undoLog.Push(undo.Entry{
		Op:        undo.OpDone,
		TaskID:    "task-json",
		Snapshot:  &domain.Task{ID: "task-json", Content: "JSON task"},
		Timestamp: time.Now(),
	}))

	out, err := runUndoCmd(t, "--json")
	require.NoError(t, err)
	assert.Contains(t, out, `"status"`)
	assert.Contains(t, out, `"undone"`)
	assert.Contains(t, out, `"task_id"`)
	assert.Contains(t, out, `"task-json"`)
}

func TestUndoCmd_JSONOutput_Delete_HasNewID(t *testing.T) {
	mb := &undoMockBackend{
		createdResult: &domain.Task{ID: "brand-new-id", Content: "Fix login bug"},
	}
	undoLog := setupUndoApp(t, mb)

	require.NoError(t, undoLog.Push(undo.Entry{
		Op:        undo.OpDelete,
		TaskID:    "old-id",
		Snapshot:  &domain.Task{ID: "old-id", Content: "Fix login bug"},
		Timestamp: time.Now(),
	}))

	out, err := runUndoCmd(t, "--json")
	require.NoError(t, err)
	assert.True(t, strings.Contains(out, `"new_id"`), "JSON should include new_id for delete undo")
	assert.Contains(t, out, "brand-new-id")
}
