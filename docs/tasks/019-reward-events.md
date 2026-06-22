# Task 019: Reward Events Persistence and XP Idempotency

## Goal

Add Reward Events as a persisted record of XP awards so that local Verify (70%) and Accepted Submission (30%) each award XP exactly once, and manual Solve awards no XP.

## Dependencies

- Task 018: Verified Status

## Likely Files

- `internal/store/store.go` — add `RewardEvent` type and interface methods
- `internal/store/sqlite.go` — implement Reward Events table and queries
- `internal/store/migrations/004_reward_events.sql` — new migration
- `internal/store/export.go` — include Reward Events in export/import
- `internal/store/sqlite_test.go` — tests for idempotency
- `internal/tui/problem_detail_screen.go` — use Reward Events when awarding XP
- `cmd/leetgo/main.go` — use Reward Events in CLI test/submit

## Implementation Notes

### Reward Event type

```go
type RewardEvent struct {
    ProblemID int
    Kind      string // "verify" or "submit"
    XP        int
    CreatedAt time.Time
}
```

### Store interface additions

```go
RecordRewardEvent(ctx context.Context, event *RewardEvent) error
HasRewardEvent(ctx context.Context, problemID int, kind string) (bool, error)
GetRewardEvents(ctx context.Context, problemID int) ([]*RewardEvent, error)
```

### XP split

- Verify reward: `store.XPForDifficulty(difficulty) * 70 / 100`
- Submit reward: `store.XPForDifficulty(difficulty) * 30 / 100`
- Manual Solve: 0 XP

### Idempotency

Before awarding XP, check `HasRewardEvent(problemID, kind)`. If true, skip award and show "already claimed" message.

### Migration `004_reward_events.sql`

```sql
CREATE TABLE IF NOT EXISTS reward_events (
    problem_id INTEGER NOT NULL,
    kind TEXT NOT NULL,
    xp INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (problem_id, kind)
);
```

## Acceptance Criteria

- `reward_events` table exists with composite primary key `(problem_id, kind)`.
- `RecordRewardEvent` is idempotent via `INSERT OR IGNORE` or `HasRewardEvent` check.
- Verify awards 70% of difficulty XP exactly once.
- Submit awards 30% of difficulty XP exactly once.
- Manual Solve awards 0 XP.
- Re-running local tests on a Verified Problem shows "already claimed".
- Re-submitting an Accepted Problem shows "already claimed".
- Export/import round-trips Reward Events.
- Tests cover all idempotency paths.

## Verification

```bash
gofmt -w internal/store internal/tui cmd/leetgo
go vet ./...
go test -count=1 ./...
```
