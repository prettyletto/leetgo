# Task 029: Solve Provenance and Practice Log Foundation

## Goal

Persist and expose the difference between Manual Solve and Accepted Solve, and introduce Practice Log as the user-facing Problem history.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/adr/0010-local-first-verification-and-reward-events.md`

Practice Log replaces user-facing Solve Log language for the full chronological learning history of a Problem.

## Dependencies

- Task 028: Solved-Gated Progression
- Existing solve log or submission persistence
- Existing migrations pattern

## Likely Files

- `internal/store/`
- `internal/store/migrations/`
- `internal/tui/solve_log_screen.go` or renamed equivalent
- `internal/tui/problem_detail_screen.go`
- CLI command handlers under `cmd/leetgo/`

## Implementation Notes

- Add enough persistence to distinguish Manual Solve from Accepted Solve.
- Keep existing historical records readable.
- Rename user-facing labels from Solve Log to Practice Log where the history includes local attempts and manual solves.
- Do not perform a broad package rename unless it is small and safe.
- Manual Solve entries should support an optional note.
- Practice Log should be chronological.

## Acceptance Criteria

- Manual Solve is recorded with provenance and optional note.
- Accepted Solve is recorded with provenance from LeetCode Submission.
- Practice Log can display Start, local Attempt, Submission Attempt, Verified, Manual Solve, and Accepted Solve entries when data exists.
- Existing solve/submission history remains visible after migration.
- User-facing TUI/CLI copy says Practice Log for the full history.
- Tests cover provenance persistence and retrieval.

## Verification

Run:

```bash
gofmt -w internal cmd
go test ./...
```
