package todoist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/abushady/tdo/internal/domain"
)

// todoistTask represents the Todoist API task JSON shape.
type todoistTask struct {
	ID           string       `json:"id"`
	Content      string       `json:"content"`
	Description  string       `json:"description"`
	Priority     int          `json:"priority"`
	Due          *todoistDue  `json:"due"`
	Labels       []string     `json:"labels"`
	ProjectID    string       `json:"project_id"`
	CreatedAt    string       `json:"created_at"`
	CommentCount int          `json:"comment_count"`
	IsCompleted  bool         `json:"is_completed"`
	ParentID     string       `json:"parent_id"`
	URL          string       `json:"url"`
}

type todoistDue struct {
	Date        string `json:"date"`
	Datetime    string `json:"datetime"`
	String      string `json:"string"`
	IsRecurring bool   `json:"is_recurring"`
}

// Priority mapping: domain <-> Todoist API.
// Todoist inverts: API 4 = highest, API 1 = none.
func priorityToAPI(p domain.Priority) int {
	switch p {
	case domain.PriorityH:
		return 4
	case domain.PriorityM:
		return 3
	case domain.PriorityL:
		return 2
	default:
		return 1
	}
}

func priorityFromAPI(p int) domain.Priority {
	switch p {
	case 4:
		return domain.PriorityH
	case 3:
		return domain.PriorityM
	case 2:
		return domain.PriorityL
	default:
		return domain.PriorityNone
	}
}

func toDomainTask(t todoistTask) domain.Task {
	task := domain.Task{
		ID:           t.ID,
		Content:      t.Content,
		Description:  t.Description,
		Priority:     priorityFromAPI(t.Priority),
		Labels:       t.Labels,
		ProjectID:    t.ProjectID,
		CommentCount: t.CommentCount,
		IsCompleted:  t.IsCompleted,
		ParentID:     t.ParentID,
		URL:          t.URL,
	}

	if task.Labels == nil {
		task.Labels = []string{}
	}

	if t.CreatedAt != "" {
		if ct, err := time.Parse(time.RFC3339, t.CreatedAt); err == nil {
			task.CreatedAt = ct
		}
	}

	if t.Due != nil {
		if t.Due.IsRecurring {
			task.Recurrence = t.Due.String
		}
		// Prefer datetime over date.
		if t.Due.Datetime != "" {
			if dt, err := time.Parse(time.RFC3339, t.Due.Datetime); err == nil {
				task.Due = &dt
			}
		} else if t.Due.Date != "" {
			if dt, err := time.Parse("2006-01-02", t.Due.Date); err == nil {
				task.Due = &dt
			}
		}
	}

	return task
}

func (c *Client) ListTasks(ctx context.Context, filter string) ([]domain.Task, error) {
	path := "/tasks"
	if filter != "" {
		path += "?" + url.Values{"filter": {filter}}.Encode()
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var raw []todoistTask
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, err
	}

	tasks := make([]domain.Task, len(raw))
	for i, t := range raw {
		tasks[i] = toDomainTask(t)
	}
	return tasks, nil
}

func (c *Client) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/tasks/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var raw todoistTask
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, err
	}

	task := toDomainTask(raw)
	return &task, nil
}

type createTaskRequest struct {
	Content     string   `json:"content"`
	Description string   `json:"description,omitempty"`
	Priority    int      `json:"priority"`
	DueString   string   `json:"due_string,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
}

func (c *Client) CreateTask(ctx context.Context, params domain.CreateParams) (*domain.Task, error) {
	reqBody := createTaskRequest{
		Content:     params.Content,
		Description: params.Description,
		Priority:    priorityToAPI(params.Priority),
		DueString:   params.DueString,
		Labels:      params.Labels,
		ProjectID:   params.ProjectID,
		ParentID:    params.ParentID,
	}
	// Use recurrence as due_string if set and due_string is empty.
	if reqBody.DueString == "" && params.Recurrence != "" {
		reqBody.DueString = params.Recurrence
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("todoist: encoding create request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/tasks", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var raw todoistTask
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, err
	}

	task := toDomainTask(raw)
	return &task, nil
}

type updateTaskRequest struct {
	Content     *string  `json:"content,omitempty"`
	Description *string  `json:"description,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	DueString   *string  `json:"due_string,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	ProjectID   *string  `json:"project_id,omitempty"`
}

func (c *Client) UpdateTask(ctx context.Context, id string, params domain.UpdateParams) error {
	req := updateTaskRequest{}
	req.Content = params.Content
	req.Description = params.Description
	if params.Priority != nil {
		p := priorityToAPI(*params.Priority)
		req.Priority = &p
	}
	req.DueString = params.DueString
	if params.Labels != nil {
		req.Labels = params.Labels
	}
	req.ProjectID = params.ProjectID

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("todoist: encoding update request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/tasks/%s", id), bytes.NewReader(body))
	if err != nil {
		return err
	}

	return drainAndClose(resp)
}

func (c *Client) CompleteTask(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/tasks/%s/close", id), nil)
	if err != nil {
		return err
	}
	return drainAndClose(resp)
}

func (c *Client) ReopenTask(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "POST", fmt.Sprintf("/tasks/%s/reopen", id), nil)
	if err != nil {
		return err
	}
	return drainAndClose(resp)
}

func (c *Client) DeleteTask(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/tasks/%s", id), nil)
	if err != nil {
		return err
	}
	return drainAndClose(resp)
}

func (c *Client) MoveTask(ctx context.Context, id string, projectID string) error {
	return c.UpdateTask(ctx, id, domain.UpdateParams{
		ProjectID: &projectID,
	})
}
