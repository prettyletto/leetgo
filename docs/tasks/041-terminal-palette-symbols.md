# Task 041: Terminal Palette and Symbol Set

## Goal

Create the shared Terminal Palette and SymbolSet foundation used by TUI and CLI styling.

## Context

Read:

- `CONTEXT.md`
- `docs/product/interface-styling.md`
- Task 040: Appearance Config

Themes should express identity through palette roles, symbols, frames, copy labels, and motion.

## Dependencies

- Task 040: Appearance Config
- Existing `internal/tui/theme.go`

## Likely Files

- `internal/tui/theme.go`
- `internal/tui/views/`
- possibly a new package if CLI also needs non-TUI rendering helpers

## Implementation Notes

Terminal Palette roles:

- Primary
- Success
- Warning
- Danger
- Muted
- Border
- XP
- Review

Symbol modes:

- Rich Symbols.
- Plain Symbols.

Status symbol helpers should cover:

- Accepted Solve.
- Verified.
- InProgress.
- Available.
- Locked.
- Review.
- Manual Solve.
- XP/Level.

Rules:

- Verified uses warning/action-needed styling.
- Locked uses muted/border styling, not Danger.
- Manual Solve uses completion styling plus provenance marker.
- Review uses Review role, not Danger.

## Acceptance Criteria

- Themes expose all Terminal Palette roles.
- Rich and Plain SymbolSets exist.
- SymbolSet selection follows config.
- Status rendering helpers can return icon+label for all Status/provenance cases.
- Tests cover Rich and Plain status labels.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
