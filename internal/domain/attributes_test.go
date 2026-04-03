package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchAttr(t *testing.T) {
	known := []string{"project", "priority", "scheduled", "due", "recur", "description", "parent"}

	tests := []struct {
		name      string
		arg       string
		known     []string
		wantAttr  string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "no colon - not an attribute",
			arg:       "hello",
			known:     known,
			wantAttr:  "",
			wantValue: "",
		},
		{
			name:      "exact match",
			arg:       "project:Work",
			known:     known,
			wantAttr:  "project",
			wantValue: "Work",
		},
		{
			name:      "unique prefix pro -> project",
			arg:       "pro:Work",
			known:     known,
			wantAttr:  "project",
			wantValue: "Work",
		},
		{
			name:      "unique prefix pri -> priority",
			arg:       "pri:H",
			known:     known,
			wantAttr:  "priority",
			wantValue: "H",
		},
		{
			name:      "unique prefix du -> due",
			arg:       "du:tomorrow",
			known:     known,
			wantAttr:  "due",
			wantValue: "tomorrow",
		},
		{
			name:      "unique prefix des -> description",
			arg:       "des:notes",
			known:     known,
			wantAttr:  "description",
			wantValue: "notes",
		},
		{
			name:      "unique prefix par -> parent",
			arg:       "par:abc",
			known:     known,
			wantAttr:  "parent",
			wantValue: "abc",
		},
		{
			name:      "unique prefix rec -> recur",
			arg:       "rec:weekly",
			known:     known,
			wantAttr:  "recur",
			wantValue: "weekly",
		},
		{
			name:      "unique prefix sc -> scheduled",
			arg:       "sc:tomorrow",
			known:     known,
			wantAttr:  "scheduled",
			wantValue: "tomorrow",
		},
		{
			name:    "ambiguous prefix d -> due/description",
			arg:     "d:tomorrow",
			known:   known,
			wantErr: true,
		},
		{
			name:    "ambiguous prefix p -> project/priority/parent",
			arg:     "p:Work",
			known:   known,
			wantErr: true,
		},
		{
			name:    "unknown alpha key errors",
			arg:     "foo:bar",
			known:   known,
			wantErr: true,
		},
		{
			name:    "dotted unknown key errors",
			arg:     "due.bfore:x",
			known:   known,
			wantErr: true,
		},
		{
			name:      "numeric key passes through",
			arg:       "10:30",
			known:     known,
			wantAttr:  "",
			wantValue: "",
		},
		{
			name:      "single char key passes through",
			arg:       "x:y",
			known:     known,
			wantAttr:  "",
			wantValue: "",
		},
		{
			name:      "mixed alphanum key passes through",
			arg:       "h264:profile",
			known:     known,
			wantAttr:  "",
			wantValue: "",
		},
		{
			name:      "empty key passes through",
			arg:       ":value",
			known:     known,
			wantAttr:  "",
			wantValue: "",
		},
		{
			name:      "all dots key passes through",
			arg:       "...:value",
			known:     known,
			wantAttr:  "",
			wantValue: "",
		},
		{
			name:      "empty value",
			arg:       "project:",
			known:     known,
			wantAttr:  "project",
			wantValue: "",
		},
		{
			name:      "value with colons",
			arg:       "due:2026-04-03T10:00:00",
			known:     known,
			wantAttr:  "due",
			wantValue: "2026-04-03T10:00:00",
		},
		{
			name:      "exact match wins over prefix ambiguity",
			arg:       "due:tomorrow",
			known:     []string{"due", "due.before", "due.after"},
			wantAttr:  "due",
			wantValue: "tomorrow",
		},
		{
			name:    "ambiguous in filter context due.b matches nothing uniquely",
			arg:     "d:tomorrow",
			known:   []string{"due.before", "due.after"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr, value, err := matchAttr(tt.arg, tt.known)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantAttr, attr)
			assert.Equal(t, tt.wantValue, value)
		})
	}
}

