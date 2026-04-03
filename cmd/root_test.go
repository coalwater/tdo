package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "exact match", input: "modify", want: "modify"},
		{name: "prefix match", input: "mod", want: "modify"},
		{name: "ambiguous d", input: "d", wantErr: true},
		{name: "no match", input: "xyz", want: ""},
		{name: "empty string", input: "", want: ""},
		{name: "alias exact ls", input: "ls", want: "ls"},
		{name: "unambiguous de -> delete", input: "de", want: "delete"},
		{name: "unambiguous do -> done", input: "do", want: "done"},
		{name: "unambiguous sta -> start", input: "sta", want: "start"},
		{name: "unambiguous sto -> stop", input: "sto", want: "stop"},
		{name: "ambiguous s", input: "s", wantErr: true},
		{name: "ambiguous st", input: "st", wantErr: true},
		{name: "unambiguous ur -> url", input: "ur", want: "url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := matchCommand(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "ambiguous")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRewriteIDArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    []string
		wantErr bool
	}{
		{
			name: "exact command passthrough",
			args: []string{"tdo", "modify", "pri:H"},
			want: []string{"tdo", "modify", "pri:H"},
		},
		{
			name: "exact command with alias",
			args: []string{"tdo", "ls"},
			want: []string{"tdo", "ls"},
		},
		{
			name: "ID-first exact",
			args: []string{"tdo", "3", "modify", "pri:H"},
			want: []string{"tdo", "modify", "--id", "3", "pri:H"},
		},
		{
			name: "ID-first abbreviated",
			args: []string{"tdo", "3", "mod", "pri:H"},
			want: []string{"tdo", "modify", "--id", "3", "pri:H"},
		},
		{
			name: "command-first abbreviated",
			args: []string{"tdo", "mod", "pri:H"},
			want: []string{"tdo", "modify", "pri:H"},
		},
		{
			name: "abbreviation del -> delete",
			args: []string{"tdo", "del", "pri:H"},
			want: []string{"tdo", "delete", "pri:H"},
		},
		{
			name: "abbreviation ann -> annotate",
			args: []string{"tdo", "ann", "note"},
			want: []string{"tdo", "annotate", "note"},
		},
		{
			name: "abbreviation don -> done",
			args: []string{"tdo", "don"},
			want: []string{"tdo", "done"},
		},
		{
			name: "abbreviation inf -> info",
			args: []string{"tdo", "inf"},
			want: []string{"tdo", "info"},
		},
		{
			name: "abbreviation ne -> next",
			args: []string{"tdo", "ne"},
			want: []string{"tdo", "next"},
		},
		{
			name: "abbreviation pro -> projects",
			args: []string{"tdo", "pro"},
			want: []string{"tdo", "projects"},
		},
		{
			name: "abbreviation ta -> tags",
			args: []string{"tdo", "ta"},
			want: []string{"tdo", "tags"},
		},
		{
			name:    "ambiguous command d",
			args:    []string{"tdo", "3", "d"},
			wantErr: true,
		},
		{
			name:    "ambiguous command s",
			args:    []string{"tdo", "3", "s"},
			wantErr: true,
		},
		{
			name:    "ambiguous command st",
			args:    []string{"tdo", "st"},
			wantErr: true,
		},
		{
			name: "unambiguous de -> delete",
			args: []string{"tdo", "de"},
			want: []string{"tdo", "delete"},
		},
		{
			name: "unambiguous do -> done",
			args: []string{"tdo", "do"},
			want: []string{"tdo", "done"},
		},
		{
			name: "unambiguous sta -> start",
			args: []string{"tdo", "sta"},
			want: []string{"tdo", "start"},
		},
		{
			name: "unambiguous sto -> stop",
			args: []string{"tdo", "sto"},
			want: []string{"tdo", "stop"},
		},
		{
			name: "minimal args",
			args: []string{"tdo"},
			want: []string{"tdo"},
		},
		{
			name: "no command match in ID-first",
			args: []string{"tdo", "3", "something"},
			want: []string{"tdo", "3", "something"},
		},
		{
			name: "ID-first with extra args before command",
			args: []string{"tdo", "3", "pri:H", "modify"},
			want: []string{"tdo", "modify", "--id", "3", "pri:H"},
		},
		{
			name: "alias abbreviation ls exact",
			args: []string{"tdo", "3", "ls"},
			want: []string{"tdo", "ls", "--id", "3"},
		},
		{
			name: "ID-first url",
			args: []string{"tdo", "3", "url"},
			want: []string{"tdo", "url", "--id", "3"},
		},
		{
			name: "abbreviation ur -> url",
			args: []string{"tdo", "ur"},
			want: []string{"tdo", "url"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RewriteIDArgs(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
