# Interface Styling Product Spec

This document defines the target interface and styling direction for Leetgo's CLI and TUI after the learning-system core is in place.

Read with:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/product/tui-screen-spec.md`
- `docs/adr/0007-dashboard-first-tui.md`
- `docs/adr/0009-screen-based-tui.md`

## Styling Goal

Leetgo should feel like a game-like RPG skill tree built for a modern terminal. The default experience should be playful, pixel-framed, and gamified, while remaining readable, fast, keyboard-first, and useful for serious practice.

The styling goal is not to make Leetgo look like a web app inside a terminal. It should look terminal-native.

## Theme Philosophy

Theme identity comes from layout, emphasis, symbols, panel shape, copy labels, and motion, not only colors.

Themes share the same information architecture:

- Dashboard still has one primary Next Action, supporting actions, Profile/gamification context, and Roadmap context.
- Roadmap Detail still shows the Unlock Path.
- Stage Detail still shows Stage progress, Problems, recommendations, and Blockers.
- Problem Detail still shows Status, Problem Brief, actions, progression context, and Practice Log.

Themes may change presentation labels and visual language, but not behavior.

## Terminal Palette

Leetgo uses a Terminal Palette: ANSI-oriented palette roles rather than fragile truecolor-only styling.

Required roles:

- `Primary`: focused action, selected node, primary callout.
- `Success`: Accepted Solve and stable completion.
- `Warning`: Verified, action-needed, pending progression.
- `Danger`: failed tests, failed Submission, destructive confirmation, invalid config/session errors.
- `Muted`: secondary text, locked-but-normal content, helper copy.
- `Border`: panels, tree connectors, frame lines.
- `XP`: rewards, Level, progress bars.
- `Review`: Review Cycle markers and Review actions.

Themes can assign different colors to the roles, but should prefer ANSI-safe values that work in common terminal backgrounds.

## Theme Identities

### RPG Skill Tree

Default Theme.

Identity:

- Game-like, mature, pixel-framed skill tree.
- Uses Pixel Frames for panels, tiles, Next Actions, and progress summaries.
- Uses Rich Symbols by default with Plain Symbols fallback.
- Uses segmented XP and progress bars.
- Uses quest/map language in section labels.
- Uses motion only for meaningful state transitions.

Labels:

- Primary Next Action: `Main Quest`.
- Secondary actions: `Side Quests`.
- Profile/gamification context: `Character HUD`.
- Roadmap context: `Map Fragment`.
- Stage presentation may label a Stage as a `Zone`, but domain/copy should still preserve Stage where clarity matters.
- Problem presentation may feel like an `Encounter` or skill tile, but domain/copy remains Problem.

Avoid:

- Childish roleplay copy everywhere.
- Emoji spam.
- Constant idle animation.
- Low-information decorative panels.

### Clean Productivity

Identity:

- Focused work queue.
- Compact information density.
- Minimal color.
- Clear labels.
- Motion nearly off.

Labels:

- Primary Next Action: `Recommended`.
- Secondary actions: `Available`.
- Upcoming locked items: `Upcoming`.
- Profile/gamification context: `Profile`.

### Cyber Dashboard

Identity:

- Terminal cockpit.
- Dense, technical, high-signal panels.
- Strong contrast and rails.
- Subtle ambient/focus motion when Motion Preference allows it.

Labels:

- Primary Next Action: `Primary Signal`.
- Secondary actions: `Secondary Targets`.
- Upcoming locked items: `Locked Signals`.
- Profile/gamification context: `Operator`.

## Symbol Modes

Leetgo supports two symbol modes.

### Rich Symbols

Default for modern terminals.

- Assumes Nerd Fonts or modern emoji rendering.
- Uses symbols/icons to communicate state and progression.
- Icons should be consistent and meaningful.
- Icons should not be decorative noise.

### Plain Symbols

Fallback for SSH, minimal terminals, broken glyph rendering, or user preference.

- ASCII-safe.
- Uses labels such as `[ACCEPTED]`, `[VERIFIED !]`, `[LOCKED]`.
- Must preserve all state information.

Config:

- `symbol_mode = "rich"` by default.
- `symbol_mode = "plain"` for fallback.

Onboarding Theme Selection should show a symbol preview and allow switching to Plain Symbols with a keytip, without adding a full extra Onboarding step.

## Motion Preference

Leetgo supports Motion Preference across CLI and TUI.

Options:

- `normal`: spinners, reward reveals, subtle Theme motion.
- `reduced`: useful spinners only, no decorative reveal.
- `off`: no animation; static output only.

Rules:

- Non-interactive command output behaves as motion off.
- Long-running interactive CLI commands may use spinners when motion is normal or reduced.
- TUI may use richer Reward Moments when motion is normal.
- Avoid constant movement while the user is reading.

Config:

- `motion_preference = "normal"` by default.

## Status Visual Hierarchy

Status styling must be consistent across Themes.

Priority and meaning:

1. Focused or Recommended item: strongest visual treatment.
2. Verified: warning/action-needed treatment because it needs Submission or Manual Solve to unlock progression.
3. InProgress: active/current treatment.
4. Available: ready but calm.
5. Accepted Solve: stable success.
6. Manual Solve: success with provenance warning marker.
7. Locked: muted, readable, normal progression.
8. Review: distinct Review role, not error styling.

Verified should be more visually urgent than Solved.

Locked should not use Danger styling.

Manual Solve should look completed but less trusted than Accepted Solve.

Review should never look like an error.

## Suggested Status Symbols

Rich Symbols, exact glyphs can be refined during implementation:

- Accepted Solve: `✓ Accepted`
- Verified: `⚠ Verified`
- InProgress: `~ In Progress`
- Available: `◆ Available`
- Locked: `🔒 Locked`
- Review: `↻ Review`
- Manual Solve: `M Manual`
- XP/Level: `★ XP`

Plain Symbols:

- `[ACCEPTED]`
- `[VERIFIED !]`
- `[IN PROGRESS]`
- `[AVAILABLE]`
- `[LOCKED]`
- `[REVIEW]`
- `[MANUAL]`

## Feedback Levels

Leetgo uses three feedback levels.

`Inline Status`:

- Short feedback inside the active screen near the related action or section.
- Example: local tests failed in Problem Detail.

`Notification`:

- Transient global message.
- Example: Theme changed.

`Reward Moment`:

- Structured feedback after meaningful progress.
- Example: Accepted Solve, Manual Solve, Review Cycle completion, Level-up, Achievement, Roadmap Completion.

Important events should become Reward Moments instead of disappearing into tiny notifications.

## Reward Moments

Reward Moments exist in both TUI and CLI.

Reward Moment content:

- What changed.
- XP gained, if any.
- Solve Duration, when available.
- Newly unlocked Problems.
- Next recommended Problem.
- Recommendation Reason.
- Follow-up actions.

Accepted Solve example:

```text
Problem Solved
#242 Valid Anagram

