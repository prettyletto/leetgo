package tui

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/generator"
	"github.com/prettyletto/leetgo/internal/leetcode"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/prettyletto/leetgo/internal/workspace"
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

func TestProblemDetail_ViewShowsStatusVerified(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusVerified
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "VERIFIED")
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

func TestProblemDetail_VerifiedPrerequisitesAreNotBlockers(t *testing.T) {
	pd, db := newTestProblemDetail(t, 15)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusVerified))
	pd.refresh()
	pd.width = 120

	missing := pd.missingPrerequisites()
	assert.NotContains(t, missing, "#1 Two Sum")
	assert.Contains(t, missing, "#167 Two Sum II")

	view := pd.View()
	assert.Contains(t, view, "Blocked by")
	assert.Contains(t, view, "#167 Two Sum II")
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

func TestProblemDetail_StartManifestMismatchDoesNotOverwriteFiles(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.cfg.Workspace = t.TempDir()
	pd.workspace = workspace.New(pd.cfg.Workspace, generator.New())

	dir := pd.problemDir()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, workspace.WriteManifest(dir, &workspace.Manifest{
		ProblemID:     49,
		Slug:          "group-anagrams",
		Roadmap:       "from-zero-to-hero",
		Stage:         "arrays-hashing",
		Language:      "go",
		StubPath:      "group_anagrams.go",
		TestsuitePath: "group_anagrams_test.go",
	}))
	stubPath := filepath.Join(dir, "two_sum.go")
	require.NoError(t, os.WriteFile(stubPath, []byte("keep me"), 0o644))

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "different Problem ID")

	stub, err := os.ReadFile(stubPath)
	require.NoError(t, err)
	assert.Equal(t, "keep me", string(stub))
}

func TestProblemDetail_MarkSolved(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	initialXP, err := pd.db.GetStats(context.Background())
	require.NoError(t, err)

	_, cmd := pd.handleMarkSolved()
	assert.Nil(t, cmd)
	assert.True(t, pd.manualSolveMode, "handleMarkSolved should enter confirmation mode")

	for _, r := range "SOLVE" {
		pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	assert.Equal(t, "SOLVE", pd.manualSolveInput)

	pd.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.False(t, pd.manualSolveMode)
	assert.Equal(t, roadmap.StatusSolved, pd.status)
	assert.Contains(t, pd.errorMsg, "Manually marked Solved")
	assert.Contains(t, pd.errorMsg, "No XP awarded")
	stats, err := pd.db.GetStats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, initialXP.TotalXP, stats.TotalXP)
}

func TestProblemDetail_DoubleMarkSolved(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusSolved

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	assert.Nil(t, cmd)
	assert.Equal(t, "Already Solved.", pd.errorMsg)
}

func TestProblemDetail_MarkSolvedLocked(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 49)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("m")})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "locked")
}

func TestProblemDetail_ManualSolveCancel(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	pd.handleMarkSolved()
	assert.True(t, pd.manualSolveMode)

	pd.handleManualSolveKey(tea.KeyMsg{Type: tea.KeyEsc})

	assert.False(t, pd.manualSolveMode)
	assert.Empty(t, pd.manualSolveInput)
	assert.Equal(t, roadmap.StatusAvailable, pd.status)
}

