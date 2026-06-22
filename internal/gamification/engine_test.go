package gamification

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestEngine(t *testing.T) (*Engine, store.Store) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	problems := []*roadmap.Problem{
		{ID: 1, Title: "Two Sum", Difficulty: roadmap.DifficultyEasy},
		{ID: 76, Title: "Minimum Window", Difficulty: roadmap.DifficultyHard},
	}
	g := roadmap.NewGraph(problems)

	return NewEngine(s, g), s
}

func TestOnProblemSolved_FirstSolve(t *testing.T) {
	engine, s := newTestEngine(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.AddXP(ctx, 10))
	require.NoError(t, s.UpdateStreak(ctx))

	unlocked, err := engine.OnProblemSolved(ctx, 1)
	require.NoError(t, err)
	assert.Contains(t, unlocked, "first_solve")
}

func TestOnProblemSolved_FirstHard(t *testing.T) {
	engine, s := newTestEngine(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 76, roadmap.StatusSolved))
	require.NoError(t, s.AddXP(ctx, 50))
	require.NoError(t, s.UpdateStreak(ctx))

	unlocked, err := engine.OnProblemSolved(ctx, 76)
	require.NoError(t, err)
	assert.Contains(t, unlocked, "first_hard")
}

func TestOnProblemSolved_LevelAchievements(t *testing.T) {
	engine, s := newTestEngine(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.AddXP(ctx, 1000))
	require.NoError(t, s.UpdateStreak(ctx))

	unlocked, err := engine.OnProblemSolved(ctx, 1)
	require.NoError(t, err)
	assert.Contains(t, unlocked, "level_5")
}
