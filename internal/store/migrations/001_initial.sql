CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS progress (
    problem_id INTEGER PRIMARY KEY,
    status TEXT NOT NULL DEFAULT 'locked',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    problem_id INTEGER NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    duration_ms INTEGER NOT NULL DEFAULT 0,
    passed BOOLEAN NOT NULL DEFAULT 0,
    self_reported TEXT,
    FOREIGN KEY (problem_id) REFERENCES progress(problem_id)
);

CREATE TABLE IF NOT EXISTS xp (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    total INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS streak (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    current INTEGER NOT NULL DEFAULT 0,
    longest INTEGER NOT NULL DEFAULT 0,
    last_date TEXT
);

CREATE TABLE IF NOT EXISTS streak_days (
    date TEXT PRIMARY KEY
);

INSERT OR IGNORE INTO xp (id, total) VALUES (1, 0);
INSERT OR IGNORE INTO streak (id, current, longest, last_date) VALUES (1, 0, 0, NULL);
INSERT OR IGNORE INTO schema_version (version) VALUES (1);
