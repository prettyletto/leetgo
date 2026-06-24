# Task 033: `leetgo next` CLI

## Goal

Add `leetgo next` so CLI users get the same primary Next Action and Recommendation Reason as the Dashboard.

## Context

Read:

- `CONTEXT.md`
- `docs/product/learning-system.md`
- `docs/tasks/031-next-action-learning-ranking.md`

The command should explain guidance by default and execute only with explicit flags.

## Dependencies

- Task 031: Next Action Learning Ranking
- Existing Start behavior

## Likely Files

- `cmd/leetgo/`
- `internal/recommendation/`
- `internal/workspace/`

## Implementation Notes

Commands:

```bash
leetgo next
leetgo next --all
leetgo next --start
```

Behavior:

- `leetgo next` prints the primary Next Action and Recommendation Reason.
- `leetgo next --all` prints the ranked set.
- `leetgo next --start` executes only if the primary action is `Start`.
- If primary action is not Start, print the mismatch and the correct command.

Do not add `--do` or vague execution flags.

## Acceptance Criteria

- `leetgo next` matches Dashboard primary action for the same store/catalog state.
- `leetgo next --all` shows ranked actions without exposing scores.
- `leetgo next --start` starts the primary Problem only when the primary action is Start.
- `leetgo next --start` does not skip Submit, ManualSolve, Review, or ConnectLeetCode actions.
- Output includes Recommendation Reason.
- Tests cover Start and non-Start primary actions.

## Verification

Run:

```bash
gofmt -w cmd internal/recommendation
go test ./...
```
