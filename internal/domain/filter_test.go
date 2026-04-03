package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFilter(t *testing.T) {
	h := PriorityH
	m := PriorityM
	l := PriorityL

	tests := []struct {
		name    string
		args    []string
		want    Filter
		wantErr bool
	}{
		{
			name: "empty input",
			args: []string{},
			want: Filter{},
		},
		{
			name: "single project",
			args: []string{"project:Work"},
			want: Filter{Project: "Work"},
		},
		{
			name: "single label",
			args: []string{"+urgent"},
			want: Filter{HasLabels: []string{"urgent"}},
		},
		{
			name: "negative label",
			args: []string{"-blocked"},
			want: Filter{NotLabels: []string{"blocked"}},
		},
		{
			name: "priority H",
			args: []string{"priority:H"},
			want: Filter{Priority: &h},
		},
		{
			name: "priority M",
			args: []string{"priority:M"},
			want: Filter{Priority: &m},
		},
		{
			name: "priority L",
			args: []string{"priority:L"},
			want: Filter{Priority: &l},
		},
		{
			name: "due.before",
			args: []string{"due.before:2026-04-10"},
			want: Filter{DueBefore: "2026-04-10"},
		},
		{
			name: "due.after",
			args: []string{"due.after:2026-03-01"},
			want: Filter{DueAfter: "2026-03-01"},
		},
		{
			name: "combined filters",
			args: []string{"project:Home", "+errands", "-deferred", "priority:M", "due.before:2026-05-01"},
			want: Filter{
				Project:   "Home",
				Priority:  &m,
				HasLabels: []string{"errands"},
				NotLabels: []string{"deferred"},
				DueBefore: "2026-05-01",
			},
		},
		{
			name: "multiple positive labels",
			args: []string{"+urgent", "+important"},
			want: Filter{HasLabels: []string{"urgent", "important"}},
		},
		// Abbreviated prefix tests
		{
			name: "abbreviated pro -> project",
			args: []string{"pro:Work"},
			want: Filter{Project: "Work"},
		},
		{
			name: "abbreviated pri -> priority",
			args: []string{"pri:H"},
			want: Filter{Priority: &h},
		},
		{
			name: "abbreviated due.b -> due.before",
			args: []string{"due.b:2026-04-10"},
			want: Filter{DueBefore: "2026-04-10"},
		},
		{
			name: "abbreviated due.a -> due.after",
			args: []string{"due.a:2026-03-01"},
			want: Filter{DueAfter: "2026-03-01"},
		},
		{
			name:    "ambiguous d: errors in filter (due.before/due.after)",
			args:    []string{"d:2026-04-10"},
			wantErr: true,
		},
		{
			name:    "ambiguous du: errors in filter (due.before/due.after)",
			args:    []string{"du:2026-04-10"},
			wantErr: true,
		},
		{
			name: "unrecognized prefix ignored in filter",
			args: []string{"foo:bar"},
			want: Filter{},
		},
		{
			name: "par: not a filter attr, ignored",
			args: []string{"par:abc"},
			want: Filter{},
		},
		{
			name: "combined abbreviated filters",
			args: []string{"pro:Home", "pri:L", "due.b:2026-05-01"},
			want: Filter{
				Project:   "Home",
				Priority:  &l,
				DueBefore: "2026-05-01",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFilter(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilterMatch(t *testing.T) {
	h := PriorityH
	m := PriorityM

	dueDate := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)

	baseTask := Task{
		ID:       "1",
		Content:  "Buy groceries",
		Project:  "Home",
		Priority: PriorityH,
		Labels:   []string{"errands", "urgent"},
		Due:      &dueDate,
	}

	noDueTask := Task{
		ID:       "2",
		Content:  "Think about life",
		Project:  "Personal",
		Priority: PriorityM,
		Labels:   []string{"reflection"},
	}

	tests := []struct {
		name   string
		filter Filter
		task   Task
		want   bool
	}{
		{
			name:   "empty filter matches everything",
			filter: Filter{},
			task:   baseTask,
			want:   true,
		},
		{
			name:   "project match case-insensitive",
			filter: Filter{Project: "home"},
			task:   baseTask,
			want:   true,
		},
		{
			name:   "project mismatch",
			filter: Filter{Project: "Work"},
			task:   baseTask,
			want:   false,
		},
		{
			name:   "label present",
			filter: Filter{HasLabels: []string{"errands"}},
			task:   baseTask,
			want:   true,
		},
		{
			name:   "label absent",
			filter: Filter{HasLabels: []string{"fitness"}},
			task:   baseTask,
			want:   false,
		},
		{
			name:   "negative label - task does not have it",
			filter: Filter{NotLabels: []string{"blocked"}},
			task:   baseTask,
			want:   true,
		},
		{
			name:   "negative label - task has it",
			filter: Filter{NotLabels: []string{"urgent"}},
			task:   baseTask,
			want:   false,
		},
		{
			name:   "priority match",
			filter: Filter{Priority: &h},
			task:   baseTask,
			want:   true,
		},
		{
			name:   "priority mismatch",
			filter: Filter{Priority: &m},
			task:   baseTask,
			want:   false,
		},
		{
			name:   "due.before - task due is before deadline",
			filter: Filter{DueBefore: "2026-04-10"},
			task:   baseTask,
			want:   true,
		},
		{
			name:   "due.before - task due is after deadline",
			filter: Filter{DueBefore: "2026-04-01"},
			task:   baseTask,
			want:   false,
		},
		{
			name:   "due.after - task due is after date",
			filter: Filter{DueAfter: "2026-04-01"},
			task:   baseTask,
			want:   true,
		},
		{
			name:   "due.after - task due is before date",
			filter: Filter{DueAfter: "2026-04-10"},
			task:   baseTask,
			want:   false,
		},
		{
			name:   "due.before - task has no due date",
			filter: Filter{DueBefore: "2026-04-10"},
			task:   noDueTask,
			want:   false,
		},
		{
			name:   "due.after - task has no due date",
			filter: Filter{DueAfter: "2026-04-01"},
			task:   noDueTask,
			want:   false,
		},
		{
			name: "compound filter - all conditions met",
			filter: Filter{
				Project:   "home",
				Priority:  &h,
				HasLabels: []string{"errands"},
				NotLabels: []string{"blocked"},
				DueBefore: "2026-04-10",
			},
			task: baseTask,
			want: true,
		},
		{
			name: "compound filter - one condition fails",
			filter: Filter{
				Project:   "home",
				Priority:  &m, // mismatch: task is H
				HasLabels: []string{"errands"},
			},
			task: baseTask,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.Match(tt.task)
			assert.Equal(t, tt.want, got)
		})
	}
}
