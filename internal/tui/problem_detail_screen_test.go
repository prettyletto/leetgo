package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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

func TestProblemDetail_UsesAdaptiveLabels(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Problem Detail")
	assert.Contains(t, view, "Status")
	assert.Contains(t, view, "Problem Brief")
	assert.Contains(t, view, "Workspace Files")
}

func TestProblemDetail_VerifiedMakesSubmitPrimary(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusVerified
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Primary: Submit")
	assert.Contains(t, view, "Manual Solve")
	assert.Contains(t, view, "enter submit (primary)")
}

func TestProblemDetail_MotionOffUsesStaticSubmittingText(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.cfg.MotionPreference = "off"
	pd.submitting = true
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Submitting to LeetCode")
	assert.NotContains(t, view, "⣾")
	assert.False(t, pd.allowsSpinnerMotion())
}

func TestProblemDetail_CleanThemeLabels(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.cfg.Theme = "clean-productivity"
	theme, err := LookupTheme(pd.cfg.Theme)
	require.NoError(t, err)
	pd.theme = theme
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Problem Detail")
	assert.Contains(t, view, "Problem Brief")
	assert.Contains(t, view, "Workspace Files")
	assert.NotContains(t, view, "Skill Tile")
}

func TestProblemDetail_PlainSymbolsPreserveStatusMeaning(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.cfg.SymbolMode = "plain"
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "[READY]")
	assert.NotContains(t, view, "◆")
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

func TestProblemDetail_DifficultyTagIsStyled(t *testing.T) {
	theme, err := LookupTheme("rpg-skill-tree")
	require.NoError(t, err)

	tag := renderProblemDetailDifficultyTag(theme, roadmap.DifficultyEasy)

	assert.Equal(t, " easy ", tag)
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
	assert.Contains(t, view, "Requires: none")
}

func TestProblemDetail_ViewShowsPrerequisitesWithContent(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 49)
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Requires:")
	assert.Contains(t, view, "Valid Anagram")
}

func TestProblemDetail_ViewShowsBlockers(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 49)
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Blocked by")
	assert.Contains(t, view, "LOCKED")
}

func TestProblemDetail_VerifiedPrerequisitesAreBlockers(t *testing.T) {
	pd, db := newTestProblemDetail(t, 15)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusVerified))
	pd.refresh()
	pd.width = 120

	missing := pd.missingPrerequisites()
	assert.Contains(t, missing, "#1 Two Sum")
	assert.Contains(t, missing, "#167 Two Sum II")

	view := pd.View()
	assert.Contains(t, view, "Blocked by")
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
	assert.Equal(t, "note", pd.manualSolvePhase)

	pd.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.False(t, pd.manualSolveMode)
	assert.Equal(t, roadmap.StatusSolved, pd.status)
	assert.Contains(t, pd.errorMsg, "Manually marked Solved")
	assert.Contains(t, pd.errorMsg, "Reward")
	assert.Contains(t, pd.errorMsg, "No XP awarded")
	stats, err := pd.db.GetStats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, initialXP.TotalXP, stats.TotalXP)

	sp, err := pd.db.GetSolveProvenance(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, sp)
	assert.Equal(t, "manual", sp.Kind)
}

func TestProblemDetail_AcceptedSolveRewardMoment(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusVerified

	pd.handleSubmitResult(submitResultMsg{
		problemID: 1,
		slug:      "two-sum",
		language:  "golang",
		result: &leetcode.SubmissionResult{
			Status:      "Accepted",
			StatusCode:  10,
			Runtime:     "1 ms",
			Memory:      "2 MB",
			PassedTests: 63,
			TotalTests:  63,
		},
		duration: 2 * time.Second,
	})

	assert.Equal(t, roadmap.StatusSolved, pd.status)
	assert.Contains(t, pd.errorMsg, "Problem Solved")
	assert.Contains(t, pd.errorMsg, "Accepted by LeetCode")
	assert.Contains(t, pd.errorMsg, "Duration: 2s")
	assert.Contains(t, pd.errorMsg, "Actions")
}

