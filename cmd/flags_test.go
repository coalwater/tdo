package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFlags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantID        string
		wantJSON      bool
		wantHelp      bool
		wantRemaining []string
	}{
		{
			name:          "empty args",
			args:          []string{},
			wantRemaining: nil,
		},
		{
			name:          "only positional",
			args:          []string{"foo", "bar"},
			wantRemaining: []string{"foo", "bar"},
		},
		{
			name:          "--id with value",
			args:          []string{"--id", "abc123", "+urgent"},
			wantID:        "abc123",
			wantRemaining: []string{"+urgent"},
		},
		{
			name:          "--id=value",
			args:          []string{"--id=abc123", "+urgent"},
			wantID:        "abc123",
			wantRemaining: []string{"+urgent"},
		},
		{
			name:          "--json",
			args:          []string{"--json", "foo"},
			wantJSON:      true,
			wantRemaining: []string{"foo"},
		},
		{
			name:          "--help",
			args:          []string{"--help"},
			wantHelp:      true,
			wantRemaining: nil,
		},
		{
			name:          "-h shorthand",
			args:          []string{"-h"},
			wantHelp:      true,
			wantRemaining: nil,
		},
		{
			name:          "minus-label passes through",
			args:          []string{"-old", "+urgent", "-work"},
			wantRemaining: []string{"-old", "+urgent", "-work"},
		},
		{
			name:          "all flags mixed with labels",
			args:          []string{"--id", "xyz", "--json", "+urgent", "-old"},
			wantID:        "xyz",
			wantJSON:      true,
			wantRemaining: []string{"+urgent", "-old"},
		},
		{
			name:          "double-dash stops extraction",
			args:          []string{"--json", "--", "--id", "foo"},
			wantJSON:      true,
			wantRemaining: []string{"--", "--id", "foo"},
		},
		{
			name:          "unknown long flags pass through",
			args:          []string{"--unknown", "val"},
			wantRemaining: []string{"--unknown", "val"},
		},
		{
			name:          "unknown short flags pass through",
			args:          []string{"-x"},
			wantRemaining: []string{"-x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotJSON, gotHelp, gotRemaining := extractFlags(tt.args)
			assert.Equal(t, tt.wantID, gotID, "id")
			assert.Equal(t, tt.wantJSON, gotJSON, "json")
			assert.Equal(t, tt.wantHelp, gotHelp, "help")
			assert.Equal(t, tt.wantRemaining, gotRemaining, "remaining")
		})
	}
}