func TestProblemDetail_ManualSolveWrongText(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	pd.handleMarkSolved()
	assert.True(t, pd.manualSolveMode)

	for _, r := range "solve" {
		pd.handleManualSolveKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	assert.Equal(t, "solve", pd.manualSolveInput)

	pd.handleManualSolveKey(tea.KeyMsg{Type: tea.KeyEnter})

	assert.True(t, pd.manualSolveMode, "should still be in confirmation mode after wrong input")
	assert.NotEqual(t, roadmap.StatusSolved, pd.status, "status should not change on wrong input")
}

func TestProblemDetail_ManualSolveViewShowsPrompt(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.width = 80

	pd.handleMarkSolved()
	view := pd.View()

	assert.Contains(t, view, "Manual Solve bypasses verification")
	assert.Contains(t, view, "Type SOLVE to confirm")
	assert.Contains(t, view, "esc")
}

func TestProblemDetail_ManualSolveViewShowsInput(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.width = 80

	pd.handleMarkSolved()
	pd.manualSolveInput = "SOL"
	view := pd.View()

	assert.Contains(t, view, "SOL")
}

func TestProblemDetail_ManualSolveBackspace(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	pd.handleMarkSolved()
	pd.manualSolveInput = "SO"
	pd.handleManualSolveKey(tea.KeyMsg{Type: tea.KeyBackspace})

	assert.Equal(t, "S", pd.manualSolveInput)
	pd.handleManualSolveKey(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "", pd.manualSolveInput)
	pd.handleManualSolveKey(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "", pd.manualSolveInput)
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
	assert.Contains(t, pd.errorMsg, "Session expired")
}

func TestProblemDetail_SubmitVerifiedAllowed(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusVerified
	pd.leetcode = nil

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "Session expired")
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
		assert.Equal(t, initialXP.TotalXP+10, stats.TotalXP)
	}

	hasSubmit, err := pd2.db.HasRewardEvent(ctx, 1, "submit")
	require.NoError(t, err)
	assert.True(t, hasSubmit)

	hasVerify, err := pd2.db.HasRewardEvent(ctx, 1, "verify")
	require.NoError(t, err)
	assert.True(t, hasVerify, "submit from InProgress should also claim verify reward")
}

func TestProblemDetail_SubmitResultAcceptedAlreadyClaimed(t *testing.T) {
	ctx := context.Background()
	pd, db := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusVerified
	require.NoError(t, db.RecordRewardEvent(ctx, &store.RewardEvent{ProblemID: 1, Kind: "submit", XP: 3}))

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

	assert.Contains(t, pd2.errorMsg, "already claimed")
	stats, err := db.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalXP)
}

func TestProblemDetail_SubmitResultAcceptedAlreadyClaimedStillSolves(t *testing.T) {
	ctx := context.Background()
	pd, db := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusVerified
	require.NoError(t, db.RecordRewardEvent(ctx, &store.RewardEvent{ProblemID: 1, Kind: "submit", XP: 3}))

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

	assert.Equal(t, roadmap.StatusSolved, pd2.status)
	progress, err := db.GetProgress(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, progress)
	assert.Equal(t, roadmap.StatusSolved, progress.Status)
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
	assert.Contains(t, pd.errorMsg, "Session expired")
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

func TestProblemDetail_EnterOnVerifiedTriggersSubmit(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusVerified
	pd.leetcode = nil

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Nil(t, cmd)
	assert.Contains(t, pd.errorMsg, "Session expired")
}

func TestProblemDetail_EnterOnSolvedOpensEditor(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusSolved
	pd.cfg.Editor = ""

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
}

func TestProblemDetail_FooterVerifiedShowsSubmitKeytip(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusVerified
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "enter submit")
}

func TestProblemDetail_FooterAvailableShowsStartKeytip(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusAvailable
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "enter start")
}

func TestProblemDetail_FooterInProgressShowsOpenKeytip(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "enter open")
}

func TestProblemDetail_SubmitFromInProgressClaimsBothRewards(t *testing.T) {
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

	assert.Equal(t, roadmap.StatusSolved, pd2.status)

	hasVerify, err := pd2.db.HasRewardEvent(ctx, 1, "verify")
	require.NoError(t, err)
	assert.True(t, hasVerify, "submitting from InProgress should claim verify reward")

	hasSubmit, err := pd2.db.HasRewardEvent(ctx, 1, "submit")
	require.NoError(t, err)
	assert.True(t, hasSubmit, "submitting from InProgress should claim submit reward")

	stats, _ := pd2.db.GetStats(ctx)
	if initialXP != nil {
		expectedXP := initialXP.TotalXP + 7 + 3
		assert.Equal(t, expectedXP, stats.TotalXP)
	}
}

func TestProblemDetail_SubmitFromVerifiedClaimsOnlySubmit(t *testing.T) {
	ctx := context.Background()
	pd, db := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusVerified

	require.NoError(t, db.RecordRewardEvent(ctx, &store.RewardEvent{ProblemID: 1, Kind: "verify", XP: 7}))
	require.NoError(t, db.AddXP(ctx, 7))

	initialXP, _ := db.GetStats(ctx)

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

	assert.Equal(t, roadmap.StatusSolved, pd2.status)

	stats, _ := pd2.db.GetStats(ctx)
	if initialXP != nil {
		assert.Equal(t, initialXP.TotalXP+3, stats.TotalXP)
	}
}

