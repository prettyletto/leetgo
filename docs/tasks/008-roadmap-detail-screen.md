# Task 008: Roadmap Detail Screen

## Goal

Implement Roadmap Detail with Unlock Path and Stage progress.

## Dependencies

- Task 003: Screen Shell Architecture
- Task 005: Theme System Foundation

## Likely Files

- `internal/tui/screens/roadmap_detail.go`
- `internal/tui/views/graph.go`
- `internal/tui/model.go`

## Implementation Notes

Roadmap Detail content:

- Roadmap title, tagline, and promise.
- Stage progress list.
- Unlock Path visualization.
- Available Problems and blockers.

Keybindings:

- `enter`: open focused Stage or Problem.
- `j/k`: move focus.
- `g`: toggle graph/list representation if both exist.
- `esc` / `backspace`: Dashboard.

## Acceptance Criteria

- Dashboard can navigate to Roadmap Detail.
- Roadmap Detail displays all Stages.
- Roadmap Detail exposes blockers for locked Problems.
- Existing graph/unlock-path behavior is preserved or improved.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