Reward
+7 XP verify
+3 XP submission

Progress
Unlocked: Group Anagrams
Next: Group Anagrams
Why: applies frequency counting to groups of strings

Actions
leetgo next --start
leetgo info group-anagrams
leetgo roadmap
```

Manual Solve example:

```text
Problem Manually Solved
#242 Valid Anagram

Progress
Unlocked: Group Anagrams

Reward
No XP for Manual Solve

Recommended
Submit later for confidence and XP
```

CLI Reward Moments should be beautiful in interactive terminals, but plain/script-friendly when stdout is not a TTY.

## Screen Direction

### Dashboard

RPG Theme target: quest board plus Character HUD.

Sections:

- `Main Quest`: primary Next Action, largest tile, Recommendation Reason, primary keytip.
- `Side Quests`: Also Available, Review Cycles, Coming Soon.
- `Character HUD`: Level, XP, Streak, Achievements.
- `Map Fragment`: current Stage, next Blocker, mini Unlock Path context.

Rules:

- The primary Next Action must always be visible.
- Dashboard should not normally require scrolling.
- Secondary content should compress before primary content.

### Roadmap Detail

RPG Theme target: world map / skill tree.

Rules:

- Default to readable staged lanes, not perfect graph geometry.
- Show Stage lanes as zones.
- Show Problems as nodes/tiles.
- Use muted styling for locked branches.
- Use stronger styling for Verified Blockers because they are one action away from unlocking.
- Focused node should show exact prerequisites, Blockers, Unlocks, and Builds toward.
- Do not attempt to draw every crossing DAG edge in the full view.

### Stage Detail

RPG Theme target: entering a zone.

Sections:

- Zone Header: Stage title, description, completion.
- Recommended Encounter: best Problem in this Stage.
- Encounter Grid: Problems as tiles grouped by Status.
- Locked Gate: Blockers and near unlocks.
- Review Shrine: active Review Cycles in this Stage.

Domain copy should still preserve `Stage` and `Problem` where precision matters.

### Problem Detail

RPG Theme target: skill tile inspection.

Sections:

- Large status tile with title, Difficulty, Problem Time Estimate.
- Problem Brief as training notes.
- Requires: prerequisites and Blockers.
- Unlocks: direct dependents.
- Builds Toward: indirect dependents.
- Action row: Start, Open, Test, Submit, Manual Solve.
- Compact Practice Log.

Section priority changes by Status:

- Before Start: Problem Brief is prominent.
- InProgress: actions and file paths become more prominent.
- Verified: Submission and Manual Solve actions become dominant.
- Solved: Practice Log, Solve Duration, Unlock impact, and Review opportunities become more prominent.
- Locked: Blockers and recommended prerequisite become prominent.

### Practice Log

Practice Log should be readable and chronological. It should use symbols for Attempt kind and result, but must remain scannable in Plain Symbols.

### Roadmap Completion

Roadmap Completion should be a Reward Moment and a revisit-able screen. It should summarize completion quality and suggested next steps.

## Responsive Layout

Leetgo uses three layout bands.

`Wide`:

- Full intended layout.
- Dashboard: Quest Board + Character HUD + Map Fragment.
- Roadmap Detail: staged lanes/tree.
- Problem Detail: multi-section layout.

`Medium`:

- Primary content first.
- Supporting panels stacked below.
- Fewer columns, same content.

`Narrow`:

- One-column.
- Primary action first.
- Context collapses into short summaries.
- No horizontal graph.
- Keytips always visible.

Critical rule:

- Never hide the primary action or current Status because of terminal width.

## Unsupported Size

Minimum supported TUI size:

- Width: 60 columns.
- Height: 18 rows.

Below this size, show Unsupported Size instead of broken layouts.

Unsupported Size copy:

```text
Leetgo needs a little more room.

