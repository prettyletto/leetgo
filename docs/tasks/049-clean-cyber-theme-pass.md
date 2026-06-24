# Task 049: Clean and Cyber Theme Pass

## Goal

Apply the shared styling system to Clean Productivity and Cyber Dashboard after RPG Theme is polished.

## Context

Read:

- `docs/product/interface-styling.md`
- Tasks 041-048

Themes should preserve behavior and information architecture while changing presentation.

## Dependencies

- Task 041: Terminal Palette and Symbol Set
- Task 042: Shared View Components
- Tasks 044-048

## Likely Files

- `internal/tui/theme.go`
- TUI screen files as needed
- tests under `internal/tui/`

## Implementation Notes

Clean Productivity:

- compact.
- minimal color.
- direct labels.
- low motion.

Cyber Dashboard:

- high contrast.
- technical labels.
- cockpit panel feel.
- subtle ambient/focus motion when Motion Preference allows.

Theme-specific labels:

- Clean: Recommended, Available, Upcoming, Profile.
- Cyber: Primary Signal, Secondary Targets, Locked Signals, Operator.

Behavior must not change by Theme.

## Acceptance Criteria

- Clean Theme has distinct compact presentation.
- Cyber Theme has distinct cockpit presentation.
- Theme-specific labels render without changing action behavior.
- Rich and Plain Symbols work in all Themes.
- Motion Preference affects Cyber ambient/focus motion.
- Tests cover theme label differences and no behavior drift.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
