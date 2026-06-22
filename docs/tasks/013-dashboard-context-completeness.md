# Task 013: Dashboard Context Completeness

## Goal

Bring Dashboard HUD and Roadmap context up to the product spec so the default post-Onboarding screen is useful without opening legacy views.

## Why This Exists

The Dashboard MVP renders the core layout, Display Name, basic stats, Next Actions, and Roadmap context. The product spec still calls for additional context that is currently missing or incomplete:

- XP progress to next Level, not just total XP.
- Latest Achievement in the Profile/game HUD when one exists.
- Next blocker summary in the Roadmap/Stage context rail.
- Maintenance actions should reflect config: Git Export should not be presented as ready when Git Export backup is disabled or has no repo configured.

## Dependencies

- Task 007: Dashboard MVP
- Task 006: Next Action Calculator

## Likely Files

- `internal/tui/dashboard_screen.go`
- `internal/recommendation/recommendation.go`
- `internal/store/store.go`
- `internal/tui/views/stats.go`
- `internal/tui/dashboard_screen_test.go`
- `internal/recommendation/recommendation_test.go`

## Implementation Notes

- Reuse existing XP helpers such as `store.LevelToXP` rather than inventing new Level math.
- Use existing Achievement storage and labels where possible.
- The blocker summary should be short: show the next locked Problem and the unsolved prerequisite(s), or state that no current blocker exists.
- If Git Export backup is disabled, either omit the Git Export maintenance action or phrase it as a setup action that sends the user to an explicit setup path. Do not imply a backup can run when no repo is configured.
- Keep narrow layout focused on Next Actions and footer; do not force all HUD details into narrow terminals.

## Acceptance Criteria

- Dashboard HUD shows Level, XP progress to next Level, Streak, solved count, current Roadmap, and Latest Achievement when available.
- Dashboard Roadmap context shows selected Roadmap, current Stage, Stage completion, and a blocker summary.
- Maintenance Next Actions respect `git_export_enabled` and `git_export_repo`.
- Tests cover XP progress display, Latest Achievement display, blocker summary, and Git Export action behavior for enabled/disabled config.

## Verification

Run:

```bash
gofmt -w internal/tui internal/recommendation
go test ./...
```
