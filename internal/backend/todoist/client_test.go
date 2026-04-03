package todoist

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/abushady/tdo/internal/backend"
	"github.com/abushady/tdo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time interface check.
var _ backend.Backend = (*Client)(nil)

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c := NewClient("test-api-key")
	c.baseURL = srv.URL
	return c
}

// wrapResults wraps data in a v1 paginated response.
func wrapResults(v any) map[string]any {
	return map[string]any{"results": v}
}

func TestAuthHeader(t *testing.T) {
	var gotAuth string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(wrapResults([]todoistTask{}))
	}))

	_, err := c.ListTasks(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, "Bearer test-api-key", gotAuth)
}

func TestListTasks(t *testing.T) {
	tests := []struct {
		name     string
		filter   string
		response []todoistTask
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "empty list",
			response: []todoistTask{},
			wantLen:  0,
		},
		{
			name:   "with tasks",
			filter: "today",
			response: []todoistTask{
				{
					ID:        "123",
					Content:   "Buy milk",
					Priority:  4,
					Labels:    []string{"errands"},
					ProjectID: "proj-1",
					AddedAt:   "2026-04-01T10:00:00Z",
					Due: &todoistDue{
						Date:   "2026-04-03",
						String: "today",
					},
				},
				{
					ID:       "456",
					Content:  "Read book",
					Priority: 1,
				},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.RequestURI()
				json.NewEncoder(w).Encode(wrapResults(tt.response))
			}))

			tasks, err := c.ListTasks(context.Background(), tt.filter)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, tasks, tt.wantLen)

			if tt.filter != "" {
				assert.Contains(t, gotPath, "filter="+tt.filter)
			}
		})
	}
}

func TestListTasksPriorityMapping(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(wrapResults([]todoistTask{
			{ID: "1", Content: "High", Priority: 4},
			{ID: "2", Content: "Medium", Priority: 3},
			{ID: "3", Content: "Low", Priority: 2},
			{ID: "4", Content: "None", Priority: 1},
		}))
	}))

	tasks, err := c.ListTasks(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, tasks, 4)

	assert.Equal(t, domain.PriorityH, tasks[0].Priority)
	assert.Equal(t, domain.PriorityM, tasks[1].Priority)
	assert.Equal(t, domain.PriorityL, tasks[2].Priority)
	assert.Equal(t, domain.PriorityNone, tasks[3].Priority)
}

func TestListTasksDueAndRecurrence(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(wrapResults([]todoistTask{
			{
				ID:       "1",
				Content:  "Datetime due",
				Priority: 1,
				Due: &todoistDue{
					Datetime: "2026-04-03T14:00:00Z",
					Date:     "2026-04-03",
					String:   "Apr 3 2pm",
				},
			},
			{
				ID:       "2",
				Content:  "Date-only due",
				Priority: 1,
				Due: &todoistDue{
					Date:   "2026-04-05",
					String: "Apr 5",
				},
			},
			{
				ID:       "3",
				Content:  "Recurring",
				Priority: 1,
				Due: &todoistDue{
					Date:        "2026-04-03",
					String:      "every day",
					IsRecurring: true,
				},
			},
			{
				ID:       "4",
				Content:  "No due",
				Priority: 1,
			},
		}))
	}))

	tasks, err := c.ListTasks(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, tasks, 4)

	require.NotNil(t, tasks[0].Due)
	assert.Equal(t, 14, tasks[0].Due.Hour())

	require.NotNil(t, tasks[1].Due)
	assert.Equal(t, 5, tasks[1].Due.Day())

	assert.Equal(t, "every day", tasks[2].Recurrence)

	assert.Nil(t, tasks[3].Due)
}

func TestCreateTask(t *testing.T) {
	var gotBody createTaskRequest
	var gotMethod string

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(todoistTask{
			ID:        "new-1",
			Content:   "Buy groceries",
			Priority:  3,
			Labels:    []string{"errands"},
			ProjectID: "proj-1",
			AddedAt:   "2026-04-03T12:00:00Z",
		})
	}))

	task, err := c.CreateTask(context.Background(), domain.CreateParams{
		Content:   "Buy groceries",
		Priority:  domain.PriorityM,
		Labels:    []string{"errands"},
		ProjectID: "proj-1",
		DueString: "tomorrow",
	})

	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "Buy groceries", gotBody.Content)
	assert.Equal(t, 3, gotBody.Priority)
	assert.Equal(t, "tomorrow", gotBody.DueString)
	assert.Equal(t, []string{"errands"}, gotBody.Labels)
	assert.Equal(t, "proj-1", gotBody.ProjectID)

	assert.Equal(t, "new-1", task.ID)
	assert.Equal(t, "Buy groceries", task.Content)
	assert.Equal(t, domain.PriorityM, task.Priority)
}

