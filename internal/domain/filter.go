package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// Filter holds parsed filter criteria. Multiple criteria are AND-ed during matching.
type Filter struct {
	Project   string
	Priority  *Priority
	HasLabels []string
	NotLabels []string
	DueBefore string
	DueAfter  string
	Limit     int
}

// filterAttrs is the known attribute list for ParseFilter.
var filterAttrs = []string{"project", "priority", "due.before", "due.after", "limit"}

// ParseFilter parses TaskWarrior-style filter arguments into a Filter.
func ParseFilter(args []string) (Filter, error) {
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
			f.DueBefore = value
		case "due.after":
			f.DueAfter = value
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
		taskDue := task.Due.Format("2006-01-02")
		if taskDue >= f.DueBefore {
			return false
		}
	}

	if f.DueAfter != "" {
		if task.Due == nil {
			return false
		}
		taskDue := task.Due.Format("2006-01-02")
		if taskDue <= f.DueAfter {
			return false
		}
	}

	return true
}
