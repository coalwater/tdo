package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Filter holds parsed filter criteria. Multiple criteria are AND-ed during matching.
type Filter struct {
	Project         string
	Priority        *Priority
	HasLabels       []string
	NotLabels       []string
	DueBefore       string
	DueAfter        string
	ScheduledBefore string
	ScheduledAfter  string
	Limit           int
}

// filterAttrs is the known attribute list for ParseFilter.
var filterAttrs = []string{"project", "priority", "due.before", "due.after", "scheduled.before", "scheduled.after", "limit"}

// ParseFilter parses TaskWarrior-style filter arguments into a Filter.
// now is used to resolve date expressions in filter values.
func ParseFilter(args []string, now time.Time) (Filter, error) {
	var f Filter
	for _, arg := range args {
		if strings.HasPrefix(arg, "+") {
			f.HasLabels = append(f.HasLabels, strings.TrimPrefix(arg, "+"))
			continue
		}
		if strings.HasPrefix(arg, "-") {
			f.NotLabels = append(f.NotLabels, strings.TrimPrefix(arg, "-"))
			continue
		}

		attr, value, err := matchAttr(arg, filterAttrs)
		if err != nil {
			return Filter{}, err
		}

		switch attr {
		case "project":
			f.Project = value
		case "priority":
			p := ParsePriority(value)
			f.Priority = &p
		case "due.before":
			resolved, err := resolveFilterDate(value, now)
			if err != nil {
				return Filter{}, fmt.Errorf("invalid due.before value %q: %w", value, err)
			}
			f.DueBefore = resolved
		case "due.after":
			resolved, err := resolveFilterDate(value, now)
			if err != nil {
				return Filter{}, fmt.Errorf("invalid due.after value %q: %w", value, err)
			}
			f.DueAfter = resolved
		case "scheduled.before":
			resolved, err := resolveFilterDate(value, now)
			if err != nil {
				return Filter{}, fmt.Errorf("invalid scheduled.before value %q: %w", value, err)
			}
			f.ScheduledBefore = resolved
		case "scheduled.after":
			resolved, err := resolveFilterDate(value, now)
			if err != nil {
				return Filter{}, fmt.Errorf("invalid scheduled.after value %q: %w", value, err)
			}
			f.ScheduledAfter = resolved
		case "limit":
			n, err := strconv.Atoi(value)
			if err != nil {
				return Filter{}, fmt.Errorf("invalid limit value %q: must be an integer", value)
			}
			if n < 0 {
				return Filter{}, fmt.Errorf("invalid limit value %q: must be non-negative", value)
			}
			f.Limit = n
		}
	}
	return f, nil
}

// resolveFilterDate resolves a filter date value. If it's already YYYY-MM-DD,
// return as-is. Otherwise try ParseDateExpr and format the result.
func resolveFilterDate(value string, now time.Time) (string, error) {
	if _, err := time.Parse("2006-01-02", value); err == nil {
		return value, nil
	}
	t, err := ParseDateExpr(value, now)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}

// Match returns true if the task satisfies all filter criteria.
func (f Filter) Match(task Task) bool {
	if f.Project != "" && !strings.EqualFold(task.Project, f.Project) {
		return false
	}

	if f.Priority != nil && task.Priority != *f.Priority {
		return false
	}

	for _, label := range f.HasLabels {
		if !task.HasLabel(label) {
			return false
		}
	}

	for _, label := range f.NotLabels {
		if task.HasLabel(label) {
			return false
		}
	}

	if f.DueBefore != "" {
		if task.Due == nil {
			return false
		}
		if task.Due.Format("2006-01-02") >= f.DueBefore {
			return false
		}
	}

	if f.DueAfter != "" {
		if task.Due == nil {
			return false
		}
		if task.Due.Format("2006-01-02") <= f.DueAfter {
			return false
		}
	}

	if f.ScheduledBefore != "" {
		if task.Scheduled == nil {
			return false
		}
		if task.Scheduled.Format("2006-01-02") >= f.ScheduledBefore {
			return false
		}
	}

	if f.ScheduledAfter != "" {
		if task.Scheduled == nil {
			return false
		}
		if task.Scheduled.Format("2006-01-02") <= f.ScheduledAfter {
			return false
		}
	}

	return true
}
