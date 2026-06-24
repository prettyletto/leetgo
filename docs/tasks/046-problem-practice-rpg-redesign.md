# Task 046: Problem Detail and Practice Log RPG Redesign

## Goal

Restyle Problem Detail as skill tile inspection and Practice Log as a readable learning history.

## Context

Read:

- `docs/product/interface-styling.md`
- `docs/product/learning-system.md`
- Task 045: Roadmap and Stage RPG Redesign

Problem Detail section priority changes by Status.

## Dependencies

- Task 041: Terminal Palette and Symbol Set
- Task 042: Shared View Components

## Likely Files

- `internal/tui/problem_detail_screen.go`
- `internal/tui/solve_log_screen.go`
- related tests

## Implementation Notes

Problem Detail RPG sections:

- large status tile.
- Problem Brief as training notes.
- Requires.
- Unlocks.
- Builds Toward.
- action row.
- compact Practice Log.

Priority by Status:

- Before Start: Problem Brief prominent.
- InProgress: actions and file paths prominent.
- Verified: Submit and Manual Solve dominant.
- Solved: Practice Log, Solve Duration, unlock impact, and Review opportunities prominent.
- Locked: Blockers and recommended prerequisite prominent.

Practice Log:

- chronological.
- status symbols for Attempts and Submission results.
- Rich and Plain Symbols supported.
- capped by default in Problem Detail.

## Acceptance Criteria

- Problem Detail uses status-aware section hierarchy.
- Verified Problem makes Submit primary and Manual Solve secondary.
- Locked Problem shows Blockers as gates and suggests prerequisite action.
- Manual Solve displays completion plus provenance marker.
- Accepted Solve displays trusted completion marker.
- Practice Log uses Practice Log wording and readable entries.
- Submit-anyway confirmation remains clear and styled as a serious warning.
- Tests cover key Status presentations.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