func TestProblemDetail_ReviewCompletionRewardMoment(t *testing.T) {
	pd, db := newTestProblemDetail(t, 1)
	ctx := context.Background()
	pd.status = roadmap.StatusSolved
	require.NoError(t, db.CreateReviewCycle(ctx, &store.ReviewCycle{ProblemID: 1, Reason: "weakness", RoadmapID: "from-zero-to-hero"}))

	sc, cmd := pd.handleTestRunResult(testRunResultMsg{problemID: 1, difficulty: roadmap.DifficultyEasy, output: "ok", passed: true, duration: time.Second})
	pd = sc.(*ProblemDetailScreen)
	assert.Nil(t, cmd)
	assert.Contains(t, pd.testResultMsg, "Review Complete")
	assert.Contains(t, pd.testResultMsg, "Tests passed")
	assert.Contains(t, pd.testResultMsg, "+5 XP")
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

	assert.Contains(t, view, "will not earn XP unless LeetCode accepts")
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

func TestProblemDetail_NarrowLayoutCyclesContextAndProgression(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.Update(tea.WindowSizeMsg{Width: 96, Height: 32})

	view := pd.View()
	assert.Contains(t, view, "Context")
	assert.Contains(t, view, "Problem Brief")
	assert.NotContains(t, view, "#167 Two Sum II (medium)")
	assert.Contains(t, view, "left/right")

	pd.Update(tea.KeyMsg{Type: tea.KeyRight})
	view = pd.View()
	assert.Contains(t, view, "Progression")
	assert.Contains(t, view, "#167 Two Sum II (medium)")
	assert.NotContains(t, view, "Requires: none")

	pd.Update(tea.KeyMsg{Type: tea.KeyLeft})
	view = pd.View()
	assert.Contains(t, view, "Context")
	assert.Contains(t, view, "Problem Brief")
}

func TestProblemDetail_NarrowContextScrollsLongBrief(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.problem.Summary = strings.Join([]string{
		"brief-line-00",
		"brief-line-01",
		"brief-line-02",
		"brief-line-03",
		"brief-line-04",
		"brief-line-05",
		"brief-line-06",
		"brief-line-07",
		"brief-line-08",
		"brief-line-09",
	}, "\n")
	pd.problem.PracticeFocus = ""
	pd.problem.WhyNow = ""
	pd.problem.UnlockImpact = ""
	pd.Update(tea.WindowSizeMsg{Width: 96, Height: 28})

	view := pd.View()
	assert.Contains(t, view, "brief-line-00")
	assert.NotContains(t, view, "brief-line-09")
	assert.Contains(t, view, "scroll")

	pd.Update(tea.KeyMsg{Type: tea.KeyDown})
	view = pd.View()
	assert.Contains(t, view, "brief-line-01")
	assert.Contains(t, view, "scroll 2/")
}

func TestProblemDetail_NoThemeCycleShortcut(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	_, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	assert.Nil(t, cmd)
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

func TestProblemDetail_ViewShowsPracticeLog(t *testing.T) {
	pd, db := newTestProblemDetail(t, 1)
	ctx := context.Background()
	solveAt := time.Now().Add(-1 * time.Hour)
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
		SubmittedAt: solveAt,
	}))
	pd.width = 120

	view := pd.View()
	assert.Contains(t, view, "Practice Log")
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
	pd.leetcode = nil

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
	assert.Nil(t, cmd)
	assert.Contains(t, pd2.testResultMsg, "already claimed")

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
	assert.Nil(t, cmd)

	assert.Equal(t, roadmap.StatusSolved, pd2.status)
	assert.Contains(t, pd2.testResultMsg, "Tests passed")

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
	assert.Nil(t, cmd)

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
	assert.Nil(t, cmd)

	assert.Equal(t, roadmap.StatusInProgress, pd2.status)

	hasVerify, err := pd2.db.HasRewardEvent(ctx, 1, "verify")
	require.NoError(t, err)
	assert.False(t, hasVerify)

	stats, _ := pd2.db.GetStats(ctx)
	if initialXP != nil {
		assert.Equal(t, initialXP.TotalXP, stats.TotalXP)
	}
}

