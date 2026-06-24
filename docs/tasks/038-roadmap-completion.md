# Task 038: Roadmap Completion

## Goal

Implement Roadmap Completion as a milestone when every Problem in the selected Roadmap is Solved.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/tasks/036-review-cycles-xp.md`

Manual Solve counts for completion. Active Review Cycles do not block completion.

## Dependencies

- Task 028: Solved-Gated Progression
- Task 029: Solve Provenance and Practice Log Foundation
- Task 030: Catalog Learning Metadata
- Task 036: Review Cycles and Review XP, if Review summary is included

## Likely Files

- `internal/roadmap/`
- `internal/recommendation/`
- `internal/tui/`
- `internal/analytics/`
- `internal/gamification/`

## Implementation Notes

Completion summary includes:

- Problems Solved.
- Accepted Solves.
- Manual Solves.
- Total XP.
- Total Solve Duration.
- Strongest Categories or Practice Focuses.
- Weaknesses to review.
- Active Review Cycles.
- Suggested next Roadmap.

Suggested next Roadmap is catalog-curated first, then ranked by progress and Weaknesses.

If a user switches Roadmaps, Problems solved before selecting the Roadmap count toward completion. The summary should distinguish solved before vs solved during when data is available.

## Acceptance Criteria

- Completion triggers when all Roadmap Problems are Solved.
- Manual Solves count for completion.
- Verified Problems do not count for completion.
- Active Review Cycles do not block completion.
- Completion summary distinguishes Accepted Solves and Manual Solves.
- Completion summary includes total Solve Duration when data exists.
- Completion recommends next Roadmap from catalog metadata.
- Dashboard can surface `ViewRoadmapCompletion` as a Next Action.
- Tests cover all Solved, one Verified remaining, Manual Solve included, and active Review Cycle non-blocking.

## Verification

Run:

```bash
gofmt -w internal
go test ./...
```
