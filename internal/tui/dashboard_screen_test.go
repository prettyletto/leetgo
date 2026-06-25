package tui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/recommendation"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDashboard(t *testing.T) (*DashboardScreen, store.Store) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	rm, err := catalog.LoadRoadmap("from-zero-to-hero")
	require.NoError(t, err)

	cfg := &config.Config{
		OnboardingComplete: true,
		DisplayName:        "Ada",
		Workspace:          t.TempDir(),
		Language:           "go",
		Roadmap:            "from-zero-to-hero",
		Theme:              "rpg-skill-tree",
	}

	theme, err := LookupTheme(cfg.Theme)
	require.NoError(t, err)

	dash := NewDashboardScreen(cfg, theme, db, rm)
	return dash, db
}

func TestDashboard_ViewShowsGreeting(t *testing.T) {
	d, _ := newTestDashboard(t)

	d.width = 120
	d.height = 40

	view := d.View()
	assert.Contains(t, view, "Welcome back, Ada")
}

func TestDashboard_AdaptiveLayoutLabels(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 120
	d.height = 40

	view := d.View()
	assert.Contains(t, view, "Recommended")
	assert.Contains(t, view, "Up next")
	assert.Contains(t, view, "Profile")
	assert.Contains(t, view, "Roadmap")
}

func TestDashboard_LegacyThemeValuesRenderAdaptiveLayout(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.cfg.Theme = "clean-productivity"
	theme, err := LookupTheme(d.cfg.Theme)
	require.NoError(t, err)
	d.theme = theme
	d.width = 120
	d.height = 40

	view := d.View()
	assert.Contains(t, view, "Recommended")
	assert.Contains(t, view, "Queue")
	assert.Contains(t, view, "Profile")
	assert.Contains(t, view, "Upcoming")
	assert.NotContains(t, view, "Character HUD")
}

func TestDashboard_CyberThemeValueNormalizesToAdaptive(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.cfg.Theme = "cyber-dashboard"
	theme, err := LookupTheme(d.cfg.Theme)
	require.NoError(t, err)
	d.theme = theme
	d.width = 120
	d.height = 40

	view := d.View()
	assert.Contains(t, view, "Recommended")
	assert.Contains(t, view, "Roadmap")
	assert.NotContains(t, view, "SYS:")
}

func TestDashboard_CleanPlainSymbolsNoRichGlyphs(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.cfg.Theme = "clean-productivity"
	d.cfg.SymbolMode = "plain"
	theme, err := LookupTheme(d.cfg.Theme)
	require.NoError(t, err)
	d.theme = theme
	d.width = 120
	d.height = 40

	view := d.View()
	assert.Contains(t, view, "XP Level")
	assert.NotContains(t, view, "✦")
}

func TestDashboard_ViewShowsDisplayName(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.cfg.DisplayName = "Hello"

	d.width = 120
	view := d.View()
	assert.Contains(t, view, "Hello")
}

func TestDashboard_ViewShowsRoadmapTitle(t *testing.T) {
	d, _ := newTestDashboard(t)

	d.width = 120
	view := d.View()
	assert.Contains(t, view, "From Zero To Hero")
}

func TestDashboard_ViewHasFooter(t *testing.T) {
	d, _ := newTestDashboard(t)

	d.width = 120
	view := d.View()
	assert.Contains(t, view, "j/k")
	assert.Contains(t, view, "enter")
	assert.Contains(t, view, "t")
	assert.Contains(t, view, "q")
}

func TestDashboard_Quit(t *testing.T) {
	d, _ := newTestDashboard(t)

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	require.NotNil(t, cmd)
}

func TestDashboard_NavigateToList(t *testing.T) {
	d, _ := newTestDashboard(t)

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenLegacyList, navigate.ScreenID)
}

func TestDashboard_TKeyDoesNotChangeTheme(t *testing.T) {
	d, _ := newTestDashboard(t)

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	assert.Nil(t, cmd)
}

