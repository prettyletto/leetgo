# Task 003: Screen Shell Architecture

## Goal

Introduce a screen-based TUI architecture while preserving current usable behavior.

## Context

The TUI must move away from one root model with many modes. See `docs/adr/0009-screen-based-tui.md` and `docs/product/tui-screen-spec.md`.

## Likely Files

- `internal/tui/model.go`
- `internal/tui/model_test.go`
- new files under `internal/tui/screens/` or `internal/tui/`
- `internal/tui/views/*`

## Implementation Notes

Introduce a small screen contract, for example:

```go
type Screen interface {
    Init() tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)
    View() string
}
```

Use whatever exact shape fits Bubble Tea best, but keep responsibilities clear:

- Root model owns config, store, global notifications, theme, and navigation.
- Screens own focused UI behavior.
- Root model delegates Update/View to active screen.

Initial screens can be minimal:

- `LegacyProblemListScreen` wrapping current list behavior.
- placeholder Dashboard screen if Onboarding is complete.

Do not implement full Onboarding in this task.

## Acceptance Criteria

- TUI still starts.
- Existing list/start/solve/submit behaviors continue to work through the shell or legacy screen.
- Tests cover navigation/delegation basics.
- Root model no longer needs to grow every future screen as another flat `viewMode` case.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
