# Task 016: Solve Log Dashboard Flow

## Goal

Make Solve Logs work from the Dashboard and Problem Detail instead of showing placeholder notifications.

## Why This Exists

The Dashboard currently exposes Solve Log actions, but both paths are placeholders:

- `enter` on the `Review Solve Log` Next Action shows `Solve Log view not yet implemented from Dashboard.`
- Pressing `s` on Dashboard shows `Solve Log detail not yet implemented.`

Problem Detail already records Solve Logs from Submissions and renders latest logs for one Problem, and the CLI already has `leetgo solve-log [limit]`. The MVP needs an in-TUI way to inspect recent Solve Logs so the Dashboard action is not dead.

## Dependencies

- Task 003: Screen Shell Architecture
- Task 006: Next Action Calculator
- Task 007: Dashboard MVP
- Task 010: Problem Detail Screen
- Solve Log persistence from SQLite migration `003_solve_logs.sql`

## Likely Files

- `internal/tui/screen.go`
- `internal/tui/root.go`
- `internal/tui/dashboard_screen.go`
- `internal/tui/problem_detail_screen.go`
- `internal/tui/solve_log_screen.go` or similar new screen
- `internal/store/sqlite.go`
- `internal/tui/*_test.go`
- `cmd/leetgo/main.go` only if CLI behavior is also broken

## Implementation Notes

- Add a `SolveLogDetail` or `SolveLogScreen` to the screen architecture.
- Dashboard `s` should navigate to the Solve Log screen.
- Dashboard `enter` on `KindInspect` / `Review Solve Log` should navigate to the same Solve Log screen.
- The Solve Log screen should show recent Solve Logs with newest first.
- Each row should include Problem ID or Slug, Language, Status, passed/total tests when available, runtime/memory when available, and submitted time.
- Empty state should say that there are no Solve Logs yet and explain that Submissions create Solve Logs.
- `esc` / `backspace` should return to Dashboard.
- Keep Problem Detail's per-Problem latest Solve Logs working.
- If Solve Logs are not being recorded from Submissions, fix that path too.
- Do not invent notes/editing unless it is already supported; this task is read-only inspection plus recording correctness.

## Acceptance Criteria

- Dashboard `s` opens a Solve Log screen, not a placeholder notification.
- Dashboard `enter` on `Review Solve Log` opens the Solve Log screen, not a placeholder notification.
- Solve Log screen shows recent Solve Logs newest first.
- Solve Log screen has a useful empty state.
- Problem Detail still shows latest Solve Logs for the current Problem.
- Accepted and rejected Submissions record Solve Logs with enough data to render status and test counts.
- Tests cover Dashboard navigation to Solve Log screen from `s` and from the Inspect Next Action.
- Tests cover Solve Log screen rendering for empty and populated stores.
- Tests cover newest-first ordering.

## Verification

Run:

```bash
gofmt -w internal/tui internal/store cmd/leetgo
go vet ./...
go test -count=1 ./...
```
