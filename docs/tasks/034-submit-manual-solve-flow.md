# Task 034: Submit and Manual Solve Flow

## Goal

Align `leetgo submit`, TUI Submission actions, and Manual Solve with solved-gated progression.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/adr/0010-local-first-verification-and-reward-events.md`

Roadmap unlock requires Solved. Verified should lead to Submit or Manual Solve.

## Dependencies

- Task 028: Solved-Gated Progression
- Task 029: Solve Provenance and Practice Log Foundation

## Likely Files

- `cmd/leetgo/`
- `internal/leetcode/`
- `internal/store/`
- `internal/tui/problem_detail_screen.go`
- `internal/tui/dashboard_screen.go`

## Implementation Notes

`leetgo submit` runs local tests by default.

If local tests fail:

- Stop before Submission.
- Record failed Local Attempt.
- Explain fix-first behavior.
- Allow `--skip-tests` override.

If local tests pass:

- Mark or keep Verified.
- Submit to LeetCode.

If LeetCode accepts:

- Mark Accepted Solve.
- Unlock dependents.
- Award eligible Submission XP.
- Show post-solve summary.

If LeetCode rejects:

- Keep Verified.
- Record failed Submission Attempt.
- Recommend continuing editing.

Manual Solve:

- Requires confirmation in TUI and CLI.
- Supports optional note.
- Supports `--yes` for scripts.
- Unlocks dependents.
- Awards no XP.

## Acceptance Criteria

- `leetgo submit` runs local tests before Submission by default.
- `leetgo submit --skip-tests` can submit after local failure.
- Accepted Submission after local failure still marks Accepted Solve and records the mismatch.
- Rejected Submission keeps Verified when local tests passed.
- Manual Solve requires confirmation unless `--yes` is provided.
- Manual Solve records optional note in Practice Log.
- Manual Solve unlocks dependents and awards no XP.
- Tests cover local pass, local fail, skip-tests accepted, rejected, and manual solve confirmation paths.

## Verification

Run:

```bash
gofmt -w cmd internal
go test ./...
```
