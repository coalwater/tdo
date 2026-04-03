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

// todoistTask represents the Todoist API v1 task JSON shape.
type todoistTask struct {
	ID          string           `json:"id"`
	Content     string           `json:"content"`
	Description string           `json:"description"`
	Priority    int              `json:"priority"`
	Due         *todoistDue      `json:"due"`
	Deadline    *todoistDeadline `json:"deadline"`
	Labels      []string         `json:"labels"`
	ProjectID   string           `json:"project_id"`
	AddedAt     string           `json:"added_at"`
	NoteCount   int              `json:"note_count"`
	Checked     bool             `json:"checked"`
	ParentID    string           `json:"parent_id"`
	URL         string           `json:"url"`
}

type todoistDeadline struct {
	Date string `json:"date"`
	Lang string `json:"lang"`
}

// paginatedResponse wraps list responses in the v1 API.
type paginatedResponse[T any] struct {
	Results    []T    `json:"results"`
	NextCursor string `json:"next_cursor,omitempty"`
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
		CommentCount: t.NoteCount,
		IsCompleted:  t.Checked,
		ParentID:     t.ParentID,
		URL:          t.URL,
	}

	if task.Labels == nil {
		task.Labels = []string{}
	}

	if t.AddedAt != "" {
		if ct, err := time.Parse(time.RFC3339, t.AddedAt); err == nil {
			task.CreatedAt = ct
		}
	}

	// Map Todoist due → domain Scheduled (when to work on it)
	if t.Due != nil {
		if t.Due.IsRecurring {
			task.Recurrence = t.Due.String
		}
		if t.Due.Datetime != "" {
			if dt, err := time.Parse(time.RFC3339, t.Due.Datetime); err == nil {
				task.Scheduled = &dt
			}
		} else if t.Due.Date != "" {
			if dt, err := time.Parse("2006-01-02", t.Due.Date); err == nil {
				task.Scheduled = &dt
			}
		}
	}

	// Map Todoist deadline → domain Due (hard deadline)
	if t.Deadline != nil && t.Deadline.Date != "" {
		if dt, err := time.Parse("2006-01-02", t.Deadline.Date); err == nil {
			task.Due = &dt
		}
	}

	return task
}

func (c *Client) ListTasks(ctx context.Context, filter string) ([]domain.Task, error) {
	var allTasks []domain.Task
	cursor := ""

	for {
		path := "/tasks"
		params := url.Values{}
		if filter != "" {
			params.Set("filter", filter)
		}
		if cursor != "" {
			params.Set("cursor", cursor)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}

		resp, err := c.doRequest(ctx, "GET", path, nil)
		if err != nil {
			return nil, err
		}

		var page paginatedResponse[todoistTask]
		if err := decodeResponse(resp, &page); err != nil {
			return nil, err
		}

		for _, t := range page.Results {
			allTasks = append(allTasks, toDomainTask(t))
		}

		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}

	return allTasks, nil
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
	Content      string   `json:"content"`
	Description  string   `json:"description,omitempty"`
	Priority     int      `json:"priority"`
	DueString    string   `json:"due_string,omitempty"`    // for scheduled (Todoist NLP)
	DeadlineDate *string  `json:"deadline_date,omitempty"` // for due/deadline (YYYY-MM-DD)
	Labels       []string `json:"labels,omitempty"`
	ProjectID    string   `json:"project_id,omitempty"`
	ParentID     string   `json:"parent_id,omitempty"`
}

func (c *Client) CreateTask(ctx context.Context, params domain.CreateParams) (*domain.Task, error) {
	reqBody := createTaskRequest{
		Content:     params.Content,
		Description: params.Description,
		Priority:    priorityToAPI(params.Priority),
		DueString:   params.ScheduledString,
		Labels:      params.Labels,
		ProjectID:   params.ProjectID,
		ParentID:    params.ParentID,
	}
	if params.DueDate != "" {
		reqBody.DeadlineDate = &params.DueDate
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
	Content      *string  `json:"content,omitempty"`
	Description  *string  `json:"description,omitempty"`
	Priority     *int     `json:"priority,omitempty"`
	DueString    *string  `json:"due_string,omitempty"`    // for scheduled
	DeadlineDate *string  `json:"deadline_date,omitempty"` // for due/deadline
	Labels       []string `json:"labels,omitempty"`
	ProjectID    *string  `json:"project_id,omitempty"`
}

func (c *Client) UpdateTask(ctx context.Context, id string, params domain.UpdateParams) error {
	req := updateTaskRequest{}
	req.Content = params.Content
	req.Description = params.Description
	if params.Priority != nil {
		p := priorityToAPI(*params.Priority)
		req.Priority = &p
	}
	req.DueString = params.ScheduledString
	req.DeadlineDate = params.DueDate
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