func TestParseAttributes(t *testing.T) {
	// Wednesday 2026-04-15 10:30:00 UTC — same as dateparse ref
	now := time.Date(2026, 4, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		args    []string
		want    ParsedAttributes
		wantErr bool
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
			name: "due date parses to resolved date",
			args: []string{"task", "due:tomorrow"},
			want: ParsedAttributes{Content: "task", DueDate: "2026-04-16"},
		},
		{
			name: "due date with time component",
			args: []string{"task", "due:today+8h"},
			want: ParsedAttributes{Content: "task", DueDate: "2026-04-15T08:00:00"},
		},
		{
			name: "due:today+12h produces datetime format",
			args: []string{"task", "due:today+12h"},
			want: ParsedAttributes{Content: "task", DueDate: "2026-04-15T12:00:00"},
		},
		{
			name: "due:now+30min produces datetime format",
			args: []string{"task", "due:now+30min"},
			want: ParsedAttributes{Content: "task", DueDate: "2026-04-15T11:00:00"},
		},
		{
			name: "due:eod produces datetime format (23:59:59)",
			args: []string{"task", "due:eod"},
			want: ParsedAttributes{Content: "task", DueDate: "2026-04-15T23:59:59"},
		},
		{
			name: "due:eom-1d produces date-only format",
			args: []string{"task", "due:eom-1d"},
			// Apr 30 23:59:59 - 1d = Apr 29 23:59:59
			want: ParsedAttributes{Content: "task", DueDate: "2026-04-29T23:59:59"},
		},
		{
			name: "due:friday+1w produces date-only (no time)",
			args: []string{"task", "due:friday+1w"},
			// next friday Apr 17 + 7d = Apr 24
			want: ParsedAttributes{Content: "task", DueDate: "2026-04-24"},
		},
		{
			name: "due:sonm resolves next month start",
			args: []string{"task", "due:sonm"},
			want: ParsedAttributes{Content: "task", DueDate: "2026-05-01"},
		},
		{
			name: "due:2d resolves relative from now",
			args: []string{"task", "due:2d"},
			// now + 2d = Apr 17 10:30:00
			want: ParsedAttributes{Content: "task", DueDate: "2026-04-17T10:30:00"},
		},
		{
			name: "due ISO date passthrough",
			args: []string{"task", "due:2026-05-01"},
			want: ParsedAttributes{Content: "task", DueDate: "2026-05-01"},
		},
		{
			name: "due ISO datetime passthrough",
			args: []string{"task", "due:2026-05-01T14:00:00"},
			want: ParsedAttributes{Content: "task", DueDate: "2026-05-01T14:00:00"},
		},
		{
			name: "scheduled attribute passed raw",
			args: []string{"task", "scheduled:tomorrow"},
			want: ParsedAttributes{Content: "task", ScheduledString: "tomorrow"},
		},
		{
			name: "scheduled with NLP string",
			args: []string{"task", "scheduled:next monday"},
			want: ParsedAttributes{Content: "task", ScheduledString: "next monday"},
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
			args: []string{"Buy", "groceries", "project:Home", "priority:M", "due:friday", "scheduled:tomorrow", "recur:weekly", "+shopping", "-old", "description:milk and eggs"},
			want: ParsedAttributes{
				Content:         "Buy groceries",
				Project:         "Home",
				Priority:        PriorityM,
				DueDate:         "2026-04-17", // next friday from Wed Apr 15
				ScheduledString: "tomorrow",
				Recurrence:      "weekly",
				Labels:          []string{"shopping"},
				RemoveLabels:    []string{"old"},
				Description:     "milk and eggs",
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
		// Abbreviated prefix tests
		{
			name: "abbreviated pro -> project",
			args: []string{"task", "pro:Next"},
			want: ParsedAttributes{Content: "task", Project: "Next"},
		},
		{
			name: "abbreviated pri -> priority",
			args: []string{"task", "pri:H"},
			want: ParsedAttributes{Content: "task", Priority: PriorityH},
		},
		{
			name: "abbreviated du -> due",
			args: []string{"task", "du:tomorrow"},
			want: ParsedAttributes{Content: "task", DueDate: "2026-04-16"},
		},
		{
			name: "abbreviated des -> description",
			args: []string{"task", "des:notes"},
			want: ParsedAttributes{Content: "task", Description: "notes"},
		},
		{
			name: "abbreviated par -> parent",
			args: []string{"task", "par:abc"},
			want: ParsedAttributes{Content: "task", ParentID: "abc"},
		},
		{
			name: "abbreviated rec -> recur",
			args: []string{"task", "rec:weekly"},
			want: ParsedAttributes{Content: "task", Recurrence: "weekly"},
		},
		{
			name: "abbreviated sc -> scheduled",
			args: []string{"task", "sc:next week"},
			want: ParsedAttributes{Content: "task", ScheduledString: "next week"},
		},
		{
			name:    "ambiguous d: errors (due/description)",
			args:    []string{"task", "d:tomorrow"},
			wantErr: true,
		},
		{
			name:    "ambiguous p: errors (project/priority/parent)",
			args:    []string{"task", "p:Work"},
			wantErr: true,
		},
		{
			name:    "unknown attr errors",
			args:    []string{"task", "foo:bar"},
			wantErr: true,
		},
		{
			name: "content with numeric colon ok",
			args: []string{"meeting", "10:30"},
			want: ParsedAttributes{Content: "meeting 10:30"},
		},
		{
			name: "combined abbreviated attrs",
			args: []string{"task", "pro:Next", "pri:M", "du:friday"},
			want: ParsedAttributes{Content: "task", Project: "Next", Priority: PriorityM, DueDate: "2026-04-17"},
		},
		{
			name:    "due with invalid expression errors",
			args:    []string{"task", "due:notadate"},
			wantErr: true,
		},
		// -- separator tests
		{
			name: "-- stops attr parsing",
			args: []string{"task", "--", "due:friday"},
			want: ParsedAttributes{Content: "task due:friday"},
		},
		{
			name: "-- stops label parsing",
			args: []string{"task", "--", "+urgent"},
			want: ParsedAttributes{Content: "task +urgent"},
		},
		{
			name: "attrs before -- still parsed",
			args: []string{"task", "project:Work", "--", "foo:bar"},
			want: ParsedAttributes{Content: "task foo:bar", Project: "Work"},
		},
		{
			name: "-- only marker becomes empty content",
			args: []string{"--"},
			want: ParsedAttributes{},
		},
		{
			name: "-- at end",
			args: []string{"task", "--"},
			want: ParsedAttributes{Content: "task"},
		},
		{
			name: "multiple colons in content after --",
			args: []string{"--", "http://example.com"},
			want: ParsedAttributes{Content: "http://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAttributes(tt.args, now)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
