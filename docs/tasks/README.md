# Task Index

These tasks are PR-sized slices for the Dashboard-first TUI rebuild. Each task should be implementable by an agent as a focused change with tests.

Read first:

- `CONTEXT.md`
- `docs/adr/0010-local-first-verification-and-reward-events.md`
- `docs/product/onboarding-dashboard.md`
- `docs/product/tui-screen-spec.md`
- `docs/product/learning-system.md`
- `docs/product/interface-styling.md`
- `docs/product/config-schema.md`
- `docs/product/tui-rebuild-plan.md`
- `docs/tasks/learning-system-sprints.md`
- `docs/tasks/interface-styling-sprints.md`

## Recommended Order

For the learning-system phase, assign tasks by sprint from [Learning System Sprint Plan](./learning-system-sprints.md). The list below remains the full task index.

For the interface-styling phase, assign tasks by sprint from [Interface Styling Sprint Plan](./interface-styling-sprints.md).

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
16. [Practice Log Dashboard Flow](./016-solve-log-dashboard-flow.md)
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
28. [Solved-Gated Progression](./028-solved-gated-progression.md)
29. [Solve Provenance and Practice Log Foundation](./029-solve-provenance-practice-log.md)
30. [Catalog Learning Metadata](./030-catalog-learning-metadata.md)
31. [Next Action Learning Ranking](./031-next-action-learning-ranking.md)
32. [Problem Brief and `leetgo info`](./032-problem-brief-info-cli.md)
33. [`leetgo next` CLI](./033-next-cli-command.md)
34. [Submit and Manual Solve Flow](./034-submit-manual-solve-flow.md)
35. [Unlock Path UI Explanations](./035-unlock-path-ui-explanations.md)
36. [Review Cycles and Review XP](./036-review-cycles-xp.md)
37. [Onboarding Session and Roadmap Handoff](./037-onboarding-session-roadmap-handoff.md)
38. [Roadmap Completion](./038-roadmap-completion.md)
39. [Learning System Audit](./039-learning-system-audit.md)
40. [Appearance Config](./040-appearance-config.md)
41. [Terminal Palette and Symbol Set](./041-terminal-palette-symbols.md)
42. [Shared View Components](./042-shared-view-components.md)
43. [Unsupported Size and Responsive Shell](./043-unsupported-size-responsive-shell.md)
44. [Dashboard RPG Redesign](./044-dashboard-rpg-redesign.md)
45. [Roadmap and Stage RPG Redesign](./045-roadmap-stage-rpg-redesign.md)
46. [Problem Detail and Practice Log RPG Redesign](./046-problem-practice-rpg-redesign.md)
47. [Reward Moments for TUI and CLI](./047-reward-moments.md)
48. [Onboarding Theme Preview and Symbol Fallback](./048-onboarding-theme-preview.md)
49. [Clean and Cyber Theme Pass](./049-clean-cyber-theme-pass.md)
50. [Interface Styling Audit](./050-interface-styling-audit.md)

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
