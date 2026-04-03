package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Resolver resolves user input to a Todoist task ID.
// Resolution order: exact ID match, positional number, fuzzy name match.
type Resolver struct {
	Tasks     []Task
	Positions map[int]string // position # -> Todoist ID
}

// ResolveResult holds the resolved task ID and a pointer to the matched task.
type ResolveResult struct {
	TaskID string
	Task   *Task
}

// Resolve resolves the given input string to a task.
func (r *Resolver) Resolve(input string) (*ResolveResult, error) {
	// 1. Exact Todoist ID match
	for i := range r.Tasks {
		if r.Tasks[i].ID == input {
			return &ResolveResult{TaskID: r.Tasks[i].ID, Task: &r.Tasks[i]}, nil
		}
	}

	// 2. Positional number
	if n, err := strconv.Atoi(input); err == nil {
		if id, ok := r.Positions[n]; ok {
			for i := range r.Tasks {
				if r.Tasks[i].ID == id {
					return &ResolveResult{TaskID: id, Task: &r.Tasks[i]}, nil
				}
			}
			return &ResolveResult{TaskID: id, Task: nil}, nil
		}
		return nil, fmt.Errorf("position %d not found", n)
	}

	// 3. Fuzzy name match (case-insensitive substring)
	lower := strings.ToLower(input)
	var matches []Task
	for _, t := range r.Tasks {
		if strings.Contains(strings.ToLower(t.Content), lower) {
			matches = append(matches, t)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no task found matching %q", input)
	case 1:
		return &ResolveResult{TaskID: matches[0].ID, Task: &matches[0]}, nil
	default:
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = fmt.Sprintf("  - %s", m.Content)
		}
		return nil, fmt.Errorf("ambiguous match for %q, multiple tasks found:\n%s", input, strings.Join(names, "\n"))
	}
}
