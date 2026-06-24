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

func newTestStageDetail(t *testing.T, stageID string) (*StageDetailScreen, store.Store) {
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

	sd := NewStageDetailScreen(cfg, theme, db, rm, stageID)
	return sd, db
}

func TestStageDetail_ViewShowsStageTitle(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "Arrays & Hashing")
}

func TestStageDetail_RPGZoneLabels(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "Zone: Arrays & Hashing")
	assert.Contains(t, view, "Encounter Grid")
	assert.Contains(t, view, "Recommended Encounter")
}

func TestStageDetail_ReviewShrine(t *testing.T) {
	sd, db := newTestStageDetail(t, "arrays-hashing")
	ctx := context.Background()
	require.NoError(t, db.CreateReviewCycle(ctx, &store.ReviewCycle{ProblemID: 1, Reason: "weakness", RoadmapID: "from-zero-to-hero"}))
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "Review Shrine")
	assert.Contains(t, view, "Two Sum")
}

func TestStageDetail_ReviewShrinePlainSymbols(t *testing.T) {
	sd, db := newTestStageDetail(t, "arrays-hashing")
	sd.cfg.SymbolMode = "plain"
	ctx := context.Background()
	require.NoError(t, db.CreateReviewCycle(ctx, &store.ReviewCycle{ProblemID: 1, Reason: "weakness", RoadmapID: "from-zero-to-hero"}))
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "Review Shrine")
	assert.Contains(t, view, "R #1 Two Sum")
	assert.NotContains(t, view, "↻")
}

func TestStageDetail_ViewShowsCompletion(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "Solved:")
	assert.Contains(t, view, "Verified:")
	assert.Contains(t, view, "Total:")
}

func TestStageDetail_ViewShowsProblems(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "Two Sum")
	assert.Contains(t, view, "Contains Duplicate")
	assert.Contains(t, view, "Group Anagrams")
}

func TestStageDetail_ViewShowsBlockers(t *testing.T) {
	sd, _ := newTestStageDetail(t, "sliding-window")
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "blocked by")
	assert.Contains(t, view, "LOCKED")
}

func TestStageDetail_ViewShowsAvailable(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "READY")
}

func TestStageDetail_ViewHasFooter(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "j/k")
	assert.Contains(t, view, "enter")
	assert.Contains(t, view, "esc")
	assert.Contains(t, view, "quit")
}

func TestStageDetail_Quit(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")

	_, cmd := sd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	require.NotNil(t, cmd)
}

func TestStageDetail_EscReturnsToRoadmapDetail(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")

	_, cmd := sd.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenRoadmapDetail, navigate.ScreenID)
}

func TestStageDetail_BackspaceReturnsToRoadmapDetail(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")

	_, cmd := sd.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenRoadmapDetail, navigate.ScreenID)
}

func TestStageDetail_FocusNavigation(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")

	if len(sd.problems) == 0 {
		t.Skip("no problems")
	}

	assert.Equal(t, 0, sd.focusIndex)
	sd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 1, sd.focusIndex)
	sd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, 0, sd.focusIndex)
}

func TestStageDetail_FocusWraps(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")

	if len(sd.problems) == 0 {
		t.Skip("no problems")
	}

	sd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	assert.Equal(t, len(sd.problems)-1, sd.focusIndex)

	sd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	assert.Equal(t, 0, sd.focusIndex)
}

func TestStageDetail_EnterGoesToProblemDetail(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")

	if len(sd.problems) == 0 {
		t.Skip("no problems")
	}

	_, cmd := sd.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)

	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenProblemDetail, navigate.ScreenID)
	assert.Greater(t, navigate.ProblemID, 0)
}

func TestStageDetail_SolvedProblems(t *testing.T) {
	sd, db := newTestStageDetail(t, "arrays-hashing")
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, db.SetProgress(ctx, 217, roadmap.StatusInProgress))

	sd.refreshProgress()
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "SOLVED")
	assert.Contains(t, view, "ACTIVE")
}

func TestStageDetail_InProgressEmphasis(t *testing.T) {
	sd, db := newTestStageDetail(t, "arrays-hashing")
	ctx := context.Background()

	require.NoError(t, db.SetProgress(ctx, 49, roadmap.StatusInProgress))

	sd.refreshProgress()
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "ACTIVE")
}

func TestStageDetail_WindowResize(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")

	sd.Update(tea.WindowSizeMsg{Width: 150, Height: 50})
	assert.Equal(t, 150, sd.width)
	assert.Equal(t, 50, sd.height)
}

func TestStageDetail_ThemeCycle(t *testing.T) {
	sd, _ := newTestStageDetail(t, "arrays-hashing")

	_, cmd := sd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(ThemeChangedMsg)
	assert.True(t, ok)
}

func TestStageDetail_UnknownStageShowsID(t *testing.T) {
	sd, _ := newTestStageDetail(t, "nonexistent-stage")
	sd.width = 120

	view := sd.View()
	assert.Contains(t, view, "Nonexistent-stage")
}

func TestStageDetail_EmptyStageShowsNoProblems(t *testing.T) {
	sd, db := newTestStageDetail(t, "bit-manipulation")
	require.NotNil(t, db)

	sd.width = 120
	view := sd.View()
	assert.Contains(t, view, "Bit Manipulation")
	assert.Contains(t, view, "No problems")
}

func (s *StageDetailScreen) refreshProgress() {
	progress, err := s.db.GetAllProgress(context.Background())
	if err == nil {
		s.progress = progress
	}
}
