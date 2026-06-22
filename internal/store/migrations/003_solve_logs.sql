CREATE TABLE IF NOT EXISTS solve_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    problem_id INTEGER NOT NULL,
    slug TEXT NOT NULL,
    language TEXT NOT NULL,
    status TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    runtime TEXT,
    memory TEXT,
    total_tests INTEGER NOT NULL DEFAULT 0,
    passed_tests INTEGER NOT NULL DEFAULT 0,
    error TEXT,
    note TEXT,
    submitted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
