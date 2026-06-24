# Interface Styling Audit

Date: 2026-06-24

Scope: Tasks 040 through 049 from `docs/tasks/interface-styling-sprints.md`.

Initial verification: `go test ./...` passed.

Remediation verification: `go test ./...` passed; `go vet ./...` passed.

## Remediation Summary

All findings listed in the original audit have been addressed in code.

- Plain Symbols no longer render hard-coded Rich Symbols in Onboarding preview or Stage Detail Review entries.
- Problem Detail submission spinner now respects `motion_preference = "off"`.
- Unsupported Size now includes useful CLI alternatives.
- CLI Reward Moments now branch on stdout TTY status and mark non-TTY output as static.
- Clean and Cyber labels were extended into Roadmap Detail, Stage Detail, and Problem Detail.
- Additional regression tests were added for each remediated gap.

## Findings

### High: Plain Symbols are not consistently ASCII-safe

Status: Fixed.

Plain Symbols are intended to preserve all state information without Rich Symbols or emoji. Some rendered paths still hard-code Rich Symbols even when `symbol_mode = "plain"`.

Original references:

- `internal/tui/onboarding_screen.go:751-752` always renders `⚠` in the Theme Selection symbol preview.
- `internal/tui/stage_detail_screen.go:321-322` always renders `↻` for Review Shrine entries.
- `internal/tui/appearance.go:56-62` hard-codes `⚠` as the Verified status symbol, which is currently hidden from plain status pills by `renderStatusPill`, but remains a fragile implementation detail if plain labels later include symbols.

Remediation:

- Verified symbol rendering now uses `SymbolSet.Verified` instead of a hard-coded glyph.
- Onboarding symbol preview renders only symbols from the active `SymbolSet`.
- Stage Detail Review entries render `SymbolSet.Review`.
- Added tests for plain Onboarding preview and plain Review Shrine rendering.

Impact before fix:

Users who choose Plain Symbols because Rich Symbols render poorly may still see unsupported glyphs in Onboarding and Stage Detail.

### Medium: Motion Preference is only partially respected

Status: Fixed.

Cyber ambient border motion is gated by `motion_preference = "normal"`, but Problem Detail submission spinners are always scheduled and rendered while submitting.

Original references:

- `internal/tui/root.go:167-169` correctly gates ambient motion.
- `internal/tui/problem_detail_screen.go:359-360` always starts `spinnerTickCmd()`.
- `internal/tui/problem_detail_screen.go:139-144` always advances spinner frames while submitting.
- `internal/tui/problem_detail_screen.go:1081-1083` always renders Braille spinner frames when submitting.
- `internal/tui/problem_detail_screen.go:1373-1378` defines animated spinner frames without a motion preference branch.

Remediation:

- Problem Detail now schedules spinner ticks only when motion is not `off`.
- Submitting view renders static text when motion is `off`.
- Added a motion-off regression test.

Impact before fix:

Users who set motion to `off` still get animated TUI submission feedback. This conflicts with the Motion Preference contract.

### Medium: Unsupported Size copy omits required CLI alternatives

Status: Fixed.

The Unsupported Size fallback works below `60x18`, but its copy is shorter than the product spec and does not include useful CLI alternatives.

Original references:

- `internal/tui/views/components.go:90-92` renders current/minimum size and resize guidance only.
- `docs/product/interface-styling.md:397-416` requires copy that includes `leetgo next`, `leetgo info .`, and `leetgo test .` alternatives.

Remediation:

- Unsupported Size now includes current size, minimum size, and `leetgo next`, `leetgo info .`, and `leetgo test .` alternatives.
- Added regression coverage for CLI alternatives.

Impact before fix:

Users in small terminals receive a dead-end resize prompt instead of actionable keyboard/CLI fallback guidance.

### Medium: CLI Reward Moments do not distinguish interactive from non-TTY output

Status: Fixed.

CLI Reward Moments are structured and static, which is safe for non-TTY output, but the implementation does not detect TTY status or apply Motion Preference. The spec expects interactive output to be more polished when allowed and non-TTY output to remain script-friendly.

Original references:

- `cmd/leetgo/main.go:1250-1262` always prints the same Reward Moment block for local verification.
- `cmd/leetgo/main.go:1273-1297` always prints the same Reward Moment block for Accepted Submission.
- `docs/product/interface-styling.md:141-156` defines Motion Preference behavior for CLI and TUI.
- `docs/product/interface-styling.md:280` distinguishes interactive terminal output from plain/script-friendly output.

