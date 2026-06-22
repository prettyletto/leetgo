# Task 027: Dashboard Next Action Ranking for Verified

## Goal

Update Dashboard Next Action ranking to include Verified Problems as "Submit" actions, ranked between Continue InProgress and Start Available.

## Dependencies

- Task 018: Verified Status
- Task 019: Reward Events
- Task 021: Dashboard Action Model

## Likely Files

- `internal/recommendation/recommendation.go` — add `submitActions`
- `internal/recommendation/recommendation_test.go`
- `internal/tui/dashboard_screen.go` — handle `KindSubmit` in enter handler
- `internal/tui/dashboard_screen_test.go`

## Implementation Notes

### New ranking order

1. Continue InProgress Problems (`KindContinue`)
2. Submit Verified Problems (`KindSubmit`) — new
3. Start Available Problems (`KindStart`)
4. Maintenance actions (`KindExport`, `KindInspect`)

### New ActionKind

Add `KindSubmit ActionKind = "submit"` to `recommendation.go`.

### submitActions function

```go
func (c *Calculator) submitActions(progress map[int]roadmap.Status, topoIndex map[int]int) []NextAction {
    // Find all Problems with StatusVerified
    // Sort by topoIndex (earlier in Roadmap first)
    // Return NextAction with KindSubmit, Title = problem title,
    // Reason = "Verified locally. Submit to LeetCode for final confirmation and bonus XP."
}
```

### Dashboard enter handler

Add case for `KindSubmit`:

```go
case recommendation.KindSubmit:
    return s, func() tea.Msg {
        return NavigateMsg{ScreenID: ScreenProblemDetail, ProblemID: action.ProblemID}
    }
```

### Dashboard card rendering

`KindSubmit` should render as `Submit` label (capitalized from "submit").

## Acceptance Criteria

- Verified Problems appear as "Submit" Next Actions in Dashboard.
- Submit actions rank after Continue and before Start.
- `enter` on Submit action opens Problem Detail.
- Submit action card shows "Verified locally. Submit to LeetCode for final confirmation and bonus XP."
- Tests cover ranking order with mix of InProgress, Verified, and Available Problems.
- Tests cover Submit action rendering.

## Verification

```bash
gofmt -w internal/recommendation internal/tui
go vet ./...
go test -count=1 ./...
```
