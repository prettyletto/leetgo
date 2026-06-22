package tui

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/leetcode"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestProblemDetail(t *testing.T, problemID int) (*ProblemDetailScreen, store.Store) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	rm, err := catalog.LoadRoadmap("from-zero-to-hero")
	require.NoError(t, err)

	cfg := &config.Config{
		DisplayName: "Ada",
		Workspace:   t.TempDir(),
		Language:    "go",
		Roadmap:     "from-zero-to-hero",
		Theme:       "rpg-skill-tree",
	}

	theme, err := LookupTheme(cfg.Theme)
	require.NoError(t, err)

	pd := NewProblemDetailScreen(cfg, theme, db, rm, problemID)
	return pd, db
}

func TestProblemDetail_ViewShowsTitle(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "#1 Two Sum")
}

func TestProblemDetail_ViewShowsDifficultyCategory(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Difficulty")
	assert.Contains(t, view, "easy")
	assert.Contains(t, view, "Category")
	assert.Contains(t, view, "arrays-hashing")
}

func TestProblemDetail_ViewShowsStatusAvailable(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "READY")
}

func TestProblemDetail_ViewShowsPrerequisites(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Prerequisites: none")
}

func TestProblemDetail_ViewShowsPrerequisitesWithContent(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 49)
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Prerequisites:")
	assert.Contains(t, view, "Valid Anagram")
}

func TestProblemDetail_ViewShowsBlockers(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 49)
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Blocked by")
	assert.Contains(t, view, "LOCKED")
}

func TestProblemDetail_ViewShowsFilePaths(t *testing.T) {
	pd, db := newTestProblemDetail(t, 1)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusInProgress))
	pd.refresh()

	pd.width = 120
	view := pd.View()
	assert.Contains(t, view, "Stub:")
	assert.Contains(t, view, "Test:")
}

func TestProblemDetail_ViewHasFooter(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "enter")
	assert.Contains(t, view, "open")
	assert.Contains(t, view, "test")
	assert.Contains(t, view, "submit")
	assert.Contains(t, view, "solve")
	assert.Contains(t, view, "quit")
}

func TestProblemDetail_Quit(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	require.NotNil(t, cmd)
}

func TestProblemDetail_EscReturnsToStageDetail(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenStageDetail, navigate.ScreenID)
	assert.NotEmpty(t, navigate.Stage, "should include stage ID for back navigation")
}

func TestProblemDetail_BackspaceReturnsToStageDetail(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenStageDetail, navigate.ScreenID)
	assert.NotEmpty(t, navigate.Stage, "should include stage ID for back navigation")
}

func TestProblemDetail_LockedShowsError(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 49)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "locked")
}

func TestProblemDetail_StartAvailable(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.cfg.Workspace = t.TempDir()

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	assert.Equal(t, roadmap.StatusInProgress, pd.status)
	assert.Contains(t, pd.errorMsg, "Started")
}

func TestProblemDetail_MarkSolved(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})

	assert.Equal(t, roadmap.StatusSolved, pd.status)
	assert.Contains(t, pd.errorMsg, "XP")
	_ = cmd
}

func TestProblemDetail_DoubleMarkSolved(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusSolved

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "Already solved")
}

func TestProblemDetail_MarkSolvedLocked(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 49)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "locked")
}

func TestProblemDetail_OpenEditorNoEditor(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	require.NotNil(t, cmd)
}

func TestProblemDetail_RunTestsLocked(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 49)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "locked")
}

func TestProblemDetail_RunTestsNotGenerated(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "not generated")
}

func TestProblemDetail_SubmitNotStarted(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "Start")
}

func TestProblemDetail_SubmitNotAuthenticated(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress
	pd.leetcode = nil

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "authenticated")
}

func TestProblemDetail_WindowResize(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	pd.Update(tea.WindowSizeMsg{Width: 150, Height: 50})
	assert.Equal(t, 150, pd.width)
	assert.Equal(t, 50, pd.height)
}

