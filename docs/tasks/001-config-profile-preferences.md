# Task 001: Config Profile and Preferences

## Goal

Extend `config.toml` to support Profile, Practice Preferences, Theme, and Git Export backup settings.

## Context

Profile and Practice Preferences live in config, not SQLite. See `docs/adr/0008-profile-preferences-in-config.md` and `docs/product/config-schema.md`.

## Likely Files

- `internal/config/config.go`
- `internal/config/config_test.go`
- `cmd/leetgo/main.go`
- `cmd/leetgo/main_test.go`
- `CONTEXT.md` only if terms change

## Implementation Notes

Add config fields:

- `OnboardingComplete bool` as `onboarding_complete`
- `DisplayName string` as `display_name`
- `Theme string` as `theme`
- `GitExportEnabled bool` as `git_export_enabled`
- `GitExportRepo string` as `git_export_repo`

Keep existing fields:

- `Workspace`
- `Editor`
- `Language`
- `Roadmap`

Default values:

- `OnboardingComplete`: false
- `DisplayName`: empty
- `Theme`: `rpg-skill-tree`
- `GitExportEnabled`: false
- `GitExportRepo`: empty
- existing defaults stay as-is

Validation rules:

- Workspace required.
- Language must be supported.
- Roadmap must exist.
- Theme must be one of `rpg-skill-tree`, `clean-productivity`, `cyber-dashboard`.
- If `OnboardingComplete` is true, Display Name is required.
- If `GitExportEnabled` is true, Git Export repo is required and should be a valid Git repository.

Update `leetgo config set` to support:

- `display-name`
- `theme`
- `git-export-enabled`
- `git-export-repo`

## Acceptance Criteria

- Existing config files without new fields still load.
- New default config includes sane defaults for all fields.
- Invalid Theme fails validation.
- Completed Onboarding without Display Name fails validation.
- Git Export enabled without repo fails validation.
- `leetgo config set theme cyber-dashboard` persists correctly.

## Verification

Run:

```bash
gofmt -w internal/config cmd/leetgo
go test ./...
```
