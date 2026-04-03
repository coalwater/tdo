package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/abushady/tdo/internal/domain"
)

const (
	tasksFile     = "tasks.json"
	projectsFile  = "projects.json"
	labelsFile    = "labels.json"
	positionsFile = "positions.json"

	staticTTL = 24 * time.Hour
)

// Cache provides file-based TTL caching for task data, projects, labels,
// and position maps.
type Cache struct {
	dir string
	ttl time.Duration
}

// New creates a Cache that stores files under dir with the given task TTL.
// Projects and labels always use a fixed 24h TTL.
func New(dir string, ttl time.Duration) *Cache {
	return &Cache{dir: dir, ttl: ttl}
}

// envelope wraps cached data with a timestamp for TTL checks.
type envelope[T any] struct {
	UpdatedAt time.Time `json:"updated_at"`
	Data      T         `json:"data"`
}

// --- Tasks ---

func (c *Cache) GetTasks() ([]domain.Task, error) {
	return get[[]domain.Task](c.dir, tasksFile, c.ttl)
}

func (c *Cache) SetTasks(tasks []domain.Task) error {
	return set(c.dir, tasksFile, tasks)
}

func (c *Cache) InvalidateTasks() error {
	path := filepath.Join(c.dir, tasksFile)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// --- Projects (24h TTL) ---

func (c *Cache) GetProjects() ([]domain.Project, error) {
	return get[[]domain.Project](c.dir, projectsFile, staticTTL)
}

func (c *Cache) SetProjects(projects []domain.Project) error {
	return set(c.dir, projectsFile, projects)
}

// --- Labels (24h TTL) ---

func (c *Cache) GetLabels() ([]domain.Label, error) {
	return get[[]domain.Label](c.dir, labelsFile, staticTTL)
}

func (c *Cache) SetLabels(labels []domain.Label) error {
	return set(c.dir, labelsFile, labels)
}

// --- Positions ---

func (c *Cache) GetPositions() (map[int]string, error) {
	return get[map[int]string](c.dir, positionsFile, c.ttl)
}

func (c *Cache) SetPositions(positions map[int]string) error {
	return set(c.dir, positionsFile, positions)
}

// --- helpers ---

func get[T any](dir, file string, ttl time.Duration) (T, error) {
	var zero T
	path := filepath.Join(dir, file)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return zero, nil
		}
		return zero, err
	}

	var env envelope[T]
	if err := json.Unmarshal(data, &env); err != nil {
		return zero, nil // treat corrupt cache as missing
	}

	if time.Since(env.UpdatedAt) > ttl {
		return zero, nil
	}

	return env.Data, nil
}

func set[T any](dir, file string, data T) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	env := envelope[T]{
		UpdatedAt: time.Now(),
		Data:      data,
	}

	b, err := json.Marshal(env)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, file), b, 0o644)
}
