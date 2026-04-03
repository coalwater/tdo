package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/abushady/tdo/internal/domain"
)

// taskJSON is the JSON-serializable representation of a task.
type taskJSON struct {
	ID           string   `json:"id"`
	Content      string   `json:"content"`
	Description  string   `json:"description,omitempty"`
	Priority     string   `json:"priority,omitempty"`
	Due          string   `json:"due,omitempty"`
	Labels       []string `json:"labels"`
	Project      string   `json:"project,omitempty"`
	ProjectID    string   `json:"project_id,omitempty"`
	Recurrence   string   `json:"recurrence,omitempty"`
	CommentCount int      `json:"comment_count"`
	Urgency      float64  `json:"urgency"`
	IsCompleted  bool     `json:"is_completed"`
	CreatedAt    string   `json:"created_at"`
	URL          string   `json:"url,omitempty"`
}

func toTaskJSON(t domain.Task, nowLabel string, now time.Time) taskJSON {
	tj := taskJSON{
		ID:           t.ID,
		Content:      t.Content,
		Description:  t.Description,
		Priority:     t.Priority.String(),
		Labels:       t.Labels,
		Project:      t.Project,
		ProjectID:    t.ProjectID,
		Recurrence:   t.Recurrence,
		CommentCount: t.CommentCount,
		Urgency:      domain.CalculateUrgency(t, nowLabel, now),
		IsCompleted:  t.IsCompleted,
		URL:          t.URL,
	}
	if tj.Labels == nil {
		tj.Labels = []string{}
	}
	if t.Due != nil {
		tj.Due = t.Due.Format("2006-01-02")
	}
	if !t.CreatedAt.IsZero() {
		tj.CreatedAt = t.CreatedAt.Format(time.RFC3339)
	}
	return tj
}

type commentJSON struct {
	ID       string `json:"id"`
	TaskID   string `json:"task_id"`
	Content  string `json:"content"`
	PostedAt string `json:"posted_at"`
}

func toCommentJSON(c domain.Comment) commentJSON {
	cj := commentJSON{
		ID:      c.ID,
		TaskID:  c.TaskID,
		Content: c.Content,
	}
	if !c.PostedAt.IsZero() {
		cj.PostedAt = c.PostedAt.Format(time.RFC3339)
	}
	return cj
}

type projectJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type labelJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// writeJSON marshals v as indented JSON and writes to w.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}
