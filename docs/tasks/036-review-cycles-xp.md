# Task 036: Review Cycles and Review XP

## Goal

Implement Review as an MVP learning action with bounded Review Cycles and small, idempotent Review XP.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- existing analytics and gamification docs/code

Review is not a Status and does not block Roadmap Completion.

## Dependencies

- Task 029: Solve Provenance and Practice Log Foundation
- Task 031: Next Action Learning Ranking
- Existing Weakness detection
- Existing Reward Event idempotency

## Likely Files

- `internal/analytics/`
- `internal/gamification/`
- `internal/store/`
- `internal/recommendation/`
- `internal/tui/`

## Implementation Notes

Review Cycle creation reasons:

- Weakness repair.
- High failed-attempt history.
- Manual Solve validation.
- Prerequisite refresh before a dependent.

Review Cycle completion proof:

- Accepted Solve Problem: local TestSuite pass during Review.
- Manual Solve Problem: Accepted Submission preferred; local TestSuite pass may complete a weaker cycle.
- Weakness repair: local TestSuite pass on the recommended Review Problem.

Review XP:

- Small amount.
- Earned at most once per Review Cycle.
- Not awarded for opening a Problem.

Review Cycles are mostly global to the user and Problem, with Roadmap context attached to the reason.

## Acceptance Criteria

- Review Cycle can be created for each MVP reason.
- Review appears as a Next Action when appropriate.
- Review Cycle completion awards Review XP once.
- Repeating the same proof does not award duplicate Review XP.
- Review does not change Problem Status.
- Review does not block Roadmap Completion.
- Manual Solve validation can recommend Accepted Submission as Review.
- Tests cover creation, completion, XP idempotency, and non-blocking completion behavior.

## Verification

Run:

```bash
gofmt -w internal
go test ./...
```
