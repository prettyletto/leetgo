# TUI Screen Spec

This spec defines the target screen behavior for the Dashboard-first TUI rebuild.

## Global TUI Rules

- The TUI opens to Onboarding when `onboarding_complete` is false or missing.
- The TUI opens to Dashboard when Onboarding is complete.
- Screens own their own keybindings and render a persistent keytip footer.
- Notifications are global and may appear over any screen.
- Theme is global and affects colors, borders, typography emphasis, and allowed animation.
- Reduced terminal width should collapse side rails before hiding primary actions.

## Screen Architecture

Target screens:

- Onboarding
- Roadmap Selection
- Theme Selection
- Dashboard
- Roadmap Detail
- Stage Detail
- Problem Detail
- Solve Log Detail, optional later

The current list, graph, heatmap, stats, and notification code should become reusable components or be moved into one of these screens.

## Onboarding

Purpose: create a Profile and collect Practice Preferences before the first Dashboard.

Steps:

1. Welcome and Display Name.
2. Git Export backup opt-in.
3. Workspace and Language confirmation.
4. Roadmap Selection.
5. Theme Selection.
6. Completion and Dashboard handoff.

### Welcome and Display Name

Primary prompt:

`Who are you, challenger?`

Inputs:

- Display Name text input.

Validation:

- Required.
- Trim whitespace.
- Max 40 visible characters.

Save:

- `display_name`.

### Git Export Backup Opt-In

Primary prompt:

`Do you want Leetgo to back up progress to a Git repo?`

Options:

- Yes, use Git Export backup.
- Not now.

If yes:

- Ask for Git repository path.
- Validate the path exists, is a directory, and contains `.git`.
- Derive Export Identity from Git email when available.
- If Git email is missing, warn that a local fallback identity will be used.

Save:

- `git_export_enabled`.
- `git_export_repo`.

### Workspace and Language Confirmation

Purpose: confirm where generated Problem files will go and which Language to use.

Inputs:

- Workspace path, prefilled with default.
- Language selector, prefilled with default `go`.

Validation:

- Workspace must be non-empty.
- Language must be one of supported generator languages.

Save:

- `workspace`.
- `language`.

### Roadmap Selection

Uses Roadmap Carousel.

Initial focus:

- Roadmap with `recommended: true`, falling back to `from-zero-to-hero`.

Keybindings:

- `<` / `left`: previous Roadmap.
- `>` / `right`: next Roadmap.
- `enter`: confirm focused Roadmap.
- `q`: quit Onboarding with no completion.

Card layout:

- Left preview card.
- Center focused card.
- Right preview card.

Focused card content:

- Title.
- Recommended label when applicable.
- Tagline.
- Audience.
- Promise.
- Problem count.
- Difficulty mix.
- Estimated hours.
- First 2-3 Stages.
- Highlights.

Save:

- `roadmap`.

### Theme Selection

Options:

- RPG skill tree, default.
- Clean productivity.
- Cyber dashboard.

Keybindings:

- `<` / `left`: previous Theme.
- `>` / `right`: next Theme.
- `enter`: confirm.

Save:

- `theme`.

Completion:

- Set `onboarding_complete = true`.
- Save config.
- Open Dashboard.

## Dashboard

Purpose: show what the user should do next.

Layout:

- Center: Next Actions.
- Left rail: Profile/game HUD.
- Right rail: Roadmap/Stage context.

Responsive behavior:

- Wide terminals show all three regions.
- Medium terminals stack rails below center.
- Narrow terminals show only Next Actions and footer, with keys to open HUD/context overlays later.

### Center: Next Actions

Rank order:

1. Continue InProgress Problems.
2. Start earliest Available Problems in selected Roadmap/Stage order.
3. Weakness-targeting review actions.
4. Maintenance actions such as Git Export or Solve Log review.

Each Next Action card includes:

- Action label: Continue, Start, Review, Export, Inspect.
- Problem title or action name.
- Stage and Category when Problem-backed.
- Why this is recommended.
- Keytip for primary action.

Primary keybindings:

- `enter`: activate focused Next Action.
- `j/k` or `up/down`: move focus.
- `r`: Roadmap Detail.
- `s`: Solve Log.
- `t`: cycle Theme.
- `q`: quit.

### Left Rail: Profile/Game HUD

Includes:

- Display Name.
- Level.
- XP progress to next Level.
- Streak.
- Solved count.
- Current Roadmap.
- Latest Achievement.

Does not include:

- Weakness list.
- Full Solve Log.
- Git Export details.

### Right Rail: Roadmap/Stage Context

Includes:

- Selected Roadmap title.
- Current Stage.
- Stage completion.
- Next blocker summary.
- Keytip to open Roadmap Detail.

Does not show the full Unlock Path graph; that belongs in Roadmap Detail.

## Roadmap Detail

Purpose: inspect the selected Roadmap's Unlock Path and Stage progress.

Content:

- Roadmap title, tagline, promise.
- Stage progress list.
- Unlock Path visualization.
- Available Problems and blockers.

Keybindings:

- `enter`: open focused Stage or Problem.
- `j/k`: move focus.
- `g`: toggle graph/list representation if both exist.
- `esc` / `backspace`: return to Dashboard.

## Stage Detail

Purpose: inspect and act on one Stage.

Content:

- Stage title and description.
- Completion count.
- Problem list for the Stage.
- Locked Problem blockers.
- Available and InProgress Problems emphasized.

Keybindings:

- `enter`: open Problem Detail.
- `j/k`: move focus.
- `esc` / `backspace`: return to Roadmap Detail.

## Problem Detail

Purpose: own Problem-specific actions.

Content:

- Problem ID, title, Difficulty, Category, Stage.
- Status.
- Prerequisites.
- Blockers when locked.
- Local file paths when generated.
- Latest Solve Log entries for this Problem.

Actions:

- Start.
- Open editor.
- Run local TestSuite.
- Submit to LeetCode.
- Mark Solved manually.

Keybindings:

- `enter`: primary action based on Status.
- `o`: open editor.
- `x`: run local TestSuite.
- `s`: submit.
- `m`: mark Solved.
- `esc` / `backspace`: return to Stage Detail.

## Theme Behavior

### RPG Skill Tree

- Default Theme.
- Strong status colors.
- Bordered cards.
- XP and Level treated as HUD elements.
- Reward animations for Level-ups and Achievements.

### Clean Productivity

- Minimal color.
- Low motion.
- High readability.
- Compact cards and panels.

### Cyber Dashboard

- High contrast.
- Neon accents.
- Subtle ambient background motion.
- Async spinners and focus effects can be more expressive.

## Migration From Current TUI

Current root modes map roughly to future screens:

- Current list view → Stage Detail / Problem browsing component.
- Current graph view → Roadmap Detail Unlock Path component.
- Current heatmap view → Dashboard secondary panel or future stats screen.
- Current stats bar → Profile/game HUD component.
- Current notifications → global notification layer.
