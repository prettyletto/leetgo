# Learning System Product Spec

This document defines the target learning experience from Onboarding through Roadmap Completion. It is the product source of truth for the next construction phase of Leetgo's guided practice system.

Read with:

- `CONTEXT.md`
- `docs/adr/0001-dag-roadmap.md`
- `docs/adr/0007-dashboard-first-tui.md`
- `docs/adr/0009-screen-based-tui.md`
- `docs/adr/0010-local-first-verification-and-reward-events.md`
- `docs/product/onboarding-dashboard.md`
- `docs/product/tui-screen-spec.md`

## Product Promise

Leetgo should feel like a guided learning companion, not a catalog browser. The app should explain what to do next, why that action matters, what it unlocks, and how it fits into the selected Roadmap.

The core loop is:

1. Choose a Roadmap during Onboarding.
2. Start the recommended Problem.
3. Work locally until Verified.
4. Submit to LeetCode or choose Manual Solve.
5. Reach Solved.
6. Unlock dependent Problems.
7. Continue through the Unlock Path.
8. Repair Weaknesses through Review Cycles.
9. Reach Roadmap Completion.

## Progression Model

The Roadmap remains a DAG. Problems are unlocked through prerequisite edges, not by Stage gates.

Status meaning:

- `Locked`: one or more Blockers remain.
- `Available`: all prerequisites are Solved in the selected Roadmap.
- `InProgress`: the user has Started the Problem.
- `Verified`: the local TestSuite passed, but the Problem is not Solved.
- `Solved`: the Problem was completed through Accepted Solve or Manual Solve.

`Verified` is local confidence only. It can earn local XP, but it does not unlock dependents.

`Accepted Solve` is a Solve backed by an Accepted LeetCode Submission. It unlocks dependents and earns Submission XP when eligible.

`Manual Solve` is a user-confirmed Solve without an Accepted Submission. It unlocks dependents, earns no XP, and can later be upgraded by an Accepted Submission.

Manual Solve requires confirmation in both TUI and CLI.

Confirmation copy:

`Mark as manually solved? This will unlock dependent Problems, but you will not earn XP unless LeetCode accepts a Submission later.`

Manual Solve may include an optional note that appears in the Practice Log.

## Recommendation Model

The Dashboard shows one primary Next Action plus a small ranked set of secondary options.

Primary Next Action priority:

1. Verified Problem needing Submission or Manual Solve.
2. Recent InProgress Problem.
3. Critical Review if a Weakness blocks current progression.
4. Newly unlocked or best Available Problem.
5. Regular Review.
6. Other Available Problems.
7. Maintenance actions such as Connect LeetCode or Roadmap Completion.

MVP Next Action kinds:

- `Start`
- `Continue`
- `Submit`
- `ManualSolve`
- `Review`
- `ConnectLeetCode`
- `ViewRoadmapCompletion`

Every Next Action must include a Recommendation Reason. The reason is user-facing text backed by a structured reason type.

MVP reason types:

- `UnlocksDependent`
- `StrengthensPracticeFocus`
- `CompletesVerified`
- `ContinuesInProgress`
- `RepairsWeakness`
- `ValidatesManualSolve`
- `CompletesRoadmap`

Do not show ranking scores to users.

The recommendation engine may use internal ranking, but the interface should explain decisions in plain language.

Example:

`Recommended: Valid Anagram`

`Why: strengthens frequency counting before Group Anagrams.`

## Gradual Learning

Gradual progression comes from denser prerequisite edges and recommendation ranking, not Stage gates.

Recommended behavior:

- Prefer newly unlocked dependents when they are not a difficulty spike.
- Prefer current Stage when options are otherwise similar.
- Prefer lower Difficulty before higher Difficulty inside the same learning area.
- Prefer direct unlock impact over indirect unlock impact.
- Use indirect dependents lightly as a motivational ranking signal.
- Avoid recommending a Hard Problem while easier related Available Problems can prepare the same Category or Practice Focus.

Users may manually start any Available Problem, even if it is not recommended. The app should preserve guidance without blocking.

Example:

`Starting Trapping Rain Water.`

`Note: Leetgo recommends Two Pointers foundations first because this is a difficulty jump.`

Locked Problems cannot be started within the guided Roadmap flow.

## Roadmap And Stage UX

The Roadmap is stored as a DAG but rendered as a tree-like Unlock Path for comprehension.

Zoom model:

