# tdo — TaskWarrior-Compatible CLI Backed by Todoist

## Context

No maintained tool provides a TaskWarrior-compatible CLI frontend backed by a cloud service. The existing `todo` bash CLI (~970 lines) proves the concept works — Todoist as backend with urgency scoring, caching, and focus management. `tdo` is a ground-up rewrite in Go that gives TaskWarrior users a familiar CLI while getting Todoist's cloud sync, mobile app, and collaboration for free.

## Design Decisions

| Decision | Choice | Why |
|---|---|---|
| Language | Go | Single binary, readable by anyone, Cobra+BubbleTea ecosystem |
| CLI framework | Cobra | Industry standard, supports TW-style `<id> <command>` routing |
| Architecture | Layered (CLI -> Domain -> Backend interface) | Testable, clean separation |
| IDs | Todoist native IDs, with positional + fuzzy name autocomplete | No mapping layer, deterministic |
| Urgency | TaskWarrior's exact formula (supported factors only) | TW users expect familiar scoring |
| Testing | Table-driven + testify + golden files, extensive TDD | Human/AI readable, mock backend |
| Claude integration | Separate (stays in ~/.claude/) | tdo is a generic tool |
| Distribution | goreleaser + GitHub Releases | Simple, cross-platform |
| Config | Env vars only (v1) | No config file complexity |

## Architecture

```
+-------------------------------+
|  CLI Layer (Cobra)            |
|  Parses args, formats output  |
+---------------+---------------+
                |
+---------------v---------------+
|  Domain Layer                 |
|  Urgency, filtering, ID map  |
|  Task model, attribute parse  |
+---------------+---------------+
                |
        +-------v-------+
        |  Backend      |
        |  Interface    |
        +---+-------+---+
            |       |
      +-----v--+ +-v------+
      |Todoist | | Mock   |
      |REST API| |(tests) |
      +--------+ +--------+
```

## Command Grammar

Pattern: `tdo [filter] <command> [args]`

```bash
# Task operations (ID before command -- TW style)
tdo <id> done
tdo <id> delete
tdo <id> modify <attrs>
tdo <id> start           # sets "now" label
tdo <id> stop            # removes "now" label
tdo <id> info
tdo <id> annotate <text>

# Creating tasks
tdo add <description> [attrs]

# Listing
tdo list [filter-expr]
tdo next                 # urgency-ranked, fits terminal height
tdo projects
tdo tags
tdo version
```

### Cobra routing for ID-first syntax

Root command's `RunE` checks if `args[0]` is a known subcommand name. If not, treats it as an ID/filter, expects `args[1]` to be the command, and dispatches accordingly.

### Attribute syntax (TW-compatible)

```
project:<name>       # assign project
priority:H|M|L       # set priority (maps to Todoist p1/p3/p2)
due:<date>           # natural language date
recur:<interval>     # daily, weekly, monthly, every 3 days, etc.
+<tag>               # add label
-<tag>               # remove label (modify only)
description:<text>   # task description/notes
```

### Filter syntax (for list command)

```bash
tdo list project:Work           # by project
tdo list +urgent                # by label
tdo list priority:H             # by priority
tdo list due.before:tomorrow    # by due date
tdo list project:Work +urgent   # compound AND
```

### ID resolution

Tried in order:
1. Exact Todoist ID match
2. Positional number from last `list`/`next` output
3. Fuzzy name match (shortest unique prefix)

Position map stored in cache dir, written after every `list`/`next`.

## Backend Interface

```go
type Backend interface {
    ListTasks(ctx context.Context, filter string) ([]domain.Task, error)
    GetTask(ctx context.Context, id string) (*domain.Task, error)
    CreateTask(ctx context.Context, params domain.CreateParams) (*domain.Task, error)
    UpdateTask(ctx context.Context, id string, params domain.UpdateParams) error
    CompleteTask(ctx context.Context, id string) error
    ReopenTask(ctx context.Context, id string) error
    DeleteTask(ctx context.Context, id string) error
    MoveTask(ctx context.Context, id string, projectID string) error

    AddComment(ctx context.Context, taskID string, text string) (*domain.Comment, error)
    ListComments(ctx context.Context, taskID string) ([]domain.Comment, error)

    ListProjects(ctx context.Context) ([]domain.Project, error)
    ListLabels(ctx context.Context) ([]domain.Label, error)
}
```

