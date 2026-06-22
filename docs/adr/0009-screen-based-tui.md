# Screen-Based TUI Architecture

The TUI will move from one expanding model with view modes to a screen-based architecture with explicit Onboarding, Roadmap Selection, Dashboard, Roadmap Detail, Stage Detail, and Problem Detail screens. This supports the Dashboard-first product direction and prevents first-run flows, carousel behavior, dashboard panels, and deeper Roadmap navigation from becoming tangled in one model.

**Considered Options**: Keep extending the current model, build a parallel TUI package, introduce explicit screens in the current TUI package.
**Consequences**: Existing list, graph, heatmap, notification, and stats components should be migrated into screens or reusable view components rather than remaining root-level modes.
