# leetgo

A gamified CLI engine for structured LeetCode practice. Generates a DAG-based roadmap, manages problem files and tests, tracks progress with XP/levels/streaks, and surfaces weaknesses — all from the terminal.

## Install

```bash
go install github.com/prettyletto/leetgo/cmd/leetgo@latest
```

Or build from source:

```bash
git clone https://github.com/prettyletto/leetgo
cd leetgo
go build -o leetgo ./cmd/leetgo
```

## Quick Start

```bash
leetgo init          # Create config at ~/.leetgo/config.toml
leetgo               # Launch the TUI
```

## Usage

### TUI

```
leetgo               # Launch interactive roadmap
leetgo start         # Same as above
```

**Keybindings:**

| Key | Action |
|-----|--------|
| `j`/`k` or arrows | Navigate problem list |
| `enter` | Start problem (generates files + opens editor) |
| `m` | Mark problem as solved (awards XP) |
| `s` | Submit solution to LeetCode |
| `g` | Toggle graph view (DAG visualization) |
| `h` | Toggle heatmap view (streak calendar) |
| `/` | Filter/search problems |
| `q` | Quit |

### CLI Commands

```bash
leetgo auth          # Connect LeetCode Session through Chrome/Chromium
leetgo export        # Export progress to JSON (~/.leetgo/exports/)
leetgo import <file> # Import progress from JSON
leetgo status        # Show progress summary
```

## Configuration

Config lives at `~/.leetgo/config.toml`:

```toml
workspace = "~/leetgo-workspace"  # Where problem files are generated
editor = "nvim"                   # Editor command (or set $VISUAL/$EDITOR)
language = "go"                   # go, python, typescript, java
```

## How It Works

1. **Roadmap**: 132 curated problems across 16 categories form a DAG with prerequisite edges. Solving "Two Sum" unlocks "3Sum", etc.
2. **Start**: Press Enter on an available problem. Leetgo generates a stub file and test file in your workspace, then opens them in your editor.
3. **Solve**: Write your solution, run the tests locally, then press `m` to mark it solved. XP is awarded based on difficulty.
4. **Submit**: Press `s` to submit your solution to LeetCode's judge (requires `leetgo auth` first; Linux/Windows need Chrome or another Chromium-based browser).
5. **Track**: Watch your XP bar grow, maintain streaks, unlock achievements, and review weakness analytics.

## Architecture

```
cmd/leetgo/          — CLI entry point
internal/
  roadmap/           — DAG types, topological sort, unlock logic
  catalog/           — Embedded YAML problem graph (132 problems)
  generator/         — Stub + test file generation (Go, Python, TS, Java)
  workspace/         — File system management
  store/             — SQLite persistence (progress, XP, streaks)
  leetcode/          — Session setup + submission client
  tui/               — Bubbletea TUI (list, graph, heatmap views)
  gamification/      — XP, levels, achievements
  analytics/         — Weakness detection, category stats
  config/            — TOML configuration
```

## Data

All data stored locally at `~/.leetgo/`:

```
~/.leetgo/
  config.toml        # Configuration
  leetgo.db          # SQLite database
  session.json       # LeetCode Session credentials
  browser-profile/   # Browser profile used by `leetgo auth`
  exports/           # JSON export/import
```

## LeetCode Session Setup

`leetgo auth` opens Chrome or another Chromium-based browser, lets you log in to LeetCode normally, extracts the required Session cookies through Chrome DevTools Protocol, then closes the browser after a successful connection.

Supported initially: Chrome, Chromium, Brave, Edge, and Vivaldi on Linux and Windows. Firefox and Safari are not supported yet.

## License

MIT
