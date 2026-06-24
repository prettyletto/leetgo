# Task 002: Roadmap Carousel Metadata

## Goal

Add Roadmap metadata required by Roadmap Carousel cards and validate it during catalog loading.

## Context

Roadmap Selection uses Roadmap Carousel cards. Cards should not hard-code product copy in TUI code. See `docs/product/onboarding-dashboard.md`.

## Likely Files

- `internal/roadmap/types.go`
- `internal/catalog/loader.go`
- `internal/catalog/loader_test.go`
- `internal/catalog/data/roadmaps/*.yaml`
- `docs/product/onboarding-dashboard.md` if metadata changes

## Implementation Notes

Add fields to `roadmap.Roadmap`:

- `Tagline string`
- `Audience string`
- `Promise string`
- `Recommended bool`
- `RoadmapTimeEstimate string`
- `DifficultyMix map[roadmap.Difficulty]int` or a dedicated struct
- `Highlights []string`

Update raw YAML parsing accordingly.

Update all bundled Roadmap YAML files:

- `from-zero-to-hero.yaml`
- `interview-sprint.yaml`
- `hard-mode.yaml`

Validation rules:

- `id`, `title`, `tagline`, `audience`, and `promise` are required.
- Only one bundled Roadmap may be recommended.
- `difficulty_mix` must add to 100 when present.
- `highlights` must contain 2-3 items.
- `roadmap_time_estimate` must be non-empty when present.

## Acceptance Criteria

- `catalog.ListRoadmaps()` returns Roadmaps with complete Carousel metadata.
- Tests fail if two bundled Roadmaps are marked recommended.
- Tests fail if a Roadmap is missing required card copy.
- Existing Problem/Stage loading still works.

## Verification

Run:

```bash
gofmt -w internal/roadmap internal/catalog
go test ./...
```
