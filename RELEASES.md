# Releases

## v0.5.1 — 2026-04-03

- fix: use TW next tag coefficient (15.0) for now label urgency

## v0.5.0 — 2026-04-03

- feat: support abbreviated attribute prefixes (e.g., `pro:Next` instead of `project:Next`)
- Added `matchAttr` helper with exact-match-wins and ambiguous-prefix error semantics
- `ParseAttributes` and `ParseFilter` now return errors on ambiguous abbreviations

## v0.4.0 — 2026-04-03

- feat: add `parent:<id>` attribute for creating subtasks
- fix: tighten cache file permissions to 0600
- chore: add security scanning and dependency updates

## v0.3.0 — 2026-04-03

- fix: compute urgency scores in table output

## v0.2.0 — 2026-04-03

- feat: add global `--json` flag for programmatic output
- test: add JSON output tests

## v0.1.0 — 2026-04-03

Initial release. TaskWarrior-compatible CLI backed by Todoist.

- Domain layer: task model, urgency scoring (TW formula), filter parser, attribute parser
- Todoist API v1 client with retry/rate-limit
- File-based TTL cache
- Colored table output with lipgloss
- CLI commands: add, list, next, modify, done, delete, start, stop, info, annotate, projects, tags
- ID-first routing (`tdo <id> <cmd>`)
- CI: lint, matrix testing, cross-platform builds
- Release: goreleaser on tag push
