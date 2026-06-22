# Task 023: Manual Solve Typed Confirmation

## Goal

Make manual Solve require typed `SOLVE` confirmation to prevent accidental no-XP bypasses.

## Dependencies

- Task 018: Verified Status

## Likely Files

- `internal/tui/problem_detail_screen.go` — add confirmation state
- `internal/tui/problem_detail_screen_test.go`

## Implementation Notes

### Current behavior

Pressing `m` immediately marks Problem Solved and awards XP.

### New behavior

1. Press `m` to initiate manual Solve.
2. Problem Detail enters confirmation mode:
   - Shows: `Manual Solve bypasses verification and awards no XP. Type SOLVE to confirm.`
   - Shows text input field.
3. User types `SOLVE` (case-sensitive).
4. On match:
   - Set Status to `Solved`.
   - Award 0 XP.
   - Show: "Manually marked Solved. No XP awarded."
5. `esc` cancels confirmation and returns to normal Problem Detail.
6. Any other text is ignored.

### UI state

Add a `manualSolveMode bool` field to `ProblemDetailScreen`. When true:
- Key input goes to the confirmation text field instead of normal keybindings.
- Footer shows `type SOLVE to confirm`, `esc to cancel`.

### Disabled states

- If Problem is already `Solved`, pressing `m` shows "Already Solved."
- If Problem is `Locked`, pressing `m` shows "Problem is locked."

## Acceptance Criteria

- Pressing `m` enters confirmation mode with clear no-XP warning.
- Typing `SOLVE` marks Problem Solved with 0 XP.
- Typing anything else does nothing.
- `esc` cancels confirmation.
- Already Solved Problems show "Already Solved."
- Locked Problems show "Problem is locked."
- Tests cover confirmation flow, cancel, already solved, and locked states.

## Verification

```bash
gofmt -w internal/tui
go vet ./...
go test -count=1 ./...
```
