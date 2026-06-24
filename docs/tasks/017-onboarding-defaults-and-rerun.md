# Task 017: Onboarding Defaults and Rerun

## Goal

Fix Onboarding startup/default behavior and add an explicit way for a user to run Onboarding again.

## Why This Exists

The current user report is that Onboarding is not initiating correctly and appears to default to `Ada`. The MVP requires fresh users to enter their own Display Name, and existing users need a clear way to rerun Onboarding when they want to change Profile and Practice Preferences.

## Dependencies

- Task 001: Config Profile and Preferences
- Task 003: Screen Shell Architecture
- Task 004: Onboarding Flow
- Task 012: Deep Navigation and Action Wiring Fixes

## Likely Files

- `cmd/leetgo/main.go`
- `internal/config/config.go`
- `internal/tui/root.go`
- `internal/tui/onboarding_screen.go`
- `internal/tui/dashboard_screen.go`
- `internal/tui/*_test.go`
- `internal/config/*_test.go`
- `README.md` if command usage changes

## Implementation Notes

### Fresh Onboarding Defaults

- A fresh config must not default `display_name` to `Ada` or any other sample name.
- If `onboarding_complete` is false, the TUI must open Onboarding even if other config fields have defaults.
- The Display Name step should start blank for a fresh config.
- Existing configs may prefill their saved Display Name when rerunning Onboarding, but there should be an easy way to clear/edit it.
- Tests should not accidentally encode `Ada` as a product default; use test-local names only where needed.

### Rerun Onboarding

Add one explicit user-facing path to rerun Onboarding. Prefer the smallest clear option:

- CLI: `leetgo onboard` or `leetgo onboarding`
- TUI: a Dashboard keybinding such as `o` for Onboarding

If adding only one path, prefer the CLI command because it is discoverable in help/docs and works even if the TUI state is wrong.

Expected behavior for rerun:

- Opens Onboarding with current Profile and Practice Preferences prefilled.
- Does not delete SQLite progress, Practice Logs, XP, Streaks, or Achievements.
- Saving Onboarding updates `config.toml` and returns to Dashboard.
- Quitting Onboarding without completion should not corrupt the existing config.

### Optional Reset Mode

If needed, add a separate explicit reset command such as `leetgo onboard --fresh` that clears Profile/Practice Preference fields before Onboarding. Do not clear progress/history.

## Acceptance Criteria

- Fresh config launches Onboarding, not Dashboard.
- Fresh Onboarding Display Name input is blank, not `Ada`.
- `onboarding_complete = false` always launches Onboarding.
- Completed Onboarding still launches Dashboard afterward.
- User can explicitly rerun Onboarding after completion.
- Rerunning Onboarding preserves SQLite progress/history.
- Rerunning Onboarding can update Display Name, Language, Workspace, Roadmap, Theme, and Git Export preferences.
- Tests cover fresh defaults, false `onboarding_complete`, completed config, and rerun command/path.
- User-facing docs mention how to rerun Onboarding.

## Verification

Run:

```bash
gofmt -w cmd/leetgo internal/config internal/tui
go vet ./...
go test -count=1 ./...
```
