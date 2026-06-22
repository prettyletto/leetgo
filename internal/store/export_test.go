package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/prettyletto/leetgo/internal/roadmap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportImport(t *testing.T) {
	ctx := context.Background()

	srcPath := filepath.Join(t.TempDir(), "src.db")
	src, err := NewSQLiteStore(srcPath)
	require.NoError(t, err)
	defer src.Close()

	require.NoError(t, src.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, src.SetProgress(ctx, 49, roadmap.StatusInProgress))
	require.NoError(t, src.SetProgress(ctx, 242, roadmap.StatusVerified))
	require.NoError(t, src.AddXP(ctx, 150))
	require.NoError(t, src.UpdateStreak(ctx))
	require.NoError(t, src.UnlockAchievement(ctx, "first_solve"))
	require.NoError(t, src.RecordAttempt(ctx, &AttemptRecord{
		ProblemID:    1,
		Timestamp:    time.Now(),
		Duration:     3 * time.Minute,
		Passed:       true,
		SelfReported: roadmap.DifficultyEasy,
	}))
	require.NoError(t, src.RecordSolveLog(ctx, &SolveLogRecord{
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
	require.NoError(t, src.RecordRewardEvent(ctx, &RewardEvent{ProblemID: 1, Kind: "verify", XP: 7}))
	require.NoError(t, src.RecordRewardEvent(ctx, &RewardEvent{ProblemID: 1, Kind: "submit", XP: 3}))
	require.NoError(t, src.RecordRewardEvent(ctx, &RewardEvent{ProblemID: 999, Kind: "verify", XP: 7}))

	data, err := src.Export(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, data.Version)
	assert.Len(t, data.Progress, 3)
	assert.Equal(t, 150, data.XP)
	assert.Contains(t, data.Achievements, "first_solve")
	assert.Len(t, data.Attempts, 1)
	assert.Len(t, data.SolveLogs, 1)
	assert.Len(t, data.RewardEvents, 3)

	dstPath := filepath.Join(t.TempDir(), "dst.db")
	dst, err := NewSQLiteStore(dstPath)
	require.NoError(t, err)
	defer dst.Close()

	require.NoError(t, dst.Import(ctx, data))

	progress, err := dst.GetAllProgress(ctx)
	require.NoError(t, err)
	assert.Equal(t, roadmap.StatusSolved, progress[1])
	assert.Equal(t, roadmap.StatusInProgress, progress[49])
	assert.Equal(t, roadmap.StatusVerified, progress[242])

	stats, err := dst.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 150, stats.TotalXP)

	achievements, err := dst.GetAchievements(ctx)
	require.NoError(t, err)
	assert.Contains(t, achievements, "first_solve")

	logs, err := dst.GetSolveLogs(ctx)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, "two-sum", logs[0].Slug)

	rewardEvents, err := dst.GetRewardEvents(ctx, 1)
	require.NoError(t, err)
	require.Len(t, rewardEvents, 2)

	hasVerify, err := dst.HasRewardEvent(ctx, 1, "verify")
	require.NoError(t, err)
	assert.True(t, hasVerify)

	hasSubmit, err := dst.HasRewardEvent(ctx, 1, "submit")
	require.NoError(t, err)
	assert.True(t, hasSubmit)

	orphanEvents, err := dst.GetRewardEvents(ctx, 999)
	require.NoError(t, err)
	require.Len(t, orphanEvents, 1)
	assert.Equal(t, "verify", orphanEvents[0].Kind)
}

func TestExportToFile(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewSQLiteStore(dbPath)
	require.NoError(t, err)
	defer s.Close()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.AddXP(ctx, 50))

	exportPath := filepath.Join(t.TempDir(), "export.json")
	require.NoError(t, ExportToFile(ctx, s, exportPath))

	assert.FileExists(t, exportPath)
}

func TestImportFromFile(t *testing.T) {
	ctx := context.Background()

	srcPath := filepath.Join(t.TempDir(), "src.db")
	src, err := NewSQLiteStore(srcPath)
	require.NoError(t, err)
	defer src.Close()

	require.NoError(t, src.SetProgress(ctx, 1, roadmap.StatusSolved))

	exportPath := filepath.Join(t.TempDir(), "export.json")
	require.NoError(t, ExportToFile(ctx, src, exportPath))

	dstPath := filepath.Join(t.TempDir(), "dst.db")
	dst, err := NewSQLiteStore(dstPath)
	require.NoError(t, err)
	defer dst.Close()

	require.NoError(t, ImportFromFile(ctx, dst, exportPath))

	progress, err := dst.GetAllProgress(ctx)
	require.NoError(t, err)
	assert.Equal(t, roadmap.StatusSolved, progress[1])
}