- `Roadmap Detail`: zoomed out view of Stages and overall progress.
- `Stage Detail`: mid zoom into one Stage, its Problems, and Blockers.
- `Problem Detail`: zoomed in view of one Problem, its prerequisites, dependents, learning content, and actions.

`Stage` remains a UX grouping and display order. Stage order affects display, summaries, and recommendation tie-breaking. Stage order does not unlock Problems.

Roadmap Detail should default to Stage-grouped Unlock Path, not a raw graph.

Roadmap Detail shows:

- Roadmap title, promise, and Roadmap Time Estimate.
- Stage progress.
- Primary Next Action for the Roadmap.
- Active Review Cycle markers.
- Coming Soon Problems with one or two Blockers.

Stage Detail shows:

- Stage title and description.
- Solved, Verified, InProgress, Available, and Locked counts.
- Recommended Problem inside the Stage.
- Recommendation Reason.
- Locked Problems with missing Blockers.
- Review Problems visually distinct from progression Problems.

Problem Detail shows:

- Problem identity: LeetCode ID, title, Difficulty, Stage, Category.
- Status and evidence: local TestSuite result and LeetCode Submission result.
- Problem Brief.
- Problem Time Estimate.
- Prerequisites and Blockers.
- Direct dependents under `Unlocks`.
- Capped indirect dependents under `Builds toward`.
- Latest Practice Log summary.
- Primary and secondary actions.

## Problem Brief

Leetgo includes its own Problem Brief. It must not copy LeetCode's full Problem statement verbatim.

Problem Brief sections:

- `Summary`: plain-language task.
- `Practice Focus`: the narrow skill the Problem trains.
- `Why now`: why this Problem appears at this point in the Roadmap.
- `Unlock Impact`: authored explanation of what this prepares.
- `Starter Hints`: progressive hints hidden until requested.

Default view shows Summary, Practice Focus, Why now, and Unlock Impact.

Hints are progressive disclosure. Hints are free, do not affect XP, and do not block Accepted Solve.

Problem Brief and Practice Focus are Roadmap-specific. The same global Problem may teach different things in different Roadmaps.

Global Problem identity should stay minimal:

- LeetCode ID.
- Title.
- Slug.
- Difficulty.

Roadmap-specific learning metadata:

- Stage.
- Category.
- Practice Focus.
- Problem Brief.
- Problem Time Estimate.
- Prerequisites.

`Unlocks` should be computed from the DAG. Authored `Unlock Impact` explains the learning value.

## Time And Practice History

Leetgo distinguishes Attempt duration from Solve Duration.

`Attempt` records local test or Submission work with result and duration.

`Solve Duration` is the total recorded practice time spent on a Problem before Solve, calculated from Attempts rather than wall-clock time.

Manual time logging is not MVP.

Attempt kinds:

- Local Attempt: `leetgo test`, pass/fail, duration.
- Submission Attempt: `leetgo submit`, LeetCode result, duration.

Practice Log is the user-facing chronological history for a Problem.

Practice Log includes:

- Start.
- Local Attempts.
- Verification.
- Submission Attempts.
- Manual Solve.
- Accepted Solve.
- Solve Duration.
- Optional notes.

Problem Detail shows a compact latest Practice Log summary by default and a full Practice Log behind an action.

Example full log:

```text
10:12 Start
10:24 Local Attempt failed
10:33 Local Attempt passed -> Verified
10:36 Submission Attempt wrong answer
10:48 Submission Attempt accepted -> Accepted Solve
```

## Submit And Verify Flow

`leetgo submit` runs local tests by default.

Flow:

1. Run local TestSuite.
2. If local tests fail, stop and recommend fixing first.
3. Allow explicit override with `--skip-tests`.
4. If local tests pass, mark or keep Verified.
5. Submit to LeetCode.
6. If accepted, mark Accepted Solve and unlock dependents.
7. If rejected, keep Verified and record the failed Submission Attempt.

Accepted Submission is stronger evidence than local tests. If a user submits with failing local tests and LeetCode accepts, mark Accepted Solve and record the TestSuite mismatch in Practice Log.

Do not let generated TestSuite issues block user progression.

## CLI Learning Surface

`leetgo info` mirrors Problem Detail in CLI form.

When run inside a Problem workspace, it shows:

- Problem identity.
- Status.
- Problem Brief.
- Problem Time Estimate.
- Prerequisites.
- Blockers.
- Unlocks and Builds toward.
- Latest Practice Log summary.
- Relevant commands.

