# Task 012: Deep Navigation and Action Wiring Fixes

## Goal

Make the Dashboard-first screen hierarchy behave as the primary TUI flow, without falling back to the legacy Problem list for Start/Continue actions.

## Why This Exists

Tasks 007-010 added the Dashboard, Roadmap Detail, Stage Detail, and Problem Detail screens, but several actions still bypass or lose state in the new hierarchy:

- Dashboard `enter` on Start/Continue currently opens the legacy list instead of the focused Problem Detail.
- Problem Detail back navigation emits `ScreenStageDetail` without the current Stage ID, so it can return to an empty/default Stage.
- Root navigation does not refresh `activeRoadmap` from config after Onboarding changes `cfg.Roadmap`.

## Dependencies

- Task 007: Dashboard MVP
- Task 008: Roadmap Detail Screen
- Task 009: Stage Detail Screen
- Task 010: Problem Detail Screen

## Likely Files

- `internal/tui/root.go`
- `internal/tui/screen.go`
- `internal/tui/dashboard_screen.go`
- `internal/tui/roadmap_detail_screen.go`
- `internal/tui/stage_detail_screen.go`
- `internal/tui/problem_detail_screen.go`
- `internal/tui/*_test.go`

## Implementation Notes

- Dashboard Start/Continue Next Actions should navigate to `ScreenProblemDetail` with `ProblemID` set.
- Problem Detail should know the Problem's Stage and include it when returning to `ScreenStageDetail`.
- Root navigation should reload or update `activeRoadmap` whenever `cfg.Roadmap` may have changed, especially after Onboarding completion.
- Keep `ScreenLegacyList` available behind the existing `l` key for now; this task should not delete legacy behavior.
- Roadmap Detail `enter` may continue opening Stage Detail, but the footer and tests should describe what actually happens.

## Acceptance Criteria

- Dashboard `enter` on a Start action opens Problem Detail for that Problem.
- Dashboard `enter` on a Continue action opens Problem Detail for that Problem.
- Problem Detail `esc` / `backspace` returns to the correct Stage Detail for the current Problem.
- Completing Onboarding with a non-default Roadmap opens Dashboard for the selected Roadmap, not the initially loaded Roadmap.
- Existing legacy list navigation remains available via `l`.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