func TestProblemDetail_SubmitLocalTestFailureStopsBeforeSubmission(t *testing.T) {
	ctx := context.Background()
	pd, db := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress

	msg := submitResultMsg{
		problemID:       1,
		slug:            "two-sum",
		language:        "golang",
		err:             assert.AnError,
		localTestRan:    true,
		localTestPassed: false,
		localTestOutput: "FAIL",
	}

	sc, cmd := pd.Update(msg)
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)
	require.NotNil(t, cmd)
	assert.Equal(t, roadmap.StatusInProgress, pd2.status)
	assert.Contains(t, pd2.errorMsg, "Local TestSuite failed")
	assert.True(t, pd2.submitAnywayMode)

	logs, err := db.GetSolveLogsForProblem(ctx, 1)
	require.NoError(t, err)
	assert.Empty(t, logs, "local test failure should not create a Submission Practice Log entry")

	attempts, err := db.GetAttempts(ctx, 1)
	require.NoError(t, err)
	require.Len(t, attempts, 1)
	assert.False(t, attempts[0].Passed)
}

func TestProblemDetail_LocalTestFailureRendersFocusedResult(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress
	pd.width = 120
	pd.height = 40

	output := "--- FAIL: TestTwoSum (0.00s)\n    two_sum_test.go:24: got [], want [0 1]\nFAIL"
	sc, cmd := pd.handleTestRunResult(testRunResultMsg{problemID: 1, difficulty: roadmap.DifficultyEasy, output: output, passed: false, duration: time.Second})
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)
	assert.Nil(t, cmd)

	view := pd2.View()
	assert.Contains(t, view, "TestSuite Failed")
	assert.Contains(t, view, "Local TestSuite failed")
	assert.Contains(t, view, "    two_sum_test.go:24")
	assert.LessOrEqual(t, maxRenderedLineWidth(view), 120)
	assert.NotContains(t, view, "Context")
	assert.NotContains(t, view, "Workspace Files")
	assert.NotContains(t, view, "Practice Log")

	pd2.Update(tea.KeyMsg{Type: tea.KeyEsc})
	view = pd2.View()
	assert.Contains(t, view, "Context")
}

func TestProblemDetail_SubmitAnywayConfirmationRenders(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.submitAnywayMode = true
	pd.submitAnywayInput = "SUB"

	view := pd.View()
	assert.Contains(t, view, "Submit Anyway")
	assert.Contains(t, view, "Local TestSuite failed")
	assert.Contains(t, view, "SUB")
}

func TestProblemDetail_SubmitAnywayEscCancels(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.submitAnywayMode = true
	pd.submitAnywayInput = "SUBMIT"

	sc, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEsc})
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)
	assert.Nil(t, cmd)
	assert.False(t, pd2.submitAnywayMode)
	assert.Empty(t, pd2.submitAnywayInput)
	assert.Contains(t, pd2.errorMsg, "cancelled")
}

func TestProblemDetail_SubmitAnywayRequiresTypedConfirmation(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusInProgress
	pd.submitAnywayMode = true
	pd.submitAnywayInput = "NOPE"

	sc, cmd := pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pd2, ok := sc.(*ProblemDetailScreen)
	require.True(t, ok)
	assert.Nil(t, cmd)
	assert.True(t, pd2.submitAnywayMode)
}

func TestProblemDetail_ManualSolveWithNote(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	pd.handleMarkSolved()
	assert.True(t, pd.manualSolveMode)
	assert.Equal(t, "", pd.manualSolvePhase)

	for _, r := range "SOLVE" {
		pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "note", pd.manualSolvePhase)

	noteText := "Solved in browser"
	for _, r := range noteText {
		pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	assert.Equal(t, noteText, pd.manualSolveNote)

	pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, pd.manualSolveMode)
	assert.Equal(t, roadmap.StatusSolved, pd.status)

	sp, err := pd.db.GetSolveProvenance(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, sp)
	assert.Equal(t, "manual", sp.Kind)
	assert.Equal(t, noteText, sp.Note)
}

func TestProblemDetail_ManualSolveNotePhaseView(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)
	pd.width = 80

	pd.handleMarkSolved()
	for _, r := range "SOLVE" {
		pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "note", pd.manualSolvePhase)

	view := pd.View()
	assert.Contains(t, view, "Optional note")
	assert.Contains(t, view, "enter")
}

