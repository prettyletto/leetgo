# Task 014: Theme Motion and Async Polish

## Goal

Finish the Theme polish and restrained motion behavior promised by Task 011.

## Why This Exists

Task 011 established Theme tokens and basic cyber ambient motion, but the acceptance criteria are only partially satisfied:

- Submission displays a static loading line, not a real animated spinner.
- Git Export has no Dashboard action flow or spinner.
- Level-up and Achievement reveal behavior is still mostly inherited from legacy notification code.
- Theme-specific motion rules need tests so RPG and clean do not accidentally gain ambient animation.

## Dependencies

- Task 005: Theme System Foundation
- Task 007: Dashboard MVP
- Task 010: Problem Detail Screen
- Task 012: Deep Navigation and Action Wiring Fixes
- Task 013: Dashboard Context Completeness

## Likely Files

- `internal/tui/root.go`
- `internal/tui/theme.go`
- `internal/tui/dashboard_screen.go`
- `internal/tui/problem_detail_screen.go`
- `internal/tui/views/notification.go`
- `internal/tui/*_test.go`

## Implementation Notes

- Use Bubble Tea ticks for spinner frames; avoid blocking the UI during async actions.
- Cyber Theme may have low-intensity ambient motion. RPG and clean should not have continuous ambient motion.
- Async actions should visibly enter and leave loading states.
- Keep motion restrained and readable; do not add noisy continuous animation to every screen.
- If Git Export cannot be run safely from Dashboard yet, implement an explicit setup/placeholder state and keep this task scoped to visible async state correctness.

## Acceptance Criteria

- Submission uses an animated spinner while in flight and clears it after a result.
- Git Export action has either a real async spinner-backed flow or an explicit non-ready setup state that does not look like a runnable backup.
- Cyber Theme has subtle ambient motion.
- Clean Theme has no ambient motion.
- RPG Theme has no continuous ambient motion by default.
- Tests cover Theme motion flags and at least one async loading-state transition.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
