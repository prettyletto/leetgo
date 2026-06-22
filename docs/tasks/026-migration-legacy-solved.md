# Task 026: Migration for Existing Solved Problems

## Goal

Migrate existing `Solved` Problems so they cannot earn duplicate Verify or Submit XP under the new Reward Events model.

## Dependencies

- Task 018: Verified Status
- Task 019: Reward Events

## Likely Files

- `internal/store/migrations/004_reward_events.sql` — migration already creates table
- `internal/store/sqlite.go` — add migration logic or post-migration hook
- `internal/store/sqlite_test.go` — migration tests

## Implementation Notes

### Migration strategy

Existing `Solved` Problems have already earned full XP under the old model. Under the new model:

- They remain `Solved` (Status does not change).
- Their existing XP total remains unchanged.
- Create synthetic Reward Events to prevent duplicate future awards:
  - `INSERT OR IGNORE INTO reward_events (problem_id, kind, xp) VALUES (?, 'verify', 0)`
  - `INSERT OR IGNORE INTO reward_events (problem_id, kind, xp) VALUES (?, 'submit', 0)`
- The `xp = 0` on synthetic events signals they are legacy guards, not real awards.

### When to run

Run migration once when the `reward_events` table is first created. Check for existing Solved Problems and insert guards.

### Implementation

In `NewSQLiteStore` or a dedicated migration function:

```go
func (s *SQLiteStore) migrateLegacySolved(ctx context.Context) error {
    rows, err := s.db.QueryContext(ctx,
        "SELECT problem_id FROM progress WHERE status = 'solved'")
    // for each row:
    // INSERT OR IGNORE INTO reward_events (problem_id, kind, xp, created_at)
    // VALUES (?, 'verify', 0, CURRENT_TIMESTAMP)
    // INSERT OR IGNORE INTO reward_events (problem_id, kind, xp, created_at)
    // VALUES (?, 'submit', 0, CURRENT_TIMESTAMP)
}
```

### Idempotency

`INSERT OR IGNORE` ensures running migration multiple times is safe.

## Acceptance Criteria

- Existing Solved Problems get both `verify` and `submit` Reward Events with `xp = 0`.
- Running migration multiple times is safe (idempotent).
- Existing XP total is unchanged.
- Future `leetgo test` on old Solved Problems shows "already claimed".
- Future `leetgo submit` on old Solved Problems shows "already claimed".
- Tests cover migration with existing Solved Problems.
- Tests cover migration idempotency.

## Verification

```bash
gofmt -w internal/store
go vet ./...
go test -count=1 ./...
```
