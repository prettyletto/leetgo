# Task 031: Next Action Learning Ranking

## Goal

Upgrade the Next Action engine to use solved-gated progression, structured Recommendation Reasons, Review actions, and gradual learning ranking.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/tasks/028-solved-gated-progression.md`
- `docs/tasks/030-catalog-learning-metadata.md`

Every Next Action needs a Recommendation Reason. Do not expose ranking scores.

## Dependencies

- Task 028: Solved-Gated Progression
- Task 030: Catalog Learning Metadata, or test doubles with equivalent fields
- Existing `internal/recommendation/`

## Likely Files

- `internal/recommendation/recommendation.go`
- `internal/recommendation/recommendation_test.go`
- `internal/analytics/`

## Implementation Notes

MVP Next Action kinds:

- `Start`
- `Continue`
- `Submit`
- `ManualSolve`
- `Review`
- `ConnectLeetCode`
- `ViewRoadmapCompletion`

MVP reason types:

- `UnlocksDependent`
- `StrengthensPracticeFocus`
- `CompletesVerified`
- `ContinuesInProgress`
- `RepairsWeakness`
- `ValidatesManualSolve`
- `CompletesRoadmap`

Primary ranking:

1. Verified Problem needing Submission or Manual Solve.
2. Recent InProgress Problem.
3. Critical Review if a Weakness blocks current progression.
4. Newly unlocked or best Available Problem.
5. Regular Review.
6. Other Available Problems.
7. Maintenance actions.

Gradual ranking rules:

- Avoid difficulty spikes when related easier Available Problems remain.
- Prefer current Stage as a tie-breaker.
- Prefer direct unlock impact.
- Use indirect unlock impact lightly.
- Prefer Weakness repair when it blocks progression.

## Acceptance Criteria

- Verified Problems produce Submit or ManualSolve actions before new Start actions.
- InProgress Problems rank before new Starts unless a Verified Problem exists.
- Available Problems never include Locked Problems.
- Recommendation Reasons include structured type and rendered text.
- Difficulty spike tests prove easier related Problems are preferred.
- Direct unlock impact affects ranking in tests.
- Review actions can rank above Start when they repair a blocking Weakness.
- Tests cover `-- no score exposed --` by asserting public action output has no ranking score field intended for UI.

## Verification

Run:

```bash
gofmt -w internal/recommendation internal/analytics
go test ./...
```
