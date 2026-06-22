package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/prettyletto/leetgo/internal/roadmap"
)

type ExportData struct {
	Version              int                 `json:"version"`
	ExportedAt           time.Time           `json:"exported_at"`
	ExportIdentity       string              `json:"export_identity,omitempty"`
	ExportIdentitySource string              `json:"export_identity_source,omitempty"`
	Progress             []ExportProgress    `json:"progress"`
	Attempts             []ExportAttempt     `json:"attempts"`
	SolveLogs            []ExportSolveLog    `json:"solve_logs"`
	RewardEvents         []ExportRewardEvent `json:"reward_events,omitempty"`
	XP                   int                 `json:"xp"`
	Streak               ExportStreak        `json:"streak"`
	StreakDays           []string            `json:"streak_days"`
	Achievements         []string            `json:"achievements"`
}

type ExportProgress struct {
	ProblemID int    `json:"problem_id"`
	Status    string `json:"status"`
}

type ExportAttempt struct {
	ProblemID    int    `json:"problem_id"`
	Timestamp    string `json:"timestamp"`
	DurationMs   int64  `json:"duration_ms"`
	Passed       bool   `json:"passed"`
	SelfReported string `json:"self_reported,omitempty"`
}

type ExportSolveLog struct {
	ProblemID   int    `json:"problem_id"`
	Slug        string `json:"slug"`
	Language    string `json:"language"`
	Status      string `json:"status"`
	StatusCode  int    `json:"status_code"`
	Runtime     string `json:"runtime,omitempty"`
	Memory      string `json:"memory,omitempty"`
	TotalTests  int    `json:"total_tests"`
	PassedTests int    `json:"passed_tests"`
	Error       string `json:"error,omitempty"`
	Note        string `json:"note,omitempty"`
	SubmittedAt string `json:"submitted_at"`
}

type ExportStreak struct {
	Current  int    `json:"current"`
	Longest  int    `json:"longest"`
	LastDate string `json:"last_date,omitempty"`
}

type ExportRewardEvent struct {
	ProblemID int    `json:"problem_id"`
	Kind      string `json:"kind"`
	XP        int    `json:"xp"`
	CreatedAt string `json:"created_at"`
}

func (s *SQLiteStore) Export(ctx context.Context) (*ExportData, error) {
	data := &ExportData{
		Version:    1,
		ExportedAt: time.Now(),
	}

	progress, err := s.GetAllProgress(ctx)
	if err != nil {
		return nil, fmt.Errorf("export progress: %w", err)
	}
	for id, status := range progress {
		data.Progress = append(data.Progress, ExportProgress{
			ProblemID: id,
			Status:    string(status),
		})
	}

	for _, p := range data.Progress {
		attempts, err := s.GetAttempts(ctx, p.ProblemID)
		if err != nil {
			return nil, fmt.Errorf("export attempts for problem %d: %w", p.ProblemID, err)
		}
		for _, a := range attempts {
			ea := ExportAttempt{
				ProblemID:  a.ProblemID,
				Timestamp:  a.Timestamp.Format(time.RFC3339),
				DurationMs: a.Duration.Milliseconds(),
				Passed:     a.Passed,
			}
			if a.SelfReported != "" {
				ea.SelfReported = string(a.SelfReported)
			}
			data.Attempts = append(data.Attempts, ea)
		}
	}

	logs, err := s.GetSolveLogs(ctx)
	if err != nil {
		return nil, fmt.Errorf("export solve logs: %w", err)
	}
	for _, log := range logs {
		data.SolveLogs = append(data.SolveLogs, ExportSolveLog{
			ProblemID:   log.ProblemID,
			Slug:        log.Slug,
			Language:    log.Language,
			Status:      log.Status,
			StatusCode:  log.StatusCode,
			Runtime:     log.Runtime,
			Memory:      log.Memory,
			TotalTests:  log.TotalTests,
			PassedTests: log.PassedTests,
			Error:       log.Error,
			Note:        log.Note,
			SubmittedAt: log.SubmittedAt.Format(time.RFC3339),
		})
	}

	rewardEvents, err := s.getAllRewardEvents(ctx)
	if err != nil {
		return nil, fmt.Errorf("export reward events: %w", err)
	}
	for _, e := range rewardEvents {
		data.RewardEvents = append(data.RewardEvents, ExportRewardEvent{
			ProblemID: e.ProblemID,
			Kind:      e.Kind,
			XP:        e.XP,
			CreatedAt: e.CreatedAt.Format(time.RFC3339),
		})
	}

	stats, err := s.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("export stats: %w", err)
	}
	data.XP = stats.TotalXP
	data.Streak = ExportStreak{
		Current: stats.Streak,
		Longest: stats.LongestStreak,
	}

	days, err := s.GetStreakDays(ctx)
	if err != nil {
		return nil, fmt.Errorf("export streak days: %w", err)
	}
	for _, d := range days {
		data.StreakDays = append(data.StreakDays, d.Format("2006-01-02"))
	}

	achievements, err := s.GetAchievements(ctx)
	if err != nil {
		return nil, fmt.Errorf("export achievements: %w", err)
	}
	data.Achievements = achievements

	return data, nil
}

