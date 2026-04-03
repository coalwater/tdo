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

## Development cycle

End-to-end workflow for any feature or fix. Execute autonomously — don't pause between steps.

1. **Todoist** — Create task (or link existing), set as current via `/todoist`
2. **Worktree** — `git worktree add .worktrees/<branch> -b <branch>` from main
3. **Tests first** — Write failing tests for the expected behavior (TDD)
4. **Implement** — Write minimal code to pass tests
5. **Verify** — `make test` + `golangci-lint run ./...` must both pass
6. **Commit** — Logical commits, conventional format (`feat:`, `fix:`, etc.)
7. **Merge** — Fast-forward merge to main, delete branch + worktree
8. **Tag** — Bump semver: `fix:` → patch, `feat:` → minor, breaking → major. `git tag v<next>`
9. **Push** — `git push origin main --tags` (confirm with user first)
10. **Close task** — Mark Todoist task complete

### Versioning

- Current: goreleaser on tag push (`.github/workflows/release.yaml`)
- Tags: `v<major>.<minor>.<patch>` — latest is `git tag --sort=-v:refname | head -1`
- Patch: bug fixes, chores, docs. Minor: new features. Major: breaking changes.

### Skip steps when they don't apply

- No Todoist task? Skip 1 and 10.
- Config-only change with no testable behavior? Skip 3.
- No version-worthy change (CI, docs)? Skip 8.

## Go env (if sandbox issues)

```bash
GOPATH=$TMPDIR/gopath GOCACHE=$TMPDIR/gocache go test ./...
```
