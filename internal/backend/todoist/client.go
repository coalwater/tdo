package todoist

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.todoist.com/api/v1"

const (
	maxRetries  = 3
	baseBackoff = 1 * time.Second
)

// APIError represents an error response from the Todoist API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("todoist: HTTP %d: %s", e.StatusCode, e.Message)
}

// Client talks to the Todoist REST API v2.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Todoist API client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    defaultBaseURL,
	}
}

// doRequest executes an HTTP request with auth, retry on 429, and error handling.
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	var attempt int
	for {
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, fmt.Errorf("todoist: building request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("todoist: executing request: %w", err)
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		resp.Body.Close()
		attempt++
		if attempt >= maxRetries {
			return nil, &APIError{StatusCode: http.StatusTooManyRequests, Message: "rate limited after retries"}
		}

		backoff := time.Duration(math.Pow(2, float64(attempt-1))) * baseBackoff
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}

		// Reset body for retry if it's seekable.
		if seeker, ok := body.(io.Seeker); ok {
			if _, err := seeker.Seek(0, io.SeekStart); err != nil {
				return nil, fmt.Errorf("todoist: resetting request body: %w", err)
			}
		}
	}
}

// decodeResponse reads and decodes a JSON response body into v.
// It closes the response body and returns an APIError for non-2xx status codes.
func decodeResponse(resp *http.Response, v any) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		msg := string(bodyBytes)
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	if v == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// drainAndClose reads and discards the response body then closes it.
// Used for responses where we only care about the status code.
func drainAndClose(resp *http.Response) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		msg := string(bodyBytes)
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return &APIError{StatusCode: resp.StatusCode, Message: msg}
	}

	return nil
}