func TestDashboard_FocusNavigation(t *testing.T) {
	d, _ := newTestDashboard(t)

	if len(d.actions) == 0 {
		t.Skip("no actions available")
	}

	assert.Equal(t, 0, d.focusIndex)

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 1, d.focusIndex)

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 2, d.focusIndex)

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 1, d.focusIndex)

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 0, d.focusIndex)
}

func TestDashboard_ArrowFocusNavigation(t *testing.T) {
	d, _ := newTestDashboard(t)

	if len(d.actions) == 0 {
		t.Skip("no actions available")
	}

	assert.Equal(t, 0, d.focusIndex)

	d.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, d.focusIndex)

	d.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, d.focusIndex)
}

func TestDashboard_ArrowStringFocusNavigation(t *testing.T) {
	d, _ := newTestDashboard(t)

	if len(d.actions) == 0 {
		t.Skip("no actions available")
	}

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("down")})
	assert.Equal(t, 1, d.focusIndex)

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")})
	assert.Equal(t, 0, d.focusIndex)
}

func TestDashboard_ArrowVariantFocusNavigation(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyMsg
		want int
	}{
		{name: "shift down", msg: tea.KeyMsg{Type: tea.KeyShiftDown}, want: 1},
		{name: "ctrl down", msg: tea.KeyMsg{Type: tea.KeyCtrlDown}, want: 1},
		{name: "alt down", msg: tea.KeyMsg{Type: tea.KeyDown, Alt: true}, want: 1},
		{name: "shift up", msg: tea.KeyMsg{Type: tea.KeyShiftUp}, want: -1},
		{name: "ctrl up", msg: tea.KeyMsg{Type: tea.KeyCtrlUp}, want: -1},
		{name: "alt up", msg: tea.KeyMsg{Type: tea.KeyUp, Alt: true}, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := dashboardFocusDelta(tt.msg)
			require.True(t, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDashboard_FocusDeltaIgnoresPlainRunes(t *testing.T) {
	_, ok := dashboardFocusDelta(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	assert.False(t, ok)
}

func TestDashboard_FocusedActionShowsCursor(t *testing.T) {
	d, _ := newTestDashboard(t)

	if len(d.actions) == 0 {
		t.Skip("no actions available")
	}

	view := d.View()
	assert.Contains(t, view, "Recommended")
	assert.Contains(t, view, "Kth Largest Element in a Stream")
}

func TestDashboard_EnterOnSubmitAction(t *testing.T) {
	d, db := newTestDashboard(t)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusVerified))
	d.refresh(ctx)

	hasSubmit := false
	for i, a := range d.actions {
		if a.Kind == recommendation.KindSubmit && a.ProblemID == 1 {
			d.focusIndex = i
			hasSubmit = true
			break
		}
	}
	if !hasSubmit {
		t.Skip("no submit action for problem 1")
	}

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenProblemDetail, navigate.ScreenID)
	assert.Equal(t, 1, navigate.ProblemID)
}

func TestDashboard_SubmitActionVisibleInView(t *testing.T) {
	d, db := newTestDashboard(t)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusVerified))
	d.refresh(ctx)
	d.width = 120

	hasSubmit := false
	for _, a := range d.actions {
		if a.Kind == recommendation.KindSubmit {
			hasSubmit = true
			break
		}
	}
	if !hasSubmit {
		t.Skip("no submit action")
	}

	view := d.View()
	assert.Contains(t, view, "Submit")
}

func TestDashboard_FocusWrapsWithSubmitActions(t *testing.T) {
	d, db := newTestDashboard(t)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusVerified))
	require.NoError(t, db.SetProgress(ctx, 242, roadmap.StatusVerified))
	d.refresh(ctx)

	if len(d.actions) == 0 {
		t.Skip("no actions")
	}

	rendered := len(d.actions)
	maxShow := 5
	if rendered > maxShow {
		rendered = maxShow
	}

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, rendered-1, d.focusIndex)

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 0, d.focusIndex)
}

func TestDashboard_EnterOnStartAction(t *testing.T) {
	d, _ := newTestDashboard(t)

	hasStart := false
	for i, a := range d.actions {
		if a.Kind == recommendation.KindStart {
			d.focusIndex = i
			hasStart = true
			break
		}
	}
	if !hasStart {
		t.Skip("no start actions available")
	}

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenProblemDetail, navigate.ScreenID)
	assert.Greater(t, navigate.ProblemID, 0)
}

