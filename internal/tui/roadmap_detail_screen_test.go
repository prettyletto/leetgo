package tui

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/leetgo/internal/catalog"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoadmapDetail(t *testing.T) (*RoadmapDetailScreen, store.Store) {
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

	rd := NewRoadmapDetailScreen(cfg, theme, db, rm)
	return rd, db
}

func TestRoadmapDetail_ViewShowsTitle(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "From Zero To Hero")
}

func TestRoadmapDetail_UsesAdaptiveLabels(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "Roadmap")
	assert.Contains(t, view, "Stage:")
	assert.Contains(t, view, "Upcoming")
}

func TestRoadmapDetail_LegacyThemeValuesRenderAdaptiveLabels(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.cfg.Theme = "clean-productivity"
	theme, err := LookupTheme(rd.cfg.Theme)
	require.NoError(t, err)
	rd.theme = theme
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "Roadmap: From Zero To Hero")
	assert.Contains(t, view, "Stage:")
	assert.Contains(t, view, "Upcoming")
	assert.NotContains(t, view, "World Map")
}

func TestRoadmapDetail_CyberThemeValueNormalizesToAdaptive(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.cfg.Theme = "cyber-dashboard"
	theme, err := LookupTheme(rd.cfg.Theme)
	require.NoError(t, err)
	rd.theme = theme
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "Roadmap")
	assert.Contains(t, view, "Stage:")
	assert.Contains(t, view, "Upcoming")
}

func TestRoadmapDetail_PlainSymbolsPreserveStatusMeaning(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.cfg.SymbolMode = "plain"
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "[READY]")
	assert.Contains(t, view, "[LOCKED]")
	assert.NotContains(t, view, "🔒")
}

func TestRoadmapDetail_ViewShowsTagline(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, rd.roadmap.Tagline)
}

func TestRoadmapDetail_ViewShowsStages(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120
	rd.height = 40

	view := rd.View()
	assert.Contains(t, view, "Arrays & Hashing")
	assert.Contains(t, view, "Two Pointers")
	assert.Contains(t, view, "Sliding Window")
	assert.Contains(t, view, "more Stage below")
}

func TestRoadmapDetail_ViewShowsProblems(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "Two Sum")
	assert.Contains(t, view, "Contains Duplicate")
	assert.Contains(t, view, "Group Anagrams")
}

func TestRoadmapDetail_ViewShowsSolvedCount(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "solved")
	assert.Contains(t, view, "/")
}

func TestRoadmapDetail_GroupTabsSwitchToSolved(t *testing.T) {
	rd, db := newTestRoadmapDetail(t)
	ctx := context.Background()
	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusSolved))
	progress, err := db.GetAllProgress(ctx)
	require.NoError(t, err)
	rd.progress = progress
	rd.width = 120
	rd.height = 40

	view := rd.View()
	assert.Contains(t, view, "[Stages]")
	assert.Contains(t, view, "Solved")
	assert.Contains(t, view, "Contains Duplicate")

	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	view = rd.View()
	assert.Contains(t, view, "[Solved]")
	assert.Contains(t, view, "Solved Problems")
	assert.Contains(t, view, "Two Sum")
	assert.NotContains(t, view, "Contains Duplicate")
}

func TestRoadmapDetail_GroupTabsSwitchWithAngles(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(">")})
	assert.Equal(t, roadmapGroupSolved, rd.groupMode)

	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("<")})
	assert.Equal(t, roadmapGroupStages, rd.groupMode)
}

func TestRoadmapDetail_ViewHasFooter(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "j/k")
	assert.Contains(t, view, "h/l")
	assert.Contains(t, view, "enter")
	assert.NotContains(t, view, "graph")
	assert.Contains(t, view, "esc")
	assert.Contains(t, view, "quit")
}

func TestRoadmapDetail_Quit(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	_, cmd := rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	require.NotNil(t, cmd)
}

func TestRoadmapDetail_EscReturnsToDashboard(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	_, cmd := rd.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenDashboard, navigate.ScreenID)
}

func TestRoadmapDetail_BackspaceReturnsToDashboard(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	_, cmd := rd.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenDashboard, navigate.ScreenID)
}

func TestRoadmapDetail_GraphShortcutDoesNothing(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	before := rd.View()

	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})

	assert.Equal(t, before, rd.View())
	assert.NotContains(t, rd.View(), "Unlock Path")
}

