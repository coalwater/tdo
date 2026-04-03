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

type todoistComment struct {
	ID       string `json:"id"`
	TaskID   string `json:"task_id"`
	Content  string `json:"content"`
	PostedAt string `json:"posted_at"`
}

func (c *Client) AddComment(ctx context.Context, taskID string, text string) (*domain.Comment, error) {
	reqBody, err := json.Marshal(map[string]string{
		"task_id": taskID,
		"content": text,
	})
	if err != nil {
		return nil, fmt.Errorf("todoist: encoding comment request: %w", err)
	}

	resp, err := c.doRequest(ctx, "POST", "/comments", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	var raw todoistComment
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, err
	}

	return toDomainComment(raw), nil
}

func (c *Client) ListComments(ctx context.Context, taskID string) ([]domain.Comment, error) {
	path := "/comments?" + url.Values{"task_id": {taskID}}.Encode()

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var raw []todoistComment
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, err
	}

	comments := make([]domain.Comment, len(raw))
	for i, c := range raw {
		comments[i] = *toDomainComment(c)
	}
	return comments, nil
}

func toDomainComment(c todoistComment) *domain.Comment {
	comment := &domain.Comment{
		ID:      c.ID,
		TaskID:  c.TaskID,
		Content: c.Content,
	}
	if c.PostedAt != "" {
		if t, err := time.Parse(time.RFC3339, c.PostedAt); err == nil {
			comment.PostedAt = t
		}
	}
	return comment
}
