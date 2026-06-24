# Task 040: Appearance Config

## Goal

Add appearance preferences needed by the styling phase: `symbol_mode` and `motion_preference`.

## Context

Read:

- `CONTEXT.md`
- `docs/product/interface-styling.md`
- `docs/product/config-schema.md`

Theme already exists. This task adds the supporting preferences for Rich/Plain Symbols and Motion Preference.

## Dependencies

- Existing config loading/saving.

## Likely Files

- `internal/config/config.go`
- `docs/product/config-schema.md`
- CLI config command tests under `cmd/leetgo/`

## Implementation Notes

Config fields:

- `symbol_mode`: `rich` or `plain`, default `rich`.
- `motion_preference`: `normal`, `reduced`, or `off`, default `normal`.

Validation:

- Reject unknown values.
- Preserve existing config behavior and defaults.

CLI config support:

- `leetgo config set symbol-mode plain`
- `leetgo config set motion reduced`

Do not add a full Appearance screen in this task.

## Acceptance Criteria

- New config fields load with defaults.
- Config save/load round-trips both fields.
- Invalid values fail validation with clear errors.
- CLI config setter can update both values.
- Existing config tests still pass.

## Verification

Run:

```bash
gofmt -w internal/config cmd/leetgo
go test ./...
```
