package tui

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/leetcode"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSolveLog(t *testing.T) (*SolveLogScreen, store.Store) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	cfg := &config.Config{
		DisplayName: "Ada",
		Workspace:   t.TempDir(),
		Language:    "go",
		Roadmap:     "from-zero-to-hero",
		Theme:       "rpg-skill-tree",
	}

	theme, err := LookupTheme(cfg.Theme)
	require.NoError(t, err)

	sl := NewSolveLogScreen(cfg, theme, db)
	return sl, db
}

func TestSolveLog_EmptyState(t *testing.T) {
	sl, _ := newTestSolveLog(t)
	sl.width = 120

	view := sl.View()
	assert.Contains(t, view, "No Solve Logs")
	assert.Contains(t, view, "submit")
}

func TestSolveLog_ViewHasFooter(t *testing.T) {
	sl, _ := newTestSolveLog(t)
	sl.width = 120

	view := sl.View()
	assert.Contains(t, view, "j/k")
	assert.Contains(t, view, "esc")
	assert.Contains(t, view, "quit")
}

func TestSolveLog_EscReturnsToDashboard(t *testing.T) {
	sl, _ := newTestSolveLog(t)

	_, cmd := sl.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenDashboard, navigate.ScreenID)
}

func TestSolveLog_Quit(t *testing.T) {
	sl, _ := newTestSolveLog(t)

	_, cmd := sl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	require.NotNil(t, cmd)
}

func TestSolveLog_NewestFirst(t *testing.T) {
	sl, db := newTestSolveLog(t)
	ctx := context.Background()

	older := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-1 * time.Hour)

	require.NoError(t, db.RecordSolveLog(ctx, &store.SolveLogRecord{
		ProblemID:   1,
		Slug:        "two-sum",
		Language:    "golang",
		Status:      "Accepted",
		StatusCode:  10,
		Runtime:     "1 ms",
		Memory:      "2 MB",
		TotalTests:  63,
		PassedTests: 63,
		SubmittedAt: older,
	}))
	require.NoError(t, db.RecordSolveLog(ctx, &store.SolveLogRecord{
		ProblemID:   15,
		Slug:        "3sum",
		Language:    "golang",
		Status:      "Wrong Answer",
		StatusCode:  11,
		TotalTests:  50,
		PassedTests: 30,
		SubmittedAt: newer,
	}))

	sl.refresh()
	assert.Equal(t, 2, len(sl.logs))
	assert.Equal(t, 15, sl.logs[0].ProblemID, "newest should be first")
	assert.Equal(t, 1, sl.logs[1].ProblemID, "older should be second")
}

func TestSolveLog_RendersAccepted(t *testing.T) {
	sl, db := newTestSolveLog(t)
	ctx := context.Background()

	require.NoError(t, db.RecordSolveLog(ctx, &store.SolveLogRecord{
		ProblemID:   1,
		Slug:        "two-sum",
		Language:    "golang",
		Status:      "Accepted",
		StatusCode:  10,
		Runtime:     "1 ms",
		Memory:      "2 MB",
		TotalTests:  63,
		PassedTests: 63,
		SubmittedAt: time.Now(),
	}))

	sl.refresh()
	sl.width = 120
	view := sl.View()
	assert.Contains(t, view, "two-sum")
	assert.Contains(t, view, "Accepted")
	assert.Contains(t, view, "1 ms")
	assert.Contains(t, view, "2 MB")
}

func TestSolveLog_RendersRejected(t *testing.T) {
	sl, db := newTestSolveLog(t)
	ctx := context.Background()

	require.NoError(t, db.RecordSolveLog(ctx, &store.SolveLogRecord{
		ProblemID:   1,
		Slug:        "two-sum",
		Language:    "golang",
		Status:      "Wrong Answer",
		StatusCode:  11,
		TotalTests:  63,
		PassedTests: 50,
		SubmittedAt: time.Now(),
	}))

	sl.refresh()
	sl.width = 120
	view := sl.View()
	assert.Contains(t, view, "Wrong Answer")
	assert.Contains(t, view, "50/63")
}

func TestSolveLog_FocusNavigation(t *testing.T) {
	sl, db := newTestSolveLog(t)
	ctx := context.Background()

	for _, id := range []int{1, 2, 3} {
		require.NoError(t, db.RecordSolveLog(ctx, &store.SolveLogRecord{
			ProblemID:   id,
			Slug:        "test",
			Language:    "golang",
			Status:      "Accepted",
			StatusCode:  10,
			SubmittedAt: time.Now(),
		}))
	}

	sl.refresh()
	assert.Equal(t, 0, sl.focusIndex)
	sl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 1, sl.focusIndex)
	sl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 0, sl.focusIndex)
}

func TestSolveLog_ThemeCycle(t *testing.T) {
	sl, _ := newTestSolveLog(t)

	_, cmd := sl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(ThemeChangedMsg)
	assert.True(t, ok)
}

func TestSolveLog_WindowResize(t *testing.T) {
	sl, _ := newTestSolveLog(t)

	sl.Update(tea.WindowSizeMsg{Width: 150, Height: 50})
	assert.Equal(t, 150, sl.width)
	assert.Equal(t, 50, sl.height)
}

func TestSolveLog_TitleRendered(t *testing.T) {
	sl, _ := newTestSolveLog(t)
	sl.width = 120

	view := sl.View()
	assert.Contains(t, view, "Solve Log")
}

func TestSolveLog_ProblemDetailStillShowsLogs(t *testing.T) {
	pd, db := newTestProblemDetail(t, 1)
	require.NotNil(t, pd)
	require.NotNil(t, db)

	ctx := context.Background()
	require.NoError(t, db.RecordSolveLog(ctx, &store.SolveLogRecord{
		ProblemID:   1,
		Slug:        "two-sum",
		Language:    "golang",
		Status:      "Accepted",
		StatusCode:  10,
		Runtime:     "1 ms",
		Memory:      "2 MB",
		TotalTests:  63,
		PassedTests: 63,
		SubmittedAt: time.Now(),
	}))

	pd.refresh()
	pd.width = 120
	view := pd.View()
	assert.Contains(t, view, "Solve Log")
	assert.Contains(t, view, "Accepted")
}

func TestSolveLog_SubmitResultRecordsLog(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	sd := submitResultMsg{
		problemID: 1,
		slug:      "two-sum",
		language:  "golang",
		result: &leetcode.SubmissionResult{
			Status:      "Accepted",
			StatusCode:  10,
			Runtime:     "0 ms",
			Memory:      "3 MB",
			TotalTests:  63,
			PassedTests: 63,
		},
	}

	sc, _ := pd.Update(sd)
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)

	sl, _ := newTestSolveLog(t)
	sl.refresh()

	assert.Equal(t, roadmap.StatusSolved, pd2.status)
	_ = sl
}
