# Learning System Sprint Plan

This plan groups tasks `028` through `039` into agent-sized sprints. Each sprint should be assigned as one coherent implementation batch. The final audit sprint should be run separately after implementation.

Read first:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/adr/0010-local-first-verification-and-reward-events.md`
- `docs/tasks/README.md`

## Sprint Rules

- A sprint agent should complete all tasks listed in the sprint unless blocked.
- A sprint agent should not start later sprints unless explicitly assigned.
- Each sprint must update tests with the implementation.
- Each sprint must run `gofmt` on changed Go files and `go test ./...` before handoff.
- Each sprint handoff must list completed tasks, changed files, test result, and any follow-up risks.
- Use domain terms from `CONTEXT.md` exactly.
- Preserve unrelated user or agent changes in the worktree.

## Sprint 1: Progression Foundation

Goal: make the domain model honest before any UX depends on it.

Tasks:

- [Task 028: Solved-Gated Progression](./028-solved-gated-progression.md)
- [Task 029: Solve Provenance and Practice Log Foundation](./029-solve-provenance-practice-log.md)
- [Task 034: Submit and Manual Solve Flow](./034-submit-manual-solve-flow.md)

Why these are together:

Progression, solve provenance, Submission, and Manual Solve all touch the same state transition boundary. Splitting them too far apart risks temporary behavior where the app unlocks incorrectly or loses why a Problem was Solved.

Expected result:

- `Verified` no longer unlocks dependents.
- `Manual Solve` unlocks dependents with no XP.
- `Accepted Solve` unlocks dependents and records trusted provenance.
- Practice Log can show the important solve path events.
- `leetgo submit` and Manual Solve follow the new product rules.

Suggested handoff prompt:

```text
Implement Sprint 1 from docs/tasks/learning-system-sprints.md. Complete Tasks 028, 029, and 034 end-to-end with tests. Do not implement later sprints. Before handoff run gofmt and go test ./..., then summarize changed files, behavior changes, and remaining risks.
```

## Sprint 2: Catalog Learning Metadata

Goal: give the app the authored learning content needed by recommendations, TUI, and CLI.

Tasks:

- [Task 030: Catalog Learning Metadata](./030-catalog-learning-metadata.md)

Why this is its own sprint:

Catalog schema and bundled YAML changes are broad and easy to review separately. This sprint creates the data contract that later UX sprints consume.

Expected result:

- Roadmap metadata supports Roadmap Time Estimate and next Roadmaps.
- Roadmap Problem metadata supports Practice Focus, Problem Brief, Problem Time Estimate, and hints.
- Loader validation catches malformed learning metadata.
- At least the first `from-zero-to-hero` Stage has representative authored metadata for UI development.

Suggested handoff prompt:

```text
Implement Sprint 2 from docs/tasks/learning-system-sprints.md. Complete Task 030 with loader validation, tests, and representative bundled metadata. Do not implement recommendation, CLI, or TUI behavior beyond what is necessary to compile. Run gofmt and go test ./... before handoff.
```

## Sprint 3: Recommendation Engine And CLI Guidance

Goal: make the guidance model usable outside the TUI first.

Tasks:

- [Task 031: Next Action Learning Ranking](./031-next-action-learning-ranking.md)
- [Task 033: `leetgo next` CLI](./033-next-cli-command.md)
- [Task 032: Problem Brief and `leetgo info`](./032-problem-brief-info-cli.md)

Why these are together:

`leetgo next` and `leetgo info` are direct consumers of the recommendation and learning metadata model. Building them together proves the domain layer works without waiting on TUI rendering.

Expected result:

- Next Actions have structured Recommendation Reasons.
- Verified, InProgress, Available, Review, and maintenance actions rank correctly.
- `leetgo next` exposes the same primary guidance expected by Dashboard.
- `leetgo info` exposes Problem Brief and progression context.
- Ranking scores are not exposed to users.

Suggested handoff prompt:

```text
Implement Sprint 3 from docs/tasks/learning-system-sprints.md. Complete Tasks 031, 033, and 032 with tests. Focus on recommendation correctness and CLI output. Do not implement TUI polish from Sprint 4. Run gofmt and go test ./... before handoff.
```

## Sprint 4: TUI Learning UX

Goal: make the Dashboard and drill-down screens explain the learning system clearly.

Tasks:

- [Task 035: Unlock Path UI Explanations](./035-unlock-path-ui-explanations.md)
- [Task 037: Onboarding Session and Roadmap Handoff](./037-onboarding-session-roadmap-handoff.md)

Why these are together:

Onboarding, Dashboard, Roadmap Detail, Stage Detail, and Problem Detail are the user's guided learning surface. They should be updated after the recommendation and metadata foundations exist.

Expected result:

- Onboarding explains the learning loop and optional LeetCode Session setup.
- Onboarding ends with the first Next Action.
- Dashboard shows one primary Next Action, Also Available, and Coming Soon.
- Roadmap Detail, Stage Detail, and Problem Detail explain Blockers, Recommendation Reasons, Unlocks, Builds toward, and Problem Briefs.
- Practice Log language is used in user-facing full-history contexts.

Suggested handoff prompt:

```text
Implement Sprint 4 from docs/tasks/learning-system-sprints.md. Complete Tasks 035 and 037 with tests. Focus on TUI clarity and consistency with docs/product/learning-system.md. Do not implement Review XP or Roadmap Completion unless required for compilation. Run gofmt and go test ./... before handoff.
```

## Sprint 5: Review And Completion

Goal: complete the MVP learning loop after progression and guidance work.

Tasks:

- [Task 036: Review Cycles and Review XP](./036-review-cycles-xp.md)
- [Task 038: Roadmap Completion](./038-roadmap-completion.md)

Why these are together:

Review Cycles and Roadmap Completion are both summary/quality layers over the core progression model. They need solve provenance, Practice Log, recommendation, and metadata work to be stable first.

Expected result:

- Review is an action, not a Status.
- Review Cycles can be created, completed, and rewarded once.
- Review XP is small and idempotent.
- Review does not block Roadmap Completion.
- Roadmap Completion summarizes Accepted Solves, Manual Solves, total XP, Solve Duration, Weaknesses, active Review Cycles, and next Roadmap recommendation.

Suggested handoff prompt:

```text
Implement Sprint 5 from docs/tasks/learning-system-sprints.md. Complete Tasks 036 and 038 with tests. Focus on Review Cycle idempotency and Roadmap Completion correctness. Run gofmt and go test ./... before handoff.
```

## Sprint 6: Learning System Audit

Goal: audit the completed implementation against the product spec and domain language.

Tasks:

- [Task 039: Learning System Audit](./039-learning-system-audit.md)

When to run:

Run this after Sprints 1 through 5 are implemented, or after any subset you want reviewed.

Expected result:

- Written audit report at `docs/audits/learning-system-audit.md`.
- Findings ordered by severity.
- File and line references where possible.
- Explicit residual risks and missing tests.
- `go test ./...` result recorded.

Suggested handoff prompt:

```text
Audit the learning-system implementation using Sprint 6 from docs/tasks/learning-system-sprints.md. Complete Task 039 only. Do not fix findings unless explicitly asked. Produce docs/audits/learning-system-audit.md with severity-ordered findings, file/line references, residual risks, missing tests, and go test ./... result.
```

## Dependency Summary

Recommended order:

1. Sprint 1: Progression Foundation.
2. Sprint 2: Catalog Learning Metadata.
3. Sprint 3: Recommendation Engine And CLI Guidance.
4. Sprint 4: TUI Learning UX.
5. Sprint 5: Review And Completion.
6. Sprint 6: Learning System Audit.

Hard dependencies:

- Sprint 3 depends on Sprint 1 and benefits from Sprint 2.
- Sprint 4 depends on Sprints 1, 2, and 3.
- Sprint 5 depends on Sprints 1 and 3, and benefits from Sprint 4.
- Sprint 6 should run last.

Parallelization note:

Sprint 2 can run in parallel with Sprint 1 if the agents coordinate on catalog/roadmap type changes. Prefer sequential execution if you want fewer merge conflicts.
