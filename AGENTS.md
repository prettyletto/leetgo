# AGENTS.md

## Project Overview

Leetgo is a Go CLI/TUI application that generates a structured LeetCode practice roadmap, manages problem files and tests, tracks progress with gamification, and surfaces weaknesses. See `CONTEXT.md` for domain terminology.

## Build & Run

```bash
go build -o leetgo ./cmd/leetgo
go run ./cmd/leetgo
go test ./...
go vet ./...
golangci-lint run
```

## Architecture

```
cmd/leetgo/          — main entry point, CLI flag parsing
internal/
  roadmap/           — DAG data structures, prerequisite logic, topological sort
  catalog/           — bundled curated problem graph (YAML/JSON loader)
  generator/         — stub + test file generation per language (Go, Python, TS, Java)
  workspace/         — file system management of the workspace directory
  store/             — SQLite persistence layer (progress, attempts, XP, streaks)
  leetcode/          — LeetCode API client, Session setup, submission
  tui/               — Bubbletea models, views, and updates (list view lives here)
  tui/views/         — graph view, stats view, heatmap view, notifications
  analytics/         — weakness detection, category-level fail rate analysis
  gamification/      — XP calculation, level progression, achievement evaluation
  config/            — user configuration (workspace path, editor, language pref)
```

## Conventions

- **Go 1.26+** with modules. No vendoring.
- **Errors**: Return errors, don't panic. Use `fmt.Errorf("context: %w", err)` for wrapping.
- **Interfaces**: Define interfaces at the consumer site, not the provider.
- **Testing**: Table-driven tests. Use `testify/assert` for assertions. Test files next to source: `foo.go` / `foo_test.go`.
- **No global state**: Pass dependencies explicitly via constructor injection.
- **Domain terms**: Use the exact terms from `CONTEXT.md`. A "Problem" is never a "Question". The "Roadmap" is never a "List".
- **SQLite**: Use `modernc.org/sqlite` (pure Go, no CGO). Schema migrations via embedded SQL files.
- **TUI**: All interactive UI uses `charmbracelet/bubbletea` + `charmbracelet/lipgloss` + `charmbracelet/bubbles`.
- **Config**: TOML format at `~/.leetgo/config.toml`. Use `pelletier/go-toml`.

## Key Decisions

See `docs/adr/` for architectural decision records.

- **DAG roadmap**: Problems form a directed acyclic graph with prerequisite edges.
- **Bundled catalog**: The problem graph ships as embedded YAML, not fetched from an API.
- **Multi-language**: Generator supports Go, Python, TypeScript, and Java at minimum.
- **Bubbletea TUI**: List view (default) + ASCII graph toggle using lipgloss.
- **SQLite + JSON export**: Local persistence with export/import for portability.
- **Browser-assisted Session**: LeetCode submission uses browser-assisted Session setup because LeetCode has no stable public OAuth flow for this CLI use case.
- **Editor**: Config `editor` field first, then `$VISUAL`, then `$EDITOR`. If none set, shows notification.

## File Generation Layout

```
~/leetgo-workspace/
  arrays-hashing/
    1-two-sum/
      two_sum.go          # Stub
      two_sum_test.go     # TestSuite
  sliding-window/
    3-longest-substring/
      longest_substring.go
      longest_substring_test.go
```

## Data Directory

```
~/.leetgo/
  config.toml           # User configuration
  leetgo.db             # SQLite database
  session.json          # LeetCode Session credentials
  exports/              # JSON export/import directory
```
