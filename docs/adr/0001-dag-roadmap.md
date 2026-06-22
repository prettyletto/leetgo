# DAG-based Roadmap

The roadmap is modeled as a directed acyclic graph (DAG) where Problems are nodes and prerequisite relationships are edges. This shapes the entire unlock/progression system, the TUI graph visualization, and the gamification engine.

We chose a DAG over a linear sequence or flat categorized list because it captures the real dependency structure between algorithmic concepts (e.g., you need Two Sum before 3Sum, BFS before shortest-path). It also enables the skill-tree-like unlock mechanic that drives gamification.

**Considered Options**: Linear sequence, categorized list with ordering, full DAG.
**Consequences**: Requires topological sorting for display, cycle detection on catalog edits, and prerequisite checking before allowing a Problem to be started.
