package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/abushady/tdo/internal/cache"
	"github.com/abushady/tdo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLCmd(t *testing.T) {
	mb := &mockBackend{
		tasks: []domain.Task{
			{ID: "task-123", Content: "Test task", URL: "https://app.todoist.com/app/task/task-123"},
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
	rootCmd.SetArgs([]string{"url", "--id", "task-123"})

	err := rootCmd.ExecuteContext(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "https://app.todoist.com/app/task/task-123\n", buf.String())
}

func TestURLCmd_MissingID(t *testing.T) {
	mb := &mockBackend{}
	cacheDir := t.TempDir()
	app = &App{
		Backend:  mb,
		Cache:    cache.New(cacheDir, 0),
		NowLabel: "now",
	}
	t.Cleanup(func() { app = nil })

	// Reset the hidden --id flag to avoid leaking from prior tests.
	_ = urlCmd.Flags().Set("id", "")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"url"})

	err := rootCmd.ExecuteContext(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "task ID is required")
}

func TestURLCmd_JSON(t *testing.T) {
	mb := &mockBackend{
		tasks: []domain.Task{
			{ID: "task-456", Content: "JSON task", URL: "https://app.todoist.com/app/task/task-456"},
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
	rootCmd.SetArgs([]string{"url", "--id", "task-456", "--json"})

	err := rootCmd.ExecuteContext(context.Background())
	require.NoError(t, err)

	var result map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "task-456", result["id"])
	assert.Equal(t, "https://app.todoist.com/app/task/task-456", result["url"])
}
