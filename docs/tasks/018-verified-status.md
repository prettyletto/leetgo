# Task 018: Add Verified Status to Domain Model

## Goal

Add `Verified` as a new Problem Status between `InProgress` and `Solved`, and update Roadmap unlocking so both `Verified` and `Solved` satisfy prerequisites.

## Dependencies

- None (foundational)

## Likely Files

- `internal/roadmap/roadmap.go` — add `StatusVerified` constant
- `internal/roadmap/graph.go` — update `IsUnlocked` to accept Verified
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
- Update `Graph.IsUnlocked` so that a prerequisite is satisfied if its status is `StatusVerified` or `StatusSolved`.
- Update all status rendering across TUI screens to show `VERIFIED` with a distinct color (use `theme.PrimaryAccent` or similar).
- Update `effectiveStatus` helpers in Roadmap Detail, Stage Detail, and Problem Detail to recognize Verified.
- Update blocker summary to not show Verified Problems as blockers.
- Update export/import to round-trip Verified status.
- Do not change XP logic in this task; that belongs in Task 019.

## Acceptance Criteria

- `StatusVerified` exists as a valid Status constant.
- `Graph.IsUnlocked` returns true when prerequisites are Verified or Solved.
- All TUI screens render Verified with a distinct label and color.
- Verified Problems do not appear as blockers in Dashboard or Problem Detail.
- Export/import round-trips Verified status.
- Tests cover IsUnlocked with Verified prerequisites.
- Tests cover status rendering for Verified.

## Verification

```bash
gofmt -w internal/roadmap internal/store internal/tui internal/recommendation
go vet ./...
go test -count=1 ./...
```
