# Single Adaptive TUI Appearance System

Leetgo's TUI now uses a single adaptive appearance system instead of multiple branded Themes.

**Considered Options**: keep the RPG/Clean/Cyber theme set, preserve alternate themes but simplify labels, move to one adaptive Theme.

**Decision**: move to one adaptive Theme that derives contrast and emphasis from the terminal environment while keeping Symbol Mode and Motion Preference as the user-facing appearance controls.

**Consequences**:

- product copy no longer branches by theme identity
- onboarding no longer asks the user to pick a Theme
- legacy saved theme values must auto-migrate to `adaptive`
- shared TUI components should use semantic tokens rather than theme-specific branches
- future polish work should improve layout density and hierarchy inside the adaptive system rather than introducing new branded Themes
