# Task 021: Dashboard Action Model Fix

## Goal

Fix Dashboard so `enter` on any Next Action opens Problem Detail only, and fix the rerender collapse bug where Dashboard loses its layout after theme/navigation changes.

## Dependencies

- Task 018: Verified Status

## Likely Files

- `internal/tui/dashboard_screen.go` — fix enter behavior, fix size preservation
- `internal/tui/root.go` — ensure size is reapplied on screen recreation
- `internal/tui/dashboard_screen_test.go`
- `internal/tui/root_test.go`

## Implementation Notes

### Dashboard enter behavior

- `enter` on Start/Continue Next Action: navigate to `ScreenProblemDetail` with `ProblemID`.
- `enter` on Submit Next Action: navigate to `ScreenProblemDetail` with `ProblemID`.
- `enter` on Export: show CLI notification (current behavior is fine).
- `enter` on Inspect: navigate to `ScreenSolveLog` (current behavior is fine).
- Dashboard must never open editor, generate files, or trigger external processes.

### Size preservation

- Root must reapply `tea.WindowSizeMsg` to every newly created screen after navigation and theme change.
- Dashboard must center content when it has width/height information.
- Dashboard must not collapse from wide layout to narrow layout on rerender.

### Arrow key handling

- Dashboard must handle `tea.KeyUp`, `tea.KeyDown`, `tea.KeyShiftUp`, `tea.KeyShiftDown`, `tea.KeyCtrlUp`, `tea.KeyCtrlDown`, and string variants `"up"`, `"down"`, `"j"`, `"k"`.
- Focused action shows explicit `>` cursor marker.

## Acceptance Criteria

- `enter` on Start/Continue/Submit Next Action opens Problem Detail, never opens editor.
- Dashboard preserves wide/medium/narrow layout across theme changes and navigation.
- Arrow keys and j/k both move focus.
- Focused action shows `>` cursor.
- Tests cover enter behavior for each action kind.
- Tests cover size preservation across theme change.
- Tests cover arrow key and j/k focus movement.

## Verification

```bash
gofmt -w internal/tui
go vet ./...
go test -count=1 ./...
```
