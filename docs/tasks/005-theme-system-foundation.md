# Task 005: Theme System Foundation

## Goal

Create a Theme system that supports RPG skill tree, clean productivity, and cyber dashboard styling.

## Context

Theme is a Practice Preference and affects all TUI screens. See `docs/product/tui-screen-spec.md`.

## Dependencies

- Task 001: Config Profile and Preferences
- Task 003: Screen Shell Architecture

## Likely Files

- `internal/tui/theme.go`
- `internal/tui/model.go`
- `internal/tui/views/*`
- `internal/config/config.go`

## Implementation Notes

Define supported Theme IDs:

- `rpg-skill-tree`
- `clean-productivity`
- `cyber-dashboard`

Create theme tokens rather than scattered raw colors:

- primary accent
- secondary accent
- border color
- muted text
- success
- warning
- danger
- panel style
- focused panel style

Do not do full animation in this task. Only make styles selectable and reusable.

## Acceptance Criteria

- Theme IDs validate through config.
- Root TUI can load selected Theme.
- Existing screens/views can render using Theme tokens.
- Tests cover Theme lookup and invalid Theme handling.

## Verification

Run:

```bash
gofmt -w internal/tui internal/config
go test ./...
```
