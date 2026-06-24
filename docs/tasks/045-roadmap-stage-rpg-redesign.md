# Task 045: Roadmap and Stage RPG Redesign

## Goal

Restyle Roadmap Detail and Stage Detail as readable RPG skill tree zoom levels.

## Context

Read:

- `docs/product/interface-styling.md`
- `docs/product/learning-system.md`
- Task 044: Dashboard RPG Redesign

Roadmap Detail should prioritize readable staged lanes over perfect graph geometry.

## Dependencies

- Task 041: Terminal Palette and Symbol Set
- Task 042: Shared View Components

## Likely Files

- `internal/tui/roadmap_detail_screen.go`
- `internal/tui/stage_detail_screen.go`
- `internal/tui/views/graph.go`
- related tests

## Implementation Notes

Roadmap Detail RPG target:

- World map / skill tree.
- Stage lanes as zones.
- Problems as nodes/tiles.
- Focused node detail with Requires, Blockers, Unlocks, Builds toward.
- Verified Blockers visually stand out.
- Locked branches muted.

Stage Detail RPG target:

- Zone Header.
- Recommended Encounter.
- Encounter Grid.
- Locked Gate.
- Review Shrine.

Do not add mouse interaction.

Do not add search/filter in this task.

## Acceptance Criteria

- Roadmap Detail defaults to staged lanes/tree presentation.
- Focused Problem exposes exact Requires/Blockers/Unlocks.
- Stage Detail groups or visually distinguishes Problems by Status.
- Review Problems are visually distinct from progression Problems.
- Verified Blockers indicate Submit or Manual Solve unlock path.
- Keyboard navigation remains working.
- Tests cover locked, verified blocker, available, solved, and review rendering.

## Verification

Run:

```bash
gofmt -w internal/tui internal/tui/views
go test ./...
```