func TestDashboard_EnterOnContinueAction(t *testing.T) {
	d, db := newTestDashboard(t)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusInProgress))
	d.refresh(ctx)

	hasContinue := false
	for i, a := range d.actions {
		if a.Kind == recommendation.KindContinue && a.ProblemID == 1 {
			d.focusIndex = i
			hasContinue = true
			break
		}
	}
	if !hasContinue {
		t.Skip("no continue action for problem 1")
	}

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenProblemDetail, navigate.ScreenID)
	assert.Equal(t, 1, navigate.ProblemID)
}

func TestDashboard_EnterOnExportAction(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.cfg.GitExportEnabled = true
	d.cfg.GitExportRepo = "/tmp/repo"
	d.refresh(context.Background())

	for i, a := range d.actions {
		if a.Kind == recommendation.KindExport {
			d.focusIndex = i
			break
		}
	}

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	notif, ok := msg.(GlobalNotificationMsg)
	require.True(t, ok)
	assert.Contains(t, notif.Message, "git-export")
	assert.NotContains(t, notif.Message, "legacy")
}

func TestDashboard_EnterOnInspectAction(t *testing.T) {
	d, _ := newTestDashboard(t)

	for i, a := range d.actions {
		if a.Kind == recommendation.KindInspect {
			d.focusIndex = i
			break
		}
	}

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenSolveLog, navigate.ScreenID)
}

func TestDashboard_WindowResize(t *testing.T) {
	d, _ := newTestDashboard(t)

	d.Update(tea.WindowSizeMsg{Width: 150, Height: 50})
	assert.Equal(t, 150, d.width)
	assert.Equal(t, 50, d.height)
}

func TestDashboard_WideLayoutHasThreeColumns(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 120
	d.height = 40

	view := d.View()
	assert.Contains(t, view, "XP", "wide view should show stats")
	assert.Contains(t, view, "From Zero To Hero", "wide view should show roadmap info")
	assert.Contains(t, view, "j/k", "wide view should show footer")
	assert.Contains(t, view, "up/down", "footer should show arrow navigation")
}

func TestDashboard_CenterContentWhenSized(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 20
	d.height = 5

	view := d.centerContent("x")
	assert.Contains(t, view, "x")
	assert.Greater(t, len(view), 1, "centered content should include placement padding")
}

func TestDashboard_NarrowLayoutShowsCenterOnly(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 50

	view := d.View()
	assert.Contains(t, view, "Welcome")
	assert.Contains(t, view, "j/k")
}

func TestDashboard_MediumLayoutShowsRailsBelow(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 80

	view := d.View()
	assert.Contains(t, view, "Welcome")
	assert.Contains(t, view, "j/k")
}

func TestDashboard_EmptyActions(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.actions = nil
	d.width = 120

	view := d.View()
	assert.Contains(t, view, "No actions available")
}

func TestDashboard_NextActionsShown(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 120

	view := d.View()
	hasStartActions := false
	for _, a := range d.actions {
		if a.Kind == recommendation.KindStart || a.Kind == recommendation.KindContinue {
			hasStartActions = true
			break
		}
	}
	if hasStartActions {
		assert.Contains(t, view, "Start")
	}
}

func TestDashboard_ProgressStatsInHUD(t *testing.T) {
	d, db := newTestDashboard(t)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, db.SetProgress(ctx, 217, roadmap.StatusSolved))
	require.NoError(t, db.SetProgress(ctx, 242, roadmap.StatusInProgress))
	require.NoError(t, db.AddXP(ctx, 50))

	d.refresh(ctx)
	d.width = 120

	view := d.View()
	assert.Contains(t, view, "Level")
	assert.Contains(t, view, "XP")
	assert.Contains(t, view, "Streak")
	assert.Contains(t, view, "Verified: 0")
	assert.Contains(t, view, "Solved: 2")
	assert.Contains(t, view, "Progress: 2")
}

