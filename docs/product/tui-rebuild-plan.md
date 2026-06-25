# TUI Rebuild Plan

This plan reflects the adaptive rewrite direction.

## Slice 1: Config Migration

Goal: collapse appearance configuration to one adaptive Theme.

Work:

- default `theme = "adaptive"`
- normalize legacy theme values during defaulting/loading
- keep symbol mode and motion preference as first-class settings

## Slice 2: Shared Appearance System

Goal: replace hardcoded theme branching with semantic adaptive components.

Work:

- one adaptive theme token set
- shared panel surfaces
- shared status pills
- shared footer and notification styling

## Slice 3: Dashboard Rewrite

Goal: make the Dashboard feel like a modern terminal workspace.

Work:

- left-aligned shell
- primary recommended action region
- compact supporting queue
- fixed-width sidebar for Profile and Roadmap context

## Slice 4: Screen Cleanup

Goal: bring deeper screens into the same adaptive system.

Work:

- Roadmap Detail labels and footer cleanup
- Stage Detail labels and footer cleanup
- Problem Detail labels and footer cleanup
- Practice Log cleanup

## Slice 5: Onboarding Simplification

Goal: remove obsolete appearance choice UX.

Work:

- remove Theme Selection from the main flow
- keep symbol-mode preview behavior where useful
- update step counts and completion flow

## Slice 6: Documentation Lock-In

Goal: make the rewrite the new source of truth.

Work:

- update `CONTEXT.md`
- update product specs
- add an ADR for the single adaptive TUI

## Success Criteria

- no user-facing multi-theme behavior remains
- legacy theme values auto-migrate safely
- Dashboard hierarchy is clearer and denser
- deeper screens use stable product language
- tests pass after the rewrite