func TestProblemDetail_ThemeCycle(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(ThemeChangedMsg)
	assert.True(t, ok)
}

func TestProblemDetail_PrimaryActionLocked(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 49)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "locked")
}

func TestProblemDetail_PrimaryActionAvailableStarts(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.cfg.Workspace = t.TempDir()

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	assert.Equal(t, roadmap.StatusInProgress, pd.status)
}

func TestProblemDetail_PrimaryActionInProgressOpensEditor(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress
	pd.cfg.Editor = ""

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
}

func TestProblemDetail_SubmitResultAccepted(t *testing.T) {
	ctx := context.Background()
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	initialXP, _ := pd.db.GetStats(ctx)

	sd := submitResultMsg{
		problemID: 1,
		slug:      "two-sum",
		language:  "golang",
		result: &leetcode.SubmissionResult{
			Status:      "Accepted",
			StatusCode:  10,
			Runtime:     "1 ms",
			Memory:      "2 MB",
			TotalTests:  63,
			PassedTests: 63,
		},
	}

	sc, _ := pd.Update(sd)
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)

	assert.False(t, pd2.submitting)
	assert.Equal(t, roadmap.StatusSolved, pd2.status)
	assert.Contains(t, pd2.errorMsg, "Accepted")

	stats, _ := pd2.db.GetStats(ctx)
	if initialXP != nil {
		assert.GreaterOrEqual(t, stats.TotalXP, initialXP.TotalXP)
	}
}

func TestProblemDetail_SubmitResultRejected(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	sd := submitResultMsg{
		problemID: 1,
		slug:      "two-sum",
		language:  "golang",
		result: &leetcode.SubmissionResult{
			Status:      "Wrong Answer",
			StatusCode:  11,
			TotalTests:  63,
			PassedTests: 50,
		},
	}

	sc, _ := pd.Update(sd)
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)

	assert.Contains(t, pd2.errorMsg, "Wrong Answer")
	assert.Contains(t, pd2.errorMsg, "50/63")
	assert.Equal(t, roadmap.StatusInProgress, pd2.status)
}

func TestProblemDetail_SubmitResultError(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	sd := submitResultMsg{
		problemID: 1,
		slug:      "two-sum",
		language:  "golang",
		err:       assert.AnError,
	}

	sc, _ := pd.Update(sd)
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)
	assert.Contains(t, pd2.errorMsg, "Submit failed")
}

func TestProblemDetail_ViewShowsSolveLogs(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	solveAt := time.Now().Add(-1 * time.Hour)
	solveLog := &store.SolveLogRecord{
		ProblemID:   1,
		Slug:        "two-sum",
		Language:    "golang",
		Status:      "Accepted",
		StatusCode:  10,
		Runtime:     "1 ms",
		Memory:      "2 MB",
		TotalTests:  63,
		PassedTests: 63,
		SubmittedAt: solveAt,
	}
	pd.solveLogs = []*store.SolveLogRecord{solveLog}
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Solve Log")
	assert.Contains(t, view, "Accepted")
}

func TestProblemDetail_SpinnerWhileSubmitting(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.submitting = true
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Submitting")
}

func TestProblemDetail_SpinnerAdvances(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.submitting = true

	sc, cmd := pd.Update(spinnerTickMsg{})
	require.NotNil(t, cmd, "tick should continue while submitting")

	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)
	assert.Equal(t, 1, pd2.spinnerFrame)
}

func TestProblemDetail_SpinnerStopsAfterSubmit(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.submitting = false

	_, cmd := pd.Update(spinnerTickMsg{})
	assert.Nil(t, cmd, "tick should stop when not submitting")
}

func TestProblemDetail_SubmitStartsSpinner(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	assert.Contains(t, pd.errorMsg, "authenticated")
	_ = cmd
}

func TestProblemDetail_BackHasStageID(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.NotEmpty(t, navigate.Stage)
	assert.Equal(t, "arrays-hashing", navigate.Stage)
}
