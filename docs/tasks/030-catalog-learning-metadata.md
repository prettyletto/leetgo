# Task 030: Catalog Learning Metadata

## Goal

Extend bundled Roadmap catalog metadata to support the learning UX: Roadmap Time Estimate, Problem Brief, Practice Focus, Problem Time Estimate, and next Roadmap recommendations.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/product/onboarding-dashboard.md`

Problem Brief and Practice Focus are Roadmap-specific. Direct unlocks and Blockers are computed from the DAG, not manually authored.

## Dependencies

- Existing catalog loader.
- Existing Roadmap YAML files.

## Likely Files

- `internal/catalog/loader.go`
- `internal/catalog/loader_test.go`
- `internal/catalog/data/roadmaps/*.yaml`
- `internal/roadmap/`

## Implementation Notes

Roadmap metadata should support:

- `promise`
- `audience`
- `roadmap_time_estimate`
- `highlights`
- `recommended`
- `next_roadmaps`

Roadmap Problem metadata should support:

- `practice_focus`
- `problem_time_estimate`
- `summary`
- `why_now`
- `unlock_impact`
- `hints`

Keep global Problem identity minimal:

- LeetCode ID.
- Title.
- Slug.
- Difficulty.

Do not duplicate computed unlock lists in YAML.

## Acceptance Criteria

- Loader parses the new Roadmap metadata fields.
- Loader parses new Roadmap Problem learning fields.
- Validation rejects missing required learning metadata for bundled MVP Roadmaps or reports a clear error.
- Tests cover valid metadata, missing required metadata, and malformed hints/time estimates.
- At least `from-zero-to-hero` has representative metadata for the first Stage sufficient to drive UI development.
- Existing Roadmaps continue to load or fail with intentional, documented validation messages.

## Verification

Run:

```bash
gofmt -w internal/catalog internal/roadmap
go test ./...
```
