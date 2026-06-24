# Task 028: Solved-Gated Progression

## Goal

Change progression so dependent Problems unlock only when prerequisites are Solved, not merely Verified.

## Context

Read:

- `CONTEXT.md`
- `docs/adr/0010-local-first-verification-and-reward-events.md`
- `docs/product/learning-system.md`

The new rule is: `Verified` is local confidence and can earn local XP, but does not unlock dependents. `Solved` unlocks dependents. Solved can come from Accepted Solve or Manual Solve.

## Dependencies

- Existing Status model.
- Existing roadmap prerequisite checks.
- Existing Reward Event behavior.

## Likely Files

- `internal/roadmap/`
- `internal/store/`
- `internal/recommendation/`
- `internal/tui/`
- tests next to changed packages

## Implementation Notes

- Audit every prerequisite/unlock check.
- Ensure Verified prerequisites remain Blockers.
- Ensure Manual Solve and Accepted Solve both satisfy prerequisites.
- Do not remove Verified. It remains a visible Status.
- Do not award XP for Manual Solve.
- Preserve idempotent Reward Events.

## Acceptance Criteria

- A Problem with a Verified prerequisite remains Locked.
- A Problem with a Manual Solve prerequisite becomes Available.
- A Problem with an Accepted Solve prerequisite becomes Available.
- Dashboard and Roadmap/Stage blocker summaries identify Verified prerequisites as Blockers.
- Existing solved progress still works across Roadmaps.
- Tests cover Verified, Manual Solve, and Accepted Solve prerequisite cases.

## Verification

Run:

```bash
gofmt -w internal
go test ./...
```
