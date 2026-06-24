# Learning System Audit

Audit date: 2026-06-23

Scope:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/adr/0010-local-first-verification-and-reward-events.md`
- Tasks `028` through `038`
- Current implementation in `cmd/leetgo`, `internal/roadmap`, `internal/recommendation`, `internal/store`, `internal/tui`, `internal/catalog`, `internal/analytics`, and `internal/gamification`

## Result

The current implementation contains substantial learning-system work: solved-gated progression, solve provenance, Practice Log surfaces, catalog learning metadata, `leetgo next`, `leetgo info`, Review Cycles, Review XP, Onboarding Session copy, and Roadmap Completion summary behavior.

During the audit and follow-up fix pass, the behavioral findings were fixed directly:

- CLI local test and submit flows no longer downgrade already Solved Problems to Verified.
- CLI and TUI local/submission attempts now record Attempt entries so Practice Log and Solve Duration have usable source data.
- Dashboard Roadmap Completion now opens a real Roadmap Completion screen instead of a placeholder notification.
- Recommendation calculation no longer creates Review Cycles as a side effect.
- Dashboard Review activation creates the Review Cycle when the user chooses the Review action.
- TUI Submission now runs the local TestSuite first and stops before LeetCode Submission when local tests fail.
- TUI Submission now offers a typed `Submit anyway` confirmation after local TestSuite failure.

## Verification

Command run:

```bash
go test ./...
```

Result:

```text
ok   github.com/prettyletto/leetgo/cmd/leetgo
ok   github.com/prettyletto/leetgo/internal/analytics
ok   github.com/prettyletto/leetgo/internal/catalog
ok   github.com/prettyletto/leetgo/internal/config
ok   github.com/prettyletto/leetgo/internal/gamification
ok   github.com/prettyletto/leetgo/internal/generator
ok   github.com/prettyletto/leetgo/internal/gitexport
ok   github.com/prettyletto/leetgo/internal/leetcode
ok   github.com/prettyletto/leetgo/internal/recommendation
ok   github.com/prettyletto/leetgo/internal/roadmap
ok   github.com/prettyletto/leetgo/internal/store
ok   github.com/prettyletto/leetgo/internal/tui
ok   github.com/prettyletto/leetgo/internal/tui/views
ok   github.com/prettyletto/leetgo/internal/workspace
```

## Fixed During Audit

### Solved Problems Could Be Downgraded To Verified

Severity: High

Files:

- `cmd/leetgo/main.go`

Problem:

`leetgo test` and the local-test phase of `leetgo submit` unconditionally marked a passing Problem as Verified. If a Problem was already Solved, a later local test pass or a submit attempt with local tests could downgrade it to Verified before Submission completed.

Why it matters:

The learning-system spec says Solved is the Roadmap progression state. Downgrading Solved to Verified can reintroduce Blockers, hide completion, and make Roadmap Completion incorrect.

Fix:

Passing local tests now mark a Problem Verified only when the current progress is not already Solved. CLI output preserves `Status: Solved` for already Solved Problems.

### Attempts Were Not Recorded For Core Practice Actions

Severity: High

Files:

- `cmd/leetgo/main.go`
- `internal/tui/model.go`
- `internal/tui/problem_detail_screen.go`

Problem:

Local test and Submission flows did not consistently record Attempts. Practice Log already rendered Attempts, and Solve Duration is defined as total recorded Attempt duration, but the main user actions did not create those records.

Why it matters:

Without Attempt records, Practice Log and Solve Duration are incomplete and Review/Weakness signals become weaker.

Fix:

CLI and TUI local test runs now record local Attempts with pass/fail and duration. CLI and TUI Submission runs now record Submission Attempts with accepted/rejected pass status and duration.

## Resolved Follow-Up Findings

### Roadmap Completion TUI Action Is Still A Placeholder

Severity: Medium, resolved

File:

- `internal/tui/dashboard_screen.go`

Original finding:

The recommendation engine can emit `ViewRoadmapCompletion`, but Dashboard activation currently shows a notification saying completion details are coming later instead of opening a real completion screen.

Impact:

CLI `leetgo status` can show a completion summary, but the TUI milestone experience is incomplete.

Fix:

Added `RoadmapCompletionScreen`, wired `ScreenCompletion` through root navigation, and changed Dashboard activation to navigate to the completion screen. Added tests for completion rendering, back navigation, and Dashboard completion activation.

### Recommendation Calculation Creates Review Cycles As A Side Effect

Severity: Medium, resolved

File:

- `internal/recommendation/recommendation.go`

Original finding:

`Calculator.Calculate` creates Review Cycles while computing Review actions. This means a read-style operation mutates persistence.

Impact:

The current logic avoids obvious duplicates by checking existing open cycles, but side effects during calculation make behavior harder to reason about and test. A future UI refresh could create learning records before the user has seen or accepted the recommendation.

Fix:

Removed Review Cycle creation from `Calculator.Calculate`. Dashboard now creates a Review Cycle only when the user activates a Review action. Added tests proving calculation does not create cycles and Review activation does.

### TUI Submit Does Not Run Local Tests First

Severity: Low, resolved

File:

- `internal/tui/problem_detail_screen.go`

Original finding:

CLI `leetgo submit` runs local tests by default, but Problem Detail's TUI submit action submits directly to LeetCode.

Impact:

CLI and TUI behavior differed. The product spec says submit should run local tests by default.

Fix:

Problem Detail submit now runs the local TestSuite before Submission. Local failure records a failed Attempt, stops before LeetCode Submission, shows the failure output, and offers a typed `Submit anyway` confirmation. Local pass records a passing Attempt, marks Verified when appropriate, and proceeds to Submission.

## Remaining Notes

### Practice Log Still Uses SolveLog Persistence Names Internally

Severity: Low

Files:

- `internal/store/store.go`
- `internal/store/sqlite.go`
- `internal/tui/solve_log_screen.go`

Finding:

User-facing copy uses Practice Log, but some internal types and table names remain `SolveLog` / `solve_logs`.

Impact:

This is acceptable under the task notes because a broad rename would be larger than necessary. The risk is terminology drift for future agents.

Recommendation:

Keep user-facing copy as Practice Log. If persistence is redesigned later, rename internal types in one focused migration/refactor.

## Coverage Notes

Confirmed by code and tests:

- Verified does not unlock dependents.
- Manual Solve unlocks dependents through Solved progress and awards no XP.
- Accepted Solve records accepted provenance and awards eligible XP.
- Verified prerequisites remain Blockers because only Solved progress is included in unlock maps.
- `leetgo next --start` refuses to execute non-Start primary actions.
- `leetgo info` shows Problem Brief, Practice Focus, Blockers, Unlocks, Builds toward, and Practice entries.
- Review XP is idempotent through Review Cycle reward state.
- Roadmap Completion is not blocked by active Review Cycles.
- Dashboard Roadmap Completion activation opens a real completion screen.
- Recommendation calculation is side-effect free for Review Cycle creation.
- TUI submit runs local tests first and stops before Submission when they fail.
- TUI submit-anyway override is behind typed confirmation after local TestSuite failure.

Missing or partial coverage:

- CLI downgrade prevention is covered indirectly by passing tests after implementation, but should get explicit tests around already Solved Problems passing local tests and submit-local-tests.
- CLI downgrade prevention is still better covered by behavior-level tests than explicit command-level regression tests.

## Residual Risk

The main remaining product risk is internal terminology drift from `SolveLog` persistence names. User-facing surfaces use Practice Log, but a future persistence refactor should rename internal types and tables in one focused migration if the compatibility cost is justified.