func TestRoadmapDetail_FocusNavigation(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	if len(rd.problems) == 0 {
		t.Skip("no problems")
	}

	assert.Equal(t, 0, rd.focusIndex)
	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 1, rd.focusIndex)
	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 0, rd.focusIndex)
}

func TestRoadmapDetail_FocusClampsAtEnds(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	if len(rd.problems) == 0 {
		t.Skip("no problems")
	}

	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 0, rd.focusIndex)

	rd.focusIndex = len(rd.problems) - 1
	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, len(rd.problems)-1, rd.focusIndex)
}

func TestRoadmapDetail_FocusWalksAllProblemsInOrder(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	if len(rd.problems) < 3 {
		t.Skip("not enough problems")
	}

	first := rd.focusIndex
	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, first+1, rd.focusIndex)
	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, first+2, rd.focusIndex)
}

func TestRoadmapDetail_FocusWalksRenderedStageOrder(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	require.GreaterOrEqual(t, len(rd.problems), 2)

	assert.Equal(t, "Two Sum", rd.problems[0].Title)
	assert.Equal(t, "Contains Duplicate", rd.problems[1].Title)

	rd.focusIndex = 0
	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	assert.Equal(t, 1, rd.focusIndex)
	assert.Equal(t, "Contains Duplicate", rd.focusedProblem().Title)
	assert.NotEqual(t, "Valid Parentheses", rd.focusedProblem().Title)
}

func TestRoadmapDetail_EnterGoesToStageDetail(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	if len(rd.problems) == 0 {
		t.Skip("no problems")
	}

	rd.focusIndex = 0

	_, cmd := rd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenStageDetail, navigate.ScreenID)
	assert.NotEmpty(t, navigate.Stage)
}

func TestRoadmapDetail_WindowResize(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	rd.Update(tea.WindowSizeMsg{Width: 150, Height: 50})
	assert.Equal(t, 150, rd.width)
	assert.Equal(t, 50, rd.height)
}

func TestRoadmapDetail_ScrollsViewport(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120
	rd.height = 32

	before := rd.View()
	assert.Contains(t, before, "Arrays & Hashing")

	rd.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	after := rd.View()

	assert.NotEqual(t, before, after)
	assert.Contains(t, after, "more Stage above")
}

func TestRoadmapDetail_ScrolledViewportKeepsStageHeader(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	content, problemLineMap, _ := rd.buildStagesContent(rd.buildStageSolvedCount())
	require.Greater(t, len(content), 4)

	rd.scrollOffset = 2
	window := rd.windowLines(content, 6, problemLineMap)

	require.NotEmpty(t, window)
	assert.Contains(t, window[0], "Arrays & Hashing")
}

func TestRoadmapDetail_WindowSourceLinesAccountForStickyHeader(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	content := []string{"Stage", "one", "two", "three"}
	problemLineMap := map[int]int{1: 1, 2: 2, 3: 3}
	rd.scrollOffset = 2

	window, sources := rd.windowLinesWithSource(content, 3, problemLineMap)

	require.Equal(t, []string{"Stage", "two", "three"}, window)
	assert.Equal(t, []int{0, 2, 3}, sources)
}

func TestRoadmapDetail_ProblemLineMapIndexesOnlyFirstWrappedLine(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	require.NotEmpty(t, rd.problems)
	rd.width = 40
	rd.problems[0].Title = "This Title Is Intentionally Long Enough To Wrap Across Multiple Lines"

	content, problemLineMap, _ := rd.buildStagesContent(rd.buildStageSolvedCount())

	entries := 0
	for _, pid := range problemLineMap {
		if pid == rd.problems[0].ID {
			entries++
		}
	}

	assert.Greater(t, len(content), len(problemLineMap))
	assert.Equal(t, 1, entries)
}