func TestProblemDetail_ManualSolveNoteEscCancels(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	pd.handleMarkSolved()
	for _, r := range "SOLVE" {
		pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, "note", pd.manualSolvePhase)

	pd.handleManualSolveKey(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, pd.manualSolveMode)
	assert.Equal(t, "", pd.manualSolvePhase)
}

func TestProblemDetail_ManualSolveNoteBackspace(t *testing.T) {
	pd, _ := newTestProblemDetail(t, 1)

	pd.handleMarkSolved()
	for _, r := range "SOLVE" {
		pd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	pd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	pd.manualSolveNote = "abc"
	pd.handleManualSolveKey(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "ab", pd.manualSolveNote)
}

func TestProblemDetail_SubmitAcceptedRecordsProvenance(t *testing.T) {
	ctx := context.Background()
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusVerified

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
	pd2, _ := sc.(*ProblemDetailScreen)

	assert.Equal(t, roadmap.StatusSolved, pd2.status)

	sp, err := pd2.db.GetSolveProvenance(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, sp)
	assert.Equal(t, "accepted", sp.Kind)
	assert.NotNil(t, sp.SolveLogID)
}

func TestProblemDetail_AcceptedUpgradesManualSolve(t *testing.T) {
	ctx := context.Background()
	pd, _ := newTestProblemDetail(t, 1)
	pd.status = roadmap.StatusSolved

	require.NoError(t, pd.db.RecordSolveProvenance(ctx, &store.SolveProvenance{
		ProblemID: 1,
		Kind:      "manual",
		Note:      "temp solve",
	}))

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
	pd2, _ := sc.(*ProblemDetailScreen)

	assert.Equal(t, roadmap.StatusSolved, pd2.status)

	sp, err := pd2.db.GetSolveProvenance(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, sp)
	assert.Equal(t, "accepted", sp.Kind)
	assert.Equal(t, "temp solve", sp.Note)
}

func TestProblemDetail_BuildPracticeLog(t *testing.T) {
	pd, db := newTestProblemDetail(t, 1)
	ctx := context.Background()

	require.NoError(t, db.RecordAttempt(ctx, &store.AttemptRecord{
		ProblemID: 1,
		Timestamp: time.Now().Add(-2 * time.Hour),
		Duration:  5 * time.Minute,
		Passed:    false,
	}))
	require.NoError(t, db.RecordAttempt(ctx, &store.AttemptRecord{
		ProblemID: 1,
		Timestamp: time.Now().Add(-1 * time.Hour),
		Duration:  3 * time.Minute,
		Passed:    true,
	}))
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
	require.NoError(t, db.RecordSolveProvenance(ctx, &store.SolveProvenance{
		ProblemID: 1,
		Kind:      "accepted",
	}))

	entries := BuildPracticeLog(db, 1)
	require.NotEmpty(t, entries)
	assert.LessOrEqual(t, len(entries), 3)

	hasAcceptedSolve := false
	hasLocalAttempt := false
	for _, e := range entries {
		if e.Kind == "Accepted Solve" {
			hasAcceptedSolve = true
		}
		if e.Kind == "Local Attempt passed" || e.Kind == "Local Attempt failed" {
			hasLocalAttempt = true
		}
	}
	assert.True(t, hasAcceptedSolve, "practice log should include accepted solve entry")
	assert.True(t, hasLocalAttempt, "practice log should include local attempt entries")

	pd.width = 120
	view := pd.View()
	assert.Contains(t, view, "Practice Log")
}

func TestProblemDetail_BuildPracticeLogKeepsThreeNewest(t *testing.T) {
	_, db := newTestProblemDetail(t, 1)
	ctx := context.Background()
	base := time.Now()

	for i := 0; i < 5; i++ {
		require.NoError(t, db.RecordAttempt(ctx, &store.AttemptRecord{
			ProblemID: 1,
			Timestamp: base.Add(time.Duration(i) * time.Minute),
			Duration:  time.Minute,
			Passed:    i%2 == 0,
		}))
	}

	entries := BuildPracticeLog(db, 1)
	require.Len(t, entries, 3)
	assert.True(t, entries[0].Timestamp.After(entries[1].Timestamp))
	assert.True(t, entries[1].Timestamp.After(entries[2].Timestamp))
	assert.Equal(t, base.Add(4*time.Minute).Unix(), entries[0].Timestamp.Unix())
}
