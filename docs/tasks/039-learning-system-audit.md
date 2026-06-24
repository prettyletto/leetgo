# Task 039: Learning System Audit

## Goal

Audit the implemented learning-system slice for product consistency, domain terminology, UX clarity, and test coverage.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/adr/0010-local-first-verification-and-reward-events.md`
- Tasks 028 through 038

This task should be done after the implementation tasks are merged or staged together.

## Dependencies

- Tasks 028 through 038, or the subset being audited.

## Likely Files

- All changed files from tasks 028 through 038.
- `docs/product/learning-system.md`
- `docs/tasks/README.md`
- `CONTEXT.md`

## Audit Checklist

Progression:

- Verified does not unlock dependents.
- Manual Solve unlocks dependents and awards no XP.
- Accepted Solve unlocks dependents and awards eligible XP.
- Verified prerequisites appear as Blockers.

Recommendation:

- Dashboard has one primary Next Action.
- Recommendation Reasons are visible and understandable.
- No ranking scores leak into UI or CLI.
- Difficulty spikes are avoided when easier related Problems are Available.
- User can start any Available Problem manually.
- Locked Problems cannot be started in guided Roadmap flow.

UX surfaces:

- Dashboard, Roadmap Detail, Stage Detail, Problem Detail, `leetgo info`, and `leetgo next` use consistent wording.
- Also Available and Coming Soon are distinct.
- Problem Brief is visible without showing hints by default.
- Practice Log replaces Solve Log in user-facing full-history contexts.

Review:

- Review is an action, not a Status.
- Review XP is idempotent per Review Cycle.
- Review Cycles do not block Roadmap Completion.

Roadmap Completion:

- Manual Solves count.
- Verified Problems do not count.
- Active Review Cycles are follow-up work, not blockers.

Docs:

- New behavior matches `CONTEXT.md` terms.
- ADR `0010` still matches implementation.
- Product spec and tasks do not contradict each other.

## Acceptance Criteria

- Produce a written audit report in `docs/audits/learning-system-audit.md`.
- Report findings ordered by severity.
- Each finding includes file/line references where possible.
- Report explicitly states residual risks and missing tests.
- Any terminology drift is either fixed or listed as a finding.
- `go test ./...` result is recorded in the audit.

## Verification

Run:

```bash
go test ./...
```
