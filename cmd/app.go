package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/abushady/tdo/internal/backend"
	"github.com/abushady/tdo/internal/backend/todoist"
	"github.com/abushady/tdo/internal/cache"
	"github.com/abushady/tdo/internal/domain"
)

// App wires together the backend, cache, and configuration for all commands.
type App struct {
	Backend  backend.Backend
	Cache    *cache.Cache
	NowLabel string
}

// app holds the lazily-initialized App instance.
var app *App

// NewApp reads environment variables and creates a fully wired App.
func NewApp() (*App, error) {
	apiKey := os.Getenv("TODOIST_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("TODOIST_API_KEY environment variable is required")
	}

	ttl := 300 * time.Second
	if v := os.Getenv("TDO_CACHE_TTL"); v != "" {
		secs, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid TDO_CACHE_TTL: %w", err)
		}
		ttl = time.Duration(secs) * time.Second
	}

	nowLabel := os.Getenv("TDO_NOW_LABEL")
	if nowLabel == "" {
		nowLabel = "now"
	}

	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("determining home directory: %w", err)
		}
		cacheDir = filepath.Join(home, ".cache")
	}
	cacheDir = filepath.Join(cacheDir, "tdo")

	client := todoist.NewClient(apiKey)

	return &App{
		Backend:  client,
		Cache:    cache.New(cacheDir, ttl),
		NowLabel: nowLabel,
	}, nil
}

// initApp lazily initializes the global app instance.
func initApp() error {
	if app != nil {
		return nil
	}
	a, err := NewApp()
	if err != nil {
		return err
	}
	app = a
	return nil
}

// GetTasks returns tasks from cache or fetches from backend.
func (a *App) GetTasks(ctx context.Context) ([]domain.Task, error) {
	tasks, err := a.Cache.GetTasks()
	if err != nil {
		return nil, fmt.Errorf("reading task cache: %w", err)
	}
	if tasks != nil {
		return tasks, nil
	}

	tasks, err = a.Backend.ListTasks(ctx, "")
	if err != nil {
		return nil, err
	}

	_ = a.Cache.SetTasks(tasks)
	return tasks, nil
}

// GetProjects returns projects from cache or fetches from backend.
func (a *App) GetProjects(ctx context.Context) ([]domain.Project, error) {
	projects, err := a.Cache.GetProjects()
	if err != nil {
		return nil, fmt.Errorf("reading project cache: %w", err)
	}
	if projects != nil {
		return projects, nil
	}

	projects, err = a.Backend.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	_ = a.Cache.SetProjects(projects)
	return projects, nil
}

// GetLabels returns labels from cache or fetches from backend.
func (a *App) GetLabels(ctx context.Context) ([]domain.Label, error) {
	labels, err := a.Cache.GetLabels()
	if err != nil {
		return nil, fmt.Errorf("reading label cache: %w", err)
	}
	if labels != nil {
		return labels, nil
	}

	labels, err = a.Backend.ListLabels(ctx)
	if err != nil {
		return nil, err
	}

	_ = a.Cache.SetLabels(labels)
	return labels, nil
}

// ResolveProjectName finds a project by name (case-insensitive).
func (a *App) ResolveProjectName(ctx context.Context, name string) (string, error) {
	projects, err := a.GetProjects(ctx)
	if err != nil {
		return "", err
	}

	lower := strings.ToLower(name)
	for _, p := range projects {
		if strings.ToLower(p.Name) == lower {
			return p.ID, nil
		}
	}
	return "", fmt.Errorf("project %q not found", name)
}

// ResolveTaskID resolves a user-provided ID (positional number, Todoist ID, or
// fuzzy name match) to a task. It loads tasks and cached positions as needed.
func (a *App) ResolveTaskID(ctx context.Context, id string) (*domain.ResolveResult, error) {
	tasks, err := a.GetTasks(ctx)
	if err != nil {
		return nil, err
	}

	positions, _ := a.Cache.GetPositions()
	if positions == nil {
		positions = make(map[int]string)
	}

	resolver := &domain.Resolver{
		Tasks:     tasks,
		Positions: positions,
	}
	return resolver.Resolve(id)
}

// EnrichProjectNames populates the Project (name) field on tasks using cached
// project data.
func (a *App) EnrichProjectNames(ctx context.Context, tasks []domain.Task) {
	projects, err := a.GetProjects(ctx)
	if err != nil {
		return
	}

	projectMap := make(map[string]string, len(projects))
	for _, p := range projects {
		projectMap[p.ID] = p.Name
	}

	for i := range tasks {
		if name, ok := projectMap[tasks[i].ProjectID]; ok {
			tasks[i].Project = name
		}
	}
}
