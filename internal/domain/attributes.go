package domain

import (
	"fmt"
	"strings"
)

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
	ParentID     string
}

// attributeAttrs is the known attribute list for ParseAttributes.
var attributeAttrs = []string{"project", "priority", "due", "recur", "description", "parent"}

// matchAttr checks if arg is a colon-prefixed attribute matching one of the
// known names. Returns the matched name, the value, and an error.
// Exact match always wins. Otherwise requires unambiguous prefix.
// Returns ("", "", nil) for no match, or error for ambiguous match.
func matchAttr(arg string, known []string) (string, string, error) {
	idx := strings.IndexByte(arg, ':')
	if idx < 0 {
		return "", "", nil
	}

	typed := arg[:idx]
	value := arg[idx+1:]

	// Exact match always wins.
	for _, k := range known {
		if k == typed {
			return k, value, nil
		}
	}

	// Collect prefix matches.
	var matches []string
	for _, k := range known {
		if strings.HasPrefix(k, typed) {
			matches = append(matches, k)
		}
	}

	switch len(matches) {
	case 0:
		return "", "", nil
	case 1:
		return matches[0], value, nil
	default:
		return "", "", fmt.Errorf("ambiguous attribute '%s:' — matches: %s", typed, strings.Join(matches, ", "))
	}
}

// ParseAttributes parses a slice of command arguments into structured
// task attributes. Anything that doesn't match a known pattern is
// collected as content words.
func ParseAttributes(args []string) (ParsedAttributes, error) {
	var p ParsedAttributes
	var contentWords []string

	for _, arg := range args {
		// Handle labels before attribute matching.
		if strings.HasPrefix(arg, "+") {
			if len(arg) > 1 {
				p.Labels = append(p.Labels, arg[1:])
			}
			continue
		}
		if strings.HasPrefix(arg, "-") {
			if len(arg) > 1 {
				p.RemoveLabels = append(p.RemoveLabels, arg[1:])
			}
			continue
		}

		attr, value, err := matchAttr(arg, attributeAttrs)
		if err != nil {
			return ParsedAttributes{}, err
		}

		switch attr {
		case "project":
			p.Project = value
		case "priority":
			p.Priority = ParsePriority(value)
		case "due":
			p.DueString = value
		case "recur":
			p.Recurrence = value
		case "description":
			p.Description = value
		case "parent":
			p.ParentID = value
		default:
			contentWords = append(contentWords, arg)
		}
	}

	p.Content = strings.Join(contentWords, " ")
	return p, nil
}
