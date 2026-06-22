# Task 022: Problem Detail Action Model

## Goal

Rework Problem Detail so `enter` changes by Status, editor launches detached, local TestSuite pass auto-Verifies, and Submission promotes to Solved.

## Dependencies

- Task 018: Verified Status
- Task 019: Reward Events

## Likely Files

- `internal/tui/problem_detail_screen.go` — rework enter, editor launch, test/submit handlers
- `internal/tui/problem_detail_screen_test.go`
- `internal/workspace/workspace.go` — may need detached editor helper

## Implementation Notes

### Enter by Status

| Status | Primary Action | Keytip |
|--------|---------------|--------|
| `Available` | Start | `enter start` |
| `InProgress` | Open | `enter open` |
| `Verified` | Submit | `enter submit` |
| `Solved` | Open | `enter open` |
| `Locked` | No action | `enter locked` |

### Start behavior

1. Generate Stub and TestSuite.
2. Write `.leetgo-problem.toml` manifest.
3. Set Status to `InProgress`.
4. Launch editor detached with Stub and TestSuite paths.
5. Show feedback: "Started. Files opened in editor."

### Detached editor launch

- Use configured `editor`, `$VISUAL`, or `$EDITOR`.
- Launch with `cmd.Start()` only; do not attach stdin/stdout/stderr.
- If no editor configured, fallback to `xdg-open <problem-dir>` on Linux.
- Leetgo stays alive on Problem Detail.

### Local TestSuite pass (key `x`)

1. Run `leetgo test` equivalent for the Problem's language.
2. If tests pass and Status is `InProgress`:
   - Set Status to `Verified`.
   - Record Verify Reward Event (70% XP).
   - Award XP if not already claimed.
   - Show: "Verified locally. +N XP."
3. If tests pass and Status is already `Verified` or `Solved`:
   - Show: "Tests passed. Reward already claimed."
4. If tests fail:
   - Show failure output.
   - Do not change Status.

### Submit (key `s` or `enter` when Verified)

1. Read Stub file.
2. Submit to LeetCode via Session.
3. If Accepted:
   - Set Status to `Solved`.
   - Record Submit Reward Event (30% XP).
   - If Verify Reward Event was not claimed (submitted from InProgress), also claim it.
   - Show: "Accepted by LeetCode. +N XP."
4. If rejected:
   - Write Solve Log with status.
   - Do not change Status.
   - Show failure details.
5. If unauthenticated/session expired:
   - Show: "Session expired. Run `leetgo auth` and try again."

## Acceptance Criteria

- `enter` on Available Problem Starts, generates files, opens editor detached.
- `enter` on InProgress Problem opens existing files.
- `enter` on Verified Problem submits to LeetCode.
- `enter` on Solved Problem opens existing files.
- `enter` on Locked Problem shows error.
- Local TestSuite pass from InProgress sets Verified and awards 70% XP.
- Local TestSuite pass from Verified/Solved shows "already claimed".
- Accepted Submission sets Solved and awards 30% XP (plus 70% if Verify was unclaimed).
- Failed Submission writes Solve Log and does not change Status.
- Editor launches detached; Leetgo stays alive.
- Tests cover all Status transitions and reward idempotency.

## Verification

```bash
gofmt -w internal/tui internal/workspace
go vet ./...
go test -count=1 ./...
```
