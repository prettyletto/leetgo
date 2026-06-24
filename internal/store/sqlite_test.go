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

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewSQLiteStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })
	return s
}

func TestProgress(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	p, err := s.GetProgress(ctx, 1)
	require.NoError(t, err)
	assert.Nil(t, p)

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusInProgress))
	p, err = s.GetProgress(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, roadmap.StatusInProgress, p.Status)

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	p, err = s.GetProgress(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, roadmap.StatusSolved, p.Status)
}

func TestGetAllProgress(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 49, roadmap.StatusInProgress))

	all, err := s.GetAllProgress(ctx)
	require.NoError(t, err)
	assert.Equal(t, roadmap.StatusSolved, all[1])
	assert.Equal(t, roadmap.StatusInProgress, all[49])
}

func TestRecordAndGetAttempts(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusInProgress))

	a := &AttemptRecord{
		ProblemID:    1,
		Timestamp:    time.Now(),
		Duration:     5 * time.Minute,
		Passed:       true,
		SelfReported: roadmap.DifficultyEasy,
	}
	require.NoError(t, s.RecordAttempt(ctx, a))

	attempts, err := s.GetAttempts(ctx, 1)
	require.NoError(t, err)
	require.Len(t, attempts, 1)
	assert.True(t, attempts[0].Passed)
	assert.Equal(t, 5*time.Minute, attempts[0].Duration)
}

func TestRecordAndGetSolveLogs(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	submittedAt := time.Now().UTC().Truncate(time.Second)

	require.NoError(t, s.RecordSolveLog(ctx, &SolveLogRecord{
		ProblemID:   1,
		Slug:        "two-sum",
		Language:    "golang",
		Status:      "Accepted",
		StatusCode:  10,
		Runtime:     "1 ms",
		Memory:      "2 MB",
		TotalTests:  63,
		PassedTests: 63,
		SubmittedAt: submittedAt,
	}))

	logs, err := s.GetSolveLogs(ctx)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, 1, logs[0].ProblemID)
	assert.Equal(t, "two-sum", logs[0].Slug)
	assert.Equal(t, "golang", logs[0].Language)
	assert.Equal(t, "Accepted", logs[0].Status)
	assert.Equal(t, 10, logs[0].StatusCode)
	assert.Equal(t, "1 ms", logs[0].Runtime)
	assert.Equal(t, 63, logs[0].PassedTests)
}

func TestXPAndLevel(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	stats, err := s.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.TotalXP)
	assert.Equal(t, 1, stats.Level)

	require.NoError(t, s.AddXP(ctx, 50))
	stats, err = s.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 50, stats.TotalXP)
	assert.Equal(t, 1, stats.Level)

	require.NoError(t, s.AddXP(ctx, 60))
	stats, err = s.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 110, stats.TotalXP)
	assert.Equal(t, 2, stats.Level)
}

func TestStatsVerifiedAndSolved(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	stats, err := s.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, stats.Verified)
	assert.Equal(t, 0, stats.Solved)

	require.NoError(t, s.SetProgress(ctx, 1, roadmap.StatusVerified))
	require.NoError(t, s.SetProgress(ctx, 2, roadmap.StatusSolved))
	require.NoError(t, s.SetProgress(ctx, 3, roadmap.StatusVerified))
	require.NoError(t, s.SetProgress(ctx, 4, roadmap.StatusInProgress))

	stats, err = s.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.Verified)
	assert.Equal(t, 1, stats.Solved)
}

func TestXPToLevel(t *testing.T) {
	tests := []struct {
		xp    int
		level int
	}{
		{0, 1},
		{99, 1},
		{100, 2},
		{299, 2},
		{300, 3},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.level, XPToLevel(tt.xp), "xp=%d", tt.xp)
	}
}

func TestStreak(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.UpdateStreak(ctx))
	stats, err := s.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Streak)

	require.NoError(t, s.UpdateStreak(ctx))
	stats, err = s.GetStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.Streak)

	days, err := s.GetStreakDays(ctx)
	require.NoError(t, err)
	assert.Len(t, days, 1)
}

func TestAchievements(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	achievements, err := s.GetAchievements(ctx)
	require.NoError(t, err)
	assert.Empty(t, achievements)

	require.NoError(t, s.UnlockAchievement(ctx, "first_solve"))
	require.NoError(t, s.UnlockAchievement(ctx, "streak_7"))

	achievements, err = s.GetAchievements(ctx)
	require.NoError(t, err)
	assert.Len(t, achievements, 2)
	assert.Contains(t, achievements, "first_solve")
	assert.Contains(t, achievements, "streak_7")

	require.NoError(t, s.UnlockAchievement(ctx, "first_solve"))
	achievements, err = s.GetAchievements(ctx)
	require.NoError(t, err)
	assert.Len(t, achievements, 2)
}

