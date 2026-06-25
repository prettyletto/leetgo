# TUI Screen Spec

This spec defines the current adaptive TUI behavior.

## Global Rules

- The TUI opens to Onboarding when `onboarding_complete` is false or stale.
- The TUI opens to Dashboard when Onboarding is complete.
- Screens own their own keybindings and render a persistent key hint footer.
- Notifications are global and may appear over any screen.
- The adaptive Theme is global, but it is not user-switchable inside the TUI.
- Reduced terminal width should collapse secondary context before hiding primary actions.

## Screen Architecture

Current screens:

- Onboarding
- Dashboard
- Roadmap Detail
- Stage Detail
- Problem Detail
- Practice Log
- Roadmap Completion

## Onboarding

Purpose: create a Profile and collect Practice Preferences before the first Dashboard.

Steps:

1. Welcome and Display Name.
2. Git Export backup opt-in.
3. Workspace and Language confirmation.
4. Roadmap Selection.
5. Optional LeetCode Session setup.
6. Completion and Dashboard handoff.

Onboarding should not include a user-facing Theme Selection step.

## Dashboard

Purpose: show what the user should do next.

Layout:

- main content column: recommended action plus supporting queue
- secondary sidebar: Profile and Roadmap context

Responsive behavior:

- wide terminals show main column plus fixed-width sidebar
- medium terminals stack the sidebar below the main content
- narrow terminals prioritize the main content and footer

Primary keybindings:

- `enter`: activate focused next action
- `j/k` or `up/down`: move focus
- `r`: open Roadmap Detail
- `s`: open Practice Log
- `l`: open legacy list view
- `q`: quit

## Roadmap Detail

Purpose: inspect the selected Roadmap's Unlock Path and Stage progress.

Keybindings:

- `enter`: open focused Stage or Problem
- `j/k`: move focus
- `g`: toggle graph/list representation if both exist
- `esc` or `backspace`: return to Dashboard

## Stage Detail

Purpose: inspect one Stage and act on its Problems.

Keybindings:

- `enter`: open focused Problem
- `j/k`: move focus
- `esc` or `backspace`: return to Roadmap Detail

## Problem Detail

Purpose: act on one Problem.

Keybindings:

- `enter`: primary action for the current Status
- `o`: open in editor
- `x`: run local tests
- `s`: submit
- `m`: manual solve flow
- `esc` or `backspace`: return to Stage Detail

## Practice Log

Purpose: review recent practice history.

Keybindings:

- `j/k`: move focus
- `esc`: return to Dashboard

## Roadmap Completion

Purpose: summarize the user's completion of a Roadmap and suggest what to do next.

Keybindings:

- `r`: open Roadmap Detail
- `esc`: return to Dashboard
- `q`: quit