## Urgency Scoring -- TaskWarrior's Exact Formula

`urgency = sum(coefficient * factor)`

### Supported factors

| Factor | Coefficient | Calculation |
|---|---|---|
| Priority H | 6.0 | 1.0 if priority is H, else 0.0 |
| Priority M | 3.9 | 1.0 if priority is M, else 0.0 |
| Priority L | 1.8 | 1.0 if priority is L, else 0.0 |
| Due date | 12.0 | See due factor below |
| Active (started) | 4.0 | 1.0 if has "now" label |
| Age | 2.0 | `min(days_since_creation / 365, 1.0)` |
| Annotations | 1.0 | 0 comments: 0.0, 1: 0.8, 2: 0.9, 3+: 1.0 |
| Tags | 1.0 | 0 labels: 0.0, 1: 0.8, 2: 0.9, 3+: 1.0 |
| Project | 1.0 | 1.0 if has project, else 0.0 |

### Dropped factors (not applicable to Todoist backend)

| Factor | TW Coefficient | Why dropped |
|---|---|---|
| Blocking | 8.0 | Dependencies not supported in v1 |
| Scheduled | 5.0 | No Todoist equivalent |
| Waiting | -3.0 | Could add later via label |
| Blocked | -5.0 | Dependencies not supported in v1 |

### Due date factor (exact TW algorithm)

```
days_overdue = (now - due) / 86400    // positive = overdue

if days_overdue >= 7.0   -> 1.0
if days_overdue >= -14.0 -> ((days_overdue + 14.0) * 0.8 / 21.0) + 0.2
else                     -> 0.2
no due date              -> 0.0
```

Linear ramp from 0.2 to 1.0 over the 21-day window (14 days before to 7 days after due).

## Caching

Location: `$XDG_CACHE_HOME/tdo/` or `~/.cache/tdo/`

| File | Content | TTL |
|---|---|---|
| `tasks.json` | All open tasks with enriched project names | 300s (configurable) |
| `projects.json` | Project list | 1 day |
| `labels.json` | Label list | 1 day |
| `positions.json` | Last listing: display # to Todoist ID | Until next list/next |

Invalidation: `add`, `done`, `delete`, `modify`, `start`, `stop` invalidate `tasks.json`. Background refresh via goroutine (non-blocking).

## Configuration

Env vars only for v1:

| Var | Required | Default | Description |
|---|---|---|---|
| `TODOIST_API_KEY` | Yes | -- | Todoist API token |
| `TDO_CACHE_TTL` | No | 300 | Cache TTL in seconds |
| `TDO_NOW_LABEL` | No | "now" | Label used for focus/start |

## Priority Mapping (Todoist <-> TW)

| TW Priority | tdo attribute | Todoist API value |
|---|---|---|
| H (high) | `priority:H` | 4 |
| M (medium) | `priority:M` | 3 |
| L (low) | `priority:L` | 2 |
| None | (default) | 1 |

Note: Todoist's numbering is inverted (4 = highest). tdo handles the translation.

## Testing Strategy

### Principles

- TDD: Tests written before implementation
- Human-readable: Table-driven tests with descriptive names that read like specs
- AI-friendly: Consistent patterns, clear input to output, no magic
- No Todoist API in unit tests: Mock backend interface everywhere

### Test layers

1. Domain tests (unit, mock backend) -- table-driven
2. CLI tests (integration, mock backend) -- golden file comparison
3. Todoist client tests (unit, httptest) -- canned responses
4. End-to-end tests (optional, CI-only) -- real API, gated behind TDO_E2E=1

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/lipgloss` | Styled terminal output |
| `github.com/stretchr/testify` | Test assertions + mocks |

Minimal dependency footprint. stdlib `net/http` for API calls, stdlib `encoding/json` for JSON.

## Out of Scope (v1)

- TUI (future, Bubble Tea)
- Contexts
- Reports / burndown / calendar
- Custom UDAs
- Dependencies (blocked/blocking)
- Undo
- Config file (env vars only)
- Homebrew tap
- Shell completions (v1.1)
