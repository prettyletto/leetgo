package store

import (
	"context"
	"time"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type Progress struct {
	ProblemID int
	Status    roadmap.Status
	UpdatedAt time.Time
}

type AttemptRecord struct {
	ID           int
	ProblemID    int
	Timestamp    time.Time
	Duration     time.Duration
	Passed       bool
	SelfReported roadmap.Difficulty
}

type SolveLogRecord struct {
	ID          int
	ProblemID   int
	Slug        string
	Language    string
	Status      string
	StatusCode  int
	Runtime     string
	Memory      string
	TotalTests  int
	PassedTests int
	Error       string
	Note        string
	SubmittedAt time.Time
}

type RewardEvent struct {
	ProblemID int
	Kind      string
	XP        int
	CreatedAt time.Time
}

type Stats struct {
	TotalXP       int
	Level         int
	Streak        int
	LongestStreak int
	Verified      int
	Solved        int
	Total         int
}

type SolveProvenance struct {
	ProblemID  int
	Kind       string
	Note       string
	SolveLogID *int
	SolvedAt   time.Time
}

type ReviewCycle struct {
	ID          int
	ProblemID   int
	Reason      string
	RoadmapID   string
	CreatedAt   time.Time
	CompletedAt *time.Time
	RewardedAt  *time.Time
}

type Store interface {
	Close() error

	GetProgress(ctx context.Context, problemID int) (*Progress, error)
	SetProgress(ctx context.Context, problemID int, status roadmap.Status) error
	GetAllProgress(ctx context.Context) (map[int]roadmap.Status, error)

	RecordAttempt(ctx context.Context, attempt *AttemptRecord) error
	GetAttempts(ctx context.Context, problemID int) ([]*AttemptRecord, error)
	RecordSolveLog(ctx context.Context, log *SolveLogRecord) error
	GetSolveLogs(ctx context.Context) ([]*SolveLogRecord, error)
	GetSolveLogsForProblem(ctx context.Context, problemID int) ([]*SolveLogRecord, error)

	RecordSolveProvenance(ctx context.Context, sp *SolveProvenance) error
	GetSolveProvenance(ctx context.Context, problemID int) (*SolveProvenance, error)
	GetSolveProvenanceAll(ctx context.Context) (map[int]*SolveProvenance, error)

	CreateReviewCycle(ctx context.Context, rc *ReviewCycle) error
	GetReviewCycles(ctx context.Context) ([]*ReviewCycle, error)
	GetReviewCyclesForProblem(ctx context.Context, problemID int) ([]*ReviewCycle, error)
	CompleteReviewCycle(ctx context.Context, id int) error
	RewardReviewCycle(ctx context.Context, id int) error

	GetStats(ctx context.Context) (*Stats, error)
	AddXP(ctx context.Context, amount int) error

	UpdateStreak(ctx context.Context) error
	GetStreakDays(ctx context.Context) ([]time.Time, error)

	UnlockAchievement(ctx context.Context, id string) error
	GetAchievements(ctx context.Context) ([]string, error)

	RecordRewardEvent(ctx context.Context, event *RewardEvent) error
	HasRewardEvent(ctx context.Context, problemID int, kind string) (bool, error)
	GetRewardEvents(ctx context.Context, problemID int) ([]*RewardEvent, error)
}
