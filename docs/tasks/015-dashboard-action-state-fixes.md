# Task 015: Dashboard Action State Fixes

## Goal

Fix the remaining Dashboard action-state issues found after Tasks 012-014 so every visible Next Action behaves predictably and no hidden action can be activated.

## Why This Exists

The audit after Tasks 012-014 found two concrete problems:

- Git Export appears as a runnable `Git Export Backup` action when enabled, but pressing `enter` navigates to the legacy Problem list. This is not a real Git Export flow and not an explicit non-ready/setup state.
- Dashboard renders only the first five Next Actions, but focus navigation cycles through all actions. This allows `enter` to activate an invisible action.

## Dependencies

- Task 012: Deep Navigation and Action Wiring Fixes
- Task 013: Dashboard Context Completeness
- Task 014: Theme Motion and Async Polish

## Likely Files

- `internal/tui/dashboard_screen.go`
- `internal/tui/dashboard_screen_test.go`
- `internal/recommendation/recommendation.go` if action wording needs to move closer to the calculator

## Implementation Notes

### Git Export Action

Choose the smallest correct fix:

- If implementing real Dashboard Git Export is small, run the existing Git Export flow asynchronously and show a spinner/loading state.
- If real Dashboard Git Export is not small, keep it as an explicit non-ready/setup state.

For the non-ready/setup state:

- Do not navigate to `ScreenLegacyList`.
- Do not label it as if backup will run immediately.
- Show a clear notification such as `Git Export from Dashboard is not ready yet. Use leetgo git-export <repo-dir> --commit.`
- The card title/reason should make the state clear, for example `Git Export Setup` or `Git Export available from CLI`.

Git Export action visibility rules:

- When `git_export_enabled` is false, omit the action.
- When `git_export_enabled` is true but `git_export_repo` is empty, show setup/non-ready wording or omit the action. Do not present it as runnable backup.
- When both are set, either run a real backup or show the explicit CLI/setup notification.

### Visible Focus Window

Dashboard focus must stay aligned with rendered actions.

Accept either approach:

- Limit focus navigation to the rendered actions only.
- Or implement scrolling/pagination so the focused action is always visible.

Do not allow `enter` to activate an action that is not visible on screen.

## Acceptance Criteria

- Pressing `enter` on Git Export never navigates to the legacy Problem list.
- Git Export is either a real async Dashboard action with visible loading state or an explicit non-ready/setup action with clear wording.
- Git Export does not appear as a runnable backup when `git_export_enabled` is false or `git_export_repo` is empty.
- Dashboard focus cannot move to an invisible Next Action.
- `enter` only activates a visible focused action.
- Tests cover Git Export disabled, enabled-without-repo, enabled-with-repo, and activation behavior.
- Tests cover more than five actions and prove invisible actions cannot be focused or activated.

## Verification

Run:

```bash
gofmt -w internal/tui internal/recommendation
go vet ./...
go test -count=1 ./...
```
