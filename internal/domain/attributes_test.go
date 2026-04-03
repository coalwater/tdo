package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAttributes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want ParsedAttributes
	}{
		{
			name: "basic content only",
			args: []string{"Fix", "the", "bug"},
			want: ParsedAttributes{Content: "Fix the bug"},
		},
		{
			name: "project attribute",
			args: []string{"task", "project:Work"},
			want: ParsedAttributes{Content: "task", Project: "Work"},
		},
		{
			name: "priority H",
			args: []string{"task", "priority:H"},
			want: ParsedAttributes{Content: "task", Priority: PriorityH},
		},
		{
			name: "priority M",
			args: []string{"task", "priority:M"},
			want: ParsedAttributes{Content: "task", Priority: PriorityM},
		},
		{
			name: "priority L",
			args: []string{"task", "priority:L"},
			want: ParsedAttributes{Content: "task", Priority: PriorityL},
		},
		{
			name: "priority case insensitive",
			args: []string{"task", "priority:h"},
			want: ParsedAttributes{Content: "task", Priority: PriorityH},
		},
		{
			name: "due date",
			args: []string{"task", "due:tomorrow"},
			want: ParsedAttributes{Content: "task", DueString: "tomorrow"},
		},
		{
			name: "recurrence",
			args: []string{"task", "recur:weekly"},
			want: ParsedAttributes{Content: "task", Recurrence: "weekly"},
		},
		{
			name: "add labels",
			args: []string{"task", "+urgent", "+bug"},
			want: ParsedAttributes{Content: "task", Labels: []string{"urgent", "bug"}},
		},
		{
			name: "remove labels",
			args: []string{"task", "-old"},
			want: ParsedAttributes{Content: "task", RemoveLabels: []string{"old"}},
		},
		{
			name: "description",
			args: []string{"task", "description:some notes"},
			want: ParsedAttributes{Content: "task", Description: "some notes"},
		},
		{
			name: "combined attributes",
			args: []string{"Buy", "groceries", "project:Home", "priority:M", "due:friday", "recur:weekly", "+shopping", "-old", "description:milk and eggs"},
			want: ParsedAttributes{
				Content:      "Buy groceries",
				Project:      "Home",
				Priority:     PriorityM,
				DueString:    "friday",
				Recurrence:   "weekly",
				Labels:       []string{"shopping"},
				RemoveLabels: []string{"old"},
				Description:  "milk and eggs",
			},
		},
		{
			name: "empty input",
			args: []string{},
			want: ParsedAttributes{},
		},
		{
			name: "nil input",
			args: nil,
			want: ParsedAttributes{},
		},
		{
			name: "content interspersed with attributes",
			args: []string{"Fix", "project:Work", "the", "priority:H", "bug"},
			want: ParsedAttributes{Content: "Fix the bug", Project: "Work", Priority: PriorityH},
		},
		{
			name: "invalid priority falls back to none",
			args: []string{"task", "priority:X"},
			want: ParsedAttributes{Content: "task", Priority: PriorityNone},
		},
		{
			name: "label with no name ignored",
			args: []string{"task", "+"},
			want: ParsedAttributes{Content: "task"},
		},
		{
			name: "remove label with no name ignored",
			args: []string{"task", "-"},
			want: ParsedAttributes{Content: "task"},
		},
		{
			name: "attribute with colon but empty value",
			args: []string{"task", "project:"},
			want: ParsedAttributes{Content: "task", Project: ""},
		},
		{
			name: "parent attribute",
			args: []string{"subtask", "parent:abc123"},
			want: ParsedAttributes{Content: "subtask", ParentID: "abc123"},
		},
		{
			name: "parent combined with other attributes",
			args: []string{"subtask", "parent:abc123", "project:Work", "priority:H", "+urgent"},
			want: ParsedAttributes{
				Content:  "subtask",
				ParentID: "abc123",
				Project:  "Work",
				Priority: PriorityH,
				Labels:   []string{"urgent"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAttributes(tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}
