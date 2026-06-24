CREATE TABLE IF NOT EXISTS solve_provenance (
    problem_id INTEGER PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('manual', 'accepted')),
    note TEXT,
    solve_log_id INTEGER,
    solved_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (solve_log_id) REFERENCES solve_logs(id)
);
