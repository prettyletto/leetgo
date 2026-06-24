# TUI Rebuild Plan

This plan breaks the Dashboard-first TUI into implementation slices.

## Slice 1: Config and Metadata Foundation

Goal: make the data model capable of supporting Onboarding and Roadmap Carousel.

Work:

- Add config fields from `docs/product/config-schema.md`.
- Add Theme validation.
- Add Git Export preference validation.
- Add Roadmap metadata fields: `tagline`, `audience`, `promise`, `recommended`, `roadmap_time_estimate`, `difficulty_mix`, `highlights`.
- Update all bundled Roadmap YAML files with metadata.
- Validate Roadmap metadata.

Acceptance:

- Existing configs still load with defaults.
- Missing `onboarding_complete` causes TUI to choose Onboarding.
- Roadmap metadata validates in tests.

## Slice 2: Screen Shell

Goal: introduce screen-based TUI architecture without rebuilding every screen at once.

Work:

- Add a `Screen` interface.
- Add root model that delegates Update/View to active screen.
- Add navigation events between screens.
- Move notifications/global footer to root shell.
- Keep current list behavior behind a temporary legacy screen if needed.

Acceptance:

- Existing TUI still opens and works through the new shell.
- Tests cover screen navigation basics.

## Slice 3: Onboarding Screens

Goal: implement first-run Profile and Practice Preferences collection.

Work:

- Welcome/Display Name screen.
- Git Export backup opt-in screen.
- Workspace/Language screen.
- Roadmap Carousel screen.
- Theme Selection screen.
- Save config and set `onboarding_complete`.

Acceptance:

- Fresh config opens Onboarding.
- Completed Onboarding writes config and opens Dashboard.
- Existing partial config pre-fills defaults.

## Slice 4: Dashboard MVP

Goal: replace list default with an action-first Dashboard.

Work:

- Add Next Action calculation.
- Add Profile/game HUD component.
- Add Roadmap/Stage context component.
- Add Theme-aware panel styling.
- Wire actions to existing Start/Test/Submit paths where possible.

Acceptance:

- Completed Profile opens Dashboard by default.
- InProgress Problems rank first.
- Available Problems rank after InProgress.
- Dashboard shows Display Name, Level, XP, Streak, solved count, Roadmap, Stage context.

## Slice 5: Roadmap, Stage, and Problem Detail

Goal: move browsing and actions into deeper screens.

Work:

- Roadmap Detail with Unlock Path and Stage progress.
- Stage Detail with filtered Problem list.
- Problem Detail with Start/Test/Submit/Practice Log actions.
- Migrate current list/graph views into these screens.

Acceptance:

- Dashboard can drill into Roadmap Detail.
- Roadmap Detail can drill into Stage Detail.
- Stage Detail can drill into Problem Detail.
- Problem Detail can Start, Test, Submit, and Mark Solved.

## Slice 6: Theme Polish and Motion

Goal: make the TUI feel deliberately designed.

Work:

- RPG skill tree Theme.
- Clean productivity Theme.
- Cyber dashboard Theme.
- Theme switching from Dashboard.
- Focus transitions.
- Submission/export spinners.
- Achievement/Level-up reveal.
- Subtle cyber ambient motion.

Acceptance:

- Theme is persisted.
- Theme changes affect all screens consistently.
- Motion is restrained and does not block readability.

## Known Rebuild Risks

- Current `internal/tui/model.go` mixes app state, list rendering, submission, editor opening, and view mode handling.
- Moving to screens should avoid duplicating Start/Test/Submit behavior.
- Existing tests rely on root model details and will need migration.

## Suggested Agent Task Boundaries

Good independent tasks:

- Config and Roadmap metadata validation.
- Roadmap Carousel component.
- Theme system and styles.
- Next Action calculator.
- Dashboard MVP layout.
- Roadmap Detail migration.
- Stage Detail migration.
- Problem Detail actions.

Avoid splitting too early:

- Root screen shell and navigation should be done first by one agent to prevent incompatible screen contracts.
