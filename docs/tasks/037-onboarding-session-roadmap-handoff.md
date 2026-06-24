# Task 037: Onboarding Session and Roadmap Handoff

## Goal

Update Onboarding so users understand the learning loop, optionally connect LeetCode Session, and land on the first Next Action.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/product/onboarding-dashboard.md`
- `docs/product/tui-screen-spec.md`

Roadmap Selection is required. LeetCode Session setup is optional but important because Accepted Solve unlocks progression and XP.

## Dependencies

- Existing Onboarding flow.
- Task 031: Next Action Learning Ranking, for first Next Action handoff.

## Likely Files

- `internal/tui/onboarding_screen.go`
- `internal/tui/roadmap_selection_screen.go`
- `internal/config/`
- `internal/leetcode/`

## Implementation Notes

Onboarding should explain:

`Start a Problem -> pass local tests -> submit to LeetCode or Manual Solve -> unlock next Problems`

Roadmap cards should show:

- Title.
- Promise.
- Audience.
- Size.
- Difficulty mix.
- Roadmap Time Estimate.
- First Stages.
- Highlights.

LeetCode Session step:

- `Connect now`.
- `Skip for now`.

Copy:

`Accepted Submissions unlock Roadmap progress and XP. You can skip and use Manual Solve, but Manual Solve earns no XP.`

Completion screen:

- Show first Next Action.
- Show Recommendation Reason.
- Actions: `Start now`, `Go to Dashboard`.

## Acceptance Criteria

- Roadmap Selection cannot be skipped.
- Default Roadmap is focused and easy to confirm.
- LeetCode Session setup can be skipped.
- Skipping Session does not block Onboarding completion.
- Final Onboarding screen shows first Next Action and Recommendation Reason.
- `Start now` starts the recommended Problem when the action kind is Start.
- Dashboard later reminds about LeetCode Session when Verified Problems are waiting.
- Tests cover Session skip, Session connect path where feasible, and first-action handoff.

## Verification

Run:

```bash
gofmt -w internal/tui internal/config internal/leetcode
go test ./...
```
