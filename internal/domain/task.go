package domain

import "time"

// Priority represents task priority levels, matching TaskWarrior's H/M/L scheme.
type Priority int

const (
	PriorityNone Priority = iota
	PriorityL
	PriorityM
	PriorityH
)

// ParsePriority converts a string like "H", "M", "L" to a Priority value.
func ParsePriority(s string) Priority {
	switch s {
	case "H", "h":
		return PriorityH
	case "M", "m":
		return PriorityM
	case "L", "l":
		return PriorityL
	default:
		return PriorityNone
	}
}

// String returns the TaskWarrior-style priority label.
func (p Priority) String() string {
	switch p {
	case PriorityH:
		return "H"
	case PriorityM:
		return "M"
	case PriorityL:
		return "L"
	default:
		return ""
	}
}

// Task is the canonical domain model for a task, independent of any backend.
type Task struct {
	ID           string
	Content      string
	Description  string
	Priority     Priority
	Due          *time.Time
	Labels       []string
	Project      string
	ProjectID    string
	CreatedAt    time.Time
	CommentCount int
	IsCompleted  bool
	Recurrence   string
	ParentID     string
	URL          string
}

// HasLabel returns true if the task has the given label.
func (t Task) HasLabel(label string) bool {
	for _, l := range t.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// Project represents a Todoist project.
type Project struct {
	ID   string
	Name string
}

// Label represents a Todoist label.
type Label struct {
	ID   string
	Name string
}

// Comment represents a task comment/annotation.
type Comment struct {
	ID       string
	TaskID   string
	Content  string
	PostedAt time.Time
}

// CreateParams holds parameters for creating a new task.
type CreateParams struct {
	Content     string
	Description string
	Priority    Priority
	DueString   string
	Labels      []string
	ProjectID   string
	Recurrence  string
	ParentID    string
}

// UpdateParams holds parameters for updating a task. Nil fields are not updated.
type UpdateParams struct {
	Content      *string
	Description  *string
	Priority     *Priority
	DueString    *string
	Labels       []string
	AddLabels    []string
	RemoveLabels []string
	ProjectID    *string
}
