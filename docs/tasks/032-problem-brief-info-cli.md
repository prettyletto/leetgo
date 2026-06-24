# Task 032: Problem Brief and `leetgo info`

## Goal

Expose Problem Brief, progression context, and Practice Log summary through `leetgo info`.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/tasks/030-catalog-learning-metadata.md`

`leetgo info` should mirror Problem Detail in CLI form.

## Dependencies

- Task 030: Catalog Learning Metadata
- Task 029: Solve Provenance and Practice Log Foundation, if Practice Log summary is included
- Problem Manifest resolution from `docs/tasks/020-problem-manifest.md`

## Likely Files

- `cmd/leetgo/`
- `internal/workspace/`
- `internal/catalog/`
- `internal/recommendation/` or a read-only progression helper

## Implementation Notes

When run inside a Problem workspace, `leetgo info` resolves the Problem from the Problem Manifest.

When run outside a Problem workspace, it accepts a Problem ID or slug.

Suggested output sections:

- Problem.
- Status.
- Brief.
- Progression.
- Practice.
- Actions.

Include:

- Summary.
- Practice Focus.
- Why now.
- Unlock Impact.
- Problem Time Estimate.
- Prerequisites.
- Blockers.
- Unlocks.
- Builds toward.
- Latest Practice Log summary when available.

Do not reveal hidden hints by default.

## Acceptance Criteria

- `leetgo info` works inside a generated Problem workspace.
- `leetgo info <id-or-slug>` works outside a Problem workspace.
- Locked Problems show Blockers and recommended first action.
- Available Problems show why they are useful and what they unlock.
- Verified Problems explain Submission or Manual Solve requirement.
- Solved Problems show Accepted vs Manual provenance when known.
- Output uses `Problem Brief`, `Practice Focus`, `Practice Log`, and other domain terms correctly.
- Tests cover inside-workspace and explicit-argument resolution.

## Verification

Run:

```bash
gofmt -w cmd internal
go test ./...
```
