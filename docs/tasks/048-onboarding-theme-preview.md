# Task 048: Onboarding Theme Preview and Symbol Fallback

## Goal

Add Theme preview and Rich/Plain Symbol fallback control to Onboarding Theme Selection.

## Context

Read:

- `docs/product/interface-styling.md`
- `docs/product/onboarding-dashboard.md`
- Task 040: Appearance Config
- Task 041: Terminal Palette and Symbol Set

Do not add a full Appearance screen.

## Dependencies

- Task 040: Appearance Config
- Task 041: Terminal Palette and Symbol Set

## Likely Files

- `internal/tui/onboarding_screen.go`
- `internal/tui/onboarding_screen_test.go`
- `internal/config/`

## Implementation Notes

Theme preview should show:

- sample Next Action tile.
- sample Status labels.
- sample progress bar.
- sample keytip footer.
- symbol rendering preview.

Symbol preview copy:

```text
Can you see these symbols?
✓ ⚠ 🔒 ★ ↻
```

Key behavior:

- `p`: toggle Plain/Rich Symbols during Theme Selection.
- Persist selected symbol mode.
- Do not add a required extra Onboarding step.

## Acceptance Criteria

- Theme Selection shows a preview for each Theme.
- Symbol preview is visible.
- `p` toggles Rich/Plain Symbols.
- Symbol mode persists to config when Onboarding completes.
- Theme selection behavior remains unchanged otherwise.
- Tests cover preview rendering and symbol-mode toggle.

## Verification

Run:

```bash
gofmt -w internal/tui internal/config
go test ./...
```
