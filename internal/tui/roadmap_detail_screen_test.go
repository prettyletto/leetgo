package tui

import (
	"context"
	"path/filepath"
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

func TestRoadmapDetail_ViewShowsTagline(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, rd.roadmap.Tagline)
}

func TestRoadmapDetail_ViewShowsStages(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	view := rd.View()

	groups := rd.groupProblemsByStage()
	for _, stage := range rd.roadmap.Stages {
		if problems, ok := groups[stage.ID]; ok && len(problems) > 0 {
			assert.Contains(t, view, stage.Title, "should show stage %s with problems", stage.Title)
		}
	}
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

func TestRoadmapDetail_ViewHasFooter(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "j/k")
	assert.Contains(t, view, "enter")
	assert.Contains(t, view, "graph")
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

func TestRoadmapDetail_ToggleViewMode(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	assert.Equal(t, rdViewList, rd.viewMode)

	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	assert.Equal(t, rdViewGraph, rd.viewMode)

	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	assert.Equal(t, rdViewList, rd.viewMode)
}

func TestRoadmapDetail_GraphViewShowsUnlockPath(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	rd.viewMode = rdViewGraph
	rd.width = 120

	view := rd.View()
	assert.Contains(t, view, "Unlock Path")
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

func TestRoadmapDetail_FocusWraps(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	if len(rd.problems) == 0 {
		t.Skip("no problems")
	}

	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, len(rd.problems)-1, rd.focusIndex)

	rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 0, rd.focusIndex)
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

	view := rd.View()
	assert.Contains(t, view, "blocked by")
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

func TestRoadmapDetail_ThemeCycle(t *testing.T) {
	rd, _ := newTestRoadmapDetail(t)

	_, cmd := rd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(ThemeChangedMsg)
	assert.True(t, ok)
}

func (s *RoadmapDetailScreen) refreshProgress() {
	progress, err := s.db.GetAllProgress(context.Background())
	if err == nil {
		s.progress = progress
	}
}
