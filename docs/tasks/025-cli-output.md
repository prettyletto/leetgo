# Task 025: CLI Test and Submit Output

## Goal

Update `leetgo test` and `leetgo submit` CLI output to use clear multi-line format showing Problem identity, Status, Reward, and next action.

## Dependencies

- Task 018: Verified Status
- Task 019: Reward Events
- Task 020: Problem Manifest

## Likely Files

- `cmd/leetgo/main.go` — update `runProblemTests` and `exportDataToGit`/submit handlers
- `cmd/leetgo/main_test.go`

## Implementation Notes

### `leetgo test .` output

On pass (first time):

```
Leetgo TestSuite passed for #1 Two Sum
Status: Verified
Reward: +70 XP
Next: run `leetgo submit .` for LeetCode confirmation and bonus XP
```

On pass (already Verified/Solved):

```
Leetgo TestSuite passed for #1 Two Sum
Status: Verified
Reward: already claimed
Next: run `leetgo submit .`
```

On fail:

```
Leetgo TestSuite failed for #1 Two Sum
Status: InProgress
<test output>
```

### `leetgo submit .` output

On Accepted (first time):

```
LeetCode Accepted for #1 Two Sum
Status: Solved
Reward: +30 XP
Runtime: 1 ms
Memory: 2 MB
```

On Accepted (already claimed):

```
LeetCode Accepted for #1 Two Sum
Status: Solved
Reward: already claimed
Runtime: 1 ms
Memory: 2 MB
```

On session expired:

```
Submission unavailable for #1 Two Sum
Status: Verified
Reward: none
Error: Session expired. Run `leetgo auth` and try again.
```

On rejected:

```
LeetCode rejected #1 Two Sum
Status: Verified
Wrong Answer (50/63 tests passed)
```

### Implementation

- Use `fmt.Printf` with clear section breaks.
- Read Reward Events to determine "already claimed" vs new award.
- Use manifest resolution from Task 020 for `.` argument.

## Acceptance Criteria

- `leetgo test .` shows multi-line output with Problem ID, Status, Reward, and next action.
- `leetgo submit .` shows multi-line output with Problem ID, Status, Reward, Runtime/Memory.
- Already-claimed rewards show "already claimed" instead of awarding again.
- Session errors show clear guidance to run `leetgo auth`.
- Tests cover output format for pass/fail/already-claimed/session-error cases.

## Verification

```bash
gofmt -w cmd/leetgo
go vet ./...
go test -count=1 ./...
```
