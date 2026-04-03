package backend

import (
	"context"

	"github.com/abushady/tdo/internal/domain"
)

// Backend defines the interface for task storage backends.
// Todoist is the primary implementation; tests use a mock.
type Backend interface {
	ListTasks(ctx context.Context, filter string) ([]domain.Task, error)
	GetTask(ctx context.Context, id string) (*domain.Task, error)
	CreateTask(ctx context.Context, params domain.CreateParams) (*domain.Task, error)
	UpdateTask(ctx context.Context, id string, params domain.UpdateParams) error
	CompleteTask(ctx context.Context, id string) error
	ReopenTask(ctx context.Context, id string) error
	DeleteTask(ctx context.Context, id string) error
	MoveTask(ctx context.Context, id string, projectID string) error

	AddComment(ctx context.Context, taskID string, text string) (*domain.Comment, error)
	ListComments(ctx context.Context, taskID string) ([]domain.Comment, error)

	ListProjects(ctx context.Context) ([]domain.Project, error)
	ListLabels(ctx context.Context) ([]domain.Label, error)
}