func TestLevelToXP(t *testing.T) {
	assert.Equal(t, 0, LevelToXP(1))
	assert.Equal(t, 100, LevelToXP(2))
	assert.Equal(t, 300, LevelToXP(3))
	assert.Equal(t, 600, LevelToXP(4))
}

func TestXPForDifficulty(t *testing.T) {
	assert.Equal(t, 10, XPForDifficulty(roadmap.DifficultyEasy))
	assert.Equal(t, 25, XPForDifficulty(roadmap.DifficultyMedium))
	assert.Equal(t, 50, XPForDifficulty(roadmap.DifficultyHard))
	assert.Equal(t, 10, XPForDifficulty("unknown"))
}

func TestRewardEvents(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	events, err := s.GetRewardEvents(ctx, 1)
	require.NoError(t, err)
	assert.Empty(t, events)

	has, err := s.HasRewardEvent(ctx, 1, "verify")
	require.NoError(t, err)
	assert.False(t, has)

	e := &RewardEvent{ProblemID: 1, Kind: "verify", XP: 7}
	require.NoError(t, s.RecordRewardEvent(ctx, e))
	assert.False(t, e.CreatedAt.IsZero())

	has, err = s.HasRewardEvent(ctx, 1, "verify")
	require.NoError(t, err)
	assert.True(t, has)

	has, err = s.HasRewardEvent(ctx, 1, "submit")
	require.NoError(t, err)
	assert.False(t, has)

	events, err = s.GetRewardEvents(ctx, 1)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "verify", events[0].Kind)
	assert.Equal(t, 7, events[0].XP)
}

func TestRewardEventIdempotency(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e := &RewardEvent{ProblemID: 1, Kind: "verify", XP: 7}
	require.NoError(t, s.RecordRewardEvent(ctx, e))

	e2 := &RewardEvent{ProblemID: 1, Kind: "verify", XP: 99}
	require.NoError(t, s.RecordRewardEvent(ctx, e2))

	events, err := s.GetRewardEvents(ctx, 1)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, 7, events[0].XP)

	require.NoError(t, s.RecordRewardEvent(ctx, &RewardEvent{ProblemID: 1, Kind: "submit", XP: 3}))
	events, err = s.GetRewardEvents(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestRewardEventMultipleProblems(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.RecordRewardEvent(ctx, &RewardEvent{ProblemID: 1, Kind: "verify", XP: 7}))
	require.NoError(t, s.RecordRewardEvent(ctx, &RewardEvent{ProblemID: 1, Kind: "submit", XP: 3}))
	require.NoError(t, s.RecordRewardEvent(ctx, &RewardEvent{ProblemID: 49, Kind: "verify", XP: 17}))

	events, err := s.GetRewardEvents(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, events, 2)

	events, err = s.GetRewardEvents(ctx, 49)
	require.NoError(t, err)
	assert.Len(t, events, 1)

	events, err = s.GetRewardEvents(ctx, 999)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestSolveProvenance_RecordAndGet(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	sp, err := s.GetSolveProvenance(ctx, 1)
	require.NoError(t, err)
	assert.Nil(t, sp)

	err = s.RecordSolveProvenance(ctx, &SolveProvenance{
		ProblemID: 1,
		Kind:      "manual",
		Note:      "Solved in browser",
	})
	require.NoError(t, err)

	sp, err = s.GetSolveProvenance(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, sp)
	assert.Equal(t, 1, sp.ProblemID)
	assert.Equal(t, "manual", sp.Kind)
	assert.Equal(t, "Solved in browser", sp.Note)
	assert.Nil(t, sp.SolveLogID)
	assert.False(t, sp.SolvedAt.IsZero())
}

func TestSolveProvenance_AcceptedWithLogID(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.RecordSolveLog(ctx, &SolveLogRecord{
		ProblemID:   1,
		Slug:        "two-sum",
		Language:    "golang",
		Status:      "Accepted",
		StatusCode:  10,
		TotalTests:  63,
		PassedTests: 63,
	}))

	logs, err := s.GetSolveLogs(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, logs)
	logID := logs[0].ID

	err = s.RecordSolveProvenance(ctx, &SolveProvenance{
		ProblemID:  1,
		Kind:       "accepted",
		SolveLogID: &logID,
	})
	require.NoError(t, err)

	sp, err := s.GetSolveProvenance(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, sp)
	assert.Equal(t, "accepted", sp.Kind)
	require.NotNil(t, sp.SolveLogID)
	assert.Equal(t, logID, *sp.SolveLogID)
}

