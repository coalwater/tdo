package domain

import "strings"

// ParsedAttributes holds the structured result of parsing TaskWarrior-style
// command arguments.
type ParsedAttributes struct {
	Content      string
	Project      string
	Priority     Priority
	DueString    string
	Recurrence   string
	Labels       []string
	RemoveLabels []string
	Description  string
}

// ParseAttributes parses a slice of command arguments into structured
// task attributes. Anything that doesn't match a known pattern is
// collected as content words.
func ParseAttributes(args []string) ParsedAttributes {
	var p ParsedAttributes
	var contentWords []string

	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "project:"):
			p.Project = arg[len("project:"):]

		case strings.HasPrefix(arg, "priority:"):
			p.Priority = ParsePriority(arg[len("priority:"):])

		case strings.HasPrefix(arg, "due:"):
			p.DueString = arg[len("due:"):]

		case strings.HasPrefix(arg, "recur:"):
			p.Recurrence = arg[len("recur:"):]

		case strings.HasPrefix(arg, "description:"):
			p.Description = arg[len("description:"):]

		case strings.HasPrefix(arg, "+"):
			if len(arg) > 1 {
				p.Labels = append(p.Labels, arg[1:])
			}

		case strings.HasPrefix(arg, "-"):
			if len(arg) > 1 {
				p.RemoveLabels = append(p.RemoveLabels, arg[1:])
			}

		default:
			contentWords = append(contentWords, arg)
		}
	}

	p.Content = strings.Join(contentWords, " ")
	return p
}
