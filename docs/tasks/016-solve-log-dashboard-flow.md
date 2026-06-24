# Task 016: Practice Log Dashboard Flow

## Renaming Note

This task was originally written for Solve Logs. User-facing full-history surfaces should now use `Practice Log`. Existing persistence names may remain if renaming them would make the implementation larger than necessary.

## Goal

Make Practice Logs work from the Dashboard and Problem Detail instead of showing placeholder notifications.

## Why This Exists

The Dashboard currently exposes Practice Log actions, but both paths are placeholders:

- `enter` on the Practice Log Next Action should open Practice Log history.
- Pressing `s` on Dashboard should open Practice Log history.

Problem Detail already records Submission history and renders latest entries for one Problem, and the CLI may already have `leetgo solve-log [limit]`. The MVP needs an in-TUI way to inspect recent Practice Log entries so the Dashboard action is not dead.

## Dependencies

- Task 003: Screen Shell Architecture
- Task 006: Next Action Calculator
- Task 007: Dashboard MVP
- Task 010: Problem Detail Screen
- Existing solve log persistence from SQLite migration `003_solve_logs.sql`

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

- Add a `PracticeLogDetail` or `PracticeLogScreen` to the screen architecture.
- Dashboard `s` should navigate to the Practice Log screen.
- Dashboard `enter` on the Practice Log action should navigate to the same Practice Log screen.
- The Practice Log screen should show recent Practice Log entries with newest first.
- Each row should include Problem ID or Slug, Language, Status, passed/total tests when available, runtime/memory when available, and submitted time.
- Empty state should say that there are no Practice Log entries yet and explain that local tests, Submissions, and Solve actions create entries.
- `esc` / `backspace` should return to Dashboard.
- Keep Problem Detail's per-Problem latest Practice Log entries working.
- If Practice Log entries are not being recorded from Submissions, fix that path too.
- Do not invent notes/editing unless it is already supported; this task is read-only inspection plus recording correctness.

## Acceptance Criteria

- Dashboard `s` opens a Practice Log screen, not a placeholder notification.
- Dashboard `enter` on the Practice Log action opens the Practice Log screen, not a placeholder notification.
- Practice Log screen shows recent entries newest first.
- Practice Log screen has a useful empty state.
- Problem Detail still shows latest Practice Log entries for the current Problem.
- Accepted and rejected Submissions record entries with enough data to render status and test counts.
- Tests cover Dashboard navigation to Practice Log screen from `s` and from the Inspect Next Action.
- Tests cover Practice Log screen rendering for empty and populated stores.
- Tests cover newest-first ordering.

## Verification

Run:

```bash
gofmt -w internal/tui internal/store cmd/leetgo
go vet ./...
go test -count=1 ./...
```
