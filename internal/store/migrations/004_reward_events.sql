CREATE TABLE IF NOT EXISTS reward_events (
    problem_id INTEGER NOT NULL,
    kind TEXT NOT NULL,
    xp INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (problem_id, kind)
);