Current: 48x14
Minimum: 60x18

Resize your terminal or use CLI commands:
leetgo next
leetgo info .
leetgo test .
```

Rules:

- Do not crash.
- Do not render broken panels.
- Allow `q` to quit.
- Show useful CLI alternatives.

## Density And Scrolling

RPG Theme should use medium density.

Rules:

- Dashboard primary tile gets the most space.
- Secondary actions are compact.
- Practice Log is capped by default.
- Problem Brief shows concise text before optional hints/extra detail.
- Roadmap Detail can scroll/zoom instead of fitting every node.

Screens that may scroll:

- Roadmap Detail.
- Stage Detail.
- Problem Detail.
- Practice Log.
- Roadmap Completion if summary grows.

Screens that should usually avoid scrolling:

- Dashboard.
- Onboarding steps.

If content scrolls:

- Footer must show scroll keytips.
- Focused item must remain visible.
- Primary action should not start below the fold.

## Shared TUI Components

Add a small shared view layer, not a large framework.

Useful components/helpers:

- `Panel`
- `PixelFrame`
- `StatusPill`
- `ProgressBar`
- `KeytipFooter`
- `RewardMoment`
- `UnsupportedSize`
- `SymbolSet`

Keep these simple and testable. Avoid introducing a complex component framework.

## Onboarding Styling

RPG Theme target: character creation plus choose your first Roadmap.

Presentation:

- Welcome: character creation.
- Display Name: character name.
- Language/workspace: loadout.
- Roadmap Selection: choose questline.
- Theme Selection: choose interface style.
- LeetCode Session: connect judge gate.
- Final screen: first quest unlocked.

Theme Selection should include a small live preview:

- sample Next Action tile.
- sample Status labels.
- sample progress bar.
- sample keytip footer.
- symbol rendering preview.

Symbol preview:

```text
Can you see these symbols?
✓ ⚠ 🔒 ★ ↻
```

Theme Selection should include a keytip to switch to Plain Symbols if Rich Symbols render incorrectly.

## MVP Boundaries

In the first styling phase:

- Polish RPG Theme first.
- Build shared components before full screen redesign.
- Implement Rich/Plain Symbols.
- Implement Motion Preference config and no-motion behavior for non-TTY CLI output where feasible.
- Implement Unsupported Size.
- Implement Reward Moment presentation foundation.
- Add Theme preview in Onboarding.

Deferred:

- Mouse interaction.
- Search/filter interaction.
- Full Appearance screen.
- Perfect graph edge layout.
- Deep internal persistence renames unrelated to presentation.