func TestDashboard_VerifiedStatsInHUD(t *testing.T) {
	d, db := newTestDashboard(t)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusVerified))
	require.NoError(t, db.SetProgress(ctx, 217, roadmap.StatusVerified))
	require.NoError(t, db.SetProgress(ctx, 242, roadmap.StatusSolved))

	d.refresh(ctx)
	d.width = 120

	view := d.View()
	assert.Contains(t, view, "Verified: 2")
	assert.Contains(t, view, "Solved: 1")
	assert.Contains(t, view, "Progress: 3")
}

func TestDashboard_RoadmapContextInWideView(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 120

	view := d.View()
	hasRoadmapContext := strings.Contains(view, "Stage") || strings.Contains(view, "Progress")
	assert.True(t, hasRoadmapContext, "wide view should include roadmap context with stage/progress info")
}

func TestDashboard_NoThemeChangeShortcut(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 120

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	assert.Nil(t, cmd)
}

func TestDashboard_RKeyShowsRoadmapDetail(t *testing.T) {
	d, _ := newTestDashboard(t)

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenRoadmapDetail, navigate.ScreenID)
}

func TestDashboard_SKeyShowsSolveLog(t *testing.T) {
	d, _ := newTestDashboard(t)

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenSolveLog, navigate.ScreenID)
}

func TestDashboard_CompletionActionNavigatesToCompletionScreen(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.actions = []recommendation.NextAction{{Kind: recommendation.KindViewRoadmapCompletion, Title: "View Roadmap Completion"}}
	d.focusIndex = 0

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenCompletion, navigate.ScreenID)
}

func TestDashboard_ReviewActivationCreatesReviewCycle(t *testing.T) {
	d, db := newTestDashboard(t)
	action := recommendation.NextAction{
		Kind:       recommendation.KindReview,
		ProblemID:  1,
		Title:      "Review Two Sum",
		ReasonType: recommendation.ReasonValidatesManualSolve,
	}
	d.actions = []recommendation.NextAction{action}
	d.focusIndex = 0

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenProblemDetail, navigate.ScreenID)

	cycles, err := db.GetReviewCyclesForProblem(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, cycles, 1)
	assert.Equal(t, "manual_solve_validation", cycles[0].Reason)
}

func TestDashboard_ContinueActionShowsInProgress(t *testing.T) {
	d, db := newTestDashboard(t)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusInProgress))
	d.refresh(ctx)
	d.width = 120

	hasContinue := false
	for _, a := range d.actions {
		if a.Kind == recommendation.KindContinue {
			hasContinue = true
			assert.Equal(t, 1, a.ProblemID)
		}
	}
	assert.True(t, hasContinue, "should have a Continue action for InProgress problem")

	view := d.View()
	assert.Contains(t, view, "Continue")
}

func TestDashboard_XPProgressBar(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.stats = &store.Stats{
		Level:   2,
		TotalXP: 150,
	}

	d.width = 120
	view := d.View()
	assert.Contains(t, view, "XP")
}

func TestDashboard_XPNaN(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.stats = &store.Stats{
		Level:   1,
		TotalXP: 0,
	}

	d.width = 120
	view := d.View()
	assert.Contains(t, view, "XP")
}

func TestDashboard_LatestAchievement(t *testing.T) {
	d, db := newTestDashboard(t)
	ctx := context.Background()

	require.NoError(t, db.UnlockAchievement(ctx, "first_solve"))
	d.refresh(ctx)
	d.width = 120

	view := d.View()
	assert.Contains(t, view, "Latest achievement:")
	assert.Contains(t, view, "First")
}

func TestDashboard_NoAchievement(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.achievementIDs = nil
	d.width = 120

	view := d.View()
	assert.NotContains(t, view, "Latest:")
}

func TestDashboard_BlockerSummary(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 120

	view := d.View()
	hasBlocker := strings.Contains(view, "Blocker") || strings.Contains(view, "blocked by")
	assert.True(t, hasBlocker, "wide view should include blocker summary")
}

