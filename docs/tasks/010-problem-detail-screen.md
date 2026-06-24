# Task 010: Problem Detail Screen

## Goal

Implement Problem Detail as the owner of Start, Test, Submit, Mark Solved, and Practice Log history actions.

## Dependencies

- Task 009: Stage Detail Screen

## Likely Files

- `internal/tui/screens/problem_detail.go`
- `internal/tui/model.go`
- existing workspace/generator/store/leetcode code

## Implementation Notes

Problem Detail content:

- Problem ID/title.
- Difficulty, Category, Stage.
- Status.
- Prerequisites and blockers.
- Generated file paths when available.
- Latest Practice Log entries for the Problem.

Actions:

- Start.
- Open editor.
- Run local TestSuite.
- Submit to LeetCode.
- Mark Solved manually.

Keybindings:

- `enter`: primary action based on Status.
- `o`: open editor.
- `x`: run local TestSuite.
- `s`: submit.
- `m`: mark Solved.
- `esc` / `backspace`: Stage Detail.

## Acceptance Criteria

- Stage Detail can open Problem Detail.
- Available Problem can be started.
- InProgress/Solved Problem can run local TestSuite.
- Submission result writes Practice Log.
- Accepted Submission marks Problem Solved and updates stats.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
