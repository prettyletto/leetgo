# Task 044: Dashboard RPG Redesign

## Goal

Restyle Dashboard as an RPG Quest Board with Character HUD and Map Fragment.

## Context

Read:

- `docs/product/interface-styling.md`
- `docs/product/learning-system.md`
- Tasks 041-043

Dashboard is the main screen after Onboarding and must keep the primary Next Action visible.

## Dependencies

- Task 041: Terminal Palette and Symbol Set
- Task 042: Shared View Components
- Task 043: Unsupported Size and Responsive Shell

## Likely Files

- `internal/tui/dashboard_screen.go`
- `internal/tui/dashboard_screen_test.go`
- `internal/tui/views/`

## Implementation Notes

RPG labels:

- `Main Quest`: primary Next Action.
- `Side Quests`: Also Available, Review, Coming Soon.
- `Character HUD`: Level, XP, Streak, Achievements.
- `Map Fragment`: current Stage, next Blocker, mini Unlock Path context.

Rules:

- Primary Next Action gets the strongest tile.
- Recommendation Reason must remain visible.
- Verified actions should feel urgent.
- Locked/Coming Soon should be muted, not Danger.
- Narrow layout shows primary action first.
- Dashboard should normally avoid scrolling.

## Acceptance Criteria

- RPG Theme Dashboard uses Quest Board/Character HUD labels.
- Primary Next Action is visually dominant.
- Also Available and Coming Soon remain distinct.
- Rich and Plain Symbols both render status meaning.
- Wide, medium, and narrow tests confirm primary action remains visible.
- Existing Dashboard action behavior remains unchanged.

## Verification

Run:

```bash
gofmt -w internal/tui internal/tui/views
go test ./...
```