func (s *SQLiteStore) Import(ctx context.Context, data *ExportData) error {
	for _, p := range data.Progress {
		if err := s.SetProgress(ctx, p.ProblemID, roadmap.Status(p.Status)); err != nil {
			return fmt.Errorf("import progress: %w", err)
		}
	}

	for _, a := range data.Attempts {
		ts, err := time.Parse(time.RFC3339, a.Timestamp)
		if err != nil {
			return fmt.Errorf("parse attempt timestamp %q: %w", a.Timestamp, err)
		}
		record := &AttemptRecord{
			ProblemID:    a.ProblemID,
			Timestamp:    ts,
			Duration:     time.Duration(a.DurationMs) * time.Millisecond,
			Passed:       a.Passed,
			SelfReported: roadmap.Difficulty(a.SelfReported),
		}
		if err := s.RecordAttempt(ctx, record); err != nil {
			return fmt.Errorf("import attempt: %w", err)
		}
	}

	for _, log := range data.SolveLogs {
		submittedAt, err := time.Parse(time.RFC3339, log.SubmittedAt)
		if err != nil {
			return fmt.Errorf("parse solve log timestamp %q: %w", log.SubmittedAt, err)
		}
		record := &SolveLogRecord{
			ProblemID:   log.ProblemID,
			Slug:        log.Slug,
			Language:    log.Language,
			Status:      log.Status,
			StatusCode:  log.StatusCode,
			Runtime:     log.Runtime,
			Memory:      log.Memory,
			TotalTests:  log.TotalTests,
			PassedTests: log.PassedTests,
			Error:       log.Error,
			Note:        log.Note,
			SubmittedAt: submittedAt,
		}
		if err := s.RecordSolveLog(ctx, record); err != nil {
			return fmt.Errorf("import solve log: %w", err)
		}
	}

	for _, e := range data.RewardEvents {
		createdAt, err := time.Parse(time.RFC3339, e.CreatedAt)
		if err != nil {
			return fmt.Errorf("parse reward event timestamp %q: %w", e.CreatedAt, err)
		}
		event := &RewardEvent{
			ProblemID: e.ProblemID,
			Kind:      e.Kind,
			XP:        e.XP,
			CreatedAt: createdAt,
		}
		if err := s.RecordRewardEvent(ctx, event); err != nil {
			return fmt.Errorf("import reward event: %w", err)
		}
	}

	if data.XP > 0 {
		if err := s.AddXP(ctx, data.XP); err != nil {
			return fmt.Errorf("import xp: %w", err)
		}
	}

	for _, id := range data.Achievements {
		if err := s.UnlockAchievement(ctx, id); err != nil {
			return fmt.Errorf("import achievement: %w", err)
		}
	}

	return nil
}

func (s *SQLiteStore) getAllRewardEvents(ctx context.Context) ([]*RewardEvent, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT problem_id, kind, xp, created_at FROM reward_events ORDER BY problem_id, created_at")
	if err != nil {
		return nil, fmt.Errorf("get all reward events: %w", err)
	}
	defer rows.Close()

	var events []*RewardEvent
	for rows.Next() {
		var e RewardEvent
		if err := rows.Scan(&e.ProblemID, &e.Kind, &e.XP, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan reward event: %w", err)
		}
		events = append(events, &e)
	}
	return events, rows.Err()
}

func ExportToFile(ctx context.Context, s *SQLiteStore, path string) error {
	data, err := s.Export(ctx)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal export: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0o644); err != nil {
		return fmt.Errorf("write export: %w", err)
	}

	return nil
}

func ImportFromFile(ctx context.Context, s *SQLiteStore, path string) error {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read import: %w", err)
	}

	var data ExportData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("unmarshal import: %w", err)
	}

	return s.Import(ctx, &data)
}
