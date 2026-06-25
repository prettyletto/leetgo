# Config Schema Target

Profile and Practice Preferences live in `~/.leetgo/config.toml`.

## Target Fields

```toml
onboarding_complete = true
display_name = "Ada"
workspace = "/home/ada/leetgo-workspace"
editor = "nvim"
language = "go"
roadmap = "from-zero-to-hero"
theme = "adaptive"
symbol_mode = "rich"
motion_preference = "normal"
git_export_enabled = true
git_export_repo = "/home/ada/progress"
```

## Field Rules

- `onboarding_complete`: explicit completion marker for Onboarding.
- `display_name`: required after Onboarding; max 40 visible characters.
- `workspace`: required path where generated Problem files live.
- `editor`: optional; fallback remains `$VISUAL`, then `$EDITOR`.
- `language`: must match a supported generator Language.
- `roadmap`: must match a bundled Roadmap ID.
- `theme`: currently `adaptive` only.
- `symbol_mode`: must be `rich` or `plain`.
- `motion_preference`: must be `normal`, `reduced`, or `off`.
- `git_export_enabled`: whether Git Export backup is enabled.
- `git_export_repo`: required when `git_export_enabled` is true.

## Backward Compatibility

Legacy configs may still contain older theme values such as `rpg-skill-tree`, `clean-productivity`, or `cyber-dashboard`.

Rules:

- Existing configs without onboarding fields should still prefill Onboarding.
- Legacy theme values should auto-migrate to `adaptive` on load/default application.
- The TUI should no longer expose user-facing theme switching.

## Validation

Config validation should fail when:

- Workspace is empty.
- Language is unsupported.
- Roadmap is unknown.
- Theme is neither `adaptive` nor a known legacy value that can be migrated.
- Symbol mode is unknown.
- Motion preference is unknown.
- Onboarding is complete but Display Name is empty.
- Git Export is enabled but repo path is empty or invalid.

## CLI Updates

Existing command:

```bash
leetgo config set <key> <value>
```

Should support these keys:

- `display-name`
- `theme`
- `symbol-mode`
- `motion`
- `git-export-enabled`
- `git-export-repo`

Behavior note:

- Setting `theme` to any legacy value should normalize to `adaptive`.
