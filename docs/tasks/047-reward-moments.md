# Task 047: Reward Moments for TUI and CLI

## Goal

Add Reward Moment presentation for meaningful progress in TUI and CLI.

## Context

Read:

- `CONTEXT.md`
- `docs/product/interface-styling.md`
- `docs/product/learning-system.md`

Reward Moments explain what changed and what the user can do next.

## Dependencies

- Task 041: Terminal Palette and Symbol Set
- Task 042: Shared View Components

## Likely Files

- `cmd/leetgo/main.go`
- `internal/tui/problem_detail_screen.go`
- `internal/tui/roadmap_completion_screen.go`
- `internal/tui/views/`
- tests in `cmd/leetgo` and `internal/tui`

## Implementation Notes

Reward Moment triggers:

- Accepted Solve.
- Manual Solve.
- Review Cycle completion.
- Level-up or Achievement unlock when available.
- Roadmap Completion.

Reward Moment content:

- Problem or Roadmap title.
- XP gained, if any.
- Solve Duration, when available.
- newly unlocked Problems.
- next recommended Problem.
- Recommendation Reason.
- follow-up actions.

CLI behavior:

- Interactive TTY can use spinner/reveal when Motion Preference allows.
- Non-TTY output must remain static and script-friendly.

## Acceptance Criteria

- Accepted Solve shows a structured Reward Moment in CLI and TUI.
- Manual Solve shows no-XP Reward Moment and follow-up Submission recommendation.
- Review Cycle completion shows Review XP and next action.
- Roadmap Completion uses Reward Moment styling.
- Non-TTY CLI output does not animate.
- Tests cover CLI output structure and TUI rendering.

## Verification

Run:

```bash
gofmt -w cmd/leetgo internal/tui internal/tui/views
go test ./...
```
