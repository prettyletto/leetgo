# Task 042: Shared View Components

## Goal

Create small shared TUI view components for consistent styling across screens.

## Context

Read:

- `docs/product/interface-styling.md`
- Task 041: Terminal Palette and Symbol Set

This task should create a small view layer, not a large framework.

## Dependencies

- Task 041: Terminal Palette and Symbol Set

## Likely Files

- `internal/tui/views/`
- `internal/tui/theme.go`
- tests under `internal/tui/views/`

## Implementation Notes

Add simple helpers/components for:

- `Panel`
- `PixelFrame`
- `StatusPill`
- `ProgressBar`
- `KeytipFooter`
- `UnsupportedSize`

Keep APIs small and composable.

`PixelFrame` should be blocky and terminal-native for RPG Theme.

`ProgressBar` should support segmented XP/progress rendering.

`KeytipFooter` should produce consistent footer formatting.

## Acceptance Criteria

- Components render deterministically in tests.
- PixelFrame differs visually from rounded/default panels.
- StatusPill uses SymbolSet and Terminal Palette roles.
- ProgressBar has empty, partial, and full test coverage.
- KeytipFooter supports compact output for narrow layouts.
- UnsupportedSize renders current/minimum size and CLI alternatives.

## Verification

Run:

```bash
gofmt -w internal/tui/views internal/tui
go test ./...
```
