# Task 018: Add Verified Status to Domain Model

## Superseded Progression Note

Task 028 supersedes the original unlock rule in this task. `Verified` no longer satisfies prerequisites. `Verified` is local confidence and may earn local XP, but only `Solved` unlocks dependent Problems. See `docs/tasks/028-solved-gated-progression.md` and `docs/adr/0010-local-first-verification-and-reward-events.md`.

## Goal

Add `Verified` as a new Problem Status between `InProgress` and `Solved`. Roadmap unlocking must treat `Verified` prerequisites as Blockers until they become Solved.

## Dependencies

- None (foundational)

## Likely Files

- `internal/roadmap/roadmap.go` — add `StatusVerified` constant
- `internal/roadmap/graph.go` — keep Verified prerequisites blocked until Solved
- `internal/store/store.go` — no interface change, Status is a string type
- `internal/store/sqlite.go` — ensure Verified is persisted correctly
- `internal/store/export.go` — handle Verified in export/import
- `internal/tui/problem_detail_screen.go` — render `VERIFIED` status label
- `internal/tui/stage_detail_screen.go` — render Verified status marker
- `internal/tui/roadmap_detail_screen.go` — render Verified status marker
- `internal/tui/dashboard_screen.go` — handle Verified in blocker summary
- `internal/recommendation/recommendation.go` — handle Verified in Next Actions
- `internal/tui/*_test.go`

## Implementation Notes

- Add `StatusVerified roadmap.Status = "verified"` to `internal/roadmap/roadmap.go`.
- Update `Graph.IsUnlocked` so that a prerequisite is satisfied only when its status is `StatusSolved`.
- Update all status rendering across TUI screens to show `VERIFIED` with a distinct color (use `theme.PrimaryAccent` or similar).
- Update `effectiveStatus` helpers in Roadmap Detail, Stage Detail, and Problem Detail to recognize Verified.
- Update blocker summary to show Verified Problems as Blockers that need Submission or Manual Solve.
- Update export/import to round-trip Verified status.
- Do not change XP logic in this task; that belongs in Task 019.

## Acceptance Criteria

- `StatusVerified` exists as a valid Status constant.
- `Graph.IsUnlocked` returns false when prerequisites are Verified and true when prerequisites are Solved.
- All TUI screens render Verified with a distinct label and color.
- Verified Problems appear as Blockers in Dashboard or Problem Detail until they become Solved.
- Export/import round-trips Verified status.
- Tests cover IsUnlocked with Verified prerequisites and Solved prerequisites.
- Tests cover status rendering for Verified.

## Verification

```bash
gofmt -w internal/roadmap internal/store internal/tui internal/recommendation
go vet ./...
go test -count=1 ./...
```
