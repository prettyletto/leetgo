# Profile and Preferences in Config

Profile and Practice Preferences are stored in `config.toml`, while progress history remains in SQLite. These values are needed before the TUI opens, are user-editable preferences rather than historical records, and should remain inspectable without database tooling.

**Considered Options**: Store everything in SQLite, split Profile into SQLite and preferences into config, store Profile and preferences in config.
**Consequences**: Onboarding primarily writes config, while SQLite continues to own progress, Attempts, Solve Logs, XP, Streaks, and Achievements.
