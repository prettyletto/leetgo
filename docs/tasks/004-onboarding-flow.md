# Task 004: Onboarding Flow

## Goal

Implement first-run Onboarding screens that collect Profile and Practice Preferences.

## Context

Onboarding appears when `onboarding_complete` is false or missing. See `docs/product/tui-screen-spec.md`.

## Dependencies

- Task 001: Config Profile and Preferences
- Task 002: Roadmap Carousel Metadata
- Task 003: Screen Shell Architecture

## Likely Files

- `internal/tui/model.go`
- `internal/tui/screens/*`
- `internal/config/config.go`
- `internal/catalog/loader.go`
- `internal/gitexport/*`

## Implementation Notes

Implement Onboarding steps:

1. Display Name input.
2. Git Export backup opt-in and repo path validation.
3. Workspace and Language confirmation.
4. Roadmap Carousel.
5. Theme Selection.
6. Save config and navigate to Dashboard.

Keep LeetCode Session setup out of Onboarding.

Roadmap Carousel behavior:

- Focus recommended Roadmap first.
- `<` / `>` and left/right rotate focus.
- `enter` confirms.

Theme Selection behavior:

- RPG skill tree focused by default.
- left/right rotate.
- `enter` confirms.

## Acceptance Criteria

- Fresh config opens Onboarding.
- Completing Onboarding writes all required config fields.
- Existing partial config pre-fills workspace/language/roadmap.
- `onboarding_complete` becomes true only after the final step.
- Quitting before completion does not mark Onboarding complete.

## Verification

Run:

```bash
gofmt -w internal/tui internal/config
go test ./...
```