func TestCompleteTask(t *testing.T) {
	var gotPath string
	var gotMethod string

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.CompleteTask(context.Background(), "task-42")
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/tasks/task-42/close", gotPath)
}

func TestReopenTask(t *testing.T) {
	var gotPath string

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.ReopenTask(context.Background(), "task-42")
	require.NoError(t, err)
	assert.Equal(t, "/tasks/task-42/reopen", gotPath)
}

func TestDeleteTask(t *testing.T) {
	var gotMethod string

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeleteTask(context.Background(), "task-42")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
}

func TestMoveTask(t *testing.T) {
	var gotBody map[string]any

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.MoveTask(context.Background(), "task-1", "proj-new")
	require.NoError(t, err)
	assert.Equal(t, "proj-new", gotBody["project_id"])
}

func TestRateLimitRetry(t *testing.T) {
	var attempts atomic.Int32

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(wrapResults([]todoistTask{
			{ID: "1", Content: "Survived", Priority: 1},
		}))
	}))

	tasks, err := c.ListTasks(context.Background(), "")
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestRateLimitExhausted(t *testing.T) {
	var attempts atomic.Int32

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))

	_, err := c.ListTasks(context.Background(), "")
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestNotFoundError(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("task not found"))
	}))

	_, err := c.GetTask(context.Background(), "nonexistent")
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Contains(t, apiErr.Message, "task not found")
}

func TestListProjects(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/projects", r.URL.Path)
		json.NewEncoder(w).Encode(wrapResults([]todoistProject{
			{ID: "p1", Name: "Inbox"},
			{ID: "p2", Name: "Work"},
		}))
	}))

	projects, err := c.ListProjects(context.Background())
	require.NoError(t, err)
	require.Len(t, projects, 2)
	assert.Equal(t, "p1", projects[0].ID)
	assert.Equal(t, "Inbox", projects[0].Name)
	assert.Equal(t, "p2", projects[1].ID)
	assert.Equal(t, "Work", projects[1].Name)
}

func TestListLabels(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/labels", r.URL.Path)
		json.NewEncoder(w).Encode(wrapResults([]todoistLabel{
			{ID: "l1", Name: "urgent"},
			{ID: "l2", Name: "home"},
		}))
	}))

	labels, err := c.ListLabels(context.Background())
	require.NoError(t, err)
	require.Len(t, labels, 2)
	assert.Equal(t, "l1", labels[0].ID)
	assert.Equal(t, "urgent", labels[0].Name)
}

func TestAddComment(t *testing.T) {
	var gotBody map[string]string

	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/comments", r.URL.Path)
		json.NewDecoder(r.Body).Decode(&gotBody)
		json.NewEncoder(w).Encode(todoistComment{
			ID:       "c1",
			ItemID:   "task-1",
			Content:  "A note",
			PostedAt: "2026-04-03T10:00:00Z",
		})
	}))

	comment, err := c.AddComment(context.Background(), "task-1", "A note")
	require.NoError(t, err)
	assert.Equal(t, "task-1", gotBody["task_id"])
	assert.Equal(t, "A note", gotBody["content"])
	assert.Equal(t, "c1", comment.ID)
	assert.Equal(t, "A note", comment.Content)
}

func TestListComments(t *testing.T) {
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "task-1", r.URL.Query().Get("task_id"))
		json.NewEncoder(w).Encode(wrapResults([]todoistComment{
			{ID: "c1", ItemID: "task-1", Content: "First", PostedAt: "2026-04-01T08:00:00Z"},
			{ID: "c2", ItemID: "task-1", Content: "Second", PostedAt: "2026-04-02T09:00:00Z"},
		}))
	}))

	comments, err := c.ListComments(context.Background(), "task-1")
	require.NoError(t, err)
	require.Len(t, comments, 2)
	assert.Equal(t, "First", comments[0].Content)
	assert.Equal(t, "Second", comments[1].Content)
}
