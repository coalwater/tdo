package todoist

import (
	"context"

	"github.com/abushady/tdo/internal/domain"
)

type todoistProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) ListProjects(ctx context.Context) ([]domain.Project, error) {
	resp, err := c.doRequest(ctx, "GET", "/projects", nil)
	if err != nil {
		return nil, err
	}

	var raw []todoistProject
	if err := decodeResponse(resp, &raw); err != nil {
		return nil, err
	}

	projects := make([]domain.Project, len(raw))
	for i, p := range raw {
		projects[i] = domain.Project{
			ID:   p.ID,
			Name: p.Name,
		}
	}
	return projects, nil
}
