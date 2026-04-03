package domain

import "strings"

// Filter holds parsed filter criteria. Multiple criteria are AND-ed during matching.
type Filter struct {
	Project   string
	Priority  *Priority
	HasLabels []string
	NotLabels []string
	DueBefore string
	DueAfter  string
}

// ParseFilter parses TaskWarrior-style filter arguments into a Filter.
func ParseFilter(args []string) Filter {
	var f Filter
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "project:"):
			f.Project = strings.TrimPrefix(arg, "project:")
		case strings.HasPrefix(arg, "priority:"):
			p := ParsePriority(strings.TrimPrefix(arg, "priority:"))
			f.Priority = &p
		case strings.HasPrefix(arg, "+"):
			f.HasLabels = append(f.HasLabels, strings.TrimPrefix(arg, "+"))
		case strings.HasPrefix(arg, "-"):
			f.NotLabels = append(f.NotLabels, strings.TrimPrefix(arg, "-"))
		case strings.HasPrefix(arg, "due.before:"):
			f.DueBefore = strings.TrimPrefix(arg, "due.before:")
		case strings.HasPrefix(arg, "due.after:"):
			f.DueAfter = strings.TrimPrefix(arg, "due.after:")
		}
	}
	return f
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
