package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/abushady/tdo/internal/backend"
	"github.com/abushady/tdo/internal/cache"
	"github.com/abushady/tdo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBackend implements backend.Backend for testing.
type mockBackend struct {
	backend.Backend
	tasks         []domain.Task
	createdParams domain.CreateParams
}

func (m *mockBackend) ListTasks(_ context.Context, _ string) ([]domain.Task, error) {
	return m.tasks, nil
}

func (m *mockBackend) CreateTask(_ context.Context, params domain.CreateParams) (*domain.Task, error) {
	m.createdParams = params
	return &domain.Task{ID: "new-task-id", Content: params.Content}, nil
}

func TestAddCmd_ParentID(t *testing.T) {
	mb := &mockBackend{
		tasks: []domain.Task{
			{ID: "parent-todoist-id", Content: "Parent task"},
		},
	}

	cacheDir := t.TempDir()
	app = &App{
		Backend:  mb,
		Cache:    cache.New(cacheDir, 0),
		NowLabel: "now",
	}
	t.Cleanup(func() { app = nil })

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"add", "child task", "parent:parent-todoist-id"})

	err := rootCmd.ExecuteContext(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "parent-todoist-id", mb.createdParams.ParentID)
	assert.Equal(t, "child task", mb.createdParams.Content)
}
