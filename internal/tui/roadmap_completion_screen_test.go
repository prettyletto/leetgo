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

func TestRoadmapCompletion_ViewShowsCompletionSummary(t *testing.T) {
	db, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	rm, err := catalog.LoadRoadmap("from-zero-to-hero")
	require.NoError(t, err)
	for id := range rm.Graph.Problems {
		require.NoError(t, db.SetProgress(context.Background(), id, roadmap.StatusSolved))
	}
	require.NoError(t, db.RecordSolveProvenance(context.Background(), &store.SolveProvenance{ProblemID: 1, Kind: "manual"}))

	theme, err := LookupTheme("rpg-skill-tree")
	require.NoError(t, err)
	s := NewRoadmapCompletionScreen(&config.Config{Theme: "rpg-skill-tree"}, theme, db, rm)

	view := s.View()
	assert.Contains(t, view, "Roadmap Completion")
	assert.Contains(t, view, "Reward")
	assert.Contains(t, view, "Roadmap complete")
	assert.Contains(t, view, "Manual Solves")
	assert.Contains(t, view, "Actions")
}

func TestRoadmapCompletion_EscReturnsToDashboard(t *testing.T) {
	db, err := store.NewSQLiteStore(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	rm, err := catalog.LoadRoadmap("from-zero-to-hero")
	require.NoError(t, err)
	theme, err := LookupTheme("rpg-skill-tree")
	require.NoError(t, err)
	s := NewRoadmapCompletionScreen(&config.Config{Theme: "rpg-skill-tree"}, theme, db, rm)

	_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	require.NotNil(t, cmd)
	msg := cmd()
	navigate, ok := msg.(NavigateMsg)
	require.True(t, ok)
	assert.Equal(t, ScreenDashboard, navigate.ScreenID)
}
