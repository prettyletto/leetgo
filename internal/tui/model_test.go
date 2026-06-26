package tui

import (
	"context"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prettyletto/leetgo/internal/config"
	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestModel(t *testing.T) *Model {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	cfg := &config.Config{
		Workspace: t.TempDir(),
		Language:  "go",
	}

	m, err := NewModel(cfg, db)
	require.NoError(t, err)
	return m
}

func pressKey(m *Model, key string) {
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
}

func pressSpecial(m *Model, key tea.KeyType) {
	m.Update(tea.KeyMsg{Type: key})
}

func TestNewModel(t *testing.T) {
	m := newTestModel(t)
	assert.NotNil(t, m.list)
	assert.NotNil(t, m.heatmapView)
	assert.NotNil(t, m.statsBar)
	assert.Equal(t, viewList, m.viewMode)
}

func TestViewModeToggle_Heatmap(t *testing.T) {
	m := newTestModel(t)
	assert.Equal(t, viewList, m.viewMode)

	pressKey(m, "h")
	assert.Equal(t, viewHeatmap, m.viewMode)

	pressKey(m, "h")
	assert.Equal(t, viewList, m.viewMode)
}

func TestGraphShortcutIsDisabled(t *testing.T) {
	m := newTestModel(t)
	assert.Equal(t, viewList, m.viewMode)

	pressKey(m, "g")

	assert.Equal(t, viewList, m.viewMode)
	assert.NotContains(t, m.renderKeytips(), "unlock path")
}

func TestView_Renders(t *testing.T) {
	m := newTestModel(t)
	view := m.View()
	assert.Contains(t, view, "Leetgo")
	assert.Contains(t, view, "Prerequisites")
}

func TestRenderStatus_Verified(t *testing.T) {
	theme, _ := LookupTheme("rpg-skill-tree")
	assert.Contains(t, renderStatus(roadmap.StatusVerified, theme), "VERIFIED")
}

func TestRenderDifficulty_IsPlainOutsideProblemDetail(t *testing.T) {
	theme, _ := LookupTheme("rpg-skill-tree")

	rendered := renderDifficulty(roadmap.DifficultyMedium, theme)

	assert.Contains(t, rendered, "Medium")
	assert.NotContains(t, rendered, "\x1b[")
}

func TestStageFilter_CyclesAndClears(t *testing.T) {
	m := newTestModel(t)
	total := len(m.list.Items())

	pressSpecial(m, tea.KeyTab)
	assert.Equal(t, 0, m.stageFilter)
	assert.Less(t, len(m.list.Items()), total)
	assert.Contains(t, m.View(), "Stage:")

	pressKey(m, "0")
	assert.Equal(t, -1, m.stageFilter)
	assert.Len(t, m.list.Items(), total)
}

func TestView_Quitting(t *testing.T) {
	m := newTestModel(t)
	m.quitting = true
	view := m.View()
	assert.Contains(t, view, "Goodbye")
}

func TestHandleMarkSolved_AlreadySolved(t *testing.T) {
	m := newTestModel(t)
	ctx := context.Background()

	m.list.Select(0)
	item := m.list.SelectedItem().(problemItem)
	require.NoError(t, m.store.SetProgress(ctx, item.problem.ID, roadmap.StatusSolved))

	m.handleMarkSolved()
}

func TestLegacyHandleMarkSolvedAwardsNoXP(t *testing.T) {
	m := newTestModel(t)
	ctx := context.Background()
	m.list.Select(0)

	m.handleMarkSolved()

	stats, err := m.store.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalXP)
}

func TestLegacyAcceptedSubmitRewardIdempotent(t *testing.T) {
	m := newTestModel(t)
	ctx := context.Background()

	m.markAcceptedSubmission(1)
	m.markAcceptedSubmission(1)

	stats, err := m.store.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalXP)

	events, err := m.store.GetRewardEvents(ctx, 1)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "submit", events[0].Kind)
	assert.Equal(t, 3, events[0].XP)
	assert.Contains(t, m.notifications.Render(), "already claimed")
}

func TestHandleSelect_LockedProblem(t *testing.T) {
	m := newTestModel(t)

	for i := 0; i < len(m.list.Items()); i++ {
		item := m.list.Items()[i].(problemItem)
		if item.status == roadmap.StatusLocked {
			m.list.Select(i)
			m.handleSelect()
			assert.Contains(t, m.notifications.Render(), "locked")
			return
		}
	}
}

func TestStubExt(t *testing.T) {
	tests := []struct {
		lang string
		ext  string
	}{
		{"go", ".go"},
		{"python", ".py"},
		{"typescript", ".ts"},
		{"java", ".java"},
		{"cpp", ".cpp"},
		{"javascript", ".js"},
		{"rust", ".rs"},
		{"csharp", ".cs"},
		{"unknown", ".go"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.ext, stubExt(tt.lang))
	}
}

func TestLeetcodeLang(t *testing.T) {
	tests := []struct {
		lang string
		lc   string
	}{
		{"go", "golang"},
		{"python", "python3"},
		{"typescript", "typescript"},
		{"java", "java"},
		{"cpp", "cpp"},
		{"javascript", "javascript"},
		{"rust", "rust"},
		{"csharp", "csharp"},
		{"unknown", "golang"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.lc, leetcodeLang(tt.lang))
	}
}
