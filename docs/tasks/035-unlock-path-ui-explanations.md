# Task 035: Unlock Path UI Explanations

## Goal

Upgrade Dashboard, Roadmap Detail, Stage Detail, and Problem Detail so users can understand recommendations, Blockers, Coming Soon Problems, Unlocks, and Builds toward.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/product/tui-screen-spec.md`

The Roadmap is a DAG but should render as a tree-like Unlock Path with zoom through existing screens.

## Dependencies

- Task 028: Solved-Gated Progression
- Task 030: Catalog Learning Metadata
- Task 031: Next Action Learning Ranking

## Likely Files

- `internal/tui/dashboard_screen.go`
- `internal/tui/roadmap_detail_screen.go`
- `internal/tui/stage_detail_screen.go`
- `internal/tui/problem_detail_screen.go`
- `internal/tui/views/`

## Implementation Notes

Dashboard should show:

- One primary Next Action.
- Recommendation Reason.
- Also Available list.
- Coming Soon list for near-locked Problems with at most two Blockers.

Roadmap Detail should show:

- Stage-grouped Unlock Path.
- Stage progress.
- Active Review markers.
- Coming Soon Problems.

Stage Detail should show:

- Counts by Status.
- Recommended Problem inside the Stage.
- Locked Problems with Blockers.
- Review Problems visually distinct from progression Problems.

Problem Detail should show:

- Problem Brief.
- Problem Time Estimate.
- Status evidence.
- Blockers.
- Unlocks.
- Builds toward.
- Latest Practice Log summary.

Avoid raw graph-first presentation as the default.

## Acceptance Criteria

- Dashboard renders exactly one primary Next Action.
- Dashboard renders Also Available separately from Coming Soon.
- Coming Soon includes near-locked Problems with one or two Blockers, including Verified Blockers.
- Roadmap Detail groups Unlock Path by Stage.
- Stage Detail explains why locked Problems are blocked.
- Problem Detail explains why a Problem is recommended, locked, Verified, manually solved, or accepted.
- UI does not expose ranking scores.
- Tests cover narrow terminal behavior enough to ensure primary actions are not hidden behind context rails.

## Verification

Run:

```bash
gofmt -w internal/tui
go test ./...
```
