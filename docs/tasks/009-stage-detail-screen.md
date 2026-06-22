# Task 009: Stage Detail Screen

## Goal

Implement Stage Detail with focused Problem browsing for one Stage.

## Dependencies

- Task 008: Roadmap Detail Screen

## Likely Files

- `internal/tui/screens/stage_detail.go`
- `internal/tui/views/*`

## Implementation Notes

Stage Detail content:

- Stage title and description.
- Completion count.
- Problems in the Stage.
- Status for each Problem.
- Blockers for locked Problems.

Keybindings:

- `enter`: open Problem Detail.
- `j/k`: move focus.
- `esc` / `backspace`: Roadmap Detail.

## Acceptance Criteria

- Roadmap Detail can open Stage Detail.
- Stage Detail filters Problems to one Stage.
- Available and InProgress Problems are visually emphasized.
- Locked Problems show blockers.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