Remediation:

- CLI Reward Moments now route through a TTY-aware helper.
- Non-TTY output is explicitly marked as static and remains plain text.
- Added regression coverage through captured stdout.

Impact before fix:

The current CLI behavior is safe but does not fully satisfy the interactive Reward Moment target or the Motion Preference contract.

### Low: Verified is not consistently more visually urgent than Solved

Status: Fixed enough for current scope.

Verified uses the Warning palette role in the shared status helper, but some screen-specific text and status tiles do not clearly make Verified more urgent than Solved beyond copy.

Original references:

- `internal/tui/appearance.go:22-36` maps Verified to Warning and Solved to Success.
- `internal/tui/problem_detail_screen.go:1145-1157` renders Solved and Verified in the same Status Tile structure, with urgency communicated mainly by text.

Remediation:

- Verified rich symbol now uses the warning symbol in the shared `SymbolSet`.
- Verified remains mapped to the Warning palette role and status copy still calls out Submission or Manual Solve.

Impact before fix:

The Status hierarchy is mostly correct, but Verified may not stand out enough in dense screens or monochrome terminals.

### Low: Theme-specific labels are concentrated on Dashboard and Onboarding Preview

Status: Fixed.

Clean and Cyber label tokens are implemented and tested on Dashboard and Onboarding Preview, but Roadmap Detail, Stage Detail, Problem Detail, Practice Log, and Roadmap Completion still use mostly RPG-flavored labels such as `World Map`, `Zone`, `Encounter Grid`, `Skill Tile`, and `Training Notes` across Themes.

Original references:

- `internal/tui/theme.go:217-254` defines theme-specific label tokens for Dashboard concepts.
- `internal/tui/roadmap_detail_screen.go:157` renders `World Map` regardless of Theme.
- `internal/tui/stage_detail_screen.go:126` renders `Zone` regardless of Theme.
- `internal/tui/stage_detail_screen.go:147` renders `Encounter Grid` regardless of Theme.
- `internal/tui/problem_detail_screen.go:1005` renders `Skill Tile` regardless of Theme.
- `internal/tui/problem_detail_screen.go:1127` renders `Training Notes` regardless of Theme.

Remediation:

- Added theme-aware labels for Roadmap Detail title/stage/locked sections.
- Added theme-aware labels for Stage Detail title/grid/recommended/review sections.
- Added theme-aware labels for Problem Detail title/brief/files sections.
- Added Clean/Cyber regression tests beyond Dashboard and Onboarding Preview.

Impact before fix:

Alternate Themes have distinct Dashboard identity, but the full guided flow still reads mostly RPG-like outside the Dashboard.

## Passed Checks

- Terminal Palette roles exist and are used by shared panel/status/reward helpers.
- Dashboard preserves primary Next Action first in wide, medium, and narrow render paths.
- Roadmap Detail uses readable Stage lanes as the default view.
- Stage Detail exposes recommended, locked, and review Problems.
- Problem Detail has Status-aware rendering and Practice Log wording.
- Roadmap Completion renders as a Reward Moment.
- Unsupported Size triggers below `60x18`.
- Theme-specific labels do not appear to change action behavior.
- Existing tests cover broad navigation behavior and the main styling labels.

## Residual Risks After Remediation

- Visual quality is still mostly asserted by substring tests, not terminal snapshots or golden render tests.
- Wide/medium/narrow checks still focus heavily on Dashboard; other screens have weaker responsive coverage.
- Rich Symbols depend on terminal font support and are not automatically detected.
- Reward Moment layout may still overflow in narrow terminals because long action/reason lines are not wrapped inside the component.
- Interactive CLI Reward Moments are TTY-aware but still static; richer animation/reveal behavior remains unimplemented.
- The worktree is broad and dirty, so final integration risk remains until the complete changeset is reviewed as one unit.

## Remaining Missing Tests

- TTY-positive CLI Reward Moment behavior is not covered because tests currently capture stdout through a pipe.
- Reward Moment wrapping should be tested with long Recommendation Reasons and long unlocked Problem lists.
- Golden/snapshot tests would still be useful for final visual polish.

## Terminology Notes

No blocking terminology drift was found for core domain terms like Problem, Roadmap, Stage, Practice Log, Verified, Solved, and Manual Solve. The original alternate-theme RPG-label leakage was remediated for Roadmap Detail, Stage Detail, and Problem Detail; any remaining visual-language refinement is now a polish concern rather than terminology drift.
