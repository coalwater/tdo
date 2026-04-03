# tdo

TaskWarrior-compatible CLI backed by Todoist.

## Build

```bash
go build -o tdo .
```

## Test

```bash
make test          # all unit + integration tests
make test-cover    # with coverage report
```

## Architecture

Layered: `cmd/` (Cobra CLI) -> `internal/domain/` (business logic) -> `internal/backend/` (interface + Todoist impl)

- `internal/domain/` — Task model, urgency scoring (TW formula), filter parser, attribute parser, ID resolver
- `internal/backend/todoist/` — Todoist API v1 client with retry/rate-limit
- `internal/cache/` — File-based TTL cache
- `internal/display/` — Colored table output with lipgloss
- `cmd/` — All CLI commands, TW-style ID-first routing

## Key patterns

- Backend is an interface — tests use mocks, not the real API
- TDD: table-driven tests with testify, golden files for display
- ID-first routing: `tdo <id> <cmd>` rewritten to `tdo <cmd> --id <id>` before Cobra dispatch
- Todoist API v1 wraps list responses in `{"results": [...]}`

## Go env (if sandbox issues)

```bash
GOPATH=$TMPDIR/gopath GOCACHE=$TMPDIR/gocache go test ./...
```