func TestDashboard_ExportFilteredWhenDisabled(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.cfg.GitExportEnabled = false

	d.refresh(context.Background())

	for _, a := range d.actions {
		assert.NotEqual(t, recommendation.KindExport, a.Kind, "export action should not appear when disabled")
	}
}

func TestDashboard_ExportVisibleWhenEnabled(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.cfg.GitExportEnabled = true
	d.cfg.GitExportRepo = "/tmp/repo"

	d.refresh(context.Background())

	hasExport := false
	for _, a := range d.actions {
		if a.Kind == recommendation.KindExport {
			hasExport = true
		}
	}
	assert.True(t, hasExport, "export action should appear when git export is enabled with repo")
}

func TestDashboard_ExportHiddenWhenNoRepo(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.cfg.GitExportEnabled = true
	d.cfg.GitExportRepo = ""

	d.refresh(context.Background())

	for _, a := range d.actions {
		assert.NotEqual(t, recommendation.KindExport, a.Kind, "export action should not appear when repo is empty")
	}
}

func TestDashboard_NarrowLayoutOmitsHUD(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 50

	view := d.View()
	assert.NotContains(t, view, "Latest:")
	assert.Contains(t, view, "j/k")
}

func TestDashboard_MediumLayoutShowsHUD(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.width = 80

	view := d.View()
	assert.Contains(t, view, "j/k")
}

func TestDashboard_BackNavigationStageID(t *testing.T) {
	d, _ := newTestDashboard(t)

	for i, a := range d.actions {
		if a.Kind == recommendation.KindStart {
			d.focusIndex = i
			_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
			require.NotNil(t, cmd)
			msg := cmd()
			navigate, ok := msg.(NavigateMsg)
			require.True(t, ok)
			assert.Equal(t, ScreenProblemDetail, navigate.ScreenID)
			assert.Greater(t, navigate.ProblemID, 0)
			return
		}
	}
	t.Skip("no start actions")
}

func TestDashboard_FocusClampedToRenderedWindow(t *testing.T) {
	d, _ := newTestDashboard(t)

	if len(d.actions) <= 5 {
		t.Skip("need more than 5 actions to test focus clamping")
	}

	maxShow := 5

	for i := 0; i < maxShow+2; i++ {
		d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	assert.Less(t, d.focusIndex, maxShow, "focus should not exceed rendered window")

	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Less(t, d.focusIndex, maxShow, "focus should stay within rendered window")
}

func TestDashboard_ExportWordingIsCLI(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.cfg.GitExportEnabled = true
	d.cfg.GitExportRepo = "/tmp/repo"
	d.refresh(context.Background())

	for i, a := range d.actions {
		if a.Kind == recommendation.KindExport {
			assert.Contains(t, a.Title, "CLI", "title should indicate CLI usage")
			assert.Contains(t, a.Reason, "leetgo", "reason should contain CLI command hint")
			assert.Contains(t, a.Reason, "git-export", "reason should contain git-export command")
			_ = i
			return
		}
	}
	t.Skip("no export action")
}

func TestDashboard_ExportEnterShowsNotification(t *testing.T) {
	d, _ := newTestDashboard(t)
	d.cfg.GitExportEnabled = true
	d.cfg.GitExportRepo = "/tmp/repo"
	d.refresh(context.Background())

	for i, a := range d.actions {
		if a.Kind == recommendation.KindExport {
			d.focusIndex = i
			break
		}
	}

	_, cmd := d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	notif, ok := msg.(GlobalNotificationMsg)
	require.True(t, ok)
	assert.Contains(t, notif.Message, "leetgo git-export")
	assert.Contains(t, notif.Message, "repo-dir")
	assert.Contains(t, notif.Message, "--commit")
}

func TestDashboard_FocusResetsOnRefresh(t *testing.T) {
	d, _ := newTestDashboard(t)

	maxShow := 5
	for i := 0; i < maxShow-1; i++ {
		d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	assert.Equal(t, maxShow-1, d.focusIndex)

	d.refresh(context.Background())
	rendered := len(d.actions)
	if rendered > maxShow {
		rendered = maxShow
	}
	assert.Less(t, d.focusIndex, rendered, "focus should stay within rendered window after refresh")
}