func TestProblemDetail_OpenEditorDetached(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	require.NotNil(t, cmd)
}

func TestProblemDetail_TestRunResultAlreadyVerified(t *testing.T) {
	ctx := context.Background()
	pd, db := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	require.NoError(t, db.RecordRewardEvent(ctx, &store.RewardEvent{ProblemID: 1, Kind: "verify", XP: 7}))
	require.NoError(t, db.AddXP(ctx, 7))

	initialXP, _ := db.GetStats(ctx)

	msg := testRunResultMsg{
		problemID:  1,
		difficulty: roadmap.DifficultyEasy,
		output:     "ok",
		passed:     true,
	}

	sc, cmd := pd.Update(msg)
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)
	require.NotNil(t, cmd)

	notifMsg := cmd()
	notif, ok := notifMsg.(GlobalNotificationMsg)
	require.True(t, ok)
	assert.Contains(t, notif.Message, "already claimed")

	stats, _ := pd2.db.GetStats(ctx)
	if initialXP != nil {
		assert.Equal(t, initialXP.TotalXP, stats.TotalXP)
	}
}

func TestProblemDetail_TestRunResultSolvedWithoutRewardDoesNotDowngrade(t *testing.T) {
	ctx := context.Background()
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusSolved

	initialXP, _ := pd.db.GetStats(ctx)

	msg := testRunResultMsg{
		problemID:  1,
		difficulty: roadmap.DifficultyEasy,
		output:     "ok",
		passed:     true,
	}

	sc, cmd := pd.Update(msg)
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)
	require.NotNil(t, cmd)

	assert.Equal(t, roadmap.StatusSolved, pd2.status)
	notifMsg := cmd()
	notif, ok := notifMsg.(GlobalNotificationMsg)
	require.True(t, ok)
	assert.Contains(t, notif.Message, "already claimed")

	hasVerify, err := pd2.db.HasRewardEvent(ctx, 1, "verify")
	require.NoError(t, err)
	assert.False(t, hasVerify)

	stats, _ := pd2.db.GetStats(ctx)
	if initialXP != nil {
		assert.Equal(t, initialXP.TotalXP, stats.TotalXP)
	}
}

func TestProblemDetail_TestRunResultPassSetsVerified(t *testing.T) {
	ctx := context.Background()
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	initialXP, _ := pd.db.GetStats(ctx)

	msg := testRunResultMsg{
		problemID:  1,
		difficulty: roadmap.DifficultyEasy,
		output:     "ok",
		passed:     true,
	}

	sc, cmd := pd.Update(msg)
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)
	require.NotNil(t, cmd)

	assert.Equal(t, roadmap.StatusVerified, pd2.status)

	hasVerify, err := pd2.db.HasRewardEvent(ctx, 1, "verify")
	require.NoError(t, err)
	assert.True(t, hasVerify)

	stats, _ := pd2.db.GetStats(ctx)
	if initialXP != nil {
		assert.Equal(t, initialXP.TotalXP+7, stats.TotalXP)
	}
}

func TestProblemDetail_TestRunResultFailDoesNotChangeStatus(t *testing.T) {
	ctx := context.Background()
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	initialXP, _ := pd.db.GetStats(ctx)

	msg := testRunResultMsg{
		problemID:  1,
		difficulty: roadmap.DifficultyEasy,
		output:     "FAIL",
		passed:     false,
	}

	sc, cmd := pd.Update(msg)
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)
	require.NotNil(t, cmd)

	assert.Equal(t, roadmap.StatusInProgress, pd2.status)

	hasVerify, err := pd2.db.HasRewardEvent(ctx, 1, "verify")
	require.NoError(t, err)
	assert.False(t, hasVerify)

	stats, _ := pd2.db.GetStats(ctx)
	if initialXP != nil {
		assert.Equal(t, initialXP.TotalXP, stats.TotalXP)
	}
}
