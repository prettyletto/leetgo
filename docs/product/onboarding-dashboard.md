# Onboarding and Dashboard Target

This document captures the target first-run and default TUI experience for the next major Leetgo UI rebuild.

## Product Shape

Leetgo should feel like a guided practice companion, not a catalog browser. The default post-Onboarding screen is an action-first Dashboard that tells the user what to do next while keeping Profile, game progress, and Roadmap context visible.

## Onboarding Flow

Onboarding appears when `onboarding_complete` is false or missing. Existing configs should be treated as a pre-filled migration flow, not silently considered complete.

Steps:

1. Welcome and Display Name.
2. Git Export backup opt-in.
3. Workspace and Language confirmation.
4. Roadmap Selection with Roadmap Carousel.
5. Theme selection.
6. Enter Dashboard.

Onboarding saves Profile and Practice Preferences to `config.toml`. It does not require LeetCode authentication; Session setup remains just-in-time when the user first submits.

## Profile and Practice Preferences

Persist in `config.toml`:

- `onboarding_complete`
- `display_name`
- `workspace`
- `language`
- `roadmap`
- `theme`
- `git_export_enabled`
- `git_export_repo`

SQLite remains responsible for progress/history: Status, Attempts, Solve Logs, XP, Streaks, and Achievements.

## Git Export Backup

If the user opts into Git Export backup during Onboarding, they must provide a valid Git repository path. If they do not have a repo ready, they choose not now and can enable it later.

Git Export backup does not make the generated Workspace a Git repo. It does not auto-push. Commits remain explicit.

## Roadmap Selection

Roadmap Selection uses a Roadmap Carousel:

- The focused Roadmap is centered.
- Neighboring Roadmaps preview left and right.
- `<` and `>` rotate focus.
- `enter` confirms the focused Roadmap.
- `from-zero-to-hero` is the first-version recommended/default Roadmap.

Roadmap cards show:

- title
- one-line promise
- intended audience
- problem count
- difficulty mix
- first 2-3 Stages
- focused-card confirm keytip

This requires richer Roadmap metadata than the current YAML files contain.

Recommended Roadmap YAML metadata:

```yaml
id: from-zero-to-hero
title: "From Zero To Hero"
tagline: "Build LeetCode fundamentals from zero into interview readiness."
audience: "New or returning users who want a guided foundation-first path."
promise: "You will learn the core patterns in the order they unlock each other."
recommended: true
estimated_hours: 80
difficulty_mix:
  easy: 35
  medium: 50
  hard: 15
highlights:
  - "Foundation-first unlock path"
  - "Broad interview coverage"
  - "Gentle difficulty ramp"
```

Field intent:

- `id`: stable machine identifier used in config and progress context.
- `title`: display name shown on cards and headers.
- `tagline`: one short card line, optimized for fast scanning.
- `audience`: who should choose this Roadmap.
- `promise`: what the Roadmap claims the user will get by following it.
- `recommended`: whether this Roadmap receives the Onboarding recommended label.
- `estimated_hours`: rough total effort signal, not a guarantee.
- `difficulty_mix`: percentage split used for card stats.
- `highlights`: 2-3 bullets that make the card feel distinct.

Derived card data:

- Problem count should be derived from `problems`, not manually stored.
- Stage preview should use the first 2-3 `stages`, not separate copy.
- Difficulty counts can be computed from Problems, but `difficulty_mix` gives a stable product-level summary for the card.

Validation rules:

- `id`, `title`, `tagline`, `audience`, and `promise` are required.
- Only one bundled Roadmap should have `recommended: true`.
- `difficulty_mix` values should add up to 100 when present.
- `highlights` should contain 2-3 items.
- `estimated_hours` must be positive when present.

## Themes

Leetgo supports three Themes:

- RPG skill tree, default
- clean productivity
- cyber dashboard

Theme is selected during Onboarding, can be switched from the Dashboard, and persists in config.

Animation rules:

- Focus transitions are allowed.
- Async states such as Submission and Git Export use spinners or similar motion.
- Rewards such as Level-ups and Achievements may animate.
- Ambient motion is Theme-specific and low-intensity: cyber may have subtle background motion; RPG and clean remain quieter.

## Dashboard Layout

The Dashboard is the default post-Onboarding screen.

Layout:

- Center: Next Actions.
- Left rail: Profile/game HUD.
- Right rail: Roadmap/Stage context.

Profile/game HUD includes:

- Display Name
- Level
- XP progress to next Level
- Streak
- solved count
- current Roadmap
- latest Achievement

Roadmap context rail includes:

- selected Roadmap title
- current Stage
- Stage completion
- blocker summary
- keytip to open Roadmap Detail

## Next Actions

Rank Next Actions in this order:

1. Continue InProgress Problems.
2. Start earliest Available Problems in selected Roadmap/Stage order.
3. Weakness-targeting review actions.
4. Maintenance actions such as Git Export or Solve Log review.

## Deep Navigation

Navigation hierarchy:

1. Dashboard.
2. Roadmap Detail.
3. Stage Detail.
4. Problem Detail.

Roadmap Detail shows the Unlock Path and Stage progress.

Stage Detail shows Problems, completion, and blockers for one Stage.

Problem Detail owns Start, local testing, Submission, and Solve Log history.

## Architecture Direction

The TUI should move to a screen-based architecture rather than continuing to expand the current root model. Target screens:

- Onboarding
- Roadmap Selection
- Dashboard
- Roadmap Detail
- Stage Detail
- Problem Detail

Existing list, graph, heatmap, stats, and notification pieces should become reusable components or be migrated into the appropriate screens.
