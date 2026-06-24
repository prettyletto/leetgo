# Task 050: Interface Styling Audit

## Goal

Audit the completed interface styling work for usability, consistency, accessibility, terminal compatibility, and test coverage.

## Context

Read:

- `CONTEXT.md`
- `docs/product/interface-styling.md`
- Tasks 040 through 049

This task should run after the styling implementation tasks are complete.

## Dependencies

- Tasks 040 through 049, or the subset being audited.

## Audit Checklist

Visual system:

- Terminal Palette roles are used consistently.
- Rich Symbols and Plain Symbols preserve the same meaning.
- Locked is not styled as Danger.
- Verified is action-needed and visually stronger than Solved.
- Manual Solve is completion plus provenance marker.
- Review is distinct and not error-styled.

Themes:

- RPG Skill Tree feels game-like, mature, and pixel-framed.
- Clean Productivity feels compact and focused.
- Cyber Dashboard feels technical and high-signal.
- Theme-specific labels do not change behavior.

Screens:

- Dashboard primary Next Action is always visible.
- Roadmap Detail uses readable staged lanes.
- Stage Detail makes recommended, locked, and review Problems clear.
- Problem Detail reweights sections by Status.
- Practice Log is readable and uses Practice Log wording.
- Roadmap Completion feels like a Reward Moment.

Responsive behavior:

- Wide, medium, and narrow layouts preserve primary action and Status.
- Below 60x18, Unsupported Size renders instead of broken layout.
- Keytips remain visible.

CLI:

- Reward Moments are structured and beautiful in interactive output.
- Non-TTY output remains script-friendly.
- Motion Preference is respected.

Accessibility:

- Plain Symbols are usable.
- Motion off/reduced is respected.
- Color is not the only carrier of state.

## Acceptance Criteria

- Produce `docs/audits/interface-styling-audit.md`.
- Findings are ordered by severity.
- Findings include file/line references where possible.
- Report includes residual risks and missing tests.
- `go test ./...` result is recorded.
- Any terminology drift is fixed or listed as a finding.

## Verification

Run:

```bash
go test ./...
```
