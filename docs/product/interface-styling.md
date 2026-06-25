# Interface Styling Product Spec

This document defines Leetgo's current TUI styling direction.

Read with:

- `CONTEXT.md`
- `docs/adr/0007-dashboard-first-tui.md`
- `docs/adr/0009-screen-based-tui.md`
- `docs/adr/0011-single-adaptive-tui.md`

## Styling Goal

Leetgo should feel modern, calm, keyboard-first, and terminal-native.

The TUI should not look like a themed mockup or a web dashboard squeezed into a terminal. It should feel like a focused local tool with clear hierarchy, adaptive contrast, and compact information density.

## Single Adaptive Theme

Leetgo now uses one adaptive Theme.

Rules:

- No alternate branded Themes.
- No RPG, clean, or cyber identity branches.
- No user-facing theme switching in the TUI.
- Legacy saved theme values are migration inputs only and should normalize to `adaptive`.

Adaptation should come from:

- terminal background brightness
- terminal color capabilities
- symbol mode
- motion preference

## Appearance System

The adaptive Theme should use semantic roles rather than screen-specific or brand-specific colors.

Required palette roles:

- `Primary`: strongest emphasis and active border.
- `Success`: stable completion.
- `Warning`: action-needed states such as Verified.
- `Danger`: failures and invalid actions.
- `Muted`: helper copy and low-priority metadata.
- `Border`: panel structure and separators.
- `XP`: progression and reward emphasis.
- `Review`: review-specific emphasis.

## Layout Principles

Leetgo should borrow modern terminal interaction patterns:

- one stable shell per screen
- left-aligned composition instead of floating centered canvases
- one dominant content region plus a fixed-width sidebar on wide terminals
- stacked sidebar on medium terminals
- content-first narrow layouts on small terminals
- subtle panel structure instead of heavy decorative frames everywhere

## Dashboard Direction

The Dashboard should be the clearest expression of the design system.

It should show:

- a primary recommended action with strong emphasis
- a compact queue of supporting actions
- a sidebar with Profile and Roadmap context
- short inline rationale for why the next action is ranked first

It should avoid:

- equal visual weight on every panel
- excessive dead space
- decorative fantasy labels
- background effects that reduce readability

## Copy Direction

Use product language, not themed language.

Preferred labels:

- `Dashboard`
- `Recommended`
- `Up next`
- `Profile`
- `Roadmap`
- `Stage`
- `Problem Detail`
- `Problem Brief`
- `Workspace Files`
- `Practice Log`

Avoid labels like:

- `Quest Board`
- `Character HUD`
- `Map Fragment`
- `World Map`
- `Zone`
- `Signal Map`

## Symbol Modes

Leetgo supports two symbol modes.

### Rich Symbols

- Default for modern terminals.
- Should improve scanning, not decorate the UI.

### Plain Symbols

- ASCII-safe fallback.
- Must preserve all state meaning.
- No hard-coded rich glyphs should leak into plain mode.

## Motion Preference

Leetgo supports `normal`, `reduced`, and `off` motion preferences.

Rules:

- `off` means no decorative animation.
- `reduced` keeps only necessary progress feedback.
- `normal` may use restrained feedback, never constant idle motion.
- Non-interactive CLI output behaves as motion off.

## Status Hierarchy

Status emphasis should remain consistent across screens.

Priority:

1. focused recommended action
2. Verified
3. InProgress
4. Available
5. Solved
6. Review
7. Locked

Rules:

- Verified should feel more urgent than Solved.
- Locked should be muted, not alarming.
- Review should be distinct from error styling.

## Shared Components

The TUI should converge on a small component language:

- panel surfaces
- status pills
- compact queue rows
- section headers
- key hint footer
- notifications
- centered overlays/modals when needed

New screen work should extend this shared system instead of creating one-off visual treatments.