When run outside a Problem workspace, it should accept a Problem ID or slug. If no Problem is specified, it may show the current Dashboard summary.

`leetgo next` prints the same primary Next Action as the Dashboard.

`leetgo next --all` may show the ranked set.

`leetgo next --start` executes only when the primary Next Action is `Start`. If the primary action is Submit, Review, or another kind, it should explain the mismatch and print the correct command.

`leetgo solve --manual` requires confirmation by default and supports `--yes` for scripts.

Suggested commands:

```bash
leetgo info
leetgo info group-anagrams
leetgo next
leetgo next --all
leetgo next --start
leetgo submit
leetgo submit --skip-tests
leetgo solve --manual
leetgo solve --manual --note "Solved in browser"
leetgo solve --manual --yes
```

## Review And Weakness Repair

Review is part of MVP.

Review is a learning action, not a Status. It revisits a previously Solved or difficult Problem to strengthen retention, repair a Weakness, or prepare for a blocked dependent.

Review does not block Roadmap Completion.

Review recommendations may be created by:

- Weakness detection.
- High failed-attempt count.
- Manual Solve needing Accepted validation.
- Important prerequisite refresh before a dependent.

Review Cycle is the bounded reward opportunity for Review XP.

Review XP rules:

- Small XP only.
- Earned at most once per Review Cycle.
- Not awarded for merely opening a Problem.
- Completion requires proof of effort.

Completion proof:

- Accepted Solve Problem: local TestSuite pass during Review completes the cycle.
- Manual Solve Problem: Accepted Submission is preferred and upgrades confidence; local TestSuite pass may complete a weaker cycle.
- Weakness repair: local TestSuite pass on the recommended Review Problem completes that cycle.

Review Cycles are mostly global to the user and Problem, with Roadmap context attached to the reason.

## Onboarding And LeetCode Session

Onboarding should explain the learning loop:

`Start a Problem -> pass local tests -> submit to LeetCode or Manual Solve -> unlock next Problems`

Roadmap Selection is required. The Default Roadmap is focused and easy to confirm.

Roadmap cards show:

- Title.
- Promise.
- Audience.
- Size.
- Difficulty mix.
- Roadmap Time Estimate.
- First Stages.
- Highlights.

Roadmap Time Estimate is a broad commitment range, not a precise completion date.

Onboarding should ask whether to connect LeetCode Session.

Copy:

`Accepted Submissions unlock Roadmap progress and XP. You can skip and use Manual Solve, but Manual Solve earns no XP.`

Skipping Session is allowed. Dashboard should remind only when relevant, especially when Verified Problems are waiting for Submission.

After Roadmap confirmation, Onboarding should show the first Next Action and Recommendation Reason with actions to Start now or go to Dashboard.

## Roadmap Completion

Roadmap Completion is reached when every Problem in a Roadmap is Solved. Manual Solve counts for completion.

Completion summary includes:

- Problems Solved.
- Accepted Solves.
- Manual Solves.
- Total XP.
- Total Solve Duration.
- Strongest Categories or Practice Focuses.
- Weaknesses to review.
- Active Review Cycles.
- Suggested next Roadmap.

Completion is not blocked by active Review Cycles.

Suggested next Roadmap should be catalog-curated first, then ranked by progress and Weaknesses.

Example:

`Recommended next: Interview Sprint`

`Why: you completed From Zero to Hero and still have Weaknesses in Sliding Window and Graphs.`

## Catalog Requirements

Roadmap metadata should support:

- `promise`
- `audience`
- `roadmap_time_estimate`
- `highlights`
- `recommended`
- `next_roadmaps`

Roadmap Problem metadata should support:

- `practice_focus`
- `problem_time_estimate`
- `summary`
- `why_now`
- `unlock_impact`
- `hints`

Computed from the DAG:

- Direct unlocks.
- Indirect builds-toward list.
- Blockers.
- Coming Soon Problems.

Validation should reject malformed Roadmaps early.

## MVP Boundaries

In MVP:

- One primary Category per Problem.
- One Stage per Problem.
- No optional Roadmap Problems.
- No Mastery domain term.
- No manual time logging.
- No ranking scores in UI.
- No starting Locked Problems inside the guided Roadmap flow.

Deferred:

- Reset stale InProgress Problems to Available.
- Optional Problems that do not count toward completion.
- Manual time logging.
- Practice Focus-level Weakness implementation if data is not ready.
- TestSuite quality issue workflow.
- `leetgo why` separate command.
