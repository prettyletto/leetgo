# Task 007: Dashboard MVP

## Goal

Implement the default post-Onboarding Dashboard using Next Actions, Profile/game HUD, and Roadmap context.

## Dependencies

- Task 001: Config Profile and Preferences
- Task 003: Screen Shell Architecture
- Task 005: Theme System Foundation
- Task 006: Next Action Calculator

## Likely Files

- `internal/tui/screens/dashboard.go`
- `internal/tui/views/*`
- `internal/tui/model.go`

## Implementation Notes

Layout:

- Center: Next Actions.
- Left rail: Profile/game HUD.
- Right rail: Roadmap/Stage context.

Dashboard keybindings:

- `enter`: activate focused Next Action.
- `j/k` or arrows: move focus.
- `r`: Roadmap Detail.
- `s`: Solve Log or Solve Log Detail placeholder.
- `t`: cycle Theme and persist.
- `q`: quit.

Use responsive behavior:

- Wide: three columns.
- Medium: center first, rails below.
- Narrow: center and footer only.

## Acceptance Criteria

- Completed Onboarding opens Dashboard by default.
- Dashboard shows Display Name.
- Dashboard shows Level, XP, Streak, solved count, and Roadmap.
- Dashboard shows at least one Next Action when Problems are available.
- Theme cycling changes and saves config.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