func TestRoadmapDetail_MovingToLastProblemScrollsSelectionIntoView(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 190
	rd.height = 30
	require.NotEmpty(t, rd.problems)

	for rd.focusIndex < len(rd.problems)-1 {
		rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}

	focusedID := rd.focusedProblem().ID
	content, problemLineMap, _ := rd.buildStagesContent(rd.buildStageSolvedCount())
	visible := rd.stagesVisible()
	rd.clampScroll(len(content), visible)
	_, sources := rd.windowLinesWithSource(content, visible, problemLineMap)
	targetLine := -1
	for i := range content {
		if problemLineMap[i] == focusedID {
			targetLine = i
			break
		}
	}

	visibleFocused := false
	for _, source := range sources {
		if problemLineMap[source] == focusedID {
			visibleFocused = true
			break
		}
	}

	assert.True(t, visibleFocused, "focused final problem should be visible after navigation: id=%d title=%q target=%d total=%d visible=%d offset=%d sources=%v", focusedID, rd.focusedProblem().Title, targetLine, len(content), visible, rd.scrollOffset, sources)
}

func TestRoadmapDetail_RendersHeapProblemsInDeclaredHeapStage(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	content, _, _ := rd.buildStagesContent(rd.buildStageSolvedCount())
	view := strings.Join(content, "\n")

	assert.Contains(t, view, "Heap & Priority Queue")
	assert.Contains(t, view, "Last Stone Weight")
}

func TestRoadmapDetail_SolvedProblemStatus(t *testing.T) {
	rd, db := newTestRoadmapDetail(t)
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, db.SetProgress(ctx, 217, roadmap.StatusInProgress))

	rd.refreshProgress()
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "SOLVED")
	assert.Contains(t, view, "ACTIVE")
}

func TestRoadmapDetail_BlockerDisplay(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120
	rd.height = 40

	lockedIdx := -1
	for i, p := range rd.problems {
		if rd.effectiveStatus(p) == roadmap.StatusLocked && len(rd.missingPrerequisites(p)) > 0 {
			lockedIdx = i
			break
		}
	}
	require.NotEqual(t, -1, lockedIdx, "should have a locked problem with blockers")

	rd.focusIndex = lockedIdx
	view := rd.View()
	assert.Contains(t, view, "Blocked")
	assert.Contains(t, view, "Problem")
	assert.Contains(t, view, "Blocked by")
	assert.Contains(t, view, "- #")
}

func TestRoadmapDetail_ListPaneWidthsUseStableTwoColumnLayout(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 190

	left, right := rd.listPaneWidths()

	assert.Equal(t, 78, left)
	assert.Equal(t, 56, right)
}

func TestRoadmapDetail_ListPaneWidthsFitNarrowTwoColumnLayout(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 110

	left, right := rd.listPaneWidths()

	assert.Equal(t, 62, left)
	assert.Equal(t, 38, right)
	assert.LessOrEqual(t, left+right+2, rd.width-8)
}

func TestRoadmapDetail_EnterOnLockedProblemGoesToStage(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	lockedIdx := -1
	for i, p := range rd.problems {
		if rd.effectiveStatus(p) == roadmap.StatusLocked && len(rd.missingPrerequisites(p)) > 0 {
			lockedIdx = i
			break
		}
	}
	require.NotEqual(t, -1, lockedIdx, "should have a locked problem with blockers")

	rd.focusIndex = lockedIdx
	_, cmd := rd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenStageDetail, navigate.ScreenID)
	assert.NotEmpty(t, navigate.Stage)
}

func TestRoadmapDetail_UnlockedProblemShowsOverview(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	unlockedIdx := -1
	for i, p := range rd.problems {
		if rd.effectiveStatus(p) != roadmap.StatusLocked {
			unlockedIdx = i
			break
		}
	}
	require.NotEqual(t, -1, unlockedIdx, "should have an unlocked problem")

	rd.focusIndex = unlockedIdx
	view := rd.View()
	assert.Contains(t, view, "Overview")
	assert.NotContains(t, view, "Press enter to open the stage.")
}

func TestRoadmapDetail_LockedProblemsHaveBlockers(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	view := rd.View()
	lockedCount := 0
	for _, p := range rd.problems {
		status := rd.effectiveStatus(p)
		if status == roadmap.StatusLocked {
			lockedCount++
		}
	}
	require.Greater(t, lockedCount, 0, "should have locked problems")
	assert.Contains(t, view, "LOCKED")
}

func TestRoadmapDetail_NoThemeCycleShortcut(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	_, cmd := rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	assert.Nil(t, cmd)
}

func (s *RoadmapDetailScreen) refreshProgress() {
	progress, err := s.db.GetAllProgress(context.Background())
	if err == nil {
		s.progress = progress
	}
}
