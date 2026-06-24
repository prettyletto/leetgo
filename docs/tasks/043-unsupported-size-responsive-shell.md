# Task 043: Unsupported Size and Responsive Shell

## Goal

Add global Unsupported Size behavior and central responsive layout bands.

## Context

Read:

- `CONTEXT.md`
- `docs/product/interface-styling.md`
- Task 042: Shared View Components

Minimum supported TUI size is 60 columns by 18 rows.

## Dependencies

- Task 042: Shared View Components

## Likely Files

- `internal/tui/root.go`
- `internal/tui/screen.go`
- `internal/tui/views/`
- screen tests under `internal/tui/`

## Implementation Notes

Layout bands:

- Wide.
- Medium.
- Narrow.
- Unsupported Size below 60x18.

Global rule:

- If terminal is below 60x18, render Unsupported Size instead of the active screen.
- Allow `q` to quit.
- Do not lose active screen state when terminal is resized back above minimum.

Unsupported Size copy should include:

- current size.
- minimum size.
- CLI alternatives: `leetgo next`, `leetgo info .`, `leetgo test .`.

## Acceptance Criteria

- Below 60x18, root renders Unsupported Size.
- At 60x18 and above, root renders the active screen.
- Resizing from unsupported back to supported restores active screen rendering.
- `q` still quits from Unsupported Size.
- Tests cover width below minimum, height below minimum, and exact minimum.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
