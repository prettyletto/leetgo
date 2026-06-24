CREATE TABLE IF NOT EXISTS review_cycles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    problem_id INTEGER NOT NULL,
    reason TEXT NOT NULL CHECK (reason IN ('weakness', 'failed_attempts', 'manual_solve_validation', 'prerequisite_refresh')),
    roadmap_id TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    rewarded_at DATETIME
);
