# Task 024: Dashboard Stats Verified/Solved Split

## Goal

Update Dashboard HUD to show separate Verified and Solved counts, and a combined Progress metric.

## Dependencies

- Task 018: Verified Status

## Likely Files

- `internal/tui/dashboard_screen.go` — update `renderHUD`
- `internal/store/store.go` — may need `GetStats` to include Verified count
- `internal/store/sqlite.go` — update stats query
- `internal/tui/dashboard_screen_test.go`

## Implementation Notes

### Current HUD

```
Level 2
[████░░░░░░] 150/300 XP
Streak: 5 days
Solved: 12/132
Roadmap: from-zero-to-hero
```

### New HUD

```
Level 2
[████░░░░░░] 150/300 XP
Streak: 5 days
Verified: 8
Solved: 12
Progress: 20/132
Roadmap: from-zero-to-hero
```

### Stats struct update

Add `Verified int` to `store.Stats`. Update `GetStats` SQL to count both:

```sql
SELECT
    COUNT(CASE WHEN status = 'verified' THEN 1 END) as verified,
    COUNT(CASE WHEN status = 'solved' THEN 1 END) as solved
FROM progress
```

### Progress calculation

`Progress = Verified + Solved` out of `Total` (total Problems in active Roadmap).

## Acceptance Criteria

- Dashboard HUD shows `Verified: X`, `Solved: Y`, and `Progress: (X+Y)/Total`.
- `store.Stats` includes `Verified` count.
- Verified and Solved counts are mutually exclusive (a Problem is one or the other).
- Tests cover stats with mix of Verified and Solved Problems.
- Tests cover HUD rendering with new stats.

## Verification

```bash
gofmt -w internal/store internal/tui
go vet ./...
go test -count=1 ./...
```
