package todoist

import (
	"context"

	"github.com/abushady/tdo/internal/domain"
)

type todoistLabel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *Client) ListLabels(ctx context.Context) ([]domain.Label, error) {
	resp, err := c.doRequest(ctx, "GET", "/labels", nil)
	if err != nil {
		return nil, err
	}

	var page paginatedResponse[todoistLabel]
	if err := decodeResponse(resp, &page); err != nil {
		return nil, err
	}

	labels := make([]domain.Label, len(page.Results))
	for i, l := range page.Results {
		labels[i] = domain.Label{
			ID:   l.ID,
			Name: l.Name,
		}
	}
	return labels, nil
}
