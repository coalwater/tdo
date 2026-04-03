package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/abushady/tdo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tmpDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "tdo-test")
}

func TestTasks(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(c *Cache)
		ttl     time.Duration
		wantNil bool
	}{
		{
			name: "set then get within TTL returns data",
			setup: func(c *Cache) {
				require.NoError(t, c.SetTasks([]domain.Task{
					{ID: "1", Content: "buy milk"},
					{ID: "2", Content: "write code", Priority: domain.PriorityH},
				}))
			},
			ttl:     time.Hour,
			wantNil: false,
		},
		{
			name: "get with no file returns nil nil",
			setup: func(_ *Cache) {
				// no setup — file doesn't exist
			},
			ttl:     time.Hour,
			wantNil: true,
		},
		{
			name: "expired TTL returns nil nil",
			setup: func(c *Cache) {
				require.NoError(t, c.SetTasks([]domain.Task{{ID: "1"}}))
				// backdate the file
				backdateFile(t, filepath.Join(c.dir, tasksFile), 2*time.Hour)
			},
			ttl:     time.Hour,
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := New(tmpDir(t), tc.ttl)
			tc.setup(c)

			got, err := c.GetTasks()
			assert.NoError(t, err)
			if tc.wantNil {
				assert.Nil(t, got)
			} else {
				assert.Len(t, got, 2)
				assert.Equal(t, "buy milk", got[0].Content)
				assert.Equal(t, domain.PriorityH, got[1].Priority)
			}
		})
	}
}

func TestInvalidateTasks(t *testing.T) {
	c := New(tmpDir(t), time.Hour)
	require.NoError(t, c.SetTasks([]domain.Task{{ID: "1"}}))

	got, err := c.GetTasks()
	require.NoError(t, err)
	require.NotNil(t, got)

	require.NoError(t, c.InvalidateTasks())

	got, err = c.GetTasks()
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestInvalidateTasksNoFile(t *testing.T) {
	c := New(tmpDir(t), time.Hour)
	// should not error when file doesn't exist
	assert.NoError(t, c.InvalidateTasks())
}

func TestPositions(t *testing.T) {
	c := New(tmpDir(t), time.Hour)
	positions := map[int]string{
		1: "abc123",
		2: "def456",
		3: "ghi789",
	}

	require.NoError(t, c.SetPositions(positions))

	got, err := c.GetPositions()
	assert.NoError(t, err)
	assert.Equal(t, positions, got)
}

func TestProjectsUse24hTTL(t *testing.T) {
	// Configure cache with a very short task TTL.
	c := New(tmpDir(t), 1*time.Millisecond)
	require.NoError(t, c.SetProjects([]domain.Project{{ID: "p1", Name: "Work"}}))

	// Backdate file by 2 hours — still within 24h TTL for projects.
	backdateFile(t, filepath.Join(c.dir, projectsFile), 2*time.Hour)

	got, err := c.GetProjects()
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "Work", got[0].Name)

	// Backdate past 24h — should expire.
	backdateFile(t, filepath.Join(c.dir, projectsFile), 25*time.Hour)

	got, err = c.GetProjects()
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestLabelsUse24hTTL(t *testing.T) {
	c := New(tmpDir(t), 1*time.Millisecond)
	require.NoError(t, c.SetLabels([]domain.Label{{ID: "l1", Name: "urgent"}}))

	// Backdate by 2 hours — still valid.
	backdateFile(t, filepath.Join(c.dir, labelsFile), 2*time.Hour)

	got, err := c.GetLabels()
	assert.NoError(t, err)
	assert.Len(t, got, 1)

	// Backdate past 24h.
	backdateFile(t, filepath.Join(c.dir, labelsFile), 25*time.Hour)

	got, err = c.GetLabels()
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestCacheCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "deeply", "nested", "cache")
	c := New(dir, time.Hour)

	require.NoError(t, c.SetTasks([]domain.Task{{ID: "1"}}))

	_, err := os.Stat(dir)
	assert.NoError(t, err, "cache directory should be created on first write")

	got, err := c.GetTasks()
	assert.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestConcurrentReads(t *testing.T) {
	c := New(tmpDir(t), time.Hour)
	require.NoError(t, c.SetTasks([]domain.Task{{ID: "1", Content: "task"}}))

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := c.GetTasks()
			assert.NoError(t, err)
			assert.Len(t, got, 1)
		}()
	}
	wg.Wait()
}

// backdateFile rewrites the envelope's UpdatedAt to (now - age).
func backdateFile(t *testing.T, path string, age time.Duration) {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))

	ts := time.Now().Add(-age)
	b, err := json.Marshal(ts)
	require.NoError(t, err)
	raw["updated_at"] = b

	out, err := json.Marshal(raw)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, out, 0o600))
}
