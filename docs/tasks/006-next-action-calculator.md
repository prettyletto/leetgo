# Task 006: Next Action Calculator

## Goal

Create a reusable domain/service layer that ranks Next Actions for the Dashboard.

## Context

Dashboard is action-first. Next Actions must be computed outside rendering code.

## Dependencies

- Existing roadmap/catalog/store behavior

## Likely Files

- new package `internal/recommendation/` or `internal/dashboard/`
- tests next to new package
- possibly `internal/analytics/`

## Implementation Notes

Define a `NextAction` type with fields like:

- ID
- Kind: Continue, Start, Review, Export, Inspect
- ProblemID optional
- Title
- Reason
- Priority
- Stage
- Category

Ranking:

1. InProgress Problems.
2. Earliest Available Problems in selected Roadmap/Stage order.
3. Weakness-targeting review actions.
4. Maintenance actions such as Git Export or Practice Log review.

Keep weakness/export actions simple if needed; first version can focus on InProgress and Available Problems.

## Acceptance Criteria

- InProgress Problem ranks before Available Problem.
- Available Problems follow Roadmap order.
- Locked Problems do not appear as Start actions.
- Tests cover empty progress, active progress, and solved prerequisites.

## Verification

Run:

```bash
gofmt -w internal
go test ./...
```
