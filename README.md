# leetgo

Leetgo is a local-first Go CLI/TUI for structured LeetCode practice. It gives you a curated DAG roadmap, generates problem workspaces and tests, tracks progress in SQLite, awards XP, and keeps the next useful action visible from the terminal.

## Status

`v0.1.0` is the first public release candidate. The core local practice loop is ready: onboarding, roadmap navigation, problem generation, local test runs, manual solves, progress tracking, and browser-assisted LeetCode Session setup.

## Install

### Go install

```bash
go install github.com/prettyletto/leetgo/cmd/leetgo@v0.1.0
```

For the latest commit on the default branch:

```bash
go install github.com/prettyletto/leetgo/cmd/leetgo@latest
```

### Build from source

```bash
git clone https://github.com/prettyletto/leetgo.git
cd leetgo
go build -o leetgo ./cmd/leetgo
./leetgo
```

## Requirements

- Go 1.26 or newer.
- A terminal with at least `60x18` cells for the TUI.
- Optional: Chrome, Chromium, Brave, Edge, or Vivaldi for `leetgo auth`.
- Optional: language toolchains for generated problem tests, depending on your configured language.

## Quick Start

```bash
leetgo init
leetgo
```

The first TUI launch walks you through onboarding: display name, workspace path, language, roadmap, theme, and optional LeetCode Session setup.

Default local paths:

```text
~/.leetgo/config.toml
~/.leetgo/leetgo.db
~/leetgo-workspace/
```

## TUI Workflow

Leetgo opens on a dashboard with your next recommended action. From there:

- Open the roadmap to see stages, locked gates, solved Problems, and upcoming blockers.
- Open a Problem Detail screen for the brief, prerequisites, unlock impact, workspace files, and progress actions.
- Start a Problem to generate a Stub and TestSuite in your workspace.
- Run local tests, submit to LeetCode, or mark a manual solve when needed.

Common keys:

| Key | Action |
| --- | --- |
| `j` / `k`, arrows | Move focus |
| `enter` | Primary action |
| `r` | Open roadmap from dashboard |
| `l` | Open/close legacy list mode |
| `h` / `l`, `<` / `>` | Switch roadmap groups where available |
| `x` | Run local TestSuite on Problem Detail |
| `s` | Submit on Problem Detail |
| `m` | Mark manual solve on Problem Detail |
| `esc` / `backspace` | Go back |
| `q` | Quit |

## CLI Commands

```bash
leetgo                         # Launch TUI
leetgo start                   # Launch TUI
leetgo init                    # Write default config
leetgo auth                    # Browser-assisted LeetCode Session setup
leetgo status                  # Print progress summary
leetgo next                    # Print next recommended action
leetgo info <id-or-slug>       # Show Problem metadata
leetgo test <id-or-slug>       # Run generated local tests
leetgo submit <id-or-slug>     # Submit current Stub to LeetCode
leetgo solve <id-or-slug>      # Record a manual solve
leetgo solve-log               # Print practice history
leetgo paths                   # Show active config/database/session paths
leetgo roadmap list            # List bundled roadmaps
leetgo config                  # Print config
leetgo export                  # Export local data to JSON
leetgo import <file>           # Import exported JSON
leetgo git-export <repo>       # Export portable progress files to a Git repo
```

## Configuration

Config is stored as TOML at `~/.leetgo/config.toml`.

Useful fields:

```toml
display_name = "Ada"
workspace = "~/leetgo-workspace"
language = "go"
roadmap = "from-zero-to-hero"
theme = "rpg-skill-tree"
editor = "nvim"
```

Editor resolution order is `editor`, then `$VISUAL`, then `$EDITOR`.

Supported generator languages in this release include Go, Python, TypeScript, JavaScript, Java, C++, Rust, and C#.

## Data And Privacy

Leetgo is local-first. Progress, XP, attempts, review cycles, and LeetCode Session credentials are stored under `~/.leetgo/`.

```text
~/.leetgo/
  config.toml
  leetgo.db
  session.json
  exports/
```

Do not commit `~/.leetgo/`, `session.json`, local database files, or generated workspace directories.

Run `leetgo paths` after installing or updating to confirm the binary is using the same `~/.leetgo/leetgo.db`. If progress appears missing, compare `leetgo paths` from the old and new binary; a different `Home` or `Data dir` means the binary is running under a different user or environment.

## LeetCode Session Setup

`leetgo auth` launches a Chromium-based browser, lets you log in to LeetCode normally, reads the required cookies through the browser debugging protocol, and writes them to `~/.leetgo/session.json`.

There is no password storage and no OAuth flow. If auth fails, rerun `leetgo auth` and complete login in the opened browser window.

## Development

```bash
go test ./...
go vet ./...
go build -o leetgo ./cmd/leetgo
```

Project layout:

```text
cmd/leetgo/          CLI entry point
internal/catalog/    embedded roadmap catalog
internal/config/     TOML config
internal/generator/  stubs and tests per language
internal/leetcode/   Session setup and submission client
internal/roadmap/    DAG and prerequisite logic
internal/store/      SQLite persistence
internal/tui/        Bubbletea screens
internal/workspace/  generated workspace management
```

## Release

This repository uses semantic version tags. The first release is `v0.1.0`.

## License

MIT. See `LICENSE`.
