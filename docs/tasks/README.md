# Task Index

These tasks are PR-sized slices for the Dashboard-first TUI rebuild. Each task should be implementable by an agent as a focused change with tests.

Read first:

- `CONTEXT.md`
- `docs/adr/0010-local-first-verification-and-reward-events.md`
- `docs/product/onboarding-dashboard.md`
- `docs/product/tui-screen-spec.md`
- `docs/product/config-schema.md`
- `docs/product/tui-rebuild-plan.md`

## Recommended Order

1. [Config Profile and Preferences](./001-config-profile-preferences.md)
2. [Roadmap Carousel Metadata](./002-roadmap-carousel-metadata.md)
3. [Screen Shell Architecture](./003-screen-shell-architecture.md)
4. [Onboarding Flow](./004-onboarding-flow.md)
5. [Theme System Foundation](./005-theme-system-foundation.md)
6. [Next Action Calculator](./006-next-action-calculator.md)
7. [Dashboard MVP](./007-dashboard-mvp.md)
8. [Roadmap Detail Screen](./008-roadmap-detail-screen.md)
9. [Stage Detail Screen](./009-stage-detail-screen.md)
10. [Problem Detail Screen](./010-problem-detail-screen.md)
11. [Theme Polish and Motion](./011-theme-polish-motion.md)
12. [Deep Navigation and Action Wiring Fixes](./012-deep-navigation-action-wiring.md)
13. [Dashboard Context Completeness](./013-dashboard-context-completeness.md)
14. [Theme Motion and Async Polish](./014-theme-motion-async-polish.md)
15. [Dashboard Action State Fixes](./015-dashboard-action-state-fixes.md)
16. [Solve Log Dashboard Flow](./016-solve-log-dashboard-flow.md)
17. [Onboarding Defaults and Rerun](./017-onboarding-defaults-and-rerun.md)
18. [Verified Status Domain Model](./018-verified-status.md)
19. [Reward Events Persistence and XP Idempotency](./019-reward-events.md)
20. [Problem Manifest Generation and CLI Resolution](./020-problem-manifest.md)
21. [Dashboard Action Model Fix](./021-dashboard-action-model.md)
22. [Problem Detail Action Model](./022-problem-detail-action-model.md)
23. [Manual Solve Typed Confirmation](./023-manual-solve-confirmation.md)
24. [Dashboard Stats Verified/Solved Split](./024-dashboard-stats-split.md)
25. [CLI Test and Submit Output](./025-cli-output.md)
26. [Migration for Existing Solved Problems](./026-migration-legacy-solved.md)
27. [Dashboard Next Action Ranking for Verified](./027-dashboard-verified-ranking.md)

## Task Rules

- Keep each task small enough for one PR.
- Preserve current behavior unless the task explicitly replaces it.
- Add or update tests in the same PR.
- Run `gofmt` and `go test ./...` before handing off.
- Do not rewrite unrelated modules.
- Use domain terms from `CONTEXT.md` exactly.

## Definition of Done

- The task's acceptance criteria pass.
- Relevant docs are updated if behavior or terms change.
- `go test ./...` passes.
- The implementation does not introduce hidden config defaults that conflict with Onboarding.
