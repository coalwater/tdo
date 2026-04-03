# Agent Guide

Instructions for AI agents contributing to tdo.

## Build & Test

```bash
# Build
go build -o tdo .

# Test
go test ./... -v -count=1

# Sandbox-safe (if permission errors)
GOPATH=$TMPDIR/gopath GOCACHE=$TMPDIR/gocache go test ./...

# Lint
gofmt -w .
go vet ./...
```

## Architecture

```
cmd/           → Cobra CLI commands (thin — parse args, call domain, format output)
internal/
  domain/      → Business logic (urgency, filters, attributes, resolver)
  backend/     → Backend interface + Todoist implementation
  cache/       → File-based TTL cache
  display/     → Terminal output formatting (lipgloss)
```

**Data flows down:** `cmd/` → `domain/` → `backend/`. Never import upward.

**Backend is an interface.** Tests mock it — never hit the Todoist API in unit tests.

## Adding a New Command

1. Create `cmd/<name>.go` with a `cobra.Command`
2. If it takes a task ID, add the hidden `--id` flag pattern:
   ```go
   func init() {
       myCmd.Flags().String("id", "", "Task ID (set automatically by ID-first routing)")
       _ = myCmd.Flags().MarkHidden("id")
   }
   ```
3. Register in `cmd/root.go` `init()` and add to `knownCommands` map
4. Support `--json` output: check `jsonOutput` global, use `writeJSON()` from `cmd/json.go`
5. Invalidate cache after mutations: `_ = app.Cache.InvalidateTasks()`

## Adding a New Domain Feature

1. Write tests first in `internal/domain/<feature>_test.go`
2. Implement in `internal/domain/<feature>.go`
3. Tests must pass before moving to integration

## Testing Standards

**TDD.** Write the test, watch it fail, then implement.

**Table-driven tests** with descriptive names:
```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
    }{
        {
            name:     "describes the scenario in plain english",
            input:    ...,
            expected: ...,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := DoThing(tt.input)
            assert.Equal(t, tt.expected, got)
        })
    }
}
```

**Assertions:** Use `github.com/stretchr/testify/assert` and `require`.

**Backend tests:** Use `net/http/httptest` with canned JSON responses. See `internal/backend/todoist/client_test.go` for the pattern.

**Golden files:** For display output tests, use files in `testdata/`.

## JSON Output

Every command supports `--json`. The global `jsonOutput` bool is set by root's persistent flag.

Pattern:
```go
if jsonOutput {
    return writeJSON(cmd.OutOrStdout(), data)
}
// else format for humans
```

JSON types are defined in `cmd/json.go`. When adding new output types, add the struct there.

## Todoist API v1

- Base URL: `https://api.todoist.com/api/v1`
- List endpoints return `{"results": [...], "next_cursor": ...}`
- Single-object endpoints return the object directly
- Priority mapping: TW H=4, M=3, L=2, None=1 (inverted)
- Task fields: `checked` (not `is_completed`), `added_at` (not `created_at`), `note_count` (not `comment_count`)

## Commit Messages

Conventional commits, no emojis:
```
feat: add shell completions
fix: handle nil due date in urgency calc
test: add filter edge case tests
chore: update dependencies
refactor: extract label merge logic
```

## Code Style

- `gofmt` all files — CI enforces this
- No trailing whitespace
- errcheck excluded in test files (see `.golangci.yml`)
- Minimal diffs — don't reformat code you didn't change