func TestSolveProvenance_UpgradeManualToAccepted(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.RecordSolveProvenance(ctx, &SolveProvenance{
		ProblemID: 1,
		Kind:      "manual",
		Note:      "temp solve",
	}))

	require.NoError(t, s.RecordSolveLog(ctx, &SolveLogRecord{
		ProblemID:  1,
		Slug:       "two-sum",
		Language:   "golang",
		Status:     "Accepted",
		StatusCode: 10,
	}))

	logs, _ := s.GetSolveLogs(ctx)
	logID := logs[0].ID

	require.NoError(t, s.RecordSolveProvenance(ctx, &SolveProvenance{
		ProblemID:  1,
		Kind:       "accepted",
		Note:       "temp solve",
		SolveLogID: &logID,
	}))

	sp, err := s.GetSolveProvenance(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, sp)
	assert.Equal(t, "accepted", sp.Kind)
	assert.Equal(t, "temp solve", sp.Note)
}

func TestSolveProvenance_GetAll(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.RecordSolveProvenance(ctx, &SolveProvenance{ProblemID: 1, Kind: "manual"}))
	require.NoError(t, s.RecordSolveProvenance(ctx, &SolveProvenance{ProblemID: 49, Kind: "accepted"}))

	all, err := s.GetSolveProvenanceAll(ctx)
	require.NoError(t, err)
	require.Len(t, all, 2)
	assert.Equal(t, "manual", all[1].Kind)
	assert.Equal(t, "accepted", all[49].Kind)
}

func TestGetSolveLogsForProblem(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.RecordSolveLog(ctx, &SolveLogRecord{
		ProblemID:  1,
		Slug:       "two-sum",
		Language:   "golang",
		Status:     "Accepted",
		StatusCode: 10,
	}))
	require.NoError(t, s.RecordSolveLog(ctx, &SolveLogRecord{
		ProblemID:  49,
		Slug:       "group-anagrams",
		Language:   "golang",
		Status:     "Wrong Answer",
		StatusCode: 11,
	}))

	logs, err := s.GetSolveLogsForProblem(ctx, 1)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, 1, logs[0].ProblemID)

	logs, err = s.GetSolveLogsForProblem(ctx, 999)
	require.NoError(t, err)
	assert.Empty(t, logs)
}

func TestReviewCycles_CreateAndGet(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	err := s.CreateReviewCycle(ctx, &ReviewCycle{
		ProblemID: 1,
		Reason:    "weakness",
		RoadmapID: "test",
	})
	require.NoError(t, err)

	cycles, err := s.GetReviewCycles(ctx)
	require.NoError(t, err)
	require.Len(t, cycles, 1)
	assert.Equal(t, 1, cycles[0].ProblemID)
	assert.Equal(t, "weakness", cycles[0].Reason)
	assert.Nil(t, cycles[0].CompletedAt)
	assert.Nil(t, cycles[0].RewardedAt)
}

func TestReviewCycles_GetForProblem(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.CreateReviewCycle(ctx, &ReviewCycle{ProblemID: 1, Reason: "weakness"}))
	require.NoError(t, s.CreateReviewCycle(ctx, &ReviewCycle{ProblemID: 49, Reason: "failed_attempts"}))

	cycles, err := s.GetReviewCyclesForProblem(ctx, 1)
	require.NoError(t, err)
	require.Len(t, cycles, 1)
	assert.Equal(t, 1, cycles[0].ProblemID)

	cycles, err = s.GetReviewCyclesForProblem(ctx, 999)
	require.NoError(t, err)
	assert.Empty(t, cycles)
}

func TestReviewCycles_CompleteAndReward(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.CreateReviewCycle(ctx, &ReviewCycle{ProblemID: 1, Reason: "weakness"}))

	cycles, _ := s.GetReviewCycles(ctx)
	require.Len(t, cycles, 1)
	id := cycles[0].ID

	require.NoError(t, s.CompleteReviewCycle(ctx, id))
	err := s.RewardReviewCycle(ctx, id)
	require.NoError(t, err)

	cycles, _ = s.GetReviewCycles(ctx)
	require.Len(t, cycles, 1)
	assert.NotNil(t, cycles[0].CompletedAt)
	assert.NotNil(t, cycles[0].RewardedAt)
}

func TestReviewCycles_RewardIdempotent(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	require.NoError(t, s.CreateReviewCycle(ctx, &ReviewCycle{ProblemID: 1, Reason: "weakness"}))
	cycles, _ := s.GetReviewCycles(ctx)
	id := cycles[0].ID

	require.NoError(t, s.CompleteReviewCycle(ctx, id))
	require.NoError(t, s.RewardReviewCycle(ctx, id))

	err := s.RewardReviewCycle(ctx, id)
	require.NoError(t, err, "rewarding again should not error")

	cycles, _ = s.GetReviewCycles(ctx)
	require.Len(t, cycles, 1)
	firstRewarded := cycles[0].RewardedAt
	require.NoError(t, s.RewardReviewCycle(ctx, id))
	cycles, _ = s.GetReviewCycles(ctx)
	assert.Equal(t, firstRewarded.Unix(), cycles[0].RewardedAt.Unix(), "re-award should not change timestamp")
}
