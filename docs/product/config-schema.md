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
theme = "rpg-skill-tree"
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
- `theme`: must be one of `rpg-skill-tree`, `clean-productivity`, `cyber-dashboard`.
- `git_export_enabled`: whether Git Export backup is enabled.
- `git_export_repo`: required when `git_export_enabled` is true.

## Backward Compatibility

Existing configs may only contain:

```toml
workspace = "..."
editor = "..."
language = "go"
roadmap = "from-zero-to-hero"
```

These configs should not be treated as Onboarding complete. The new Onboarding should prefill existing values and ask only for missing Profile and backup/theme choices.

## Validation

Config validation should fail when:

- Workspace is empty.
- Language is unsupported.
- Roadmap is unknown.
- Theme is unknown.
- Onboarding is complete but Display Name is empty.
- Git Export is enabled but repo path is empty or invalid.

## CLI Updates

Existing command:

```bash
leetgo config set <key> <value>
```

Should support new keys:

- `display-name`
- `theme`
- `git-export-enabled`
- `git-export-repo`

Potential future commands:

```bash
leetgo onboarding reset
leetgo theme list
leetgo theme use cyber-dashboard
```
