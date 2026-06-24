# Interface Styling Sprint Plan

This plan groups Tasks `040` through `050` into implementation sprints. The final audit sprint should run separately after implementation.

Read first:

- `CONTEXT.md`
- `docs/product/interface-styling.md`
- `docs/product/learning-system.md`
- `docs/tasks/README.md`

## Sprint Rules

- Complete all tasks in the assigned sprint unless blocked.
- Do not start later sprints unless explicitly assigned.
- Preserve behavior unless a task explicitly changes presentation behavior.
- Add or update tests in the same sprint.
- Run `gofmt` on changed Go files and `go test ./...` before handoff.
- Handoff must include changed files, completed tasks, test result, and remaining risks.

## Sprint 1: Appearance Foundation

Goal: add config and shared styling primitives before redesigning screens.

Tasks:

- [Task 040: Appearance Config](./040-appearance-config.md)
- [Task 041: Terminal Palette and Symbol Set](./041-terminal-palette-symbols.md)
- [Task 042: Shared View Components](./042-shared-view-components.md)
- [Task 043: Unsupported Size and Responsive Shell](./043-unsupported-size-responsive-shell.md)

Expected result:

- `symbol_mode` and `motion_preference` exist.
- Terminal Palette roles are available.
- Rich/Plain SymbolSets are available.
- PixelFrame, StatusPill, ProgressBar, KeytipFooter, and UnsupportedSize helpers exist.
- Root renders Unsupported Size below 60x18.

Suggested handoff prompt:

```text
Implement Sprint 1 from docs/tasks/interface-styling-sprints.md. Complete Tasks 040, 041, 042, and 043 with tests. Do not redesign screens beyond what is needed to adopt the shared primitives. Run gofmt and go test ./... before handoff.
```

## Sprint 2: RPG Core Screens

Goal: make the main guided learning flow feel like the RPG skill tree Theme.

Tasks:

- [Task 044: Dashboard RPG Redesign](./044-dashboard-rpg-redesign.md)
- [Task 045: Roadmap and Stage RPG Redesign](./045-roadmap-stage-rpg-redesign.md)
- [Task 046: Problem Detail and Practice Log RPG Redesign](./046-problem-practice-rpg-redesign.md)

Expected result:

- Dashboard feels like Quest Board + Character HUD.
- Roadmap Detail feels like a world map / skill tree.
- Stage Detail feels like a zone.
- Problem Detail feels like skill tile inspection.
- Practice Log is readable and themed.

Suggested handoff prompt:

```text
Implement Sprint 2 from docs/tasks/interface-styling-sprints.md. Complete Tasks 044, 045, and 046 with tests. Focus on RPG Theme presentation while preserving behavior. Run gofmt and go test ./... before handoff.
```

## Sprint 3: Reward Moments And Onboarding Preview

Goal: make progress feel satisfying and make Theme choice understandable.

Tasks:

- [Task 047: Reward Moments for TUI and CLI](./047-reward-moments.md)
- [Task 048: Onboarding Theme Preview and Symbol Fallback](./048-onboarding-theme-preview.md)

Expected result:

- Accepted Solve, Manual Solve, Review Cycle completion, and Roadmap Completion have Reward Moments.
- CLI Reward Moments are styled but non-TTY safe.
- Onboarding Theme Selection shows previews and symbol fallback.

Suggested handoff prompt:

```text
Implement Sprint 3 from docs/tasks/interface-styling-sprints.md. Complete Tasks 047 and 048 with tests. Focus on Reward Moments, CLI/TUI consistency, Theme previews, and symbol fallback. Run gofmt and go test ./... before handoff.
```

## Sprint 4: Alternate Theme Pass

Goal: apply the styling system to Clean Productivity and Cyber Dashboard without changing behavior.

Tasks:

- [Task 049: Clean and Cyber Theme Pass](./049-clean-cyber-theme-pass.md)

Expected result:

- Clean Theme is compact, minimal, and focused.
- Cyber Theme is technical, high-signal, and cockpit-like.
- Theme-specific labels render correctly.
- Behavior remains identical across Themes.

Suggested handoff prompt:

```text
Implement Sprint 4 from docs/tasks/interface-styling-sprints.md. Complete Task 049 with tests. Do not change learning behavior. Run gofmt and go test ./... before handoff.
```

## Sprint 5: Interface Styling Audit

Goal: audit the completed styling implementation.

Tasks:

- [Task 050: Interface Styling Audit](./050-interface-styling-audit.md)

Expected result:

- Written audit report at `docs/audits/interface-styling-audit.md`.
- Severity-ordered findings.
- Residual risks and missing tests.
- `go test ./...` result recorded.

Suggested handoff prompt:

```text
Audit the interface styling implementation using Sprint 5 from docs/tasks/interface-styling-sprints.md. Complete Task 050 only. Do not fix findings unless explicitly asked. Produce docs/audits/interface-styling-audit.md with severity-ordered findings, file/line references, residual risks, missing tests, and go test ./... result.
```

## Recommended Order

1. Sprint 1: Appearance Foundation.
2. Sprint 2: RPG Core Screens.
3. Sprint 3: Reward Moments And Onboarding Preview.
4. Sprint 4: Alternate Theme Pass.
5. Sprint 5: Interface Styling Audit.

Parallelization note:

Avoid parallelizing Sprint 1 and Sprint 2 because screen redesign depends on shared primitives. Sprint 3 can start after enough Reward Moment primitives exist, but sequential execution is safer.
