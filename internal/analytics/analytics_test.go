package analytics

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/prettyletto/leetgo/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAnalytics(t *testing.T) (*Analytics, store.Store) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := store.NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	problems := []*roadmap.Problem{
		{ID: 1, Title: "Two Sum", Category: "arrays", Difficulty: roadmap.DifficultyEasy},
		{ID: 2, Title: "Add Two Numbers", Category: "linked-list", Difficulty: roadmap.DifficultyMedium},
		{ID: 3, Title: "Longest Substring", Category: "sliding-window", Difficulty: roadmap.DifficultyMedium},
	}
	g := roadmap.NewGraph(problems)

	return New(s, g), s
}

func TestCategoryStats(t *testing.T) {
	a, s := newTestAnalytics(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 2, roadmap.StatusInProgress))

	require.NoError(t, s.RecordAttempt(ctx, &store.AttemptRecord{
		ProblemID: 1,
		Timestamp: time.Now(),
		Passed:    true,
	}))

	require.NoError(t, s.RecordAttempt(ctx, &store.AttemptRecord{
		ProblemID: 2,
		Timestamp: time.Now(),
		Passed:    false,
	}))
	require.NoError(t, s.RecordAttempt(ctx, &store.AttemptRecord{
		ProblemID: 2,
		Timestamp: time.Now(),
		Passed:    false,
	}))

	stats, err := a.CategoryStats(ctx)
	require.NoError(t, err)
	require.Len(t, stats, 3)

	var arrays, linkedList CategoryStats
	for _, s := range stats {
		if s.Category == "arrays" {
			arrays = s
		} else if s.Category == "linked-list" {
			linkedList = s
		}
	}

	assert.Equal(t, 1, arrays.Solved)
	assert.Equal(t, 1, arrays.Attempts)
	assert.Equal(t, 0, arrays.Failures)
	assert.Equal(t, 0.0, arrays.FailRate)

	assert.Equal(t, 0, linkedList.Solved)
	assert.Equal(t, 2, linkedList.Attempts)
	assert.Equal(t, 2, linkedList.Failures)
	assert.Equal(t, 1.0, linkedList.FailRate)
}

func TestDetectWeaknesses(t *testing.T) {
	a, s := newTestAnalytics(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		require.NoError(t, s.RecordAttempt(ctx, &store.AttemptRecord{
			ProblemID:    2,
			Timestamp:    time.Now(),
			Passed:       false,
			SelfReported: roadmap.DifficultyHard,
		}))
	}

	weaknesses, err := a.DetectWeaknesses(ctx)
	require.NoError(t, err)

	assert.NotEmpty(t, weaknesses)
	assert.Equal(t, roadmap.Category("linked-list"), weaknesses[0].Category)
}

func TestDetectWeaknesses_MinAttempts(t *testing.T) {
	a, s := newTestAnalytics(t)
	ctx := context.Background()

	require.NoError(t, s.RecordAttempt(ctx, &store.AttemptRecord{
		ProblemID: 1,
		Timestamp: time.Now(),
		Passed:    false,
	}))

	weaknesses, err := a.DetectWeaknesses(ctx)
	require.NoError(t, err)
	assert.Empty(t, weaknesses)
}
