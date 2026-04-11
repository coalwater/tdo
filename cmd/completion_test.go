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

// mockBackendForCompletion implements backend.Backend for testing completion.
type mockBackendForCompletion struct {
	backend.Backend
	projects []domain.Project
	labels   []domain.Label
}

func (m *mockBackendForCompletion) ListProjects(_ context.Context) ([]domain.Project, error) {
	return m.projects, nil
}

func (m *mockBackendForCompletion) ListLabels(_ context.Context) ([]domain.Label, error) {
	return m.labels, nil
}

func TestCompletionCmd_ValidShells(t *testing.T) {
	tests := []struct {
		name  string
		shell string
	}{
		{name: "bash completion", shell: "bash"},
		{name: "zsh completion", shell: "zsh"},
		{name: "fish completion", shell: "fish"},
		{name: "powershell completion", shell: "powershell"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := &mockBackendForCompletion{}

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
			rootCmd.SetArgs([]string{"completion", tt.shell})

			err := rootCmd.ExecuteContext(context.Background())
			require.NoError(t, err, "completion command should succeed for valid shell: %s", tt.shell)
		})
	}
}

func TestCompletionCmd_InvalidShell(t *testing.T) {
	mb := &mockBackendForCompletion{}

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
	rootCmd.SetArgs([]string{"completion", "invalid-shell"})

	err := rootCmd.ExecuteContext(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
}

func TestCompletionCmd_MissingArgument(t *testing.T) {
	mb := &mockBackendForCompletion{}

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
	rootCmd.SetArgs([]string{"completion"})

	err := rootCmd.ExecuteContext(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestCompletionCmd_TooManyArguments(t *testing.T) {
	mb := &mockBackendForCompletion{}

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
	rootCmd.SetArgs([]string{"completion", "bash", "extra"})

	err := rootCmd.ExecuteContext(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestGetProjectNames(t *testing.T) {
	mb := &mockBackendForCompletion{
		projects: []domain.Project{
			{Name: "Project A"},
			{Name: "Project B"},
			{Name: "Work"},
		},
	}

	cacheDir := t.TempDir()
	app = &App{
		Backend:  mb,
		Cache:    cache.New(cacheDir, 0),
		NowLabel: "now",
	}
	t.Cleanup(func() { app = nil })

	names, err := getProjectNames(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"Project A", "Project B", "Work"}, names)
}

func TestGetLabelNames(t *testing.T) {
	mb := &mockBackendForCompletion{
		labels: []domain.Label{
			{Name: "urgent"},
			{Name: "bug"},
			{Name: "feature"},
		},
	}

	cacheDir := t.TempDir()
	app = &App{
		Backend:  mb,
		Cache:    cache.New(cacheDir, 0),
		NowLabel: "now",
	}
	t.Cleanup(func() { app = nil })

	names, err := getLabelNames(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []string{"urgent", "bug", "feature"}, names)
}
