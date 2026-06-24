package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/prettyletto/leetgo/internal/roadmap"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set journal mode: %w", err)
	}

	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

func (s *SQLiteStore) migrate() error {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, entry := range entries {
		data, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		if _, err := s.db.Exec(string(data)); err != nil {
			return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) GetProgress(ctx context.Context, problemID int) (*Progress, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT problem_id, status, updated_at FROM progress WHERE problem_id = ?", problemID)

	var p Progress
	var status string
	err := row.Scan(&p.ProblemID, &status, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get progress: %w", err)
	}
	p.Status = roadmap.Status(status)
	return &p, nil
}

func (s *SQLiteStore) SetProgress(ctx context.Context, problemID int, status roadmap.Status) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO progress (problem_id, status, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(problem_id) DO UPDATE SET status = excluded.status, updated_at = excluded.updated_at`,
		problemID, string(status), time.Now())
	if err != nil {
		return fmt.Errorf("set progress: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetAllProgress(ctx context.Context) (map[int]roadmap.Status, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT problem_id, status FROM progress")
	if err != nil {
		return nil, fmt.Errorf("get all progress: %w", err)
	}
	defer rows.Close()

	result := make(map[int]roadmap.Status)
	for rows.Next() {
		var id int
		var status string
		if err := rows.Scan(&id, &status); err != nil {
			return nil, fmt.Errorf("scan progress: %w", err)
		}
		result[id] = roadmap.Status(status)
	}
	return result, rows.Err()
}

func (s *SQLiteStore) RecordAttempt(ctx context.Context, attempt *AttemptRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO attempts (problem_id, timestamp, duration_ms, passed, self_reported)
		 VALUES (?, ?, ?, ?, ?)`,
		attempt.ProblemID, attempt.Timestamp, attempt.Duration.Milliseconds(),
		attempt.Passed, attempt.SelfReported)
	if err != nil {
		return fmt.Errorf("record attempt: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetAttempts(ctx context.Context, problemID int) ([]*AttemptRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, problem_id, timestamp, duration_ms, passed, self_reported FROM attempts WHERE problem_id = ? ORDER BY timestamp DESC",
		problemID)
	if err != nil {
		return nil, fmt.Errorf("get attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*AttemptRecord
	for rows.Next() {
		var a AttemptRecord
		var durationMs int64
		var selfReported sql.NullString
		if err := rows.Scan(&a.ID, &a.ProblemID, &a.Timestamp, &durationMs, &a.Passed, &selfReported); err != nil {
			return nil, fmt.Errorf("scan attempt: %w", err)
		}
		a.Duration = time.Duration(durationMs) * time.Millisecond
		if selfReported.Valid {
			a.SelfReported = roadmap.Difficulty(selfReported.String)
		}
		attempts = append(attempts, &a)
	}
	return attempts, rows.Err()
}

func (s *SQLiteStore) RecordSolveLog(ctx context.Context, log *SolveLogRecord) error {
	if log.SubmittedAt.IsZero() {
		log.SubmittedAt = time.Now()
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO solve_logs (
			problem_id, slug, language, status, status_code, runtime, memory,
			total_tests, passed_tests, error, note, submitted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.ProblemID, log.Slug, log.Language, log.Status, log.StatusCode,
		log.Runtime, log.Memory, log.TotalTests, log.PassedTests, log.Error,
		log.Note, log.SubmittedAt)
	if err != nil {
		return fmt.Errorf("record solve log: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetSolveLogs(ctx context.Context) ([]*SolveLogRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, problem_id, slug, language, status, status_code, runtime, memory,
			total_tests, passed_tests, error, note, submitted_at
		 FROM solve_logs ORDER BY submitted_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("get solve logs: %w", err)
	}
	defer rows.Close()

	var logs []*SolveLogRecord
	for rows.Next() {
		var log SolveLogRecord
		var runtime, memory, errText, note sql.NullString
		if err := rows.Scan(&log.ID, &log.ProblemID, &log.Slug, &log.Language, &log.Status,
			&log.StatusCode, &runtime, &memory, &log.TotalTests, &log.PassedTests,
			&errText, &note, &log.SubmittedAt); err != nil {
			return nil, fmt.Errorf("scan solve log: %w", err)
		}
		if runtime.Valid {
			log.Runtime = runtime.String
		}
		if memory.Valid {
			log.Memory = memory.String
		}
		if errText.Valid {
			log.Error = errText.String
		}
		if note.Valid {
			log.Note = note.String
		}
		logs = append(logs, &log)
	}
	return logs, rows.Err()
}

func (s *SQLiteStore) GetSolveLogsForProblem(ctx context.Context, problemID int) ([]*SolveLogRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, problem_id, slug, language, status, status_code, runtime, memory,
			total_tests, passed_tests, error, note, submitted_at
		 FROM solve_logs WHERE problem_id = ? ORDER BY submitted_at DESC, id DESC`, problemID)
	if err != nil {
		return nil, fmt.Errorf("get solve logs for problem: %w", err)
	}
	defer rows.Close()

	var logs []*SolveLogRecord
	for rows.Next() {
		var log SolveLogRecord
		var runtime, memory, errText, note sql.NullString
		if err := rows.Scan(&log.ID, &log.ProblemID, &log.Slug, &log.Language, &log.Status,
			&log.StatusCode, &runtime, &memory, &log.TotalTests, &log.PassedTests,
			&errText, &note, &log.SubmittedAt); err != nil {
			return nil, fmt.Errorf("scan solve log: %w", err)
		}
		if runtime.Valid {
			log.Runtime = runtime.String
		}
		if memory.Valid {
			log.Memory = memory.String
		}
		if errText.Valid {
			log.Error = errText.String
		}
		if note.Valid {
			log.Note = note.String
		}
		logs = append(logs, &log)
	}
	return logs, rows.Err()
}

func (s *SQLiteStore) RecordSolveProvenance(ctx context.Context, sp *SolveProvenance) error {
	if sp.SolvedAt.IsZero() {
		sp.SolvedAt = time.Now()
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO solve_provenance (problem_id, kind, note, solve_log_id, solved_at)
		 VALUES (?, ?, ?, ?, ?)`,
		sp.ProblemID, sp.Kind, nullString(sp.Note), sp.SolveLogID, sp.SolvedAt)
	if err != nil {
		return fmt.Errorf("record solve provenance: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetSolveProvenance(ctx context.Context, problemID int) (*SolveProvenance, error) {
	var sp SolveProvenance
	var note sql.NullString
	var solveLogID sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		"SELECT problem_id, kind, note, solve_log_id, solved_at FROM solve_provenance WHERE problem_id = ?",
		problemID).Scan(&sp.ProblemID, &sp.Kind, &note, &solveLogID, &sp.SolvedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get solve provenance: %w", err)
	}
	if note.Valid {
		sp.Note = note.String
	}
	if solveLogID.Valid {
		id := int(solveLogID.Int64)
		sp.SolveLogID = &id
	}
	return &sp, nil
}

func (s *SQLiteStore) GetSolveProvenanceAll(ctx context.Context) (map[int]*SolveProvenance, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT problem_id, kind, note, solve_log_id, solved_at FROM solve_provenance")
	if err != nil {
		return nil, fmt.Errorf("get all solve provenance: %w", err)
	}
	defer rows.Close()

	result := make(map[int]*SolveProvenance)
	for rows.Next() {
		var sp SolveProvenance
		var note sql.NullString
		var solveLogID sql.NullInt64
		if err := rows.Scan(&sp.ProblemID, &sp.Kind, &note, &solveLogID, &sp.SolvedAt); err != nil {
			return nil, fmt.Errorf("scan solve provenance: %w", err)
		}
		if note.Valid {
			sp.Note = note.String
		}
		if solveLogID.Valid {
			id := int(solveLogID.Int64)
			sp.SolveLogID = &id
		}
		result[sp.ProblemID] = &sp
	}
	return result, rows.Err()
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (s *SQLiteStore) CreateReviewCycle(ctx context.Context, rc *ReviewCycle) error {
	if rc.CreatedAt.IsZero() {
		rc.CreatedAt = time.Now()
	}
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO review_cycles (problem_id, reason, roadmap_id, created_at) VALUES (?, ?, ?, ?)",
		rc.ProblemID, rc.Reason, rc.RoadmapID, rc.CreatedAt)
	if err != nil {
		return fmt.Errorf("create review cycle: %w", err)
	}
	id, _ := result.LastInsertId()
	rc.ID = int(id)
	return nil
}

func (s *SQLiteStore) GetReviewCycles(ctx context.Context) ([]*ReviewCycle, error) {
	return s.queryReviewCycles(ctx, "SELECT id, problem_id, reason, roadmap_id, created_at, completed_at, rewarded_at FROM review_cycles ORDER BY created_at DESC")
}

func (s *SQLiteStore) GetReviewCyclesForProblem(ctx context.Context, problemID int) ([]*ReviewCycle, error) {
	return s.queryReviewCycles(ctx, "SELECT id, problem_id, reason, roadmap_id, created_at, completed_at, rewarded_at FROM review_cycles WHERE problem_id = ? ORDER BY created_at DESC", problemID)
}

func (s *SQLiteStore) queryReviewCycles(ctx context.Context, query string, args ...interface{}) ([]*ReviewCycle, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query review cycles: %w", err)
	}
	defer rows.Close()

	var cycles []*ReviewCycle
	for rows.Next() {
		var rc ReviewCycle
		var completedAt, rewardedAt sql.NullTime
		if err := rows.Scan(&rc.ID, &rc.ProblemID, &rc.Reason, &rc.RoadmapID, &rc.CreatedAt, &completedAt, &rewardedAt); err != nil {
			return nil, fmt.Errorf("scan review cycle: %w", err)
		}
		if completedAt.Valid {
			rc.CompletedAt = &completedAt.Time
		}
		if rewardedAt.Valid {
			rc.RewardedAt = &rewardedAt.Time
		}
		cycles = append(cycles, &rc)
	}
	return cycles, rows.Err()
}

func (s *SQLiteStore) CompleteReviewCycle(ctx context.Context, id int) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, "UPDATE review_cycles SET completed_at = ? WHERE id = ? AND completed_at IS NULL", now, id)
	if err != nil {
		return fmt.Errorf("complete review cycle: %w", err)
	}
	return nil
}

func (s *SQLiteStore) RewardReviewCycle(ctx context.Context, id int) error {
	now := time.Now()
	_, err := s.db.ExecContext(ctx, "UPDATE review_cycles SET rewarded_at = ? WHERE id = ? AND rewarded_at IS NULL", now, id)
	if err != nil {
		return fmt.Errorf("reward review cycle: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetStats(ctx context.Context) (*Stats, error) {
	var stats Stats

	if err := s.db.QueryRowContext(ctx, "SELECT total FROM xp WHERE id = 1").Scan(&stats.TotalXP); err != nil {
		return nil, fmt.Errorf("get xp: %w", err)
	}
	stats.Level = XPToLevel(stats.TotalXP)

	row := s.db.QueryRowContext(ctx, "SELECT current, longest FROM streak WHERE id = 1")
	if err := row.Scan(&stats.Streak, &stats.LongestStreak); err != nil {
		return nil, fmt.Errorf("get streak: %w", err)
	}

	if err := s.db.QueryRowContext(ctx,
		`SELECT
			COUNT(CASE WHEN status = 'verified' THEN 1 END) as verified,
			COUNT(CASE WHEN status = 'solved' THEN 1 END) as solved
		 FROM progress`).Scan(&stats.Verified, &stats.Solved); err != nil {
		return nil, fmt.Errorf("get solved/verified count: %w", err)
	}

	return &stats, nil
}

func (s *SQLiteStore) AddXP(ctx context.Context, amount int) error {
	_, err := s.db.ExecContext(ctx, "UPDATE xp SET total = total + ? WHERE id = 1", amount)
	if err != nil {
		return fmt.Errorf("add xp: %w", err)
	}
	return nil
}

func (s *SQLiteStore) UpdateStreak(ctx context.Context) error {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var lastDate sql.NullString
	var current int
	if err := tx.QueryRowContext(ctx, "SELECT last_date, current FROM streak WHERE id = 1").Scan(&lastDate, &current); err != nil {
		return fmt.Errorf("get streak: %w", err)
	}

	if lastDate.Valid && lastDate.String == today {
		return nil
	}

	if lastDate.Valid && lastDate.String == yesterday {
		current++
	} else {
		current = 1
	}

	var longest int
	if err := tx.QueryRowContext(ctx, "SELECT longest FROM streak WHERE id = 1").Scan(&longest); err != nil {
		return fmt.Errorf("get longest: %w", err)
	}
	if current > longest {
		longest = current
	}

	if _, err := tx.ExecContext(ctx, "UPDATE streak SET current = ?, longest = ?, last_date = ? WHERE id = 1", current, longest, today); err != nil {
		return fmt.Errorf("update streak: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO streak_days (date) VALUES (?)", today); err != nil {
		return fmt.Errorf("record streak day: %w", err)
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetStreakDays(ctx context.Context) ([]time.Time, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT date FROM streak_days ORDER BY date")
	if err != nil {
		return nil, fmt.Errorf("get streak days: %w", err)
	}
	defer rows.Close()

	var days []time.Time
	for rows.Next() {
		var dateStr string
		if err := rows.Scan(&dateStr); err != nil {
			return nil, fmt.Errorf("scan streak day: %w", err)
		}
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, fmt.Errorf("parse date: %w", err)
		}
		days = append(days, t)
	}
	return days, rows.Err()
}

func (s *SQLiteStore) UnlockAchievement(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT OR IGNORE INTO achievements (id, unlocked_at) VALUES (?, ?)",
		id, time.Now())
	if err != nil {
		return fmt.Errorf("unlock achievement: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetAchievements(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id FROM achievements ORDER BY unlocked_at")
	if err != nil {
		return nil, fmt.Errorf("get achievements: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *SQLiteStore) RecordRewardEvent(ctx context.Context, event *RewardEvent) error {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	_, err := s.db.ExecContext(ctx,
		"INSERT OR IGNORE INTO reward_events (problem_id, kind, xp, created_at) VALUES (?, ?, ?, ?)",
		event.ProblemID, event.Kind, event.XP, event.CreatedAt)
	if err != nil {
		return fmt.Errorf("record reward event: %w", err)
	}
	return nil
}

func (s *SQLiteStore) HasRewardEvent(ctx context.Context, problemID int, kind string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM reward_events WHERE problem_id = ? AND kind = ?",
		problemID, kind).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check reward event: %w", err)
	}
	return count > 0, nil
}

func (s *SQLiteStore) GetRewardEvents(ctx context.Context, problemID int) ([]*RewardEvent, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT problem_id, kind, xp, created_at FROM reward_events WHERE problem_id = ? ORDER BY created_at",
		problemID)
	if err != nil {
		return nil, fmt.Errorf("get reward events: %w", err)
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

func XPToLevel(xp int) int {
	level := 1
	required := 100
	for xp >= required {
		xp -= required
		level++
		required = level * 100
	}
	return level
}

func LevelToXP(level int) int {
	total := 0
	for i := 1; i < level; i++ {
		total += i * 100
	}
	return total
}

func XPForDifficulty(d roadmap.Difficulty) int {
	switch d {
	case roadmap.DifficultyEasy:
		return 10
	case roadmap.DifficultyMedium:
		return 25
	case roadmap.DifficultyHard:
		return 50
	default:
		return 10
	}
}
