# Task 011: Theme Polish and Motion

## Goal

Polish the TUI visual design and add restrained motion according to Theme rules.

## Dependencies

- Task 005: Theme System Foundation
- Task 007: Dashboard MVP
- Task 008/009/010 detail screens preferred

## Likely Files

- `internal/tui/theme.go`
- `internal/tui/screens/*`
- `internal/tui/views/*`

## Implementation Notes

Themes:

- RPG skill tree: default; game HUD, strong status colors, bordered cards, reward effects.
- Clean productivity: minimal, compact, low-motion, high readability.
- Cyber dashboard: neon accents, high contrast, subtle ambient background motion.

Motion:

- Focus transitions.
- Submission spinner.
- Git Export spinner.
- Level-up/Achievement reveal.
- Low-intensity cyber ambient motion only.

Avoid continuous noisy animation in RPG or clean themes.

## Acceptance Criteria

- All screens respect selected Theme.
- Dashboard can cycle Theme and persist it.
- Async actions show visible loading state.
- Cyber Theme has subtle ambient motion.
- Clean Theme has no ambient motion.
- Tests cover Theme selection/persistence where practical.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
